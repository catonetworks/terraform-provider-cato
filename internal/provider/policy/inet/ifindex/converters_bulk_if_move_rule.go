package ifindex

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func resolveBulkRuleEnablement(prior, plan types.Map, index *cato_go_sdk.IfwRulesIndexPolicy) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics
	if index == nil || index.Policy == nil || index.Policy.InternetFirewall == nil {
		diags.AddError("Could not read Internet Firewall rule enablement", "Cato returned an empty Internet Firewall policy.")
		return prior, diags
	}

	enabledByName := make(map[string]bool)
	for _, item := range index.Policy.InternetFirewall.Policy.Rules {
		enabledByName[item.Rule.Name] = item.Rule.Enabled
	}

	rules := prior.Elements()
	for key, planned := range plan.Elements() {
		if planned.IsNull() || planned.IsUnknown() {
			continue
		}
		enabled := planned.(types.Object).Attributes()["enabled"]
		attrs := rules[key].(types.Object).Attributes()
		if enabled.IsUnknown() {
			name := attrs["rule_name"].(types.String).ValueString()
			apiEnabled, ok := enabledByName[name]
			if !ok {
				diags.AddError("Internet firewall rule not found", fmt.Sprintf("Could not read enablement for rule %q.", name))
				return prior, diags
			}
			enabled = types.BoolValue(apiEnabled)
		}
		attrs["enabled"] = enabled
		rule, ruleDiags := types.ObjectValue(IfwRuleIndexResourceAttrTypes, attrs)
		diags.Append(ruleDiags...)
		rules[key] = rule
	}
	state, mapDiags := types.MapValue(IfwRuleIndexResourceObjectTypes, rules)
	diags.Append(mapDiags...)
	return state, diags
}

func buildIfwSectionMoves(
	sections []IfwRulesSectionDataIndex,
	sectionIDs map[string]string,
	protectedSectionIDs map[string]struct{},
	startAfterID string,
) []ifwSectionMove {
	moves := make([]ifwSectionMove, 0, len(sections))
	currentSectionID := ""

	for _, section := range sections {
		sectionID := sectionIDs[section.SectionName]
		input := cato_models.PolicyMoveSectionInput{ID: sectionID}

		if currentSectionID == "" {
			if startAfterID != "" {
				ref := startAfterID
				input.To = &cato_models.PolicySectionPositionInput{
					Ref:      &ref,
					Position: "AFTER_SECTION",
				}
			} else {
				input.To = &cato_models.PolicySectionPositionInput{Position: "LAST_IN_POLICY"}
			}
		} else {
			ref := currentSectionID
			input.To = &cato_models.PolicySectionPositionInput{
				Ref:      &ref,
				Position: "AFTER_SECTION",
			}
		}

		_, protected := protectedSectionIDs[sectionID]
		moves = append(moves, ifwSectionMove{
			section:   section,
			input:     input,
			protected: protected,
		})
		currentSectionID = sectionID
	}

	return moves
}

