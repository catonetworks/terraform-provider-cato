package sbrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// SocketBypassFromToAttrTypes is a from/to range.
var SocketBypassFromToAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// SocketBypassActionAttrTypes is the bypass action selector.
var SocketBypassActionAttrTypes = map[string]attr.Type{
	"port":          types.StringType,
	"event_enabled": types.BoolType,
}

// SocketBypassSiteAttrTypes is the site scope selector.
var SocketBypassSiteAttrTypes = map[string]attr.Type{
	"site":  types.SetType{ElemType: idname.RefObjectType},
	"group": types.SetType{ElemType: idname.RefObjectType},
}

// SocketBypassSourceAttrTypes is the source selector.
var SocketBypassSourceAttrTypes = map[string]attr.Type{
	"floating_subnet":     types.SetType{ElemType: idname.RefObjectType},
	"global_ip_range":     types.SetType{ElemType: idname.RefObjectType},
	"group":               types.SetType{ElemType: idname.RefObjectType},
	"host":                types.SetType{ElemType: idname.RefObjectType},
	"ip":                  types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: types.ObjectType{AttrTypes: SocketBypassFromToAttrTypes}},
	"network_interface":   types.SetType{ElemType: idname.RefObjectType},
	"site_network_subnet": types.SetType{ElemType: idname.RefObjectType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"vlan":                types.ListType{ElemType: types.Int64Type},
}

// SocketBypassDestinationAttrTypes is the destination selector.
var SocketBypassDestinationAttrTypes = map[string]attr.Type{
	"application":     types.SetType{ElemType: idname.RefObjectType},
	"custom_app":      types.SetType{ElemType: idname.RefObjectType},
	"domain":          types.ListType{ElemType: types.StringType},
	"fqdn":            types.ListType{ElemType: types.StringType},
	"global_ip_range": types.SetType{ElemType: idname.RefObjectType},
	"ip":              types.ListType{ElemType: types.StringType},
	"ip_range":        types.ListType{ElemType: types.ObjectType{AttrTypes: SocketBypassFromToAttrTypes}},
	"subnet":          types.ListType{ElemType: types.StringType},
}

// SocketBypassCustomServiceAttrTypes is one custom service entry.
var SocketBypassCustomServiceAttrTypes = map[string]attr.Type{
	"port":       types.ListType{ElemType: types.StringType},
	"port_range": types.ObjectType{AttrTypes: SocketBypassFromToAttrTypes},
	"protocol":   types.StringType,
}

// SocketBypassServiceAttrTypes is the service selector.
var SocketBypassServiceAttrTypes = map[string]attr.Type{
	"simple": types.ListType{ElemType: types.StringType},
	"custom": types.ListType{ElemType: types.ObjectType{AttrTypes: SocketBypassCustomServiceAttrTypes}},
}

// SocketBypassExceptionAttrTypes is one exception entry.
var SocketBypassExceptionAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"site":        types.ObjectType{AttrTypes: SocketBypassSiteAttrTypes},
	"source":      types.ObjectType{AttrTypes: SocketBypassSourceAttrTypes},
	"destination": types.ObjectType{AttrTypes: SocketBypassDestinationAttrTypes},
	"service":     types.ObjectType{AttrTypes: SocketBypassServiceAttrTypes},
}

// SocketBypassRuleRuleAttrTypes defines the nested `rule` object for socket bypass rules.
var SocketBypassRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"enabled":     types.BoolType,
	"action":      types.ObjectType{AttrTypes: SocketBypassActionAttrTypes},
	"site":        types.ObjectType{AttrTypes: SocketBypassSiteAttrTypes},
	"source":      types.ObjectType{AttrTypes: SocketBypassSourceAttrTypes},
	"destination": types.ObjectType{AttrTypes: SocketBypassDestinationAttrTypes},
	"service":     types.ObjectType{AttrTypes: SocketBypassServiceAttrTypes},
	"exception":   types.ListType{ElemType: types.ObjectType{AttrTypes: SocketBypassExceptionAttrTypes}},
}

// SocketBypassRuleAttrTypes is the top-level resource attribute map.
var SocketBypassRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: SocketBypassRuleRuleAttrTypes},
}

// SocketBypassRule is the Terraform model for cato_socket_bypass_rule.
type SocketBypassRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// SocketBypassRuleRulePlan maps the nested rule block (ObjectAs).
type SocketBypassRuleRulePlan struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Action      types.Object `tfsdk:"action"`
	Site        types.Object `tfsdk:"site"`
	Source      types.Object `tfsdk:"source"`
	Destination types.Object `tfsdk:"destination"`
	Service     types.Object `tfsdk:"service"`
	Exception   types.List   `tfsdk:"exception"`
}

// SocketBypassActionPlan maps the action selector.
type SocketBypassActionPlan struct {
	Port         types.String `tfsdk:"port"`
	EventEnabled types.Bool   `tfsdk:"event_enabled"`
}

// SocketBypassSitePlan maps the site scope selector.
type SocketBypassSitePlan struct {
	Site  types.Set `tfsdk:"site"`
	Group types.Set `tfsdk:"group"`
}

// SocketBypassSourcePlan maps the source selector.
type SocketBypassSourcePlan struct {
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range"`
	Group             types.Set  `tfsdk:"group"`
	Host              types.Set  `tfsdk:"host"`
	IP                types.List `tfsdk:"ip"`
	IPRange           types.List `tfsdk:"ip_range"`
	NetworkInterface  types.Set  `tfsdk:"network_interface"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet"`
	Subnet            types.List `tfsdk:"subnet"`
	Vlan              types.List `tfsdk:"vlan"`
}

// SocketBypassDestinationPlan maps the destination selector.
type SocketBypassDestinationPlan struct {
	Application   types.Set  `tfsdk:"application"`
	CustomApp     types.Set  `tfsdk:"custom_app"`
	Domain        types.List `tfsdk:"domain"`
	Fqdn          types.List `tfsdk:"fqdn"`
	GlobalIPRange types.Set  `tfsdk:"global_ip_range"`
	IP            types.List `tfsdk:"ip"`
	IPRange       types.List `tfsdk:"ip_range"`
	Subnet        types.List `tfsdk:"subnet"`
}

// SocketBypassCustomServicePlan maps one custom service entry.
type SocketBypassCustomServicePlan struct {
	Port      types.List   `tfsdk:"port"`
	PortRange types.Object `tfsdk:"port_range"`
	Protocol  types.String `tfsdk:"protocol"`
}

// SocketBypassServicePlan maps the service selector.
type SocketBypassServicePlan struct {
	Simple types.List `tfsdk:"simple"`
	Custom types.List `tfsdk:"custom"`
}

// SocketBypassFromToPlan maps a from/to range.
type SocketBypassFromToPlan struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

// SocketBypassExceptionPlan maps one exception entry.
type SocketBypassExceptionPlan struct {
	Name        types.String `tfsdk:"name"`
	Site        types.Object `tfsdk:"site"`
	Source      types.Object `tfsdk:"source"`
	Destination types.Object `tfsdk:"destination"`
	Service     types.Object `tfsdk:"service"`
}
