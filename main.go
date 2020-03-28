package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	addr = flag.String("http", ":8080", "http address to listen on")
)

func main() {
	// TODO: tests
	r := mux.NewRouter()
	r.Methods("GET").Path("/v1/user").HandlerFunc(listUsers)
	r.Methods("GET").Path("/v1/user/{id}").HandlerFunc(getUser)
	r.Methods("POST").Path("/v1/user").HandlerFunc(createUser)
	r.Methods("PUT").Path("/v1/user/{id}").HandlerFunc(modifyUser)
	log.Fatal(http.ListenAndServe(*addr, r))
}

type User struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Email   string `json:"email"`
	// FIXME: [LATER] only store a hash of the password
	Password   string     `json:"password"`
	Birthday   time.Time  `json:"birthday"`
	Address    string     `json:"address"`
	Phone      *string    `json:"phone",omitempty`
	Technology string     `json:"technology"`
	Deleted    *time.Time `json:"deleted,omitempty"`
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: query: technology=*|php|go|java|js,active=true|false|*
	// TODO: query: pagination - ideally automatically mapped to the Postgres query & to the response (UsersList type? HTTP headers?)
	// TODO: Content-Type, Accepted

	mockLock.Lock()
	defer mockLock.Unlock()

	err := json.NewEncoder(w).Encode(mockUsers)
	if err != nil {
		log.Print("listUsers: ", err)
		// TODO: if not too late, write 500 to w
		return
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	// TODO: quick fail if id empty or invalid?

	mockLock.Lock()
	defer mockLock.Unlock()

	found := mockUsers.findActive(id)
	// TODO: exit mutex early, serialization doesn't need to be in critical section

	if found == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := json.NewEncoder(w).Encode(found)
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
	// TODO: return all validation errors, not just the first one
	switch {
	case !strings.Contains(u.Email, "@"):
		// TODO: also verify there's something before '@' and after '@'
		// TODO: consider more advanced validation, though this is tricky; if applicable, consider sending confirmation email
		err = errors.New(".email is not a valid email address")
	case !validTechnology[u.Technology]:
		err = errors.New(".technology must be one of: go java js php")
	case u.Deleted != nil:
		err = errors.New(".deleted must be empty")
	}
	// TODO: .birthday probably shouldn't be in future
	// TODO: validate .phone if provided (there's some pkg for this IIRC)
	// TODO: arguably, non-optional fields should also be non-empty, though
	// question is how far we want to go with validation, e.g. is "x" a
	// valid address?
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "error: ", err)
		return
	}

	mockLock.Lock()
	defer mockLock.Unlock()

	// Make sure user with the same email doesn't already exist
	conflict := mockUsers.findActive(u.Email)
	if conflict != nil {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, "error: user with the same .email already exists")
		return
	}

	mockUsers = append(mockUsers, u)

	// FIXME: base URL below should be customizable via flag
	w.Header().Add("Location", "/v1/user/"+u.Email)
	w.WriteHeader(http.StatusNoContent)
}

func modifyUser(w http.ResponseWriter, r *http.Request) {
	panic("NIY")
}

var mockLock sync.Mutex
var mockUsers = Users{
	{
		Name: "John", Surname: "Smith",
		Email:    "john@smith.name",
		Password: "iAmJohnny",
		// FIXME: provide location
		Birthday:   time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC),
		Address:    "some city, some street 1",
		Phone:      nil,
		Technology: "php",
		Deleted:    nil,
	},
	{
		Name: "Anne J. Marie", Surname: "von Flick",
		Email:    "Anne-Marie@vonflick.de",
		Password: `foo123!"%`,
		// FIXME: provide location
		Birthday:   time.Date(2000, 12, 12, 0, 0, 0, 0, time.UTC),
		Address:    "Anneville",
		Phone:      newString("+48 123-456-789"),
		Technology: "java",
		Deleted:    nil,
	},
	{
		Name:     "Robert'); DROP TABLE Users;--",
		Surname:  "Tables",                // TODO: fun surname (Unicode? Zalgo?)
		Email:    "bobby.tables@xkcd.com", // TODO: fun email
		Password: "TODO: harder",          // TODO: fun password
		// FIXME: provide location
		Birthday:   time.Date(2007, 10, 7, 0, 0, 0, 0, time.UTC), // TODO: fun date (Feb?)
		Address:    "Wherever",                                   // TODO: fun address (Unicode? Zalgo?)
		Phone:      nil,                                          // TODO: fun phone
		Technology: "js",
		Deleted:    nil,
	},
	{
		Name: "Dorothy", Surname: "Deleted de 1",
		Email:    "bobby.tables@xkcd.com",
		Password: "",
		// FIXME: provide location
		Birthday:   time.Date(2000, 12, 12, 0, 0, 0, 0, time.UTC),
		Address:    "Dorothyville 1",
		Phone:      newString("+48 1"),
		Technology: "go",
		// FIXME: provide location?
		Deleted: newTime(time.Date(2020, 3, 28, 12, 21, 1, 0, time.UTC)),
	},
	{
		Name: "Dorothy", Surname: "Deleted de 2",
		Email:    "bobby.tables@xkcd.com",
		Password: "",
		// FIXME: provide location
		Birthday:   time.Date(2000, 12, 12, 0, 0, 0, 0, time.UTC),
		Address:    "Dorothyville 2",
		Phone:      newString("+48 2"),
		Technology: "go",
		// FIXME: provide location?
		Deleted: newTime(time.Date(2020, 3, 28, 12, 21, 2, 0, time.UTC)),
	},
}

func newString(v string) *string     { return &v }
func newTime(v time.Time) *time.Time { return &v }

type Users []User

func (us Users) findActive(id string) *User {
	for _, u := range us {
		if u.Deleted == nil && u.Email == id {
			return &u
		}
	}
	return nil
}
