package alwaysonindex

import (
	"cmp"
	"context"
	"slices"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func parseRulesIndex(ctx context.Context, result *cato.PolicyZtnaAlwaysOnPolicy, diags *diag.Diagnostics) *RulesIndex {
	if result == nil {
		diags.AddError("Invalid Always On policy response", "API did not return a policy data")
		return nil
	}
	policy := result.GetPolicy().GetZtnaAlwaysOn().GetPolicy()
	if policy.GetSections() == nil || policy.GetRules() == nil {
		diags.AddError("Invalid Always On policy response", "API did not return policy sections and rules")
		return nil
	}
	changes := policy.GetRevision().GetChanges()
	if changes < 0 {
		diags.AddError("Invalid Always On policy response", "API returned a draft change count outside the supported int64 range")
		return nil
	}

	sections, sectionNames := parseSections(ctx, policy, diags)
	if diags.HasError() {
		return nil
	}

	rules := parseRules(ctx, policy, sections, sectionNames, diags)
	if diags.HasError() {
		return nil
	}

	sectionData, sectionDiags := types.MapValue(SectionObjectType, sections)
	diags.Append(sectionDiags...)
	ruleData, ruleDiags := types.MapValue(RuleObjectType, rules)
	diags.Append(ruleDiags...)
	if diags.HasError() {
		return nil
	}
	return &RulesIndex{SectionData: sectionData, RuleData: ruleData, Changes: types.Int64Value(changes)}
}
func parseSections(ctx context.Context, policy *cato.PolicyZtnaAlwaysOnPolicy_Policy_ZtnaAlwaysOn_Policy,
	diags *diag.Diagnostics) (sections map[string]attr.Value, sectionNames map[string]string) {
	sections = make(map[string]attr.Value, len(policy.GetSections()))
	sectionNames = make(map[string]string, len(policy.GetSections()))

	for index, row := range policy.GetSections() {
		section := row.GetSection()
		id, name := section.GetID(), section.GetName()
		if id == "" || name == "" || sections[name] != nil || sectionNames[id] != "" {
			diags.AddError("Invalid Always On policy response", "API returned a section with a missing or duplicate ID/name")
			return nil, nil
		}
		sectionData := Section{
			ID:    types.StringValue(id),
			Index: types.Int64Value(int64(index) + 1),
		}
		value, valueDiags := types.ObjectValueFrom(ctx, SectionAttrTypes, sectionData)
		diags.Append(valueDiags...)
		sections[name] = value
		sectionNames[id] = name
	}
	return sections, sectionNames
}
func parseRules(ctx context.Context, policy *cato.PolicyZtnaAlwaysOnPolicy_Policy_ZtnaAlwaysOn_Policy,
	sections map[string]attr.Value, sectionNames map[string]string, diags *diag.Diagnostics,
) (rules map[string]attr.Value) {
	rules = make(map[string]attr.Value, len(policy.GetRules()))
	ruleIDs := make(map[string]bool, len(policy.GetRules()))
	sectionPositions := make(map[string]int64, len(sections))
	for _, row := range policy.GetRules() {
		rule := row.GetRule()
		id, name := rule.GetID(), rule.GetName()
		parentID := rule.GetSection().GetID()
		parentName := sectionNames[parentID]
		if id == "" || name == "" || rules[name] != nil || ruleIDs[id] || parentName == "" {
			diags.AddError("Invalid Always On policy response", "API returned a rule with a missing or duplicate ID/name or missing parent section")
			return nil
		}
		sectionPositions[parentID]++
		ruleData := Rule{
			ID:             types.StringValue(id),
			SectionName:    types.StringValue(parentName),
			IndexInSection: types.Int64Value(sectionPositions[parentID]),
		}
		value, valueDiags := types.ObjectValueFrom(ctx, RuleAttrTypes, ruleData)
		diags.Append(valueDiags...)
		rules[name] = value
		ruleIDs[id] = true
	}
	return rules
}

// buildMoveSectionInput prepares SDK input values.
func buildMoveSectionInput(currentSections []nameID, newPosition int,
) (resultInput cato_models.PolicyMoveSectionInput) {
	// Prepare input for API to move the section
	resultInput = cato_models.PolicyMoveSectionInput{
		ID: currentSections[newPosition].id,
		To: &cato_models.PolicySectionPositionInput{},
	}
	if newPosition == len(currentSections)-1 { // move to LAST_IN_POLICY
		resultInput.To.Position = cato_models.PolicySectionPositionEnumLastInPolicy
	} else {
		resultInput.To.Position = cato_models.PolicySectionPositionEnumBeforeSection
		resultInput.To.Ref = new(currentSections[newPosition+1].id)
	}
	return resultInput
}

// buildMoveRuleInput prepares SDK input values.
func buildMoveRuleInput(currentRules []nameID, newPosition int, sectionID string,
) (resultInput cato_models.PolicyMoveRuleInput) {
	// Prepare input for API to move the section
	resultInput = cato_models.PolicyMoveRuleInput{
		ID: currentRules[newPosition].id,
		To: &cato_models.PolicyRulePositionInput{},
	}
	if newPosition == len(currentRules)-1 { // move to LAST_IN_POLICY
		resultInput.To.Position = new(cato_models.PolicyRulePositionEnumLastInSection)
		resultInput.To.Ref = new(sectionID)
	} else {
		resultInput.To.Position = new(cato_models.PolicyRulePositionEnumBeforeRule)
		resultInput.To.Ref = new(currentRules[newPosition+1].id)
	}
	return resultInput
}

// makeSectionIDMap makes map[sectionName]sectionID
func makeSectionIDMap(sections map[string]Section) map[string]string {
	out := make(map[string]string)
	for secName, data := range sections {
		out[secName] = data.ID.ValueString()
	}
	return out
}

// makeRuleIDMap makes map[ruleName]ruleID
func makeRuleIDMap(rules map[string]Rule) map[string]string {
	out := make(map[string]string)
	for ruleName, data := range rules {
		out[ruleName] = data.ID.ValueString()
	}
	return out
}

// makeSectionIndexMap makes map[sectionName]sectionIndexInPolicy
func makeSectionIndexMap(sections map[string]Section) map[string]int64 {
	out := make(map[string]int64)
	for secName, data := range sections {
		out[secName] = data.Index.ValueInt64()
	}
	return out
}

// makeRuleIndexMap makes map[sectionName]map[ruleName]ruleIndexInSection
func makeRuleIndexMap(rules map[string]Rule) map[string]map[string]int64 {
	sections := make(map[string]map[string]int64)
	for ruleName, ruleData := range rules {
		sectionName := ruleData.SectionName.ValueString()
		indexInSection := ruleData.IndexInSection.ValueInt64()

		ruleMap, ok := sections[sectionName]
		if !ok {
			ruleMap = make(map[string]int64)
		}
		ruleMap[ruleName] = indexInSection

		sections[sectionName] = ruleMap
	}
	return sections
}

// buildSectionOrder extracts current map of sections from config and plan
// and returns ordered current and target slices in sectionOrder struct
func buildSectionOrder(ctx context.Context, state, config RulesIndex, diags *diag.Diagnostics,
) (order sectionOrder, sectionIDs map[string]string) {
	var currentSections, targetSections map[string]Section

	if !state.SectionData.IsNull() {
		diags.Append(state.SectionData.ElementsAs(ctx, &currentSections, false)...)
	}
	if !config.SectionData.IsNull() {
		diags.Append(config.SectionData.ElementsAs(ctx, &targetSections, false)...)
	}
	if diags.HasError() {
		return sectionOrder{}, nil
	}
	if len(currentSections) == 0 && len(targetSections) == 0 {
		return sectionOrder{}, nil
	}

	sectionIDMap := makeSectionIDMap(currentSections)
	currentSectionIndexMap := makeSectionIndexMap(currentSections)
	targetSectionIndexMap := makeSectionIndexMap(targetSections)
	validateSections(sectionIDMap, currentSectionIndexMap, targetSectionIndexMap, diags)
	if diags.HasError() {
		return sectionOrder{}, nil
	}

	orderedSections := sectionOrder{
		current: makeSectionSlice(sectionIDMap, currentSectionIndexMap),
		target:  makeSectionSlice(sectionIDMap, targetSectionIndexMap),
	}

	return orderedSections, sectionIDMap
}

func makeSectionSlice(sectionIDs map[string]string, sectionIndexes map[string]int64) []nameID {
	sections := make([]nameID, 0, len(sectionIDs))
	for name, id := range sectionIDs {
		sections = append(sections, nameID{name: name, id: id})
	}
	slices.SortFunc(sections, func(a, b nameID) int {
		return cmp.Compare(sectionIndexes[a.name], sectionIndexes[b.name])
	})
	return sections
}

// buildRuleOrder extracts current list of rules in sections from config and plan
// and returns a map where for each section there are ordered current and target slices of rules
func buildRuleOrder(ctx context.Context, state, config RulesIndex, sectionIDs map[string]string, diags *diag.Diagnostics,
) map[string]ruleOrder {
	var currentRules, targetRules map[string]Rule

	if !state.RuleData.IsNull() {
		diags.Append(state.RuleData.ElementsAs(ctx, &currentRules, false)...)
	}
	if !config.RuleData.IsNull() {
		diags.Append(config.RuleData.ElementsAs(ctx, &targetRules, false)...)
	}
	if diags.HasError() {
		return nil
	}
	if len(currentRules) == 0 && len(targetRules) == 0 {
		return nil
	}

	ruleIDs := makeRuleIDMap(currentRules)
	currentRuleIndexMap := makeRuleIndexMap(currentRules)
	targetRuleIndexMap := makeRuleIndexMap(targetRules)
	validateRules(sectionIDs, currentRuleIndexMap, targetRuleIndexMap, diags)
	if diags.HasError() {
		return nil
	}

	ruleOrderMap := make(map[string]ruleOrder)

	for secName, currentRuleMap := range currentRuleIndexMap {
		rules := ruleOrderMap[secName]
		rules.sectionID = sectionIDs[secName]
		for ruleName := range currentRuleMap {
			rules.current = append(rules.current, nameID{name: ruleName, id: ruleIDs[ruleName]})
		}
		// sort current by index
		slices.SortFunc(rules.current, func(a, b nameID) int {
			return cmp.Compare(currentRuleMap[a.name], currentRuleMap[b.name])
		})
		ruleOrderMap[secName] = rules
	}

	for secName, targetRuleMap := range targetRuleIndexMap {
		rules := ruleOrderMap[secName]
		rules.sectionID = sectionIDs[secName]
		for ruleName := range targetRuleMap {
			rules.target = append(rules.target, nameID{name: ruleName, id: ruleIDs[ruleName]})
		}
		// sort target by index
		slices.SortFunc(rules.target, func(a, b nameID) int {
			return cmp.Compare(targetRuleMap[a.name], targetRuleMap[b.name])
		})
		ruleOrderMap[secName] = rules
	}
	return ruleOrderMap
}
