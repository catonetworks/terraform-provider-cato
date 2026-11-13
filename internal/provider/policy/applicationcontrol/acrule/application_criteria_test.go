package acrule

// Regression tests for the bug classes found during refarch→dev replication
// (2026-11-10): null application/source into non-nullable inputs, the
// application_type/application_criteria surfaces, activity granular
// conditions, and CONTENT_TYPE file-attribute refs.

import (
	"context"
	"strings"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	iprange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
)

func nullApplicationObject() types.Object {
	return types.ObjectValueMust(acApplicationAttrTypes, map[string]attr.Value{
		"application":              types.SetNull(idname.RefObjectType),
		"custom_app":               types.SetNull(idname.RefObjectType),
		"app_category":             types.SetNull(idname.RefObjectType),
		"custom_category":          types.SetNull(idname.RefObjectType),
		"sanctioned_apps_category": types.SetNull(idname.RefObjectType),
		"domain":                   types.ListNull(types.StringType),
		"fqdn":                     types.ListNull(types.StringType),
		"ip":                       types.ListNull(types.StringType),
		"subnet":                   types.ListNull(types.StringType),
		"ip_range":                 types.ListNull(iprange.RangeObjectType),
		"global_ip_range":          types.SetNull(idname.RefObjectType),
		"application_type":         types.ListNull(types.StringType),
	})
}

func setTypedRuleAttr(t *testing.T, plan *ApplicationControlRule, typedRuleAttribute, key string, val attr.Value) {
	t.Helper()
	ruleAttrs := plan.Rule.Attributes()
	typedAttrs := ruleAttrs[typedRuleAttribute].(types.Object).Attributes()
	typedAttrs[key] = val
	ruleAttrs[typedRuleAttribute] = types.ObjectValueMust(applicationControlTypedRuleAttrTypes, typedAttrs)
	plan.Rule = types.ObjectValueMust(ApplicationControlRuleRuleAttrTypes, ruleAttrs)
}

// The API rejects a null source (`ApplicationControlSourceInput!`); an omitted
// source block must be sent as an empty object, never null.
func TestHydrateAddDefaultsEmptySource(t *testing.T) {
	t.Parallel()
	plan := newMinimalApplicationControlRulePlan(cato_models.ApplicationControlRuleTypeApplication, "application_rule")
	got, diags := hydrateApplicationControlAddRuleInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if got.Rule.ApplicationRule.Source == nil {
		t.Fatal("expected non-nil source input for an omitted source block")
	}
}

// A rule with no application selector at all must fail with the friendly
// validation error, not reach the API as `application: null`.
func TestHydrateAddRequiresApplicationSelector(t *testing.T) {
	t.Parallel()
	plan := newMinimalApplicationControlRulePlan(cato_models.ApplicationControlRuleTypeApplication, "application_rule")
	setTypedRuleAttr(t, &plan, "application_rule", "application", nullApplicationObject())
	_, diags := hydrateApplicationControlAddRuleInput(context.Background(), plan)
	if !diags.HasError() {
		t.Fatal("expected a validation error for a rule with no application selector")
	}
	found := false
	for _, d := range diags.Errors() {
		if strings.Contains(d.Summary()+d.Detail(), "application selection required") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'application selection required' error, got: %+v", diags)
	}
}

// application_type must flow into the input (the selector for type/criteria
// rules like "all CLOUD_APPLICATIONs supporting PCI DSS").
func TestHydrateAddApplicationType(t *testing.T) {
	t.Parallel()
	plan := newMinimalApplicationControlRulePlan(cato_models.ApplicationControlRuleTypeApplication, "application_rule")
	got, diags := hydrateApplicationControlAddRuleInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	app := got.Rule.ApplicationRule.Application
	if app == nil || len(app.ApplicationType) != 1 || app.ApplicationType[0] != cato_models.ApplicationTypeCloudApplication {
		t.Fatalf("expected applicationType [CLOUD_APPLICATION], got %+v", app)
	}
}

func TestCriteriaInputFromPlan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	crit := types.ObjectValueMust(acCriteriaAttrTypes, map[string]attr.Value{
		"compliance_attributes": types.MapValueMust(types.StringType, map[string]attr.Value{
			"pciDss": types.StringValue("SUPPORTED"),
		}),
		"security_attributes": types.MapNull(types.StringType),
		"origin_country":      types.SetNull(idname.RefObjectType),
		"risk": types.ListValueMust(types.ObjectType{AttrTypes: acCriteriaRiskAttrTypes}, []attr.Value{
			types.ObjectValueMust(acCriteriaRiskAttrTypes, map[string]attr.Value{
				"risk":          types.StringValue("7"),
				"risk_operator": types.StringValue("GREATER_THAN"),
			}),
		}),
		"ai_risk": types.ListNull(types.StringType),
	})
	var diags diag.Diagnostics
	out, diags := acCriteriaInputFromPlan(ctx, crit, "application_rule.application_criteria", diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if out.Attributes.ComplianceAttributes.PciDss != cato_models.ApplicationControlAttributeValueSupported {
		t.Fatalf("expected pciDss SUPPORTED, got %q", out.Attributes.ComplianceAttributes.PciDss)
	}
	// omitted keys stay ANY
	if out.Attributes.ComplianceAttributes.Iso27001 != cato_models.ApplicationControlAttributeValueAny {
		t.Fatalf("expected iso27001 ANY default, got %q", out.Attributes.ComplianceAttributes.Iso27001)
	}
	if len(out.Risk) != 1 || string(out.Risk[0].Risk) != "7" || out.Risk[0].RiskOperator != "GREATER_THAN" {
		t.Fatalf("unexpected risk rows: %+v", out.Risk)
	}
}

