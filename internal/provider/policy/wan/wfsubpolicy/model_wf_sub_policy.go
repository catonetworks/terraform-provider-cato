package wfsubpolicy

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WanFirewallSubPolicy is the Terraform model for the cato_wf_sub_policy
// resource. A sub-policy owns a SUB_POLICY_SCOPE rule (exposed here as the
// embedded scope object, reusing the cato_wf_rule schema and hydrators) plus an
// API-managed cleanup rule. The Cato API has no updateSubPolicy mutation, so
// name and description are RequiresReplace.
type WanFirewallSubPolicy struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	At            types.Object `tfsdk:"at"`              // *policyrule.PolicyRulePositionInput
	ScopeRuleID   types.String `tfsdk:"scope_rule_id"`   // computed SUB_POLICY_SCOPE rule id
	Scope         types.Object `tfsdk:"scope"`           // PolicyPolicyWanFirewallPolicyRulesRule
	CleanupRuleID types.String `tfsdk:"cleanup_rule_id"` // computed id of the API-created cleanup rule
	CleanupRule   types.Object `tfsdk:"cleanup_rule"`    // PolicyPolicyWanFirewallPolicyRulesRule
}
