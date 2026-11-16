package policy

import (
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/require"
)

type testMutationError struct{ code, message *string }

func (e testMutationError) GetErrorCode() *string    { return e.code }
func (e testMutationError) GetErrorMessage() *string { return e.message }

func TestMutationResultRequiresExplicitSuccess(t *testing.T) {
	t.Parallel()
	success, failure := cato_models.PolicyMutationStatusSuccess, cato_models.PolicyMutationStatusFailure
	for _, test := range []struct {
		name      string
		status    *cato_models.PolicyMutationStatus
		errors    []testMutationError
		wantError bool
	}{
		{name: "success", status: &success},
		{name: "missing status", wantError: true},
		{name: "failure without details", status: &failure, wantError: true},
		{name: "error despite success", status: &success, errors: []testMutationError{{code: new("Rejected")}}, wantError: true},
		{name: "message-only error", status: &failure, errors: []testMutationError{{message: new("policy locked")}}, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := MutationResult(test.status, test.errors)
			if test.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPublishResultAllowsOnlyExplicitMissingRevision(t *testing.T) {
	failure, success := cato_models.PolicyMutationStatusFailure, cato_models.PolicyMutationStatusSuccess
	missing := testMutationError{code: new("PolicyRevisionNotFound")}
	require.NoError(t, PublishResult(&failure, []testMutationError{missing}))
	require.Error(t, PublishResult[testMutationError](nil, nil))
	require.Error(t, PublishResult(&success, []testMutationError{missing}))
	require.Error(t, PublishResult(&failure, []testMutationError{missing, {code: new("AccessDenied")}}))
}

// Generated SDK error getters accept nil pointers. A malformed error list must
// still fail rather than turn an otherwise successful status into success.
func TestMutationResultRejectsNilSDKError(t *testing.T) {
	t.Parallel()
	for _, status := range []cato_models.PolicyMutationStatus{
		cato_models.PolicyMutationStatusSuccess,
		cato_models.PolicyMutationStatusFailure,
	} {
		apiErrors := []*cato.PolicyInternetFirewallAddSubPolicy_Policy_InternetFirewall_AddSubPolicy_Errors{nil}
		require.Error(t, MutationResult(&status, apiErrors))
		require.Error(t, PublishResult(&status, apiErrors))
	}
}
