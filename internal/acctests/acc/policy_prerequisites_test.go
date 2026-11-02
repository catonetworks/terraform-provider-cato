//go:build acctest

package acc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestPolicyAccessDenied(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       bool
	}{
		{"denied", `{"graphqlErrors":[{"message":"Exception while fetching data (/policy/socketBypass) : Access denied: Insufficient privileges for account 123","path":["policy","socketBypass"]}]}`, true},
		{"wrong account", `{"graphqlErrors":[{"message":"Exception while fetching data (/policy/socketBypass) : Access denied: Insufficient privileges for account 456","path":["policy","socketBypass"]}]}`, false},
		{"wrong policy", `{"graphqlErrors":[{"message":"Exception while fetching data (/policy/socketBypass) : Access denied: Insufficient privileges for account 123","path":["policy","terminalServer"]}]}`, false},
		{"mixed", `{"graphqlErrors":[{"message":"Exception while fetching data (/policy/socketBypass) : Access denied: Insufficient privileges for account 123","path":["policy","socketBypass"]},{"message":"Internal server error"}]}`, false},
		{"network", `{"networkErrors":{"code":503},"graphqlErrors":[{"message":"Exception while fetching data (/policy/socketBypass) : Access denied: Insufficient privileges for account 123","path":["policy","socketBypass"]}]}`, false},
		{"nil entry", `{"graphqlErrors":[null]}`, false},
		{"empty", `{"graphqlErrors":[]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var response clientv2.ErrorResponse
			require.NoError(t, json.Unmarshal([]byte(tc.body), &response))
			require.Equal(t, tc.want, policyAccessDenied(&response, "socketBypass", "123"))
		})
	}
	require.False(t, policyAccessDenied(nil, "socketBypass", "123"))
	require.False(t, policyAccessDenied(errors.New("Access denied"), "socketBypass", "123"))
}

func TestRequirePolicyAccess(t *testing.T) {
	originalClient, originalAccount := catoClient, CatoAccountID
	defer func() { catoClient, CatoAccountID = originalClient, originalAccount }()
	CatoAccountID = "123"
	for _, tc := range []struct {
		name, body string
		skipped    bool
	}{
		{"allowed", `{"data":{"policy":{"socketBypass":{"policy":{"enabled":true}}}}}`, false},
		{"denied", `{"errors":[{"message":"Exception while fetching data (/policy/socketBypass) : Access denied: Insufficient privileges for account 123","path":["policy","socketBypass"]}]}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() { require.NoError(t, r.Body.Close()) }()
				var body struct {
					Query     string         `json:"query"`
					Variables map[string]any `json:"variables"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, "123", body.Variables["accountId"])
				require.Contains(t, body.Query, "socketBypass { policy(input: {revision: {type: PUBLIC}})")
				require.NotContains(t, body.Query, "mutation")
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			var err error
			catoClient, err = cato.New(server.URL, "test-token", "wrong-default", nil, nil)
			require.NoError(t, err)
			skipped := false
			t.Run("prerequisite", func(t *testing.T) {
				defer func() { skipped = t.Skipped() }()
				RequirePolicyAccess(t, "socketBypass")
			})
			require.Equal(t, tc.skipped, skipped)
		})
	}
}
