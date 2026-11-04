package wfsubpolicy

import (
	"context"
	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	wfrule "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func TestNewWfSubPolicyResource(t *testing.T) {
	if r := NewWfSubPolicyResource(); r == nil {
		t.Fatal("expected resource instance, got nil")
	} else if _, ok := r.(*wfSubPolicyResource); !ok {
		t.Fatalf("expected *wfSubPolicyResource, got %T", r)
	}
}

func TestIfSubPolicyMetadata(t *testing.T) {
	r := &wfSubPolicyResource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, resp)
	if resp.TypeName != "cato_wf_sub_policy" {
		t.Fatalf("expected cato_wf_sub_policy, got %q", resp.TypeName)
	}
}

func TestIfSubPolicyConfigure(t *testing.T) {
	r := &wfSubPolicyResource{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{}, resp)
	if r.client != nil {
		t.Fatal("expected nil client when provider data nil")
	}
	client := &catoclient.Service{AccountId: "123"}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, resp)
	if r.client != client {
		t.Fatal("expected client to be set")
	}
}

func TestIfSubPolicyImportState(t *testing.T) {
	ctx := context.Background()
	r := &wfSubPolicyResource{}
	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	resp.State.Set(ctx, newIfSubPolicyModel(""))
	r.ImportState(ctx, resource.ImportStateRequest{ID: "sub-123"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var imported WanFirewallSubPolicy
	resp.State.Get(ctx, &imported)
	if imported.ID.ValueString() != "sub-123" {
		t.Fatalf("expected imported id sub-123, got %q", imported.ID.ValueString())
	}
}

func TestIfSubPolicyCreateAddError(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(wfrule.EmptyWanFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallAddSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(nil, apperror.AssertErr("add failed")).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newIfSubPolicyPlan(ctx, t)}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for add error")
	}
}

func TestIfSubPolicyCreateAddStatusFailure(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(wfrule.EmptyWanFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallAddSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(ifAddSubPolicyResponse(cato_models.PolicyMutationStatusFailure, "boom"), nil).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newIfSubPolicyPlan(ctx, t)}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for failure status")
	}
}

func TestIfSubPolicyCreateSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(wfrule.EmptyWanFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallAddSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(ifAddSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallPublishPolicyRevision(mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("sub-1", "test-sub", "a sub", "scope-1", "scope"), nil).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newIfSubPolicyPlan(ctx, t)}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}

	var state WanFirewallSubPolicy
	resp.State.Get(ctx, &state)
	if state.ID.ValueString() != "sub-1" {
		t.Fatalf("expected sub id sub-1, got %q", state.ID.ValueString())
	}
	if state.ScopeRuleID.ValueString() != "scope-1" {
		t.Fatalf("expected scope rule id scope-1, got %q", state.ScopeRuleID.ValueString())
	}
}

func TestIfSubPolicyCreateNotFoundAfterPublish(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(wfrule.EmptyWanFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallAddSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(ifAddSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallPublishPolicyRevision(mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(wfrule.EmptyWanFirewallPolicyResponse(), nil).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newIfSubPolicyPlan(ctx, t)}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics when created sub-policy not found")
	}
}

func TestIfSubPolicyReadSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("sub-1", "renamed", "desc", "scope-1", "scope-name"), nil).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	state := newIfSubPolicyStateWithID(ctx, t)
	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var got WanFirewallSubPolicy
	resp.State.Get(ctx, &got)
	if got.Name.ValueString() != "renamed" {
		t.Fatalf("expected name renamed, got %q", got.Name.ValueString())
	}
}

func TestIfSubPolicyReadRemovesMissing(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(wfrule.EmptyWanFirewallPolicyResponse(), nil).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	state := newIfSubPolicyStateWithID(ctx, t)
	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatal("expected state removed when sub-policy missing")
	}
}

func TestIfSubPolicyUpdateScopeSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewallUpdateRule(mock.Anything, mock.Anything, "account-123").
		Return(wfrule.SuccessfulUpdateRuleResponse("scope-1"), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallPublishPolicyRevision(mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("sub-1", "test-sub", "desc", "scope-1", "scope-name"), nil).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	req := resource.UpdateRequest{
		Plan:  newIfSubPolicyPlanWithID(ctx, t, "sub-1", "scope-1"),
		State: newIfSubPolicyStateWithID(ctx, t),
	}
	r.Update(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestIfSubPolicyUpdateScopeError(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewallUpdateRule(mock.Anything, mock.Anything, "account-123").
		Return(nil, apperror.AssertErr("update failed")).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	req := resource.UpdateRequest{
		Plan:  newIfSubPolicyPlanWithID(ctx, t, "sub-1", "scope-1"),
		State: newIfSubPolicyStateWithID(ctx, t),
	}
	r.Update(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for update error")
	}
}

func TestIfSubPolicyDeleteSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewallRemoveSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(ifRemoveSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallPublishPolicyRevision(mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: newIfSubPolicyStateWithID(ctx, t)}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestIfSubPolicyDeleteStatusFailure(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewallRemoveSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(ifRemoveSubPolicyResponse(cato_models.PolicyMutationStatusFailure, "cannot"), nil).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: newIfSubPolicyStateWithID(ctx, t)}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for remove failure status")
	}
}

// ---- helpers ----

func getIfSubPolicySchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()
	r := &wfSubPolicyResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	return resp.Schema
}

func newIfSubPolicyModel(id string) WanFirewallSubPolicy {
	scope := wfrule.NewMinimalWanFwRuleModel(id)
	return WanFirewallSubPolicy{
		ID:          convert.NullableString(id),
		Name:        types.StringValue("test-sub"),
		Description: types.StringValue("a sub"),
		ScopeRuleID: types.StringNull(),
		At: types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		}),
		Scope:         scope.Rule,
		CleanupRuleID: types.StringNull(),
		CleanupRule:   types.ObjectNull(wfrule.WanFirewallRuleRuleAttrTypes),
	}
}

func newIfSubPolicyPlan(ctx context.Context, t *testing.T) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: getIfSubPolicySchema(ctx, t)}
	if diags := plan.Set(ctx, newIfSubPolicyModel("")); diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}
	return plan
}

func newIfSubPolicyPlanWithID(ctx context.Context, t *testing.T, id, scopeID string) tfsdk.Plan {
	t.Helper()
	m := newIfSubPolicyModel(id)
	m.ScopeRuleID = types.StringValue(scopeID)
	plan := tfsdk.Plan{Schema: getIfSubPolicySchema(ctx, t)}
	if diags := plan.Set(ctx, m); diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}
	return plan
}

func newIfSubPolicyStateWithID(ctx context.Context, t *testing.T) tfsdk.State {
	t.Helper()
	m := newIfSubPolicyModel("sub-1")
	m.ScopeRuleID = types.StringValue("scope-1")
	state := tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}
	if diags := state.Set(ctx, m); diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	return state
}

func ifAddSubPolicyResponse(status cato_models.PolicyMutationStatus, errMsg string) *cato_go_sdk.PolicyWanFirewallAddSubPolicy {
	var errs []*cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy_WanFirewall_AddSubPolicy_Errors
	if errMsg != "" {
		code := "ERR"
		errs = append(errs, &cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy_WanFirewall_AddSubPolicy_Errors{
			ErrorCode: &code, ErrorMessage: &errMsg,
		})
	}
	return &cato_go_sdk.PolicyWanFirewallAddSubPolicy{
		Policy: cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy{
			WanFirewall: &cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy_WanFirewall{
				AddSubPolicy: cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy_WanFirewall_AddSubPolicy{
					Status: status, Errors: errs,
				},
			},
		},
	}
}