func buildIfwRuleIndexStateData(
	ruleFromPlan IfwRulesRuleDataIndex,
	ruleNameIDMap map[string]string,
	ruleNameDescriptionMap map[string]string,
	ruleNameEnabledMap map[string]bool,
) (basetypes.ObjectValue, diag.Diagnostics) {
	enabled := ruleFromPlan.Enabled
	if enabled.IsNull() || enabled.IsUnknown() {
		enabled = types.BoolValue(ruleNameEnabledMap[ruleFromPlan.RuleName])
	}

	return types.ObjectValue(
		IfwRuleIndexResourceAttrTypes,
		map[string]attr.Value{
			"id":               types.StringValue(ruleNameIDMap[ruleFromPlan.RuleName]),
			"index_in_section": types.Int64Value(ruleFromPlan.IndexInSection),
			"section_name":     types.StringValue(ruleFromPlan.SectionName),
			"rule_name":        types.StringValue(ruleFromPlan.RuleName),
			"description":      types.StringValue(ruleNameDescriptionMap[ruleFromPlan.RuleName]),
			"enabled":          enabled,
		},
	)
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	state *IfwRulesIndex,
	priorDiags *diag.Diagnostics,
	sectionIDMap map[string]string,
	ruleIDMap map[string]string,
	ruleDescriptionMap map[string]string,
	ruleIndexAPIData *cato_go_sdk.IfwRulesIndexPolicy,
	configuredEnabled map[string]struct{},
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	ruleEnabledMap := make(map[string]bool)
	for _, v := range ruleIndexAPIData.Policy.InternetFirewall.Policy.Rules {
		ruleIDMap[v.Rule.Name] = v.Rule.ID
		ruleDescriptionMap[v.Rule.Name] = v.Rule.Description
		ruleEnabledMap[v.Rule.Name] = v.Rule.Enabled
	}

	// Update only computed fields in existing state, preserving planned configuration
	sectionObjectMap := make(map[string]attr.Value)
	if !state.SectionData.IsNull() && !state.SectionData.IsUnknown() {
		sectionDataMapElements := state.SectionData.Elements()
		for sectionName, sectionValue := range sectionDataMapElements {
			sectionObject := sectionValue.(types.Object)
			var existingSection IfwRulesSectionItemIndex
			diags = append(diags, sectionObject.As(ctx, &existingSection, basetypes.ObjectAsOptions{})...)

			// Preserve planned values, only update computed ID
			sectionIndexStateData, sectionDiags := types.ObjectValue(
				IfwSectionIndexResourceAttrTypes,
				map[string]attr.Value{
					"id":            types.StringValue(sectionIDMap[sectionName]),
					"section_name":  existingSection.SectionName,  // Preserve planned value
					"section_index": existingSection.SectionIndex, // Preserve planned value
				},
			)
			diags = append(diags, sectionDiags...)
			sectionObjectMap[sectionName] = sectionIndexStateData
		}
	}

	sectionObjectsMap, sectionMapDiags := types.MapValue(
		IfwSectionIndexResourceObjectTypes,
		sectionObjectMap,
	)
	diags = append(diags, sectionMapDiags...)
	state.SectionData = sectionObjectsMap

	// Update only computed fields in existing rule state, preserving planned configuration
	ruleObjectMap := make(map[string]attr.Value)
	if !state.RuleData.IsNull() && !state.RuleData.IsUnknown() {
		ruleDataMapElements := state.RuleData.Elements()
		for ruleName, ruleValue := range ruleDataMapElements {
			ruleObject := ruleValue.(types.Object)
			var existingRule IfwRulesRuleItemIndex
			diags = append(diags, ruleObject.As(ctx, &existingRule, basetypes.ObjectAsOptions{})...)

			enabled := existingRule.Enabled
			if _, configured := configuredEnabled[ruleName]; !configured {
				enabled = types.BoolValue(ruleEnabledMap[ruleName])
			}

			// Preserve planned values, only update computed fields
			ruleIndexStateData, ruleDiags := types.ObjectValue(
				IfwRuleIndexResourceAttrTypes,
				map[string]attr.Value{
					"id":               types.StringValue(ruleIDMap[ruleName]),
					"index_in_section": existingRule.IndexInSection,                     // Preserve planned value
					"section_name":     existingRule.SectionName,                        // Preserve planned value
					"rule_name":        existingRule.RuleName,                           // Preserve planned value
					"description":      types.StringValue(ruleDescriptionMap[ruleName]), // Update computed value
					"enabled":          enabled,
				},
			)
			diags = append(diags, ruleDiags...)
			ruleObjectMap[ruleName] = ruleIndexStateData
		}
	}

	ruleObjectsMap, ruleMapDiags := types.MapValue(
		IfwRuleIndexResourceObjectTypes,
		ruleObjectMap,
	)
	diags = append(diags, ruleMapDiags...)
	state.RuleData = ruleObjectsMap
}

