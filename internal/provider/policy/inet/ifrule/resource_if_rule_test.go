package ifrule

import (
	"context"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func TestNewInternetFwRuleResource(t *testing.T) {
	r := NewInternetFwRuleResource()

	if r == nil {
		t.Fatal("expected resource instance, got nil")
	}

	if _, ok := r.(*InternetFwRuleResource); !ok {
		t.Fatalf("expected *internetFwRuleResource, got %T", r)
	}
}

func TestInternetFwRuleMetadata(t *testing.T) {
	ctx := context.Background()
	r := &InternetFwRuleResource{}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, resp)

	if resp.TypeName != "cato_if_rule" {
		t.Fatalf("expected type name cato_if_rule, got %q", resp.TypeName)
	}
}

func TestInternetFwRuleConfigureNilProviderData(t *testing.T) {
	r := &InternetFwRuleResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{}, resp)

	if r.Client != nil {
		t.Fatal("expected client to remain nil when provider data is nil")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestInternetFwRuleConfigureSetsClient(t *testing.T) {
	client := &catoclient.Service{AccountId: "123"}
	r := &InternetFwRuleResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, resp)

	if r.Client != client {
		t.Fatal("expected resource client to be set from provider data")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestInternetFwRuleImportState(t *testing.T) {
	ctx := context.Background()
	r := &InternetFwRuleResource{}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: GetInternetFwRuleSchema(ctx, t)},
	}

	diags := resp.State.Set(ctx, InternetFirewallRule{
		Rule: types.ObjectNull(InternetFirewallRuleRuleAttrTypes),
		At:   types.ObjectNull(policyrule.PositionAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("unexpected seed state diagnostics: %+v", diags)
	}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "rule-123"}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}

	var imported InternetFirewallRule
	diags = resp.State.Get(ctx, &imported)
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}

	ruleModel := PolicyPolicyInternetFirewallPolicyRulesRule{}
	diags = imported.Rule.As(ctx, &ruleModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("unexpected rule diagnostics: %+v", diags)
	}

	if ruleModel.ID.ValueString() != "rule-123" {
		t.Fatalf("expected imported rule id rule-123, got %q", ruleModel.ID.ValueString())
	}
}

func TestUseStateExceptionDeviceAttributesUsesMatchingException(t *testing.T) {
	ctx := context.Background()

	topLevelDeviceAttributes := ifwDeviceAttributesObject([]string{"top-level-category"})
	exceptionDeviceAttributes := ifwDeviceAttributesObject([]string{"exception-category"})

	model := NewMinimalInternetFwRuleModel("rule-123")
	ruleAttrs := model.Rule.Attributes()
	ruleAttrs["device_attributes"] = topLevelDeviceAttributes
	ruleAttrs["exceptions"] = types.SetValueMust(IfwExceptionObjectType, []attr.Value{
		types.ObjectValueMust(IfwExceptionAttrTypes, map[string]attr.Value{
			"name":              types.StringValue("exception-a"),
			"source":            emptyIfwSourceObject(),
			"country":           types.SetNull(idname.RefObjectType),
			"device":            types.SetNull(idname.RefObjectType),
			"device_attributes": exceptionDeviceAttributes,
			"device_os":         types.ListNull(types.StringType),
			"destination":       emptyIfwDestinationObject(),
			"service":           emptyIfwServiceObject(),
			"connection_origin": types.StringNull(),
		}),
	})
	model.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	got := useStateExceptionDeviceAttributes(ctx, model, "exception-a")
	if got.IsNull() || got.IsUnknown() {
		t.Fatal("expected exception device attributes, got null or unknown")
	}

	category, ok := got.Attributes()["category"].(types.List)
	if !ok {
		t.Fatalf("expected category to be types.List, got %T", got.Attributes()["category"])
	}

	var categories []string
	diags := category.ElementsAs(ctx, &categories, false)
	if diags.HasError() {
		t.Fatalf("unexpected category diagnostics: %+v", diags)
	}
	if len(categories) != 1 || categories[0] != "exception-category" {
		t.Fatalf("expected exception category, got %#v", categories)
	}
}

