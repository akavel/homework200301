package main

import (
	"time"
)

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
