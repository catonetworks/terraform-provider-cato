package alwaysonindex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type sectionDataValidator struct{}

// ValidateMap checks section names and requires unique indexes from 1 to the section count.
func (v sectionDataValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var sections map[string]types.Object
	decodeDiags := req.ConfigValue.ElementsAs(ctx, &sections, false)
	resp.Diagnostics.Append(decodeDiags...)
	if decodeDiags.HasError() {
		return
	}
	maxIndex := int64(len(sections))
	seen := make(map[int64]string)
	for name, object := range sections {
		if name == "" {
			resp.Diagnostics.AddAttributeError(req.Path.AtMapKey(name), "Invalid section name", "Section name must not be empty.")
			continue
		}
		if object.IsNull() || object.IsUnknown() {
			continue
		}
		var section Section
		decodeDiags := object.As(ctx, &section, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(decodeDiags...)
		if decodeDiags.HasError() {
			return
		}
		if section.Index.IsNull() || section.Index.IsUnknown() {
			continue
		}
		index := section.Index.ValueInt64()
		indexPath := req.Path.AtMapKey(name).AtName("section_index")
		if index < 1 || index > maxIndex {
			resp.Diagnostics.AddAttributeError(indexPath, "Invalid section index",
				fmt.Sprintf("Section %q must have an index between 1 and %d; got %d.", name, maxIndex, index))
			continue
		}
		if previous, exists := seen[index]; exists {
			resp.Diagnostics.AddAttributeError(indexPath, "Duplicate section index",
				fmt.Sprintf("Sections %q and %q both use index %d.", previous, name, index))
			continue
		}
		seen[index] = name
	}
}

func (v sectionDataValidator) Description(_ context.Context) string {
	return "Section names must not be empty, and section indexes must form a sequence from 1 to the number of sections."
}

func (v sectionDataValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

type ruleDataValidator struct{}

// ValidateMap validates that
// - rule name is not empty
// - section_name is not empty
// - for each section_name, the index_in_section must be a sequence of numbers from 1 to number of rules in the section
func (v ruleDataValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var objects map[string]types.Object
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &objects, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rules := make(map[string]Rule)
	sectionCounts := make(map[string]int64)
	unresolvedSections := false
	for name, object := range objects {
		if name == "" {
			resp.Diagnostics.AddAttributeError(req.Path.AtMapKey(name), "Invalid rule name", "Rule name must not be empty.")
			continue
		}
		if object.IsNull() || object.IsUnknown() {
			unresolvedSections = true
			continue
		}
		var rule Rule
		decodeDiags := object.As(ctx, &rule, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(decodeDiags...)
		if decodeDiags.HasError() {
			return
		}
		if rule.SectionName.IsNull() || rule.SectionName.IsUnknown() {
			unresolvedSections = true
			continue
		}
		section := rule.SectionName.ValueString()
		if section == "" {
			resp.Diagnostics.AddAttributeError(req.Path.AtMapKey(name).AtName("section_name"),
				"Invalid section name", "Section name must not be empty.")
			continue
		}
		sectionCounts[section]++
		rules[name] = rule
	}
	validateRuleDataIndexes(rules, sectionCounts, unresolvedSections, req.Path, &resp.Diagnostics)
}

func validateRuleDataIndexes(rules map[string]Rule, sectionCounts map[string]int64,
	unresolvedSections bool, attributePath path.Path, diags *diag.Diagnostics,
) {
	seen := make(map[string]map[int64]string)
	for name, rule := range rules {
		if rule.IndexInSection.IsNull() || rule.IndexInSection.IsUnknown() {
			continue
		}
		section, index := rule.SectionName.ValueString(), rule.IndexInSection.ValueInt64()
		indexPath := attributePath.AtMapKey(name).AtName("index_in_section")
		if index < 1 || (!unresolvedSections && index > sectionCounts[section]) {
			diags.AddAttributeError(indexPath, "Invalid rule index",
				fmt.Sprintf("Rule %q in section %q must have an index from 1 to the number of rules in the section (%d); got %d.",
					name, section, sectionCounts[section], index))
			continue
		}
		if seen[section] == nil {
			seen[section] = make(map[int64]string)
		}
		if previous, exists := seen[section][index]; exists {
			diags.AddAttributeError(indexPath, "Duplicate rule index",
				fmt.Sprintf("Rules %q and %q in section %q both use index %d.", previous, name, section, index))
			continue
		}
		seen[section][index] = name
	}
}
func (v ruleDataValidator) Description(_ context.Context) string {
	return "Rule and section names must not be empty, and indexes in each section must form a sequence from 1 to its rule count."
}
func (v ruleDataValidator) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func validateSections(sectionIDs map[string]string, current, target map[string]int64, diags *diag.Diagnostics) {
	if len(current) != len(target) {
		if len(current) > len(target) {
			diags.AddError("Unexpected always-on policy sections",
				fmt.Sprintf("Expected: %d sections, got: %d", len(target), len(current)))
			return
		}
		diags.AddError("Number of section does not match", fmt.Sprintf("Number of sections current: %d, target: %d",
			len(current), len(target)))
		return
	}

	checkIdx := func(sectionIDs map[string]string, sectionIndexes map[string]int64) bool {
		maxIdx := int64(len(sectionIndexes))
		idxSeen := make(map[int64]string)

		for name, index := range sectionIndexes {
			_, found := sectionIDs[name]
			if !found {
				diags.AddError("Section ID not found", fmt.Sprintf("ID of section '%s' not found in TF state", name))
				return false
			}
			if index < 1 || index > maxIdx {
				diags.AddError("Invalid section index",
					fmt.Sprintf("section index must be between 1 and %d, got %d", maxIdx, index))
				return false
			}
			if nameSeen, exist := idxSeen[index]; exist {
				diags.AddError("Duplicate section index", fmt.Sprintf("section index '%d' was used multiple times ('%s', '%s')",
					index, nameSeen, name))
				return false
			}
			idxSeen[index] = name
		}
		return true
	}

	if !checkIdx(sectionIDs, current) {
		return
	}
	if !checkIdx(sectionIDs, target) {
		return
	}
}

// validateRules validates rule maps
// sectionIDs, ruleIDs = map[name]id
// current, target = map[sectionName]map[ruleName]indexInSection
func validateRules(sectionIDs map[string]string, current, target map[string]map[string]int64, diags *diag.Diagnostics) {
	if len(current) != len(target) {
		diags.AddError("Number of sections does not match", fmt.Sprintf("Current sections: %d target: %d", len(current), len(target)))
		return
	}

	// check rules are matching
	for secName, currentRuleMap := range current {
		targetRuleMap, ok := target[secName]
		if !ok {
			diags.AddError("Section not found", fmt.Sprintf("section '%s' not found in the target", secName))
			return
		}
		if len(currentRuleMap) != len(targetRuleMap) {
			if len(currentRuleMap) > len(targetRuleMap) {
				diags.AddError("Unexpected always-on rules",
					fmt.Sprintf("Unexpected always-on rules found in section '%s' expected: %d rules, got: %d",
						secName, len(targetRuleMap), len(currentRuleMap)))
				return
			}
			diags.AddError("Number of rules does not match", fmt.Sprintf("Rules section '%s' current: %d, target: %d",
				secName, len(currentRuleMap), len(targetRuleMap)))
			return
		}
		for curName := range currentRuleMap {
			if _, ok := targetRuleMap[curName]; !ok {
				diags.AddError("Rule not found", fmt.Sprintf("Rule '%s' in section '%s' not found in target",
					curName, secName))
				return
			}
		}
	}

	for secName, currentRuleMap := range current {
		if _, ok := sectionIDs[secName]; !ok {
			diags.AddError("Section ID not found", fmt.Sprintf("ID of section '%s' not found in the state", secName))
			return
		}
		if !checkRuleIndexes(currentRuleMap, diags) {
			return
		}
	}
	for secName, targetRuleMap := range target {
		if _, ok := sectionIDs[secName]; !ok {
			diags.AddError("Section ID not found", fmt.Sprintf("ID of section '%s' not found in the plan", secName))
			return
		}
		if !checkRuleIndexes(targetRuleMap, diags) {
			return
		}
	}
}

func checkRuleIndexes(ruleMap map[string]int64, diags *diag.Diagnostics) bool {
	seen := make(map[int64]struct{})
	for name, idx := range ruleMap {
		if idx < 1 || idx > int64(len(ruleMap)) {
			diags.AddError("Invalid rule index",
				fmt.Sprintf("Rule index must be between 1 and %d, got %d for rule %s", len(ruleMap), idx, name))
			return false
		}
		if _, ok := seen[idx]; ok {
			diags.AddError("Duplicate rule index",
				fmt.Sprintf("Duplicate rule index %d, for rule %s", idx, name))
			return false
		}
		seen[idx] = struct{}{}
	}
	return true
}
