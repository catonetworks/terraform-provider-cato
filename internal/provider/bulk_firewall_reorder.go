package provider

import (
	"fmt"
	"sort"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type bulkPolicySection struct {
	ID         string
	Name       string
	Properties []cato_models.PolicyElementPropertiesEnum
}

type bulkPolicyRule struct {
	ID          string
	Name        string
	SectionID   string
	SectionName string
	Description string
	Enabled     bool
	Index       int64
	Properties  []cato_models.PolicyElementPropertiesEnum
}

type bulkPlanSection struct {
	Index int64
	Name  string
}

type bulkPlanRule struct {
	Index       int64
	Name        string
	SectionName string
	Description string
	Enabled     bool
}

func buildBulkFirewallReorderInput(currentSections []bulkPolicySection, currentRules []bulkPolicyRule, planSections []bulkPlanSection, planRules []bulkPlanRule, sectionToStartAfterID string) (cato_models.PolicyReorderInput, map[string]string, map[string]string, error) {
	sectionIDByName := make(map[string]string, len(currentSections))
	sectionByID := make(map[string]bulkPolicySection, len(currentSections))
	for _, section := range currentSections {
		sectionIDByName[section.Name] = section.ID
		sectionByID[section.ID] = section
	}

	ruleIDByName := make(map[string]string, len(currentRules))
	ruleByName := make(map[string]bulkPolicyRule, len(currentRules))
	for _, rule := range currentRules {
		ruleIDByName[rule.Name] = rule.ID
		ruleByName[rule.Name] = rule
	}

	if sectionToStartAfterID != "" {
		if _, ok := sectionByID[sectionToStartAfterID]; !ok {
			return cato_models.PolicyReorderInput{}, nil, nil, fmt.Errorf("SectionToStartAfterId %q not found", sectionToStartAfterID)
		}
	}

	sort.Slice(planSections, func(i, j int) bool {
		return planSections[i].Index < planSections[j].Index
	})
	sort.Slice(planRules, func(i, j int) bool {
		if planRules[i].SectionName != planRules[j].SectionName {
			return planRules[i].SectionName < planRules[j].SectionName
		}
		return planRules[i].Index < planRules[j].Index
	})

	plannedSectionIDs := make(map[string]struct{}, len(planSections))
	orderedPlannedSections := make([]bulkPolicySection, 0, len(planSections))
	for _, planned := range planSections {
		sectionID, ok := sectionIDByName[planned.Name]
		if !ok {
			return cato_models.PolicyReorderInput{}, nil, nil, fmt.Errorf("section %q not found", planned.Name)
		}
		plannedSectionIDs[sectionID] = struct{}{}
		orderedPlannedSections = append(orderedPlannedSections, sectionByID[sectionID])
	}

	baseSections := make([]bulkPolicySection, 0, len(currentSections))
	for _, section := range currentSections {
		if _, planned := plannedSectionIDs[section.ID]; planned {
			continue
		}
		baseSections = append(baseSections, section)
	}

	insertAt := len(baseSections)
	if sectionToStartAfterID != "" {
		insertAt = -1
		for i, section := range baseSections {
			if section.ID == sectionToStartAfterID {
				insertAt = i + 1
				break
			}
		}
		if insertAt == -1 {
			return cato_models.PolicyReorderInput{}, nil, nil, fmt.Errorf("SectionToStartAfterId %q cannot reference a section being reordered", sectionToStartAfterID)
		}
	}

	desiredSections := make([]bulkPolicySection, 0, len(currentSections))
	desiredSections = append(desiredSections, baseSections[:insertAt]...)
	desiredSections = append(desiredSections, orderedPlannedSections...)
	desiredSections = append(desiredSections, baseSections[insertAt:]...)
	desiredSections = keepSystemSectionsAtCurrentIndexes(currentSections, desiredSections)

	plannedRuleIDs := make(map[string]struct{}, len(planRules))
	plannedRulesBySection := make(map[string][]bulkPolicyRule)
	for _, planned := range planRules {
		sectionID, ok := sectionIDByName[planned.SectionName]
		if !ok {
			return cato_models.PolicyReorderInput{}, nil, nil, fmt.Errorf("section %q not found for rule %q", planned.SectionName, planned.Name)
		}
		rule, ok := ruleByName[planned.Name]
		if !ok {
			return cato_models.PolicyReorderInput{}, nil, nil, fmt.Errorf("rule %q not found", planned.Name)
		}
		rule.SectionID = sectionID
		rule.SectionName = planned.SectionName
		plannedRuleIDs[rule.ID] = struct{}{}
		plannedRulesBySection[sectionID] = append(plannedRulesBySection[sectionID], rule)
	}

	currentRulesBySection := make(map[string][]bulkPolicyRule)
	sort.Slice(currentRules, func(i, j int) bool {
		return currentRules[i].Index < currentRules[j].Index
	})
	for _, rule := range currentRules {
		currentRulesBySection[rule.SectionID] = append(currentRulesBySection[rule.SectionID], rule)
	}

	sections := make([]*cato_models.PolicyReorderSectionInput, 0, len(desiredSections))
	for _, section := range desiredSections {
		orderedRules := reorderSectionRules(currentRulesBySection[section.ID], plannedRulesBySection[section.ID], plannedRuleIDs)
		rules := make([]*cato_models.PolicyReorderRuleInput, 0, len(orderedRules))
		for _, rule := range orderedRules {
			rules = append(rules, &cato_models.PolicyReorderRuleInput{
				Ref:      policyElementRef(rule.ID),
				SubRules: []*cato_models.PolicyReorderSubRuleInput{},
			})
		}

		sections = append(sections, &cato_models.PolicyReorderSectionInput{
			Ref:   policyElementRef(section.ID),
			Rules: rules,
		})
	}

	return cato_models.PolicyReorderInput{Sections: sections}, sectionIDByName, ruleIDByName, nil
}

func keepSystemSectionsAtCurrentIndexes(currentSections []bulkPolicySection, desiredSections []bulkPolicySection) []bulkPolicySection {
	desiredNonSystem := make([]bulkPolicySection, 0, len(desiredSections))
	for _, section := range desiredSections {
		if !hasSystemProperty(section.Properties) {
			desiredNonSystem = append(desiredNonSystem, section)
		}
	}

	result := make([]bulkPolicySection, 0, len(currentSections))
	nextNonSystem := 0
	for _, current := range currentSections {
		if hasSystemProperty(current.Properties) {
			result = append(result, current)
			continue
		}
		if nextNonSystem < len(desiredNonSystem) {
			result = append(result, desiredNonSystem[nextNonSystem])
			nextNonSystem++
		}
	}
	result = append(result, desiredNonSystem[nextNonSystem:]...)

	return result
}

func reorderSectionRules(currentRules []bulkPolicyRule, plannedRules []bulkPolicyRule, plannedRuleIDs map[string]struct{}) []bulkPolicyRule {
	desiredNonSystem := make([]bulkPolicyRule, 0, len(currentRules)+len(plannedRules))
	desiredNonSystem = append(desiredNonSystem, plannedRules...)
	for _, rule := range currentRules {
		if hasSystemProperty(rule.Properties) {
			continue
		}
		if _, planned := plannedRuleIDs[rule.ID]; planned {
			continue
		}
		desiredNonSystem = append(desiredNonSystem, rule)
	}

	result := make([]bulkPolicyRule, 0, len(currentRules)+len(plannedRules))
	nextNonSystem := 0
	for _, current := range currentRules {
		if hasSystemProperty(current.Properties) {
			result = append(result, current)
			continue
		}
		if nextNonSystem < len(desiredNonSystem) {
			result = append(result, desiredNonSystem[nextNonSystem])
			nextNonSystem++
		}
	}
	result = append(result, desiredNonSystem[nextNonSystem:]...)

	return result
}

func hasSystemProperty(properties []cato_models.PolicyElementPropertiesEnum) bool {
	for _, property := range properties {
		if property == cato_models.PolicyElementPropertiesEnumSystem {
			return true
		}
	}
	return false
}

func policyElementRef(id string) *cato_models.PolicyElementRefInput {
	return &cato_models.PolicyElementRefInput{
		By:    cato_models.ObjectRefByID,
		Input: id,
	}
}
