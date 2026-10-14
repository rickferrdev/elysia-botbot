package server

import "errors"

var (
	ErrNotFound = errors.New("server configuration not found")
	ErrConflict = errors.New("server configuration conflict")
)
