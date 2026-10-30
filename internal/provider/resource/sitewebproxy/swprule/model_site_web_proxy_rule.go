package swprule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// resourceModel contains Terraform values for site_web_proxy_rule.
type resourceModel struct {
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// AtModel contains Terraform values for at.
type AtModel struct {
	Position types.String `tfsdk:"position"`
	Ref      types.String `tfsdk:"ref"`
}

var atAttrTypes = map[string]attr.Type{
	"position": types.StringType,
	"ref":      types.StringType,
}

// RuleModel contains Terraform values for rule.
type RuleModel struct {
	AllowedDestinationProtocol  types.String `tfsdk:"allowed_destination_protocol"`
	AssociatedSite              types.Set    `tfsdk:"associated_site"`
	AuthenticationConfig        types.Object `tfsdk:"authentication_config"`
	Description                 types.String `tfsdk:"description"`
	Enabled                     types.Bool   `tfsdk:"enabled"`
	Fqdn                        types.String `tfsdk:"fqdn"`
	ID                          types.String `tfsdk:"id"`
	Index                       types.Int64  `tfsdk:"index"`
	Name                        types.String `tfsdk:"name"`
	Port                        types.String `tfsdk:"port"`
	Section                     types.Object `tfsdk:"section"`
	ShouldAssociateWithAllSites types.Bool   `tfsdk:"should_associate_with_all_sites"`
}

var ruleAttrTypes = map[string]attr.Type{
	"allowed_destination_protocol":    types.StringType,
	"associated_site":                 types.SetType{ElemType: idname.RefObjectType},
	"authentication_config":           types.ObjectType{AttrTypes: ruleAuthenticationConfigAttrTypes},
	"description":                     types.StringType,
	"enabled":                         types.BoolType,
	"fqdn":                            types.StringType,
	"id":                              types.StringType,
	"index":                           types.Int64Type,
	"name":                            types.StringType,
	"port":                            types.StringType,
	"section":                         types.ObjectType{AttrTypes: ruleSectionAttrTypes},
	"should_associate_with_all_sites": types.BoolType,
}

// RuleAuthenticationConfigModel contains Terraform values for rule.authentication_config.
type RuleAuthenticationConfigModel struct {
	KerberosConfig types.Object `tfsdk:"kerberos_config"`
	Method         types.String `tfsdk:"method"`
}

var ruleAuthenticationConfigAttrTypes = map[string]attr.Type{
	"kerberos_config": types.ObjectType{AttrTypes: ruleAuthenticationConfigKerberosConfigAttrTypes},
	"method":          types.StringType,
}

// RuleAuthenticationConfigKerberosConfigModel contains Terraform values for rule.authentication_config.kerberos_config.
type RuleAuthenticationConfigKerberosConfigModel struct {
	EncryptedKeytab types.String `tfsdk:"encrypted_keytab"`
	IsEnabled       types.Bool   `tfsdk:"is_enabled"`
}

var ruleAuthenticationConfigKerberosConfigAttrTypes = map[string]attr.Type{
	"encrypted_keytab": types.StringType,
	"is_enabled":       types.BoolType,
}

// RuleSectionModel contains Terraform values for rule.section.
type RuleSectionModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var ruleSectionAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}
