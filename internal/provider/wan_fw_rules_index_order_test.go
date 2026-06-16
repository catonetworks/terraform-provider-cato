package provider

import (
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
)

func TestBuildWanPolicySectionOrderWithoutAnchor(t *testing.T) {
	t.Parallel()

	apiSections := []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections{
		{Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections_Section{ID: "1", Name: "managed"}},
		{Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections_Section{ID: "2", Name: "other"}},
	}
	planned := []WanRulesSectionDataIndex{
		{SectionName: "managed", SectionIndex: 1},
	}

	got := buildWanPolicySectionOrder(apiSections, planned, "")
	want := []string{"managed", "other"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestBuildWanPolicySectionOrderWithAnchor(t *testing.T) {
	t.Parallel()

	apiSections := []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections{
		{Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections_Section{ID: "1", Name: "before"}},
		{Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections_Section{ID: "2", Name: "anchor"}},
		{Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections_Section{ID: "3", Name: "after"}},
	}
	planned := []WanRulesSectionDataIndex{
		{SectionName: "managed", SectionIndex: 1},
	}

	got := buildWanPolicySectionOrder(apiSections, planned, "2")
	want := []string{"before", "anchor", "managed", "after"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}
