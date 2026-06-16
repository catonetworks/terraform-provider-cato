package provider

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
)

func buildWanPolicySectionOrder(
	apiSections []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections,
	sectionListFromPlan []WanRulesSectionDataIndex,
	sectionToStartAfterID string,
) []string {
	plannedSet := make(map[string]struct{}, len(sectionListFromPlan))
	plannedNames := make([]string, 0, len(sectionListFromPlan))
	for _, section := range sectionListFromPlan {
		plannedNames = append(plannedNames, section.SectionName)
		plannedSet[section.SectionName] = struct{}{}
	}

	if sectionToStartAfterID == "" {
		result := append([]string{}, plannedNames...)
		for _, apiSection := range apiSections {
			sectionName := apiSection.Section.Name
			if _, planned := plannedSet[sectionName]; planned {
				continue
			}
			result = append(result, sectionName)
		}
		return result
	}

	anchorIdx := -1
	for i, apiSection := range apiSections {
		if apiSection.Section.ID == sectionToStartAfterID {
			anchorIdx = i
			break
		}
	}
	if anchorIdx < 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(apiSections)+len(plannedNames))
	result := make([]string, 0, len(apiSections)+len(plannedNames))
	add := func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}

	for i := 0; i <= anchorIdx; i++ {
		sectionName := apiSections[i].Section.Name
		if _, planned := plannedSet[sectionName]; planned {
			continue
		}
		add(sectionName)
	}
	for _, sectionName := range plannedNames {
		add(sectionName)
	}
	for i := anchorIdx + 1; i < len(apiSections); i++ {
		add(apiSections[i].Section.Name)
	}

	return result
}
