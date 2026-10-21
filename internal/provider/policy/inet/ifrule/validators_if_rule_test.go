package ifrule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

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
	objectType := types.ObjectType{AttrTypes: policyrule.TrackingAlertAttrTypes}

	testCases := []struct {
		name  string
		value types.Object
	}{
		{name: "null", value: types.ObjectNull(policyrule.TrackingAlertAttrTypes)},
		{name: "unknown", value: types.ObjectUnknown(policyrule.TrackingAlertAttrTypes)},
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
			policyrule.TrackingAlertAttrTypes,
			map[string]attr.Value{
				"enabled":            types.BoolNull(),
				"frequency":          types.StringNull(),
				"subscription_group": types.SetNull(idname.RefObjectType),
				"webhook":            types.SetNull(idname.RefObjectType),
				"mailing_list":       types.SetNull(idname.RefObjectType),
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
			policyrule.TrackingAlertAttrTypes,
			map[string]attr.Value{
				"enabled":            types.BoolValue(true),
				"frequency":          types.StringValue("DAILY"),
				"subscription_group": types.SetNull(idname.RefObjectType),
				"webhook":            types.SetNull(idname.RefObjectType),
				"mailing_list":       types.SetNull(idname.RefObjectType),
			},
		),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestInternetFirewallExceptionModifierResolvesUnknownCountryID(t *testing.T) {
	t.Parallel()
	ref := func(id types.String) types.Object {
		return types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"name": types.StringValue("country-a"), "id": id})
	}
	exceptionTypes := map[string]attr.Type{"country": types.SetType{ElemType: idname.RefObjectType}}
	exception := func(id types.String) types.Set {
		country := types.SetValueMust(idname.RefObjectType, []attr.Value{ref(id)})
		object := types.ObjectValueMust(exceptionTypes, map[string]attr.Value{"country": country})
		return types.SetValueMust(types.ObjectType{AttrTypes: exceptionTypes}, []attr.Value{object})
	}
	config, plan := exception(types.StringNull()), exception(types.StringUnknown())
	resp := &planmodifier.SetResponse{PlanValue: plan}
	IfwExceptionsSetModifier().PlanModifySet(context.Background(), planmodifier.SetRequest{
		ConfigValue: config, PlanValue: plan, StateValue: types.SetNull(types.ObjectType{AttrTypes: exceptionTypes}),
	}, resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.True(t, config.Equal(resp.PlanValue), "expected configured name to be retained and unknown ID resolved to null")
}
