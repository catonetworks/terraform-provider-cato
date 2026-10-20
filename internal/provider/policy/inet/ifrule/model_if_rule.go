//nolint:lll
package ifrule

import (
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/activeperiod"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyschedule"
)

type InternetFirewallRule struct {
	Rule        types.Object `tfsdk:"rule" json:"rule,omitempty"` // PolicyPolicyInternetFirewallPolicyRulesRule
	At          types.Object `tfsdk:"at" json:"at,omitempty"`     // *policyrule.PolicyRulePositionInput
	SubPolicyID types.String `tfsdk:"sub_policy_id" json:"sub_policy_id,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRule struct {
	ID          types.String `tfsdk:"id" json:"id,omitempty"`
	Name        types.String `tfsdk:"name" json:"name,omitempty"`
	Description types.String `tfsdk:"description" json:"description,omitempty"`
	Enabled     types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	// Section          types.Object `tfsdk:"section" json:"section,omitempty"` // PolicyPolicyInternetFirewallPolicyRulesRuleSection
	Source           types.Object `tfsdk:"source" json:"source,omitempty"` // PolicyPolicyInternetFirewallPolicyRulesRuleSource
	ConnectionOrigin types.String `tfsdk:"connection_origin" json:"connection_origin,omitempty"`
	Country          types.Set    `tfsdk:"country" json:"country,omitempty"`                    // []PolicyPolicyInternetFirewallPolicyRulesRuleCountry
	Device           types.Set    `tfsdk:"device" json:"device,omitempty"`                      // []PolicyPolicyInternetFirewallPolicyRulesRuleDevice
	DeviceAttributes types.Object `tfsdk:"device_attributes" json:"deviceAttributes,omitempty"` // PolicyPolicyInternetFirewallPolicyRulesRuleDeviceAttributes
	DeviceOs         types.List   `tfsdk:"device_os" json:"device_os,omitempty"`
	Destination      types.Object `tfsdk:"destination" json:"destination,omitempty"` // PolicyPolicyInternetFirewallPolicyRulesRuleDestination
	Service          types.Object `tfsdk:"service" json:"service,omitempty"`         // PolicyPolicyInternetFirewallPolicyRulesRuleService
	Action           types.String `tfsdk:"action" json:"action,omitempty"`
	Tracking         types.Object `tfsdk:"tracking" json:"tracking,omitempty"`           // policyrule.PolicyTracking
	ActivePeriod     types.Object `tfsdk:"active_period" json:"active_period,omitempty"` // PolicyPolicyInternetFirewallPolicyRulesRuleActivePeriod
	Schedule         types.Object `tfsdk:"schedule" json:"schedule,omitempty"`           // PolicyPolicyInternetFirewallPolicyRulesRuleSchedule
	Exceptions       types.Set    `tfsdk:"exceptions" json:"exceptions,omitempty"`       // []*PolicyPolicyInternetFirewallPolicyRulesRuleExceptions
}

type PolicyPolicyInternetFirewallPolicyRulesRuleActivePeriod struct {
	EffectiveFrom    types.String `tfsdk:"effective_from" json:"effective_from,omitempty"`
	ExpiresAt        types.String `tfsdk:"expires_at" json:"expires_at,omitempty"`
	UseEffectiveFrom types.Bool   `tfsdk:"use_effective_from" json:"use_effective_from,omitempty"`
	UseExpiresAt     types.Bool   `tfsdk:"use_expires_at" json:"use_expires_at,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleCountry struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDevice struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDeviceAttributes struct {
	Category     types.List `tfsdk:"category" json:"category,omitempty"`
	Type         types.List `tfsdk:"type" json:"type,omitempty"`
	Model        types.List `tfsdk:"model" json:"model,omitempty"`
	Manufacturer types.List `tfsdk:"manufacturer" json:"manufacturer,omitempty"`
	Os           types.List `tfsdk:"os" json:"os,omitempty"`
	OsVersion    types.List `tfsdk:"os_version" json:"osVersion,omitempty"`
}

