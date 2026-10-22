package parulebulk

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
)

// convertMoveToPositionValues converts between the Terraform model and SDK values.
func convertMoveToPositionValues(
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

// convertHydratePrivAccessRuleBulkStateValues converts between the Terraform model and SDK values.
func convertHydratePrivAccessRuleBulkStateValues(
	ctx context.Context,
	diags diag.Diagnostics,
	apiRulesTf map[string]types.Object,
) (resultRulesMap basetypes.MapValue) {
	resultRulesMap, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: PrivateAccessBulkRuleTypes}, apiRulesTf)
	diags.Append(objDiags...)
	return resultRulesMap
}

// convertHydratePrivAccessRuleBulkStateValues2 converts between the Terraform model and SDK values.
func convertHydratePrivAccessRuleBulkStateValues2(
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

// convertHydratePrivAccessRuleBulkStateValues3 converts between the Terraform model and SDK values.
func convertHydratePrivAccessRuleBulkStateValues3(
	ctx context.Context,
	diags diag.Diagnostics,
	obj types.Object,
) (resultTfRule PrivateAccessBulkRule) {
	diags.Append(obj.As(ctx, &resultTfRule, basetypes.ObjectAsOptions{})...)
	return resultTfRule
}

// convertHydratePrivAccessRuleBulkStateValues4 converts between the Terraform model and SDK values.
func convertHydratePrivAccessRuleBulkStateValues4(
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

// convertHydratePrivAccessRuleBulkStateValues5 converts between the Terraform model and SDK values.
func convertHydratePrivAccessRuleBulkStateValues5(
	ctx context.Context,
	diags diag.Diagnostics,
	tfRule *PrivateAccessBulkRule,
	stateRule PrivateAccessBulkRule,
) (resultRuleObj basetypes.ObjectValue) {
	tfRule.Index = stateRule.Index
	resultRuleObj, objDiags := types.ObjectValueFrom(ctx, PrivateAccessBulkRuleTypes, tfRule)
	diags.Append(objDiags...)
	return resultRuleObj
}
