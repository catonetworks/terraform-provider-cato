//go:build acctest

package acc

import (
	"errors"
	"fmt"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type firewallCleanupElement struct {
	name string
	id   string
}

type firewallCleanupPlan struct {
	rules       []firewallCleanupElement
	sections    []firewallCleanupElement
	subPolicies []firewallCleanupElement
	blocked     []error
}

type firewallCleanupOperations struct {
	deleteRule      func(firewallCleanupElement) (bool, error)
	deleteSection   func(firewallCleanupElement) (bool, error)
	deleteSubPolicy func(firewallCleanupElement) (bool, error)
	publish         func() error
}

type firewallPayloadError interface {
	GetErrorCode() *string
	GetErrorMessage() *string
}

func deleteInternetFirewallElements(t *testing.T) error {
	client := GetClient(t)
	result, err := client.PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, CatoAccountID)
	if err != nil {
		return fmt.Errorf("reading Internet Firewall policy: %w", err)
	}
	if revision := result.GetPolicy().GetInternetFirewall().GetPolicy().GetRevision(); revision != nil {
		if err = validateExistingPolicyRevision("Internet Firewall", revision.GetID(), revision.GetChanges()); err != nil {
			return err
		}
	}

	operations := firewallCleanupOperations{
		deleteRule: func(element firewallCleanupElement) (bool, error) {
			input := cato_models.InternetFirewallRemoveRuleInput{ID: element.id}
			response, callErr := client.PolicyInternetFirewallRemoveRule(
				ctx,
				&cato_models.InternetFirewallPolicyMutationInput{},
				input,
				CatoAccountID,
			)
			action := fmt.Sprintf("deleting Internet Firewall rule %s (%s)", element.name, element.id)
			if response == nil {
				return false, missingFirewallMutationResponse(action, callErr)
			}
			payload := response.GetPolicy().GetInternetFirewall().GetRemoveRule()
			return firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
		},
		deleteSection: func(element firewallCleanupElement) (bool, error) {
			input := cato_models.PolicyRemoveSectionInput{ID: element.id}
			response, callErr := client.PolicyInternetFirewallRemoveSection(
				ctx,
				&cato_models.InternetFirewallPolicyMutationInput{},
				input,
				CatoAccountID,
			)
			action := fmt.Sprintf("deleting Internet Firewall section %s (%s)", element.name, element.id)
			if response == nil {
				return false, missingFirewallMutationResponse(action, callErr)
			}
			payload := response.GetPolicy().GetInternetFirewall().GetRemoveSection()
			return firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
		},
		deleteSubPolicy: func(element firewallCleanupElement) (bool, error) {
			input := cato_models.InternetFirewallRemoveSubPolicyInput{
				Ref: &cato_models.InternetFirewallPolicyRefInput{
					By:    cato_models.ObjectRefByID,
					Input: element.id,
				},
			}
			response, callErr := client.PolicyInternetFirewallRemoveSubPolicy(
				ctx,
				&cato_models.InternetFirewallPolicyMutationInput{},
				input,
				CatoAccountID,
			)
			action := fmt.Sprintf("deleting Internet Firewall sub-policy %s (%s)", element.name, element.id)
			if response == nil {
				return false, missingFirewallMutationResponse(action, callErr)
			}
			payload := response.GetPolicy().GetInternetFirewall().GetRemoveSubPolicy()
			return firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
		},
		publish: func() error {
			response, callErr := client.PolicyInternetFirewallPublishPolicyRevision(
				ctx,
				&cato_models.InternetFirewallPolicyMutationInput{},
				&cato_models.PolicyPublishRevisionInput{},
				CatoAccountID,
			)
			const action = "publishing Internet Firewall policy revision"
			if response == nil {
				return missingFirewallMutationResponse(action, callErr)
			}
			payload := response.GetPolicy().GetInternetFirewall().GetPublishPolicyRevision()
			_, publishErr := firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
			return publishErr
		},
	}

	return runFirewallCleanup(internetFirewallCleanupPlan(result), operations)
}

