package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WanFirewallSubPolicy is the Terraform model for the cato_wf_sub_policy
// resource. It mirrors InternetFirewallSubPolicy but reuses the cato_wf_rule
// schema and hydrators for the embedded scope object. The Cato API has no
// updateSubPolicy mutation, so name and description are RequiresReplace.
type WanFirewallSubPolicy struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	At          types.Object `tfsdk:"at"`            // *PolicyRulePositionInput
	ScopeRuleID types.String `tfsdk:"scope_rule_id"` // computed SUB_POLICY_SCOPE rule id
	Scope       types.Object `tfsdk:"scope"`         // PolicyPolicyWanFirewallPolicyRulesRule
}
