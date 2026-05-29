package provider

import (
	"context"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

func TestNewInternetFwRuleResource(t *testing.T) {
	r := NewInternetFwRuleResource()

	if r == nil {
		t.Fatal("expected resource instance, got nil")
	}

	if _, ok := r.(*internetFwRuleResource); !ok {
		t.Fatalf("expected *internetFwRuleResource, got %T", r)
	}
}

func TestInternetFwRuleMetadata(t *testing.T) {
	ctx := context.Background()
	r := &internetFwRuleResource{}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, resp)

	if resp.TypeName != "cato_if_rule" {
		t.Fatalf("expected type name cato_if_rule, got %q", resp.TypeName)
	}
}

func TestInternetFwRuleConfigureNilProviderData(t *testing.T) {
	r := &internetFwRuleResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{}, resp)

	if r.client != nil {
		t.Fatal("expected client to remain nil when provider data is nil")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestInternetFwRuleConfigureSetsClient(t *testing.T) {
	client := &catoClientData{AccountId: "123"}
	r := &internetFwRuleResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, resp)

	if r.client != client {
		t.Fatal("expected resource client to be set from provider data")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestInternetFwRuleImportState(t *testing.T) {
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

	model := newMinimalInternetFwRuleModel("rule-123")
	ruleAttrs := model.Rule.Attributes()
	ruleAttrs["device_attributes"] = topLevelDeviceAttributes
	ruleAttrs["exceptions"] = types.SetValueMust(IfwExceptionObjectType, []attr.Value{
		types.ObjectValueMust(IfwExceptionAttrTypes, map[string]attr.Value{
			"name":              types.StringValue("exception-a"),
			"source":            emptyIfwSourceObject(),
			"country":           types.SetNull(NameIDObjectType),
			"device":            types.SetNull(NameIDObjectType),
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

	model := newMinimalInternetFwRuleModel("rule-123")
	ruleAttrs := model.Rule.Attributes()
	ruleAttrs["exceptions"] = types.SetValueMust(IfwExceptionObjectType, []attr.Value{
		types.ObjectValueMust(IfwExceptionAttrTypes, map[string]attr.Value{
			"name":              types.StringValue("exception-a"),
			"source":            emptyIfwSourceObject(),
			"country":           types.SetNull(NameIDObjectType),
			"device":            types.SetNull(NameIDObjectType),
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

	model := newMinimalInternetFwRuleModel("")
	ruleAttrs := model.Rule.Attributes()
	destinationAttrs := ruleAttrs["destination"].(types.Object).Attributes()
	destinationAttrs["application"] = types.SetValueMust(NameIDObjectType, []attr.Value{
		types.ObjectValueMust(NameIDAttrTypes, map[string]attr.Value{
			"id":   types.StringNull(),
			"name": types.StringNull(),
		}),
	})
	ruleAttrs["destination"] = types.ObjectValueMust(IfwDestAttrTypes, destinationAttrs)
	model.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	_, diags := hydrateIfwRuleAPI(ctx, model)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for invalid destination application object reference")
	}
}

func TestHydrateIfwRuleAPIFailsOnInvalidSourceHostRef(t *testing.T) {
	ctx := context.Background()

	model := newMinimalInternetFwRuleModel("")
	ruleAttrs := model.Rule.Attributes()
	sourceAttrs := ruleAttrs["source"].(types.Object).Attributes()
	sourceAttrs["host"] = types.SetValueMust(NameIDObjectType, []attr.Value{
		types.ObjectValueMust(NameIDAttrTypes, map[string]attr.Value{
			"id":   types.StringNull(),
			"name": types.StringNull(),
		}),
	})
	ruleAttrs["source"] = types.ObjectValueMust(IfwSourceAttrTypes, sourceAttrs)
	model.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	_, diags := hydrateIfwRuleAPI(ctx, model)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for invalid source host object reference")
	}
}

func TestHydrateIfwRuleStateKeepsServiceNullWhenAPIReturnsEmpty(t *testing.T) {
	ctx := context.Background()

	state := newMinimalInternetFwRuleModel("rule-123")
	ruleAttrs := state.Rule.Attributes()
	ruleAttrs["service"] = types.ObjectNull(IfwServiceAttrTypes)
	state.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs)

	currentRule := minimalAPIRule("test-ifw-rule", 10)
	hydrated := hydrateIfwRuleState(ctx, state, &currentRule)

	if !hydrated.Service.IsNull() {
		t.Fatalf("expected hydrated service to be null when API returns empty service, got %+v", hydrated.Service)
	}
}

func TestHydrateIfwRuleStatePreservesServiceObjectShapeWhenAPIReturnsEmpty(t *testing.T) {
	ctx := context.Background()

	state := newMinimalInternetFwRuleModel("rule-123")
	currentRule := minimalAPIRule("test-ifw-rule", 10)
	hydrated := hydrateIfwRuleState(ctx, state, &currentRule)

	if hydrated.Service.IsNull() || hydrated.Service.IsUnknown() {
		t.Fatalf("expected hydrated service object shape to be preserved, got %+v", hydrated.Service)
	}

	if !hydrated.Service.Equal(emptyIfwServiceObject()) {
		t.Fatalf("expected hydrated service to match prior object shape, got %+v", hydrated.Service)
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

	r := &internetFwRuleResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
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
		Return(nil, assertErr("add failed")).
		Once()

	r := &internetFwRuleResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newInternetFwRulePlan(ctx, t, "")}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getInternetFwRuleSchema(ctx, t)}}

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
		Return(successfulAddRuleResponse("rule-123"), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(internetFirewallPolicyResponseWithRule(minimalAPIRule("test-ifw-rule", 10)), nil).
		Once()

	r := &internetFwRuleResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newInternetFwRulePlan(ctx, t, "")}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getInternetFwRuleSchema(ctx, t)}}

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
		Return(emptyInternetFirewallPolicyResponse(), nil).
		Once()

	r := &internetFwRuleResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
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
		Return(internetFirewallPolicyResponseWithRule(minimalAPIRule("updated-name", 11)), nil).
		Once()

	r := &internetFwRuleResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
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
		Return(nil, assertErr("move failed")).
		Once()

	r := &internetFwRuleResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
	}
	req := resource.UpdateRequest{
		Plan:  newInternetFwRulePlan(ctx, t, "rule-123"),
		State: resourceState,
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: getInternetFwRuleSchema(ctx, t)}}

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
		Return(successfulUpdateRuleResponse("rule-123"), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").
		Return(internetFirewallPolicyResponseWithRule(minimalAPIRule("test-ifw-rule", 12)), nil).
		Once()

	r := &internetFwRuleResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
	}
	req := resource.UpdateRequest{
		Plan:  newInternetFwRulePlan(ctx, t, "rule-123"),
		State: resourceState,
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: getInternetFwRuleSchema(ctx, t)}}

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

func TestRuleAlertValidatorDescription(t *testing.T) {
	v := ruleAlertValidator{}
	got := v.Description(context.Background())

	if got == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestRuleAlertValidatorMarkdownDescription(t *testing.T) {
	ctx := context.Background()
	v := ruleAlertValidator{}

	if got, want := v.MarkdownDescription(ctx), v.Description(ctx); got != want {
		t.Fatalf("expected markdown description to match description\nwant: %q\ngot:  %q", want, got)
	}
}

func TestRuleAlertValidatorValidateObjectSkipsNullAndUnknown(t *testing.T) {
	ctx := context.Background()
	v := ruleAlertValidator{}
	objectType := types.ObjectType{AttrTypes: TrackingAlertAttrTypes}

	testCases := []struct {
		name  string
		value types.Object
	}{
		{name: "null", value: types.ObjectNull(TrackingAlertAttrTypes)},
		{name: "unknown", value: types.ObjectUnknown(TrackingAlertAttrTypes)},
		{name: "unknown_configured_type", value: types.ObjectUnknown(objectType.AttrTypes)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.ObjectResponse{}
			v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: tc.value}, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("expected no diagnostics, got %+v", resp.Diagnostics)
			}
		})
	}
}

