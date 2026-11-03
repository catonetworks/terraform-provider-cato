package tsindex

import (
	"context"
	"fmt"

	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

// decodeTsOrderingInputs decodes the section_data / rule_data plan maps into the
// shared flat-move planner inputs.
func decodeTsOrderingInputs(
	ctx context.Context, plan TsRulesIndex,
) (sectionIndexes map[string]int64, ruleRows map[string]sharedpolicy.FlatRuleRow, err error) {
	sectionIndexes = map[string]int64{}
	if !plan.SectionData.IsNull() && !plan.SectionData.IsUnknown() {
		entries := map[string]TsRulesIndexSectionEntry{}
		if diags := plan.SectionData.ElementsAs(ctx, &entries, false); diags.HasError() {
			return nil, nil, fmt.Errorf("decoding section_data: %v", diags.Errors())
		}
		for name, e := range entries {
			sectionIndexes[name] = e.SectionIndex.ValueInt64()
		}
	}
	ruleRows = map[string]sharedpolicy.FlatRuleRow{}
	if !plan.RuleData.IsNull() && !plan.RuleData.IsUnknown() {
		entries := map[string]TsRulesIndexRuleEntry{}
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
