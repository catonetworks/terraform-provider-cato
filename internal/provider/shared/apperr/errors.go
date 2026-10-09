package apperr

import "errors"

type APIErrors interface {
	GetErrorCode() *string
	GetErrorMessage() *string
}

// ErrConfig generic configuration error
var ErrConfig = errors.New("configuration error")
