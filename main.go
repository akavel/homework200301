package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	addr = flag.String("http", ":8080", "http address to listen on")
)

func main() {
	r := mux.NewRouter()
	r.Methods("GET").HandleFunc("/user", listUsers)
	r.Methods("GET").HandleFunc("/user/{id}", getUser)
	r.Methods("POST").HandleFunc("/user", createUser)
	r.Methods("PUT").HandleFunc("/user/{id}", modifyUser)
	http.Handle("/v1/", r)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
