package main

import (
	"net/http"
)

func main() {
	c := NewContext(5000, "workdir")

	http.HandleFunc("/rest/", (&RestService{c: c}).HandleFunc)
	http.HandleFunc("/", NewWebService(c).HandleFunc)

	http.ListenAndServe(c.getServerUri(), nil)
}
