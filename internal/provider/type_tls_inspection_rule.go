package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TlsInspectionRule represents the top-level resource structure
type TlsInspectionRule struct {
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
	ID   types.String `tfsdk:"id"`
}

// Policy_Policy_TlsInspect_Policy_Rules_Rule represents the rule structure
type Policy_Policy_TlsInspect_Policy_Rules_Rule struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	Index                      types.Int64  `tfsdk:"index"`
	Enabled                    types.Bool   `tfsdk:"enabled"`
	Action                     types.String `tfsdk:"action"`
	UntrustedCertificateAction types.String `tfsdk:"untrusted_certificate_action"`
	ConnectionOrigin           types.String `tfsdk:"connection_origin"`
	Source                     types.Object `tfsdk:"source"`
	Country                    types.Set    `tfsdk:"country"`
	DevicePostureProfile       types.Set    `tfsdk:"device_posture_profile"`
	Platform                   types.String `tfsdk:"platform"`
	Application                types.Object `tfsdk:"application"`
}

// Policy_Policy_TlsInspect_Policy_Rules_Rule_Source represents the source criteria
type Policy_Policy_TlsInspect_Policy_Rules_Rule_Source struct {
	IP                types.List `tfsdk:"ip"`
	Subnet            types.List `tfsdk:"subnet"`
	Host              types.Set  `tfsdk:"host"`
	Site              types.Set  `tfsdk:"site"`
	IPRange           types.List `tfsdk:"ip_range"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range"`
	NetworkInterface  types.Set  `tfsdk:"network_interface"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet"`
	User              types.Set  `tfsdk:"user"`
	UsersGroup        types.Set  `tfsdk:"users_group"`
	Group             types.Set  `tfsdk:"group"`
	SystemGroup       types.Set  `tfsdk:"system_group"`
}

// Policy_Policy_TlsInspect_Policy_Rules_Rule_Application represents the application criteria
type Policy_Policy_TlsInspect_Policy_Rules_Rule_Application struct {
	Application        types.Set    `tfsdk:"application"`
	CustomApp          types.Set    `tfsdk:"custom_app"`
	AppCategory        types.Set    `tfsdk:"app_category"`
	CustomCategory     types.Set    `tfsdk:"custom_category"`
	Domain             types.List   `tfsdk:"domain"`
	Fqdn               types.List   `tfsdk:"fqdn"`
	IP                 types.List   `tfsdk:"ip"`
	Subnet             types.List   `tfsdk:"subnet"`
	IPRange            types.List   `tfsdk:"ip_range"`
	GlobalIPRange      types.Set    `tfsdk:"global_ip_range"`
	RemoteAsn          types.List   `tfsdk:"remote_asn"`
	Service            types.Set    `tfsdk:"service"`
	CustomService      types.Object `tfsdk:"custom_service"`
	CustomServiceIp    types.Object `tfsdk:"custom_service_ip"`
	TlsInspectCategory types.String `tfsdk:"tls_inspect_category"`
	Country            types.Set    `tfsdk:"country"`
}

// Policy_Policy_TlsInspect_Policy_Rules_Rule_Application_CustomService represents custom service definition
type Policy_Policy_TlsInspect_Policy_Rules_Rule_Application_CustomService struct {
	Port      types.List   `tfsdk:"port"`
	PortRange types.Object `tfsdk:"port_range"`
	Protocol  types.String `tfsdk:"protocol"`
}

// Policy_Policy_TlsInspect_Policy_Rules_Rule_Application_CustomServiceIp represents custom service IP definition
type Policy_Policy_TlsInspect_Policy_Rules_Rule_Application_CustomServiceIp struct {
	Name    types.String `tfsdk:"name"`
	IP      types.String `tfsdk:"ip"`
	IPRange types.Object `tfsdk:"ip_range"`
}

// TlsSourceAttrTypes defines the attribute types for TLS source criteria
var TlsSourceAttrTypes = map[string]attr.Type{
	"ip":                  types.ListType{ElemType: types.StringType},
	"subnet":              types.ListType{ElemType: types.StringType},
	"host":                types.SetType{ElemType: NameIDObjectType},
	"site":                types.SetType{ElemType: NameIDObjectType},
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

// TlsApplicationAttrTypes defines the attribute types for TLS application criteria
var TlsApplicationAttrTypes = map[string]attr.Type{
	"application":          types.SetType{ElemType: NameIDObjectType},
	"custom_app":           types.SetType{ElemType: NameIDObjectType},
	"app_category":         types.SetType{ElemType: NameIDObjectType},
	"custom_category":      types.SetType{ElemType: NameIDObjectType},
	"domain":               types.ListType{ElemType: types.StringType},
	"fqdn":                 types.ListType{ElemType: types.StringType},
	"ip":                   types.ListType{ElemType: types.StringType},
	"subnet":               types.ListType{ElemType: types.StringType},
	"ip_range":             types.ListType{ElemType: FromToObjectType},
	"global_ip_range":      types.SetType{ElemType: NameIDObjectType},
	"remote_asn":           types.ListType{ElemType: types.StringType},
	"service":              types.SetType{ElemType: NameIDObjectType},
	"custom_service":       types.ObjectType{AttrTypes: CustomServiceAttrTypes},
	"custom_service_ip":    types.ObjectType{AttrTypes: CustomServiceIpAttrTypes},
	"tls_inspect_category": types.StringType,
	"country":              types.SetType{ElemType: NameIDObjectType},
}

// Shared type definitions for nested objects with ID and Name
type NameIDRef struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type FromTo struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

// TlsInspectionRuleRuleAttrTypes defines the attribute types for the TLS rule object
var TlsInspectionRuleRuleAttrTypes = map[string]attr.Type{
	"id":                           types.StringType,
	"name":                         types.StringType,
	"description":                  types.StringType,
	"index":                        types.Int64Type,
	"enabled":                      types.BoolType,
	"action":                       types.StringType,
	"untrusted_certificate_action": types.StringType,
	"connection_origin":            types.StringType,
	"source":                       types.ObjectType{AttrTypes: TlsSourceAttrTypes},
	"country":                      types.SetType{ElemType: NameIDObjectType},
	"device_posture_profile":       types.SetType{ElemType: NameIDObjectType},
	"platform":                     types.StringType,
	"application":                  types.ObjectType{AttrTypes: TlsApplicationAttrTypes},
}
