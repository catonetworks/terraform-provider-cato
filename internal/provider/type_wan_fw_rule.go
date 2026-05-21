//nolint:lll
package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WanFirewallRule struct {
	Rule types.Object `tfsdk:"rule" json:"rule,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRule
	At   types.Object `tfsdk:"at" json:"at,omitempty"`     // *PolicyRulePositionInput
}

// type PolicyRulePositionInput struct {
// 	Position types.String `tfsdk:"position"`
// 	Ref      types.String `tfsdk:"ref"`
// }

type PolicyPolicyWanFirewallPolicyRulesRule struct {
	ID          types.String `tfsdk:"id" json:"id,omitempty"`
	Name        types.String `tfsdk:"name" json:"name,omitempty"`
	Description types.String `tfsdk:"description" json:"description,omitempty"`
	Enabled     types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	// Section          types.Object `tfsdk:"section" json:"section,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRuleSection
	Source           types.Object `tfsdk:"source" json:"source,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRuleSource
	ConnectionOrigin types.String `tfsdk:"connection_origin" json:"connectionOrigin,omitempty"`
	ActivePeriod     types.Object `tfsdk:"active_period" json:"activePeriod,omitempty"`         // PolicyPolicyWanFirewallPolicyRulesRuleActivePeriod
	Country          types.Set    `tfsdk:"country" json:"country,omitempty"`                    // []PolicyPolicyWanFirewallPolicyRulesRuleCountry
	Device           types.Set    `tfsdk:"device" json:"device,omitempty"`                      // []PolicyPolicyWanFirewallPolicyRulesRuleDevice
	DeviceAttributes types.Object `tfsdk:"device_attributes" json:"deviceAttributes,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRuleDeviceAttributes
	DeviceOs         types.List   `tfsdk:"device_os" json:"deviceOS,omitempty"`
	Destination      types.Object `tfsdk:"destination" json:"destination,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRuleDestination
	Application      types.Object `tfsdk:"application" json:"application,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRuleApplication
	Service          types.Object `tfsdk:"service" json:"service,omitempty"`         // PolicyPolicyWanFirewallPolicyRulesRuleService
	Action           types.String `tfsdk:"action" json:"action,omitempty"`
	Tracking         types.Object `tfsdk:"tracking" json:"tracking,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRuleTracking
	Schedule         types.Object `tfsdk:"schedule" json:"schedule,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRuleSchedule
	Direction        types.String `tfsdk:"direction" json:"direction,omitempty"`
	Exceptions       types.Set    `tfsdk:"exceptions" json:"exceptions,omitempty"` // []*PolicyPolicyWanFirewallPolicyRulesRuleExceptions
}

