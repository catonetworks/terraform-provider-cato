//go:build acctest

package acc

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type cleanupPolicyElement struct {
	kind string
	name string
	id   string
}

type cleanupMutationResult struct {
	status            *cato_models.PolicyMutationStatus
	payloadErrors     any
	payloadErrorCount int
}

type cleanupRemoveFunc func(string) (cleanupMutationResult, error)

func deleteTLSInspectionElements(t *testing.T) error {
	t.Helper()

	client := GetClient(t)
	result, err := client.Tlsinspectpolicy(ctx, CatoAccountID)
	if err != nil {
		return fmt.Errorf("reading TLS Inspection policy: %w", err)
	}

	policy := result.GetPolicy().GetTLSInspect().GetPolicy()
	if err = validateTLSInspectionRevision(policy); err != nil {
		return err
	}
	rules := make([]cleanupPolicyElement, 0)
	for _, wrapper := range policy.GetRules() {
		if wrapper == nil {
			continue
		}
		rule := wrapper.GetRule()
		if acctestRE.MatchString(rule.GetName()) {
			rules = append(rules, cleanupPolicyElement{
				kind: "TLS Inspection rule",
				name: rule.GetName(),
				id:   rule.GetID(),
			})
		}
	}

	sections := make([]cleanupPolicyElement, 0)
	for _, wrapper := range policy.GetSections() {
		if wrapper == nil {
			continue
		}
		section := wrapper.GetSection()
		if acctestRE.MatchString(section.GetName()) {
			sections = append(sections, cleanupPolicyElement{
				kind: "TLS Inspection section",
				name: section.GetName(),
				id:   section.GetID(),
			})
		}
	}

	rulesChanged, cleanupErr := removeCleanupPolicyElements(rules, func(id string) (cleanupMutationResult, error) {
		response, removeErr := client.PolicyTLSInspectRemoveRule(
			ctx,
			cato_models.TLSInspectRemoveRuleInput{ID: id},
			CatoAccountID,
		)
		if removeErr != nil {
			return cleanupMutationResult{}, removeErr
		}
		if response == nil {
			return cleanupMutationResult{}, errors.New("missing mutation response")
		}

		payload := response.GetPolicy().GetTLSInspect().GetRemoveRule()
		payloadErrors := payload.GetErrors()
		return cleanupMutationResult{
			status:            payload.GetStatus(),
			payloadErrors:     payloadErrors,
			payloadErrorCount: len(payloadErrors),
		}, nil
	})

	sectionsChanged, sectionErr := removeCleanupPolicyElements(
		sections,
		func(id string) (cleanupMutationResult, error) {
			response, removeErr := client.PolicyTLSInspectRemoveSection(
				ctx,
				cato_models.PolicyRemoveSectionInput{ID: id},
				CatoAccountID,
			)
			if removeErr != nil {
				return cleanupMutationResult{}, removeErr
			}
			if response == nil {
				return cleanupMutationResult{}, errors.New("missing mutation response")
			}

			payload := response.GetPolicy().GetTLSInspect().GetRemoveSection()
			payloadErrors := payload.GetErrors()
			return cleanupMutationResult{
				status:            payload.GetStatus(),
				payloadErrors:     payloadErrors,
				payloadErrorCount: len(payloadErrors),
			}, nil
		},
	)
	cleanupErr = errors.Join(cleanupErr, sectionErr)

	if !rulesChanged && !sectionsChanged {
		return cleanupErr
	}

	publishErr := publishTLSInspectionCleanup(client)
	return errors.Join(cleanupErr, publishErr)
}

