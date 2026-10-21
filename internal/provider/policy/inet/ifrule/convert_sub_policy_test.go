package ifrule

import "testing"

func TestIfwSubPolicyCleanupRuleID(t *testing.T) {
	body := ifwPolicyWithSubPolicy("sub-1", "MySub")

	if got := ifwSubPolicyCleanupRuleID(body, "sub-1"); got != "cleanup-1" {
		t.Fatalf("expected cleanup-1, got %q", got)
	}
	if got := ifwSubPolicyCleanupRuleID(body, "missing"); got != "" {
		t.Fatalf("expected empty for missing sub-policy, got %q", got)
	}
}