func deleteWanFirewallElements(t *testing.T) error {
	client := GetClient(t)
	result, err := client.PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, CatoAccountID)
	if err != nil {
		return fmt.Errorf("reading WAN Firewall policy: %w", err)
	}
	if revision := result.GetPolicy().GetWanFirewall().GetPolicy().GetRevision(); revision != nil {
		if err = validateExistingPolicyRevision("WAN Firewall", revision.GetID(), revision.GetChanges()); err != nil {
			return err
		}
	}

	operations := firewallCleanupOperations{
		deleteRule: func(element firewallCleanupElement) (bool, error) {
			input := cato_models.WanFirewallRemoveRuleInput{ID: element.id}
			response, callErr := client.PolicyWanFirewallRemoveRule(ctx, input, CatoAccountID)
			action := fmt.Sprintf("deleting WAN Firewall rule %s (%s)", element.name, element.id)
			if response == nil {
				return false, missingFirewallMutationResponse(action, callErr)
			}
			payload := response.GetPolicy().GetWanFirewall().GetRemoveRule()
			return firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
		},
		deleteSection: func(element firewallCleanupElement) (bool, error) {
			input := cato_models.PolicyRemoveSectionInput{ID: element.id}
			response, callErr := client.PolicyWanFirewallRemoveSection(ctx, input, CatoAccountID)
			action := fmt.Sprintf("deleting WAN Firewall section %s (%s)", element.name, element.id)
			if response == nil {
				return false, missingFirewallMutationResponse(action, callErr)
			}
			payload := response.GetPolicy().GetWanFirewall().GetRemoveSection()
			return firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
		},
		deleteSubPolicy: func(element firewallCleanupElement) (bool, error) {
			input := cato_models.WanFirewallRemoveSubPolicyInput{
				Ref: &cato_models.WanFirewallPolicyRefInput{
					By:    cato_models.ObjectRefByID,
					Input: element.id,
				},
			}
			response, callErr := client.PolicyWanFirewallRemoveSubPolicy(ctx, input, CatoAccountID)
			action := fmt.Sprintf("deleting WAN Firewall sub-policy %s (%s)", element.name, element.id)
			if response == nil {
				return false, missingFirewallMutationResponse(action, callErr)
			}
			payload := response.GetPolicy().GetWanFirewall().GetRemoveSubPolicy()
			return firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
		},
		publish: func() error {
			response, callErr := client.PolicyWanFirewallPublishPolicyRevision(
				ctx,
				&cato_models.PolicyPublishRevisionInput{},
				CatoAccountID,
			)
			const action = "publishing WAN Firewall policy revision"
			if response == nil {
				return missingFirewallMutationResponse(action, callErr)
			}
			payload := response.GetPolicy().GetWanFirewall().GetPublishPolicyRevision()
			_, publishErr := firewallMutationResult(action, callErr, payload.GetStatus(), payload.GetErrors())
			return publishErr
		},
	}

	return runFirewallCleanup(wanFirewallCleanupPlan(result), operations)
}

func runFirewallCleanup(plan firewallCleanupPlan, operations firewallCleanupOperations) error {
	rulesChanged, ruleErr := deleteFirewallCleanupElements(plan.rules, operations.deleteRule)
	sectionsChanged, sectionErr := deleteFirewallCleanupElements(plan.sections, operations.deleteSection)
	subPoliciesChanged, subPolicyErr := deleteFirewallCleanupElements(plan.subPolicies, operations.deleteSubPolicy)

	deleteErr := errors.Join(errors.Join(plan.blocked...), ruleErr, sectionErr, subPolicyErr)
	if !rulesChanged && !sectionsChanged && !subPoliciesChanged {
		return deleteErr
	}

	return errors.Join(deleteErr, operations.publish())
}

func deleteFirewallCleanupElements(
	elements []firewallCleanupElement,
	deleteElement func(firewallCleanupElement) (bool, error),
) (bool, error) {
	changed := false
	var deleteErrs []error
	for _, element := range elements {
		elementChanged, err := deleteElement(element)
		if elementChanged {
			changed = true
		}
		if err != nil {
			deleteErrs = append(deleteErrs, err)
		}
	}

	return changed, errors.Join(deleteErrs...)
}

func internetFirewallCleanupPlan(result *cato.Policy) firewallCleanupPlan {
	policy := result.GetPolicy().GetInternetFirewall().GetPolicy()
	plan := firewallCleanupPlan{}
	unownedSubPolicies := make(map[string]string)

	for _, wrapper := range policy.GetRules() {
		if wrapper == nil || *wrapper.GetRuleType() != cato_models.PolicyRuleTypeEnumPolicyRule {
			continue
		}
		if !isFirewallCleanupCandidate(wrapper.GetRule().GetName(), wrapper.GetProperties()) {
			if subPolicy := wrapper.GetSubPolicy(); subPolicy != nil && subPolicy.GetID() != "" {
				unownedSubPolicies[subPolicy.GetID()] = wrapper.GetRule().GetName()
			}
			continue
		}
		plan.rules = append(plan.rules, firewallCleanupElement{
			name: wrapper.GetRule().GetName(),
			id:   wrapper.GetRule().GetID(),
		})
	}
	for _, wrapper := range policy.GetSections() {
		if wrapper == nil ||
			!isFirewallCleanupCandidate(wrapper.GetSection().GetName(), wrapper.GetProperties()) {
			continue
		}
		plan.sections = append(plan.sections, firewallCleanupElement{
			name: wrapper.GetSection().GetName(),
			id:   wrapper.GetSection().GetID(),
		})
	}
	for _, wrapper := range policy.GetSubPolicies() {
		if wrapper == nil ||
			!acctestRE.MatchString(wrapper.GetPolicy().GetName()) ||
			isReadOnlySubPolicy(wrapper.GetProperties()) {
			continue
		}
		if childName, blocked := unownedSubPolicies[wrapper.GetPolicy().GetID()]; blocked {
			plan.blocked = append(plan.blocked, fmt.Errorf(
				"refusing to delete Internet Firewall sub-policy %s (%s): contains non-acctest rule %s",
				wrapper.GetPolicy().GetName(),
				wrapper.GetPolicy().GetID(),
				childName,
			))
			continue
		}
		plan.subPolicies = append(plan.subPolicies, firewallCleanupElement{
			name: wrapper.GetPolicy().GetName(),
			id:   wrapper.GetPolicy().GetID(),
		})
	}

	return plan
}

