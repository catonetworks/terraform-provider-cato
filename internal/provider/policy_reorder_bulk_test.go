package provider

import (
	"context"
	"errors"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

func TestBuildPolicyReorderInput_twoRulesSwapOrder(t *testing.T) {
	t.Parallel()
	sections := []BulkPolicySectionRef{{ID: "s1", Name: "Sec"}}
	rules := []BulkPolicyRuleRow{
		{SectionID: "s1", SectionName: "Sec", RuleID: "r1", RuleName: "A", Index: 1},
		{SectionID: "s1", SectionName: "Sec", RuleID: "r2", RuleName: "B", Index: 2},
	}
	planned := []BulkPlannedRuleIndex{
		{SectionName: "Sec", RuleName: "B", IndexInSection: 1},
		{SectionName: "Sec", RuleName: "A", IndexInSection: 2},
	}
	in, err := buildPolicyReorderInput(sections, rules, planned)
	require.NoError(t, err)
	require.Len(t, in.Sections, 1)
	require.Len(t, in.Sections[0].Rules, 2)
	require.Equal(t, "r2", in.Sections[0].Rules[0].Ref.Input)
	require.Equal(t, "r1", in.Sections[0].Rules[1].Ref.Input)
}

func TestBuildPolicyReorderInput_skipsSectionsWithNoRules(t *testing.T) {
	t.Parallel()
	sections := []BulkPolicySectionRef{
		{ID: "empty", Name: "EmptySec"},
		{ID: "s1", Name: "Sec"},
	}
	rules := []BulkPolicyRuleRow{
		{SectionID: "s1", SectionName: "Sec", RuleID: "r1", RuleName: "A", Index: 1},
	}
	planned := []BulkPlannedRuleIndex{{SectionName: "Sec", RuleName: "A", IndexInSection: 1}}
	in, err := buildPolicyReorderInput(sections, rules, planned)
	require.NoError(t, err)
	require.Len(t, in.Sections, 1)
	require.Equal(t, "s1", in.Sections[0].Ref.Input)
}

func TestBuildPolicyReorderInput_plannedUnknownRule(t *testing.T) {
	t.Parallel()
	sections := []BulkPolicySectionRef{{ID: "s1", Name: "Sec"}}
	rules := []BulkPolicyRuleRow{
		{SectionID: "s1", SectionName: "Sec", RuleID: "r1", RuleName: "A", Index: 1},
	}
	planned := []BulkPlannedRuleIndex{
		{SectionName: "Sec", RuleName: "missing", IndexInSection: 1},
	}
	_, err := buildPolicyReorderInput(sections, rules, planned)
	require.Error(t, err)
}

func TestBuildPolicyReorderInput_crossSectionPlannedRules(t *testing.T) {
	t.Parallel()
	sections := []BulkPolicySectionRef{
		{ID: "sa", Name: "SecA"},
		{ID: "sb", Name: "SecB"},
	}
	rules := []BulkPolicyRuleRow{
		{SectionID: "sa", SectionName: "SecA", RuleID: "idX", RuleName: "X", Index: 1},
		{SectionID: "sb", SectionName: "SecB", RuleID: "idY", RuleName: "Y", Index: 1},
	}
	// Move Y into SecA above X; SecB ends up with no rules in the reorder payload.
	planned := []BulkPlannedRuleIndex{
		{SectionName: "SecA", RuleName: "Y", IndexInSection: 1},
		{SectionName: "SecA", RuleName: "X", IndexInSection: 2},
	}
	in, err := buildPolicyReorderInput(sections, rules, planned)
	require.NoError(t, err)
	require.Len(t, in.Sections, 1)
	require.Equal(t, "sa", in.Sections[0].Ref.Input)
	require.Len(t, in.Sections[0].Rules, 2)
	require.Equal(t, "idY", in.Sections[0].Rules[0].Ref.Input)
	require.Equal(t, "idX", in.Sections[0].Rules[1].Ref.Input)
}

func TestInternetFirewallReorderError_propagatesCallError(t *testing.T) {
	t.Parallel()
	err := internetFirewallReorderError(nil, errors.New("network"))
	require.ErrorContains(t, err, "network")
}

func TestInternetFirewallBulkPolicyClient_ReorderInvoked(t *testing.T) {
	t.Parallel()
	m := mocks.NewInternetFirewallBulkPolicyClient(t)
	ctx := context.Background()
	accountID := "acc-1"
	reorderIn := cato_models.PolicyReorderInput{
		Sections: []*cato_models.PolicyReorderSectionInput{
			{
				Ref: &cato_models.PolicyElementRefInput{By: cato_models.ObjectRefByID, Input: "s1"},
				Rules: []*cato_models.PolicyReorderRuleInput{
					{
						Ref:      &cato_models.PolicyElementRefInput{By: cato_models.ObjectRefByID, Input: "r1"},
						SubRules: []*cato_models.PolicyReorderSubRuleInput{},
					},
				},
			},
		},
	}

	m.On("PolicyInternetFirewallReorderPolicy", ctx, mock.Anything, reorderIn, accountID).
		Return(nil, errors.New("injected")).
		Once()

	_, err := m.PolicyInternetFirewallReorderPolicy(ctx, nil, reorderIn, accountID)
	require.ErrorContains(t, err, "injected")
}

func TestWanFirewallBulkPolicyClient_ReorderInvoked(t *testing.T) {
	t.Parallel()
	m := mocks.NewWanFirewallBulkPolicyClient(t)
	ctx := context.Background()
	accountID := "acc-2"
	reorderIn := cato_models.PolicyReorderInput{Sections: []*cato_models.PolicyReorderSectionInput{}}

	m.On("PolicyWanFirewallReorderPolicy", ctx, mock.Anything, reorderIn, accountID).
		Return(nil, errors.New("injected")).
		Once()

	_, err := m.PolicyWanFirewallReorderPolicy(ctx, nil, reorderIn, accountID)
	require.ErrorContains(t, err, "injected")
}
