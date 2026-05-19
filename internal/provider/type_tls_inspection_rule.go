package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TLSInspectionRule represents the top-level resource structure
type TLSInspectionRule struct {
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
	ID   types.String `tfsdk:"id"`
}

// PolicyPolicyTLSInspectPolicyRulesRule represents the rule structure
type PolicyPolicyTLSInspectPolicyRulesRule struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
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

// PolicyPolicyTLSInspectPolicyRulesRuleSource represents the source criteria
type PolicyPolicyTLSInspectPolicyRulesRuleSource struct {
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

// PolicyPolicyTLSInspectPolicyRulesRuleApplication represents the application criteria
type PolicyPolicyTLSInspectPolicyRulesRuleApplication struct {
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
	CustomServiceIP    types.Object `tfsdk:"custom_service_ip"`
	TLSInspectCategory types.String `tfsdk:"tls_inspect_category"`
	Country            types.Set    `tfsdk:"country"`
}

// PolicyPolicyTLSInspectPolicyRulesRuleApplicationCustomService represents custom service definition
type PolicyPolicyTLSInspectPolicyRulesRuleApplicationCustomService struct {
	Port      types.List   `tfsdk:"port"`
	PortRange types.Object `tfsdk:"port_range"`
	Protocol  types.String `tfsdk:"protocol"`
}

// PolicyPolicyTLSInspectPolicyRulesRuleApplicationCustomServiceIP represents custom service IP definition
type PolicyPolicyTLSInspectPolicyRulesRuleApplicationCustomServiceIP struct {
	Name    types.String `tfsdk:"name"`
	IP      types.String `tfsdk:"ip"`
	IPRange types.Object `tfsdk:"ip_range"`
}

// TLSSourceAttrTypes defines the attribute types for TLS source criteria
var TLSSourceAttrTypes = map[string]attr.Type{
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

// TLSApplicationAttrTypes defines the attribute types for TLS application criteria
var TLSApplicationAttrTypes = map[string]attr.Type{
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
	"custom_service_ip":    types.ObjectType{AttrTypes: CustomServiceIPAttrTypes},
	"tls_inspect_category": types.StringType,
	"country":              types.SetType{ElemType: NameIDObjectType},
}

// NameIDRef defines a nested object with ID and name fields.
type NameIDRef struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type FromTo struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

// TLSInspectionRuleRuleAttrTypes defines the attribute types for the TLS rule object
var TLSInspectionRuleRuleAttrTypes = map[string]attr.Type{
	"id":                           types.StringType,
	"name":                         types.StringType,
	"description":                  types.StringType,
	"enabled":                      types.BoolType,
	"action":                       types.StringType,
	"untrusted_certificate_action": types.StringType,
	"connection_origin":            types.StringType,
	"source":                       types.ObjectType{AttrTypes: TLSSourceAttrTypes},
	"country":                      types.SetType{ElemType: NameIDObjectType},
	"device_posture_profile":       types.SetType{ElemType: NameIDObjectType},
	"platform":                     types.StringType,
	"application":                  types.ObjectType{AttrTypes: TLSApplicationAttrTypes},
}
