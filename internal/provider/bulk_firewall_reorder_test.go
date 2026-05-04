package provider

import (
	"strings"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestBuildBulkFirewallReorderInputBuildsFullPolicyOrder(t *testing.T) {
	currentSections := []bulkPolicySection{
		{ID: "s-system", Name: "System", Properties: []cato_models.PolicyElementPropertiesEnum{cato_models.PolicyElementPropertiesEnumSystem}},
		{ID: "s-keep", Name: "Keep"},
		{ID: "s-alpha", Name: "Alpha"},
		{ID: "s-beta", Name: "Beta"},
		{ID: "s-tail", Name: "Tail"},
	}
	currentRules := []bulkPolicyRule{
		{ID: "r-system", Name: "System Rule", SectionID: "s-system", SectionName: "System", Index: 1, Properties: []cato_models.PolicyElementPropertiesEnum{cato_models.PolicyElementPropertiesEnumSystem}},
		{ID: "r-keep", Name: "Keep Rule", SectionID: "s-keep", SectionName: "Keep", Index: 1},
		{ID: "r-alpha-1", Name: "Alpha One", SectionID: "s-alpha", SectionName: "Alpha", Index: 1},
		{ID: "r-alpha-2", Name: "Alpha Two", SectionID: "s-alpha", SectionName: "Alpha", Index: 2},
		{ID: "r-beta-1", Name: "Beta One", SectionID: "s-beta", SectionName: "Beta", Index: 1},
	}
	planSections := []bulkPlanSection{
		{Index: 2, Name: "Beta"},
		{Index: 1, Name: "Alpha"},
	}
	planRules := []bulkPlanRule{
		{Index: 2, Name: "Beta One", SectionName: "Alpha"},
		{Index: 1, Name: "Alpha Two", SectionName: "Alpha"},
	}

	got, sectionIDs, ruleIDs, err := buildBulkFirewallReorderInput(currentSections, currentRules, planSections, planRules, "s-keep")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertStringSliceEqual(t, "sections", reorderSectionIDs(got), []string{"s-system", "s-keep", "s-alpha", "s-beta", "s-tail"})
	assertStringSliceEqual(t, "system rules", reorderRuleIDs(got, "s-system"), []string{"r-system"})
	assertStringSliceEqual(t, "keep rules", reorderRuleIDs(got, "s-keep"), []string{"r-keep"})
	assertStringSliceEqual(t, "alpha rules", reorderRuleIDs(got, "s-alpha"), []string{"r-alpha-2", "r-beta-1", "r-alpha-1"})
	assertStringSliceEqual(t, "beta rules", reorderRuleIDs(got, "s-beta"), []string{})

	if sectionIDs["Alpha"] != "s-alpha" {
		t.Fatalf("expected section name map to contain Alpha -> s-alpha, got %q", sectionIDs["Alpha"])
	}
	if ruleIDs["Beta One"] != "r-beta-1" {
		t.Fatalf("expected rule name map to contain Beta One -> r-beta-1, got %q", ruleIDs["Beta One"])
	}
}

func TestBuildBulkFirewallReorderInputKeepsSystemRulesAtCurrentIndexes(t *testing.T) {
	currentSections := []bulkPolicySection{
		{ID: "s-main", Name: "Main"},
	}
	currentRules := []bulkPolicyRule{
		{ID: "r-one", Name: "One", SectionID: "s-main", SectionName: "Main", Index: 1},
		{ID: "r-system", Name: "System Rule", SectionID: "s-main", SectionName: "Main", Index: 2, Properties: []cato_models.PolicyElementPropertiesEnum{cato_models.PolicyElementPropertiesEnumSystem}},
		{ID: "r-two", Name: "Two", SectionID: "s-main", SectionName: "Main", Index: 3},
	}
	planSections := []bulkPlanSection{
		{Index: 1, Name: "Main"},
	}
	planRules := []bulkPlanRule{
		{Index: 1, Name: "Two", SectionName: "Main"},
		{Index: 2, Name: "One", SectionName: "Main"},
	}

	got, _, _, err := buildBulkFirewallReorderInput(currentSections, currentRules, planSections, planRules, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertStringSliceEqual(t, "main rules", reorderRuleIDs(got, "s-main"), []string{"r-two", "r-system", "r-one"})
}

func TestBuildBulkFirewallReorderInputReturnsReferenceErrors(t *testing.T) {
	currentSections := []bulkPolicySection{
		{ID: "s-one", Name: "One"},
		{ID: "s-two", Name: "Two"},
	}
	currentRules := []bulkPolicyRule{
		{ID: "r-one", Name: "Rule One", SectionID: "s-one", SectionName: "One", Index: 1},
	}

	tests := map[string]struct {
		planSections          []bulkPlanSection
		planRules             []bulkPlanRule
		sectionToStartAfterID string
		wantErr               string
	}{
		"missing start after section": {
			planSections:          []bulkPlanSection{{Index: 1, Name: "One"}},
			sectionToStartAfterID: "s-missing",
			wantErr:               "not found",
		},
		"start after section is also being reordered": {
			planSections:          []bulkPlanSection{{Index: 1, Name: "One"}},
			sectionToStartAfterID: "s-one",
			wantErr:               "cannot reference a section being reordered",
		},
		"missing planned section": {
			planSections: []bulkPlanSection{{Index: 1, Name: "Missing"}},
			wantErr:      "section \"Missing\" not found",
		},
		"rule references missing section": {
			planSections: []bulkPlanSection{{Index: 1, Name: "One"}},
			planRules:    []bulkPlanRule{{Index: 1, Name: "Rule One", SectionName: "Missing"}},
			wantErr:      "section \"Missing\" not found for rule \"Rule One\"",
		},
		"missing planned rule": {
			planSections: []bulkPlanSection{{Index: 1, Name: "One"}},
			planRules:    []bulkPlanRule{{Index: 1, Name: "Missing Rule", SectionName: "One"}},
			wantErr:      "rule \"Missing Rule\" not found",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, _, _, err := buildBulkFirewallReorderInput(currentSections, currentRules, tt.planSections, tt.planRules, tt.sectionToStartAfterID)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error to contain %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func reorderSectionIDs(input cato_models.PolicyReorderInput) []string {
	ids := make([]string, 0, len(input.Sections))
	for _, section := range input.Sections {
		ids = append(ids, section.Ref.Input)
	}
	return ids
}

func reorderRuleIDs(input cato_models.PolicyReorderInput, sectionID string) []string {
	for _, section := range input.Sections {
		if section.Ref.Input != sectionID {
			continue
		}
		ids := make([]string, 0, len(section.Rules))
		for _, rule := range section.Rules {
			ids = append(ids, rule.Ref.Input)
		}
		return ids
	}
	return nil
}

func assertStringSliceEqual(t *testing.T, name string, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: expected %v, got %v", name, want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: expected %v, got %v", name, want, got)
		}
	}
}
