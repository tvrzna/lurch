package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	conf := LoadConfig(os.Args)
	c := NewContext(conf)

	server := http.NewServeMux()

	server.HandleFunc("/rest/", (&RestService{c: c}).HandleFunc)
	server.HandleFunc("/", NewWebService(c).HandleFunc)

	log.Print("-- lurch started on ", conf.getServerUri())
	http.ListenAndServe(conf.getServerUri(), server)
	log.Print("-- lurch finished")
}