type PolicyPolicyWanFirewallPolicyRulesRuleActivePeriod struct {
	EffectiveFrom    types.String `tfsdk:"effective_from" json:"effective_from,omitempty"`
	ExpiresAt        types.String `tfsdk:"expires_at" json:"expires_at,omitempty"`
	UseEffectiveFrom types.Bool   `tfsdk:"use_effective_from" json:"use_effective_from,omitempty"`
	UseExpiresAt     types.Bool   `tfsdk:"use_expires_at" json:"use_expires_at,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleCountry struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDevice struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDeviceAttributes struct {
	Category     types.List `tfsdk:"category" json:"category,omitempty"`
	Type         types.List `tfsdk:"type" json:"type,omitempty"`
	Model        types.List `tfsdk:"model" json:"model,omitempty"`
	Manufacturer types.List `tfsdk:"manufacturer" json:"manufacturer,omitempty"`
	Os           types.List `tfsdk:"os" json:"os,omitempty"`
	OsVersion    types.List `tfsdk:"os_version" json:"osVersion,omitempty"`
}

// DeviceAttributesInput struct for converting from tfsdk to cato_models
type DeviceAttributesInput struct {
	Category     []string `tfsdk:"category" json:"category"`
	Manufacturer []string `tfsdk:"manufacturer" json:"manufacturer"`
	Model        []string `tfsdk:"model" json:"model"`
	Os           []string `tfsdk:"os" json:"os"`
	OsVersion    []string `tfsdk:"os_version" json:"osVersion"`
	Type         []string `tfsdk:"type" json:"type"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSection struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleApplication struct {
	Application            types.Set  `tfsdk:"application" json:"application,omitempty"`                           // []*PolicyPolicyWanFirewallPolicyRulesRuleApplicationApplication
	CustomApp              types.Set  `tfsdk:"custom_app" json:"custom_app,omitempty"`                             // []*PolicyPolicyWanFirewallPolicyRulesRuleApplicationCustomApp
	AppCategory            types.Set  `tfsdk:"app_category" json:"app_category,omitempty"`                         // []*PolicyPolicyWanFirewallPolicyRulesRuleApplicationAppCategory
	CustomCategory         types.Set  `tfsdk:"custom_category" json:"custom_category,omitempty"`                   // []*PolicyPolicyWanFirewallPolicyRulesRuleApplicationCustomCategory
	SanctionedAppsCategory types.Set  `tfsdk:"sanctioned_apps_category" json:"sanctioned_apps_category,omitempty"` // []*PolicyPolicyWanFirewallPolicyRulesRuleApplicationSanctionedAppsCategory
	Domain                 types.List `tfsdk:"domain" json:"domain,omitempty"`
	Fqdn                   types.List `tfsdk:"fqdn" json:"fqdn,omitempty"`
	IP                     types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet                 types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange                types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`               // []*PolicyPolicyWanFirewallPolicyRulesRuleApplicationIPRange
	GlobalIPRange          types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"` // []*PolicyPolicyWanFirewallPolicyRulesRuleApplicationGlobalIPRange
}

type PolicyPolicyWanFirewallPolicyRulesRuleApplicationApplication struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleApplicationCustomApp struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleApplicationAppCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleApplicationCustomCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleApplicationSanctionedAppsCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleApplicationIPRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleApplicationGlobalIPRange struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSource struct {
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"` // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceHost
	Site              types.Set  `tfsdk:"site" json:"site,omitempty"` // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceSite
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`                       // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceIPRange
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`         // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceGlobalIPRange
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`     // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceNetworkInterface
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"` // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceSiteNetworkSubnet
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`         // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceFloatingSubnet
	User              types.Set  `tfsdk:"user" json:"user,omitempty"`                               // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceUser
	UsersGroup        types.Set  `tfsdk:"users_group" json:"users_group,omitempty"`                 // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceUsersGroup
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`                             // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceGroup
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`               // []*PolicyPolicyWanFirewallPolicyRulesRuleSourceSystemGroup
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceHost struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceSite struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceIPRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceGlobalIPRange struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceNetworkInterface struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceSiteNetworkSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceUser struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceUsersGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceSystemGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSourceFloatingSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestination struct {
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`                                   // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationIP
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"`                               // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationHost
	Site              types.Set  `tfsdk:"site" json:"site,omitempty"`                               // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationSite
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`                           // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationSubnet
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`                       // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationIPRange
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`         // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationGlobalIPRange
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`     // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationNetworkInterface
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"` // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationSiteNetworkSubnet
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`         // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationFloatingSubnet
	User              types.Set  `tfsdk:"user" json:"user,omitempty"`                               // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationUser
	UsersGroup        types.Set  `tfsdk:"users_group" json:"users_group,omitempty"`                 // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationUsersGroup
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`                             // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationGroup
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`               // []*PolicyPolicyWanFirewallPolicyRulesRuleDestinationSystemGroup
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationHost struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationSite struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationIPRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationGlobalIPRange struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationNetworkInterface struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationSiteNetworkSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationUser struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationUsersGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationSystemGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleDestinationFloatingSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleService struct {
	Standard types.Set  `tfsdk:"standard" json:"standard,omitempty"` // []*PolicyPolicyWanFirewallPolicyRulesRuleServiceStandard
	Custom   types.List `tfsdk:"custom" json:"custom,omitempty"`     // []*PolicyPolicyWanFirewallPolicyRulesRuleServiceCustom
}

type PolicyPolicyWanFirewallPolicyRulesRuleServiceStandard struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleServiceCustom struct {
	Port      types.List   `tfsdk:"port" json:"port,omitempty"`
	PortRange types.Object `tfsdk:"port_range" json:"port_range,omitempty"` // *PolicyPolicyWanFirewallPolicyRulesRuleServiceCustomPortRange
	Protocol  types.String `tfsdk:"protocol" json:"protocol,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleServiceCustomPortRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleTracking struct {
	Event types.Object `tfsdk:"event" json:"event,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRuleTrackingEvent
	Alert types.Object `tfsdk:"alert" json:"alert,omitempty"` // PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlert
}

type PolicyPolicyWanFirewallPolicyRulesRuleTrackingEvent struct {
	Enabled types.Bool `tfsdk:"enabled" json:"enabled,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlert struct {
	Enabled           types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	Frequency         types.String `tfsdk:"frequency" json:"frequency,omitempty"`
	SubscriptionGroup types.Set    `tfsdk:"subscription_group" json:"subscription_group,omitempty"` // []*PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlertSubscriptionGroup
	Webhook           types.Set    `tfsdk:"webhook" json:"webhook,omitempty"`                       // []*PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlertWebhook
	MailingList       types.Set    `tfsdk:"mailing_list" json:"mailing_list,omitempty"`             // []*PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlertMailingList
}

type PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlertSubscriptionGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlertWebhook struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleTrackingAlertMailingList struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleSchedule struct {
	ActiveOn        types.String `tfsdk:"active_on" json:"active_on,omitempty"`
	CustomTimeframe types.Object `tfsdk:"custom_timeframe" json:"custom_timeframe,omitempty"` // *PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe
	CustomRecurring types.Object `tfsdk:"custom_recurring" json:"custom_recurring,omitempty"` // *PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring
}

type PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomTimeframe struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyWanFirewallPolicyRulesRuleScheduleCustomRecurring struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
	Days types.List   `tfsdk:"days" json:"days,omitempty"` // []DayOfWeek
}

