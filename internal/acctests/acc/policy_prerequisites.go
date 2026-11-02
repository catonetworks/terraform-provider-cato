//go:build acctest

package acc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
)

// RequirePolicyAccess checks optional policy prerequisites before creating fixtures.
// Only a structured denial for this exact policy/account is an environment skip.
func RequirePolicyAccess(t *testing.T, policy string) {
	t.Helper()
	switch policy {
	case "siteWebProxy", "socketBypass", "terminalServer":
	default:
		t.Fatalf("unsupported policy prerequisite %q", policy)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	document := fmt.Sprintf(`query acceptancePolicyAccess($accountId: ID!) {
 policy(accountId: $accountId) { %s { policy(input: {revision: {type: PUBLIC}}) { enabled } } }
}`, policy)
	type policyAccess struct {
		Policy *struct {
			Enabled bool `json:"enabled" graphql:"enabled"`
		} `json:"policy" graphql:"policy"`
	}
	var result struct {
		Policy struct {
			SiteWebProxy   *policyAccess `json:"siteWebProxy" graphql:"siteWebProxy"`
			SocketBypass   *policyAccess `json:"socketBypass" graphql:"socketBypass"`
			TerminalServer *policyAccess `json:"terminalServer" graphql:"terminalServer"`
		} `json:"policy" graphql:"policy"`
	}

	err := GetClient(t).Client.Post(ctx, "acceptancePolicyAccess", document, &result, map[string]any{"accountId": CatoAccountID})
	if policyAccessDenied(err, policy, CatoAccountID) {
		t.Skipf("skipping test: CI account %s lacks access to %s; run with an authorized account for live coverage", CatoAccountID, policy)
	}
	if err != nil {
		t.Fatalf("check %s policy access: %v", policy, err)
	}
	selected := map[string]*policyAccess{
		"siteWebProxy":   result.Policy.SiteWebProxy,
		"socketBypass":   result.Policy.SocketBypass,
		"terminalServer": result.Policy.TerminalServer,
	}[policy]
	if selected == nil || selected.Policy == nil {
		t.Fatalf("check %s policy access: API returned no policy", policy)
	}
}

func policyAccessDenied(err error, policy, account string) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.NetworkError != nil || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
		return false
	}
	expected := fmt.Sprintf("Exception while fetching data (/policy/%s) : "+
		"Access denied: Insufficient privileges for account %s", policy, account)
	for _, item := range *response.GqlErrors {
		if item == nil || item.Message != expected || len(item.Path) != 2 || item.Path.String() != "policy."+policy {
			return false
		}
	}
	return true
}
