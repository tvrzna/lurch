package main

import (
	"net/http"
)

func main() {
	c := NewContext("workdir")

	http.HandleFunc("/rest/", (&RestService{c: c}).HandleFunc)

	http.ListenAndServe(":5000", nil)
}
