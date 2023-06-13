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
	Name   string        `json:"name"`
	Builds []DomainBuild `json:"builds"`
}

type DomainBuild struct {
	Name      string      `json:"name"`
	Status    BuildStatus `json:"status"`
	StartDate time.Time   `json:"startDate"`
	EndDate   time.Time   `json:"endDate"`
	Output    *string     `json:"output"`
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
	w.Header().Add("content-type", "application/json")

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
	case "builds":
		if params[ParamProject] != "" {
			if params[ParamParam] == "build" {
				s.initBuild(params[ParamProject], w, r)
				return
			} else if params[ParamParam] == "interrupt" {
				s.interruptBuild(params[ParamProject], params[ParamParam2], w, r)
				return
			} else {
				s.buildDetail(params[ParamProject], params[ParamParam], w, r)
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

// List all projects and provide their build history
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
		builds, err := s.c.ListBuilds(p)
		if err != nil {
			s.message(w, "could not list builds", http.StatusInternalServerError)
			return
		}

		result[i] = s.getProjectDetails(p.name, builds)
	}

	data, _ := json.Marshal(result)
	w.Write(data)
}

// Get project details and history of builds
func (s RestService) getProjectDetails(name string, builds []*Build) DomainProject {
	project := DomainProject{Name: name, Builds: make([]DomainBuild, len(builds))}
	for j, b := range builds {
		status := b.Status()
		if s.c.IsBeingBuilt(b) {
			status = InProgress
		}
		project.Builds[j] = DomainBuild{Name: b.name, Status: status, StartDate: b.StartDate(), EndDate: b.EndDate()}
	}
	return project
}

// List project details
func (s RestService) listProject(projectName string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.message(w, "", http.StatusMethodNotAllowed)
		return
	}

	builds, err := s.c.ListBuilds(s.c.OpenProject(projectName))
	if err != nil {
		s.message(w, "could not list builds", http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(s.getProjectDetails(projectName, builds))
	w.Write(data)
}

// Initialize new build
func (s RestService) initBuild(projectName string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.message(w, "", http.StatusMethodNotAllowed)
		return
	}
	if s.c.Build(s.c.OpenProject(projectName)) {
		s.message(w, "build enqueued", http.StatusOK)
	} else {
		s.message(w, "build could not be enqueued", http.StatusBadRequest)
	}
}

func (s RestService) interruptBuild(projectName, buildNumber string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.message(w, "", http.StatusMethodNotAllowed)
		return
	}

	b := s.c.OpenBuild(s.c.OpenProject(projectName), buildNumber)
	s.c.Interrupt(b)

	s.message(w, "build interrupted", http.StatusOK)

}

// Gets details of project's build
func (s RestService) buildDetail(projectName, buildNumber string, w http.ResponseWriter, r *http.Request) {
	if buildNumber == "" {
		s.message(w, "", http.StatusNotFound)
		return
	}

	b := s.c.OpenBuild(s.c.OpenProject(projectName), buildNumber)

	status := b.Status()
	if s.c.IsBeingBuilt(b) {
		status = InProgress
	}

	output, _ := b.ReadOutput()

	data, _ := json.Marshal(DomainBuild{Name: b.name, Status: status, StartDate: b.StartDate(), EndDate: b.EndDate(), Output: &output})
	w.Write(data)
}
