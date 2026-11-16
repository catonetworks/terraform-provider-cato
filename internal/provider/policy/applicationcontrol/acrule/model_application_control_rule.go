package acrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	applicationcontrol2 "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/applicationcontrol"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyschedule"
)

// ApplicationControlAccessMethodAttrTypes is one access-method row.
var ApplicationControlAccessMethodAttrTypes = map[string]attr.Type{
	"access_method": types.StringType,
	"operator":      types.StringType,
	"value":         types.StringType,
}

var applicationControlActivityAttrTypes = map[string]attr.Type{
	"activity":  idname.RefObjectType,
	"field":     idname.RefObjectType,
	"operator":  types.StringType,
	"value":     types.StringType,
	"value_set": idname.RefObjectType,
}

// applicationControlTypedRuleAttrTypes is shared shape for application/data/file nested blocks.
var applicationControlTypedRuleAttrTypes = map[string]attr.Type{
	"action":                       types.StringType,
	"severity":                     types.StringType,
	"schedule":                     policyschedule.ScheduleObjectType,
	"source":                       applicationcontrol2.ApplicationControlSourceObjectType,
	"tracking":                     policyrule.TrackingObjectType,
	"device":                       types.SetType{ElemType: idname.RefObjectType},
	"access_method":                types.ListType{ElemType: types.ObjectType{AttrTypes: ApplicationControlAccessMethodAttrTypes}},
	"application":                  acApplicationObjectType,
	"application_activity":         types.ListType{ElemType: types.ObjectType{AttrTypes: applicationControlActivityAttrTypes}},
	"application_activity_satisfy": types.StringType,
	"application_criteria":         types.ObjectType{AttrTypes: acCriteriaAttrTypes},
	"application_criteria_satisfy": types.StringType,
	"action_config":                types.ObjectType{AttrTypes: applicationControlActionConfigAttrTypes},
	"file_attribute":               types.ListType{ElemType: types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}},
	"file_attribute_satisfy":       types.StringType,
	"dlp_profile":                  types.ObjectType{AttrTypes: applicationControlDlpProfileAttrTypes},
}

var applicationControlActionConfigAttrTypes = map[string]attr.Type{
	"user_notification": types.SetType{ElemType: idname.RefObjectType},
}

var applicationControlDlpProfileAttrTypes = map[string]attr.Type{
	"content_profile": types.SetType{ElemType: idname.RefObjectType},
	"edm_profile":     types.SetType{ElemType: idname.RefObjectType},
}

var applicationControlFileAttributeAttrTypes = map[string]attr.Type{
	"file_attribute":            types.StringType,
	"operator":                  types.StringType,
	"value":                     types.StringType,
	"content_type_values":       types.SetType{ElemType: idname.RefObjectType},
	"content_type_group_values": types.SetType{ElemType: idname.RefObjectType},
}

// ApplicationControlRuleRuleAttrTypes defines nested `rule` object.
var ApplicationControlRuleRuleAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"name":             types.StringType,
	"description":      types.StringType,
	"enabled":          types.BoolType,
	"rule_type":        types.StringType,
	"application_rule": types.ObjectType{AttrTypes: applicationControlTypedRuleAttrTypes},
	"data_rule":        types.ObjectType{AttrTypes: applicationControlTypedRuleAttrTypes},
	"file_rule":        types.ObjectType{AttrTypes: applicationControlTypedRuleAttrTypes},
}

// ApplicationControlRuleAttrTypes is the top-level resource attribute map.
var ApplicationControlRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: ApplicationControlRuleRuleAttrTypes},
}

// ApplicationControlRule is the Terraform model for cato_application_control_rule.
type ApplicationControlRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// ApplicationControlRuleRulePlan maps the nested rule block.
type ApplicationControlRuleRulePlan struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	RuleType        types.String `tfsdk:"rule_type"`
	ApplicationRule types.Object `tfsdk:"application_rule"`
	DataRule        types.Object `tfsdk:"data_rule"`
	FileRule        types.Object `tfsdk:"file_rule"`
}

// ApplicationControlTypedRulePlan maps application_rule / data_rule / file_rule blocks.
type ApplicationControlTypedRulePlan struct {
	Action                     types.String `tfsdk:"action"`
	Severity                   types.String `tfsdk:"severity"`
	Schedule                   types.Object `tfsdk:"schedule"`
	Source                     types.Object `tfsdk:"source"`
	Tracking                   types.Object `tfsdk:"tracking"`
	Device                     types.Set    `tfsdk:"device"`
	AccessMethod               types.List   `tfsdk:"access_method"`
	Application                types.Object `tfsdk:"application"`
	ApplicationActivity        types.List   `tfsdk:"application_activity"`
	ApplicationActivitySatisfy types.String `tfsdk:"application_activity_satisfy"`
	ApplicationCriteria        types.Object `tfsdk:"application_criteria"`
	ApplicationCriteriaSatisfy types.String `tfsdk:"application_criteria_satisfy"`
	ActionConfig               types.Object `tfsdk:"action_config"`
	FileAttribute              types.List   `tfsdk:"file_attribute"`
	FileAttributeSatisfy       types.String `tfsdk:"file_attribute_satisfy"`
	DlpProfile                 types.Object `tfsdk:"dlp_profile"`
}

// ApplicationControlActivityPlan is one application_activity element. The
// optional field/operator/value/value_set carry activity granular conditions
// (e.g. Full Path URL CONTAINS "x") — several catalog activities REQUIRE a
// value (the API rejects a bare ref with "activity: <name> must have a value").
type ApplicationControlActivityPlan struct {
	Activity types.Object `tfsdk:"activity"`
	Field    types.Object `tfsdk:"field"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
	ValueSet types.Object `tfsdk:"value_set"`
}

// ApplicationControlAccessMethodPlan is one access_method element.
type ApplicationControlAccessMethodPlan struct {
	AccessMethod types.String `tfsdk:"access_method"`
	Operator     types.String `tfsdk:"operator"`
	Value        types.String `tfsdk:"value"`
}
