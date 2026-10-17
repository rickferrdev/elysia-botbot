package autoreactions

import "errors"

var (
	ErrNotFound = errors.New("autoreaction not found")
	ErrConflict = errors.New("autoreaction conflict")
)
