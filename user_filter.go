package main

import (
	"errors"
	"net/url"
)

type UserFilter struct {
	Technology string
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

	return f, nil
}
