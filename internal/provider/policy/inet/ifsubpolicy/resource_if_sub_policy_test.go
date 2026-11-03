package ifsubpolicy

import (
	"context"
	"encoding/json"
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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func TestNewIfSubPolicyResource(t *testing.T) {
	if r := NewIfSubPolicyResource(); r == nil {
		t.Fatal("expected resource instance, got nil")
	} else if _, ok := r.(*ifSubPolicyResource); !ok {
		t.Fatalf("expected *ifSubPolicyResource, got %T", r)
	}
}

func TestIfSubPolicyMetadata(t *testing.T) {
	r := &ifSubPolicyResource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, resp)
	if resp.TypeName != "cato_if_sub_policy" {
		t.Fatalf("expected cato_if_sub_policy, got %q", resp.TypeName)
	}
}

func TestIfSubPolicyConfigure(t *testing.T) {
	r := &ifSubPolicyResource{}
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
	r := &ifSubPolicyResource{}
	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	resp.State.Set(ctx, newIfSubPolicyModel(""))
	r.ImportState(ctx, resource.ImportStateRequest{ID: "sub-123"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var imported InternetFirewallSubPolicy
	resp.State.Get(ctx, &imported)
	if imported.ID.ValueString() != "sub-123" {
		t.Fatalf("expected imported id sub-123, got %q", imported.ID.ValueString())
	}
}

func TestIfSubPolicyCreateAddError(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifrule.EmptyInternetFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallAddSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, apperror.AssertErr("add failed")).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newIfSubPolicyPlan(ctx, t)}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for add error")
	}
}

func TestIfSubPolicyCreateAddStatusFailure(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifrule.EmptyInternetFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallAddSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifAddSubPolicyResponse(cato_models.PolicyMutationStatusFailure, "boom"), nil).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newIfSubPolicyPlan(ctx, t)}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for failure status")
	}
}

func TestIfSubPolicyCreateSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifrule.EmptyInternetFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallAddSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifAddSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(successfulSubPolicyPublish(t), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("test-sub", "a sub", "scope"), nil).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	plan := newIfSubPolicyPlan(ctx, t)
	var replacement InternetFirewallSubPolicy
	plan.Get(ctx, &replacement)
	replacement.CleanupRuleID = types.StringValue("previous-resource-cleanup")
	plan.Set(ctx, replacement)
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}

	var state InternetFirewallSubPolicy
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
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifrule.EmptyInternetFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallAddSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifAddSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(successfulSubPolicyPublish(t), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifrule.EmptyInternetFirewallPolicyResponse(), nil).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newIfSubPolicyPlan(ctx, t)}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics when created sub-policy not found")
	}
}

func TestIfSubPolicyReadSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("renamed", "desc", "scope-name"), nil).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	state := newIfSubPolicyStateWithID(ctx, t)
	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var got InternetFirewallSubPolicy
	resp.State.Get(ctx, &got)
	if got.Name.ValueString() != "renamed" {
		t.Fatalf("expected name renamed, got %q", got.Name.ValueString())
	}
}

func TestIfSubPolicyReadRemovesMissing(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifrule.EmptyInternetFirewallPolicyResponse(), nil).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
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
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewallUpdateRule(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifrule.SuccessfulUpdateRuleResponse("scope-1"), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(successfulSubPolicyPublish(t), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("test-sub", "desc", "scope-name"), nil).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
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
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewallUpdateRule(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, apperror.AssertErr("update failed")).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
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
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewallRemoveSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifRemoveSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(successfulSubPolicyPublish(t), nil).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: newIfSubPolicyStateWithID(ctx, t)}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestIfSubPolicyDeleteStatusFailure(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewallRemoveSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifRemoveSubPolicyResponse(cato_models.PolicyMutationStatusFailure, "cannot"), nil).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: newIfSubPolicyStateWithID(ctx, t)}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for remove failure status")
	}
}

// ---- helpers ----

func getIfSubPolicySchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()
	r := &ifSubPolicyResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	return resp.Schema
}

