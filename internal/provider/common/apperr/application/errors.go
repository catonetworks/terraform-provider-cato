// Package application defines errors that do not depend on Terraform or the SDK.
package application

import "errors"

// Error retains the diagnostic context of a failed operation and its cause.
type Error struct {
	Summary string
	Detail  string
	Cause   error
}

func (e *Error) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return e.Cause.Error()
}
func (e *Error) Unwrap() error { return e.Cause }
func Wrap(summary string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Summary: summary, Cause: err}
}

var ErrAbsent = errors.New("remote resource absent")

// New constructs an operation error with a diagnostic detail that has no underlying cause.
func New(summary, detail string) error { return Wrap(summary, errors.New(detail)) }
