package network

import (
	"errors"
	"fmt"
)

var (
	ErrRuleNotFound     = errors.New("rule is absent from the editable revision")
	ErrNoDraft          = errors.New("no draft revision exists")
	ErrRevisionConflict = errors.New("policy revision is busy")
	ErrInvalidResponse  = errors.New("API response is incomplete or invalid")
	ErrRejected         = errors.New("API rejected the operation")
	ErrUnavailable      = errors.New("API request did not return a usable response")
	ErrStillPublished   = errors.New("rule is still present in the published policy")
)

// CatoAPIAdapterError exposes only the information use cases need to classify
// an API failure. ErrorKind returns ErrRejected, ErrUnavailable, or ErrInvalidResponse.
type CatoAPIAdapterError interface {
	error
	ErrorKind() error
	ErrorCodes() []string
}

// UseCaseError records the application classification while preserving the
// original cause for errors.Is/As. Error never renders the adapter's raw message.
type UseCaseError struct {
	Kind  error
	Cause error
}

func (e *UseCaseError) Error() string        { return fmt.Sprintf("use case failed: %v", e.Kind) }
func (e *UseCaseError) Unwrap() error        { return e.Cause }
func (e *UseCaseError) Is(target error) bool { return target == e.Kind }

type Stage string

const (
	StageCoordinate Stage = "coordinate"
	StageRemove     Stage = "remove"
	StagePublish    Stage = "publish"
	StageVerify     Stage = "verify"
)

// DeleteError records progress without introducing Terraform state or diagnostics.
type DeleteError struct {
	Stage            Stage
	RemovalCompleted bool
	Cause            error
}

func (e *DeleteError) Error() string {
	return fmt.Sprintf("delete private-access rule at %s: %v", e.Stage, e.Cause)
}
func (e *DeleteError) Unwrap() error { return e.Cause }

type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return e.Field + " must be set" }

// classifyPrivateAccessPolicyError assigns workflow meaning to API error codes.
// The adapter reports codes without deciding whether to tolerate or retry them.
func classifyPrivateAccessPolicyError(err error) error {
	if err == nil {
		return nil
	}
	classified := classifyCatoAPIError(err)
	var apiError CatoAPIAdapterError
	if !errors.As(err, &apiError) || classified.Kind != ErrRejected {
		return classified
	}
	codes := apiError.ErrorCodes()
	if len(codes) == 0 {
		return classified
	}
	var kind error
	for _, code := range codes {
		var current error
		switch code {
		case "ruleNotExist", "RuleNotFound":
			current = ErrRuleNotFound
		case "PolicyRevisionNotFound":
			current = ErrNoDraft
		case "reorderPolicyBlockedByActiveSessions":
			current = ErrRevisionConflict
		default:
			current = ErrRejected
		}
		if kind != nil && kind != current {
			return classified // Mixed failures cannot be hidden by an acceptable code.
		}
		kind = current
	}
	classified.Kind = kind
	return classified
}

// Callers handle nil before converting the result to error, avoiding typed nils.
func classifyCatoAPIError(err error) *UseCaseError {
	if err == nil {
		return nil
	}
	kind := ErrUnavailable
	var apiError CatoAPIAdapterError
	if errors.As(err, &apiError) {
		switch apiKind := apiError.ErrorKind(); apiKind {
		case ErrRejected, ErrUnavailable, ErrInvalidResponse:
			kind = apiKind
		}
	}
	return &UseCaseError{Kind: kind, Cause: err}
}
