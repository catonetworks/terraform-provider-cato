package provider

import (
	"context"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestNewApplicationControlRuleResource(t *testing.T) {
	t.Parallel()
	r := NewApplicationControlRuleResource()
	if r == nil {
		t.Fatal("expected resource instance, got nil")
	}
	if _, ok := r.(*applicationControlRuleResource); !ok {
		t.Fatalf("expected *applicationControlRuleResource, got %T", r)
	}
}

func TestApplicationControlRuleMetadata(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &applicationControlRuleResource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, resp)
	if resp.TypeName != "cato_application_control_rule" {
		t.Fatalf("expected type name cato_application_control_rule, got %q", resp.TypeName)
	}
}

func TestApplicationControlRuleConfigureNilProviderData(t *testing.T) {
	t.Parallel()
	r := &applicationControlRuleResource{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{}, resp)
	if r.client != nil {
		t.Fatal("expected client to remain nil when provider data is nil")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestApplicationControlRuleConfigureSetsClient(t *testing.T) {
	t.Parallel()
	client := &catoClientData{AccountId: "123"}
	r := &applicationControlRuleResource{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, resp)
	if r.client != client {
		t.Fatal("expected resource client to be set from provider data")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestApplicationControlRuleImportState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &applicationControlRuleResource{}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: getApplicationControlRuleSchema(ctx, t)},
	}
	diags := resp.State.Set(ctx, ApplicationControlRule{
		Rule: types.ObjectNull(ApplicationControlRuleRuleAttrTypes),
		At:   types.ObjectNull(PositionAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("unexpected seed state diagnostics: %+v", diags)
	}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "rule-123"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var imported ApplicationControlRule
	diags = resp.State.Get(ctx, &imported)
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	ruleModel := ApplicationControlRuleRulePlan{}
	diags = imported.Rule.As(ctx, &ruleModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("unexpected rule diagnostics: %+v", diags)
	}
	if ruleModel.ID.ValueString() != "rule-123" {
		t.Fatalf("expected imported rule id rule-123, got %q", ruleModel.ID.ValueString())
	}
}

func TestHydrateApplicationControlAddRuleInputDefaultsFileCriteria(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	plan := newMinimalApplicationControlRulePlan(cato_models.ApplicationControlRuleTypeFile, "file_rule")
	got, diags := hydrateApplicationControlAddRuleInput(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	fileRule := got.Rule.FileRule
	if fileRule == nil {
		t.Fatal("expected file rule payload")
	}
	if fileRule.ApplicationCriteria == nil {
		t.Fatal("expected non-nil file rule application criteria")
	}
	if fileRule.ApplicationContext == nil {
		t.Fatal("expected non-nil file rule application context")
	}
	if fileRule.ApplicationCriteriaSatisfy != cato_models.ApplicationControlSatisfyAll {
		t.Fatalf("expected file rule application criteria satisfy ALL, got %q", fileRule.ApplicationCriteriaSatisfy)
	}
}

func TestHydrateApplicationControlAddRuleInputDefaultsDataCriteria(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	plan := newMinimalApplicationControlRulePlan(cato_models.ApplicationControlRuleTypeData, "data_rule")
	dataRuleAttrs := plan.Rule.Attributes()["data_rule"].(types.Object).Attributes()
	dataRuleAttrs["dlp_profile"] = types.ObjectValueMust(applicationControlDlpProfileAttrTypes, map[string]attr.Value{
		"content_profile": types.SetValueMust(NameIDObjectType, []attr.Value{
			types.ObjectValueMust(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue("content-profile-1"),
				"name": types.StringNull(),
			}),
		}),
		"edm_profile": types.SetNull(NameIDObjectType),
	})
	ruleAttrs := plan.Rule.Attributes()
	ruleAttrs["data_rule"] = types.ObjectValueMust(applicationControlTypedRuleAttrTypes, dataRuleAttrs)
	plan.Rule = types.ObjectValueMust(ApplicationControlRuleRuleAttrTypes, ruleAttrs)

	got, diags := hydrateApplicationControlAddRuleInput(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	dataRule := got.Rule.DataRule
	if dataRule == nil {
		t.Fatal("expected data rule payload")
	}
	if dataRule.ApplicationCriteria == nil {
		t.Fatal("expected non-nil data rule application criteria")
	}
	if dataRule.ApplicationContext == nil {
		t.Fatal("expected non-nil data rule application context")
	}
	if dataRule.ApplicationCriteriaSatisfy != cato_models.ApplicationControlSatisfyAll {
		t.Fatalf("expected data rule application criteria satisfy ALL, got %q", dataRule.ApplicationCriteriaSatisfy)
	}
}

func getApplicationControlRuleSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()
	r := &applicationControlRuleResource{}
	sr := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, sr)
	return sr.Schema
}

func newMinimalApplicationControlRulePlan(
	ruleType cato_models.ApplicationControlRuleType,
	typedRuleAttribute string,
) ApplicationControlRule {
	typedRule := types.ObjectValueMust(applicationControlTypedRuleAttrTypes, map[string]attr.Value{
		"action":                       types.StringValue(string(cato_models.ApplicationControlActionMonitor)),
		"severity":                     types.StringValue(string(cato_models.ApplicationControlSeverityMedium)),
		"schedule":                     types.ObjectNull(ScheduleAttrTypes),
		"source":                       types.ObjectNull(ApplicationControlSourceAttrTypes),
		"tracking":                     types.ObjectNull(TrackingAttrTypes),
		"device":                       types.SetNull(NameIDObjectType),
		"access_method":                types.ListNull(types.ObjectType{AttrTypes: ApplicationControlAccessMethodAttrTypes}),
		"application":                  types.ObjectNull(WanApplicationAttrTypes),
		"application_activity":         types.ListNull(types.ObjectType{AttrTypes: applicationControlActivityAttrTypes}),
		"application_activity_satisfy": types.StringNull(),
		"action_config":                types.ObjectNull(applicationControlActionConfigAttrTypes),
		"file_attribute":               types.ListNull(types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}),
		"file_attribute_satisfy":       types.StringNull(),
		"dlp_profile":                  types.ObjectNull(applicationControlDlpProfileAttrTypes),
	})

	ruleAttrs := map[string]attr.Value{
		"id":               types.StringNull(),
		"name":             types.StringValue("test-rule"),
		"description":      types.StringValue(""),
		"enabled":          types.BoolValue(true),
		"rule_type":        types.StringValue(string(ruleType)),
		"application_rule": types.ObjectNull(applicationControlTypedRuleAttrTypes),
		"data_rule":        types.ObjectNull(applicationControlTypedRuleAttrTypes),
		"file_rule":        types.ObjectNull(applicationControlTypedRuleAttrTypes),
	}
	ruleAttrs[typedRuleAttribute] = typedRule

	return ApplicationControlRule{
		Rule: types.ObjectValueMust(ApplicationControlRuleRuleAttrTypes, ruleAttrs),
		At: types.ObjectValueMust(PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		}),
	}
}
