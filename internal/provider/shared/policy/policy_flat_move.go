package policy

import "sort"

// FlatMoveOp is one planned move operation (section or rule) for a flat
// (sections + rules, no sub-policies) policy, in execution order.
type FlatMoveOp struct {
	IsSection   bool
	Name        string
	SectionName string // for rules: the target section
}

// FlatRuleRow is one rule ordering row.
type FlatRuleRow struct {
	IndexInSection int64
	SectionName    string
}

// BuildFlatMoveSequence turns ordering inputs into a deterministic move
// sequence for policies without a reorderPolicy mutation: sections in
// section_index order (each moved LAST_IN_POLICY, so the sequence yields
// ascending order), then per section (same order) its rules in
// index_in_section order (each moved LAST_IN_SECTION). Rules whose section is
// not in sectionIndexes are grouped after, in name order, so the sequence is
// total and deterministic. Used by the Always-On, Application Control, and
// App Tenant Restriction bulk move resources.
func BuildFlatMoveSequence(sectionIndexes map[string]int64, ruleRows map[string]FlatRuleRow) []FlatMoveOp {
	ops := make([]FlatMoveOp, 0, len(sectionIndexes)+len(ruleRows))

	sectionNames := make([]string, 0, len(sectionIndexes))
	for name := range sectionIndexes {
		sectionNames = append(sectionNames, name)
	}
	sort.Slice(sectionNames, func(i, j int) bool {
		if sectionIndexes[sectionNames[i]] != sectionIndexes[sectionNames[j]] {
			return sectionIndexes[sectionNames[i]] < sectionIndexes[sectionNames[j]]
		}
		return sectionNames[i] < sectionNames[j]
	})
	for _, name := range sectionNames {
		ops = append(ops, FlatMoveOp{IsSection: true, Name: name})
	}

	bySection := map[string][]string{}
	for ruleName := range ruleRows {
		bySection[ruleRows[ruleName].SectionName] = append(bySection[ruleRows[ruleName].SectionName], ruleName)
	}
	orderedSections := append([]string{}, sectionNames...)
	var extraSections []string
	for sectionName := range bySection {
		if _, ok := sectionIndexes[sectionName]; !ok {
			extraSections = append(extraSections, sectionName)
		}
	}
	sort.Strings(extraSections)
	orderedSections = append(orderedSections, extraSections...)

	for _, sectionName := range orderedSections {
		ruleNames := bySection[sectionName]
		sort.Slice(ruleNames, func(i, j int) bool {
			ri, rj := ruleRows[ruleNames[i]], ruleRows[ruleNames[j]]
			if ri.IndexInSection != rj.IndexInSection {
				return ri.IndexInSection < rj.IndexInSection
			}
			return ruleNames[i] < ruleNames[j]
		})
		for _, ruleName := range ruleNames {
			ops = append(ops, FlatMoveOp{IsSection: false, Name: ruleName, SectionName: sectionName})
		}
	}
	return ops
}
