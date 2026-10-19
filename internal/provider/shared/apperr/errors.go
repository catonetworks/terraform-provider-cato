package apperr

import "errors"

type APIErrors interface {
	GetErrorCode() *string
	GetErrorMessage() *string
}

// ErrConfig generic configuration error
var ErrConfig = errors.New("configuration error")

func AssertErr(message string) error {
	return &TestError{message: message}
}

type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}