func wanFirewallCleanupPlan(result *cato.Policy) firewallCleanupPlan {
	policy := result.GetPolicy().GetWanFirewall().GetPolicy()
	plan := firewallCleanupPlan{}
	unownedSubPolicies := make(map[string]string)

	for _, wrapper := range policy.GetRules() {
		if wrapper == nil || *wrapper.GetRuleType() != cato_models.PolicyRuleTypeEnumPolicyRule {
			continue
		}
		if !isFirewallCleanupCandidate(wrapper.GetRule().GetName(), wrapper.GetProperties()) {
			if subPolicy := wrapper.GetSubPolicy(); subPolicy != nil && subPolicy.GetID() != "" {
				unownedSubPolicies[subPolicy.GetID()] = wrapper.GetRule().GetName()
			}
			continue
		}
		plan.rules = append(plan.rules, firewallCleanupElement{
			name: wrapper.GetRule().GetName(),
			id:   wrapper.GetRule().GetID(),
		})
	}
	for _, wrapper := range policy.GetSections() {
		if wrapper == nil ||
			!isFirewallCleanupCandidate(wrapper.GetSection().GetName(), wrapper.GetProperties()) {
			continue
		}
		plan.sections = append(plan.sections, firewallCleanupElement{
			name: wrapper.GetSection().GetName(),
			id:   wrapper.GetSection().GetID(),
		})
	}
	for _, wrapper := range policy.GetSubPolicies() {
		if wrapper == nil ||
			!acctestRE.MatchString(wrapper.GetPolicy().GetName()) ||
			isReadOnlySubPolicy(wrapper.GetProperties()) {
			continue
		}
		if childName, blocked := unownedSubPolicies[wrapper.GetPolicy().GetID()]; blocked {
			plan.blocked = append(plan.blocked, fmt.Errorf(
				"refusing to delete WAN Firewall sub-policy %s (%s): contains non-acctest rule %s",
				wrapper.GetPolicy().GetName(),
				wrapper.GetPolicy().GetID(),
				childName,
			))
			continue
		}
		plan.subPolicies = append(plan.subPolicies, firewallCleanupElement{
			name: wrapper.GetPolicy().GetName(),
			id:   wrapper.GetPolicy().GetID(),
		})
	}

	return plan
}

func isFirewallCleanupCandidate(name string, properties []cato_models.PolicyElementPropertiesEnum) bool {
	if !acctestRE.MatchString(name) {
		return false
	}
	for _, property := range properties {
		if property == cato_models.PolicyElementPropertiesEnumSystem {
			return false
		}
	}
	return true
}

func isReadOnlySubPolicy(properties []cato_models.SubPolicyProperty) bool {
	for _, property := range properties {
		if property == cato_models.SubPolicyPropertyReadOnly {
			return true
		}
	}
	return false
}

func firewallMutationResult[T firewallPayloadError](
	action string,
	callErr error,
	status *cato_models.PolicyMutationStatus,
	payloadErrors []T,
) (bool, error) {
	var result []error
	if callErr != nil {
		result = append(result, fmt.Errorf("%s: %w", action, callErr))
	}
	succeeded := status != nil && *status == cato_models.PolicyMutationStatusSuccess
	if !succeeded {
		result = append(result, fmt.Errorf("%s: mutation status %s", action, firewallMutationStatus(status)))
	}
	for _, payloadErr := range payloadErrors {
		result = append(result, fmt.Errorf(
			"%s: API error %s: %s",
			action,
			valueOrUnknown(payloadErr.GetErrorCode()),
			valueOrUnknown(payloadErr.GetErrorMessage()),
		))
	}
	return succeeded, errors.Join(result...)
}

func missingFirewallMutationResponse(action string, callErr error) error {
	return errors.Join(
		func() error {
			if callErr == nil {
				return nil
			}
			return fmt.Errorf("%s: %w", action, callErr)
		}(),
		fmt.Errorf("%s: missing mutation response", action),
	)
}

func firewallMutationStatus(status *cato_models.PolicyMutationStatus) string {
	if status == nil {
		return "<nil>"
	}
	return string(*status)
}

func valueOrUnknown(value *string) string {
	if value == nil || *value == "" {
		return "<unknown>"
	}
	return *value
}

func validateExistingPolicyRevision(policyName, revisionID string, changes int64) error {
	if changes == 0 {
		return nil
	}
	return fmt.Errorf(
		"%s policy has existing draft revision %q with %d changes; refusing to publish or discard unowned changes",
		policyName,
		revisionID,
		changes,
	)
}
