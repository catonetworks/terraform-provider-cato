package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WanFirewallRule struct {
	Rule types.Object `tfsdk:"rule" json:"rule,omitempty"` //Policy_Policy_WanFirewall_Policy_Rules_Rule
	At   types.Object `tfsdk:"at" json:"at,omitempty"`     //*PolicyRulePositionInput
}

// type PolicyRulePositionInput struct {
// 	Position types.String `tfsdk:"position"`
// 	Ref      types.String `tfsdk:"ref"`
// }

type Policy_Policy_WanFirewall_Policy_Rules_Rule struct {
	ID          types.String `tfsdk:"id" json:"id,omitempty"`
	Name        types.String `tfsdk:"name" json:"name,omitempty"`
	Description types.String `tfsdk:"description" json:"description,omitempty"`
	Index       types.Int64  `tfsdk:"index" json:"index,omitempty"`
	Enabled     types.Bool   `tfsdk:"enabled" json:"enable,omitempty"`
	// Section          types.Object `tfsdk:"section" json:"section,omitempty"` //Policy_Policy_WanFirewall_Policy_Rules_Rule_Section
	Source           types.Object `tfsdk:"source" json:"source,omitempty"` //Policy_Policy_WanFirewall_Policy_Rules_Rule_Source
	ConnectionOrigin types.String `tfsdk:"connection_origin" json:"connection_origin,omitempty"`
	Country          types.Set    `tfsdk:"country" json:"country,omitempty"` //[]Policy_Policy_WanFirewall_Policy_Rules_Rule_Country
	Device           types.Set    `tfsdk:"device" json:"device,omitempty"`   //[]Policy_Policy_WanFirewall_Policy_Rules_Rule_Device
	DeviceOs         types.List   `tfsdk:"device_os" json:"device_os,omitempty"`
	Destination      types.Object `tfsdk:"destination" json:"destination,omitempty"` //Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination
	Application      types.Object `tfsdk:"application" json:"application,omitempty"` //Policy_Policy_WanFirewall_Policy_Rules_Rule_Application
	Service          types.Object `tfsdk:"service" json:"service,omitempty"`         //Policy_Policy_WanFirewall_Policy_Rules_Rule_Service
	Action           types.String `tfsdk:"action" json:"action,omitempty"`
	Tracking         types.Object `tfsdk:"tracking" json:"tracking,omitempty"` //Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking
	Schedule         types.Object `tfsdk:"schedule" json:"schedule,omitempty"` //Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule
	Direction        types.String `tfsdk:"direction" json:"direction,omitempty"`
	Exceptions       types.Set    `tfsdk:"exceptions" json:"exceptions,omitempty"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Exceptions
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Country struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Device struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Section struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Application struct {
	Application            types.Set  `tfsdk:"application" json:"application,omitempty"`                           //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_Application
	CustomApp              types.Set  `tfsdk:"custom_app" json:"custom_app,omitempty"`                             //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomApp
	AppCategory            types.Set  `tfsdk:"app_category" json:"app_category,omitempty"`                         //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_AppCategory
	CustomCategory         types.Set  `tfsdk:"custom_category" json:"custom_category,omitempty"`                   //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomCategory
	SanctionedAppsCategory types.Set  `tfsdk:"sanctioned_apps_category" json:"sanctioned_apps_category,omitempty"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_SanctionedAppsCategory
	Domain                 types.List `tfsdk:"domain" json:"domain,omitempty"`
	Fqdn                   types.List `tfsdk:"fqdn" json:"fqdn,omitempty"`
	IP                     types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet                 types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange                types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`               //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_IPRange
	GlobalIPRange          types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_GlobalIPRange
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_Application struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomApp struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_AppCategory struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomCategory struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_SanctionedAppsCategory struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_IPRange struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_GlobalIPRange struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source struct {
	IP                types.List `tfsdk:"ip"`
	Host              types.Set  `tfsdk:"host"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Host
	Site              types.Set  `tfsdk:"site"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Site
	Subnet            types.List `tfsdk:"subnet"`
	IPRange           types.List `tfsdk:"ip_range"`            //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_IPRange
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range"`     //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_GlobalIPRange
	NetworkInterface  types.Set  `tfsdk:"network_interface"`   //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_NetworkInterface
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet"`     //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_FloatingSubnet
	User              types.Set  `tfsdk:"user"`                //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_User
	UsersGroup        types.Set  `tfsdk:"users_group"`         //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_UsersGroup
	Group             types.Set  `tfsdk:"group"`               //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Group
	SystemGroup       types.Set  `tfsdk:"system_group"`        //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SystemGroup
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Host struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Site struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_IPRange struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_GlobalIPRange struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_NetworkInterface struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_User struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_UsersGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Group struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SystemGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_FloatingSubnet struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination struct {
	IP                types.List `tfsdk:"ip"`
	Host              types.Set  `tfsdk:"host"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Host
	Site              types.Set  `tfsdk:"site"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Site
	Subnet            types.List `tfsdk:"subnet"`
	IPRange           types.List `tfsdk:"ip_range"`            //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_IPRange
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range"`     //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
	NetworkInterface  types.Set  `tfsdk:"network_interface"`   //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_NetworkInterface
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SiteNetworkSubnet
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet"`     //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_FloatingSubnet
	User              types.Set  `tfsdk:"user"`                //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_User
	UsersGroup        types.Set  `tfsdk:"users_group"`         //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_UsersGroup
	Group             types.Set  `tfsdk:"group"`               //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Group
	SystemGroup       types.Set  `tfsdk:"system_group"`        //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SystemGroup
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Host struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Site struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_IPRange struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_GlobalIPRange struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_NetworkInterface struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SiteNetworkSubnet struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_User struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_UsersGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Group struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SystemGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_FloatingSubnet struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Service struct {
	Standard types.Set  `tfsdk:"standard"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Standard
	Custom   types.List `tfsdk:"custom"`   //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Standard struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom struct {
	Port      types.List   `tfsdk:"port"`
	PortRange types.Object `tfsdk:"port_range"` //*Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom_PortRange
	Protocol  types.String `tfsdk:"protocol"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom_PortRange struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking struct {
	Event types.Object `tfsdk:"event"` //Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Event
	Alert types.Object `tfsdk:"alert"` //Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Event struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	Frequency         types.String `tfsdk:"frequency"`
	SubscriptionGroup types.Set    `tfsdk:"subscription_group"` //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
	Webhook           types.Set    `tfsdk:"webhook"`            //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_Webhook
	MailingList       types.Set    `tfsdk:"mailing_list"`       //[]*Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_MailingList
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_Webhook struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_MailingList struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule struct {
	ActiveOn        types.String `tfsdk:"active_on"`
	CustomTimeframe types.Object `tfsdk:"custom_timeframe"` //*Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe
	CustomRecurring types.Object `tfsdk:"custom_recurring"` //*Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomRecurring
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomRecurring struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
	Days types.List   `tfsdk:"days"` //[]DayOfWeek
}

type Policy_Policy_WanFirewall_Policy_Rules_Rule_Exceptions struct {
	Name             types.String `tfsdk:"name"`
	Source           types.Object `tfsdk:"source"`
	ConnectionOrigin types.String `tfsdk:"connection_origin"`
	Country          types.Set    `tfsdk:"country"`
	Device           types.Set    `tfsdk:"device"`
	DeviceOs         types.List   `tfsdk:"device_os"`
	Destination      types.Object `tfsdk:"destination"`
	Application      types.Object `tfsdk:"application"`
	Service          types.Object `tfsdk:"service"`
	Direction        types.String `tfsdk:"direction"`
}

// Generic object types used to write back to state
var WanFirewallRuleObjectType = types.ObjectType{AttrTypes: InternetFirewallRuleAttrTypes}
var WanFirewallRuleAttrTypes = map[string]attr.Type{
	"rule": WanFirewallRuleRuleObjectType,
	"at":   PositionObjectType,
}

var WanFirewallRuleRuleObjectType = types.ObjectType{AttrTypes: InternetFirewallRuleRuleAttrTypes}
var WanFirewallRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"index":       types.Int64Type,
	"enabled":     types.BoolType,
	// "section":           NameIDObjectType,
	"source":            WanSourceObjectType,
	"connection_origin": types.StringType,
	"country":           types.SetType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device":            types.SetType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device_os":         types.ListType{ElemType: types.StringType},
	"application":       WanApplicationObjectType,
	"destination":       types.ObjectType{AttrTypes: WanDestAttrTypes},
	"service":           types.ObjectType{AttrTypes: WanServiceAttrTypes},
	"action":            types.StringType,
	"tracking":          TrackingObjectType,
	"schedule":          ScheduleObjectType,
	"direction":         types.StringType,
	"exceptions":        types.SetType{ElemType: types.ObjectType{AttrTypes: WanExceptionAttrTypes}},
}

var WanServiceObjectType = types.ObjectType{AttrTypes: WanServiceAttrTypes}
var WanServiceAttrTypes = map[string]attr.Type{
	"standard": types.SetType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"custom":   types.ListType{ElemType: types.ObjectType{AttrTypes: CustomServiceAttrTypes}},
}

var WanSourceObjectType = types.ObjectType{AttrTypes: WanSourceAttrTypes}
var WanSourceAttrTypes = map[string]attr.Type{
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

var WanDestObjectType = types.ObjectType{AttrTypes: WanDestAttrTypes}
var WanDestAttrTypes = map[string]attr.Type{
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

var WanApplicationObjectType = types.ObjectType{AttrTypes: WanSourceAttrTypes}
var WanApplicationAttrTypes = map[string]attr.Type{
	"application":              types.SetType{ElemType: NameIDObjectType},
	"custom_app":               types.SetType{ElemType: NameIDObjectType},
	"app_category":             types.SetType{ElemType: NameIDObjectType},
	"custom_category":          types.SetType{ElemType: NameIDObjectType},
	"sanctioned_apps_category": types.SetType{ElemType: NameIDObjectType},
	"domain":                   types.ListType{ElemType: types.StringType},
	"fqdn":                     types.ListType{ElemType: types.StringType},
	"ip":                       types.ListType{ElemType: types.StringType},
	"subnet":                   types.ListType{ElemType: types.StringType},
	"ip_range":                 types.ListType{ElemType: FromToObjectType},
	"global_ip_range":          types.SetType{ElemType: NameIDObjectType},
}

var WanExceptionObjectType = types.ObjectType{AttrTypes: WanExceptionAttrTypes}
var WanExceptionAttrTypes = map[string]attr.Type{
	"name":    types.StringType,
	"source":  types.ObjectType{AttrTypes: WanSourceAttrTypes},
	"country": types.SetType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device":  types.SetType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	// "device_attributes": types.ObjectType{AttrTypes: DeviceAttrAttrTypes},
	"device_os":         types.ListType{ElemType: types.StringType},
	"destination":       types.ObjectType{AttrTypes: WanDestAttrTypes},
	"application":       types.ObjectType{AttrTypes: WanApplicationAttrTypes},
	"service":           types.ObjectType{AttrTypes: WanServiceAttrTypes},
	"direction":         types.StringType,
	"connection_origin": types.StringType,
}

var WanDeviceAttrObjectType = types.ObjectType{AttrTypes: WanDeviceAttrAttrTypes}
var WanDeviceAttrAttrTypes = map[string]attr.Type{
	"category":     types.ListType{ElemType: types.StringType},
	"type":         types.ListType{ElemType: types.StringType},
	"model":        types.ListType{ElemType: types.StringType},
	"manufacturer": types.ListType{ElemType: types.StringType},
	"os":           types.ListType{ElemType: types.StringType},
	"osVersion":    types.ListType{ElemType: types.StringType},
}
