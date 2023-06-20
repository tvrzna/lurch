package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
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
	} else if strings.HasPrefix(r.URL.Path, "/download") {
		s.downloadArtifact(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *WebService) loadIndex(w http.ResponseWriter, r *http.Request) {
	p := &PageContext{s: s, ProjectVersion: s.c.conf.GetVersion()}
	w.Header().Set("content-type", "text/html")

	projects, err := s.c.ListProjects()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}

	for _, proj := range projects {
		p.Projects = append(p.Projects, proj.name)
	}

	if err := s.layout.Execute(w, p); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
}

func (s *WebService) downloadArtifact(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) > 3 {
		p := s.c.OpenProject(path[2])
		j := s.c.OpenJob(p, path[3])

		f, err := os.Open(j.ArtifactPath())
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		defer f.Close()

		w.Header().Set("content-type", s.getMimeType(j.ArtifactPath()))
		w.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s_%s.tar.gz\"", p.name, j.name))

		w.WriteHeader(http.StatusOK)
		if _, err := io.Copy(w, f); err != nil {
			return
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (s *WebService) getMimeType(path string) string {
	return mime.TypeByExtension(path[strings.LastIndex(path, "."):])
}
