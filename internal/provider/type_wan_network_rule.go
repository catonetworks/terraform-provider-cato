//nolint:lll
package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WanNetworkRule represents the top-level resource
type WanNetworkRule struct {
	Rule types.Object `tfsdk:"rule" json:"rule,omitempty"` // PolicyPolicyWanNetworkPolicyRulesRule
	At   types.Object `tfsdk:"at" json:"at,omitempty"`     // *PolicyRulePositionInput
}

// PolicyPolicyWanNetworkPolicyRulesRule represents a WAN Network rule
type PolicyPolicyWanNetworkPolicyRulesRule struct {
	ID                types.String `tfsdk:"id" json:"id,omitempty"`
	Name              types.String `tfsdk:"name" json:"name,omitempty"`
	Description       types.String `tfsdk:"description" json:"description,omitempty"`
	Enabled           types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	RuleType          types.String `tfsdk:"rule_type" json:"ruleType,omitempty"`
	RouteType         types.String `tfsdk:"route_type" json:"routeType,omitempty"`
	Source            types.Object `tfsdk:"source" json:"source,omitempty"`                        // PolicyPolicyWanNetworkPolicyRulesRuleSource
	Destination       types.Object `tfsdk:"destination" json:"destination,omitempty"`              // PolicyPolicyWanNetworkPolicyRulesRuleDestination
	Application       types.Object `tfsdk:"application" json:"application,omitempty"`              // PolicyPolicyWanNetworkPolicyRulesRuleApplication
	Configuration     types.Object `tfsdk:"configuration" json:"configuration,omitempty"`          // PolicyPolicyWanNetworkPolicyRulesRuleConfiguration
	BandwidthPriority types.Object `tfsdk:"bandwidth_priority" json:"bandwidthPriority,omitempty"` // PolicyPolicyWanNetworkPolicyRulesRuleBandwidthPriority
	Exceptions        types.Set    `tfsdk:"exceptions" json:"exceptions,omitempty"`                // []*PolicyPolicyWanNetworkPolicyRulesRuleExceptions
}

