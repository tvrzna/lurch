package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	ParamAction  = "action"
	ParamProject = "projectName"
	ParamParam   = "param"
	ParamParam2  = "param2"
)

type DomainProject struct {
	Name   string            `json:"name"`
	Jobs   []DomainJob       `json:"jobs"`
	Params map[string]string `json:"params,omitempty"`
}

type DomainJob struct {
	Name      string            `json:"name"`
	Status    JobStatus         `json:"status"`
	StartDate time.Time         `json:"startDate"`
	EndDate   time.Time         `json:"endDate"`
	Output    string            `json:"output,omitempty"`
	Params    map[string]string `json:"params,omitempty"`
}

type DomainStatus struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type RestService struct {
	c *Context
}

func (s RestService) message(w http.ResponseWriter, message string, httpCode int) {
	w.WriteHeader(httpCode)
	if message == "" {
		message = http.StatusText(httpCode)
	}

	e := json.NewEncoder(w)
	if err := e.Encode(&DomainStatus{Message: message, Code: httpCode}); err != nil {
		log.Print("could not marshal error message '", message, "'")
		w.WriteHeader(500)
		w.Write([]byte("system error"))
	}
}

func (s RestService) HandleFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	params := s.parseUrl(r.URL.Path)

	switch params[ParamAction] {
	case "projects":
		if params[ParamProject] == "" {
			s.listProjects(w, r)
			return
		} else {
			s.listProject(params[ParamProject], w, r)
			return
		}
	case "jobs":
		if params[ParamProject] != "" {
			if params[ParamParam] == "start" {
				s.startJob(params[ParamProject], w, r)
				return
			} else if params[ParamParam] == "interrupt" {
				s.interruptJob(params[ParamProject], params[ParamParam2], w, r)
				return
			} else {
				s.jobDetail(params[ParamProject], params[ParamParam], w, r)
				return
			}
		}
	}
	s.message(w, "", http.StatusNotFound)
}

func (s RestService) parseUrl(url string) map[string]string {
	result := make(map[string]string)

	urls := strings.Split(url, "/")
	if len(urls) >= 3 {
		result[ParamAction] = strings.TrimSpace(urls[2])
	}
	if len(urls) >= 4 {
		result[ParamProject] = strings.TrimSpace(urls[3])
	}
	if len(urls) >= 5 {
		result[ParamParam] = strings.TrimSpace(urls[4])
	}
	if len(urls) >= 6 {
		result[ParamParam2] = strings.TrimSpace(urls[5])
	}

	return result
}

// List all projects and provide their job history
func (s RestService) listProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.message(w, "", http.StatusMethodNotAllowed)
		return
	}

	projects, err := s.c.ListProjects()
	if err != nil {
		s.message(w, "could not list projects", http.StatusInternalServerError)
		return
	}
	result := make([]DomainProject, len(projects))

	for i, p := range projects {
		jobs, err := s.c.ListJobs(p)
		if err != nil {
			s.message(w, "could not list jobs", http.StatusInternalServerError)
			return
		}
		p.LoadParams()
		result[i] = s.getProjectDetails(p, jobs)
	}

	e := json.NewEncoder(w)
	e.Encode(result)
}

// Get project details and history of jobs
func (s RestService) getProjectDetails(p *Project, jobs []*Job) DomainProject {
	project := DomainProject{Name: p.name, Jobs: make([]DomainJob, len(jobs)), Params: p.params}
	for j, b := range jobs {
		status := b.Status()
		if s.c.IsBeingBuilt(b) {
			status = InProgress
		}
		project.Jobs[j] = DomainJob{Name: b.name, Status: status, StartDate: b.StartDate(), EndDate: b.EndDate()}
	}
	return project
}

// List project details
func (s RestService) listProject(projectName string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.message(w, "", http.StatusMethodNotAllowed)
		return
	}

	p := s.c.OpenProject(projectName)
	p.LoadParams()
	jobs, err := s.c.ListJobs(p)
	if err != nil {
		s.message(w, "could not list jobs", http.StatusInternalServerError)
		return
	}

	e := json.NewEncoder(w)
	e.Encode(s.getProjectDetails(p, jobs))
}

// Start new job
func (s RestService) startJob(projectName string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.message(w, "", http.StatusMethodNotAllowed)
		return
	}

	var t DomainJob
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&t)

	if buildNo := s.c.StartJob(s.c.OpenProject(projectName), t.Params); buildNo != "" {
		s.message(w, fmt.Sprintf("job #%s enqueued", buildNo), http.StatusOK)
	} else {
		s.message(w, "job could not be enqueued", http.StatusBadRequest)
	}
}

func (s RestService) interruptJob(projectName, jobNumber string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.message(w, "", http.StatusMethodNotAllowed)
		return
	}

	b := s.c.OpenJob(s.c.OpenProject(projectName), jobNumber)
	s.c.Interrupt(b)

	s.message(w, "job interrupted", http.StatusOK)
}

// Gets details of project's job
func (s RestService) jobDetail(projectName, jobNumber string, w http.ResponseWriter, r *http.Request) {
	if jobNumber == "" {
		s.message(w, "", http.StatusNotFound)
		return
	}

	b := s.c.OpenJob(s.c.OpenProject(projectName), jobNumber)

	status := b.Status()
	if s.c.IsBeingBuilt(b) {
		status = InProgress
	}

	output, _ := b.ReadOutput()

	e := json.NewEncoder(w)
	e.Encode(DomainJob{Name: b.name, Status: status, StartDate: b.StartDate(), EndDate: b.EndDate(), Output: output})
}
