package main

import (
	"net/http"
	"os"
)

func main() {
	conf := LoadConfig(os.Args)
	c := NewContext(conf)

	server := http.NewServeMux()

	server.HandleFunc("/rest/", (&RestService{c: c}).HandleFunc)
	server.HandleFunc("/", NewWebService(c).HandleFunc)

	http.ListenAndServe(c.conf.getServerUri(), server)
}