func TestUseStateExceptionDeviceAttributesReturnsNullWithoutMatch(t *testing.T) {
	ctx := context.Background()

	model := NewMinimalInternetFwRuleModel("rule-123")
	ruleAttrs := model.Rule.Attributes()
	ruleAttrs["exceptions"] = types.SetValueMust(IfwExceptionObjectType, []attr.Value{
		types.ObjectValueMust(IfwExceptionAttrTypes, map[string]attr.Value{
			"name":              types.StringValue("exception-a"),
			"source":            emptyIfwSourceObject(),
			"country":           types.SetNull(idname.RefObjectType),
			"device":            types.SetNull(idname.RefObjectType),
			"device_attributes": ifwDeviceAttributesObject([]string{"exception-category"}),
			"device_os":         types.ListNull(types.StringType),
			"destination":       emptyIfwDestinationObject(),
			"service":           emptyIfwServiceObject(),
			"connection_origin": types.StringNull(),
		}),
	})
	model.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	got := useStateExceptionDeviceAttributes(ctx, model, "exception-b")
	if !got.IsNull() {
		t.Fatalf("expected null device attributes for missing exception, got %+v", got)
	}
}

func TestHydrateIfwRuleAPIFailsOnInvalidDestinationApplicationRef(t *testing.T) {
	ctx := context.Background()

	model := NewMinimalInternetFwRuleModel("")
	ruleAttrs := model.Rule.Attributes()
	destinationAttrs := ruleAttrs["destination"].(types.Object).Attributes()
	destinationAttrs["application"] = types.SetValueMust(idname.RefObjectType, []attr.Value{
		types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringNull(),
			"name": types.StringNull(),
		}),
	})
	ruleAttrs["destination"] = types.ObjectValueMust(IfwDestAttrTypes, destinationAttrs)
	model.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	_, diags := HydrateIfwRuleAPI(ctx, model)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for invalid destination application object reference")
	}
}

func TestHydrateIfwRuleAPIFailsOnInvalidSourceHostRef(t *testing.T) {
	ctx := context.Background()

	model := NewMinimalInternetFwRuleModel("")
	ruleAttrs := model.Rule.Attributes()
	sourceAttrs := ruleAttrs["source"].(types.Object).Attributes()
	sourceAttrs["host"] = types.SetValueMust(idname.RefObjectType, []attr.Value{
		types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringNull(),
			"name": types.StringNull(),
		}),
	})
	ruleAttrs["source"] = types.ObjectValueMust(IfwSourceAttrTypes, sourceAttrs)
	model.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	_, diags := HydrateIfwRuleAPI(ctx, model)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for invalid source host object reference")
	}
}

func TestHydrateIfwRuleStateKeepsServiceNullWhenAPIReturnsEmpty(t *testing.T) {
	ctx := context.Background()

	state := NewMinimalInternetFwRuleModel("rule-123")
	ruleAttrs := state.Rule.Attributes()
	ruleAttrs["service"] = types.ObjectNull(IfwServiceAttrTypes)
	state.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	currentRule := MinimalAPIRule("test-ifw-rule", 10)
	hydrated := HydrateIfwRuleState(ctx, state, &currentRule)

	if !hydrated.Service.IsNull() {
		t.Fatalf("expected hydrated service to be null when API returns empty service, got %+v", hydrated.Service)
	}
}

func TestHydrateIfwRuleStatePreservesServiceObjectShapeWhenAPIReturnsEmpty(t *testing.T) {
	ctx := context.Background()

	state := NewMinimalInternetFwRuleModel("rule-123")
	currentRule := MinimalAPIRule("test-ifw-rule", 10)
	hydrated := HydrateIfwRuleState(ctx, state, &currentRule)

	if hydrated.Service.IsNull() || hydrated.Service.IsUnknown() {
		t.Fatalf("expected hydrated service object shape to be preserved, got %+v", hydrated.Service)
	}

	if !hydrated.Service.Equal(emptyIfwServiceObject()) {
		t.Fatalf("expected hydrated service to match prior object shape, got %+v", hydrated.Service)
	}
}

func TestHydrateIfwRuleStatePreservesNullCustomServiceWhenAPIReturnsEmptyCustom(t *testing.T) {
	ctx := context.Background()

	state := NewMinimalInternetFwRuleModel("rule-123")
	ruleAttrs := state.Rule.Attributes()
	ruleAttrs["service"] = types.ObjectValueMust(IfwServiceAttrTypes, map[string]attr.Value{
		"standard": types.SetValueMust(idname.RefObjectType, []attr.Value{
			types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue("svc-1"),
				"name": types.StringValue("SMTP"),
			}),
		}),
		"custom": types.ListNull(customservice.CustomServiceObjectType),
	})
	state.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	currentRule := MinimalAPIRule("test-ifw-rule", 10)
	currentRule.Service = cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service{
		Standard: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard{
			{
				ID:   "svc-1",
				Name: "SMTP",
			},
		},
		Custom: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom{},
	}

	hydrated := HydrateIfwRuleState(ctx, state, &currentRule)
	serviceAttrs := hydrated.Service.Attributes()
	if !serviceAttrs["custom"].(types.List).IsNull() {
		t.Fatalf("expected hydrated service.custom to remain null, got %+v", serviceAttrs["custom"])
	}
}

