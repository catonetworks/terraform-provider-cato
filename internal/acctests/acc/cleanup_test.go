//go:build acctest

package acc

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

var acctestRE = regexp.MustCompile(`^acctest_`)

func TestCleanupAccTestResources(t *testing.T) {
	if os.Getenv("ACCTEST_CLEANUP") != "true" {
		t.Log("Skipping cleanup of test resources. Set ACCTEST_CLEANUP=true to enable.")
		return
	}
	var errors []error
	var helpers = []func(t *testing.T) error{
		deletePrivateAccessRules,
		deleteSocketLanPolicyResources,
		deleteSites,
	}

	GetClient(t)

	run := func(helper func(t *testing.T) error) {
		if err := helper(t); err != nil {
			errors = append(errors, err)
		}
	}

	for _, helper := range helpers {
		run(helper)
	}

	if len(errors) > 0 {
		t.Fatalf("cleanup errors: %v", errors)
	}
}

func deleteSites(t *testing.T) error {
	sites := getEntities(t, resSite)
	for _, site := range sites {
		if acctestRE.MatchString(site.Name) {
			if _, err := catoClient.SiteRemoveSite(ctx, site.ID, CatoAccountID); err != nil {
				return fmt.Errorf("deleting site %s (%s): %v", site.Name, site.ID, err)
			}
		}
	}
	return nil
}

func deleteSocketLanPolicyResources(t *testing.T) error {
	client := GetClient(t)
	result, err := client.PolicySocketLanPolicy(ctx, CatoAccountID, nil)
	if err != nil {
		return fmt.Errorf("reading socket LAN policy: %w", err)
	}

	policy := result.GetPolicy().GetSocketLan().GetPolicy()
	if policy == nil {
		return nil
	}

	firewallRuleIDs := make(map[string]struct{})
	networkRuleIDs := make(map[string]struct{})
	hasChanges := false
	for _, ruleWrapper := range policy.GetRules() {
		if ruleWrapper == nil {
			continue
		}

		rule := ruleWrapper.GetRule()
		ruleName := rule.GetName()
		isAcctestRule := acctestRE.MatchString(ruleName)

		for _, firewallWrapper := range rule.GetFirewall() {
			if firewallWrapper == nil {
				continue
			}
			firewallRule := firewallWrapper.GetRule()
			if !isAcctestRule && !acctestRE.MatchString(firewallRule.GetName()) {
				continue
			}
			if firewallRule.GetID() != "" {
				firewallRuleIDs[firewallRule.GetID()] = struct{}{}
			}
		}

		if isAcctestRule && ruleWrapper.RuleType == cato_models.PolicyRuleTypeEnumPolicyRule && rule.GetID() != "" {
			networkRuleIDs[rule.GetID()] = struct{}{}
		}
	}

	for ruleID := range firewallRuleIDs {
		input := cato_models.SocketLanFirewallRemoveRuleInput{ID: ruleID}
		result, err := client.PolicySocketLanFirewallRemoveRule(ctx, CatoAccountID, nil, input)
		if err != nil {
			return fmt.Errorf("deleting socket LAN firewall rule %s: %w", ruleID, err)
		}
		remove := result.GetPolicy().GetSocketLan().GetFirewall().GetRemoveRule()
		if err := checkSocketLanMutation(
			"deleting socket LAN firewall rule",
			remove.GetStatus(),
			len(remove.GetErrors()),
		); err != nil {
			return fmt.Errorf("%s %s: %w", "deleting socket LAN firewall rule", ruleID, err)
		}
	}

	for ruleID := range networkRuleIDs {
		input := cato_models.SocketLanRemoveRuleInput{ID: ruleID}
		result, err := client.PolicySocketLanRemoveRule(ctx, nil, input, CatoAccountID)
		if err != nil {
			return fmt.Errorf("deleting socket LAN network rule %s: %w", ruleID, err)
		}
		remove := result.GetPolicy().GetSocketLan().GetRemoveRule()
		if err := checkSocketLanMutation(
			"deleting socket LAN network rule",
			remove.GetStatus(),
			len(remove.GetErrors()),
		); err != nil {
			return fmt.Errorf("%s %s: %w", "deleting socket LAN network rule", ruleID, err)
		}
	}

	for _, subPolicy := range policy.GetSubPolicies() {
		if subPolicy == nil || !acctestRE.MatchString(subPolicy.GetPolicy().GetName()) {
			continue
		}

		input := cato_models.SocketLanRemoveSubPolicyInput{
			Ref: &cato_models.SocketLanPolicyRefInput{
				By:    cato_models.ObjectRefByID,
				Input: subPolicy.GetPolicy().GetID(),
			},
		}
		result, err := client.PolicySocketLanRemoveSubPolicy(ctx, nil, input, CatoAccountID)
		if err != nil {
			return fmt.Errorf("deleting socket LAN sub-policy %s: %w", subPolicy.GetPolicy().GetID(), err)
		}
		remove := result.GetPolicy().GetSocketLan().GetRemoveSubPolicy()
		if err := checkSocketLanMutation(
			"deleting socket LAN sub-policy",
			remove.GetStatus(),
			len(remove.GetErrors()),
		); err != nil {
			return fmt.Errorf("%s %s: %w", "deleting socket LAN sub-policy", subPolicy.GetPolicy().GetID(), err)
		}
		hasChanges = true
	}

	for _, section := range policy.GetSections() {
		if section == nil || !acctestRE.MatchString(section.Section.GetName()) {
			continue
		}

		input := cato_models.PolicyRemoveSectionInput{ID: section.Section.GetID()}
		result, err := client.PolicySocketLanRemoveSection(ctx, nil, input, CatoAccountID)
		if err != nil {
			return fmt.Errorf("deleting socket LAN section %s: %w", section.Section.GetID(), err)
		}
		remove := result.GetPolicy().GetSocketLan().GetRemoveSection()
		if err := checkSocketLanMutation(
			"deleting socket LAN section",
			remove.GetStatus(),
			len(remove.GetErrors()),
		); err != nil {
			return fmt.Errorf("%s %s: %w", "deleting socket LAN section", section.Section.GetID(), err)
		}
		hasChanges = true
	}

	if len(firewallRuleIDs) == 0 && len(networkRuleIDs) == 0 && !hasChanges {
		return nil
	}

	publishResult, err := client.PolicySocketLanPublishPolicyRevision(
		ctx,
		nil,
		&cato_models.PolicyPublishRevisionInput{},
		CatoAccountID,
	)
	if err != nil {
		return fmt.Errorf("publishing socket LAN policy revision: %w", err)
	}
	publish := publishResult.GetPolicy().GetSocketLan().GetPublishPolicyRevision()
	if err := checkSocketLanMutation(
		"publishing socket LAN policy revision",
		publish.GetStatus(),
		len(publish.GetErrors()),
	); err != nil {
		return err
	}

	return nil
}