func TestRuleAlertValidatorValidateObjectRequiresEnabledAndFrequency(t *testing.T) {
	ctx := context.Background()
	v := ruleAlertValidator{}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, validator.ObjectRequest{
		ConfigValue: types.ObjectValueMust(
			TrackingAlertAttrTypes,
			map[string]attr.Value{
				"enabled":            types.BoolNull(),
				"frequency":          types.StringNull(),
				"subscription_group": types.SetNull(NameIDObjectType),
				"webhook":            types.SetNull(NameIDObjectType),
				"mailing_list":       types.SetNull(NameIDObjectType),
			},
		),
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for missing enabled and frequency")
	}
	if got := len(resp.Diagnostics.Errors()); got != 2 {
		t.Fatalf("expected 2 validation errors, got %d", got)
	}
}

func TestRuleAlertValidatorValidateObjectAcceptsValidAlert(t *testing.T) {
	ctx := context.Background()
	v := ruleAlertValidator{}
	resp := &validator.ObjectResponse{}

	v.ValidateObject(ctx, validator.ObjectRequest{
		ConfigValue: types.ObjectValueMust(
			TrackingAlertAttrTypes,
			map[string]attr.Value{
				"enabled":            types.BoolValue(true),
				"frequency":          types.StringValue("DAILY"),
				"subscription_group": types.SetNull(NameIDObjectType),
				"webhook":            types.SetNull(NameIDObjectType),
				"mailing_list":       types.SetNull(NameIDObjectType),
			},
		),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func getInternetFwRuleSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &internetFwRuleResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	return resp.Schema
}

func newInternetFwRulePlan(ctx context.Context, t *testing.T, ruleID string) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{Schema: getInternetFwRuleSchema(ctx, t)}
	diags := plan.Set(ctx, newMinimalInternetFwRuleModel(ruleID))
	if diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}

	return plan
}

