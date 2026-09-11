//go:build acctest

package acc

import (
	"errors"
	"fmt"
	"os"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

func TestCleanupAccTestResources(t *testing.T) {
	if os.Getenv("ACCTEST_CLEANUP") != "true" {
		t.Log("Skipping cleanup of test resources. Set ACCTEST_CLEANUP=true to enable.")
		return
	}
	var cleanupErrors []error

	GetClient(t)

	run := func(helper func(t *testing.T) error) bool {
		if err := helper(t); err != nil {
			cleanupErrors = append(cleanupErrors, err)
			return false
		}
		return true
	}

	policyClean := true
	for _, helper := range []func(t *testing.T) error{
		deletePrivateAccessRules,
		deleteInternetFirewallElements,
		deleteWanFirewallElements,
		deleteSocketLanRules,
		deleteTLSInspectionElements,
		deleteWanNetworkElements,
		deleteApplicationControlElements,
		deleteAppTenantRestrictionElements,
	} {
		if !run(helper) {
			policyClean = false
		}
	}

	if policyClean {
		run(deleteSitesAndDependencies)
		run(deleteAcctestGlobalIPRanges)
	}
	run(deleteAcctestAccounts)

	if len(cleanupErrors) > 0 {
		t.Fatalf("cleanup errors: %v", cleanupErrors)
	}
}

func deleteSocketLanRules(t *testing.T) error {
	client := GetClient(t)
	result, err := client.PolicySocketLanPolicy(ctx, CatoAccountID, nil)
	if err != nil {
		return fmt.Errorf("reading Socket LAN policy: %v", err)
	}

	policy := result.GetPolicy().GetSocketLan().GetPolicy()
	if revision := policy.GetRevision(); revision != nil {
		if err = validateExistingPolicyRevision("Socket LAN", revision.GetID(), revision.GetChanges()); err != nil {
			return err
		}
	}
	rulesChanged, ruleErr := cleanupSocketLanRules(client, policy.GetRules())
	sectionsChanged, sectionErr := cleanupSocketLanSections(client, policy.GetSections())
	subPoliciesChanged, subPolicyErr := cleanupSocketLanSubPolicies(client, policy.GetSubPolicies())
	cleanupErr := errors.Join(ruleErr, sectionErr, subPolicyErr)
	if !rulesChanged && !sectionsChanged && !subPoliciesChanged {
		return cleanupErr
	}

	return errors.Join(cleanupErr, publishSocketLanPolicy(client))
}

func cleanupSocketLanRules(
	client *cato.Client,
	rules []*cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules,
) (bool, error) {
	changed := false
	var cleanupErr error
	for _, ruleWrapper := range rules {
		if ruleWrapper == nil || *ruleWrapper.GetRuleType() != cato_models.PolicyRuleTypeEnumPolicyRule {
			continue
		}

		rule := ruleWrapper.GetRule()
		deleteParent := acctestRE.MatchString(rule.GetName())
		firewallChanged, deleteErr := deleteSocketLanFirewallRules(client, rule.GetFirewall())
		if deleteErr != nil {
			cleanupErr = errors.Join(cleanupErr, deleteErr)
		}
		if firewallChanged {
			changed = true
		}

		if !deleteParent {
			continue
		}
		if hasUnownedSocketLanFirewallRules(rule.GetFirewall()) {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf(
				"refusing to delete Socket LAN network rule %s (%s): contains non-acctest firewall rules",
				rule.GetName(),
				rule.GetID(),
			))
			continue
		}
		if deleteErr != nil {
			continue
		}

		if deleteErr = deleteSocketLanNetworkRule(client, rule); deleteErr != nil {
			cleanupErr = errors.Join(cleanupErr, deleteErr)
			continue
		}
		changed = true
	}
	return changed, cleanupErr
}

func cleanupSocketLanSections(
	client *cato.Client,
	sections []*cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections,
) (bool, error) {
	changed := false
	var cleanupErr error
	for _, wrapper := range sections {
		if wrapper == nil || !isFirewallCleanupCandidate(
			wrapper.GetSection().GetName(),
			wrapper.GetProperties(),
		) {
			continue
		}
		sectionChanged, deleteErr := deleteSocketLanSection(client, wrapper.GetSection())
		if deleteErr != nil {
			cleanupErr = errors.Join(cleanupErr, deleteErr)
		}
		if sectionChanged {
			changed = true
		}
	}
	return changed, cleanupErr
}

func cleanupSocketLanSubPolicies(
	client *cato.Client,
	subPolicies []*cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_SubPolicies,
) (bool, error) {
	changed := false
	var cleanupErr error
	for _, wrapper := range subPolicies {
		if wrapper == nil ||
			!acctestRE.MatchString(wrapper.GetPolicy().GetName()) ||
			isReadOnlySubPolicy(wrapper.GetProperties()) {
			continue
		}
		subPolicyChanged, deleteErr := deleteSocketLanSubPolicy(client, wrapper.GetPolicy())
		if deleteErr != nil {
			cleanupErr = errors.Join(cleanupErr, deleteErr)
		}
		if subPolicyChanged {
			changed = true
		}
	}
	return changed, cleanupErr
}

