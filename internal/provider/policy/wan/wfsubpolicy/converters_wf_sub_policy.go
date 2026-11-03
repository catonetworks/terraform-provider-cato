package wfsubpolicy

import (
	"context"
	"fmt"
	"maps"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	wfrule "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
)

// wanSubPolicyInfo returns the sub-policy info block for the given sub-policy id,
// or nil when it is not present in the policy snapshot.
func wanSubPolicyInfo(
	body *cato_go_sdk.Policy,
	subID string,
) *cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies_Policy {
	if body == nil || subID == "" {
		return nil
	}
	for _, sp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetSubPolicies() {
		if sp.GetPolicy().GetID() == subID {
			return sp.GetPolicy()
		}
	}
	return nil
}

// derefStr safely dereferences an optional string pointer.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// wanScopeRule returns the SUB_POLICY_SCOPE rule owned by the given sub-policy,
// or nil when not found.
func wanScopeRule(body *cato_go_sdk.Policy, subID string) *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule {
	if body == nil || subID == "" {
		return nil
	}
	for _, rp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetRules() {
		rt := rp.GetRuleType()
		sp := rp.GetSubPolicy()
		if rt != nil && *rt == subPolicyScopeRuleType && sp != nil && sp.GetID() == subID {
			return rp.GetRule()
		}
	}
	return nil
}

// wanCleanupRule returns the sub-policy's cleanup rule: ALWAYS the LAST rule
// (max rule.index) inside the sub-policy. The API auto-creates the cleanup rule
// at the bottom; if an admin later moves or renames rules so a different rule
// sits last, the cleanup identity follows the position — every read/update
// re-resolves the last rule and refreshes cleanup_rule_id to it. This is also
// what populates the cleanup on `terraform import`.
func wanCleanupRule(
	body *cato_go_sdk.Policy,
	subID string,
) *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule {
	if body == nil || subID == "" {
		return nil
	}
	var last *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule
	for _, rp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetRules() {
		sp := rp.GetSubPolicy()
		if sp == nil || sp.GetID() != subID {
			continue
		}
		if rt := rp.GetRuleType(); rt != nil && *rt == subPolicyScopeRuleType {
			continue
		}
		rule := rp.GetRule()
		if last == nil || rule.GetIndex() > last.GetIndex() {
			last = rule
		}
	}
	return last
}

// cleanupNameConfigured reports whether the cleanup_rule object carries a
// user-supplied (known, non-empty) name — i.e. the config renames the cleanup
// rule rather than keeping the API-generated "<sub-policy name> - Cleanup Rule".
func cleanupNameConfigured(cleanupObj types.Object) bool {
	if cleanupObj.IsNull() || cleanupObj.IsUnknown() {
		return false
	}
	nameAttr, ok := cleanupObj.Attributes()["name"]
	if !ok || nameAttr.IsNull() || nameAttr.IsUnknown() {
		return false
	}
	name, ok := nameAttr.(types.String)
	return ok && name.ValueString() != ""
}

// cleanupUpdateErrorDetail formats a cleanup-rule update failure. The cleanup
// is identified positionally (the sub-policy's LAST rule), so if an admin moves
// the cleanup rule in the CMA, the provider targets whatever rule now sits last
// and renaming it to the configured cleanup name collides with the real cleanup
// rule elsewhere in the sub-policy. Make that remediation obvious instead of
// surfacing a bare ruleNameAlreadyExists.
func cleanupUpdateErrorDetail(errCode, errMessage, configuredName, targetRuleID string) string {
	detail := fmt.Sprintf("%s : %s", errCode, errMessage)
	if errCode == "ruleNameAlreadyExists" {
		detail += fmt.Sprintf(
			"\n\nThe cleanup rule is identified as the sub-policy's LAST rule (currently rule id %s). "+
				"A different rule named %q already exists in this sub-policy, which usually means the "+
				"cleanup rule was manually moved out of the last position in the CMA. "+
				"Move the cleanup rule back to the bottom of the sub-policy (or re-export the policy JSON) and retry. "+
				"The cleanup rule is managed by this resource and is not intended to be moved outside Terraform.",
			targetRuleID, configuredName,
		)
	}
	return detail
}