func newMinimalInternetFwRuleModel(ruleID string) InternetFirewallRule {
	ruleAttrs := map[string]attr.Value{
		"id":                nullableString(ruleID),
		"name":              types.StringValue("test-ifw-rule"),
		"description":       types.StringNull(),
		"enabled":           types.BoolValue(true),
		"source":            emptyIfwSourceObject(),
		"connection_origin": types.StringNull(),
		"active_period":     types.ObjectNull(ActivePeriodAttrTypes),
		"country":           types.SetNull(NameIDObjectType),
		"device":            types.SetNull(NameIDObjectType),
		"device_attributes": types.ObjectNull(IfwDeviceAttrAttrTypes),
		"device_os":         types.ListNull(types.StringType),
		"destination":       emptyIfwDestinationObject(),
		"service":           emptyIfwServiceObject(),
		"action":            types.StringValue("ALLOW"),
		"tracking":          minimalTrackingObject(),
		"schedule":          types.ObjectNull(ScheduleAttrTypes),
		"exceptions":        types.SetNull(IfwExceptionObjectType),
	}

	return InternetFirewallRule{
		Rule: types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs),
		At: types.ObjectValueMust(PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		}),
	}
}

func newInternetFwRuleStateWithID(ctx context.Context, t *testing.T) tfsdk.State {
	t.Helper()

	state := tfsdk.State{Schema: getInternetFwRuleSchema(ctx, t)}
	diags := state.Set(ctx, newMinimalInternetFwRuleModel("rule-123"))
	if diags.HasError() {
		t.Fatalf("unexpected seed state diagnostics: %+v", diags)
	}

	return state
}

