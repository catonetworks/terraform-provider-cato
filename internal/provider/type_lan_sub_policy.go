package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
)

// LanFirewallSubPolicy is the Terraform model for the cato_lf_sub_policy
// resource.
type LanFirewallSubPolicy struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	At          types.Object `tfsdk:"at"`            // PolicyRulePositionInput
	ScopeRuleID types.String `tfsdk:"scope_rule_id"` // computed SUB_POLICY_SCOPE rule id
	Scope       types.Object `tfsdk:"scope"`         // LanFirewallSubPolicyScope
}

type LanFirewallSubPolicyScope struct {
	Description types.String `tfsdk:"description" json:"description,omitempty"`
	Destination types.Object `tfsdk:"destination" json:"destination,omitempty"` // SocketLanDestination
	Direction   types.String `tfsdk:"direction" json:"direction,omitempty"`
	Enabled     types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name" json:"name,omitempty"`
	NAT         types.Object `tfsdk:"nat" json:"nat,omitempty"`         // PolicyNatSettings
	Service     types.Object `tfsdk:"service" json:"service,omitempty"` // PolicyService (service name or port/range/protocol)
	Site        types.Object `tfsdk:"site" json:"site,omitempty"`       // PolicySite (sites or groups)
	Source      types.Object `tfsdk:"source" json:"source,omitempty"`   // SocketLanSource
	// Transport = LAN
}

var LanFirewallSubPolicyScopeTypes = map[string]attr.Type{
	"description": types.StringType,
	"destination": types.ObjectType{AttrTypes: SocketLanDestinationAttrTypes},
	"direction":   types.StringType,
	"enabled":     types.BoolType,
	"id":          types.StringType,
	"name":        types.StringType,
	"nat":         types.ObjectType{AttrTypes: PolicyNatSettingsTypes},
	"service":     types.ObjectType{AttrTypes: PolicyServiceTypes},
	"site":        types.ObjectType{AttrTypes: PolicySiteTypes},
	"source":      types.ObjectType{AttrTypes: SocketLanSourceAttrTypes},
}

type PolicyNatSettings struct {
	Enabled types.Bool   `tfsdk:"enabled" json:"enabled,omitempty"`
	NatType types.String `tfsdk:"nat_type" json:"nat_type,omitempty"`
}

var PolicyNatSettingsTypes = map[string]attr.Type{
	"enabled":  types.BoolType,
	"nat_type": types.StringType,
}

type PolicyService struct {
	Custom types.List `tfsdk:"custom" json:"custom,omitempty"`
	Simple types.Set  `tfsdk:"simple" json:"simple,omitempty"`
}

var PolicyServiceTypes = map[string]attr.Type{
	"custom": types.ListType{ElemType: types.ObjectType{AttrTypes: PolicyCustomServiceTypes}},
	"simple": types.SetType{ElemType: SimpleServiceObjectType},
}

// PolicyCustomService represents custom service definition
type PolicyCustomService struct {
	Port      types.List   `tfsdk:"port" json:"port,omitempty"`
	PortRange types.Object `tfsdk:"port_range" json:"port_range,omitempty"`
	Protocol  types.String `tfsdk:"protocol" json:"protocol,omitempty"`
}

var PolicyCustomServiceTypes = map[string]attr.Type{
	"port":       types.ListType{ElemType: types.StringType},
	"port_range": FromToObjectType,
	"protocol":   types.StringType,
}

type PolicySite struct {
	Group types.Set `tfsdk:"group" json:"group,omitempty"` // []IDNameRefModel
	Site  types.Set `tfsdk:"site" json:"site,omitempty"`   // []IDNameRefModel
}

var PolicySiteTypes = map[string]attr.Type{
	"group": types.SetType{ElemType: types.ObjectType{AttrTypes: parse.IDNameRefModelTypes}},
	"site":  types.SetType{ElemType: types.ObjectType{AttrTypes: parse.IDNameRefModelTypes}},
}
