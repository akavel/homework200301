package main

import (
	"errors"
	"strings"
	"time"
)

var validTechnology = map[string]bool{
	"go": true, "java": true, "js": true, "php": true,
}

// FIXME: make Technology an enum or FK in Postgres
// TODO: [LATER] pg-related annotations & fields are internal and should not be exposed in exported type (see orm.Table?)
type User struct {
	// ID is the field required by go-pg ORM as the primary key
	ID int64 `json:"-"`

	Name    string `json:"name" pg:",notnull"`
	Surname string `json:"surname" pg:",notnull"`
	Email   string `json:"email" pg:",notnull"`
	// FIXME: [LATER] only store a hash of the password
	Password   string     `json:"password" pg:",notnull"`
	Birthday   time.Time  `json:"birthday" pg:",notnull"`
	Address    string     `json:"address" pg:",notnull"`
	Phone      *string    `json:"phone",omitempty`
	Technology string     `json:"technology" pg:",notnull"`
	Deleted    *time.Time `json:"deleted,omitempty"`
}

// Validate checks if User fields have allowed values. If not, an error is
// returned with a detailed message.
func (u *User) Validate() error {
	// TODO: return all validation errors, not just the first one
	switch {
	case !strings.Contains(u.Email, "@"):
		// TODO: also verify there's something before '@' and after '@'
		// TODO: consider more advanced validation, though this is tricky; if
		// applicable, consider sending confirmation email instead
		return errors.New(".email is not a valid email address")
	case !validTechnology[u.Technology]:
		return errors.New(".technology must be one of: go java js php")
	case u.Deleted != nil:
		return errors.New(".deleted must be empty")
	default:
		return nil
	}
	// TODO: .birthday probably shouldn't be in future
	// TODO: validate .phone contents format if field provided (there's some pkg for this IIRC)
	// TODO: arguably, non-optional fields should also be non-empty, though
	// question is how far we want to go with validation, e.g. is "x" a
	// valid address?
}
