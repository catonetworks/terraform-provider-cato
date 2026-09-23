package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestPolicyErrLooksLikeConcurrentRevisionBlock(t *testing.T) {
	t.Parallel()
	require.True(t, policyErrLooksLikeConcurrentRevisionBlock("reorderPolicyBlockedByActiveSessions: x"))
	require.True(t, policyErrLooksLikeConcurrentRevisionBlock("Cannot reorder policy while other active revisions exist"))
	require.False(t, policyErrLooksLikeConcurrentRevisionBlock(""))
	require.False(t, policyErrLooksLikeConcurrentRevisionBlock("some other API failure"))
}

func TestDiscardPolicyRevisionTargetsExplicitRevision(t *testing.T) {
	t.Parallel()

	const (
		accountID  = "account-1"
		revisionID = "revision-1"
	)
	tests := map[string]struct {
		operation             string
		policyMutationInput   string
		response              string
		discardPolicyRevision func(context.Context, *cato.Client, string, string) error
	}{
		"internet_firewall": {
			operation:             "policyInternetFirewallDiscardPolicyRevision",
			policyMutationInput:   "internetFirewallPolicyMutationInput",
			response:              `{"data":{"policy":{"internetFirewall":{"discardPolicyRevision":{"errors":[]}}}}}`,
			discardPolicyRevision: discardInternetFirewallPolicyRevision,
		},
		"wan_firewall": {
			operation:             "policyWanFirewallDiscardPolicyRevision",
			policyMutationInput:   "wanFirewallPolicyMutationInput",
			response:              `{"data":{"policy":{"wanFirewall":{"discardPolicyRevision":{"errors":[]}}}}}`,
			discardPolicyRevision: discardWanFirewallPolicyRevision,
		},
		"wan_network": {
			operation:             "policyWanNetworkDiscardPolicyRevision",
			policyMutationInput:   "wanNetworkPolicyMutationInput",
			response:              `{"data":{"policy":{"wanNetwork":{"discardPolicyRevision":{"errors":[]}}}}}`,
			discardPolicyRevision: discardWanNetworkPolicyRevision,
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var request struct {
				OperationName string         `json:"operationName"`
				Variables     map[string]any `json:"variables"`
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() { require.NoError(t, r.Body.Close()) }()
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			client, err := cato.New(server.URL, "test-token", accountID, nil, nil)
			require.NoError(t, err)
			require.NoError(t, tt.discardPolicyRevision(context.Background(), client, accountID, revisionID))
			require.Equal(t, tt.operation, request.OperationName)
			require.Equal(t, accountID, request.Variables["accountId"])
			require.Equal(t, revisionID, nestedRequestString(t, request.Variables, "policyDiscardRevisionInput", "id"))
			require.Equal(t, revisionID, nestedRequestString(t, request.Variables, tt.policyMutationInput, "revision", "id"))
		})
	}
}

func nestedRequestString(t *testing.T, root map[string]any, path ...string) string {
	t.Helper()

	var current any = root
	for _, element := range path {
		values, ok := current.(map[string]any)
		require.Truef(t, ok, "request path %v does not contain an object at %q", path, element)
		current, ok = values[element]
		require.Truef(t, ok, "request path %v is missing %q", path, element)
	}
	value, ok := current.(string)
	require.Truef(t, ok, "request path %v does not contain a string", path)
	return value
}
