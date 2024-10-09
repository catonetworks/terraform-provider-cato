package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type InternetFirewallRule struct {
	Rule types.Object `tfsdk:"rule"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule
	At   types.Object `tfsdk:"at"`   //*PolicyRulePositionInput
}

type PolicyRulePositionInput struct {
	Position types.String `tfsdk:"position"`
	Ref      types.String `tfsdk:"ref"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Index            types.Int64  `tfsdk:"index"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Section          types.Object `tfsdk:"section"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Section
	Source           types.Object `tfsdk:"source"`  //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source
	ConnectionOrigin types.String `tfsdk:"connection_origin"`
	Country          types.List   `tfsdk:"country"` //[]Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country
	Device           types.List   `tfsdk:"device"`  //[]Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device
	DeviceOs         types.List   `tfsdk:"device_os"`
	Destination      types.Object `tfsdk:"destination"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination
	Service          types.Object `tfsdk:"service"`     //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service
	Action           types.String `tfsdk:"action"`
	Tracking         types.Object `tfsdk:"tracking"`   //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking
	Schedule         types.Object `tfsdk:"schedule"`   //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule
	Exceptions       types.List   `tfsdk:"exceptions"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Section struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination struct {
	Application            types.List `tfsdk:"application"`              //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application
	CustomApp              types.List `tfsdk:"custom_app"`               //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp
	AppCategory            types.List `tfsdk:"app_category"`             //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory
	CustomCategory         types.List `tfsdk:"custom_category"`          //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory
	SanctionedAppsCategory types.List `tfsdk:"sanctioned_apps_category"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory
	Country                types.List `tfsdk:"country"`                  //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country
	Domain                 types.List `tfsdk:"domain"`
	Fqdn                   types.List `tfsdk:"fqdn"`
	IP                     types.List `tfsdk:"ip"`
	Subnet                 types.List `tfsdk:"subnet"`
	IPRange                types.List `tfsdk:"ip_range"`        //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
	GlobalIPRange          types.List `tfsdk:"global_ip_range"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
	RemoteAsn              types.List `tfsdk:"remote_asn"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source struct {
	IP                types.List `tfsdk:"ip"`
	Host              types.List `tfsdk:"host"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host
	Site              types.List `tfsdk:"site"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site
	Subnet            types.List `tfsdk:"subnet"`
	IPRange           types.List `tfsdk:"ip_range"`            //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange
	GlobalIPRange     types.List `tfsdk:"global_ip_range"`     //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange
	NetworkInterface  types.List `tfsdk:"network_interface"`   //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface
	SiteNetworkSubnet types.List `tfsdk:"site_network_subnet"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
	FloatingSubnet    types.List `tfsdk:"floating_subnet"`     //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet
	User              types.List `tfsdk:"user"`                //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User
	UsersGroup        types.List `tfsdk:"users_group"`         //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup
	Group             types.List `tfsdk:"group"`               //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group
	SystemGroup       types.List `tfsdk:"system_group"`        //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service struct {
	Standard types.List `tfsdk:"standard"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard
	Custom   types.List `tfsdk:"custom"`   //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom struct {
	Port      types.List   `tfsdk:"port"`
	PortRange types.Object `tfsdk:"port_range"` //*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange
	Protocol  types.String `tfsdk:"protocol"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking struct {
	Event types.Object `tfsdk:"event"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event
	Alert types.Object `tfsdk:"alert"` //Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	Frequency         types.String `tfsdk:"frequency"`
	SubscriptionGroup types.List   `tfsdk:"subscription_group"` //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
	Webhook           types.List   `tfsdk:"webhook"`            //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_Webhook
	MailingList       types.List   `tfsdk:"mailing_list"`       //[]*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_MailingList
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_Webhook struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_MailingList struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule struct {
	ActiveOn        types.String `tfsdk:"active_on"`
	CustomTimeframe types.Object `tfsdk:"custom_timeframe"` //*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe
	CustomRecurring types.Object `tfsdk:"custom_recurring"` //*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomRecurring
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomRecurring struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
	Days types.List   `tfsdk:"days"` //[]DayOfWeek
}

type DayOfWeek types.String

type Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions struct {
	Name             types.String `tfsdk:"name"` ///////
	Source           types.Object `tfsdk:"source"`
	ConnectionOrigin types.String `tfsdk:"connection_origin"` ///////
	Country          types.List   `tfsdk:"country"`
	Device           types.List   `tfsdk:"device"`
	DeviceOs         types.List   `tfsdk:"device_os"`
	Destination      types.Object `tfsdk:"destination"`
	Service          types.Object `tfsdk:"service"`
}

type OperatingSystem types.String
