package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestServer_PostAndPutUser_Validation(t *testing.T) {
	defaultEndpoints := []string{
		"POST /v1/user",
		"PUT /v1/user/john@smith.com",
	}
	tests := []struct {
		comment       string
		endpoints     []string
		rq            string
		wantStatus    int
		wantReplyWith string
	}{
		// INVALID REQUESTS
		// Non-JSON
		{
			comment:    "non-JSON input",
			endpoints:  defaultEndpoints,
			rq:         `hello world!`,
			wantStatus: http.StatusBadRequest,
		},
		// Incorrect field values
		{
			comment:   "invalid User: invalid email (no @)",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john.smith.com",
"name": "John",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "go"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "email",
		},
		{
			comment:   "invalid User: invalid value in .technology",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"name": "John",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "haskell"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "technology",
		},
		{
			comment:   "invalid User: unwanted 'deleted' field",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"name": "John",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "go",
"deleted": "1950-01-01T00:00:00Z"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "deleted",
		},
		{
			comment: "mismatch between .email and URL",
			endpoints: []string{
				"PUT /v1/user/greg@example.com",
			},
			rq: `
{
"email": "john@smith.com",
"name": "John",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "go"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "email",
		},
		// Missing fields
		{
			comment:   "invalid User: no email",
			endpoints: defaultEndpoints,
			rq: `
{
"name": "John",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "go"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "email",
		},
		{
			comment:   "invalid User: no name",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "go"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "name",
		},
		{
			comment:   "invalid User: no surname",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"name": "John",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "go"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "surname",
		},
		{
			comment:   "invalid User: no password",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"name": "John",
"surname": "Smith",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "go"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "password",
		},
		{
			comment:   "invalid User: no birthday",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"name": "John",
"surname": "Smith",
"password": "some pwd",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "go"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "birthday",
		},
		{
			comment:   "invalid User: no address",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"name": "John",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"phone": "111 222 333",
"technology": "go"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "address",
		},
		{
			comment:   "invalid User: no technology",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"name": "John",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333"
}
`,
			wantStatus:    http.StatusBadRequest,
			wantReplyWith: "technology",
		},

		// VALID REQUESTS
		{
			comment:   "correct User, with optional .phone present",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"name": "John",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"phone": "111 222 333",
"technology": "go"
}
`,
			wantStatus: http.StatusNoContent,
		},
		{
			comment:   "correct User, with optional .phone absent",
			endpoints: defaultEndpoints,
			rq: `
{
"email": "john@smith.com",
"name": "John",
"surname": "Smith",
"password": "some pwd",
"birthday": "1950-01-01T00:00:00Z",
"address": "Some Street 17\nSome City",
"technology": "go"
}
`,
			wantStatus: http.StatusNoContent,
		},
	}

	srv := Server{DB: nullDB{}}
	r := mux.NewRouter()
	srv.RegisterAt(r)
	listener := httptest.NewServer(r)
	client := listener.Client()
	defer listener.Close()

	for _, tt := range tests {
		for _, e := range tt.endpoints {
			var (
				s      = strings.Split(e, " ")
				method = s[0]
				path   = s[1]
			)

			req, err := http.NewRequest(method, listener.URL+path, strings.NewReader(tt.rq))
			if err != nil {
				t.Errorf("%s %q: request building error: %s", e, tt.comment, err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			rs, err := client.Do(req)
			if err != nil {
				t.Errorf("%s %q: HTTP query error: %s", e, tt.comment, err)
				continue
			}
			if rs.StatusCode != tt.wantStatus {
				t.Errorf("%s %q: want status %v, got %v (%v)", e, tt.comment, tt.wantStatus, rs.StatusCode, rs.Status)
			}
			body, err := ioutil.ReadAll(rs.Body)
			rs.Body.Close()
			if err != nil {
				t.Errorf("%s %q: error reading Body: %s", e, tt.comment, err)
				continue
			}
			if !strings.Contains(string(body), tt.wantReplyWith) {
				t.Errorf("%s %q: want reply with: %q, got:\n%s", e, tt.comment, tt.wantReplyWith, string(body))
			}
		}
	}
}

type nullDB struct{}

func (db nullDB) ListUsers(filter UserFilter) ([]*User, error) { return nil, nil }
func (db nullDB) GetUser(email string) (*User, error)          { return nil, nil }
func (db nullDB) CreateUser(u *User) error                     { return nil }
func (db nullDB) ModifyUser(u *User) error                     { return nil }
func (db nullDB) DeleteUser(email string) error                { return nil }
func (db nullDB) Close() error                                 { return nil }

func TestServer_ListUsers(t *testing.T) {
	tests := []struct {
		query      string
		wantFilter *UserFilter
		wantStatus int
	}{
		{
			query: "", // default filter
			wantFilter: &UserFilter{
				Technology: "*",
				Deleted:    newBool(false),
			},
			wantStatus: http.StatusOK,
		},
		{
			query: "?technology=go",
			wantFilter: &UserFilter{
				Technology: "go",
				Deleted:    newBool(false),
			},
			wantStatus: http.StatusOK,
		},
		{
			query:      "?technology=FOOBAR",
			wantStatus: http.StatusBadRequest,
		},
		{
			query: "?deleted=*",
			wantFilter: &UserFilter{
				// TODO: make semantics of UserFilter fields more consistent, e.g. Technology:nil instead of "*"
				Technology: "*",
				Deleted:    nil,
			},
			wantStatus: http.StatusOK,
		},
		{
			query: "?deleted=true",
			wantFilter: &UserFilter{
				Technology: "*",
				Deleted:    newBool(true),
			},
			wantStatus: http.StatusOK,
		},
		{
			query: "?technology=java&deleted=*",
			wantFilter: &UserFilter{
				Technology: "java",
				Deleted:    nil,
			},
			wantStatus: http.StatusOK,
		},
		{
			query: "?technology=java&deleted=yes",
			wantFilter: &UserFilter{
				Technology: "*",
				Deleted:    newBool(true),
			},
			wantStatus: http.StatusOK,
		},
		{
			query:      "?deleted=FOOBAR",
			wantStatus: http.StatusBadRequest,
		},
		{
			query:      "?technology=FOOBAR&deleted=yes",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
	}
}

type callbackDB struct {
	listUsers  func(filter UserFilter) ([]*User, error)
	getUser    func(email string) (*User, error)
	createUser func(u *User) error
	modifyUser func(u *User) error
	deleteUser func(email string) error
	close      func() error
}

func (db callbackDB) ListUsers(filter UserFilter) ([]*User, error) { return db.listUsers(filter) }
func (db callbackDB) GetUser(email string) (*User, error)          { return db.getUser(email) }
func (db callbackDB) CreateUser(u *User) error                     { return db.createUser(u) }
func (db callbackDB) ModifyUser(u *User) error                     { return db.modifyUser(u) }
func (db callbackDB) DeleteUser(email string) error                { return db.deleteUser(email) }
func (db callbackDB) Close() error                                 { return db.close() }
