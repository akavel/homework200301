package main

type ErrConflict struct{ wraperr }
type ErrNotFound struct{ wraperr }

// wraperr is a helper type, allowing to easily wrap errors in "tagged" types.
type wraperr struct{ err error }

func (e wraperr) Error() string { return e.err.Error() }
func (e wraperr) Unwrap() error { return e.err }