// DeviceAttributesInputIfw converts device attributes from tfsdk to cato_models.
type DeviceAttributesInputIfw struct {
	Category     []string `tfsdk:"category" json:"category"`
	Manufacturer []string `tfsdk:"manufacturer" json:"manufacturer"`
	Model        []string `tfsdk:"model" json:"model"`
	Os           []string `tfsdk:"os" json:"os"`
	OsVersion    []string `tfsdk:"os_version" json:"osVersion"`
	Type         []string `tfsdk:"type" json:"type"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSection struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDestination struct {
	Application            types.Set  `tfsdk:"application" json:"application,omitempty"`                           // []*PolicyPolicyInternetFirewallPolicyRulesRuleDestinationApplication
	CustomApp              types.Set  `tfsdk:"custom_app" json:"custom_app,omitempty"`                             // []*PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCustomApp
	AppCategory            types.Set  `tfsdk:"app_category" json:"app_category,omitempty"`                         // []*PolicyPolicyInternetFirewallPolicyRulesRuleDestinationAppCategory
	CustomCategory         types.Set  `tfsdk:"custom_category" json:"custom_category,omitempty"`                   // []*PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCustomCategory
	SanctionedAppsCategory types.Set  `tfsdk:"sanctioned_apps_category" json:"sanctioned_apps_category,omitempty"` // []*PolicyPolicyInternetFirewallPolicyRulesRuleDestinationSanctionedAppsCategory
	Country                types.Set  `tfsdk:"country" json:"country,omitempty"`                                   // []*PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCountry
	Domain                 types.List `tfsdk:"domain" json:"domain,omitempty"`
	Fqdn                   types.List `tfsdk:"fqdn" json:"fqdn,omitempty"`
	IP                     types.List `tfsdk:"ip" json:"ip,omitempty"`
	Subnet                 types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange                types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`               // []*ipaddr.Range
	GlobalIPRange          types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"` // []*PolicyPolicyInternetFirewallPolicyRulesRuleDestinationGlobalIPRange
	RemoteAsn              types.List `tfsdk:"remote_asn" json:"remote_asn,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDestinationApplication struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCustomApp struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDestinationAppCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCustomCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDestinationSanctionedAppsCategory struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCountry struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleDestinationGlobalIPRange struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSource struct {
	IP                types.List `tfsdk:"ip" json:"ip,omitempty"`
	Host              types.Set  `tfsdk:"host" json:"host,omitempty"` // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
	Site              types.Set  `tfsdk:"site" json:"site,omitempty"` // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceSite
	Subnet            types.List `tfsdk:"subnet" json:"subnet,omitempty"`
	IPRange           types.List `tfsdk:"ip_range" json:"ip_range,omitempty"`                       // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceIPRange
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range" json:"global_ip_range,omitempty"`         // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceGlobalIPRange
	NetworkInterface  types.Set  `tfsdk:"network_interface" json:"network_interface,omitempty"`     // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceNetworkInterface
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet" json:"site_network_subnet,omitempty"` // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceSiteNetworkSubnet
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet" json:"floating_subnet,omitempty"`         // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceFloatingSubnet
	User              types.Set  `tfsdk:"user" json:"user,omitempty"`                               // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceUser
	UsersGroup        types.Set  `tfsdk:"users_group" json:"users_group,omitempty"`                 // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceUsersGroup
	Group             types.Set  `tfsdk:"group" json:"group,omitempty"`                             // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceGroup
	SystemGroup       types.Set  `tfsdk:"system_group" json:"system_group,omitempty"`               // []*PolicyPolicyInternetFirewallPolicyRulesRuleSourceSystemGroup
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceSite struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceIPRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceGlobalIPRange struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceNetworkInterface struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceSiteNetworkSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceUser struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceUsersGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceSystemGroup struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSourceFloatingSubnet struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleService struct {
	Standard types.Set  `tfsdk:"standard" json:"standard,omitempty"` // []*PolicyPolicyInternetFirewallPolicyRulesRuleServiceStandard
	Custom   types.List `tfsdk:"custom" json:"custom,omitempty"`     // []*customservice.CustomService
}

type PolicyPolicyInternetFirewallPolicyRulesRuleServiceStandard struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleSchedule struct {
	ActiveOn        types.String `tfsdk:"active_on" json:"active_on,omitempty"`
	CustomTimeframe types.Object `tfsdk:"custom_timeframe" json:"custom_timeframe,omitempty"` // *PolicyPolicyInternetFirewallPolicyRulesRuleScheduleCustomTimeframe
	CustomRecurring types.Object `tfsdk:"custom_recurring" json:"custom_recurring,omitempty"` // *PolicyPolicyInternetFirewallPolicyRulesRuleScheduleCustomRecurring
}

type PolicyPolicyInternetFirewallPolicyRulesRuleScheduleCustomTimeframe struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

type PolicyPolicyInternetFirewallPolicyRulesRuleScheduleCustomRecurring struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
	Days types.List   `tfsdk:"days" json:"days,omitempty"` // []DayOfWeek
}

