package ifrule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
