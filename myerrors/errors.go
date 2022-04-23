package myerrors

import "errors"

var (
	ErrBodyNotFound      = errors.New("couldn't get body")
	ErrCtxValue          = errors.New("failed to retrieve value from context")
	ErrInvalidInput      = errors.New("invalid input")
	ErrNegativeCounter   = errors.New("input exceeds counter: counter cannot be negative")
	ErrNonNumericCounter = errors.New("counter is non-numeric")
	ErrNotFound          = errors.New("failed to retrieve data")
	ErrUserNotFound      = errors.New("user not found")
)
