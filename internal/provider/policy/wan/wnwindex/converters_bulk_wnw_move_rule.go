package wnwindex

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// convertMoveWanNetworkRulesAndSectionsValues converts between the Terraform model and SDK values.
func convertMoveWanNetworkRulesAndSectionsValues(
	ctx context.Context,
	plan WanNetworkRulesIndex,
	priorDiags *diag.Diagnostics,
	sectionIDList map[string]string,
	sectionIndexAPIData *cato_go_sdk.WanNetworkPolicy,
) (resultSectionListFromPlan []WanNetworkRulesSectionDataIndex) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	for _, v := range sectionIndexAPIData.Policy.WanNetwork.Policy.Sections {
		sectionIDList[v.Section.Name] = v.Section.ID
	}

	resultSectionListFromPlan = make([]WanNetworkRulesSectionDataIndex, 0)

	// Convert map to slice for processing
	sectionDataMapElements := plan.SectionData.Elements()
	for _, sectionValue := range sectionDataMapElements {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex WanNetworkRulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionDataTmp := WanNetworkRulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		}
		resultSectionListFromPlan = append(resultSectionListFromPlan, sectionDataTmp)
	}
	return resultSectionListFromPlan
}

// convertMoveWanNetworkRulesAndSectionsValues2 converts between the Terraform model and SDK values.
func convertMoveWanNetworkRulesAndSectionsValues2(
	priorSectionObjectsMap *basetypes.MapValue,
	priorRuleObjectsMap *basetypes.MapValue,
	priorDiags *diag.Diagnostics,
	ruleObjectMap map[string]attr.Value,
	sectionObjectMap map[string]attr.Value,
) {
	sectionObjectsMap := *priorSectionObjectsMap
	defer func() { *priorSectionObjectsMap = sectionObjectsMap }()
	ruleObjectsMap := *priorRuleObjectsMap
	defer func() { *priorRuleObjectsMap = ruleObjectsMap }()
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	var sectionMapDiags diag.Diagnostics
	sectionObjectsMap, sectionMapDiags = types.MapValue(
		WanNetworkSectionIndexResourceObjectTypes,
		sectionObjectMap,
	)
	diags = append(diags, sectionMapDiags...)

	var ruleMapDiags diag.Diagnostics
	ruleObjectsMap, ruleMapDiags = types.MapValue(
		WanNetworkRuleIndexResourceObjectTypes,
		ruleObjectMap,
	)
	diags = append(diags, ruleMapDiags...)
}

// convertMoveWanNetworkRulesAndSectionsValues3 converts between the Terraform model and SDK values.
func convertMoveWanNetworkRulesAndSectionsValues3(
	plan WanNetworkRulesIndex,
	priorListOfSectionNames *[]string,
	sectionIDList map[string]string,
	originalCurrentSectionID *string,
	workingSectionName WanNetworkRulesSectionDataIndex,
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
	// initializing currentSectionId to the SectionToStartAfterID otherwise set to id of first section for next in list
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

// convertMoveWanNetworkRulesAndSectionsValues4 converts between the Terraform model and SDK values.
func convertMoveWanNetworkRulesAndSectionsValues4(
	priorDiags *diag.Diagnostics,
	sectionIDList map[string]string,
	priorCurrentSectionID *string,
	sectionObjectMap map[string]attr.Value,
	workingSectionName WanNetworkRulesSectionDataIndex,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	currentSectionID := *priorCurrentSectionID
	defer func() { *priorCurrentSectionID = currentSectionID }()
	sectionIndexStateData, diagsSection := types.ObjectValue(
		WanNetworkSectionIndexResourceAttrTypes,
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

// convertMoveWanNetworkRulesAndSectionsValues5 converts between the Terraform model and SDK values.
func convertMoveWanNetworkRulesAndSectionsValues5(
	priorDiags *diag.Diagnostics,
	ruleObjectMap map[string]attr.Value,
	ruleListFromPlan []WanNetworkRulesRuleDataIndex,
	ruleNameIDMap map[string]string,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	for _, ruleFromPlan := range ruleListFromPlan {
		ruleID := ruleNameIDMap[ruleFromPlan.RuleName]
		ruleIndexStateData, diagsSection := types.ObjectValue(
			WanNetworkRuleIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":               types.StringValue(ruleID),
				"index_in_section": types.Int64Value(ruleFromPlan.IndexInSection),
				"section_name":     types.StringValue(ruleFromPlan.SectionName),
				"rule_name":        types.StringValue(ruleFromPlan.RuleName),
				"description":      types.StringValue(ruleFromPlan.Description),
				"enabled":          types.BoolValue(ruleFromPlan.Enabled),
			},
		)
		diags = append(diags, diagsSection...)
		ruleObjectMap[ruleFromPlan.RuleName] = ruleIndexStateData
	}
}

// convertMoveWanNetworkRulesAndSectionsValues6 converts between the Terraform model and SDK values.
func convertMoveWanNetworkRulesAndSectionsValues6(
	ctx context.Context,
	priorDiags *diag.Diagnostics,
	priorRuleListFromPlan *[]WanNetworkRulesRuleDataIndex,
	ruleValue attr.Value,
) (resultRulenDataTmp WanNetworkRulesRuleDataIndex) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	ruleListFromPlan := *priorRuleListFromPlan
	defer func() { *priorRuleListFromPlan = ruleListFromPlan }()
	ruleObject := ruleValue.(types.Object)
	var planSourceRuleIndex WanNetworkRulesRuleItemIndex
	diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

	resultRulenDataTmp = WanNetworkRulesRuleDataIndex{
		IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
		RuleName:       planSourceRuleIndex.RuleName.ValueString(),
		SectionName:    planSourceRuleIndex.SectionName.ValueString(),
		Description:    planSourceRuleIndex.Description.ValueString(),
		Enabled:        planSourceRuleIndex.Enabled.ValueBool(),
	}
	ruleListFromPlan = append(ruleListFromPlan, resultRulenDataTmp)
	return resultRulenDataTmp
}

// convertMoveWanNetworkRulesAndSectionsValues8 converts between the Terraform model and SDK values.
func convertMoveWanNetworkRulesAndSectionsValues8(
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
