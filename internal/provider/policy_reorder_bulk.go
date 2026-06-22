package provider

import (
	"fmt"
	"sort"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func gqlOptionalStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BulkPolicySectionRef is one section row from Policy*FirewallSectionsIndex.
type BulkPolicySectionRef struct {
	ID   string
	Name string
}

// BulkPolicyRuleRow is one rule row from Policy*FirewallRulesIndex.
type BulkPolicyRuleRow struct {
	SectionID   string
	SectionName string
	RuleID      string
	RuleName    string
	Index       int64
}

// BulkPlannedRuleIndex is one rule placement from Terraform rule_data.
type BulkPlannedRuleIndex struct {
	SectionName    string
	RuleName       string
	IndexInSection int64
}

func sectionIDSet(sections []BulkPolicySectionRef) map[string]struct{} {
	sectionIDs := make(map[string]struct{}, len(sections))
	for _, s := range sections {
		sectionIDs[s.ID] = struct{}{}
	}
	return sectionIDs
}

func indexRulesBySectionID(sectionIDs map[string]struct{}, rules []BulkPolicyRuleRow) (map[string][]BulkPolicyRuleRow, error) {
	rulesBySectionID := make(map[string][]BulkPolicyRuleRow)
	for _, rw := range rules {
		if _, ok := sectionIDs[rw.SectionID]; !ok {
			return nil, fmt.Errorf("rule %q references unknown section id %q", rw.RuleName, rw.SectionID)
		}
		rulesBySectionID[rw.SectionID] = append(rulesBySectionID[rw.SectionID], rw)
	}
	for sid, list := range rulesBySectionID {
		sort.Slice(list, func(i, j int) bool {
			if list[i].Index != list[j].Index {
				return list[i].Index < list[j].Index
			}
			return list[i].RuleName < list[j].RuleName
		})
		rulesBySectionID[sid] = list
	}
	return rulesBySectionID, nil
}

func indexPlannedBySectionName(planned []BulkPlannedRuleIndex) map[string][]BulkPlannedRuleIndex {
	plannedBySection := make(map[string][]BulkPlannedRuleIndex)
	for _, p := range planned {
		plannedBySection[p.SectionName] = append(plannedBySection[p.SectionName], p)
	}
	for sn, plist := range plannedBySection {
		sort.Slice(plist, func(i, j int) bool {
			if plist[i].IndexInSection != plist[j].IndexInSection {
				return plist[i].IndexInSection < plist[j].IndexInSection
			}
			return plist[i].RuleName < plist[j].RuleName
		})
		plannedBySection[sn] = plist
	}
	return plannedBySection
}

func ruleNameToIDGlobal(rules []BulkPolicyRuleRow) map[string]string {
	out := make(map[string]string, len(rules))
	for _, r := range rules {
		out[r.RuleName] = r.RuleID
	}
	return out
}

func buildSectionReorderInput(
	sec BulkPolicySectionRef,
	apiRules []BulkPolicyRuleRow,
	plannedForSec []BulkPlannedRuleIndex,
	globalNameToID map[string]string,
	allPlanned []BulkPlannedRuleIndex,
) (*cato_models.PolicyReorderSectionInput, error) {
	plannedNames := make(map[string]struct{}, len(plannedForSec))
	for _, p := range plannedForSec {
		plannedNames[p.RuleName] = struct{}{}
	}

	plannedToOtherSection := make(map[string]struct{})
	for _, p := range allPlanned {
		if p.SectionName != sec.Name {
			plannedToOtherSection[p.RuleName] = struct{}{}
		}
	}

	for _, p := range plannedForSec {
		if _, ok := globalNameToID[p.RuleName]; !ok {
			return nil, fmt.Errorf(
				"planned rule %q in section %q not found in current policy", p.RuleName, sec.Name)
		}
	}

	orderedPlannedIDs := make([]string, 0, len(plannedForSec))
	for _, p := range plannedForSec {
		orderedPlannedIDs = append(orderedPlannedIDs, globalNameToID[p.RuleName])
	}

	tailIDs := make([]string, 0, len(apiRules))
	for _, r := range apiRules {
		if _, ok := plannedNames[r.RuleName]; ok {
			continue
		}
		if _, ok := plannedToOtherSection[r.RuleName]; ok {
			continue
		}
		tailIDs = append(tailIDs, r.RuleID)
	}

	finalIDs := append(append([]string{}, orderedPlannedIDs...), tailIDs...)

	ruleInputs := make([]*cato_models.PolicyReorderRuleInput, 0, len(finalIDs))
	for _, rid := range finalIDs {
		idCopy := rid
		ruleInputs = append(ruleInputs, &cato_models.PolicyReorderRuleInput{
			Ref: &cato_models.PolicyElementRefInput{
				By:    cato_models.ObjectRefByID,
				Input: idCopy,
			},
			SubRules: []*cato_models.PolicyReorderSubRuleInput{},
		})
	}

	secID := sec.ID
	return &cato_models.PolicyReorderSectionInput{
		Ref: &cato_models.PolicyElementRefInput{
			By:    cato_models.ObjectRefByID,
			Input: secID,
		},
		Rules: ruleInputs,
	}, nil
}

// buildPolicyReorderInput builds a full PolicyReorderInput from the current API
// snapshot plus Terraform planned indices. Planned rules are stacked first in
// index_in_section order; remaining rules in the section keep their relative API order.
// Every section from the index must appear in the output (WAN and IF return
// reorderPolicyMissingSections if any section is omitted), including sections whose
// final rule list is empty after moves.
func buildPolicyReorderInput(
	sections []BulkPolicySectionRef,
	rules []BulkPolicyRuleRow,
	planned []BulkPlannedRuleIndex,
) (cato_models.PolicyReorderInput, error) {
	sectionIDs := sectionIDSet(sections)
	rulesBySectionID, err := indexRulesBySectionID(sectionIDs, rules)
	if err != nil {
		return cato_models.PolicyReorderInput{}, err
	}
	plannedBySection := indexPlannedBySectionName(planned)
	globalNameToID := ruleNameToIDGlobal(rules)

	outSections := make([]*cato_models.PolicyReorderSectionInput, 0, len(sections))
	for _, sec := range sections {
		secInput, err := buildSectionReorderInput(
			sec,
			rulesBySectionID[sec.ID],
			plannedBySection[sec.Name],
			globalNameToID,
			planned,
		)
		if err != nil {
			return cato_models.PolicyReorderInput{}, err
		}
		outSections = append(outSections, secInput)
	}

	return cato_models.PolicyReorderInput{Sections: outSections}, nil
}

func internetFirewallReorderError(resp *cato_go_sdk.PolicyInternetFirewallReorderPolicy, callErr error) error {
	if callErr != nil {
		return callErr
	}
	if resp == nil || resp.GetPolicy() == nil || resp.GetPolicy().GetInternetFirewall() == nil {
		return fmt.Errorf("internet firewall reorderPolicy returned empty policy payload")
	}
	rp := resp.GetPolicy().GetInternetFirewall().GetReorderPolicy()
	if rp == nil {
		return fmt.Errorf("internet firewall reorderPolicy returned empty reorder payload")
	}
	if errs := rp.GetErrors(); len(errs) > 0 {
		var b strings.Builder
		for _, e := range errs {
			if e == nil {
				continue
			}
			if b.Len() > 0 {
				b.WriteString("; ")
			}
			_, _ = fmt.Fprintf(&b, "%s: %s", gqlOptionalStr(e.GetErrorCode()), gqlOptionalStr(e.GetErrorMessage()))
		}
		return fmt.Errorf("internet firewall reorderPolicy errors: %s", b.String())
	}
	if st := rp.GetStatus(); st != nil && *st != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("internet firewall reorderPolicy status %q", string(*st))
	}
	return nil
}

func wanFirewallReorderError(resp *cato_go_sdk.PolicyWanFirewallReorderPolicy, callErr error) error {
	if callErr != nil {
		return callErr
	}
	if resp == nil || resp.GetPolicy() == nil || resp.GetPolicy().GetWanFirewall() == nil {
		return fmt.Errorf("WAN firewall reorderPolicy returned empty policy payload")
	}
	rp := resp.GetPolicy().GetWanFirewall().GetReorderPolicy()
	if rp == nil {
		return fmt.Errorf("WAN firewall reorderPolicy returned empty reorder payload")
	}
	if errs := rp.GetErrors(); len(errs) > 0 {
		var b strings.Builder
		for _, e := range errs {
			if e == nil {
				continue
			}
			if b.Len() > 0 {
				b.WriteString("; ")
			}
			_, _ = fmt.Fprintf(&b, "%s: %s", gqlOptionalStr(e.GetErrorCode()), gqlOptionalStr(e.GetErrorMessage()))
		}
		return fmt.Errorf("WAN firewall reorderPolicy errors: %s", b.String())
	}
	if st := rp.GetStatus(); st != nil && *st != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("WAN firewall reorderPolicy status %q", string(*st))
	}
	return nil
}
