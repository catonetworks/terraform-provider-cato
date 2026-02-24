package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SocketLanNetworkRule represents the top-level resource structure
type SocketLanNetworkRule struct {
	Rule types.Object `tfsdk:"rule" json:"rule,omitempty"`
	At   types.Object `tfsdk:"at" json:"at,omitempty"`
}

// SocketLanNetworkRuleData represents the rule data within the resource
type SocketLanNetworkRuleData struct {
	ID          types.String `tfsdk:"id" json:"id,omitempty"`
	Name        types.String `tfsdk:"name" json:"name,omitempty"`
	Description types.String `tfsdk:"description" json:"description,omitempty"`
	Index       types.Int64  `tfsdk:"index" json:"index,omitempty"`
	Enabled     types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	Direction   types.String `tfsdk:"direction" json:"direction,omitempty"`
	Transport   types.String `tfsdk:"transport" json:"transport,omitempty"`
	Site        types.Object `tfsdk:"site" json:"site,omitempty"`
	Source      types.Object `tfsdk:"source" json:"source,omitempty"`
	Destination types.Object `tfsdk:"destination" json:"destination,omitempty"`
	Service     types.Object `tfsdk:"service" json:"service,omitempty"`
	Nat         types.Object `tfsdk:"nat" json:"nat,omitempty"`
}

// SocketLanSite represents the site scope for the rule
type SocketLanSite struct {
	Site  types.Set `tfsdk:"site" json:"site,omitempty"`
	Group types.Set `tfsdk:"group" json:"group,omitempty"`
}

// SocketLanSource represents source criteria for network rules
type SocketLanSource struct {
	Vlan              types.List `tfsdk:"vlan" json:"vlan,omitempty"`
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"`
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"`
}

// SocketLanDestination represents destination criteria for network rules
type SocketLanDestination struct {
	Vlan              types.List `tfsdk:"vlan" json:"vlan,omitempty"`
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"`
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"`
}

// SocketLanService represents service criteria for network rules
type SocketLanService struct {
	Simple types.Set  `tfsdk:"simple" json:"simple,omitempty"`
	Custom types.List `tfsdk:"custom" json:"custom,omitempty"`
}

// SocketLanNat represents NAT settings for network rules
type SocketLanNat struct {
	Enabled types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	NatType types.String `tfsdk:"nat_type" json:"nat_type,omitempty"`
}

// SocketLanSimpleService represents a simple named service
type SocketLanSimpleService struct {
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

// Attr types for state management

var SocketLanNetworkRuleObjectType = types.ObjectType{AttrTypes: SocketLanNetworkRuleAttrTypes}
var SocketLanNetworkRuleAttrTypes = map[string]attr.Type{
	"rule": SocketLanNetworkRuleRuleObjectType,
	"at":   PositionObjectType,
}

var SocketLanNetworkRuleRuleObjectType = types.ObjectType{AttrTypes: SocketLanNetworkRuleRuleAttrTypes}
var SocketLanNetworkRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"index":       types.Int64Type,
	"enabled":     types.BoolType,
	"direction":   types.StringType,
	"transport":   types.StringType,
	"site":        SocketLanSiteObjectType,
	"source":      SocketLanSourceObjectType,
	"destination": SocketLanDestinationObjectType,
	"service":     SocketLanServiceObjectType,
	"nat":         SocketLanNatObjectType,
}

var SocketLanSiteObjectType = types.ObjectType{AttrTypes: SocketLanSiteAttrTypes}
var SocketLanSiteAttrTypes = map[string]attr.Type{
	"site":  types.SetType{ElemType: NameIDObjectType},
	"group": types.SetType{ElemType: NameIDObjectType},
}

var SocketLanSourceObjectType = types.ObjectType{AttrTypes: SocketLanSourceAttrTypes}
var SocketLanSourceAttrTypes = map[string]attr.Type{
	"vlan":                types.ListType{ElemType: types.Int64Type},
	"ip":                  types.ListType{ElemType: types.StringType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: FromToObjectType},
	"host":                types.SetType{ElemType: NameIDObjectType},
	"group":               types.SetType{ElemType: NameIDObjectType},
	"system_group":        types.SetType{ElemType: NameIDObjectType},
	"network_interface":   types.SetType{ElemType: NameIDObjectType},
	"global_ip_range":     types.SetType{ElemType: NameIDObjectType},
	"floating_subnet":     types.SetType{ElemType: NameIDObjectType},
	"site_network_subnet": types.SetType{ElemType: NameIDObjectType},
}

var SocketLanDestinationObjectType = types.ObjectType{AttrTypes: SocketLanDestinationAttrTypes}
var SocketLanDestinationAttrTypes = map[string]attr.Type{
	"vlan":                types.ListType{ElemType: types.Int64Type},
	"ip":                  types.ListType{ElemType: types.StringType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: FromToObjectType},
	"host":                types.SetType{ElemType: NameIDObjectType},
	"group":               types.SetType{ElemType: NameIDObjectType},
	"system_group":        types.SetType{ElemType: NameIDObjectType},
	"network_interface":   types.SetType{ElemType: NameIDObjectType},
	"global_ip_range":     types.SetType{ElemType: NameIDObjectType},
	"floating_subnet":     types.SetType{ElemType: NameIDObjectType},
	"site_network_subnet": types.SetType{ElemType: NameIDObjectType},
}

var SocketLanServiceObjectType = types.ObjectType{AttrTypes: SocketLanServiceAttrTypes}
var SocketLanServiceAttrTypes = map[string]attr.Type{
	"simple": types.SetType{ElemType: SimpleServiceObjectType},
	"custom": types.ListType{ElemType: CustomServiceObjectType},
}

var SimpleServiceObjectType = types.ObjectType{AttrTypes: SimpleServiceAttrTypes}
var SimpleServiceAttrTypes = map[string]attr.Type{
	"name": types.StringType,
}

var SocketLanNatObjectType = types.ObjectType{AttrTypes: SocketLanNatAttrTypes}
var SocketLanNatAttrTypes = map[string]attr.Type{
	"enabled":  types.BoolType,
	"nat_type": types.StringType,
}
