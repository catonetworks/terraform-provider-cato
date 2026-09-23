//go:build acctest

package acc

import (
	"errors"
	"fmt"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

type applicationPolicyCleanupOperations struct {
	removeRule    cleanupRemoveFunc
	removeSection cleanupRemoveFunc
	publish       func() error
}

func deleteApplicationControlElements(t *testing.T) error {
	client := GetClient(t)
	result, err := client.ApplicationControlPolicy(ctx, CatoAccountID)
	if err != nil {
		return fmt.Errorf("reading Application Control policy: %w", err)
	}
	if result == nil {
		return errors.New("reading Application Control policy: missing response")
	}
	if revision := result.GetPolicy().GetApplicationControl().GetPolicy().GetRevision(); revision != nil {
		if err = validateExistingPolicyRevision("Application Control", revision.GetID(), revision.GetChanges()); err != nil {
			return err
		}
	}

	rules, sections := applicationControlCleanupElements(result)
	operations := applicationPolicyCleanupOperations{
		removeRule: func(id string) (cleanupMutationResult, error) {
			response, removeErr := client.PolicyApplicationControlRemoveRule(
				ctx,
				cato_models.ApplicationControlRemoveRuleInput{ID: id},
				CatoAccountID,
			)
			if removeErr != nil {
				return cleanupMutationResult{}, removeErr
			}
			if response == nil {
				return cleanupMutationResult{}, errors.New("missing mutation response")
			}
			payload := response.GetPolicy().GetApplicationControl().GetRemoveRule()
			payloadErrors := payload.GetErrors()
			return cleanupMutationResult{
				status:            payload.GetStatus(),
				payloadErrors:     payloadErrors,
				payloadErrorCount: len(payloadErrors),
			}, nil
		},
		removeSection: func(id string) (cleanupMutationResult, error) {
			response, removeErr := client.PolicyApplicationControlRemoveSection(
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
			payload := response.GetPolicy().GetApplicationControl().GetRemoveSection()
			payloadErrors := payload.GetErrors()
			return cleanupMutationResult{
				status:            payload.GetStatus(),
				payloadErrors:     payloadErrors,
				payloadErrorCount: len(payloadErrors),
			}, nil
		},
		publish: func() error {
			response, publishErr := client.PolicyApplicationControlPublishPolicyRevision(ctx, CatoAccountID)
			if publishErr != nil {
				return fmt.Errorf("publishing Application Control policy revision: %w", publishErr)
			}
			if response == nil {
				return errors.New("publishing Application Control policy revision: missing response")
			}
			payload := response.GetPolicy().GetApplicationControl().GetPublishPolicyRevision()
			payloadErrors := payload.GetErrors()
			return validateCleanupMutation("publishing Application Control policy revision", cleanupMutationResult{
				status:            payload.GetStatus(),
				payloadErrors:     payloadErrors,
				payloadErrorCount: len(payloadErrors),
			})
		},
	}

	// Application Control is a singleton policy. Global cleanup cannot restore its
	// baseline because no pre-test baseline is stored; only owned elements are removed.
	return runApplicationPolicyCleanup(rules, sections, operations)
}

func deleteAppTenantRestrictionElements(t *testing.T) error {
	client := GetClient(t)
	result, err := client.AppTenantRestrictionPolicy(ctx, CatoAccountID)
	if err != nil {
		return fmt.Errorf("reading App Tenant Restriction policy: %w", err)
	}
	if result == nil {
		return errors.New("reading App Tenant Restriction policy: missing response")
	}
	if revision := result.GetPolicy().GetAppTenantRestriction().GetPolicy().GetRevision(); revision != nil {
		if err = validateExistingPolicyRevision(
			"App Tenant Restriction",
			revision.GetID(),
			revision.GetChanges(),
		); err != nil {
			return err
		}
	}

	rules, sections := appTenantRestrictionCleanupElements(result)
	operations := applicationPolicyCleanupOperations{
		removeRule: func(id string) (cleanupMutationResult, error) {
			response, removeErr := client.PolicyAppTenantRestrictionRemoveRule(
				ctx,
				cato_models.AppTenantRestrictionRemoveRuleInput{ID: id},
				CatoAccountID,
			)
			if removeErr != nil {
				return cleanupMutationResult{}, removeErr
			}
			if response == nil {
				return cleanupMutationResult{}, errors.New("missing mutation response")
			}
			payload := response.GetPolicy().GetAppTenantRestriction().GetRemoveRule()
			payloadErrors := payload.GetErrors()
			return cleanupMutationResult{
				status:            payload.GetStatus(),
				payloadErrors:     payloadErrors,
				payloadErrorCount: len(payloadErrors),
			}, nil
		},
		removeSection: func(id string) (cleanupMutationResult, error) {
			response, removeErr := client.PolicyAppTenantRestrictionRemoveSection(
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
			payload := response.GetPolicy().GetAppTenantRestriction().GetRemoveSection()
			payloadErrors := payload.GetErrors()
			return cleanupMutationResult{
				status:            payload.GetStatus(),
				payloadErrors:     payloadErrors,
				payloadErrorCount: len(payloadErrors),
			}, nil
		},
		publish: func() error {
			response, publishErr := client.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, CatoAccountID)
			if publishErr != nil {
				return fmt.Errorf("publishing App Tenant Restriction policy revision: %w", publishErr)
			}
			if response == nil {
				return errors.New("publishing App Tenant Restriction policy revision: missing response")
			}
			payload := response.GetPolicy().GetAppTenantRestriction().GetPublishPolicyRevision()
			payloadErrors := payload.GetErrors()
			return validateCleanupMutation("publishing App Tenant Restriction policy revision", cleanupMutationResult{
				status:            payload.GetStatus(),
				payloadErrors:     payloadErrors,
				payloadErrorCount: len(payloadErrors),
			})
		},
	}

	return runApplicationPolicyCleanup(rules, sections, operations)
}

func runApplicationPolicyCleanup(
	rules []cleanupPolicyElement,
	sections []cleanupPolicyElement,
	operations applicationPolicyCleanupOperations,
) error {
	rulesChanged, ruleErr := removeCleanupPolicyElements(rules, operations.removeRule)
	sectionsChanged, sectionErr := removeCleanupPolicyElements(sections, operations.removeSection)
	cleanupErr := errors.Join(ruleErr, sectionErr)
	if !rulesChanged && !sectionsChanged {
		return cleanupErr
	}
	return errors.Join(cleanupErr, operations.publish())
}

func applicationControlCleanupElements(
	result *cato.ApplicationControlPolicy,
) (rules []cleanupPolicyElement, sections []cleanupPolicyElement) {
	policy := result.GetPolicy().GetApplicationControl().GetPolicy()
	for _, wrapper := range policy.GetRules() {
		if wrapper == nil || !acctestRE.MatchString(wrapper.GetRule().GetName()) {
			continue
		}
		rule := wrapper.GetRule()
		rules = append(rules, cleanupPolicyElement{
			kind: "Application Control rule",
			name: rule.GetName(),
			id:   rule.GetID(),
		})
	}
	for _, wrapper := range policy.GetSections() {
		if wrapper == nil || !acctestRE.MatchString(wrapper.GetSection().GetName()) {
			continue
		}
		section := wrapper.GetSection()
		sections = append(sections, cleanupPolicyElement{
			kind: "Application Control section",
			name: section.GetName(),
			id:   section.GetID(),
		})
	}
	return rules, sections
}

func appTenantRestrictionCleanupElements(
	result *cato.AppTenantRestrictionPolicy,
) (rules []cleanupPolicyElement, sections []cleanupPolicyElement) {
	policy := result.GetPolicy().GetAppTenantRestriction().GetPolicy()
	for _, wrapper := range policy.GetRules() {
		if wrapper == nil || !acctestRE.MatchString(wrapper.GetRule().GetName()) {
			continue
		}
		rule := wrapper.GetRule()
		rules = append(rules, cleanupPolicyElement{
			kind: "App Tenant Restriction rule",
			name: rule.GetName(),
			id:   rule.GetID(),
		})
	}
	for _, wrapper := range policy.GetSections() {
		if wrapper == nil || !acctestRE.MatchString(wrapper.GetSection().GetName()) {
			continue
		}
		section := wrapper.GetSection()
		sections = append(sections, cleanupPolicyElement{
			kind: "App Tenant Restriction section",
			name: section.GetName(),
			id:   section.GetID(),
		})
	}
	return rules, sections
}
