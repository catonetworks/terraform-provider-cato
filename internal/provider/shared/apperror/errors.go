package apperror

import "errors"

type APIErrors interface {
	GetErrorCode() *string
	GetErrorMessage() *string
}

var (
	// ErrConfig generic configuration error
	ErrConfig = errors.New("configuration error")

	// ErrAPIResponseParse failed to parse API response
	ErrAPIResponseParse = errors.New("failed to parse API response")
)

// PolicyRevisionNotFound error when publishing policy changes
const PolicyRevisionNotFound = "PolicyRevisionNotFound"

func AssertErr(message string) error {
	return &TestError{message: message}
}

type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}
