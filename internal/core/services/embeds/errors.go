package embeds

import "errors"

var (
	ErrNotFound     = errors.New("embed not found")
	ErrInvalidInput = errors.New("invalid embed input")
	ErrConflict     = errors.New("autoreaction conflict")
	ErrMaxSize      = errors.New("file has reached maximum size")
	ErrBadRequest   = errors.New("request for cdn with errors")
)