type PolicyPolicyWanFirewallPolicyRulesRuleExceptions struct {
	Name             types.String `tfsdk:"name" json:"name,omitempty"`
	Source           types.Object `tfsdk:"source" json:"source,omitempty"`
	ConnectionOrigin types.String `tfsdk:"connection_origin" json:"connectionOrigin,omitempty"`
	Country          types.Set    `tfsdk:"country" json:"country,omitempty"`
	Device           types.Set    `tfsdk:"device" json:"device,omitempty"`
	DeviceAttributes types.Object `tfsdk:"device_attributes" json:"deviceAttributes,omitempty"`
	DeviceOs         types.List   `tfsdk:"device_os" json:"deviceOS,omitempty"`
	Destination      types.Object `tfsdk:"destination" json:"destination,omitempty"`
	Application      types.Object `tfsdk:"application" json:"application,omitempty"`
	Service          types.Object `tfsdk:"service" json:"service,omitempty"`
	Direction        types.String `tfsdk:"direction" json:"direction,omitempty"`
}

// WanFirewallRuleObjectType defines the top-level WAN Firewall rule object type.
var WanFirewallRuleObjectType = types.ObjectType{AttrTypes: WanFirewallRuleAttrTypes}
var WanFirewallRuleAttrTypes = map[string]attr.Type{
	"rule": WanFirewallRuleRuleObjectType,
	"at":   PositionObjectType,
}

var WanFirewallRuleRuleObjectType = types.ObjectType{AttrTypes: WanFirewallRuleRuleAttrTypes}
var WanFirewallRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"enabled":     types.BoolType,
	// "section":           NameIDObjectType,
	"source":            WanSourceObjectType,
	"connection_origin": types.StringType,
	"active_period":     types.ObjectType{AttrTypes: ActivePeriodAttrTypes},
	"country":           types.SetType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device":            types.SetType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device_attributes": types.ObjectType{AttrTypes: WanDeviceAttrAttrTypes},
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

var WanApplicationObjectType = types.ObjectType{AttrTypes: WanApplicationAttrTypes}
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
	"name":              types.StringType,
	"source":            types.ObjectType{AttrTypes: WanSourceAttrTypes},
	"country":           types.SetType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device":            types.SetType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"device_attributes": types.ObjectType{AttrTypes: WanDeviceAttrAttrTypes},
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
	"os_version":   types.ListType{ElemType: types.StringType},
}