// PolicyPolicyWanNetworkPolicyRulesRuleSource represents source matching criteria
type PolicyPolicyWanNetworkPolicyRulesRuleSource struct {
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"`
	Site              types.Set  `tfsdk:"site" json:"site,omitempty"`
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`
	User              types.Set  `tfsdk:"user" json:"user,omitempty"`
	UsersGroup        types.Set  `tfsdk:"users_group" json:"users_group,omitempty"`
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceHost struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceSite struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceIPRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceGlobalIPRange struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceNetworkInterface struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceSiteNetworkSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceFloatingSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceUser struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceUsersGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleSourceSystemGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleDestination represents destination matching criteria
type PolicyPolicyWanNetworkPolicyRulesRuleDestination struct {
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"`
	Site              types.Set  `tfsdk:"site" json:"site,omitempty"`
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`
	User              types.Set  `tfsdk:"user" json:"user,omitempty"`
	UsersGroup        types.Set  `tfsdk:"users_group" json:"users_group,omitempty"`
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationHost struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationSite struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationIPRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationGlobalIPRange struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationNetworkInterface struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationSiteNetworkSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationFloatingSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationUser struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationUsersGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleDestinationSystemGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleApplication represents application matching criteria
type PolicyPolicyWanNetworkPolicyRulesRuleApplication struct {
	Application     types.Set  `tfsdk:"application" json:"application,omitempty"`
	CustomApp       types.Set  `tfsdk:"custom_app" json:"custom_app,omitempty"`
	AppCategory     types.Set  `tfsdk:"app_category" json:"app_category,omitempty"`
	CustomCategory  types.Set  `tfsdk:"custom_category" json:"custom_category,omitempty"`
	Domain          types.List `tfsdk:"domain" json:"domain,omitempty"`
	Fqdn            types.List `tfsdk:"fqdn" json:"fqdn,omitempty"`
	Service         types.Set  `tfsdk:"service" json:"service,omitempty"`
	CustomService   types.List `tfsdk:"custom_service" json:"custom_service,omitempty"`
	CustomServiceIP types.List `tfsdk:"custom_service_ip" json:"custom_service_ip,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomService represents custom service definition
type PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomService struct {
	Port      types.List   `tfsdk:"port" json:"port,omitempty"`
	PortRange types.Object `tfsdk:"port_range" json:"port_range,omitempty"`
	Protocol  types.String `tfsdk:"protocol" json:"protocol,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServiceIP represents custom service IP definition
type PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServiceIP struct {
	Name    types.String `tfsdk:"name" json:"name,omitempty"`
	IP      types.String `tfsdk:"ip" json:"ip,omitempty"`
	IPRange types.Object `tfsdk:"ip_range" json:"ip_range,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleApplicationApplication represents an application ID/name reference object.
type PolicyPolicyWanNetworkPolicyRulesRuleApplicationApplication struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomApp struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleApplicationAppCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleApplicationService struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServicePortRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServiceIPIPRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleConfiguration represents WAN network configuration
type PolicyPolicyWanNetworkPolicyRulesRuleConfiguration struct {
	ActiveTCPAcceleration bool         `tfsdk:"active_tcp_acceleration" json:"activeTcpAcceleration,omitempty"`
	PacketLossMitigation  bool         `tfsdk:"packet_loss_mitigation" json:"packetLossMitigation,omitempty"`
	PreserveSourcePort    bool         `tfsdk:"preserve_source_port" json:"preserveSourcePort,omitempty"`
	PrimaryTransport      types.Object `tfsdk:"primary_transport" json:"primaryTransport,omitempty"`
	SecondaryTransport    types.Object `tfsdk:"secondary_transport" json:"secondaryTransport,omitempty"`
	AllocationIP          types.Set    `tfsdk:"allocation_ip" json:"allocationIp,omitempty"`
	PopLocation           types.Set    `tfsdk:"pop_location" json:"popLocation,omitempty"`
	BackhaulingSite       types.Set    `tfsdk:"backhauling_site" json:"backhaulingSite,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleConfigurationTransport represents transport configuration
type PolicyPolicyWanNetworkPolicyRulesRuleConfigurationTransport struct {
	TransportType          types.String `tfsdk:"transport_type" json:"transportType,omitempty"`
	PrimaryInterfaceRole   types.String `tfsdk:"primary_interface_role" json:"primaryInterfaceRole,omitempty"`
	SecondaryInterfaceRole types.String `tfsdk:"secondary_interface_role" json:"secondaryInterfaceRole,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleBandwidthPriority represents bandwidth priority
type PolicyPolicyWanNetworkPolicyRulesRuleBandwidthPriority struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleConfigurationAllocationIP represents an allocation IP ID/name reference object.
type PolicyPolicyWanNetworkPolicyRulesRuleConfigurationAllocationIP struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleConfigurationPopLocation struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanNetworkPolicyRulesRuleConfigurationBackhaulingSite struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

// PolicyPolicyWanNetworkPolicyRulesRuleExceptions represents rule exceptions
type PolicyPolicyWanNetworkPolicyRulesRuleExceptions struct {
	Name        types.String `tfsdk:"name" json:"name,omitempty"`
	Source      types.Object `tfsdk:"source" json:"source,omitempty"`
	Destination types.Object `tfsdk:"destination" json:"destination,omitempty"`
	Application types.Object `tfsdk:"application" json:"application,omitempty"`
}

// AttrTypes maps for ObjectType definitions

var WanNetworkRuleObjectType = types.ObjectType{AttrTypes: WanNetworkRuleAttrTypes}
var WanNetworkRuleAttrTypes = map[string]attr.Type{
	"rule": WanNetworkRuleRuleObjectType,
	"at":   PositionObjectType,
}

var WanNetworkRuleRuleObjectType = types.ObjectType{AttrTypes: WanNetworkRuleRuleAttrTypes}
var WanNetworkRuleRuleAttrTypes = map[string]attr.Type{
	"id":                 types.StringType,
	"name":               types.StringType,
	"description":        types.StringType,
	"enabled":            types.BoolType,
	"rule_type":          types.StringType,
	"route_type":         types.StringType,
	"source":             WanNetworkSourceObjectType,
	"destination":        WanNetworkDestObjectType,
	"application":        WanNetworkApplicationObjectType,
	"configuration":      WanNetworkConfigurationObjectType,
	"bandwidth_priority": BandwidthPriorityObjectType,
	"exceptions":         types.SetType{ElemType: WanNetworkExceptionObjectType},
}

var WanNetworkSourceObjectType = types.ObjectType{AttrTypes: WanNetworkSourceAttrTypes}
var WanNetworkSourceAttrTypes = map[string]attr.Type{
	"ip":                  types.ListType{ElemType: types.StringType},
	"host":                types.SetType{ElemType: NameIDObjectType},
	"site":                types.SetType{ElemType: NameIDObjectType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: FromToObjectType},
	"global_ip_range":     types.SetType{ElemType: NameIDObjectType},
	"network_interface":   types.SetType{ElemType: NameIDObjectType},
	"site_network_subnet": types.SetType{ElemType: NameIDObjectType},
	"floating_subnet":     types.SetType{ElemType: NameIDObjectType},
	"user":                types.SetType{ElemType: NameIDObjectType},
	"users_group":         types.SetType{ElemType: NameIDObjectType},
	"group":               types.SetType{ElemType: NameIDObjectType},
	"system_group":        types.SetType{ElemType: NameIDObjectType},
}

var WanNetworkDestObjectType = types.ObjectType{AttrTypes: WanNetworkDestAttrTypes}
var WanNetworkDestAttrTypes = map[string]attr.Type{
	"ip":                  types.ListType{ElemType: types.StringType},
	"host":                types.SetType{ElemType: NameIDObjectType},
	"site":                types.SetType{ElemType: NameIDObjectType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: FromToObjectType},
	"global_ip_range":     types.SetType{ElemType: NameIDObjectType},
	"network_interface":   types.SetType{ElemType: NameIDObjectType},
	"site_network_subnet": types.SetType{ElemType: NameIDObjectType},
	"floating_subnet":     types.SetType{ElemType: NameIDObjectType},
	"user":                types.SetType{ElemType: NameIDObjectType},
	"users_group":         types.SetType{ElemType: NameIDObjectType},
	"group":               types.SetType{ElemType: NameIDObjectType},
	"system_group":        types.SetType{ElemType: NameIDObjectType},
}

var WanNetworkApplicationObjectType = types.ObjectType{AttrTypes: WanNetworkApplicationAttrTypes}
var WanNetworkApplicationAttrTypes = map[string]attr.Type{
	"application":       types.SetType{ElemType: NameIDObjectType},
	"custom_app":        types.SetType{ElemType: NameIDObjectType},
	"app_category":      types.SetType{ElemType: NameIDObjectType},
	"custom_category":   types.SetType{ElemType: NameIDObjectType},
	"domain":            types.ListType{ElemType: types.StringType},
	"fqdn":              types.ListType{ElemType: types.StringType},
	"service":           types.SetType{ElemType: NameIDObjectType},
	"custom_service":    types.ListType{ElemType: WanNetworkCustomServiceObjectType},
	"custom_service_ip": types.ListType{ElemType: WanNetworkCustomServiceIPObjectType},
}

// WanNetworkExceptionApplicationObjectType is used for exceptions (matches WAN Network application schema)
var WanNetworkExceptionApplicationObjectType = types.ObjectType{AttrTypes: WanNetworkExceptionApplicationAttrTypes}
var WanNetworkExceptionApplicationAttrTypes = map[string]attr.Type{
	"application":       types.SetType{ElemType: NameIDObjectType},
	"custom_app":        types.SetType{ElemType: NameIDObjectType},
	"app_category":      types.SetType{ElemType: NameIDObjectType},
	"custom_category":   types.SetType{ElemType: NameIDObjectType},
	"domain":            types.ListType{ElemType: types.StringType},
	"fqdn":              types.ListType{ElemType: types.StringType},
	"service":           types.SetType{ElemType: NameIDObjectType},
	"custom_service":    types.ListType{ElemType: WanNetworkCustomServiceObjectType},
	"custom_service_ip": types.ListType{ElemType: WanNetworkCustomServiceIPObjectType},
}

var WanNetworkCustomServiceObjectType = types.ObjectType{AttrTypes: WanNetworkCustomServiceAttrTypes}
var WanNetworkCustomServiceAttrTypes = map[string]attr.Type{
	"port":       types.ListType{ElemType: types.StringType},
	"port_range": FromToObjectType,
	"protocol":   types.StringType,
}

var WanNetworkCustomServiceIPObjectType = types.ObjectType{AttrTypes: WanNetworkCustomServiceIPAttrTypes}
var WanNetworkCustomServiceIPAttrTypes = map[string]attr.Type{
	"name":     types.StringType,
	"ip":       types.StringType,
	"ip_range": FromToObjectType,
}

var WanNetworkConfigurationObjectType = types.ObjectType{AttrTypes: WanNetworkConfigurationAttrTypes}
var WanNetworkConfigurationAttrTypes = map[string]attr.Type{
	"active_tcp_acceleration": types.BoolType,
	"packet_loss_mitigation":  types.BoolType,
	"preserve_source_port":    types.BoolType,
	"primary_transport":       WanNetworkTransportObjectType,
	"secondary_transport":     WanNetworkTransportObjectType,
	"allocation_ip":           types.SetType{ElemType: NameIDObjectType},
	"pop_location":            types.SetType{ElemType: NameIDObjectType},
	"backhauling_site":        types.SetType{ElemType: NameIDObjectType},
}

var WanNetworkTransportObjectType = types.ObjectType{AttrTypes: WanNetworkTransportAttrTypes}
var WanNetworkTransportAttrTypes = map[string]attr.Type{
	"transport_type":           types.StringType,
	"primary_interface_role":   types.StringType,
	"secondary_interface_role": types.StringType,
}

var BandwidthPriorityObjectType = types.ObjectType{AttrTypes: BandwidthPriorityAttrTypes}
var BandwidthPriorityAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

var WanNetworkExceptionObjectType = types.ObjectType{AttrTypes: WanNetworkExceptionAttrTypes}
var WanNetworkExceptionAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"source":      WanNetworkSourceObjectType,
	"destination": WanNetworkDestObjectType,
	"application": WanNetworkExceptionApplicationObjectType,
}
