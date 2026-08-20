//go:build acctest

package acc

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type testFirewallPayloadError struct {
	code    *string
	message *string
}

func (e *testFirewallPayloadError) GetErrorCode() *string {
	return e.code
}

func (e *testFirewallPayloadError) GetErrorMessage() *string {
	return e.message
}

func TestRunFirewallCleanupPublishesAfterPartialSuccess(t *testing.T) {
	deleteErr := errors.New("delete failed")
	var calls []string
	plan := firewallCleanupPlan{
		rules:       []firewallCleanupElement{{name: "rule"}},
		sections:    []firewallCleanupElement{{name: "section"}},
		subPolicies: []firewallCleanupElement{{name: "sub-policy"}},
	}
	operations := firewallCleanupOperations{
		deleteRule: func(element firewallCleanupElement) (bool, error) {
			calls = append(calls, "rule:"+element.name)
			return false, deleteErr
		},
		deleteSection: func(element firewallCleanupElement) (bool, error) {
			calls = append(calls, "section:"+element.name)
			return true, nil
		},
		deleteSubPolicy: func(element firewallCleanupElement) (bool, error) {
			calls = append(calls, "sub-policy:"+element.name)
			return true, nil
		},
		publish: func() error {
			calls = append(calls, "publish")
			return nil
		},
	}

	err := runFirewallCleanup(plan, operations)

	if !errors.Is(err, deleteErr) {
		t.Fatalf("runFirewallCleanup() error = %v, want joined delete error", err)
	}
	wantCalls := []string{"rule:rule", "section:section", "sub-policy:sub-policy", "publish"}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", calls, wantCalls)
	}
}

func TestRunFirewallCleanupDoesNotPublishWithoutSuccessfulMutation(t *testing.T) {
	deleteErr := errors.New("delete failed")
	published := false
	plan := firewallCleanupPlan{
		rules: []firewallCleanupElement{{name: "rule"}},
	}
	operations := firewallCleanupOperations{
		deleteRule: func(firewallCleanupElement) (bool, error) { return false, deleteErr },
		deleteSection: func(firewallCleanupElement) (bool, error) {
			t.Fatal("unexpected section deletion")
			return false, nil
		},
		deleteSubPolicy: func(firewallCleanupElement) (bool, error) {
			t.Fatal("unexpected sub-policy deletion")
			return false, nil
		},
		publish: func() error {
			published = true
			return nil
		},
	}

	err := runFirewallCleanup(plan, operations)

	if !errors.Is(err, deleteErr) {
		t.Fatalf("runFirewallCleanup() error = %v, want delete error", err)
	}
	if published {
		t.Fatal("runFirewallCleanup() published without a successful mutation")
	}
}

func TestIsFirewallCleanupCandidate(t *testing.T) {
	tests := []struct {
		name       string
		element    string
		properties []cato_models.PolicyElementPropertiesEnum
		want       bool
	}{
		{name: "matching user element", element: "acctest_rule", want: true},
		{name: "different prefix", element: "production_rule", want: false},
		{
			name:       "system element",
			element:    "acctest_system_rule",
			properties: []cato_models.PolicyElementPropertiesEnum{cato_models.PolicyElementPropertiesEnumSystem},
			want:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isFirewallCleanupCandidate(test.element, test.properties); got != test.want {
				t.Fatalf("isFirewallCleanupCandidate() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestFirewallMutationResultIncludesStatusAndPayloadErrors(t *testing.T) {
	status := cato_models.PolicyMutationStatusFailure
	code := "INVALID"
	message := "cannot delete"

	changed, err := firewallMutationResult(
		"deleting rule",
		nil,
		&status,
		[]*testFirewallPayloadError{{code: &code, message: &message}},
	)

	if changed {
		t.Fatal("firewallMutationResult() changed = true for failed status")
	}
	if err == nil {
		t.Fatal("firewallMutationResult() error = nil")
	}
	for _, want := range []string{"mutation status FAILURE", "API error INVALID: cannot delete"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("firewallMutationResult() error = %q, want substring %q", err, want)
		}
	}
}

func TestRunApplicationPolicyCleanupOrdersDeletesAndPublishes(t *testing.T) {
	var calls []string
	operations := applicationPolicyCleanupOperations{
		removeRule: func(id string) (cleanupMutationResult, error) {
			calls = append(calls, "rule:"+id)
			return successfulCleanupMutation(), nil
		},
		removeSection: func(id string) (cleanupMutationResult, error) {
			calls = append(calls, "section:"+id)
			return successfulCleanupMutation(), nil
		},
		publish: func() error {
			calls = append(calls, "publish")
			return nil
		},
	}

	err := runApplicationPolicyCleanup(
		[]cleanupPolicyElement{{id: "rule-id"}},
		[]cleanupPolicyElement{{id: "section-id"}},
		operations,
	)
	if err != nil {
		t.Fatalf("runApplicationPolicyCleanup() error = %v", err)
	}
	want := []string{"rule:rule-id", "section:section-id", "publish"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestHasUnownedSocketLanFirewallRules(t *testing.T) {
	firewallRule := func(name string) *cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall {
		return &cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall{
			Rule: cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule{
				Name: name,
			},
		}
	}

	if hasUnownedSocketLanFirewallRules([]*cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall{
		firewallRule("acctest_child"),
	}) {
		t.Fatal("owned firewall rule reported as unowned")
	}
	if !hasUnownedSocketLanFirewallRules([]*cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall{
		firewallRule("production_child"),
	}) {
		t.Fatal("unowned firewall rule not detected")
	}
}

func successfulCleanupMutation() cleanupMutationResult {
	status := cato_models.PolicyMutationStatusSuccess
	return cleanupMutationResult{status: &status}
}
