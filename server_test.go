package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestServer_PostUser_Validation(t *testing.T) {
	tests := []struct {
		comment       string
		rq            string
		wantStatus    int
		wantReplyWith string
	}{
		// INVALID REQUESTS
		{
			comment:    "non-JSON input",
			rq:         `hello world!`,
			wantStatus: http.StatusBadRequest,
		},
		{
			comment: "invalid User: no email",
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
			comment: "invalid User: invalid email (no @)",
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
			comment: "invalid User: no name",
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
			comment: "invalid User: no surname",
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
			comment: "invalid User: no password",
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
			comment: "invalid User: no birthday",
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
			wantReplyWith: "birthday",
		},
		{
			comment: "invalid User: no address",
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
			comment: "invalid User: no technology",
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
			comment: "correct User, with optional .phone present",
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
			comment: "correct User, with optional .phone absent",
			rq: `
{
"email": "greg@smith.com",
"name": "Greg",
"surname": "Smith",
"password": "some pwd",
"birthday": "1966-01-01T00:00:00Z",
"address": "Tiny Town 23",
"technology": "js"
}
`,
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		srv := Server{DB: nullDB{}}
		r := mux.NewRouter()
		srv.RegisterAt(r)
		listener := httptest.NewServer(r)
		client := listener.Client()

		rs, err := client.Post(listener.URL+"/v1/user", "application/json", strings.NewReader(tt.rq))
		if err != nil {
			t.Errorf("%q: HTTP query error: %s", tt.comment, err)
			listener.Close()
			continue
		}
		if rs.StatusCode != tt.wantStatus {
			t.Errorf("%q: want status %v, got %v (%v)", tt.comment, tt.wantStatus, rs.StatusCode, rs.Status)
		}
		body, err := ioutil.ReadAll(rs.Body)
		rs.Body.Close()
		if err != nil {
			t.Errorf("%q: error reading Body: %s", tt.comment, err)
			listener.Close()
			continue
		}
		if !strings.Contains(string(body), tt.wantReplyWith) {
			t.Errorf("%q: want reply with: %q, got:\n%s", tt.comment, tt.wantReplyWith, string(body))
		}

		listener.Close()
	}
}

type nullDB struct{}

func (db nullDB) ListUsers(filter UserFilter) ([]*User, error) { return nil, nil }
func (db nullDB) GetUser(email string) (*User, error)          { return nil, nil }
func (db nullDB) CreateUser(u *User) error                     { return nil }
func (db nullDB) ModifyUser(u *User) error                     { return nil }
func (db nullDB) DeleteUser(email string) error                { return nil }
func (db nullDB) Close() error                                 { return nil }