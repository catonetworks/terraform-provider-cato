package tlsindex

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func buildTLSRuleIndexStateData(
	ruleFromPlan RulesRuleDataIndex,
	ruleNameIDMap map[string]string,
	ruleNameDescriptionMap map[string]string,
	ruleNameEnabledMap map[string]bool,
) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		TLSRuleIndexResourceAttrTypes,
		map[string]attr.Value{
			"id":               types.StringValue(ruleNameIDMap[ruleFromPlan.RuleName]),
			"index_in_section": types.Int64Value(ruleFromPlan.IndexInSection),
			"section_name":     types.StringValue(ruleFromPlan.SectionName),
			"rule_name":        types.StringValue(ruleFromPlan.RuleName),
			"description":      types.StringValue(ruleNameDescriptionMap[ruleFromPlan.RuleName]),
			"enabled":          types.BoolValue(ruleNameEnabledMap[ruleFromPlan.RuleName]),
		},
	)
}

// convertMoveTLSRulesAndSectionsValues converts between the Terraform model and SDK values.
func convertMoveTLSRulesAndSectionsValues(
	ctx context.Context,
	plan RulesIndex,
	priorDiags *[]diag.Diagnostic,
	sectionIDList map[string]string,
	sectionIndexAPIData *cato_go_sdk.Tlsinspectpolicy,
) (resultSectionListFromPlan []RulesSectionDataIndex) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	for _, v := range sectionIndexAPIData.Policy.TLSInspect.Policy.Sections {
		sectionIDList[v.Section.Name] = v.Section.ID
	}

	resultSectionListFromPlan = make([]RulesSectionDataIndex, 0)

	// Convert map to slice for processing
	sectionDataMapElements := plan.SectionData.Elements()
	for _, sectionValue := range sectionDataMapElements {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex RulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionDataTmp := RulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		}
		resultSectionListFromPlan = append(resultSectionListFromPlan, sectionDataTmp)
	}
	return resultSectionListFromPlan
}

// convertMoveTLSRulesAndSectionsValues2 converts between the Terraform model and SDK values.
func convertMoveTLSRulesAndSectionsValues2(
	priorDiags *[]diag.Diagnostic,
	ruleObjectMap map[string]attr.Value,
	sectionObjectMap map[string]attr.Value,
) (resultSectionObjectsMap basetypes.MapValue, resultRuleObjectsMap basetypes.MapValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultSectionObjectsMap, sectionMapDiags := types.MapValue(
		TLSSectionIndexResourceObjectTypes,
		sectionObjectMap,
	)
	diags = append(diags, sectionMapDiags...)

	resultRuleObjectsMap, ruleMapDiags := types.MapValue(
		TLSRuleIndexResourceObjectTypes,
		ruleObjectMap,
	)
	diags = append(diags, ruleMapDiags...)
	return resultSectionObjectsMap, resultRuleObjectsMap
}

// convertMoveTLSRulesAndSectionsValues3 converts between the Terraform model and SDK values.
func convertMoveTLSRulesAndSectionsValues3(
	plan RulesIndex,
	priorListOfSectionNames *[]string,
	sectionIDList map[string]string,
	originalCurrentSectionID *string,
	workingSectionName RulesSectionDataIndex,
) (resultPolicyMoveSectionInputInt cato_models.PolicyMoveSectionInput) {
	currentSectionID := *originalCurrentSectionID
	defer func() { *originalCurrentSectionID = currentSectionID }()

	listOfSectionNames := *priorListOfSectionNames
	defer func() { *priorListOfSectionNames = listOfSectionNames }()
	listOfSectionNames = append(listOfSectionNames, workingSectionName.SectionName)
	resultPolicyMoveSectionInputInt = cato_models.PolicyMoveSectionInput{
		ID: sectionIDList[workingSectionName.SectionName],
	}

	// For the first element, check for sectionToStartAfterId, if not, start at last LAST_IN_POLICY
	// initializing currentSectionID to the SectionToStartAfterID otherwise set to id of first section for next in list
	if currentSectionID == "" {
		if plan.SectionToStartAfterID.ValueString() != "" {
			resultPolicyMoveSectionInputInt.To = &cato_models.PolicySectionPositionInput{
				Ref:      plan.SectionToStartAfterID.ValueStringPointer(),
				Position: "AFTER_SECTION",
			}
		} else {
			resultPolicyMoveSectionInputInt.To = &cato_models.PolicySectionPositionInput{
				Position: "LAST_IN_POLICY",
			}
		}
	} else {
		resultPolicyMoveSectionInputInt.To = &cato_models.PolicySectionPositionInput{
			Ref:      &currentSectionID,
			Position: "AFTER_SECTION",
		}
	}
	return resultPolicyMoveSectionInputInt
}

// convertMoveTLSRulesAndSectionsValues4 converts between the Terraform model and SDK values.
func convertMoveTLSRulesAndSectionsValues4(
	priorDiags *[]diag.Diagnostic,
	sectionIDList map[string]string,
	priorCurrentSectionID *string,
	sectionObjectMap map[string]attr.Value,
	workingSectionName RulesSectionDataIndex,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	currentSectionID := *priorCurrentSectionID
	defer func() { *priorCurrentSectionID = currentSectionID }()
	sectionIndexStateData, diagsSection := types.ObjectValue(
		TLSSectionIndexResourceAttrTypes,
		map[string]attr.Value{
			"id":            types.StringValue(sectionIDList[workingSectionName.SectionName]),
			"section_name":  types.StringValue(workingSectionName.SectionName),
			"section_index": types.Int64Value(workingSectionName.SectionIndex),
		},
	)
	diags = append(diags, diagsSection...)

	sectionObjectMap[workingSectionName.SectionName] = sectionIndexStateData

	currentSectionID = sectionIDList[workingSectionName.SectionName]
}

// convertMoveTLSRulesAndSectionsValues5 converts between the Terraform model and SDK values.
func convertMoveTLSRulesAndSectionsValues5(
	ctx context.Context,
	priorDiags *[]diag.Diagnostic,
	priorRuleListFromPlan *[]RulesRuleDataIndex,
	ruleValue attr.Value,
) (resultRulenDataTmp RulesRuleDataIndex) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	ruleListFromPlan := *priorRuleListFromPlan
	defer func() { *priorRuleListFromPlan = ruleListFromPlan }()
	ruleObject := ruleValue.(types.Object)
	var planSourceRuleIndex RulesRuleItemIndex
	diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

	resultRulenDataTmp = RulesRuleDataIndex{
		IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
		RuleName:       planSourceRuleIndex.RuleName.ValueString(),
		SectionName:    planSourceRuleIndex.SectionName.ValueString(),
		Description:    planSourceRuleIndex.Description.ValueString(),
		Enabled:        planSourceRuleIndex.Enabled.ValueBool(),
	}
	ruleListFromPlan = append(ruleListFromPlan, resultRulenDataTmp)
	return resultRulenDataTmp
}

// convertMoveTLSRulesAndSectionsValues7 converts between the Terraform model and SDK values.
func convertMoveTLSRulesAndSectionsValues7(
	ruleNameIDMap map[string]string,
	mapRuleIndexToRuleName map[int64]string,
	x int,
	toPosition *cato_models.PolicyRulePositionInput,
) (resultMoveRuleConfig cato_models.PolicyMoveRuleInput) {
	resultMoveRuleConfig = cato_models.PolicyMoveRuleInput{
		ID: ruleNameIDMap[mapRuleIndexToRuleName[int64(x)]],
		To: toPosition,
	}
	return resultMoveRuleConfig
}
