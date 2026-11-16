package lfsubpolicy

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// lanCleanupRule preserves a previously discovered cleanup rule by ID. Only
// initial discovery/import falls back to the last rule inside the sub-policy.
func lanCleanupRule(
	policy *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
	subID, trackedID string,
) *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule {
	if policy == nil || subID == "" {
		return nil
	}
	var last *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule
	for _, rp := range policy.GetRules() {
		sp := rp.GetSubPolicy()
		if sp == nil || sp.GetID() != subID {
			continue
		}
		if rt := rp.GetRuleType(); rt != nil && *rt == cato_models.PolicyRuleTypeEnumSubPolicyScope {
			continue
		}
		rule := rp.GetRule()
		if trackedID != "" {
			if rule.GetID() == trackedID {
				return rule
			}
			continue
		}
		if last == nil || rule.GetIndex() > last.GetIndex() {
			last = rule
		}
	}
	return last
}

// lanCleanupObject encodes the cleanup rule into the model shape.
func lanCleanupObject(
	cleanup *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule,
	diags *diag.Diagnostics,
) basetypes.ObjectValue {
	obj, d := types.ObjectValue(LanFirewallSubPolicyCleanupRuleTypes, map[string]attr.Value{
		"id":          types.StringValue(cleanup.GetID()),
		"name":        types.StringValue(cleanup.GetName()),
		"description": types.StringValue(cleanup.GetDescription()),
		"enabled":     types.BoolValue(cleanup.GetEnabled()),
	})
	diags.Append(d...)
	return obj
}

// lanCleanupUpdateInput builds a partial update for the cleanup rule from the
// configured cleanup_rule object. Only configured (non-null) fields are sent.
// Returns nil when nothing is configured.
func lanCleanupUpdateInput(
	ctx context.Context,
	desired types.Object,
	cleanupID string,
	diags *diag.Diagnostics,
) *cato_models.SocketLanUpdateRuleInput {
	if desired.IsNull() || desired.IsUnknown() {
		return nil
	}
	cfg := LanFirewallSubPolicyCleanupRule{}
	diags.Append(desired.As(ctx, &cfg, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return nil
	}
	data := &cato_models.SocketLanUpdateRuleDataInput{}
	touched := false
	if !cfg.Name.IsNull() && !cfg.Name.IsUnknown() && cfg.Name.ValueString() != "" {
		data.Name = cfg.Name.ValueStringPointer()
		touched = true
	}
	if !cfg.Description.IsNull() && !cfg.Description.IsUnknown() {
		data.Description = cfg.Description.ValueStringPointer()
		touched = true
	}
	if !cfg.Enabled.IsNull() && !cfg.Enabled.IsUnknown() {
		data.Enabled = cfg.Enabled.ValueBoolPointer()
		touched = true
	}
	if !touched {
		return nil
	}
	return &cato_models.SocketLanUpdateRuleInput{ID: cleanupID, Rule: data}
}
