package domain

import (
	"errors"
)

var (
	// ErrNotFound indicates that the queried module is not found.
	ErrNotFound = errors.New("not found")
)