func newIfSubPolicyModel(id string) InternetFirewallSubPolicy {
	scope := ifrule.NewMinimalInternetFwRuleModel(id)
	return InternetFirewallSubPolicy{
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
		CleanupRule:   types.ObjectNull(ifrule.InternetFirewallRuleRuleAttrTypes),
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

func ifAddSubPolicyResponse(status cato_models.PolicyMutationStatus, errMsg string) *cato_go_sdk.PolicyInternetFirewallAddSubPolicy {
	var errs []*cato_go_sdk.PolicyInternetFirewallAddSubPolicy_Policy_InternetFirewall_AddSubPolicy_Errors
	if errMsg != "" {
		code := "ERR"
		errs = append(errs, &cato_go_sdk.PolicyInternetFirewallAddSubPolicy_Policy_InternetFirewall_AddSubPolicy_Errors{
			ErrorCode: &code, ErrorMessage: &errMsg,
		})
	}
	return &cato_go_sdk.PolicyInternetFirewallAddSubPolicy{
		Policy: cato_go_sdk.PolicyInternetFirewallAddSubPolicy_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallAddSubPolicy_Policy_InternetFirewall{
				AddSubPolicy: cato_go_sdk.PolicyInternetFirewallAddSubPolicy_Policy_InternetFirewall_AddSubPolicy{
					Status: status, Errors: errs,
				},
			},
		},
	}
}

func ifRemoveSubPolicyResponse(status cato_models.PolicyMutationStatus, errMsg string) *cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy {
	var errs []*cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy_Policy_InternetFirewall_RemoveSubPolicy_Errors
	if errMsg != "" {
		code := "ERR"
		errs = append(errs, &cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy_Policy_InternetFirewall_RemoveSubPolicy_Errors{
			ErrorCode: &code, ErrorMessage: &errMsg,
		})
	}
	return &cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy{
		Policy: cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy_Policy_InternetFirewall{
				RemoveSubPolicy: cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy_Policy_InternetFirewall_RemoveSubPolicy{
					Status: status, Errors: errs,
				},
			},
		},
	}
}

func ifSubPolicyResponse(subName, subDesc, scopeName string) *cato_go_sdk.Policy {
	const subID = "sub-1"
	const scopeID = "scope-1"
	scopeRule := ifrule.MinimalAPIRule(scopeName, 0)
	scopeRule.ID = scopeID
	scopeRule.Name = scopeName
	// The API auto-creates a cleanup rule inside every sub-policy; it is the
	// sub-policy's last rule (highest index).
	cleanupRule := ifrule.MinimalAPIRule(subName+" - Cleanup Rule", 1)
	cleanupRule.ID = "cleanup-1"
	return &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{
				Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy{
					SubPolicies: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_SubPolicies{
						{
							Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy_SubPolicies_Policy{
								ID: subID, Name: subName, Description: subDesc,
								Enabled: true, PolicyLevel: cato_models.PolicyLevelEnumSubPolicy,
							},
						},
					},
					Rules: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{
						{
							RuleType:  cato_models.PolicyRuleTypeEnumSubPolicyScope,
							SubPolicy: &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
							Rule:      scopeRule,
						},
						{
							RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
							SubPolicy: &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_SubPolicy{ID: subID, Name: subName},
							Rule:      cleanupRule,
						},
					},
				},
			},
		},
	}
}

var _ = basetypes.ObjectAsOptions{}

// A stored cleanup identity must survive manual moves and renames in CMA.
func TestIfwCleanupRulePreservesStoredIdentity(t *testing.T) {
	body := ifSubPolicyResponse("test-sub", "desc", "scope-name")
	moved := ifrule.MinimalAPIRule("Moved User Rule", 5)
	moved.ID = "user-9"
	body.Policy.InternetFirewall.Policy.Rules = append(
		body.Policy.InternetFirewall.Policy.Rules,
		&cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{
			RuleType:  cato_models.PolicyRuleTypeEnumPolicyRule,
			SubPolicy: &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_SubPolicy{ID: "sub-1", Name: "test-sub"},
			Rule:      moved,
		},
	)
	got := ifwCleanupRule(body, "sub-1", "cleanup-1")
	if got == nil || got.ID != "cleanup-1" {
		t.Fatalf("expected stored cleanup-1 despite external reorder, got %+v", got)
	}
	if ifwCleanupRule(body, "", "") != nil {
		t.Fatal("expected nil for empty sub id")
	}
}

// Read hydrates the cleanup rule and refreshes cleanup_rule_id — the same path
// that populates the cleanup after `terraform import` (passthrough id + Read).
func TestIfSubPolicyReadHydratesCleanup(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("test-sub", "desc", "scope-name"), nil).Once()

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	state := newIfSubPolicyStateWithID(ctx, t)
	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var got InternetFirewallSubPolicy
	resp.State.Get(ctx, &got)
	if got.CleanupRuleID.ValueString() != "cleanup-1" {
		t.Fatalf("expected cleanup_rule_id cleanup-1, got %q", got.CleanupRuleID.ValueString())
	}
	if got.CleanupRule.IsNull() {
		t.Fatal("expected cleanup_rule hydrated, got null")
	}
}

