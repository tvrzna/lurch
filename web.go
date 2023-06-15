package main

import (
	"embed"
	"html/template"
	"log"
	"mime"
	"net/http"
	"strings"
)

//go:embed www
var www embed.FS

type WebService struct {
	c      *Context
	layout *template.Template
}

func NewWebService(c *Context) *WebService {
	result := &WebService{c: c}

	tpl, err := template.ParseFS(www, "www/template.html")
	if err != nil {
		log.Fatal(err)
	}
	result.layout = tpl

	return result
}

type PageContext struct {
	s              *WebService
	ProjectVersion string
	Projects       []string
}

func (p *PageContext) UrlFor(path string) string {
	return p.s.c.conf.getAppUrl() + "/" + path
}

func (s *WebService) HandleFunc(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/static/") {
		f, _ := www.ReadFile("www" + r.URL.Path)
		w.Header().Set("content-type", s.getMimeType(r.URL.Path))
		w.Write(f)
		return
	} else if r.URL.Path == "" || r.URL.Path == "/" || r.URL.Path == "index.html" {
		s.loadIndex(w, r)
	} else {
		// TODO: handle the other files
		w.WriteHeader(404)
	}
}

func (s *WebService) loadIndex(w http.ResponseWriter, r *http.Request) {
	p := &PageContext{s: s, ProjectVersion: s.c.conf.GetVersion()}
	w.Header().Set("content-type", "text/html")

	projects, err := s.c.ListProjects()
	if err != nil {
		w.WriteHeader(500)
		log.Println(err)
	}

	for _, proj := range projects {
		p.Projects = append(p.Projects, proj.name)
	}

	if err := s.layout.Execute(w, p); err != nil {
		w.WriteHeader(500)
		log.Println(err)
	}
}

func (s *WebService) getMimeType(path string) string {
	return mime.TypeByExtension(path[strings.LastIndex(path, "."):])
}
