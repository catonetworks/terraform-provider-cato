package catoclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/require"
)

func TestSplitTunnelUpdateDNSExclusionRequest(t *testing.T) {
	for _, tc := range []struct {
		name      string
		exclusion *cato_models.SplitTunnelDNSExclusionUpdateInput
		expected  any
	}{
		{"omitted stays omitted", nil, nil},
		{"empty clears domains", &cato_models.SplitTunnelDNSExclusionUpdateInput{Domain: []string{}}, map[string]any{"domain": []any{}}},
		{"populated domains", &cato_models.SplitTunnelDNSExclusionUpdateInput{Domain: []string{"example.com"}}, map[string]any{"domain": []any{"example.com"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var request struct {
				OperationName string         `json:"operationName"`
				Query         string         `json:"query"`
				Variables     map[string]any `json:"variables"`
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() { require.NoError(t, r.Body.Close()) }()
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{"data":{"policy":{"splitTunnel":{"updateRule":{"status":"SUCCESS","errors":[]}}}}}`))
				require.NoError(t, err)
			}))
			defer server.Close()
			sdk, err := cato.New(server.URL, "test-token", "account-1", nil, nil)
			require.NoError(t, err)
			description := "preserve sibling fields"
			result, err := NewProviderSDKClient(sdk).PolicySplitTunnelUpdateRule(context.Background(), cato_models.SplitTunnelUpdateRuleInput{
				ID: "rule-1", Rule: &cato_models.SplitTunnelUpdateRuleDataInput{Description: &description, DNSExclusion: tc.exclusion},
			}, "account-1")
			require.NoError(t, err)
			require.NotNil(t, result.GetPolicy().GetSplitTunnel().GetUpdateRule().GetStatus())
			require.Equal(t, cato_models.PolicyMutationStatusSuccess, *result.GetPolicy().GetSplitTunnel().GetUpdateRule().GetStatus())
			require.Equal(t, "policySplitTunnelUpdateRule", request.OperationName)
			require.Equal(t, cato.PolicySplitTunnelUpdateRuleDocument, request.Query)
			require.Equal(t, "account-1", request.Variables["accountId"])
			require.Nil(t, request.Variables["splitTunnelPolicyMutationInput"])
			input := request.Variables["splitTunnelUpdateRuleInput"].(map[string]any)
			require.Equal(t, "rule-1", input["id"])
			rule := input["rule"].(map[string]any)
			require.Equal(t, description, rule["description"])
			require.Equal(t, tc.expected, rule["dnsExclusion"])
		})
	}
}