func TestCleanupNameConfigured(t *testing.T) {
	if cleanupNameConfigured(types.ObjectNull(ifrule.InternetFirewallRuleRuleAttrTypes)) {
		t.Fatal("null object must not count as configured")
	}
	m := ifrule.NewMinimalInternetFwRuleModel("x")
	attrs := m.Rule.Attributes()
	attrs["name"] = types.StringValue("Custom Cleanup Name")
	obj := types.ObjectValueMust(ifrule.InternetFirewallRuleRuleAttrTypes, attrs)
	if !cleanupNameConfigured(obj) {
		t.Fatal("explicit name must count as configured")
	}
	attrs["name"] = types.StringNull()
	obj = types.ObjectValueMust(ifrule.InternetFirewallRuleRuleAttrTypes, attrs)
	if cleanupNameConfigured(obj) {
		t.Fatal("null name must not count as configured")
	}
}

func successfulSubPolicyPublish(t *testing.T) *cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision {
	t.Helper()
	var response cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision
	if err := json.Unmarshal([]byte(`{"policy":{"internetFirewall":{"publishPolicyRevision":{"status":"SUCCESS"}}}}`), &response); err != nil {
		t.Fatal(err)
	}
	return &response
}

func TestMissingTrackedCleanupPreservesSubPolicyState(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("test-sub", "desc", "scope-name"), nil).Once()
	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	state := newIfSubPolicyStateWithID(ctx, t)
	var model InternetFirewallSubPolicy
	state.Get(ctx, &model)
	model.CleanupRuleID = types.StringValue("missing-cleanup")
	state.Set(ctx, model)
	response := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("missing tracked cleanup must require reconciliation")
	}
	if !state.Raw.Equal(response.State.Raw) {
		t.Fatal("existing sub-policy state must not be removed")
	}
}

func TestSubPolicyPublishRejectsPayloadFailure(t *testing.T) {
	for _, payload := range []string{`{"status":"FAILURE","errors":[]}`, `{}`, `{"status":"SUCCESS","errors":[{"errorCode":"Rejected"}]}`} {
		ctx := context.Background()
		client := mocks.NewInternetFirewallSubPolicyClient(t)
		var response cato_go_sdk.PolicyInternetFirewallPublishPolicyRevision
		if err := json.Unmarshal([]byte(`{"policy":{"internetFirewall":{"publishPolicyRevision":`+payload+`}}}`), &response); err != nil {
			t.Fatal(err)
		}
		client.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").Return(&response, nil).Once()
		r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: client}
		if err := r.publish(ctx); err == nil {
			t.Fatalf("must reject payload %s", payload)
		}
	}
}

// Changing `at` must reposition the scope rule via moveRule — the same
// contract as cato_if_rule/cato_if_section (input-only on read, applied on
// update when modified).
func TestIfSubPolicyUpdateAtChangeMovesScope(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyInternetFirewallMoveRule(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Run(func(ctx context.Context, mutIn *cato_models.InternetFirewallPolicyMutationInput, in cato_models.PolicyMoveRuleInput, accountID string, _ ...clientv2.RequestInterceptor) {
			if in.ID != "scope-1" {
				t.Fatalf("moveRule targeted %q, want scope-1", in.ID)
			}
			if in.To == nil || in.To.Position == nil || *in.To.Position != cato_models.PolicyRulePositionEnumFirstInPolicy {
				t.Fatalf("moveRule position = %+v, want FIRST_IN_POLICY", in.To)
			}
		}).
		Return(ifMoveRuleResponse(cato_models.PolicyMutationStatusSuccess), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallUpdateRule(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifrule.SuccessfulUpdateRuleResponse("scope-1"), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(successfulSubPolicyPublish(t), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("test-sub", "desc", "scope-name"), nil).Once()

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

	r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: newIfSubPolicyStateWithID(ctx, t)}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func ifMoveRuleResponse(status cato_models.PolicyMutationStatus) *cato_go_sdk.PolicyInternetFirewallMoveRule {
	return &cato_go_sdk.PolicyInternetFirewallMoveRule{
		Policy: cato_go_sdk.PolicyInternetFirewallMoveRule_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallMoveRule_Policy_InternetFirewall{
				MoveRule: cato_go_sdk.PolicyInternetFirewallMoveRule_Policy_InternetFirewall_MoveRule{
					Status: status,
				},
			},
		},
	}
}
