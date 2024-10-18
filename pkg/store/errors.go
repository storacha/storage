package store

import "errors"

// ErrNotFound is returned when something is not found in the store.
var ErrNotFound = errors.New("not found")