func TestHydrateIfwRuleStatePreservesNullIDForSourceUsersGroup(t *testing.T) {
	ctx := context.Background()

	state := NewMinimalInternetFwRuleModel("rule-123")
	ruleAttrs := state.Rule.Attributes()
	sourceAttrs := ruleAttrs["source"].(types.Object).Attributes()
	sourceAttrs["users_group"] = types.SetValueMust(idname.RefObjectType, []attr.Value{
		types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringNull(),
			"name": types.StringValue("Corp Users"),
		}),
	})
	ruleAttrs["source"] = types.ObjectValueMust(IfwSourceAttrTypes, sourceAttrs)
	state.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	currentRule := MinimalAPIRule("test-ifw-rule", 10)
	currentRule.Source.UsersGroup = []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup{
		{
			ID:   "ug-123",
			Name: "Corp Users",
		},
	}

	hydrated := HydrateIfwRuleState(ctx, state, &currentRule)
	hydratedUsersGroup := hydrated.Source.Attributes()["users_group"].(types.Set)
	hydratedUserGroupObj := hydratedUsersGroup.Elements()[0].(types.Object)
	hydratedUserGroupID := hydratedUserGroupObj.Attributes()["id"].(types.String)

	if !hydratedUserGroupID.IsNull() {
		t.Fatalf("expected hydrated users_group id to remain null for planned name-only ref, got %+v", hydratedUserGroupID)
	}
}

func TestHydrateIfwRuleStateUsesAPIIDWhenPlannedSourceUsersGroupIDIsUnknown(t *testing.T) {
	ctx := context.Background()

	state := NewMinimalInternetFwRuleModel("rule-123")
	ruleAttrs := state.Rule.Attributes()
	sourceAttrs := ruleAttrs["source"].(types.Object).Attributes()
	sourceAttrs["users_group"] = types.SetValueMust(idname.RefObjectType, []attr.Value{
		types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringUnknown(),
			"name": types.StringValue("Corp Users"),
		}),
	})
	ruleAttrs["source"] = types.ObjectValueMust(IfwSourceAttrTypes, sourceAttrs)
	state.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	currentRule := MinimalAPIRule("test-ifw-rule", 10)
	currentRule.Source.UsersGroup = []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup{
		{
			ID:   "ug-123",
			Name: "Corp Users",
		},
	}

	hydrated := HydrateIfwRuleState(ctx, state, &currentRule)
	hydratedUsersGroup := hydrated.Source.Attributes()["users_group"].(types.Set)
	hydratedUserGroupObj := hydratedUsersGroup.Elements()[0].(types.Object)
	hydratedUserGroupID := hydratedUserGroupObj.Attributes()["id"].(types.String)

	if hydratedUserGroupID.IsUnknown() || hydratedUserGroupID.ValueString() != "ug-123" {
		t.Fatalf("expected hydrated users_group id ug-123 from API, got %+v", hydratedUserGroupID)
	}
}

