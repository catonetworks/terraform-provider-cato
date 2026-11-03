package swprule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// SiteWebProxyKerberosConfigAttrTypes is the Kerberos auth selector.
var SiteWebProxyKerberosConfigAttrTypes = map[string]attr.Type{
	"is_enabled":       types.BoolType,
	"encrypted_keytab": types.StringType,
}

// SiteWebProxyAuthenticationConfigAttrTypes is the authentication selector.
var SiteWebProxyAuthenticationConfigAttrTypes = map[string]attr.Type{
	"method":          types.StringType,
	"kerberos_config": types.ObjectType{AttrTypes: SiteWebProxyKerberosConfigAttrTypes},
}

// SiteWebProxyRuleRuleAttrTypes defines the nested `rule` object for site web proxy rules.
var SiteWebProxyRuleRuleAttrTypes = map[string]attr.Type{
	"id":                              types.StringType,
	"name":                            types.StringType,
	"description":                     types.StringType,
	"enabled":                         types.BoolType,
	"fqdn":                            types.StringType,
	"port":                            types.StringType,
	"allowed_destination_protocol":    types.StringType,
	"should_associate_with_all_sites": types.BoolType,
	"associated_site":                 types.SetType{ElemType: idname.RefObjectType},
	"authentication_config":           types.ObjectType{AttrTypes: SiteWebProxyAuthenticationConfigAttrTypes},
}

// SiteWebProxyRuleAttrTypes is the top-level resource attribute map.
var SiteWebProxyRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: SiteWebProxyRuleRuleAttrTypes},
}

// SiteWebProxyRule is the Terraform model for cato_site_web_proxy_rule.
type SiteWebProxyRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// SiteWebProxyRuleRulePlan maps the nested rule block (ObjectAs).
type SiteWebProxyRuleRulePlan struct {
	ID                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	Enabled                     types.Bool   `tfsdk:"enabled"`
	Fqdn                        types.String `tfsdk:"fqdn"`
	Port                        types.String `tfsdk:"port"`
	AllowedDestinationProtocol  types.String `tfsdk:"allowed_destination_protocol"`
	ShouldAssociateWithAllSites types.Bool   `tfsdk:"should_associate_with_all_sites"`
	AssociatedSite              types.Set    `tfsdk:"associated_site"`
	AuthenticationConfig        types.Object `tfsdk:"authentication_config"`
}

// SiteWebProxyAuthenticationConfigPlan maps the authentication selector.
type SiteWebProxyAuthenticationConfigPlan struct {
	Method         types.String `tfsdk:"method"`
	KerberosConfig types.Object `tfsdk:"kerberos_config"`
}

// SiteWebProxyKerberosConfigPlan maps the Kerberos selector.
type SiteWebProxyKerberosConfigPlan struct {
	IsEnabled       types.Bool   `tfsdk:"is_enabled"`
	EncryptedKeytab types.String `tfsdk:"encrypted_keytab"`
}
