package network

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// This fake implements the application contract without importing the adapter.
type catoAPIErrorStub struct {
	kind  error
	codes []string
	cause error
}

var _ CatoAPIAdapterError = catoAPIErrorStub{}

func (e catoAPIErrorStub) Error() string        { return "sensitive backend details" }
func (e catoAPIErrorStub) ErrorKind() error     { return e.kind }
func (e catoAPIErrorStub) ErrorCodes() []string { return e.codes }
func (e catoAPIErrorStub) Unwrap() error        { return e.cause }

func rejectedCatoAPIError(codes ...string) error {
	return catoAPIErrorStub{kind: ErrRejected, codes: codes}
}

func TestClassifyPrivateAccessPolicyError(t *testing.T) {
	t.Parallel()
	require.NoError(t, classifyPrivateAccessPolicyError(nil))
	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"absent rule", rejectedCatoAPIError("ruleNotExist"), ErrRuleNotFound},
		{"alternate absent rule", rejectedCatoAPIError("RuleNotFound"), ErrRuleNotFound},
		{"no draft", rejectedCatoAPIError("PolicyRevisionNotFound"), ErrNoDraft},
		{"conflict", rejectedCatoAPIError("reorderPolicyBlockedByActiveSessions"), ErrRevisionConflict},
		{"wrapped conflict", fmt.Errorf("sensitive wrapper: %w", rejectedCatoAPIError("reorderPolicyBlockedByActiveSessions")), ErrRevisionConflict},
		{"mixed failures", rejectedCatoAPIError("RuleNotFound", "PermissionDenied"), ErrRejected},
		{"different recognized failures", rejectedCatoAPIError("RuleNotFound", "PolicyRevisionNotFound"), ErrRejected},
		{"same classification", rejectedCatoAPIError("ruleNotExist", "RuleNotFound"), ErrRuleNotFound},
		{"unknown code", rejectedCatoAPIError("PermissionDenied"), ErrRejected},
		{"no codes", rejectedCatoAPIError(), ErrRejected},
		{"invalid response", catoAPIErrorStub{kind: ErrInvalidResponse, codes: []string{"RuleNotFound"}}, ErrInvalidResponse},
		{"unavailable", catoAPIErrorStub{kind: ErrUnavailable}, ErrUnavailable},
		{"unknown error", errors.New("sensitive backend details"), ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := classifyPrivateAccessPolicyError(tc.err)
			var failure *UseCaseError
			require.ErrorAs(t, err, &failure)
			require.Equal(t, tc.want, failure.Kind)
			require.ErrorIs(t, err, tc.want)
			require.Equal(t, tc.err, failure.Cause)
			require.NotContains(t, err.Error(), "sensitive")
		})
	}
}

func TestUseCaseErrorPreservesCancellationAndAPIContract(t *testing.T) {
	t.Parallel()
	source := catoAPIErrorStub{kind: ErrUnavailable, cause: context.Canceled}
	err := classifyPrivateAccessPolicyError(source)
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, err, ErrUnavailable)
	var apiError CatoAPIAdapterError
	require.ErrorAs(t, err, &apiError)
	require.Equal(t, source, apiError)
	require.NotContains(t, err.Error(), "sensitive")
}