func TestInternetFwRuleDelete(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallPolicyClient(t)
	resourceState := newInternetFwRuleStateWithID(ctx, t)

	mockClient.EXPECT().
		PolicyInternetFirewallRemoveRule(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()

	r := &InternetFwRuleResource{
		Client:    &catoclient.Service{AccountId: "account-123"},
		IfwClient: mockClient,
	}
	req := resource.DeleteRequest{State: resourceState}
	resp := &resource.DeleteResponse{}

	r.Delete(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestInternetFwRuleCreate(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallPolicyClient(t)

	mockClient.EXPECT().
		PolicyInternetFirewallAddRule(mock.Anything, mock.Anything, "account-123").
		Return(nil, apperror.AssertErr("add failed")).
		Once()

	r := &InternetFwRuleResource{
		Client:    &catoclient.Service{AccountId: "account-123"},
		IfwClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newInternetFwRulePlan(ctx, t, "")}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: GetInternetFwRuleSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for add rule error")
	}
}

func TestInternetFwRuleCreateSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallPolicyClient(t)

	mockClient.EXPECT().
		PolicyInternetFirewallAddRule(mock.Anything, mock.Anything, "account-123").
		Return(SuccessfulAddRuleResponse("rule-123"), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(InternetFirewallPolicyResponseWithRule(MinimalAPIRule("test-ifw-rule", 10)), nil).
		Once()

	r := &InternetFwRuleResource{
		Client:    &catoclient.Service{AccountId: "account-123"},
		IfwClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newInternetFwRulePlan(ctx, t, "")}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: GetInternetFwRuleSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}

	assertRuleState(ctx, t, resp.State, "rule-123", "test-ifw-rule")
}

func TestInternetFwRuleReadRemovesMissingResource(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallPolicyClient(t)
	resourceState := newInternetFwRuleStateWithID(ctx, t)

	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(EmptyInternetFirewallPolicyResponse(), nil).
		Once()

	r := &InternetFwRuleResource{
		Client:    &catoclient.Service{AccountId: "account-123"},
		IfwClient: mockClient,
	}
	req := resource.ReadRequest{State: resourceState}
	resp := &resource.ReadResponse{State: resourceState}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatal("expected state to be removed when rule is missing")
	}
}

func TestInternetFwRuleReadSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallPolicyClient(t)
	resourceState := newInternetFwRuleStateWithID(ctx, t)

	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(InternetFirewallPolicyResponseWithRule(MinimalAPIRule("updated-name", 11)), nil).
		Once()

	r := &InternetFwRuleResource{
		Client:    &catoclient.Service{AccountId: "account-123"},
		IfwClient: mockClient,
	}
	req := resource.ReadRequest{State: resourceState}
	resp := &resource.ReadResponse{State: resourceState}

	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}

	assertRuleState(ctx, t, resp.State, "rule-123", "updated-name")
}

func TestInternetFwRuleUpdate(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallPolicyClient(t)
	resourceState := newInternetFwRuleStateWithID(ctx, t)

	mockClient.EXPECT().
		PolicyInternetFirewallMoveRule(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, apperror.AssertErr("move failed")).
		Once()

	r := &InternetFwRuleResource{
		Client:    &catoclient.Service{AccountId: "account-123"},
		IfwClient: mockClient,
	}
	req := resource.UpdateRequest{
		Plan:  newInternetFwRulePlan(ctx, t, "rule-123"),
		State: resourceState,
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: GetInternetFwRuleSchema(ctx, t)}}

	r.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for move rule error")
	}
}

func TestInternetFwRuleUpdateSuccess(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewInternetFirewallPolicyClient(t)
	resourceState := newInternetFwRuleStateWithID(ctx, t)

	mockClient.EXPECT().
		PolicyInternetFirewallMoveRule(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(successfulMoveRuleResponse("rule-123"), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallUpdateRule(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(SuccessfulUpdateRuleResponse("rule-123"), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(InternetFirewallPolicyResponseWithRule(MinimalAPIRule("test-ifw-rule", 12)), nil).
		Once()

	r := &InternetFwRuleResource{
		Client:    &catoclient.Service{AccountId: "account-123"},
		IfwClient: mockClient,
	}
	req := resource.UpdateRequest{
		Plan:  newInternetFwRulePlan(ctx, t, "rule-123"),
		State: resourceState,
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: GetInternetFwRuleSchema(ctx, t)}}

	r.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}

	assertRuleState(ctx, t, resp.State, "rule-123", "test-ifw-rule")
}

func TestInternetFwRuleExceptionsPlanModifiersPreserveEmptyStateSet(t *testing.T) {
	ctx := context.Background()
	exceptionsAttr := getInternetFwRuleExceptionsAttribute(ctx, t)

	if got := len(exceptionsAttr.PlanModifiers); got != 1 {
		t.Fatalf("expected 1 exception plan modifier, got %d", got)
	}

	exceptionObjectType := types.ObjectType{AttrTypes: IfwExceptionAttrTypes}
	emptyState := types.SetValueMust(exceptionObjectType, []attr.Value{})

	req := planmodifier.SetRequest{
		ConfigValue: types.SetNull(exceptionObjectType),
		PlanValue:   types.SetUnknown(exceptionObjectType),
		State:       tfsdk.State{Raw: tftypes.NewValue(tftypes.Bool, true)},
		StateValue:  emptyState,
	}

	for _, modifier := range exceptionsAttr.PlanModifiers {
		resp := &planmodifier.SetResponse{PlanValue: req.PlanValue}
		modifier.PlanModifySet(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
		}
		if !resp.PlanValue.IsNull() && !resp.PlanValue.IsUnknown() {
			req.PlanValue = resp.PlanValue
		}
	}

	if req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		t.Fatalf("expected known empty set after plan modifiers, got %v", req.PlanValue)
	}
	if !req.PlanValue.Equal(emptyState) {
		t.Fatalf("expected plan value to equal empty state set, got %v", req.PlanValue)
	}
}

func newInternetFwRulePlan(ctx context.Context, t *testing.T, ruleID string) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{Schema: GetInternetFwRuleSchema(ctx, t)}
	diags := plan.Set(ctx, NewMinimalInternetFwRuleModel(ruleID))
	if diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}

	return plan
}

