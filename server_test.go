package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

const validJohnSmith = `
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
`

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
			rq:            validJohnSmith,
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
			comment:    "correct User, with optional .phone present",
			endpoints:  defaultEndpoints,
			rq:         validJohnSmith,
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
				Technology: "java",
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
		gotFilter := make(chan *UserFilter, 1)
		srv := Server{
			DB: callbackDB{
				listUsers: func(filter UserFilter) ([]*User, error) {
					gotFilter <- &filter
					// close chan to force panic if func called more than once
					close(gotFilter)
					return nil, nil
				},
			},
		}
		r := mux.NewRouter()
		srv.RegisterAt(r)
		listener := httptest.NewServer(r)
		client := listener.Client()

		rs, err := client.Get(listener.URL + "/v1/user" + tt.query)
		listener.Close()
		if err != nil {
			t.Errorf("%q: HTTP request error: %s", tt.query, err)
			continue
		}
		if rs.StatusCode != tt.wantStatus {
			t.Errorf("%q: want status %v, got %v (%v)", tt.query, tt.wantStatus, rs.StatusCode, rs.Status)
		}
		if tt.wantFilter != nil {
			select {
			case f := <-gotFilter:
				if !reflect.DeepEqual(f, tt.wantFilter) {
					t.Errorf("%q: bad filter:\nwant: %s\nhave: %s",
						tt.query, dumpJSON(tt.wantFilter), dumpJSON(f))
				}
			default:
				t.Errorf("%q: listUsers not called", tt.query)
			}
		}
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

func dumpJSON(v interface{}) string {
	buf, _ := json.Marshal(v)
	return string(buf)
}

func TestServer_GetUser(t *testing.T) {
	tests := []struct {
		comment    string
		email      string
		mockResult *User
		mockError  error
		wantStatus int
	}{
		{
			comment:    "existing user",
			email:      "rick@black.name",
			mockResult: &User{Email: newString("rick@black.name")},
			wantStatus: http.StatusOK,
		},
		{
			comment:    "no such user",
			email:      "nobody@nowhere.gov",
			mockResult: nil,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		gotEmail := make(chan string, 1)
		srv := Server{
			DB: callbackDB{
				getUser: func(email string) (*User, error) {
					gotEmail <- email
					// close chan to force panic if func called more than once
					close(gotEmail)
					return tt.mockResult, tt.mockError
				},
			},
		}
		r := mux.NewRouter()
		srv.RegisterAt(r)
		listener := httptest.NewServer(r)
		client := listener.Client()

		rs, err := client.Get(listener.URL + "/v1/user/" + tt.email)
		if err != nil {
			t.Errorf("%q: HTTP request error: %s", tt.comment, err)
			listener.Close()
			continue
		}
		// Verify HTTP status of response
		if rs.StatusCode != tt.wantStatus {
			t.Errorf("%q: want status %v, got %v (%v)", tt.comment, tt.wantStatus, rs.StatusCode, rs.Status)
		}
		// Verify body of response
		body, err := ioutil.ReadAll(rs.Body)
		rs.Body.Close()
		listener.Close()
		if err != nil {
			t.Errorf("%q: error reading Body: %s", tt.comment, err)
			continue
		}
		var gotUser *User
		err = json.Unmarshal(body, &gotUser)
		if len(body) > 0 && err != nil {
			t.Errorf("%q: error unmarshalling response as JSON: %s, in:\n%s",
				tt.comment, err, string(body))
		}
		if !reflect.DeepEqual(gotUser, tt.mockResult) {
			t.Errorf("%q: bad response:\nwant: %s\nhave: %s\nraw:  %s",
				tt.comment, dumpJSON(tt.mockResult), dumpJSON(gotUser), string(body))
		}
		// Verify email decoded from URL
		select {
		case e := <-gotEmail:
			if e != tt.email {
				t.Errorf("%q: bad decoded email:\nwant: %s\nhave: %s", tt.comment, tt.email, e)
			}
		default:
			t.Errorf("%q: getUser not called", tt.comment)
		}
	}
}

func newString(v string) *string { return &v }

func TestServer_VariousErrors(t *testing.T) {
	tests := []struct {
		comment    string
		rq         string // "METHOD URL[ BODY]"
		db         callbackDB
		wantStatus int
	}{
		{
			comment: "listUsers error",
			rq:      `GET /v1/user`,
			db: callbackDB{
				listUsers: func(_ UserFilter) ([]*User, error) {
					return nil, errors.New("FAKE ERROR")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			comment: "getUser error",
			rq:      `GET /v1/user/foo@bar.name`,
			db: callbackDB{
				getUser: func(_ string) (*User, error) {
					return nil, errors.New("FAKE ERROR")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			comment: "createUser unspecified error",
			rq:      `POST /v1/user ` + validJohnSmith,
			db: callbackDB{
				createUser: func(_ *User) error {
					return errors.New("FAKE ERROR")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			comment: "createUser Conflict error",
			rq:      `POST /v1/user ` + validJohnSmith,
			db: callbackDB{
				createUser: func(_ *User) error {
					return ErrConflict{wraperr{errors.New("FAKE ERROR")}}
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			comment: "modifyUser unspecified error",
			rq:      `PUT /v1/user/john@smith.com ` + validJohnSmith,
			db: callbackDB{
				modifyUser: func(_ *User) error {
					return errors.New("FAKE ERROR")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			comment: "createUser NotFound error",
			rq:      `PUT /v1/user/john@smith.com ` + validJohnSmith,
			db: callbackDB{
				modifyUser: func(_ *User) error {
					return ErrNotFound{wraperr{errors.New("FAKE ERROR")}}
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			comment: "deleteUser unspecified error",
			rq:      `DELETE /v1/user/john@smith.com`,
			db: callbackDB{
				deleteUser: func(_ string) error {
					return errors.New("FAKE ERROR")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			comment: "deleteUser NotFound error",
			rq:      `DELETE /v1/user/john@smith.com`,
			db: callbackDB{
				deleteUser: func(_ string) error {
					return ErrNotFound{wraperr{errors.New("FAKE ERROR")}}
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		// Prepare server
		srv := Server{DB: tt.db}
		r := mux.NewRouter()
		srv.RegisterAt(r)
		listener := httptest.NewServer(r)
		client := listener.Client()

		// Prepare request
		var (
			query  = strings.SplitN(tt.rq, " ", 3)
			method = query[0]
			path   = query[1]
			body   io.Reader
		)
		if len(query) >= 3 {
			body = strings.NewReader(query[2])
		}
		rq, err := http.NewRequest(method, listener.URL+path, body)
		if err != nil {
			t.Errorf("%q: request building error: %s", tt.comment, err)
			listener.Close()
			continue
		}
		rq.Header.Set("Content-Type", "application/json")

		// RUN TEST
		rs, err := client.Do(rq)
		if err != nil {
			t.Errorf("%q: HTTP query error: %s", tt.comment, err)
			listener.Close()
			continue
		}
		rs.Body.Close()

		// Verify HTTP status of response
		if rs.StatusCode != tt.wantStatus {
			t.Errorf("%q: want status %v, got %v (%v)", tt.comment, tt.wantStatus, rs.StatusCode, rs.Status)
		}
		listener.Close()
	}
}
