package rpfindex

import (
	"context"
	"fmt"

	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

// decodeRpfOrderingInputs decodes the section_data / rule_data plan maps into the
// shared flat-move planner inputs.
func decodeRpfOrderingInputs(ctx context.Context, plan RpfRulesIndex) (map[string]int64, map[string]sharedpolicy.FlatRuleRow, error) {
	sectionIndexes := map[string]int64{}
	if !plan.SectionData.IsNull() && !plan.SectionData.IsUnknown() {
		entries := map[string]RpfRulesIndexSectionEntry{}
		if diags := plan.SectionData.ElementsAs(ctx, &entries, false); diags.HasError() {
			return nil, nil, fmt.Errorf("decoding section_data: %v", diags.Errors())
		}
		for name, e := range entries {
			sectionIndexes[name] = e.SectionIndex.ValueInt64()
		}
	}
	ruleRows := map[string]sharedpolicy.FlatRuleRow{}
	if !plan.RuleData.IsNull() && !plan.RuleData.IsUnknown() {
		entries := map[string]RpfRulesIndexRuleEntry{}
		if diags := plan.RuleData.ElementsAs(ctx, &entries, false); diags.HasError() {
			return nil, nil, fmt.Errorf("decoding rule_data: %v", diags.Errors())
		}
		for name, e := range entries {
			ruleRows[name] = sharedpolicy.FlatRuleRow{
				IndexInSection: e.IndexInSection.ValueInt64(),
				SectionName:    e.SectionName.ValueString(),
			}
		}
	}
	return sectionIndexes, ruleRows, nil
}
