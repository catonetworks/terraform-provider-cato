package policy

import (
	"fmt"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// MutationError is the common error projection returned by policy mutations.
type MutationError interface {
	GetErrorCode() *string
	GetErrorMessage() *string
}

// MutationResult rejects missing statuses and failures without error details.
// A successful HTTP/GraphQL response alone does not establish mutation success.
func MutationResult[T MutationError](status *cato_models.PolicyMutationStatus, apiErrors []T) error {
	if len(apiErrors) > 0 {
		messages := make([]string, 0, len(apiErrors))
		for _, entry := range apiErrors {
			code, message := "", ""
			if entry.GetErrorCode() != nil {
				code = *entry.GetErrorCode()
			}
			if entry.GetErrorMessage() != nil {
				message = *entry.GetErrorMessage()
			}
			messages = append(messages, fmt.Sprintf("%s: %s", code, message))
		}
		return fmt.Errorf("policy mutation errors: %s", strings.Join(messages, "; "))
	}
	if status == nil {
		return fmt.Errorf("policy mutation did not return a status")
	}
	if *status != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("policy mutation returned status %q", *status)
	}
	return nil
}

// PublishResult also accepts the explicit no-draft response for a no-op publication.
func PublishResult[T MutationError](status *cato_models.PolicyMutationStatus, apiErrors []T) error {
	if status != nil && *status == cato_models.PolicyMutationStatusFailure && len(apiErrors) == 1 {
		if code := apiErrors[0].GetErrorCode(); code != nil && *code == "PolicyRevisionNotFound" {
			return nil
		}
	}
	return MutationResult(status, apiErrors)
}