// convertMoveIfwRulesAndSectionsValues converts between the Terraform model and SDK values.
func convertMoveIfwRulesAndSectionsValues(
	ctx context.Context,
	plan IfwRulesIndex,
	priorDiags *[]diag.Diagnostic,
) (resultSectionListFromPlan []IfwRulesSectionDataIndex) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultSectionListFromPlan = make([]IfwRulesSectionDataIndex, 0)

	// Convert map to slice for processing
	sectionDataMapElements := plan.SectionData.Elements()
	for _, sectionValue := range sectionDataMapElements {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex IfwRulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionDataTmp := IfwRulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		}
		resultSectionListFromPlan = append(resultSectionListFromPlan, sectionDataTmp)
	}
	return resultSectionListFromPlan
}

// convertMoveIfwRulesAndSectionsValues2 converts between the Terraform model and SDK values.
func convertMoveIfwRulesAndSectionsValues2(
	priorDiags *[]diag.Diagnostic,
	ruleObjectMap map[string]attr.Value,
	sectionObjectMap map[string]attr.Value,
) (resultSectionObjectsMap basetypes.MapValue, resultRuleObjectsMap basetypes.MapValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultSectionObjectsMap, sectionMapDiags := types.MapValue(
		IfwSectionIndexResourceObjectTypes,
		sectionObjectMap,
	)
	diags = append(diags, sectionMapDiags...)

	resultRuleObjectsMap, ruleMapDiags := types.MapValue(
		IfwRuleIndexResourceObjectTypes,
		ruleObjectMap,
	)
	diags = append(diags, ruleMapDiags...)
	return resultSectionObjectsMap, resultRuleObjectsMap
}

// convertMoveIfwRulesAndSectionsValues4 converts between the Terraform model and SDK values.
func convertMoveIfwRulesAndSectionsValues4(
	priorDiags *[]diag.Diagnostic,
	sectionIDList map[string]string,
	sectionObjectMap map[string]attr.Value,
	move ifwSectionMove,
) (
	resultWorkingSectionName IfwRulesSectionDataIndex,
	resultPolicyMoveSectionInputInt cato_models.PolicyMoveSectionInput,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultWorkingSectionName = move.section
	resultPolicyMoveSectionInputInt = move.input

	sectionIndexStateData, diagsSection := types.ObjectValue(
		IfwSectionIndexResourceAttrTypes,
		map[string]attr.Value{
			"id":            types.StringValue(sectionIDList[resultWorkingSectionName.SectionName]),
			"section_name":  types.StringValue(resultWorkingSectionName.SectionName),
			"section_index": types.Int64Value(resultWorkingSectionName.SectionIndex),
		},
	)
	diags = append(diags, diagsSection...)
	sectionObjectMap[resultWorkingSectionName.SectionName] = sectionIndexStateData
	return resultWorkingSectionName, resultPolicyMoveSectionInputInt
}

// convertMoveIfwRulesAndSectionsValues5 converts between the Terraform model and SDK values.
func convertMoveIfwRulesAndSectionsValues5(
	ctx context.Context,
	plan IfwRulesIndex,
	priorDiags *[]diag.Diagnostic,
	sectionListFromPlan []IfwRulesSectionDataIndex,
) (resultRuleListFromPlan []IfwRulesRuleDataIndex, resultSectionIndexMap map[string]int64) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultRuleListFromPlan = make([]IfwRulesRuleDataIndex, 0)

	// Convert map to slice for processing
	ruleDataMapElements := plan.RuleData.Elements()
	for _, ruleValue := range ruleDataMapElements {
		ruleObject := ruleValue.(types.Object)
		var planSourceRuleIndex IfwRulesRuleItemIndex
		diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		rulenDataTmp := IfwRulesRuleDataIndex{
			IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
			RuleName:       planSourceRuleIndex.RuleName.ValueString(),
			SectionName:    planSourceRuleIndex.SectionName.ValueString(),
			Description:    planSourceRuleIndex.Description.ValueString(),
			Enabled:        planSourceRuleIndex.Enabled,
		}
		resultRuleListFromPlan = append(resultRuleListFromPlan, rulenDataTmp)
	}

	// Create a map of section name to section index for proper ordering
	resultSectionIndexMap = make(map[string]int64)
	for _, section := range sectionListFromPlan {
		resultSectionIndexMap[section.SectionName] = section.SectionIndex
	}
	return resultRuleListFromPlan, resultSectionIndexMap
}
