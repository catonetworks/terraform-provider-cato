package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

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
