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
	"github.com/jamiealquiza/envy" // TODO: fork it to avoid cobra & pflag dependencies, contribute back upstream
)

var (
	addr   = flag.String("http", ":8080", "http address to listen on")
	rqlog  = flag.String("rqlog", "requests.log", "file to log requests information into")
	dbconn = flag.String("dbconn", "postgres://login:password@localhost:5432/users_db", "Postgres database connection string")
)

func main() {
	// TODO: write tests, run with `go test -race`

	envy.Parse("USERS") // Propagate env variables into flags
	flag.Parse()

	// Connect to Postgres DB
	dbopt, err := pg.ParseURL(*dbconn)
	if err != nil {
		log.Fatalf("parsing -dbconn flag value: %s", err)
	}
	dbopt.ApplicationName = "users_go"
	// TODO: [LATER] add timeouts etc. to dbopt
	db, err := ConnectPostgres(dbopt)
	if err != nil {
		log.Fatalf("initializing Postgres DB: %s", err)
	}
	// TODO: [LATER] Close() will never happen now (needs HTTP server soft shutdown)
	// TODO: [LATER] log any error from Close()
	defer db.Close()

	rqLogger, err := NewRequestLogger(*rqlog)
	if err != nil {
		log.Fatal(err)
	}

	srv := Server{
		DB: db,
		// FIXME: base URL below should be customizable via a separate flag
		BaseURL: *addr,
	}

	r := mux.NewRouter()
	r.Use(rqLogger.WrapHTTPHandler)
	srv.RegisterAt(r)
	log.Fatal(http.ListenAndServe(*addr, r))
}

type Server struct {
	DB      Database
	BaseURL string
}

// Database represents a set of operations required of a database to be usable
// by the Server. The User object passed to the functions is assumed to be
// internally consistent. However, a type implementing Database is responsible
// for ensuring database consistency, including:
//
// - there can be only one active (non-deleted) user with same Email value
//
// - all operations on a Database must be safe for concurrent use
//
// When it makes sense, the operations are expected to return an error that can
// be converted to ErrNotFound or ErrConflict using errors.As.
//
// TODO: [LATER] introduce Context to methods, to allow timeouts control
type Database interface {
	// ListUsers is expected to return a list of users matching the provided
	// filter.
	ListUsers(filter UserFilter) ([]*User, error)
	GetUser(email string) (*User, error)
	CreateUser(u *User) error
	ModifyUser(u *User) error
	// DeleteUser is expected to be a "soft delete", setting User.Deleted to
	// non-nil value.
	DeleteUser(email string) error

	Close() error
}

func (s *Server) RegisterAt(r *mux.Router) {
	r.Methods("GET").Path("/v1/user").HandlerFunc(s.listUsers)
	r.Methods("GET").Path("/v1/user/{email}").HandlerFunc(s.getUser)
	r.Methods("POST").Path("/v1/user").HandlerFunc(s.createUser)
	r.Methods("PUT").Path("/v1/user/{email}").HandlerFunc(s.modifyUser)
	r.Methods("DELETE").Path("/v1/user/{email}").HandlerFunc(s.deleteUser)
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query into filters
	// TODO: pagination - ideally automatically mapped to the Postgres query & to the response (UsersList type? HTTP headers?)
	filter, err := NewUserFilter(r.URL.Query())
	if err != nil {
		RespondError(w, http.StatusBadRequest, err)
		return
	}

	users, err := s.DB.ListUsers(filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err)
		return
	}
	// Let's print `[]` instead of `null` in the JSON response in case of empty results list
	if users == nil {
		users = []*User{}
	}
	RespondJSON(w, http.StatusOK, users)
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	// TODO: quick fail if email empty or invalid?

	found, err := s.DB.GetUser(email)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err)
		return
	}
	if found != nil {
		RespondJSON(w, http.StatusOK, found)
	} else {
		RespondJSON(w, http.StatusNotFound, nil)
	}
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err)
		return
	}

	// Validate fields
	err = u.Validate()
	if err != nil {
		RespondError(w, http.StatusBadRequest, err)
		return
	}

	err = s.DB.CreateUser(&u)
	if err != nil {
		if errors.As(err, &ErrConflict{}) {
			// FIXME: below message is currently too much of a leap of faith; need to make the whole path more robust
			RespondError(w, http.StatusConflict, errors.New("user with the same .email already exists"))
			return
		}
		RespondError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Add("Location", s.BaseURL+"/v1/user/"+*u.Email)
	RespondJSON(w, http.StatusNoContent, nil)
}

func (s *Server) modifyUser(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	// TODO: quick fail if email empty or invalid?

	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err)
		return
	}

	// Validate fields
	err = u.Validate()
	if err != nil {
		RespondError(w, http.StatusBadRequest, err)
		return
	}
	if *u.Email != email {
		RespondError(w, http.StatusBadRequest, errors.New(".email field does not match the value in the URL"))
		return
	}

	// TODO: [LATER] consider adding data versioning to User to let clients avoid race conditions

	err = s.DB.ModifyUser(&u)
	if err != nil {
		// TODO: [LATER] below block is ugly, find a nicer way of translating errors (helper func?)
		if errors.As(err, &ErrNotFound{}) {
			RespondError(w, http.StatusNotFound, err)
			return
		}
		RespondError(w, http.StatusInternalServerError, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	// TODO: quick fail if email empty or invalid?

	err := s.DB.DeleteUser(email)
	if err != nil {
		if errors.As(err, &ErrNotFound{}) {
			RespondError(w, http.StatusNotFound, err)
			return
		}
		RespondError(w, http.StatusInternalServerError, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

// RespondError writes the error message from err into w, and sets the HTTP
// status of the response.
//
// TODO: emit JSON-formatted errors? (via RespondJSON)
func RespondError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	// TODO: [LATER] maybe return/log error if Fprint failed
	fmt.Fprint(w, "error: ", err)
}

// RespondJSON marshals non-nil obj into w as JSON, and sets the HTTP status of
// the response. If obj is a literal nil, only the response status and headers
// are set (no data is serialized).
func RespondJSON(w http.ResponseWriter, status int, obj interface{}) {
	// TODO: should we also set Content-Type if not printing any contents (obj==nil)?
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if obj == nil {
		return
	}
	buf, err := json.Marshal(obj)
	if err != nil {
		log.Printf("BUG: RespondJSON called with an object of type %T that doesn't serialize to JSON: %s", obj, err)
		return
	}
	// Note: ignoring write errors, as we don't want info every time client
	// decided to ignore us and disconnect
	w.Write(buf)
}
