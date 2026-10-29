package rpfrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// resourceModel contains Terraform values for remote_port_forward_rule.
type resourceModel struct {
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// AtModel contains Terraform values for at.
type AtModel struct {
	Position types.String `tfsdk:"position"`
	Ref      types.String `tfsdk:"ref"`
}

var atAttrTypes = map[string]attr.Type{
	"position": types.StringType,
	"ref":      types.StringType,
}

// RuleModel contains Terraform values for rule.
type RuleModel struct {
	Description       types.String `tfsdk:"description"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	ExternalIP        types.Object `tfsdk:"external_ip"`
	ExternalPortRange types.Object `tfsdk:"external_port_range"`
	ForwardIcmp       types.Bool   `tfsdk:"forward_icmp"`
	ID                types.String `tfsdk:"id"`
	Index             types.Int64  `tfsdk:"index"`
	InternalIP        types.String `tfsdk:"internal_ip"`
	InternalPortRange types.Object `tfsdk:"internal_port_range"`
	Name              types.String `tfsdk:"name"`
	RemoteIps         types.Object `tfsdk:"remote_ips"`
	RestrictionType   types.String `tfsdk:"restriction_type"`
	Section           types.Object `tfsdk:"section"`
	Tracking          types.Object `tfsdk:"tracking"`
}

var ruleAttrTypes = map[string]attr.Type{
	"description":         types.StringType,
	"enabled":             types.BoolType,
	"external_ip":         idname.RefObjectType,
	"external_port_range": types.ObjectType{AttrTypes: ruleExternalPortRangeAttrTypes},
	"forward_icmp":        types.BoolType,
	"id":                  types.StringType,
	"index":               types.Int64Type,
	"internal_ip":         types.StringType,
	"internal_port_range": types.ObjectType{AttrTypes: ruleInternalPortRangeAttrTypes},
	"name":                types.StringType,
	"remote_ips":          types.ObjectType{AttrTypes: ruleRemoteIpsAttrTypes},
	"restriction_type":    types.StringType,
	"section":             types.ObjectType{AttrTypes: ruleSectionAttrTypes},
	"tracking":            types.ObjectType{AttrTypes: ruleTrackingAttrTypes},
}

// RuleExternalPortRangeModel contains Terraform values for rule.external_port_range.
type RuleExternalPortRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var ruleExternalPortRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// RuleInternalPortRangeModel contains Terraform values for rule.internal_port_range.
type RuleInternalPortRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var ruleInternalPortRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// RuleRemoteIpsModel contains Terraform values for rule.remote_ips.
type RuleRemoteIpsModel struct {
	GlobalIPRange types.Set `tfsdk:"global_ip_range"`
	IP            types.Set `tfsdk:"ip"`
	IPRange       types.Set `tfsdk:"ip_range"`
	Subnet        types.Set `tfsdk:"subnet"`
}

var ruleRemoteIpsAttrTypes = map[string]attr.Type{
	"global_ip_range": types.SetType{ElemType: idname.RefObjectType},
	"ip":              types.SetType{ElemType: types.StringType},
	"ip_range":        types.SetType{ElemType: types.ObjectType{AttrTypes: ruleRemoteIpsIPRangeAttrTypes}},
	"subnet":          types.SetType{ElemType: types.StringType},
}

// RuleRemoteIpsIPRangeModel contains Terraform values for rule.remote_ips.ip_range.
type RuleRemoteIpsIPRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var ruleRemoteIpsIPRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// RuleSectionModel contains Terraform values for rule.section.
type RuleSectionModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var ruleSectionAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

// RuleTrackingModel contains Terraform values for rule.tracking.
type RuleTrackingModel struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	Frequency         types.String `tfsdk:"frequency"`
	MailingList       types.Set    `tfsdk:"mailing_list"`
	SubscriptionGroup types.Set    `tfsdk:"subscription_group"`
	Webhook           types.Set    `tfsdk:"webhook"`
}

var ruleTrackingAttrTypes = map[string]attr.Type{
	"enabled":            types.BoolType,
	"frequency":          types.StringType,
	"mailing_list":       types.SetType{ElemType: idname.RefObjectType},
	"subscription_group": types.SetType{ElemType: idname.RefObjectType},
	"webhook":            types.SetType{ElemType: idname.RefObjectType},
}