func deleteWanNetworkElements(t *testing.T) error {
	t.Helper()

	client := GetClient(t)
	result, err := client.WanNetworkPolicy(ctx, CatoAccountID)
	if err != nil {
		return fmt.Errorf("reading WAN Network policy: %w", err)
	}

	policy := result.GetPolicy().GetWanNetwork().GetPolicy()
	if err = validateWanNetworkRevision(policy); err != nil {
		return err
	}
	rules := make([]cleanupPolicyElement, 0)
	for _, wrapper := range policy.GetRules() {
		if wrapper == nil {
			continue
		}
		rule := wrapper.GetRule()
		if acctestRE.MatchString(rule.GetName()) {
			rules = append(rules, cleanupPolicyElement{
				kind: "WAN Network rule",
				name: rule.GetName(),
				id:   rule.GetID(),
			})
		}
	}

	sections := make([]cleanupPolicyElement, 0)
	for _, wrapper := range policy.GetSections() {
		if wrapper == nil {
			continue
		}
		section := wrapper.GetSection()
		if acctestRE.MatchString(section.GetName()) {
			sections = append(sections, cleanupPolicyElement{
				kind: "WAN Network section",
				name: section.GetName(),
				id:   section.GetID(),
			})
		}
	}

	rulesChanged, cleanupErr := removeCleanupPolicyElements(rules, func(id string) (cleanupMutationResult, error) {
		response, removeErr := client.PolicyWanNetworkRemoveRule(
			ctx,
			cato_models.WanNetworkRemoveRuleInput{ID: id},
			CatoAccountID,
		)
		if removeErr != nil {
			return cleanupMutationResult{}, removeErr
		}
		if response == nil {
			return cleanupMutationResult{}, errors.New("missing mutation response")
		}

		payload := response.GetPolicy().GetWanNetwork().GetRemoveRule()
		payloadErrors := payload.GetErrors()
		return cleanupMutationResult{
			status:            payload.GetStatus(),
			payloadErrors:     payloadErrors,
			payloadErrorCount: len(payloadErrors),
		}, nil
	})

	sectionsChanged, sectionErr := removeCleanupPolicyElements(
		sections,
		func(id string) (cleanupMutationResult, error) {
			response, removeErr := client.PolicyWanNetworkRemoveSection(
				ctx,
				cato_models.PolicyRemoveSectionInput{ID: id},
				CatoAccountID,
			)
			if removeErr != nil {
				return cleanupMutationResult{}, removeErr
			}
			if response == nil {
				return cleanupMutationResult{}, errors.New("missing mutation response")
			}

			payload := response.GetPolicy().GetWanNetwork().GetRemoveSection()
			payloadErrors := payload.GetErrors()
			return cleanupMutationResult{
				status:            payload.GetStatus(),
				payloadErrors:     payloadErrors,
				payloadErrorCount: len(payloadErrors),
			}, nil
		},
	)
	cleanupErr = errors.Join(cleanupErr, sectionErr)

	if !rulesChanged && !sectionsChanged {
		return cleanupErr
	}

	publishErr := publishWanNetworkCleanup(client)
	return errors.Join(cleanupErr, publishErr)
}

func removeCleanupPolicyElements(
	elements []cleanupPolicyElement,
	remove cleanupRemoveFunc,
) (bool, error) {
	changed := false
	var cleanupErr error

	for _, element := range elements {
		action := fmt.Sprintf("deleting %s %s (%s)", element.kind, element.name, element.id)
		result, err := remove(element.id)
		if err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("%s: %w", action, err))
			continue
		}
		if err = validateCleanupMutation(action, result); err != nil {
			cleanupErr = errors.Join(cleanupErr, err)
			continue
		}
		changed = true
	}

	return changed, cleanupErr
}

func publishTLSInspectionCleanup(client *cato.Client) error {
	response, err := client.PolicyTLSInspectPublishPolicyRevision(ctx, CatoAccountID)
	if err != nil {
		return fmt.Errorf("publishing TLS Inspection policy revision: %w", err)
	}
	if response == nil {
		return errors.New("publishing TLS Inspection policy revision: missing mutation response")
	}

	payload := response.GetPolicy().GetTLSInspect().GetPublishPolicyRevision()
	payloadErrors := payload.GetErrors()
	return validateCleanupMutation("publishing TLS Inspection policy revision", cleanupMutationResult{
		status:            payload.GetStatus(),
		payloadErrors:     payloadErrors,
		payloadErrorCount: len(payloadErrors),
	})
}

func publishWanNetworkCleanup(client *cato.Client) error {
	response, err := client.PolicyWanNetworkPublishPolicyRevision(ctx, CatoAccountID)
	if err != nil {
		return fmt.Errorf("publishing WAN Network policy revision: %w", err)
	}
	if response == nil {
		return errors.New("publishing WAN Network policy revision: missing mutation response")
	}

	payload := response.GetPolicy().GetWanNetwork().GetPublishPolicyRevision()
	payloadErrors := payload.GetErrors()
	return validateCleanupMutation("publishing WAN Network policy revision", cleanupMutationResult{
		status:            payload.GetStatus(),
		payloadErrors:     payloadErrors,
		payloadErrorCount: len(payloadErrors),
	})
}

func validateCleanupMutation(action string, result cleanupMutationResult) error {
	if result.status != nil &&
		*result.status == cato_models.PolicyMutationStatusSuccess &&
		result.payloadErrorCount == 0 {
		return nil
	}

	status := "<nil>"
	if result.status != nil {
		status = string(*result.status)
	}

	payloadErrors, err := json.Marshal(result.payloadErrors)
	if err != nil {
		return fmt.Errorf(
			"%s: mutation status %s, payload errors could not be encoded: %v",
			action,
			status,
			err,
		)
	}
	return fmt.Errorf("%s: mutation status %s, payload errors %s", action, status, payloadErrors)
}

func validateTLSInspectionRevision(policy *cato.Tlsinspectpolicy_Policy_TLSInspect_Policy) error {
	revision := policy.GetRevision()
	if revision == nil {
		return nil
	}
	return validateExistingPolicyRevision("TLS Inspection", revision.GetID(), revision.GetChanges())
}

func validateWanNetworkRevision(policy *cato.WanNetworkPolicy_Policy_WanNetwork_Policy) error {
	revision := policy.GetRevision()
	if revision == nil {
		return nil
	}
	return validateExistingPolicyRevision("WAN Network", revision.GetID(), revision.GetChanges())
}