func deleteSocketLanFirewallRules(
	client *cato.Client,
	firewallRules []*cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall,
) (bool, error) {
	changed := false
	for _, firewallWrapper := range firewallRules {
		if firewallWrapper == nil {
			continue
		}

		firewallRule := firewallWrapper.GetRule()
		if !acctestRE.MatchString(firewallRule.GetName()) {
			continue
		}

		input := cato_models.SocketLanFirewallRemoveRuleInput{ID: firewallRule.GetID()}
		result, callErr := client.PolicySocketLanFirewallRemoveRule(ctx, CatoAccountID, nil, input)
		action := fmt.Sprintf(
			"deleting Socket LAN firewall rule %s (%s)",
			firewallRule.GetName(),
			firewallRule.GetID(),
		)
		if result == nil {
			return changed, missingFirewallMutationResponse(action, callErr)
		}
		payload := result.GetPolicy().GetSocketLan().GetFirewall().GetRemoveRule()
		succeeded, mutationErr := firewallMutationResult(
			action,
			callErr,
			payload.GetStatus(),
			payload.GetErrors(),
		)
		if mutationErr != nil {
			return changed, mutationErr
		}
		changed = changed || succeeded
	}

	return changed, nil
}

func hasUnownedSocketLanFirewallRules(
	firewallRules []*cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall,
) bool {
	for _, wrapper := range firewallRules {
		if wrapper != nil && !acctestRE.MatchString(wrapper.GetRule().GetName()) {
			return true
		}
	}
	return false
}

func deleteSocketLanNetworkRule(
	client *cato.Client,
	rule *cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule,
) error {
	input := cato_models.SocketLanRemoveRuleInput{ID: rule.GetID()}
	result, callErr := client.PolicySocketLanRemoveRule(ctx, nil, input, CatoAccountID)
	action := fmt.Sprintf("deleting Socket LAN network rule %s (%s)", rule.GetName(), rule.GetID())
	if result == nil {
		return missingFirewallMutationResponse(action, callErr)
	}
	payload := result.GetPolicy().GetSocketLan().GetRemoveRule()
	_, mutationErr := firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
	return mutationErr
}

func deleteSocketLanSection(
	client *cato.Client,
	section *cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_Sections_Section,
) (bool, error) {
	input := cato_models.PolicyRemoveSectionInput{ID: section.GetID()}
	result, callErr := client.PolicySocketLanRemoveSection(ctx, nil, input, CatoAccountID)
	action := fmt.Sprintf("deleting Socket LAN section %s (%s)", section.GetName(), section.GetID())
	if result == nil {
		return false, missingFirewallMutationResponse(action, callErr)
	}
	payload := result.GetPolicy().GetSocketLan().GetRemoveSection()
	return firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
}

func deleteSocketLanSubPolicy(
	client *cato.Client,
	subPolicy *cato.PolicySocketLanPolicy_Policy_SocketLan_Policy_SubPolicies_Policy,
) (bool, error) {
	input := cato_models.SocketLanRemoveSubPolicyInput{
		Ref: &cato_models.SocketLanPolicyRefInput{
			By:    cato_models.ObjectRefByID,
			Input: subPolicy.GetID(),
		},
	}
	result, callErr := client.PolicySocketLanRemoveSubPolicy(ctx, nil, input, CatoAccountID)
	action := fmt.Sprintf("deleting Socket LAN sub-policy %s (%s)", subPolicy.GetName(), subPolicy.GetID())
	if result == nil {
		return false, missingFirewallMutationResponse(action, callErr)
	}
	payload := result.GetPolicy().GetSocketLan().GetRemoveSubPolicy()
	return firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
}

func publishSocketLanPolicy(client *cato.Client) error {
	result, err := client.PolicySocketLanPublishPolicyRevision(
		ctx, nil, &cato_models.PolicyPublishRevisionInput{}, CatoAccountID,
	)
	const action = "publishing Socket LAN policy revision"
	if result == nil {
		return missingFirewallMutationResponse(action, err)
	}
	payload := result.GetPolicy().GetSocketLan().GetPublishPolicyRevision()
	_, publishErr := firewallMutationResult(action, err, payload.GetStatus(), payload.GetErrors())
	return publishErr
}

func deletePrivateAccessRules(t *testing.T) error {
	client := GetClient(t)
	result, err := client.PolicyReadPrivateAccessPolicy(ctx, CatoAccountID)
	if err != nil {
		return err
	}
	policy := result.GetPolicy().GetPrivateAccess().GetPolicy()
	if revision := policy.GetRevision(); revision != nil {
		if err = validateExistingPolicyRevision("Private Access", revision.GetID(), revision.GetChanges()); err != nil {
			return err
		}
	}
	rules := policy.GetRules()

	changed := false
	var cleanupErr error
	for _, rule := range rules {
		if rule == nil || !acctestRE.MatchString(rule.Rule.GetName()) {
			continue
		}

		input := cato_models.PrivateAccessRemoveRuleInput{ID: rule.Rule.ID}
		removeResult, callErr := client.PolicyPrivateAccessDeleteRule(ctx, CatoAccountID, input)
		action := fmt.Sprintf("deleting private access rule %s (%s)", rule.Rule.GetName(), rule.Rule.ID)
		if removeResult == nil {
			cleanupErr = errors.Join(cleanupErr, missingFirewallMutationResponse(action, callErr))
			continue
		}
		payload := removeResult.GetPolicy().GetPrivateAccess().GetRemoveRule()
		succeeded, mutationErr := firewallMutationResult(
			action,
			callErr,
			payload.GetStatus(),
			payload.GetErrors(),
		)
		if mutationErr != nil {
			cleanupErr = errors.Join(cleanupErr, mutationErr)
		}
		changed = changed || succeeded
	}
	if !changed {
		return cleanupErr
	}

	publishResult, callErr := client.PolicyPrivateAccessPublishRevision(ctx, CatoAccountID)
	const action = "publishing private access revision"
	if publishResult == nil {
		return errors.Join(cleanupErr, missingFirewallMutationResponse(action, callErr))
	}
	payload := publishResult.GetPolicy().GetPrivateAccess().GetPublishPolicyRevision()
	_, publishErr := firewallMutationResult(
		action,
		callErr,
		payload.GetStatus(),
		payload.GetErrors(),
	)
	return errors.Join(cleanupErr, publishErr)
}
