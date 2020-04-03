package main

import (
	"errors"
	"net/url"
)

// UserFilter describes criteria for selecting User objects.
type UserFilter struct {
	Technology *string // nil matches any value, non-nil matches equal value in User
	Deleted    *bool   // nil matches any value, true matches deleted User (User.Deleted!=nil), false matches active User (User.Deleted==nil)
}

// NewUserFilter creates a UserFilter based on URL query. If the query cannot
// be translated to a valid UserFilter, an error is returned.
func NewUserFilter(query url.Values) (UserFilter, error) {
	f := UserFilter{}

	switch v := query.Get("technology"); {
	case v == "" || v == "*":
		f.Technology = nil
	case validTechnology[v]:
		f.Technology = &v
	default:
		// TODO: [LATER] avoid duplication of valid technology values in lists
		return UserFilter{}, errors.New("'technology' query parameter must be one of: * go java js php")
	}

	switch v := query.Get("deleted"); v {
	case "*":
		f.Deleted = nil
	case "yes", "true":
		f.Deleted = newBool(true)
	case "", "no", "false":
		f.Deleted = newBool(false)
	default:
		return UserFilter{}, errors.New("'deleted' query parameter must be one of: * yes no true false")
	}

	return f, nil
}

func newBool(v bool) *bool { return &v }