func ifRemoveSubPolicyResponse(status cato_models.PolicyMutationStatus, errMsg string) *cato_go_sdk.PolicyWanFirewallRemoveSubPolicy {
	var errs []*cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy_WanFirewall_RemoveSubPolicy_Errors
	if errMsg != "" {
		code := "ERR"
		errs = append(errs, &cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy_WanFirewall_RemoveSubPolicy_Errors{
			ErrorCode: &code, ErrorMessage: &errMsg,
		})
	}
	return &cato_go_sdk.PolicyWanFirewallRemoveSubPolicy{
		Policy: cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy{
			WanFirewall: &cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy_WanFirewall{
				RemoveSubPolicy: cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy_WanFirewall_RemoveSubPolicy{
					Status: status, Errors: errs,
				},
			},
		},
	}
}

func ifSubPolicyResponse(subID, subName, subDesc, scopeID, scopeName string) *cato_go_sdk.Policy {
	scopeRule := wfrule.MinimalAPIRule(scopeName, 0)
	scopeRule.ID = scopeID
	scopeRule.Name = scopeName
	// The API auto-creates a cleanup rule inside every sub-policy; it is the
	// sub-policy's last rule (highest index).
	cleanupRule := wfrule.MinimalAPIRule(subName+" - Cleanup Rule", 1)
	cleanupRule.ID = "cleanup-1"
	return &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			WanFirewall: &cato_go_sdk.Policy_Policy_WanFirewall{
				Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy{
					SubPolicies: []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies{
						{
							Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies_Policy{
								ID: subID, Name: subName, Description: subDesc,
								Enabled: true, PolicyLevel: cato_models.PolicyLevelEnumSubPolicy,
							},
						},
					},
					Rules: []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules{
						{
							RuleType:  cato_models.PolicyRuleTypeEnumSubPolicyScope,
							SubPolicy: &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
							Rule:      scopeRule,
						},
						{
							RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
							SubPolicy: &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
							Rule:      cleanupRule,
						},
					},
				},
			},
		},
	}
}

var _ = basetypes.ObjectAsOptions{}

// The cleanup identity follows position: whatever rule sits LAST in the
// sub-policy is the cleanup, even after manual moves/renames in CMA.
func TestWanCleanupRuleFollowsLastPosition(t *testing.T) {
	body := ifSubPolicyResponse("sub-1", "test-sub", "desc", "scope-1", "scope-name")
	moved := wfrule.MinimalAPIRule("Moved User Rule", 5)
	moved.ID = "user-9"
	body.Policy.WanFirewall.Policy.Rules = append(
		body.Policy.WanFirewall.Policy.Rules,
		&cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules{
			RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
			SubPolicy: &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_SubPolicy{ID: "sub-1", Name: "test-sub"},
			Rule:      moved,
		},
	)
	got := wanCleanupRule(body, "sub-1")
	if got == nil || got.ID != "user-9" {
		t.Fatalf("expected last rule user-9 as cleanup, got %+v", got)
	}
	if wanCleanupRule(body, "") != nil {
		t.Fatal("expected nil for empty sub id")
	}
}

// Read hydrates the cleanup rule and refreshes cleanup_rule_id — the same path
// that populates the cleanup after `terraform import` (passthrough id + Read).
func TestIfSubPolicyReadHydratesCleanup(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("sub-1", "test-sub", "desc", "scope-1", "scope-name"), nil).Once()

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	state := newIfSubPolicyStateWithID(ctx, t)
	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var got WanFirewallSubPolicy
	resp.State.Get(ctx, &got)
	if got.CleanupRuleID.ValueString() != "cleanup-1" {
		t.Fatalf("expected cleanup_rule_id cleanup-1, got %q", got.CleanupRuleID.ValueString())
	}
	if got.CleanupRule.IsNull() {
		t.Fatal("expected cleanup_rule hydrated, got null")
	}
}

