package tomatoes

import "errors"

var (
	InternalError            = errors.New("internal error")
	InsufficientBalanceError = errors.New("insufficient balance")
)
