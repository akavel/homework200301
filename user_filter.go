package main

import (
	"errors"
	"net/url"
)

type UserFilter struct {
	Technology string
	Active     *bool
}

func NewUserFilter(query url.Values) (UserFilter, error) {
	f := UserFilter{}

	v := query.Get("technology")
	switch {
	case v == "" || v == "*":
		f.Technology = "*"
	case validTechnology[v]:
		f.Technology = v
	default:
		// TODO: [LATER] avoid duplication of valid technology values in lists
		return UserFilter{}, errors.New("'technology' query parameter must be one of: * go java js php")
	}

	v = query.Get("active")
	switch v {
	case "*":
		f.Active = nil
	case "", "yes", "true":
		f.Active = newBool(true)
	case "no", "false":
		f.Active = newBool(false)
	default:
		return UserFilter{}, errors.New("'active' query parameter must be one of: * yes no true false")
	}

	return f, nil
}

func newBool(v bool) *bool { return &v }