func TestCriteriaInputRejectsExplicitANYAndUnknownKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for name, m := range map[string]map[string]attr.Value{
		"explicit ANY": {"pciDss": types.StringValue("ANY")},
		"unknown key":  {"pciDss2": types.StringValue("SUPPORTED")},
	} {
		crit := types.ObjectValueMust(acCriteriaAttrTypes, map[string]attr.Value{
			"compliance_attributes": types.MapValueMust(types.StringType, m),
			"security_attributes":   types.MapNull(types.StringType),
			"origin_country":        types.SetNull(idname.RefObjectType),
			"risk":                  types.ListNull(types.ObjectType{AttrTypes: acCriteriaRiskAttrTypes}),
			"ai_risk":               types.ListNull(types.StringType),
		})
		var diags diag.Diagnostics
		_, diags = acCriteriaInputFromPlan(ctx, crit, "p", diags)
		if !diags.HasError() {
			t.Fatalf("%s: expected a validation error", name)
		}
	}
}

// Activity granular conditions (field/operator/value) must reach the input —
// several catalog activities (e.g. full_path_url) REQUIRE a value.
func TestHydrateAddActivityConditions(t *testing.T) {
	t.Parallel()
	plan := newMinimalApplicationControlRulePlan(cato_models.ApplicationControlRuleTypeApplication, "application_rule")
	activity := types.ListValueMust(types.ObjectType{AttrTypes: applicationControlActivityAttrTypes}, []attr.Value{
		types.ObjectValueMust(applicationControlActivityAttrTypes, map[string]attr.Value{
			"activity": types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
				"id": types.StringValue("full_path_url"), "name": types.StringNull(),
			}),
			"field": types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
				"id": types.StringValue("full_url"), "name": types.StringNull(),
			}),
			"operator":  types.StringValue("CONTAINS"),
			"value":     types.StringValue("catonetworks.sharepoint.com"),
			"value_set": types.ObjectNull(idname.RefAttrTypes),
		}),
	})
	setTypedRuleAttr(t, &plan, "application_rule", "application_activity", activity)
	got, diags := hydrateApplicationControlAddRuleInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	rows := got.Rule.ApplicationRule.ApplicationActivity
	if len(rows) != 1 {
		t.Fatalf("expected 1 activity row, got %d", len(rows))
	}
	row := rows[0]
	if row.Activity == nil || row.Activity.Input != "full_path_url" {
		t.Fatalf("unexpected activity ref: %+v", row.Activity)
	}
	if row.Field == nil || row.Field.Input != "full_url" {
		t.Fatalf("expected field ref full_url, got %+v", row.Field)
	}
	if row.Operator == nil || *row.Operator != "CONTAINS" {
		t.Fatalf("expected operator CONTAINS, got %+v", row.Operator)
	}
	if row.Value == nil || *row.Value != "catonetworks.sharepoint.com" {
		t.Fatalf("expected condition value, got %+v", row.Value)
	}
}

// CONTENT_TYPE file-attribute rows select via content-type refs and the API
// requires `value` to stay null for them.
func TestFileAttributesContentTypeRefs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rows := types.ListValueMust(types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}, []attr.Value{
		types.ObjectValueMust(applicationControlFileAttributeAttrTypes, map[string]attr.Value{
			"file_attribute":      types.StringValue("CONTENT_TYPE"),
			"operator":            types.StringValue("IS"),
			"value":               types.StringNull(),
			"content_type_values": types.SetNull(idname.RefObjectType),
			"content_type_group_values": types.SetValueMust(idname.RefObjectType, []attr.Value{
				types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
					"id": types.StringNull(), "name": types.StringValue("Binaries"),
				}),
			}),
		}),
	})
	var diags diag.Diagnostics
	out, diags := acFileAttributesAdd(ctx, rows, diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 row, got %d", len(out))
	}
	if out[0].Value != nil {
		t.Fatalf("expected nil value for CONTENT_TYPE row, got %q", *out[0].Value)
	}
	if len(out[0].ContentTypeGroupValues) != 1 || out[0].ContentTypeGroupValues[0].Input != "Binaries" {
		t.Fatalf("unexpected content type group refs: %+v", out[0].ContentTypeGroupValues)
	}
}

// Data rules without a dlp_profile must send an EMPTY DlpProfileInput, not
// error — the API itself decides (live tenants carry such rules; some
// accounts' validation rejects them, which then surfaces as the API error).
func TestHydrateDataRuleAllowsMissingDlpProfile(t *testing.T) {
	t.Parallel()
	plan := newMinimalApplicationControlRulePlan(cato_models.ApplicationControlRuleTypeData, "data_rule")
	got, diags := hydrateApplicationControlAddRuleInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	dlp := got.Rule.DataRule.DlpProfile
	if dlp == nil {
		t.Fatal("expected non-nil (empty) dlp profile input")
	}
	if len(dlp.ContentProfile) != 0 || len(dlp.EdmProfile) != 0 {
		t.Fatalf("expected empty dlp profile lists, got %+v", dlp)
	}
}
