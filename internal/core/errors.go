package core

import (
	"errors"
	"fmt"
)

type (
	ErrorCode    string
	ErrorMessage string
	ErrorStatus  string
)

type AppError struct {
	Message ErrorMessage
	Code    ErrorCode
	Status  ErrorStatus
	Cause   error
}

func (err *AppError) Error() string {
	if err.Cause != nil {
		return fmt.Sprintf("[%s]: %s: %s\n", err.Code, err.Message, err.Cause.Error())
	}
	return fmt.Sprintf("[%s]: %s", err.Code, err.Message)
}

func (err *AppError) Unwrap() error {
	return err.Cause
}

func (err *AppError) Is(target error) bool {
	targetError, ok := target.(*AppError)
	return ok && err.Code == targetError.Code
}

func IsCode(err error, code ErrorCode) bool {
	var e *AppError
	if !errors.As(err, &e) {
		return false
	}

	return e.Code == code
}

func NewAppError(message ErrorMessage, code ErrorCode, cause error) *AppError {
	return &AppError{
		Message: message,
		Cause:   cause,
		Code:    code,
	}
}
