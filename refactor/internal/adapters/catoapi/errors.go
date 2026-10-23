package catoapi

import (
	"fmt"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
)

var _ network.CatoAPIAdapterError = (*CatoAPIError)(nil)

// CatoAPIError carries API failure details through the application contract.
// Cause remains available through errors.Is/As but is never included in Error().
type CatoAPIError struct {
	Operation string
	Kind      error
	Codes     []string
	Cause     error
}

func (e *CatoAPIError) Error() string        { return fmt.Sprintf("%s: %v", e.Operation, e.Kind) }
func (e *CatoAPIError) Unwrap() error        { return e.Cause }
func (e *CatoAPIError) Is(target error) bool { return target == e.Kind }
func (e *CatoAPIError) ErrorKind() error     { return e.Kind }
func (e *CatoAPIError) ErrorCodes() []string { return e.Codes }

func requestError(operation string, cause error) error {
	return &CatoAPIError{Operation: operation, Kind: network.ErrUnavailable, Cause: cause}
}

func invalidResponse(operation string) error {
	return &CatoAPIError{Operation: operation, Kind: network.ErrInvalidResponse}
}