func newInternetFwRuleStateWithID(ctx context.Context, t *testing.T) tfsdk.State {
	t.Helper()

	state := tfsdk.State{Schema: GetInternetFwRuleSchema(ctx, t)}
	diags := state.Set(ctx, NewMinimalInternetFwRuleModel("rule-123"))
	if diags.HasError() {
		t.Fatalf("unexpected seed state diagnostics: %+v", diags)
	}

	return state
}

func ifwDeviceAttributesObject(categories []string) types.Object {
	categoryValues := make([]attr.Value, 0, len(categories))
	for _, category := range categories {
		categoryValues = append(categoryValues, types.StringValue(category))
	}

	return types.ObjectValueMust(IfwDeviceAttrAttrTypes, map[string]attr.Value{
		"category":     types.ListValueMust(types.StringType, categoryValues),
		"type":         types.ListNull(types.StringType),
		"model":        types.ListNull(types.StringType),
		"manufacturer": types.ListNull(types.StringType),
		"os":           types.ListNull(types.StringType),
		"os_version":   types.ListNull(types.StringType),
	})
}

func successfulMoveRuleResponse(ruleID string) *cato_go_sdk.PolicyInternetFirewallMoveRule {
	return &cato_go_sdk.PolicyInternetFirewallMoveRule{
		Policy: cato_go_sdk.PolicyInternetFirewallMoveRule_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallMoveRule_Policy_InternetFirewall{
				MoveRule: cato_go_sdk.PolicyInternetFirewallMoveRule_Policy_InternetFirewall_MoveRule{
					Status: "SUCCESS",
					Rule: &cato_go_sdk.PolicyInternetFirewallMoveRule_Policy_InternetFirewall_MoveRule_Rule{
						Rule: cato_go_sdk.PolicyInternetFirewallMoveRule_Policy_InternetFirewall_MoveRule_Rule_Rule{
							ID: ruleID,
						},
					},
				},
			},
		},
	}
}

func assertRuleState(ctx context.Context, t *testing.T, state tfsdk.State, wantID, wantName string) {
	t.Helper()

	var model InternetFirewallRule
	diags := state.Get(ctx, &model)
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}

	rule := PolicyPolicyInternetFirewallPolicyRulesRule{}
	diags = model.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("unexpected rule diagnostics: %+v", diags)
	}

	if got := rule.ID.ValueString(); got != wantID {
		t.Fatalf("expected rule id %q, got %q", wantID, got)
	}
	if got := rule.Name.ValueString(); got != wantName {
		t.Fatalf("expected rule name %q, got %q", wantName, got)
	}
}

func getInternetFwRuleExceptionsAttribute(ctx context.Context, t *testing.T) schema.SetNestedAttribute {
	t.Helper()

	ruleAttr := getInternetFwRuleRuleAttribute(ctx, t)
	exceptionsAttr, ok := ruleAttr.Attributes["exceptions"].(schema.SetNestedAttribute)
	if !ok {
		t.Fatalf("rule.exceptions attribute is not a SetNestedAttribute")
	}

	return exceptionsAttr
}

func getInternetFwRuleRuleAttribute(ctx context.Context, t *testing.T) schema.SingleNestedAttribute {
	t.Helper()

	s := GetInternetFwRuleSchema(ctx, t)
	ruleAttr, ok := s.Attributes["rule"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("rule attribute is not a SingleNestedAttribute")
	}

	return ruleAttr
}

