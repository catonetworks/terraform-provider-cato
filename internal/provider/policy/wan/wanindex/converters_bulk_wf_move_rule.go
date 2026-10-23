package wanindex

import (
	"context"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func buildWanRuleIndexStateData(
	ruleFromPlan WanRulesRuleDataIndex,
	ruleNameIDMap map[string]string,
	ruleNameDescriptionMap map[string]string,
	ruleNameEnabledMap map[string]bool,
) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		WanRuleIndexResourceAttrTypes,
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

func wanFirewallMoveSectionError(resp *cato_go_sdk.PolicyWanFirewallMoveSection, callErr error) error {
	if callErr != nil {
		return callErr
	}
	if resp == nil || resp.GetPolicy() == nil || resp.GetPolicy().GetWanFirewall() == nil {
		return nil
	}
	ms := resp.GetPolicy().GetWanFirewall().GetMoveSection()
	if ms == nil {
		return nil
	}
	if errs := ms.GetErrors(); len(errs) > 0 {
		var b strings.Builder
		for _, e := range errs {
			if e == nil {
				continue
			}
			if b.Len() > 0 {
				b.WriteString("; ")
			}
			_, _ = fmt.Fprintf(&b, "%s: %s", utils.GqlOptionalStr(e.GetErrorCode()), utils.GqlOptionalStr(e.GetErrorMessage()))
		}
		return fmt.Errorf("WAN firewall moveSection errors: %s", b.String())
	}
	if st := ms.GetStatus(); st != nil && *st != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("WAN firewall moveSection status %q", string(*st))
	}
	return nil
}

// convertMoveWanRulesAndSectionsValues converts between the Terraform model and SDK values.
func convertMoveWanRulesAndSectionsValues(
	ctx context.Context,
	plan WanRulesIndex,
	priorDiags *[]diag.Diagnostic,
	sectionIDList map[string]string,
	sectionIndexAPIData *cato_go_sdk.WanSectionsIndexPolicy,
) (resultSectionListFromPlan []WanRulesSectionDataIndex) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	for _, v := range sectionIndexAPIData.Policy.WanFirewall.Policy.Sections {
		sectionIDList[v.Section.Name] = v.Section.ID
	}

	resultSectionListFromPlan = make([]WanRulesSectionDataIndex, 0)

	// Convert map to slice for processing
	sectionDataMapElements := plan.SectionData.Elements()
	for _, sectionValue := range sectionDataMapElements {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex WanRulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionDataTmp := WanRulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		}
		resultSectionListFromPlan = append(resultSectionListFromPlan, sectionDataTmp)
	}
	return resultSectionListFromPlan
}

// convertMoveWanRulesAndSectionsValues2 converts between the Terraform model and SDK values.
func convertMoveWanRulesAndSectionsValues2(
	priorDiags *[]diag.Diagnostic,
	ruleObjectMap map[string]attr.Value,
	sectionObjectMap map[string]attr.Value,
) (resultSectionObjectsMap basetypes.MapValue, resultRuleObjectsMap basetypes.MapValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultSectionObjectsMap, sectionMapDiags := types.MapValue(
		WanSectionIndexResourceObjectTypes,
		sectionObjectMap,
	)
	diags = append(diags, sectionMapDiags...)

	resultRuleObjectsMap, ruleMapDiags := types.MapValue(
		WanRuleIndexResourceObjectTypes,
		ruleObjectMap,
	)
	diags = append(diags, ruleMapDiags...)
	return resultSectionObjectsMap, resultRuleObjectsMap
}

// convertMoveWanRulesAndSectionsValues3 converts between the Terraform model and SDK values.
func convertMoveWanRulesAndSectionsValues3(
	plan WanRulesIndex,
	sectionIDList map[string]string,
	originalCurrentSectionID *string,
	workingSectionName WanRulesSectionDataIndex,
) (resultPolicyMoveSectionInputInt cato_models.PolicyMoveSectionInput) {
	currentSectionID := *originalCurrentSectionID
	defer func() { *originalCurrentSectionID = currentSectionID }()

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

// convertMoveWanRulesAndSectionsValues4 converts between the Terraform model and SDK values.
func convertMoveWanRulesAndSectionsValues4(
	priorDiags *[]diag.Diagnostic,
	sectionIDList map[string]string,
	priorCurrentSectionID *string,
	sectionObjectMap map[string]attr.Value,
	workingSectionName WanRulesSectionDataIndex,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	currentSectionID := *priorCurrentSectionID
	defer func() { *priorCurrentSectionID = currentSectionID }()
	sectionIndexStateData, diagsSection := types.ObjectValue(
		WanSectionIndexResourceAttrTypes,
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

// convertMoveWanRulesAndSectionsValues5 converts between the Terraform model and SDK values.
func convertMoveWanRulesAndSectionsValues5(
	ctx context.Context,
	priorDiags *[]diag.Diagnostic,
	priorRuleListFromPlan *[]WanRulesRuleDataIndex,
	ruleValue attr.Value,
) (resultRulenDataTmp WanRulesRuleDataIndex) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	ruleListFromPlan := *priorRuleListFromPlan
	defer func() { *priorRuleListFromPlan = ruleListFromPlan }()
	ruleObject := ruleValue.(types.Object)
	var planSourceRuleIndex WanRulesRuleItemIndex
	diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

	resultRulenDataTmp = WanRulesRuleDataIndex{
		IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
		RuleName:       planSourceRuleIndex.RuleName.ValueString(),
		SectionName:    planSourceRuleIndex.SectionName.ValueString(),
		Description:    planSourceRuleIndex.Description.ValueString(),
		Enabled:        planSourceRuleIndex.Enabled.ValueBool(),
	}
	ruleListFromPlan = append(ruleListFromPlan, resultRulenDataTmp)
	return resultRulenDataTmp
}
