package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SocketLanFirewallRule represents the top-level resource structure for firewall rules
type SocketLanFirewallRule struct {
	Rule types.Object `tfsdk:"rule" json:"rule,omitempty"`
	At   types.Object `tfsdk:"at" json:"at,omitempty"`
}

// SocketLanFirewallRulePositionInput represents the position input for firewall rules
type SocketLanFirewallRulePositionInput struct {
	Position types.String `tfsdk:"position" json:"position,omitempty"`
	Ref      types.String `tfsdk:"ref" json:"ref,omitempty"`
}

// SocketLanFirewallRuleData represents the rule data within the firewall rule resource
type SocketLanFirewallRuleData struct {
	ID          types.String `tfsdk:"id" json:"id,omitempty"`
	Name        types.String `tfsdk:"name" json:"name,omitempty"`
	Description types.String `tfsdk:"description" json:"description,omitempty"`
	Index       types.Int64  `tfsdk:"index" json:"index,omitempty"`
	Enabled     types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	Direction   types.String `tfsdk:"direction" json:"direction,omitempty"`
	Action      types.String `tfsdk:"action" json:"action,omitempty"`
	Source      types.Object `tfsdk:"source" json:"source,omitempty"`
	Destination types.Object `tfsdk:"destination" json:"destination,omitempty"`
	Application types.Object `tfsdk:"application" json:"application,omitempty"`
	Service     types.Object `tfsdk:"service" json:"service,omitempty"`
	Tracking    types.Object `tfsdk:"tracking" json:"tracking,omitempty"`
}

// SocketLanFirewallSource represents source criteria for firewall rules (includes mac and site)
type SocketLanFirewallSource struct {
	Vlan              types.List `tfsdk:"vlan" json:"vlan,omitempty"`
	Mac               types.List `tfsdk:"mac" json:"mac,omitempty"`
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"`
	Site              types.Set  `tfsdk:"site" json:"site,omitempty"`
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"`
}

// SocketLanFirewallDestination represents destination criteria for firewall rules (includes site)
type SocketLanFirewallDestination struct {
	Vlan              types.List `tfsdk:"vlan" json:"vlan,omitempty"`
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"`
	Site              types.Set  `tfsdk:"site" json:"site,omitempty"`
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"`
}

// SocketLanFirewallApplication represents application matching criteria
type SocketLanFirewallApplication struct {
	Application   types.Set  `tfsdk:"application" json:"application,omitempty"`
	CustomApp     types.Set  `tfsdk:"custom_app" json:"custom_app,omitempty"`
	Domain        types.List `tfsdk:"domain" json:"domain,omitempty"`
	Fqdn          types.List `tfsdk:"fqdn" json:"fqdn,omitempty"`
	IP            types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet        types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange       types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`
	GlobalIPRange types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`
}

// SocketLanFirewallService represents service criteria for firewall rules (includes standard)
type SocketLanFirewallService struct {
	Simple   types.Set  `tfsdk:"simple" json:"simple,omitempty"`
	Standard types.Set  `tfsdk:"standard" json:"standard,omitempty"`
	Custom   types.List `tfsdk:"custom" json:"custom,omitempty"`
}


// Attr types for state management

var SocketLanFirewallRuleObjectType = types.ObjectType{AttrTypes: SocketLanFirewallRuleAttrTypes}
var SocketLanFirewallRuleAttrTypes = map[string]attr.Type{
	"rule": SocketLanFirewallRuleRuleObjectType,
	"at":   PositionObjectType,
}

var SocketLanFirewallRuleRuleObjectType = types.ObjectType{AttrTypes: SocketLanFirewallRuleRuleAttrTypes}
var SocketLanFirewallRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"index":       types.Int64Type,
	"enabled":     types.BoolType,
	"direction":   types.StringType,
	"action":      types.StringType,
	"source":      SocketLanFirewallSourceObjectType,
	"destination": SocketLanFirewallDestinationObjectType,
	"application": SocketLanFirewallApplicationObjectType,
	"service":     SocketLanFirewallServiceObjectType,
	"tracking":    TrackingObjectType,
}

var SocketLanFirewallSourceObjectType = types.ObjectType{AttrTypes: SocketLanFirewallSourceAttrTypes}
var SocketLanFirewallSourceAttrTypes = map[string]attr.Type{
	"vlan":                types.ListType{ElemType: types.Int64Type},
	"mac":                 types.ListType{ElemType: types.StringType},
	"ip":                  types.ListType{ElemType: types.StringType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: FromToObjectType},
	"host":                types.SetType{ElemType: NameIDObjectType},
	"site":                types.SetType{ElemType: NameIDObjectType},
	"group":               types.SetType{ElemType: NameIDObjectType},
	"system_group":        types.SetType{ElemType: NameIDObjectType},
	"network_interface":   types.SetType{ElemType: NameIDObjectType},
	"global_ip_range":     types.SetType{ElemType: NameIDObjectType},
	"floating_subnet":     types.SetType{ElemType: NameIDObjectType},
	"site_network_subnet": types.SetType{ElemType: NameIDObjectType},
}

var SocketLanFirewallDestinationObjectType = types.ObjectType{AttrTypes: SocketLanFirewallDestinationAttrTypes}
var SocketLanFirewallDestinationAttrTypes = map[string]attr.Type{
	"vlan":                types.ListType{ElemType: types.Int64Type},
	"ip":                  types.ListType{ElemType: types.StringType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: FromToObjectType},
	"host":                types.SetType{ElemType: NameIDObjectType},
	"site":                types.SetType{ElemType: NameIDObjectType},
	"group":               types.SetType{ElemType: NameIDObjectType},
	"system_group":        types.SetType{ElemType: NameIDObjectType},
	"network_interface":   types.SetType{ElemType: NameIDObjectType},
	"global_ip_range":     types.SetType{ElemType: NameIDObjectType},
	"floating_subnet":     types.SetType{ElemType: NameIDObjectType},
	"site_network_subnet": types.SetType{ElemType: NameIDObjectType},
}

var SocketLanFirewallApplicationObjectType = types.ObjectType{AttrTypes: SocketLanFirewallApplicationAttrTypes}
var SocketLanFirewallApplicationAttrTypes = map[string]attr.Type{
	"application":     types.SetType{ElemType: NameIDObjectType},
	"custom_app":      types.SetType{ElemType: NameIDObjectType},
	"domain":          types.ListType{ElemType: types.StringType},
	"fqdn":            types.ListType{ElemType: types.StringType},
	"ip":              types.ListType{ElemType: types.StringType},
	"subnet":          types.ListType{ElemType: types.StringType},
	"ip_range":        types.ListType{ElemType: FromToObjectType},
	"global_ip_range": types.SetType{ElemType: NameIDObjectType},
}

var SocketLanFirewallServiceObjectType = types.ObjectType{AttrTypes: SocketLanFirewallServiceAttrTypes}
var SocketLanFirewallServiceAttrTypes = map[string]attr.Type{
	"simple":   types.SetType{ElemType: SimpleServiceObjectType},
	"standard": types.SetType{ElemType: NameIDObjectType},
	"custom":   types.ListType{ElemType: CustomServiceObjectType},
}
