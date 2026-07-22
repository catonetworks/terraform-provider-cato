package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// InternetFirewallSubPolicy is the Terraform model for the cato_if_sub_policy
// resource. A sub-policy owns a SUB_POLICY_SCOPE rule (exposed here as the
// embedded scope object, reusing the cato_if_rule schema and hydrators) plus an
// API-managed cleanup rule. The Cato API has no updateSubPolicy mutation, so
// name and description are RequiresReplace.
type InternetFirewallSubPolicy struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	At          types.Object `tfsdk:"at"`            // *PolicyRulePositionInput
	ScopeRuleID types.String `tfsdk:"scope_rule_id"` // computed SUB_POLICY_SCOPE rule id
	Scope       types.Object `tfsdk:"scope"`         // PolicyPolicyInternetFirewallPolicyRulesRule
}
