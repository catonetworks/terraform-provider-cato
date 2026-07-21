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
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

func TestNewWfSubPolicyResource(t *testing.T) {
	if r := NewWfSubPolicyResource(); r == nil {
		t.Fatal("expected resource instance, got nil")
	} else if _, ok := r.(*wfSubPolicyResource); !ok {
		t.Fatalf("expected *wfSubPolicyResource, got %T", r)
	}
}

func TestWfSubPolicyMetadata(t *testing.T) {
	r := &wfSubPolicyResource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, resp)
	if resp.TypeName != "cato_wf_sub_policy" {
		t.Fatalf("expected cato_wf_sub_policy, got %q", resp.TypeName)
	}
}

func TestWfSubPolicyImportState(t *testing.T) {
	ctx := context.Background()
	r := &wfSubPolicyResource{}
	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: getWfSubPolicySchema(ctx, t)}}
	resp.State.Set(ctx, newWfSubPolicyModel(""))
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

func TestWfSubPolicyCreateSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(emptyWanFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallAddSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(wanAddSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallPublishPolicyRevision(mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(wanSubPolicyResponse("sub-1", "test-sub", "a sub", "scope-1", "scope"), nil).Once()

	r := &wfSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getWfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newWfSubPolicyPlan(ctx, t, "")}, resp)
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

func TestWfSubPolicyCreateAddError(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(emptyWanFirewallPolicyResponse(), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallAddSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(nil, assertErr("add failed")).Once()

	r := &wfSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getWfSubPolicySchema(ctx, t)}}
	r.Create(ctx, resource.CreateRequest{Plan: newWfSubPolicyPlan(ctx, t, "")}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for add error")
	}
}

func TestWfSubPolicyReadSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(wanSubPolicyResponse("sub-1", "renamed", "desc", "scope-1", "scope-name"), nil).Once()

	r := &wfSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
	state := newWfSubPolicyStateWithID(ctx, t)
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

func TestWfSubPolicyReadRemovesMissing(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewall(mock.Anything, mock.Anything, "account-123").
		Return(emptyWanFirewallPolicyResponse(), nil).Once()

	r := &wfSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
	state := newWfSubPolicyStateWithID(ctx, t)
	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatal("expected state removed when sub-policy missing")
	}
}

func TestWfSubPolicyDeleteSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewallRemoveSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(wanRemoveSubPolicyResponse(cato_models.PolicyMutationStatusSuccess, ""), nil).Once()
	mockClient.EXPECT().PolicyWanFirewallPublishPolicyRevision(mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).Once()

	r := &wfSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: newWfSubPolicyStateWithID(ctx, t)}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestWfSubPolicyDeleteStatusFailure(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewWanFirewallSubPolicyClient(t)
	mockClient.EXPECT().PolicyWanFirewallRemoveSubPolicy(mock.Anything, mock.Anything, "account-123").
		Return(wanRemoveSubPolicyResponse(cato_models.PolicyMutationStatusFailure, "cannot"), nil).Once()

	r := &wfSubPolicyResource{client: &catoClientData{AccountId: "account-123"}, subPolyClient: mockClient}
	resp := &resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: newWfSubPolicyStateWithID(ctx, t)}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for remove failure status")
	}
}

// ---- helpers ----

func getWfSubPolicySchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()
	r := &wfSubPolicyResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	return resp.Schema
}

