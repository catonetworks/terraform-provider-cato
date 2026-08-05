package provider

import (
	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Name        types.String `tfsdk:"name" json:"name,omitempty"`
	NAT         types.Object `tfsdk:"nat" json:"nat,omitempty"`         // PolicyNatSettings
	Service     types.Object `tfsdk:"service" json:"service,omitempty"` // PolicyService (service name or port/range/protocol)
	Site        types.Object `tfsdk:"site" json:"site,omitempty"`       // PolicySite (sites or groups)
	Source      types.Object `tfsdk:"source" json:"source,omitempty"`   // SocketLanSource
	// Transport = LAN
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
	Custom types.Set `tfsdk:"custom" json:"custom,omitempty"` // []ProtocolPort
	Simple types.Set `tfsdk:"simple" json:"simple,omitempty"`
}

var PolicyServiceTypes = map[string]attr.Type{
	"custom": types.SetType{ElemType: types.ObjectType{AttrTypes: ProtocolPortTypes}},
	"simple": types.SetType{ElemType: types.StringType},
}

type PolicySite struct {
	Group types.Set `tfsdk:"group" json:"group,omitempty"` // []IDNameRefModel
	Site  types.Set `tfsdk:"site" json:"site,omitempty"`   // []IDNameRefModel
}

var PolicySiteTypes = map[string]attr.Type{
	"group": types.SetType{ElemType: types.ObjectType{AttrTypes: parse.IDNameRefModelTypes}},
	"site":  types.SetType{ElemType: types.ObjectType{AttrTypes: parse.IDNameRefModelTypes}},
}
