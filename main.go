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
	r.Methods("GET").Path("/v1/user").HandlerFunc(listUsers)
	r.Methods("GET").Path("/v1/user/{id}").HandlerFunc(getUser)
	r.Methods("POST").Path("/v1/user").HandlerFunc(createUser)
	r.Methods("PUT").Path("/v1/user/{id}").HandlerFunc(modifyUser)
	log.Fatal(http.ListenAndServe(*addr, r))
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
	// panic("NIY")
}

func getUser(w http.ResponseWriter, r *http.Request) {
	panic("NIY")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	panic("NIY")
}

func modifyUser(w http.ResponseWriter, r *http.Request) {
	panic("NIY")
}
