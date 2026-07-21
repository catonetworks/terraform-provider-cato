package provider

import (
	"context"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
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
		Policy: &cato_go_sdk.Policy_Policy{
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

func TestWanSubPolicyCleanupRuleID(t *testing.T) {
	subID, subName := "sub-w", "MyWanSub"
	rules := []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules{
		{
			RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
			SubPolicy: &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
			Rule:      cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{ID: "wuser-1", Name: "wan-user"},
		},
		{
			RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
			SubPolicy: &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
			Rule:      cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{ID: "wcleanup-1", Name: subName + " - Cleanup Rule"},
		},
	}
	body := &cato_go_sdk.Policy{
		Policy: &cato_go_sdk.Policy_Policy{
			WanFirewall: &cato_go_sdk.Policy_Policy_WanFirewall{
				Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy{
					SubPolicies: []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies{
						{Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies_Policy{ID: subID, Name: subName}},
					},
					Rules: rules,
				},
			},
		},
	}

	if got := wanSubPolicyCleanupRuleID(body, subID); got != "wcleanup-1" {
		t.Fatalf("expected wcleanup-1, got %q", got)
	}
	if got := wanSubPolicyCleanupRuleID(body, "nope"); got != "" {
		t.Fatalf("expected empty for missing sub-policy, got %q", got)
	}
}

func TestInternetFwRuleImportStateComposite(t *testing.T) {
	ctx := context.Background()
	r := &internetFwRuleResource{}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: getInternetFwRuleSchema(ctx, t)},
	}
	diags := resp.State.Set(ctx, InternetFirewallRule{
		Rule: types.ObjectNull(InternetFirewallRuleRuleAttrTypes),
		At:   types.ObjectNull(PositionAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("unexpected seed state diagnostics: %+v", diags)
	}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "sub-42/rule-9"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}

	var imported InternetFirewallRule
	diags = resp.State.Get(ctx, &imported)
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	if imported.SubPolicyID.ValueString() != "sub-42" {
		t.Fatalf("expected sub_policy_id sub-42, got %q", imported.SubPolicyID.ValueString())
	}
	ruleModel := PolicyPolicyInternetFirewallPolicyRulesRule{}
	imported.Rule.As(ctx, &ruleModel, basetypes.ObjectAsOptions{})
	if ruleModel.ID.ValueString() != "rule-9" {
		t.Fatalf("expected rule id rule-9, got %q", ruleModel.ID.ValueString())
	}
}

func TestInternetFwRuleImportStateCompositeInvalid(t *testing.T) {
	ctx := context.Background()
	r := &internetFwRuleResource{}
	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: getInternetFwRuleSchema(ctx, t)}}
	_ = resp.State.Set(ctx, InternetFirewallRule{
		Rule: types.ObjectNull(InternetFirewallRuleRuleAttrTypes),
		At:   types.ObjectNull(PositionAttrTypes),
	})

	r.ImportState(ctx, resource.ImportStateRequest{ID: "sub-42/"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for malformed composite import id")
	}
}

func newInternetFwRulePlanWithSubPolicy(ctx context.Context, t *testing.T, subID string) tfsdk.Plan {
	t.Helper()
	model := newMinimalInternetFwRuleModel("")
	model.SubPolicyID = types.StringValue(subID)
	plan := tfsdk.Plan{Schema: getInternetFwRuleSchema(ctx, t)}
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
			return successfulAddRuleResponse("rule-123"), nil
		}).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()
	// Second PolicyInternetFirewall call hydrates state after publish.
	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(internetFirewallPolicyResponseWithRule(minimalAPIRule("test-ifw-rule", 10)), nil).
		Once()

	r := &internetFwRuleResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newInternetFwRulePlanWithSubPolicy(ctx, t, "sub-1")}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getInternetFwRuleSchema(ctx, t)}}

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
		Return(emptyInternetFirewallPolicyResponse(), nil).
		Once()

	r := &internetFwRuleResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newInternetFwRulePlanWithSubPolicy(ctx, t, "sub-1")}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getInternetFwRuleSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics when sub-policy cleanup anchor is missing")
	}
}