// cleanupConfiguredName returns the name carried by the cleanup_rule config
// object, or "" when unset.
func cleanupConfiguredName(cleanupObj types.Object) string {
	if cleanupObj.IsNull() || cleanupObj.IsUnknown() {
		return ""
	}
	if nameAttr, ok := cleanupObj.Attributes()["name"].(types.String); ok && !nameAttr.IsNull() && !nameAttr.IsUnknown() {
		return nameAttr.ValueString()
	}
	return ""
}

// hydrateCleanupState encodes the cleanup rule into the model, refreshing
// cleanup_rule_id. The prior cleanup object (state or plan) seeds the hydrator
// so omitted optional fields stay stable, mirroring the scope handling.
func hydrateCleanupState(
	ctx context.Context,
	model *WanFirewallSubPolicy,
	cleanup *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule,
) (basetypes.ObjectValue, diag.Diagnostics) {
	model.CleanupRuleID = types.StringValue(cleanup.GetID())
	cleanupRule := wfrule.Rule{Rule: model.CleanupRule, At: model.At}
	cleanupState, hsDiags := wfrule.HydrateWanRuleState(ctx, cleanupRule, cleanup)
	cleanupState.ID = types.StringValue(cleanup.GetID())
	obj, objDiags := types.ObjectValueFrom(ctx, wfrule.WanFirewallRuleRuleAttrTypes, cleanupState)
	hsDiags.Append(objDiags...)
	return obj, hsDiags
}

// wanSubPolicyIDByName returns the id of the sub-policy with the given name that
// is not present in the exclude set. It is used to identify a freshly created
// sub-policy by diffing the sub-policy list before and after creation.
func wanSubPolicyIDByName(body *cato_go_sdk.Policy, name string, exclude map[string]struct{}) string {
	if body == nil {
		return ""
	}
	for _, sp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetSubPolicies() {
		info := sp.GetPolicy()
		if info.GetName() != name {
			continue
		}
		if _, skip := exclude[info.GetID()]; skip {
			continue
		}
		return info.GetID()
	}
	return ""
}

// wanSubPolicyIDs returns the set of sub-policy ids currently in the snapshot.
func wanSubPolicyIDs(body *cato_go_sdk.Policy) map[string]struct{} {
	ids := map[string]struct{}{}
	if body == nil {
		return ids
	}
	for _, sp := range body.GetPolicy().GetWanFirewall().GetPolicy().GetSubPolicies() {
		ids[sp.GetPolicy().GetID()] = struct{}{}
	}
	return ids
}

// objectRefByID builds an ObjectRefBy=ID reference input.
func objectRefByID(id string) *cato_models.WanFirewallPolicyRefInput {
	return &cato_models.WanFirewallPolicyRefInput{By: cato_models.ObjectRefByID, Input: id}
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	plan WanFirewallSubPolicy,
) (resultCreateScope basetypes.ObjectValue, resultDiags diag.Diagnostics) {
	scopeAttrs := maps.Clone(plan.Scope.Attributes())
	scopeAttrs["action"] = types.StringValue(string(cato_models.WanFirewallActionEnumAllow))
	scopeAttrs["name"] = plan.Name
	scopeAttrs["description"] = plan.Description
	resultCreateScope, resultDiags = types.ObjectValue(wfrule.WanFirewallRuleRuleAttrTypes, scopeAttrs)
	return resultCreateScope, resultDiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	plan WanFirewallSubPolicy,
	hydrated wfrule.HydrateWanAPITypes,
	before *cato_go_sdk.Policy,
) (resultExisting map[string]struct{}, resultAddInput cato_models.WanFirewallAddSubPolicyInput) {
	resultExisting = wanSubPolicyIDs(before)

	resultAddInput = cato_models.WanFirewallAddSubPolicyInput{
		At: hydrated.Create.At,
		Policy: &cato_models.WanFirewallAddSubPolicyDataInput{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
		},
		Scope: hydrated.Create.Rule,
	}
	// scope.tracking is a non-nullable PolicyTrackingInput! in
	// WanFirewallAddSubPolicyInput, but the API rejects enabled tracking
	// settings on a scoping rule — so when the (optional) tracking attribute is
	// omitted, send the schema's documented disabled default instead of null.
	if resultAddInput.Scope != nil && resultAddInput.Scope.Tracking == nil {
		resultAddInput.Scope.Tracking = &cato_models.PolicyTrackingInput{
			Event: &cato_models.PolicyRuleTrackingEventInput{Enabled: false},
			Alert: &cato_models.PolicyRuleTrackingAlertInput{
				Enabled:           false,
				Frequency:         cato_models.PolicyRuleTrackingFrequencyEnumHourly,
				MailingList:       []*cato_models.SubscriptionMailingListRefInput{},
				SubscriptionGroup: []*cato_models.SubscriptionGroupRefInput{},
				Webhook:           []*cato_models.SubscriptionWebhookRefInput{},
			},
		}
	}
	return resultExisting, resultAddInput
}

