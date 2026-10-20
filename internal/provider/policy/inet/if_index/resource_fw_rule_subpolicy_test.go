package inetindex

import (
	"context"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	inetfirewall "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/if_rule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// ifwPolicyWithSubPolicy builds a policy snapshot containing a single sub-policy
// with its scope rule, cleanup rule, and one user rule, all owned by subID.
func ifwPolicyWithSubPolicy(subID, subName string) *cato_go_sdk.Policy {
	rules := []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{
		{
			RuleType:  cato_models.PolicyRuleTypeEnumSubPolicyScope,
			SubPolicy: &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
			Rule:      cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{ID: "scope-1", Name: subName + "-scope"},
		},
		{
			RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
			SubPolicy: &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
			Rule:      cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{ID: "user-1", Name: "user-rule"},
		},
		{
			RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
			SubPolicy: &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
			Rule:      cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{ID: "cleanup-1", Name: subName + " - Cleanup Rule"},
		},
	}
	return &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{
				Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy{
					SubPolicies: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_SubPolicies{
						{Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy_SubPolicies_Policy{ID: subID, Name: subName}},
					},
					Rules: rules,
				},
			},
		},
	}
}

func TestIfwSubPolicyCleanupRuleID(t *testing.T) {
	body := ifwPolicyWithSubPolicy("sub-1", "MySub")

	if got := ifwSubPolicyCleanupRuleID(body, "sub-1"); got != "cleanup-1" {
		t.Fatalf("expected cleanup-1, got %q", got)
	}
	if got := ifwSubPolicyCleanupRuleID(body, "missing"); got != "" {
		t.Fatalf("expected empty for missing sub-policy, got %q", got)
	}
}

func newInternetFwRulePlanWithSubPolicy(ctx context.Context, t *testing.T, subID string) tfsdk.Plan {
	t.Helper()
	model := inetfirewall.NewMinimalInternetFwRuleModel("")
	model.SubPolicyID = types.StringValue(subID)
	plan := tfsdk.Plan{Schema: inetfirewall.GetInternetFwRuleSchema(ctx, t)}
	if diags := plan.Set(ctx, model); diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}
	return plan
}

func TestInternetFwRuleCreateIntoSubPolicy(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallPolicyClient(t)

	// First PolicyInternetFirewall call resolves the cleanup anchor.
	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifwPolicyWithSubPolicy("sub-1", "MySub"), nil).
		Once()

	var capturedAt *cato_models.PolicyRulePositionInput
	mockClient.EXPECT().
		PolicyInternetFirewallAddRule(mock.Anything, mock.Anything, "account-123").
		RunAndReturn(func(_ context.Context, in cato_models.InternetFirewallAddRuleInput, _ string, _ ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicyInternetFirewallAddRule, error) {
			capturedAt = in.At
			return inetfirewall.SuccessfulAddRuleResponse("rule-123"), nil
		}).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()
	// Second PolicyInternetFirewall call hydrates state after publish.
	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(inetfirewall.InternetFirewallPolicyResponseWithRule(inetfirewall.MinimalAPIRule("test-ifw-rule", 10)), nil).
		Once()

	r := &inetfirewall.InternetFwRuleResource{
		Client:    &catoclient.Service{AccountId: "account-123"},
		IfwClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newInternetFwRulePlanWithSubPolicy(ctx, t, "sub-1")}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: inetfirewall.GetInternetFwRuleSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	if capturedAt == nil || capturedAt.Position == nil {
		t.Fatal("expected AddRule to receive an anchored position")
	}
	if *capturedAt.Position != cato_models.PolicyRulePositionEnumBeforeRule {
		t.Fatalf("expected BEFORE_RULE, got %v", *capturedAt.Position)
	}
	if capturedAt.Ref == nil || *capturedAt.Ref != "cleanup-1" {
		t.Fatalf("expected anchor ref cleanup-1, got %v", capturedAt.Ref)
	}
}

func TestInternetFwRuleCreateSubPolicyNotFound(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallPolicyClient(t)

	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(inetfirewall.EmptyInternetFirewallPolicyResponse(), nil).
		Once()

	r := &inetfirewall.InternetFwRuleResource{
		Client:    &catoclient.Service{AccountId: "account-123"},
		IfwClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newInternetFwRulePlanWithSubPolicy(ctx, t, "sub-1")}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: inetfirewall.GetInternetFwRuleSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics when sub-policy cleanup anchor is missing")
	}
}
