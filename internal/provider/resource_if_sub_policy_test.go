package provider

import (
	"context"
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
	client := &catoClientData{AccountId: "123"}
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
		Return(emptyInternetFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallAddSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, assertErr("add failed")).Once()

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
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
		Return(emptyInternetFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallAddSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifAddSubPolicyResponse(cato_models.PolicyMutationStatusFailure, "boom"), nil).Once()

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
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
		Return(emptyInternetFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallAddSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifAddSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("sub-1", "test-sub", "a sub", "scope-1", "scope"), nil).Once()

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newIfSubPolicyPlan(ctx, t)}, resp)
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
		Return(emptyInternetFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallAddSubPolicy(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(ifAddSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(emptyInternetFirewallPolicyResponse(), nil).Once()

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
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
		Return(ifSubPolicyResponse("sub-1", "renamed", "desc", "scope-1", "scope-name"), nil).Once()

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
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
		Return(emptyInternetFirewallPolicyResponse(), nil).Once()

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
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
		Return(successfulUpdateRuleResponse("scope-1"), nil).Once()
	mockClient.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()
	mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(ifSubPolicyResponse("sub-1", "test-sub", "desc", "scope-1", "scope-name"), nil).Once()

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
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
		Return(nil, assertErr("update failed")).Once()

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
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
		Return(nil, nil).Once()

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
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

	r := &ifSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
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
	scope := newMinimalInternetFwRuleModel(id)
	return InternetFirewallSubPolicy{
		ID:          nullableString(id),
		Name:        types.StringValue("test-sub"),
		Description: types.StringValue("a sub"),
		ScopeRuleID: types.StringNull(),
		At: types.ObjectValueMust(PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		}),
		Scope: scope.Rule,
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
		Policy: &cato_go_sdk.PolicyInternetFirewallAddSubPolicy_Policy{
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
		Policy: &cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy_Policy_InternetFirewall{
				RemoveSubPolicy: cato_go_sdk.PolicyInternetFirewallRemoveSubPolicy_Policy_InternetFirewall_RemoveSubPolicy{
					Status: status, Errors: errs,
				},
			},
		},
	}
}

func ifSubPolicyResponse(subID, subName, subDesc, scopeID, scopeName string) *cato_go_sdk.Policy {
	scopeRule := minimalAPIRule(scopeName, 0)
	scopeRule.ID = scopeID
	scopeRule.Name = scopeName
	return &cato_go_sdk.Policy{
		Policy: &cato_go_sdk.Policy_Policy{
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
					},
				},
			},
		},
	}
}

var _ = basetypes.ObjectAsOptions{}