func emptyWanScopeObject(id string) types.Object {
	return types.ObjectValueMust(WanFirewallRuleRuleAttrTypes, map[string]attr.Value{
		"id":                nullableString(id),
		"name":              types.StringValue("test-wf-sub-scope"),
		"description":       types.StringNull(),
		"enabled":           types.BoolValue(true),
		"source":            types.ObjectNull(WanSourceAttrTypes),
		"connection_origin": types.StringValue("ANY"),
		"active_period":     types.ObjectNull(ActivePeriodAttrTypes),
		"country":           types.SetNull(NameIDObjectType),
		"device":            types.SetNull(NameIDObjectType),
		"device_attributes": types.ObjectNull(WanDeviceAttrAttrTypes),
		"device_os":         types.ListNull(types.StringType),
		"application":       types.ObjectNull(WanApplicationAttrTypes),
		"destination":       types.ObjectNull(WanDestAttrTypes),
		"service":           types.ObjectNull(WanServiceAttrTypes),
		"action":            types.StringValue("ALLOW"),
		"tracking":          types.ObjectNull(TrackingAttrTypes),
		"schedule":          types.ObjectNull(ScheduleAttrTypes),
		"direction":         types.StringValue("TO"),
		"exceptions":        types.SetNull(WanExceptionObjectType),
	})
}

func newWfSubPolicyModel(id string) WanFirewallSubPolicy {
	return WanFirewallSubPolicy{
		ID:          nullableString(id),
		Name:        types.StringValue("test-sub"),
		Description: types.StringValue("a sub"),
		ScopeRuleID: types.StringNull(),
		At: types.ObjectValueMust(PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		}),
		Scope: emptyWanScopeObject(id),
	}
}

func newWfSubPolicyPlan(ctx context.Context, t *testing.T, id string) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: getWfSubPolicySchema(ctx, t)}
	if diags := plan.Set(ctx, newWfSubPolicyModel(id)); diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}
	return plan
}

func newWfSubPolicyStateWithID(ctx context.Context, t *testing.T) tfsdk.State {
	t.Helper()
	m := newWfSubPolicyModel("sub-1")
	m.ScopeRuleID = types.StringValue("scope-1")
	state := tfsdk.State{Schema: getWfSubPolicySchema(ctx, t)}
	if diags := state.Set(ctx, m); diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	return state
}

func emptyWanFirewallPolicyResponse() *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: &cato_go_sdk.Policy_Policy{
			WanFirewall: &cato_go_sdk.Policy_Policy_WanFirewall{
				Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules{},
				},
			},
		},
	}
}

func wanAddSubPolicyResponse(status cato_models.PolicyMutationStatus, errMsg string) *cato_go_sdk.PolicyWanFirewallAddSubPolicy {
	var errs []*cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy_WanFirewall_AddSubPolicy_Errors
	if errMsg != "" {
		code := "ERR"
		errs = append(errs, &cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy_WanFirewall_AddSubPolicy_Errors{ErrorCode: &code, ErrorMessage: &errMsg})
	}
	return &cato_go_sdk.PolicyWanFirewallAddSubPolicy{
		Policy: &cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy{
			WanFirewall: &cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy_WanFirewall{
				AddSubPolicy: cato_go_sdk.PolicyWanFirewallAddSubPolicy_Policy_WanFirewall_AddSubPolicy{Status: status, Errors: errs},
			},
		},
	}
}

func wanRemoveSubPolicyResponse(status cato_models.PolicyMutationStatus, errMsg string) *cato_go_sdk.PolicyWanFirewallRemoveSubPolicy {
	var errs []*cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy_WanFirewall_RemoveSubPolicy_Errors
	if errMsg != "" {
		code := "ERR"
		errs = append(errs, &cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy_WanFirewall_RemoveSubPolicy_Errors{ErrorCode: &code, ErrorMessage: &errMsg})
	}
	return &cato_go_sdk.PolicyWanFirewallRemoveSubPolicy{
		Policy: &cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy{
			WanFirewall: &cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy_WanFirewall{
				RemoveSubPolicy: cato_go_sdk.PolicyWanFirewallRemoveSubPolicy_Policy_WanFirewall_RemoveSubPolicy{Status: status, Errors: errs},
			},
		},
	}
}

func wanSubPolicyResponse(subID, subName, subDesc, scopeID, scopeName string) *cato_go_sdk.Policy {
	scopeRule := cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{
		ID:      scopeID,
		Name:    scopeName,
		Enabled: true,
	}
	return &cato_go_sdk.Policy{
		Policy: &cato_go_sdk.Policy_Policy{
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
					},
				},
			},
		},
	}
}