func checkSocketLanMutation(operation string, status *cato_models.PolicyMutationStatus, errorCount int) error {
	if errorCount > 0 {
		return fmt.Errorf("%s returned %d API error(s)", operation, errorCount)
	}
	if status == nil {
		return fmt.Errorf("%s returned no mutation status", operation)
	}
	if *status != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("%s returned mutation status %q", operation, *status)
	}
	return nil
}

func deletePrivateAccessRules(t *testing.T) error {
	client := GetClient(t)
	result, err := client.PolicyReadPrivateAccessPolicy(ctx, CatoAccountID, nil)
	if err != nil {
		return err
	}
	rules := result.GetPolicy().GetPrivateAccess().GetPolicy().GetRules()
	if len(rules) == 0 {
		return nil
	}

	for _, rule := range rules {
		if !acctestRE.MatchString(rule.Rule.GetName()) {
			continue
		}

		input := cato_models.PrivateAccessRemoveRuleInput{ID: rule.Rule.ID}
		_, err = client.PolicyPrivateAccessDeleteRule(context.Background(), CatoAccountID, input)
		if err != nil {
			return fmt.Errorf("deleting private access rule %s (%s): %v", rule.Rule.GetName(), rule.Rule.ID, err)
		}
	}
	if _, err = client.PolicyPrivateAccessPublishRevision(ctx, CatoAccountID); err != nil {
		return fmt.Errorf("publishing private access revision: %v", err)
	}

	return nil
}