func TestCleanupNameConfigured(t *testing.T) {
	if cleanupNameConfigured(types.ObjectNull(wfrule.WanFirewallRuleRuleAttrTypes)) {
		t.Fatal("null object must not count as configured")
	}
	m := wfrule.NewMinimalWanFwRuleModel("x")
	attrs := m.Rule.Attributes()
	attrs["name"] = types.StringValue("Custom Cleanup Name")
	obj := types.ObjectValueMust(wfrule.WanFirewallRuleRuleAttrTypes, attrs)
	if !cleanupNameConfigured(obj) {
		t.Fatal("explicit name must count as configured")
	}
	attrs["name"] = types.StringNull()
	obj = types.ObjectValueMust(wfrule.WanFirewallRuleRuleAttrTypes, attrs)
	if cleanupNameConfigured(obj) {
		t.Fatal("null name must not count as configured")
	}
}

func TestCleanupUpdateErrorDetail(t *testing.T) {
	plain := cleanupUpdateErrorDetail("someOtherError", "boom", "X - Cleanup Rule", "r-1")
	if plain != "someOtherError : boom" {
		t.Fatalf("unexpected plain detail: %q", plain)
	}
	moved := cleanupUpdateErrorDetail("ruleNameAlreadyExists", "Rule with the same name already exists", "X - Cleanup Rule", "r-1")
	for _, want := range []string{"LAST rule", "r-1", `"X - Cleanup Rule"`, "manually moved", "not intended to be moved outside Terraform"} {
		if !containsStr(moved, want) {
			t.Fatalf("expected detail to contain %q, got: %s", want, moved)
		}
	}
}

func containsStr(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

// Changing `at` must reposition the scope rule via moveRule — the same
// contract as cato_wf_rule/cato_wf_section (input-only on read, applied on
// update when modified).
func TestWfSubPolicyUpdateAtChangeMovesScope(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewallMoveRule(mock.Anything, mock.Anything, "account-123").
		Run(func(ctx context.Context, in cato_models.PolicyMoveRuleInput, accountID string, _ ...clientv2.RequestInterceptor) {
			if in.ID != "scope-1" {
				t.Fatalf("moveRule targeted %q, want scope-1", in.ID)
			}
			if in.To == nil || in.To.Position == nil || *in.To.Position != cato_models.PolicyRulePositionEnumFirstInPolicy {
				t.Fatalf("moveRule position = %+v, want FIRST_IN_POLICY", in.To)
			}
		}).
		Return(wfMoveRuleResponse(cato_models.PolicyMutationStatusSuccess), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallUpdateRule(mock.Anything, mock.Anything, "account-123").
		Return(wfrule.SuccessfulUpdateRuleResponse("scope-1"), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallPublishPolicyRevision(mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("sub-1", "test-sub", "desc", "scope-1", "scope-name"), nil).Once()

	m := newIfSubPolicyModel("sub-1")
	m.ScopeRuleID = types.StringValue("scope-1")
	m.At = types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{
		"position": types.StringValue("FIRST_IN_POLICY"),
		"ref":      types.StringNull(),
	})
	plan := tfsdk.Plan{Schema: getIfSubPolicySchema(ctx, t)}
	if diags := plan.Set(ctx, m); diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}

	r := &wfSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: newIfSubPolicyStateWithID(ctx, t)}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func wfMoveRuleResponse(status cato_models.PolicyMutationStatus) *cato_go_sdk.PolicyWanFirewallMoveRule {
	return &cato_go_sdk.PolicyWanFirewallMoveRule{
		Policy: cato_go_sdk.PolicyWanFirewallMoveRule_Policy{
			WanFirewall: &cato_go_sdk.PolicyWanFirewallMoveRule_Policy_WanFirewall{
				MoveRule: cato_go_sdk.PolicyWanFirewallMoveRule_Policy_WanFirewall_MoveRule{
					Status: status,
				},
			},
		},
	}
}
