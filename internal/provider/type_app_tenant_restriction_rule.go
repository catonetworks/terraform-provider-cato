package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AppTenantRestrictionHeaderAttrTypes is a single HTTP header name/value pair.
var AppTenantRestrictionHeaderAttrTypes = map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}

// AppTenantRestrictionRuleRuleAttrTypes defines the nested `rule` object for app tenant restriction rules.
var AppTenantRestrictionRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"enabled":     types.BoolType,
	"action":      types.StringType,
	"severity":    types.StringType,
	"application": NameIDObjectType,
	"headers":     types.ListType{ElemType: types.ObjectType{AttrTypes: AppTenantRestrictionHeaderAttrTypes}},
	"schedule":    ScheduleObjectType,
	"source":      ApplicationControlSourceObjectType,
}

// AppTenantRestrictionRuleAttrTypes is the top-level resource attribute map.
var AppTenantRestrictionRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   PositionObjectType,
	"rule": types.ObjectType{AttrTypes: AppTenantRestrictionRuleRuleAttrTypes},
}

// AppTenantRestrictionRule is the Terraform model for cato_app_tenant_restriction_rule.
type AppTenantRestrictionRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// AppTenantRestrictionRuleRulePlan maps the nested rule block (ObjectAs).
type AppTenantRestrictionRuleRulePlan struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Action      types.String `tfsdk:"action"`
	Severity    types.String `tfsdk:"severity"`
	Application types.Object `tfsdk:"application"`
	Headers     types.List   `tfsdk:"headers"`
	Schedule    types.Object `tfsdk:"schedule"`
	Source      types.Object `tfsdk:"source"`
}

// AppTenantRestrictionHeaderPlan is one element of `headers`.
type AppTenantRestrictionHeaderPlan struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
