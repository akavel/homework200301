package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-pg/pg/v9"
	"github.com/gorilla/mux"
)

var (
	addr = flag.String("http", ":8080", "http address to listen on")
)

var (
	// FIXME: don't use global var, create proper wrapper object instead
	db Database
)

func main() {
	// TODO: write tests, run with `go test -race`

	// FIXME: pass Postgres options via env vars (esp. password) - probably as a standard connection string
	// db = NewMockDB()
	var err error
	db, err = ConnectPostgres(&pg.Options{
		Addr:            "localhost:5432",
		User:            "homework",
		Password:        "DazBMyGQdKqKG",
		Database:        "users_db",
		ApplicationName: "users_go",
		// TODO: [LATER] add timeouts etc.
	})
	if err != nil {
		log.Fatalf("initializing Postgres DB: %s", err)
	}
	// TODO: [LATER] Close() will never happen now (needs HTTP server soft shutdown)
	// TODO: [LATER] log any error from Close()
	defer db.Close()

	r := mux.NewRouter()
	r.Methods("GET").Path("/v1/user").HandlerFunc(listUsers)
	r.Methods("GET").Path("/v1/user/{id}").HandlerFunc(getUser)
	r.Methods("POST").Path("/v1/user").HandlerFunc(createUser)
	r.Methods("PUT").Path("/v1/user/{id}").HandlerFunc(modifyUser)
	r.Methods("DELETE").Path("/v1/user/{id}").HandlerFunc(deleteUser)
	log.Fatal(http.ListenAndServe(*addr, r))
}

// TODO: [LATER] introduce Context to methods, to allow timeouts control
type Database interface {
	ListUsers() ([]*User, error) // TODO: listing options (query)
	GetUser(email string) (*User, error)
	CreateUser(u *User) error
	DeleteUser(email string) error

	Close() error
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: query: technology=*|php|go|java|js,active=true|false|*
	// TODO: query: pagination - ideally automatically mapped to the Postgres query & to the response (UsersList type? HTTP headers?)
	// TODO: Content-Type, Accepted

	users, err := db.ListUsers()
	if err != nil {
		// TODO: return JSON-formatted errors?
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: [LATER] maybe log error if Fprintf failed, here and everywhere else
		fmt.Fprint(w, "error: ", err)
		return
	}
	// Let's print `[]` instead of `null` in the JSON response in case of empty results list
	if users == nil {
		users = []*User{}
	}

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Print("listUsers: ", err)
		// TODO: if not too late, write 500 to w
		return
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	// TODO: quick fail if id empty or invalid?

	found, err := db.GetUser(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "error: ", err)
		return
	}

	if found == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(found)
	if err != nil {
		log.Printf("getUser[%q]: %s", id, err)
		// TODO: if not too late, write 500 to w
		return
	}
}

var validTechnology = map[string]bool{
	"go": true, "java": true, "js": true, "php": true,
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		// TODO: return JSON-formatted errors?
		w.WriteHeader(http.StatusBadRequest)
		// TODO: maybe log Fprintf error, here and everywhere else
		fmt.Fprint(w, "error: ", err)
		return
	}

	// Validate fields
	err = u.Validate()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "error: ", err)
		return
	}

	err = db.CreateUser(&u)
	if err != nil {
		if errors.As(err, &ErrConflict{}) {
			w.WriteHeader(http.StatusConflict)
			// FIXME: below message is currently too much of a leap of faith; need to make the whole path more robust
			fmt.Fprint(w, "error: user with the same .email already exists")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "error: ", err)
		return
	}

	// FIXME: base URL below should be customizable via flag
	w.Header().Add("Location", "/v1/user/"+u.Email)
	w.WriteHeader(http.StatusNoContent)
}

func modifyUser(w http.ResponseWriter, r *http.Request) {
	panic("NIY")
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: [LATER] rename 'id' var & param to 'email' for naming consistency
	id := mux.Vars(r)["id"]
	// TODO: quick fail if id empty or invalid?

	err := db.DeleteUser(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "error: ", err)
		return
	}

	// if found == nil {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	fmt.Fprint(w, "error: user not found")
	// 	return
	// }

	w.WriteHeader(http.StatusNoContent)
}
