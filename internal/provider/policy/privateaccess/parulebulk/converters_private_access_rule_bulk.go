package parulebulk

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
)

// privateAccessMoveRuleInput updates local rule order and builds the SDK move request.
func privateAccessMoveRuleInput(
	currentRules []*PrivateAccessBulkRule,
	ruleID string,
	newPosition int,
	myRule *PrivateAccessBulkRule,
	curPossition int,
) (resultInput cato_models.PolicyMoveRuleInput) {
	for i := curPossition; i > newPosition; i-- {
		currentRules[i] = currentRules[i-1]
	}
	currentRules[newPosition] = myRule

	// Prepare input for API to move the rule

	if newPosition == 0 {
		resultInput = cato_models.PolicyMoveRuleInput{
			ID: ruleID,
			To: &cato_models.PolicyRulePositionInput{Position: new(cato_models.PolicyRulePositionEnumFirstInPolicy)},
		}
	} else {
		resultInput = cato_models.PolicyMoveRuleInput{
			ID: ruleID,
			To: &cato_models.PolicyRulePositionInput{
				Position: new(cato_models.PolicyRulePositionEnumAfterRule),
				Ref:      convert.KnownStringPointer(currentRules[newPosition-1].ID),
			},
		}
	}
	return resultInput
}

// privateAccessRuleMapFromObjects encodes rule objects as the Terraform rule_data map.
func privateAccessRuleMapFromObjects(
	ctx context.Context,
	diags *diag.Diagnostics,
	apiRulesTf map[string]types.Object,
) (resultRulesMap basetypes.MapValue) {
	resultRulesMap, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: PrivateAccessBulkRuleTypes}, apiRulesTf)
	diags.Append(objDiags...)
	return resultRulesMap
}

// privateAccessBulkStateFromRules builds bulk state while preserving the configured publish value.
func privateAccessBulkStateFromRules(
	plan *PrivateAccessRuleBulkModel,
	rulesMap basetypes.MapValue,
) (resultState *PrivateAccessRuleBulkModel) {
	resultState = &PrivateAccessRuleBulkModel{
		RuleData: rulesMap,
		Publish:  types.Int64Value(plan.Publish.ValueInt64()),
		ID:       types.StringValue("0"),
	}
	return resultState
}

// privateAccessRuleFromObject decodes a configured Terraform rule and appends diagnostics.
func privateAccessRuleFromObject(
	ctx context.Context,
	diags *diag.Diagnostics,
	obj types.Object,
) (resultTfRule PrivateAccessBulkRule) {
	diags.Append(obj.As(ctx, &resultTfRule, basetypes.ObjectAsOptions{})...)
	return resultTfRule
}

// privateAccessRuleFromResponse matches an SDK rule with its configured Terraform rule.
func privateAccessRuleFromResponse(
	stateRules map[string]PrivateAccessBulkRule,
	polRule *cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules,
) (
	resultAPIRule cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule,
	resultTfRule *PrivateAccessBulkRule,
	resultStateRule PrivateAccessBulkRule,
	resultOk bool,
) {
	resultAPIRule = polRule.Rule
	resultTfRule = &PrivateAccessBulkRule{
		ID:       types.StringValue(resultAPIRule.ID),
		Name:     types.StringValue(resultAPIRule.Name),
		CMAIndex: resultAPIRule.Index,
	}
	resultStateRule, resultOk = stateRules[resultAPIRule.Name]
	return resultAPIRule, resultTfRule, resultStateRule, resultOk
}

// privateAccessRuleObjectFromModel encodes an SDK-enriched rule with its configured index.
func privateAccessRuleObjectFromModel(
	ctx context.Context,
	diags *diag.Diagnostics,
	tfRule *PrivateAccessBulkRule,
	stateRule PrivateAccessBulkRule,
) (resultRuleObj basetypes.ObjectValue) {
	tfRule.Index = stateRule.Index
	resultRuleObj, objDiags := types.ObjectValueFrom(ctx, PrivateAccessBulkRuleTypes, tfRule)
	diags.Append(objDiags...)
	return resultRuleObj
}

// includeUnmanagedRules preserves other policy rules in the ordering algorithm,
// while requiring every configured rule to be present in the hydrated state.
func includeUnmanagedRules(
	stateRules map[string]PrivateAccessBulkRule,
	apiRulesTf map[string]types.Object,
	apiRulesGo map[string]*PrivateAccessBulkRule,
	unmanaged []*PrivateAccessBulkRule,
) error {
	unmanagedIndex := int64(len(stateRules))
	for _, rule := range stateRules {
		if rule.Index.ValueInt64() >= unmanagedIndex {
			unmanagedIndex = rule.Index.ValueInt64() + 1
		}
	}
	for name := range stateRules {
		if _, ok := apiRulesTf[name]; !ok {
			return fmt.Errorf("configured private access rule %q not found", name)
		}
	}
	slices.SortFunc(unmanaged, func(a, b *PrivateAccessBulkRule) int { return cmp.Compare(a.CMAIndex, b.CMAIndex) })
	for i, rule := range unmanaged {
		rule.Index = types.Int64Value(unmanagedIndex + int64(i))
		apiRulesGo[rule.Name.ValueString()] = rule
	}

	return nil
}
