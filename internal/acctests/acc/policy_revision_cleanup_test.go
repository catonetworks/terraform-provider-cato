//go:build acctest

package acc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestCleanupMutationErrors(t *testing.T) {
	code, message := "PolicyRevisionNotFound", "Revision was not found"
	missing := &cato.PolicyWanFirewallDiscardPolicyRevision_Policy_WanFirewall_DiscardPolicyRevision_Errors{ErrorCode: &code, ErrorMessage: &message}
	require.NoError(t, cleanupMutationErrors([]*cato.PolicyWanFirewallDiscardPolicyRevision_Policy_WanFirewall_DiscardPolicyRevision_Errors{missing}))
	otherCode, otherMessage := "PolicyDiscardRevisionFailure", "Cannot discard this revision"
	other := &cato.PolicyWanFirewallDiscardPolicyRevision_Policy_WanFirewall_DiscardPolicyRevision_Errors{ErrorCode: &otherCode, ErrorMessage: &otherMessage}
	err := cleanupMutationErrors([]*cato.PolicyWanFirewallDiscardPolicyRevision_Policy_WanFirewall_DiscardPolicyRevision_Errors{missing, other})
	require.ErrorContains(t, err, otherCode)
	require.ErrorContains(t, err, otherMessage)
	require.Error(t, cleanupMutationErrors([]*cato.PolicyWanFirewallDiscardPolicyRevision_Policy_WanFirewall_DiscardPolicyRevision_Errors{nil}))
}

func TestDiscardWanFirewallPolicyRevision(t *testing.T) {
	originalAccount := CatoAccountID
	CatoAccountID = "123"
	defer func() { CatoAccountID = originalAccount }()
	for _, tc := range []struct {
		name, code, message string
		wantError           bool
	}{
		{"already removed", "PolicyRevisionNotFound", "Revision was not found", false},
		{"failure details", "PolicyDiscardRevisionFailure", "Cannot discard this revision", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() { require.NoError(t, r.Body.Close()) }()
				var body struct {
					Variables map[string]any `json:"variables"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, "123", body.Variables["accountId"])
				require.Equal(t, map[string]any{"id": "revision-1"}, body.Variables["policyDiscardRevisionInput"])
				require.Equal(t, map[string]any{"revision": map[string]any{"id": "revision-1"}}, body.Variables["wanFirewallPolicyMutationInput"])
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"data":{"policy":{"wanFirewall":{"discardPolicyRevision":{"status":"FAILURE","errors":[{"errorCode":%q,"errorMessage":%q}]}}}}}`, tc.code, tc.message)
			}))
			defer server.Close()
			client, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
			require.NoError(t, err)
			err = discardWanFirewallPolicyRevision(client, "revision-1")
			if tc.wantError {
				require.ErrorContains(t, err, tc.code)
				require.ErrorContains(t, err, tc.message)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