type DayOfWeek types.String

type PolicyPolicyInternetFirewallPolicyRulesRuleExceptions struct {
	Name             types.String `tfsdk:"name" json:"name,omitempty"` // // / //
	Source           types.Object `tfsdk:"source" json:"source,omitempty"`
	ConnectionOrigin types.String `tfsdk:"connection_origin" json:"connection_origin,omitempty"` // // / //
	Country          types.Set    `tfsdk:"country" json:"country,omitempty"`
	Device           types.Set    `tfsdk:"device" json:"device,omitempty"`
	DeviceAttributes types.Object `tfsdk:"device_attributes" json:"deviceAttributes,omitempty"`
	DeviceOs         types.List   `tfsdk:"device_os" json:"device_os,omitempty"`
	Destination      types.Object `tfsdk:"destination" json:"destination,omitempty"`
	Service          types.Object `tfsdk:"service" json:"service,omitempty"`
}

type OperatingSystem types.String

// InternetFirewallRuleObjectType defines the top-level Internet Firewall rule object type.
var InternetFirewallRuleObjectType = types.ObjectType{AttrTypes: InternetFirewallRuleAttrTypes}
var InternetFirewallRuleAttrTypes = map[string]attr.Type{
	"rule": InternetFirewallRuleRuleObjectType,
	"at":   policyrule.PositionObjectType,
}

var InternetFirewallRuleRuleObjectType = types.ObjectType{AttrTypes: InternetFirewallRuleRuleAttrTypes}
var InternetFirewallRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"enabled":     types.BoolType,
	// "section":           idname.RefObjectType,
	"source":            IfwSourceObjectType,
	"connection_origin": types.StringType,
	"active_period":     types.ObjectType{AttrTypes: activeperiod.ActivePeriodAttrTypes},
	"country":           types.SetType{ElemType: types.ObjectType{AttrTypes: idname.RefAttrTypes}},
	"device":            types.SetType{ElemType: types.ObjectType{AttrTypes: idname.RefAttrTypes}},
	"device_attributes": types.ObjectType{AttrTypes: IfwDeviceAttrAttrTypes},
	"device_os":         types.ListType{ElemType: types.StringType},
	"destination":       types.ObjectType{AttrTypes: IfwDestAttrTypes},
	"service":           types.ObjectType{AttrTypes: IfwServiceAttrTypes},
	"action":            types.StringType,
	"tracking":          policyrule.TrackingObjectType,
	"schedule":          policyschedule.ScheduleObjectType,
	"exceptions":        types.SetType{ElemType: types.ObjectType{AttrTypes: IfwExceptionAttrTypes}},
}

var IfwServiceObjectType = types.ObjectType{AttrTypes: IfwServiceAttrTypes}
var IfwServiceAttrTypes = map[string]attr.Type{
	"standard": types.SetType{ElemType: types.ObjectType{AttrTypes: idname.RefAttrTypes}},
	"custom":   types.ListType{ElemType: types.ObjectType{AttrTypes: customservice.CustomServiceAttrTypes}},
}

