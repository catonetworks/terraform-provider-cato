package catoapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	privateaccess "github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
)

func TestPublishPrivateAccessPolicyRevision(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, response string
		want           error
	}{
		{"success", `{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS"}}}}}`, nil},
		{"no draft", `{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[{"errorCode":"PolicyRevisionNotFound"}]}}}}}`, privateaccess.ErrRejected},
		{"mixed failure", `{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[{"errorCode":"PolicyRevisionNotFound"},{"errorCode":"PermissionDenied"}]}}}}}`, privateaccess.ErrRejected},
		{"missing status", `{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{}}}}}`, privateaccess.ErrInvalidResponse},
		{"null response", `{"data":{"policy":null}}`, privateaccess.ErrInvalidResponse},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, requests := testCatoAPI(t, tc.response)
			err := NewAdapter(client).PublishPrivateAccessPolicyRevision(context.Background(), "account-123")
			if tc.want == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.want)
			}
			if tc.want == privateaccess.ErrRejected {
				require.NotErrorIs(t, err, privateaccess.ErrNoDraft)
			}
			request := <-requests
			require.Equal(t, "policyPrivateAccessPublishRevision", request.OperationName)
			require.JSONEq(t, `"account-123"`, string(request.Variables["accountID"]))
		})
	}
}

func TestQueryPrivateAccessPolicy(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, response string
		exists         bool
		want           error
	}{
		{"absent", `{"data":{"policy":{"privateAccess":{"policy":{"rules":[]}}}}}`, false, nil},
		{"present", `{"data":{"policy":{"privateAccess":{"policy":{"rules":[{"rule":{"id":"rule-456"}}]}}}}}`, true, nil},
		{"other rule", `{"data":{"policy":{"privateAccess":{"policy":{"rules":[{"rule":{"id":"other"}}]}}}}}`, false, nil},
		{"missing policy", `{"data":{"policy":{"privateAccess":{"policy":null}}}}`, false, privateaccess.ErrInvalidResponse},
		{"missing rules", `{"data":{"policy":{"privateAccess":{"policy":{}}}}}`, false, privateaccess.ErrInvalidResponse},
		{"null rules", `{"data":{"policy":{"privateAccess":{"policy":{"rules":null}}}}}`, false, privateaccess.ErrInvalidResponse},
		{"nil row", `{"data":{"policy":{"privateAccess":{"policy":{"rules":[null]}}}}}`, false, privateaccess.ErrInvalidResponse},
		{"nil rule", `{"data":{"policy":{"privateAccess":{"policy":{"rules":[{"rule":null}]}}}}}`, false, privateaccess.ErrInvalidResponse},
		{"missing id", `{"data":{"policy":{"privateAccess":{"policy":{"rules":[{"rule":{}}]}}}}}`, false, privateaccess.ErrInvalidResponse},
		{"graphql", `{"errors":[{"message":"sensitive backend details"}]}`, false, privateaccess.ErrUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, requests := testCatoAPI(t, tc.response)
			policy, err := NewAdapter(client).QueryPrivateAccessPolicy(context.Background(), "account-123", entities.PublicPolicyRevision)
			if tc.want == nil {
				require.NoError(t, err)
				require.Equal(t, tc.exists, policy.HasRule("rule-456"))
			} else {
				require.ErrorIs(t, err, tc.want)
			}
			request := <-requests
			require.Equal(t, "refactorPrivateAccessPolicy", request.OperationName)
			require.Contains(t, request.Query, "type: $revision")
			require.JSONEq(t, `"PUBLIC"`, string(request.Variables["revision"]))
			require.JSONEq(t, `"account-123"`, string(request.Variables["accountID"]))
		})
	}
}
