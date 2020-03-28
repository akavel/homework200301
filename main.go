package main

import (
	"flag"
	"log"
	"net/http"
	"time"

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

type User struct {
	Name    string
	Surname string
	Email   string
	// FIXME: only store a hash of the password
	Password   string
	Birthday   time.Time
	Address    string
	Phone      *string
	Technology string
	Deleted    *time.Time
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

var mockUsers = []User{
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
