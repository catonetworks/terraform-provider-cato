package sbrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// resourceModel contains Terraform values for socket_bypass_rule.
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
	Action      types.Object `tfsdk:"action"`
	Description types.String `tfsdk:"description"`
	Destination types.Object `tfsdk:"destination"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Exception   types.Set    `tfsdk:"exception"`
	ID          types.String `tfsdk:"id"`
	Index       types.Int64  `tfsdk:"index"`
	Name        types.String `tfsdk:"name"`
	Section     types.Object `tfsdk:"section"`
	Service     types.Object `tfsdk:"service"`
	Site        types.Object `tfsdk:"site"`
	Source      types.Object `tfsdk:"source"`
}

var ruleAttrTypes = map[string]attr.Type{
	"action":      types.ObjectType{AttrTypes: ruleActionAttrTypes},
	"description": types.StringType,
	"destination": types.ObjectType{AttrTypes: ruleDestinationAttrTypes},
	"enabled":     types.BoolType,
	"exception":   types.SetType{ElemType: types.ObjectType{AttrTypes: ruleExceptionAttrTypes}},
	"id":          types.StringType,
	"index":       types.Int64Type,
	"name":        types.StringType,
	"section":     types.ObjectType{AttrTypes: ruleSectionAttrTypes},
	"service":     types.ObjectType{AttrTypes: ruleServiceAttrTypes},
	"site":        types.ObjectType{AttrTypes: ruleSiteAttrTypes},
	"source":      types.ObjectType{AttrTypes: ruleSourceAttrTypes},
}

// RuleActionModel contains Terraform values for rule.action.
type RuleActionModel struct {
	Event types.Object `tfsdk:"event"`
	Port  types.String `tfsdk:"port"`
}

var ruleActionAttrTypes = map[string]attr.Type{
	"event": types.ObjectType{AttrTypes: ruleActionEventAttrTypes},
	"port":  types.StringType,
}

// RuleActionEventModel contains Terraform values for rule.action.event.
type RuleActionEventModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

var ruleActionEventAttrTypes = map[string]attr.Type{
	"enabled": types.BoolType,
}

// RuleDestinationModel contains Terraform values for rule.destination.
type RuleDestinationModel struct {
	Application   types.Set `tfsdk:"application"`
	CustomApp     types.Set `tfsdk:"custom_app"`
	Domain        types.Set `tfsdk:"domain"`
	Fqdn          types.Set `tfsdk:"fqdn"`
	GlobalIPRange types.Set `tfsdk:"global_ip_range"`
	IP            types.Set `tfsdk:"ip"`
	IPRange       types.Set `tfsdk:"ip_range"`
	Subnet        types.Set `tfsdk:"subnet"`
}

var ruleDestinationAttrTypes = map[string]attr.Type{
	"application":     types.SetType{ElemType: idname.RefObjectType},
	"custom_app":      types.SetType{ElemType: idname.RefObjectType},
	"domain":          types.SetType{ElemType: types.StringType},
	"fqdn":            types.SetType{ElemType: types.StringType},
	"global_ip_range": types.SetType{ElemType: idname.RefObjectType},
	"ip":              types.SetType{ElemType: types.StringType},
	"ip_range":        types.SetType{ElemType: types.ObjectType{AttrTypes: ruleDestinationIPRangeAttrTypes}},
	"subnet":          types.SetType{ElemType: types.StringType},
}

// RuleDestinationIPRangeModel contains Terraform values for rule.destination.ip_range.
type RuleDestinationIPRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var ruleDestinationIPRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// RuleExceptionModel contains Terraform values for rule.exception.
type RuleExceptionModel struct {
	Destination types.Object `tfsdk:"destination"`
	Name        types.String `tfsdk:"name"`
	Service     types.Object `tfsdk:"service"`
	Site        types.Object `tfsdk:"site"`
	Source      types.Object `tfsdk:"source"`
}

var ruleExceptionAttrTypes = map[string]attr.Type{
	"destination": types.ObjectType{AttrTypes: ruleExceptionDestinationAttrTypes},
	"name":        types.StringType,
	"service":     types.ObjectType{AttrTypes: ruleExceptionServiceAttrTypes},
	"site":        types.ObjectType{AttrTypes: ruleExceptionSiteAttrTypes},
	"source":      types.ObjectType{AttrTypes: ruleExceptionSourceAttrTypes},
}

// RuleExceptionDestinationModel contains Terraform values for rule.exception.destination.
type RuleExceptionDestinationModel struct {
	Application   types.Set `tfsdk:"application"`
	CustomApp     types.Set `tfsdk:"custom_app"`
	Domain        types.Set `tfsdk:"domain"`
	Fqdn          types.Set `tfsdk:"fqdn"`
	GlobalIPRange types.Set `tfsdk:"global_ip_range"`
	IP            types.Set `tfsdk:"ip"`
	IPRange       types.Set `tfsdk:"ip_range"`
	Subnet        types.Set `tfsdk:"subnet"`
}

var ruleExceptionDestinationAttrTypes = map[string]attr.Type{
	"application":     types.SetType{ElemType: idname.RefObjectType},
	"custom_app":      types.SetType{ElemType: idname.RefObjectType},
	"domain":          types.SetType{ElemType: types.StringType},
	"fqdn":            types.SetType{ElemType: types.StringType},
	"global_ip_range": types.SetType{ElemType: idname.RefObjectType},
	"ip":              types.SetType{ElemType: types.StringType},
	"ip_range":        types.SetType{ElemType: types.ObjectType{AttrTypes: ruleExceptionDestinationIPRangeAttrTypes}},
	"subnet":          types.SetType{ElemType: types.StringType},
}

// RuleExceptionDestinationIPRangeModel contains Terraform values for rule.exception.destination.ip_range.
type RuleExceptionDestinationIPRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var ruleExceptionDestinationIPRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// RuleExceptionServiceModel contains Terraform values for rule.exception.service.
type RuleExceptionServiceModel struct {
	Custom types.Set `tfsdk:"custom"`
	Simple types.Set `tfsdk:"simple"`
}

var ruleExceptionServiceAttrTypes = map[string]attr.Type{
	"custom": types.SetType{ElemType: types.ObjectType{AttrTypes: ruleExceptionServiceCustomAttrTypes}},
	"simple": types.SetType{ElemType: types.ObjectType{AttrTypes: ruleExceptionServiceSimpleAttrTypes}},
}

// RuleExceptionServiceCustomModel contains Terraform values for rule.exception.service.custom.
type RuleExceptionServiceCustomModel struct {
	Port      types.Set    `tfsdk:"port"`
	PortRange types.Object `tfsdk:"port_range"`
	Protocol  types.String `tfsdk:"protocol"`
}

var ruleExceptionServiceCustomAttrTypes = map[string]attr.Type{
	"port":       types.SetType{ElemType: types.StringType},
	"port_range": types.ObjectType{AttrTypes: ruleExceptionServiceCustomPortRangeAttrTypes},
	"protocol":   types.StringType,
}

// RuleExceptionServiceCustomPortRangeModel contains Terraform values for rule.exception.service.custom.port_range.
type RuleExceptionServiceCustomPortRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var ruleExceptionServiceCustomPortRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// RuleExceptionServiceSimpleModel contains Terraform values for rule.exception.service.simple.
type RuleExceptionServiceSimpleModel struct {
	Name types.String `tfsdk:"name"`
}

var ruleExceptionServiceSimpleAttrTypes = map[string]attr.Type{
	"name": types.StringType,
}

// RuleExceptionSiteModel contains Terraform values for rule.exception.site.
type RuleExceptionSiteModel struct {
	Group types.Set `tfsdk:"group"`
	Site  types.Set `tfsdk:"site"`
}

var ruleExceptionSiteAttrTypes = map[string]attr.Type{
	"group": types.SetType{ElemType: idname.RefObjectType},
	"site":  types.SetType{ElemType: idname.RefObjectType},
}

// RuleExceptionSourceModel contains Terraform values for rule.exception.source.
type RuleExceptionSourceModel struct {
	FloatingSubnet    types.Set `tfsdk:"floating_subnet"`
	GlobalIPRange     types.Set `tfsdk:"global_ip_range"`
	Group             types.Set `tfsdk:"group"`
	Host              types.Set `tfsdk:"host"`
	IP                types.Set `tfsdk:"ip"`
	IPRange           types.Set `tfsdk:"ip_range"`
	NetworkInterface  types.Set `tfsdk:"network_interface"`
	SiteNetworkSubnet types.Set `tfsdk:"site_network_subnet"`
	Subnet            types.Set `tfsdk:"subnet"`
	Vlan              types.Set `tfsdk:"vlan"`
}

var ruleExceptionSourceAttrTypes = map[string]attr.Type{
	"floating_subnet":     types.SetType{ElemType: idname.RefObjectType},
	"global_ip_range":     types.SetType{ElemType: idname.RefObjectType},
	"group":               types.SetType{ElemType: idname.RefObjectType},
	"host":                types.SetType{ElemType: idname.RefObjectType},
	"ip":                  types.SetType{ElemType: types.StringType},
	"ip_range":            types.SetType{ElemType: types.ObjectType{AttrTypes: ruleExceptionSourceIPRangeAttrTypes}},
	"network_interface":   types.SetType{ElemType: idname.RefObjectType},
	"site_network_subnet": types.SetType{ElemType: idname.RefObjectType},
	"subnet":              types.SetType{ElemType: types.StringType},
	"vlan":                types.SetType{ElemType: types.Int64Type},
}

// RuleExceptionSourceIPRangeModel contains Terraform values for rule.exception.source.ip_range.
type RuleExceptionSourceIPRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var ruleExceptionSourceIPRangeAttrTypes = map[string]attr.Type{
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

// RuleServiceModel contains Terraform values for rule.service.
type RuleServiceModel struct {
	Custom types.Set `tfsdk:"custom"`
	Simple types.Set `tfsdk:"simple"`
}

var ruleServiceAttrTypes = map[string]attr.Type{
	"custom": types.SetType{ElemType: types.ObjectType{AttrTypes: ruleServiceCustomAttrTypes}},
	"simple": types.SetType{ElemType: types.ObjectType{AttrTypes: ruleServiceSimpleAttrTypes}},
}

// RuleServiceCustomModel contains Terraform values for rule.service.custom.
type RuleServiceCustomModel struct {
	Port      types.Set    `tfsdk:"port"`
	PortRange types.Object `tfsdk:"port_range"`
	Protocol  types.String `tfsdk:"protocol"`
}

var ruleServiceCustomAttrTypes = map[string]attr.Type{
	"port":       types.SetType{ElemType: types.StringType},
	"port_range": types.ObjectType{AttrTypes: ruleServiceCustomPortRangeAttrTypes},
	"protocol":   types.StringType,
}

// RuleServiceCustomPortRangeModel contains Terraform values for rule.service.custom.port_range.
type RuleServiceCustomPortRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var ruleServiceCustomPortRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// RuleServiceSimpleModel contains Terraform values for rule.service.simple.
type RuleServiceSimpleModel struct {
	Name types.String `tfsdk:"name"`
}

var ruleServiceSimpleAttrTypes = map[string]attr.Type{
	"name": types.StringType,
}

// RuleSiteModel contains Terraform values for rule.site.
type RuleSiteModel struct {
	Group types.Set `tfsdk:"group"`
	Site  types.Set `tfsdk:"site"`
}

var ruleSiteAttrTypes = map[string]attr.Type{
	"group": types.SetType{ElemType: idname.RefObjectType},
	"site":  types.SetType{ElemType: idname.RefObjectType},
}

// RuleSourceModel contains Terraform values for rule.source.
type RuleSourceModel struct {
	FloatingSubnet    types.Set `tfsdk:"floating_subnet"`
	GlobalIPRange     types.Set `tfsdk:"global_ip_range"`
	Group             types.Set `tfsdk:"group"`
	Host              types.Set `tfsdk:"host"`
	IP                types.Set `tfsdk:"ip"`
	IPRange           types.Set `tfsdk:"ip_range"`
	NetworkInterface  types.Set `tfsdk:"network_interface"`
	SiteNetworkSubnet types.Set `tfsdk:"site_network_subnet"`
	Subnet            types.Set `tfsdk:"subnet"`
	Vlan              types.Set `tfsdk:"vlan"`
}

var ruleSourceAttrTypes = map[string]attr.Type{
	"floating_subnet":     types.SetType{ElemType: idname.RefObjectType},
	"global_ip_range":     types.SetType{ElemType: idname.RefObjectType},
	"group":               types.SetType{ElemType: idname.RefObjectType},
	"host":                types.SetType{ElemType: idname.RefObjectType},
	"ip":                  types.SetType{ElemType: types.StringType},
	"ip_range":            types.SetType{ElemType: types.ObjectType{AttrTypes: ruleSourceIPRangeAttrTypes}},
	"network_interface":   types.SetType{ElemType: idname.RefObjectType},
	"site_network_subnet": types.SetType{ElemType: idname.RefObjectType},
	"subnet":              types.SetType{ElemType: types.StringType},
	"vlan":                types.SetType{ElemType: types.Int64Type},
}

// RuleSourceIPRangeModel contains Terraform values for rule.source.ip_range.
type RuleSourceIPRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

var ruleSourceIPRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}
