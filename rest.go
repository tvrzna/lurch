package main

import (
	"encoding/json"
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
	Name string      `json:"name"`
	Jobs []DomainJob `json:"jobs"`
}

type DomainJob struct {
	Name      string    `json:"name"`
	Status    JobStatus `json:"status"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Output    *string   `json:"output"`
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
	if data, err := json.Marshal(&DomainStatus{Message: message, Code: httpCode}); err == nil {
		w.Write(data)
	} else {
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

		result[i] = s.getProjectDetails(p.name, jobs)
	}

	data, _ := json.Marshal(result)
	w.Write(data)
}

// Get project details and history of jobs
func (s RestService) getProjectDetails(name string, jobs []*Job) DomainProject {
	project := DomainProject{Name: name, Jobs: make([]DomainJob, len(jobs))}
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

	jobs, err := s.c.ListJobs(s.c.OpenProject(projectName))
	if err != nil {
		s.message(w, "could not list jobs", http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(s.getProjectDetails(projectName, jobs))
	w.Write(data)
}

// Start new job
func (s RestService) startJob(projectName string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.message(w, "", http.StatusMethodNotAllowed)
		return
	}
	if s.c.StartJob(s.c.OpenProject(projectName)) {
		s.message(w, "job enqueued", http.StatusOK)
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

	data, _ := json.Marshal(DomainJob{Name: b.name, Status: status, StartDate: b.StartDate(), EndDate: b.EndDate(), Output: &output})
	w.Write(data)
}