// wanFirewallSubPolicyStateFromPolicy finds the created sub-policy and encodes its published scope rule.
func wanFirewallSubPolicyStateFromPolicy(
	ctx context.Context,
	plan *WanFirewallSubPolicy,
	priorDiags *diag.Diagnostics,
	scopeRule wfrule.Rule,
	existing map[string]struct{},
	after *cato_go_sdk.Policy,
	diagnostics *diag.Diagnostics,
) (resultScopeObj basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	subID := wanSubPolicyIDByName(after, plan.Name.ValueString(), existing)
	if subID == "" {
		diagnostics.AddError("Sub-Policy Not Found", "Created WAN Firewall sub-policy could not be located after publish.")
		return resultScopeObj
	}
	scope := wanScopeRule(after, subID)
	if scope == nil {
		diagnostics.AddError("Scope Rule Not Found", fmt.Sprintf("No SUB_POLICY_SCOPE rule found for sub-policy %s.", subID))
		return resultScopeObj
	}

	plan.ID = types.StringValue(subID)
	plan.ScopeRuleID = types.StringValue(scope.GetID())
	scopeState, hsDiags := wfrule.HydrateWanRuleState(ctx, scopeRule, scope)
	diags.Append(hsDiags...)
	scopeState.ID = types.StringValue(scope.GetID())
	var objDiags diag.Diagnostics
	resultScopeObj, objDiags = types.ObjectValueFrom(ctx, wfrule.WanFirewallRuleRuleAttrTypes, scopeState)
	diags.Append(objDiags...)
	return resultScopeObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	state *WanFirewallSubPolicy,
	info *cato_go_sdk.Policy_Policy_WanFirewall_Policy_SubPolicies_Policy,
	scope *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule,
) (resultScopeObj basetypes.ObjectValue, resultDiags diag.Diagnostics) {
	state.Name = types.StringValue(info.GetName())
	// description is Required, so config always holds a value ("" when unset);
	// writing null here on an empty API description would diff against "" and,
	// with RequiresReplace, force needless replacement on every plan.
	state.Description = types.StringValue(info.GetDescription())
	state.ScopeRuleID = types.StringValue(scope.GetID())

	scopeRule := wfrule.Rule{Rule: state.Scope, At: state.At}
	scopeState, hsDiags := wfrule.HydrateWanRuleState(ctx, scopeRule, scope)
	resultDiags.Append(hsDiags...)
	scopeState.ID = types.StringValue(scope.GetID())
	resultScopeObj, objDiags := types.ObjectValueFrom(ctx, wfrule.WanFirewallRuleRuleAttrTypes, scopeState)
	resultDiags.Append(objDiags...)
	return resultScopeObj, resultDiags
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan WanFirewallSubPolicy,
	scopeRule wfrule.Rule,
	priorDiags *diag.Diagnostics,
	body *cato_go_sdk.Policy,
	diagnostics *diag.Diagnostics,
) (resultScopeObj basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	scope := wanScopeRule(body, plan.ID.ValueString())
	if scope == nil {
		diagnostics.AddError("Scope Rule Not Found", "Sub-policy scope rule not found after update.")
		return resultScopeObj
	}
	scopeState, hsDiags := wfrule.HydrateWanRuleState(ctx, scopeRule, scope)
	diags.Append(hsDiags...)
	scopeState.ID = types.StringValue(scope.GetID())
	var objDiags diag.Diagnostics
	resultScopeObj, objDiags = types.ObjectValueFrom(ctx, wfrule.WanFirewallRuleRuleAttrTypes, scopeState)
	diags.Append(objDiags...)
	return resultScopeObj
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	state WanFirewallSubPolicy,
) (resultRemoveInput cato_models.WanFirewallRemoveSubPolicyInput) {
	resultRemoveInput = cato_models.WanFirewallRemoveSubPolicyInput{Ref: objectRefByID(state.ID.ValueString())}
	return resultRemoveInput
}