var IfwSourceObjectType = types.ObjectType{AttrTypes: IfwSourceAttrTypes}
var IfwSourceAttrTypes = map[string]attr.Type{
	"ip":                  types.ListType{ElemType: types.StringType},
	"host":                types.SetType{ElemType: idname.RefObjectType},
	"site":                types.SetType{ElemType: idname.RefObjectType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"ip_range":            types.ListType{ElemType: iprange.RangeObjectType},
	"global_ip_range":     types.SetType{ElemType: idname.RefObjectType},
	"network_interface":   types.SetType{ElemType: idname.RefObjectType},
	"site_network_subnet": types.SetType{ElemType: idname.RefObjectType},
	"floating_subnet":     types.SetType{ElemType: idname.RefObjectType},
	"user":                types.SetType{ElemType: idname.RefObjectType},
	"users_group":         types.SetType{ElemType: idname.RefObjectType},
	"group":               types.SetType{ElemType: idname.RefObjectType},
	"system_group":        types.SetType{ElemType: idname.RefObjectType},
}

var IfwDestObjectType = types.ObjectType{AttrTypes: IfwDestAttrTypes}
var IfwDestAttrTypes = map[string]attr.Type{
	"application":              types.SetType{ElemType: idname.RefObjectType},
	"custom_app":               types.SetType{ElemType: idname.RefObjectType},
	"app_category":             types.SetType{ElemType: idname.RefObjectType},
	"custom_category":          types.SetType{ElemType: idname.RefObjectType},
	"sanctioned_apps_category": types.SetType{ElemType: idname.RefObjectType},
	"country":                  types.SetType{ElemType: idname.RefObjectType},
	"domain":                   types.ListType{ElemType: types.StringType},
	"fqdn":                     types.ListType{ElemType: types.StringType},
	"ip":                       types.ListType{ElemType: types.StringType},
	"subnet":                   types.ListType{ElemType: types.StringType},
	"ip_range":                 types.ListType{ElemType: iprange.RangeObjectType},
	"global_ip_range":          types.SetType{ElemType: idname.RefObjectType},
	"remote_asn":               types.ListType{ElemType: types.StringType},
}

var IfwExceptionObjectType = types.ObjectType{AttrTypes: IfwExceptionAttrTypes}
var IfwExceptionAttrTypes = map[string]attr.Type{
	"name":              types.StringType,
	"source":            types.ObjectType{AttrTypes: IfwSourceAttrTypes},
	"country":           types.SetType{ElemType: types.ObjectType{AttrTypes: idname.RefAttrTypes}},
	"device":            types.SetType{ElemType: types.ObjectType{AttrTypes: idname.RefAttrTypes}},
	"device_attributes": types.ObjectType{AttrTypes: IfwDeviceAttrAttrTypes},
	"device_os":         types.ListType{ElemType: types.StringType},
	"destination":       types.ObjectType{AttrTypes: IfwDestAttrTypes},
	"service":           types.ObjectType{AttrTypes: IfwServiceAttrTypes},
	"connection_origin": types.StringType,
}

var IfwDeviceAttrObjectType = types.ObjectType{AttrTypes: IfwDeviceAttrAttrTypes}
var IfwDeviceAttrAttrTypes = map[string]attr.Type{
	"category":     types.ListType{ElemType: types.StringType},
	"type":         types.ListType{ElemType: types.StringType},
	"model":        types.ListType{ElemType: types.StringType},
	"manufacturer": types.ListType{ElemType: types.StringType},
	"os":           types.ListType{ElemType: types.StringType},
	"os_version":   types.ListType{ElemType: types.StringType},
}

// HydrateIfwAPITypes create sub-types for both create and update calls to populate both entries
type HydrateIfwAPITypes struct {
	Create cato_models.InternetFirewallAddRuleInput
	Update cato_models.InternetFirewallUpdateRuleInput
}
