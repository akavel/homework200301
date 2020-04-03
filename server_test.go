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
	tests := []struct {
		comment       string
		rq            string
		wantStatus    int
		wantReplyWith string
	}{
		// INVALID REQUESTS
		// Non-JSON
		{
			comment:    "non-JSON input",
			rq:         `hello world!`,
			wantStatus: http.StatusBadRequest,
		},
		// Incorrect field values
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
			comment: "invalid User: invalid value in .technology",
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
			comment: "invalid User: unwanted 'deleted' field",
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
		// Missing fields
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

	for _, tt := range tests {
		srv := Server{DB: nullDB{}}
		r := mux.NewRouter()
		srv.RegisterAt(r)
		listener := httptest.NewServer(r)
		client := listener.Client()

		// Test POST
		rs, err := client.Post(listener.URL+"/v1/user", "application/json", strings.NewReader(tt.rq))
		if err != nil {
			t.Errorf("POST %q: HTTP query error: %s", tt.comment, err)
			listener.Close()
			continue
		}
		if rs.StatusCode != tt.wantStatus {
			t.Errorf("POST %q: want status %v, got %v (%v)", tt.comment, tt.wantStatus, rs.StatusCode, rs.Status)
		}
		body, err := ioutil.ReadAll(rs.Body)
		rs.Body.Close()
		if err != nil {
			t.Errorf("POST %q: error reading Body: %s", tt.comment, err)
			listener.Close()
			continue
		}
		if !strings.Contains(string(body), tt.wantReplyWith) {
			t.Errorf("POST %q: want reply with: %q, got:\n%s", tt.comment, tt.wantReplyWith, string(body))
		}

		// Test PUT
		req, err := http.NewRequest("PUT", listener.URL+"/v1/user/john@smith.com", strings.NewReader(tt.rq))
		if err != nil {
			t.Errorf("PUT %q: request building error: %s", tt.comment, err)
			listener.Close()
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		rs, err = client.Do(req)
		if err != nil {
			t.Errorf("PUT %q: HTTP query error: %s", tt.comment, err)
			listener.Close()
			continue
		}
		if rs.StatusCode != tt.wantStatus {
			t.Errorf("PUT %q: want status %v, got %v (%v)", tt.comment, tt.wantStatus, rs.StatusCode, rs.Status)
		}
		body, err = ioutil.ReadAll(rs.Body)
		rs.Body.Close()
		if err != nil {
			t.Errorf("PUT %q: error reading Body: %s", tt.comment, err)
			listener.Close()
			continue
		}
		if !strings.Contains(string(body), tt.wantReplyWith) {
			t.Errorf("PUT %q: want reply with: %q, got:\n%s", tt.comment, tt.wantReplyWith, string(body))
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
