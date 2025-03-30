package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InternetFirewallRule struct {
	Rule types.Object `tfsdk:"rule" json:"rule,omitempty"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule
	At   types.Object `tfsdk:"at" json:"at,omitempty"`     //*PolicyRulePositionInput
}

type PolicyRulePositionInput struct {
	Position types.String `tfsdk:"position" json:"position,omitempty"`
	Ref      types.String `tfsdk:"ref" json:"ref,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule struct {
	ID               types.String `tfsdk:"id" json:"id,omitempty"`
	Name             types.String `tfsdk:"name" json:"name,omitempty"`
	Description      types.String `tfsdk:"description" json:"description,omitempty"`
	Index            types.Int64  `tfsdk:"index" json:"index,omitempty"`
	Enabled          types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	Section          types.Object `tfsdk:"section" json:"section,omitempty"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Section
	Source           types.Object `tfsdk:"source" json:"source,omitempty"`   //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source
	ConnectionOrigin types.String `tfsdk:"connection_origin" json:"connection_origin,omitempty"`
	Country          types.List   `tfsdk:"country" json:"country,omitempty"` //[]Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country
	Device           types.List   `tfsdk:"device" json:"device,omitempty"`   //[]Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device
	DeviceOs         types.List   `tfsdk:"device_os" json:"device_os,omitempty"`
	Destination      types.Object `tfsdk:"destination" json:"destination,omitempty"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination
	Service          types.Object `tfsdk:"service" json:"service,omitempty"`         //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service
	Action           types.String `tfsdk:"action" json:"action,omitempty"`
	Tracking         types.Object `tfsdk:"tracking" json:"tracking,omitempty"`     //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking
	Schedule         types.Object `tfsdk:"schedule" json:"schedule,omitempty"`     //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule
	Exceptions       types.List   `tfsdk:"exceptions" json:"exceptions,omitempty"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Section struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination struct {
	Application            types.List `tfsdk:"application" json:"application,omitempty"`                           //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application
	CustomApp              types.List `tfsdk:"custom_app" json:"custom_app,omitempty"`                             //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp
	AppCategory            types.List `tfsdk:"app_category" json:"app_category,omitempty"`                         //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory
	CustomCategory         types.List `tfsdk:"custom_category" json:"custom_category,omitempty"`                   //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory
	SanctionedAppsCategory types.List `tfsdk:"sanctioned_apps_category" json:"sanctioned_apps_category,omitempty"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory
	Country                types.List `tfsdk:"country" json:"country,omitempty"`                                   //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country
	Domain                 types.List `tfsdk:"domain" json:"domain,omitempty"`
	Fqdn                   types.List `tfsdk:"fqdn" json:"fqdn,omitempty"`
	IP                     types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet                 types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange                types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`               //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
	GlobalIPRange          types.List `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
	RemoteAsn              types.List `tfsdk:"remote_asn" json:"remote_asn,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source struct {
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Host              types.List `tfsdk:"host" json:"host,omitempty"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host
	Site              types.List `tfsdk:"site" json:"site,omitempty"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`                       //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange
	GlobalIPRange     types.List `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`         //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange
	NetworkInterface  types.List `tfsdk:"network_interface" json:"network_interface,omitempty"`     //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface
	SiteNetworkSubnet types.List `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
	FloatingSubnet    types.List `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`         //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet
	User              types.List `tfsdk:"user" json:"user,omitempty"`                               //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User
	UsersGroup        types.List `tfsdk:"users_group" json:"users_group,omitempty"`                 //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup
	Group             types.List `tfsdk:"group" json:"group,omitempty"`                             //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group
	SystemGroup       types.List `tfsdk:"system_group" json:"system_group,omitempty"`               //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service struct {
	Standard types.List `tfsdk:"standard" json:"standard,omitempty"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard
	Custom   types.List `tfsdk:"custom" json:"custom,omitempty"`     //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom struct {
	Port      types.List   `tfsdk:"port" json:"port,omitempty"`
	PortRange types.Object `tfsdk:"port_range" json:"port_range,omitempty"` //*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange
	Protocol  types.String `tfsdk:"protocol" json:"protocol,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking struct {
	Event types.Object `tfsdk:"event" json:"event,omitempty"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event
	Alert types.Object `tfsdk:"alert" json:"alert,omitempty"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event struct {
	Enabled types.Bool `tfsdk:"enabled" json:"enabled,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert struct {
	Enabled           types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	Frequency         types.String `tfsdk:"frequency" json:"frequency,omitempty"`
	SubscriptionGroup types.List   `tfsdk:"subscription_group" json:"subscription_group,omitempty"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
	Webhook           types.List   `tfsdk:"webhook" json:"webhook,omitempty"`                       //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_Webhook
	MailingList       types.List   `tfsdk:"mailing_list" json:"mailing_list,omitempty"`             //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_MailingList
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_Webhook struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_MailingList struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule struct {
	ActiveOn        types.String `tfsdk:"active_on" json:"active_on,omitempty"`
	CustomTimeframe types.Object `tfsdk:"custom_timeframe" json:"custom_timeframe,omitempty"` //*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe
	CustomRecurring types.Object `tfsdk:"custom_recurring" json:"custom_recurring,omitempty"` //*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomRecurring
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomRecurring struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
	Days types.List   `tfsdk:"days" json:"days,omitempty"` //[]DayOfWeek
}

type DayOfWeek types.String

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions struct {
	Name             types.String `tfsdk:"name" json:"name,omitempty"` ///////
	Source           types.Object `tfsdk:"source" json:"source,omitempty"`
	ConnectionOrigin types.String `tfsdk:"connection_origin" json:"connection_origin,omitempty"` ///////
	Country          types.List   `tfsdk:"country" json:"country,omitempty"`
	Device           types.List   `tfsdk:"device" json:"device,omitempty"`
	DeviceOs         types.List   `tfsdk:"device_os" json:"device_os,omitempty"`
	Destination      types.Object `tfsdk:"destination" json:"destination,omitempty"`
	Service          types.Object `tfsdk:"service" json:"service,omitempty"`
}

type OperatingSystem types.String

// Generic object types used to write back to state
var InternetFirewallRuleObjectType = types.ObjectType{AttrTypes: InternetFirewallRuleAttrTypes}
var InternetFirewallRuleAttrTypes = map[string]attr.Type{
	"rule": InternetFirewallRuleRuleObjectType,
	"at":   PositionObjectType,
}

var PositionObjectType = types.ObjectType{AttrTypes: PositionAttrTypes}
var PositionAttrTypes = map[string]attr.Type{
	"position": types.StringType,
	"ref":      types.StringType,
}

var InternetFirewallRuleRuleObjectType = types.ObjectType{AttrTypes: InternetFirewallRuleRuleAttrTypes}
var InternetFirewallRuleRuleAttrTypes = map[string]attr.Type{
	"id":                types.StringType,
	"name":              types.StringType,
	"description":       types.StringType,
	"index":             types.Int64Type,
	"enabled":           types.BoolType,
	"section":           NameIDObjectType,
	"source":            SourceObjectType,
	"connection_origin": types.StringType,
	"country":           types.ListType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device":            types.ListType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device_os":         types.ListType{ElemType: types.StringType},
	"destination":       types.ObjectType{AttrTypes: DestAttrTypes},
	"service":           types.ObjectType{AttrTypes: ServiceAttrTypes},
	"action":            types.StringType,
	"tracking":          TrackingObjectType,
	"schedule":          ScheduleObjectType,
	"exceptions":        types.ListType{ElemType: types.ObjectType{AttrTypes: ExceptionAttrTypes}},
}

var ServiceObjectType = types.ObjectType{AttrTypes: ServiceAttrTypes}
var ServiceAttrTypes = map[string]attr.Type{
	"standard": types.ListType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"custom":   types.ListType{ElemType: types.ObjectType{AttrTypes: CustomServiceAttrTypes}},
}
var CustomServiceObjectType = types.ObjectType{AttrTypes: CustomServiceAttrTypes}
var CustomServiceAttrTypes = map[string]attr.Type{
	"port":       types.ListType{ElemType: types.StringType},
	"port_range": FromToObjectType,
	"protocol":   types.StringType,
}
var NameIDObjectType = types.ObjectType{AttrTypes: NameIDAttrTypes}
var NameIDAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"id":   types.StringType,
}
var FromToObjectType = types.ObjectType{AttrTypes: FromToAttrTypes}
var FromToAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}
var FromToDaysObjectType = types.ObjectType{AttrTypes: FromToAttrTypes}
var FromToDaysAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
	"days": types.ListType{ElemType: types.StringType},
}

// Rule -> Tracking
var TrackingObjectType = types.ObjectType{AttrTypes: TrackingAttrTypes}
var TrackingAttrTypes = map[string]attr.Type{
	"event": types.ObjectType{AttrTypes: TrackingEventAttrTypes},
	"alert": types.ObjectType{AttrTypes: TrackingAlertAttrTypes},
}

var TrackingEventObjectType = types.ObjectType{AttrTypes: TrackingAttrTypes}
var TrackingEventAttrTypes = map[string]attr.Type{
	"enabled": types.BoolType,
}
var TrackingAlertObjectType = types.ObjectType{AttrTypes: TrackingAttrTypes}
var TrackingAlertAttrTypes = map[string]attr.Type{
	"enabled":            types.BoolType,
	"frequency":          types.StringType,
	"subscription_group": types.ListType{ElemType: NameIDObjectType},
	"webhook":            types.ListType{ElemType: NameIDObjectType},
	"mailing_list":       types.ListType{ElemType: NameIDObjectType},
}

var ScheduleObjectType = types.ObjectType{AttrTypes: ScheduleAttrTypes}
var ScheduleAttrTypes = map[string]attr.Type{
	"active_on":        types.StringType,
	"custom_timeframe": types.ObjectType{AttrTypes: FromToAttrTypes},
	"custom_recurring": types.ObjectType{AttrTypes: FromToDaysAttrTypes},
}

var SourceObjectType = types.ObjectType{AttrTypes: SourceAttrTypes}
var SourceAttrTypes = map[string]attr.Type{
	"ip":                  types.ListType{ElemType: types.StringType},
	"host":                types.ListType{ElemType: NameIDObjectType},
	"site":                types.ListType{ElemType: NameIDObjectType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: FromToObjectType},
	"global_ip_range":     types.ListType{ElemType: NameIDObjectType},
	"network_interface":   types.ListType{ElemType: NameIDObjectType},
	"site_network_subnet": types.ListType{ElemType: NameIDObjectType},
	"floating_subnet":     types.ListType{ElemType: NameIDObjectType},
	"user":                types.ListType{ElemType: NameIDObjectType},
	"users_group":         types.ListType{ElemType: NameIDObjectType},
	"group":               types.ListType{ElemType: NameIDObjectType},
	"system_group":        types.ListType{ElemType: NameIDObjectType},
}

var DestObjectType = types.ObjectType{AttrTypes: DestAttrTypes}
var DestAttrTypes = map[string]attr.Type{
	"application":              types.ListType{ElemType: NameIDObjectType},
	"custom_app":               types.ListType{ElemType: NameIDObjectType},
	"app_category":             types.ListType{ElemType: NameIDObjectType},
	"custom_category":          types.ListType{ElemType: NameIDObjectType},
	"sanctioned_apps_category": types.ListType{ElemType: NameIDObjectType},
	"country":                  types.ListType{ElemType: NameIDObjectType},
	"domain":                   types.ListType{ElemType: types.StringType},
	"fqdn":                     types.ListType{ElemType: types.StringType},
	"ip":                       types.ListType{ElemType: types.StringType},
	"subnet":                   types.ListType{ElemType: types.StringType},
	"ip_range":                 types.ListType{ElemType: FromToObjectType},
	"global_ip_range":          types.ListType{ElemType: NameIDObjectType},
	"remote_asn":               types.ListType{ElemType: types.StringType},
}

var ExceptionObjectType = types.ObjectType{AttrTypes: ExceptionAttrTypes}
var ExceptionAttrTypes = map[string]attr.Type{
	"name":    types.StringType,
	"source":  types.ObjectType{AttrTypes: SourceAttrTypes},
	"country": types.ListType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device":  types.ListType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	// "device_attributes": types.ObjectType{AttrTypes: DeviceAttrAttrTypes},
	"device_os":         types.ListType{ElemType: types.StringType},
	"destination":       types.ObjectType{AttrTypes: DestAttrTypes},
	"service":           types.ObjectType{AttrTypes: ServiceAttrTypes},
	"connection_origin": types.StringType,
}

var DeviceAttrObjectType = types.ObjectType{AttrTypes: DeviceAttrAttrTypes}
var DeviceAttrAttrTypes = map[string]attr.Type{
	"category":     types.ListType{ElemType: types.StringType},
	"type":         types.ListType{ElemType: types.StringType},
	"model":        types.ListType{ElemType: types.StringType},
	"manufacturer": types.ListType{ElemType: types.StringType},
	"os":           types.ListType{ElemType: types.StringType},
	"osVersion":    types.ListType{ElemType: types.StringType},
}