// Exercise parent, nested-object, and field modifiers in framework order. Set
// members intentionally have different configuration and state ordering.
func TestInternetFwRuleUserGroupPlanPreservesIDsWhenReordered(t *testing.T) {
	ctx := context.Background()
	rule := GetInternetFwRuleSchema(ctx, t).Attributes["rule"].(schema.SingleNestedAttribute)
	source := rule.Attributes["source"].(schema.SingleNestedAttribute)
	groups := source.Attributes["users_group"].(schema.SetNestedAttribute)
	ref := func(name string, id types.String) types.Object {
		return types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"name": types.StringValue(name), "id": id})
	}
	stateElements := []attr.Value{ref("group-a", types.StringValue("id-a")), ref("group-b", types.StringValue("id-b"))}
	configElements := []attr.Value{ref("group-b", types.StringNull()), ref("group-a", types.StringNull())}
	planElements := []attr.Value{ref("group-b", types.StringUnknown()), ref("group-a", types.StringUnknown())}
	makeSource := func(elements []attr.Value) types.Object {
		attrs := emptyIfwSourceObject().Attributes()
		attrs["users_group"] = types.SetValueMust(idname.RefObjectType, elements)
		return types.ObjectValueMust(IfwSourceAttrTypes, attrs)
	}
	config, state, plan := makeSource(configElements), makeSource(stateElements), makeSource(planElements)
	sourcePath := path.Root("rule").AtName("source")
	for _, modifier := range source.PlanModifiers {
		resp := planmodifier.ObjectResponse{PlanValue: plan}
		modifier.PlanModifyObject(ctx, planmodifier.ObjectRequest{Path: sourcePath, ConfigValue: config, StateValue: state, PlanValue: plan}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("source diagnostics: %v", resp.Diagnostics)
		}
		plan = resp.PlanValue
	}
	plannedGroups := plan.Attributes()["users_group"].(types.Set)
	result := make([]attr.Value, 0, 2)
	for i, elem := range plannedGroups.Elements() {
		plannedRef := elem.(types.Object)
		configRef, stateRef := configElements[i].(types.Object), stateElements[i].(types.Object)
		refPath := sourcePath.AtName("users_group").AtSetValue(elem)
		for _, modifier := range groups.NestedObject.PlanModifiers {
			resp := planmodifier.ObjectResponse{PlanValue: plannedRef}
			modifier.PlanModifyObject(ctx, planmodifier.ObjectRequest{Path: refPath, ConfigValue: configRef, StateValue: stateRef, PlanValue: plannedRef}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("reference diagnostics: %v", resp.Diagnostics)
			}
			plannedRef = resp.PlanValue
		}
		attrs := plannedRef.Attributes()
		for _, field := range []string{"name", "id"} {
			attribute := groups.NestedObject.Attributes[field].(schema.StringAttribute)
			for _, modifier := range attribute.PlanModifiers {
				resp := planmodifier.StringResponse{PlanValue: attrs[field].(types.String)}
				modifier.PlanModifyString(ctx, planmodifier.StringRequest{Path: refPath.AtName(field), ConfigValue: configRef.Attributes()[field].(types.String), StateValue: stateRef.Attributes()[field].(types.String), PlanValue: resp.PlanValue}, &resp)
				if resp.Diagnostics.HasError() {
					t.Fatalf("field diagnostics: %v", resp.Diagnostics)
				}
				attrs[field] = resp.PlanValue
			}
		}
		result = append(result, types.ObjectValueMust(idname.RefAttrTypes, attrs))
	}
	got := types.SetValueMust(idname.RefObjectType, result)
	want := types.SetValueMust(idname.RefObjectType, stateElements)
	if !got.Equal(want) {
		t.Fatalf("unchanged groups must retain their own IDs: got %s, want %s", got, want)
	}
}

func TestInternetFwRuleImportStateComposite(t *testing.T) {
	ctx := context.Background()
	r := &InternetFwRuleResource{}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: GetInternetFwRuleSchema(ctx, t)},
	}
	diags := resp.State.Set(ctx, InternetFirewallRule{
		Rule: types.ObjectNull(InternetFirewallRuleRuleAttrTypes),
		At:   types.ObjectNull(policyrule.PositionAttrTypes),
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
	r := &InternetFwRuleResource{}
	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: GetInternetFwRuleSchema(ctx, t)}}
	_ = resp.State.Set(ctx, InternetFirewallRule{
		Rule: types.ObjectNull(InternetFirewallRuleRuleAttrTypes),
		At:   types.ObjectNull(policyrule.PositionAttrTypes),
	})

	r.ImportState(ctx, resource.ImportStateRequest{ID: "sub-42/"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for malformed composite import id")
	}
}