func emptyIfwSourceObject() types.Object {
	return types.ObjectValueMust(IfwSourceAttrTypes, map[string]attr.Value{
		"ip":                  types.ListNull(types.StringType),
		"host":                types.SetNull(NameIDObjectType),
		"site":                types.SetNull(NameIDObjectType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(FromToObjectType),
		"global_ip_range":     types.SetNull(NameIDObjectType),
		"network_interface":   types.SetNull(NameIDObjectType),
		"site_network_subnet": types.SetNull(NameIDObjectType),
		"floating_subnet":     types.SetNull(NameIDObjectType),
		"user":                types.SetNull(NameIDObjectType),
		"users_group":         types.SetNull(NameIDObjectType),
		"group":               types.SetNull(NameIDObjectType),
		"system_group":        types.SetNull(NameIDObjectType),
	})
}

func emptyIfwDestinationObject() types.Object {
	return types.ObjectValueMust(IfwDestAttrTypes, map[string]attr.Value{
		"application":              types.SetNull(NameIDObjectType),
		"custom_app":               types.SetNull(NameIDObjectType),
		"app_category":             types.SetNull(NameIDObjectType),
		"custom_category":          types.SetNull(NameIDObjectType),
		"sanctioned_apps_category": types.SetNull(NameIDObjectType),
		"country":                  types.SetNull(NameIDObjectType),
		"domain":                   types.ListNull(types.StringType),
		"fqdn":                     types.ListNull(types.StringType),
		"ip":                       types.ListNull(types.StringType),
		"subnet":                   types.ListNull(types.StringType),
		"ip_range":                 types.ListNull(FromToObjectType),
		"global_ip_range":          types.SetNull(NameIDObjectType),
		"remote_asn":               types.ListNull(types.StringType),
	})
}

func emptyIfwServiceObject() types.Object {
	return types.ObjectValueMust(IfwServiceAttrTypes, map[string]attr.Value{
		"standard": types.SetNull(NameIDObjectType),
		"custom":   types.ListNull(CustomServiceObjectType),
	})
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

func minimalTrackingObject() types.Object {
	return types.ObjectValueMust(TrackingAttrTypes, map[string]attr.Value{
		"event": types.ObjectValueMust(TrackingEventAttrTypes, map[string]attr.Value{
			"enabled": types.BoolValue(false),
		}),
		"alert": types.ObjectNull(TrackingAlertAttrTypes),
	})
}

func emptyInternetFirewallPolicyResponse() *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: &cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{
				Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{},
				},
			},
		},
	}
}

func internetFirewallPolicyResponseWithRule(rule cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule) *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: &cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{
				Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{
						{
							Rule: rule,
						},
					},
				},
			},
		},
	}
}

func minimalAPIRule(name string, index int64) cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule {
	return cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{
		ID:               "rule-123",
		Name:             name,
		Index:            index,
		Enabled:          true,
		Action:           cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule_Rule_Rule{}.Action,
		ConnectionOrigin: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule_Rule_Rule{}.ConnectionOrigin,
	}
}

func successfulAddRuleResponse(ruleID string) *cato_go_sdk.PolicyInternetFirewallAddRule {
	return &cato_go_sdk.PolicyInternetFirewallAddRule{
		Policy: &cato_go_sdk.PolicyInternetFirewallAddRule_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall{
				AddRule: cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule{
					Status: "SUCCESS",
					Rule: &cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule_Rule{
						Rule: cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule_Rule_Rule{
							ID: ruleID,
						},
					},
				},
			},
		},
	}
}

func successfulMoveRuleResponse(ruleID string) *cato_go_sdk.PolicyInternetFirewallMoveRule {
	return &cato_go_sdk.PolicyInternetFirewallMoveRule{
		Policy: &cato_go_sdk.PolicyInternetFirewallMoveRule_Policy{
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

func successfulUpdateRuleResponse(ruleID string) *cato_go_sdk.PolicyInternetFirewallUpdateRule {
	return &cato_go_sdk.PolicyInternetFirewallUpdateRule{
		Policy: &cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall{
				UpdateRule: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule{
					Status: "SUCCESS",
					Rule: &cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule_Rule{
						Rule: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule_Rule_Rule{
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

func nullableString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}

func assertErr(message string) error {
	return &testError{message: message}
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
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

	s := getInternetFwRuleSchema(ctx, t)
	ruleAttr, ok := s.Attributes["rule"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("rule attribute is not a SingleNestedAttribute")
	}

	return ruleAttr
}
