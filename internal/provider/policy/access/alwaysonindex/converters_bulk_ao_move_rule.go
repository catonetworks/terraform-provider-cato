package alwaysonindex

import (
	"context"
	"fmt"
)

// decodeAoOrderingInputs decodes the section_data / rule_data plan maps into
// the move-sequence planner inputs.
func decodeAoOrderingInputs(
	ctx context.Context, plan AoRulesIndex,
) (sectionIndexes map[string]int64, ruleRows map[string]AoRuleRow, err error) {
	sectionIndexes = map[string]int64{}
	if !plan.SectionData.IsNull() && !plan.SectionData.IsUnknown() {
		entries := map[string]AoSectionDataEntry{}
		if diags := plan.SectionData.ElementsAs(ctx, &entries, false); diags.HasError() {
			return nil, nil, fmt.Errorf("decoding section_data: %v", diags.Errors())
		}
		for name, e := range entries {
			sectionIndexes[name] = e.SectionIndex.ValueInt64()
		}
	}
	ruleRows = map[string]AoRuleRow{}
	if !plan.RuleData.IsNull() && !plan.RuleData.IsUnknown() {
		entries := map[string]AoRuleDataEntry{}
		if diags := plan.RuleData.ElementsAs(ctx, &entries, false); diags.HasError() {
			return nil, nil, fmt.Errorf("decoding rule_data: %v", diags.Errors())
		}
		for name, e := range entries {
			ruleRows[name] = AoRuleRow{
				IndexInSection: e.IndexInSection.ValueInt64(),
				SectionName:    e.SectionName.ValueString(),
			}
		}
	}
	return sectionIndexes, ruleRows, nil
}
