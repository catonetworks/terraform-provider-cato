//go:build acctest

package acc

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
)

// CleanupFirewallAndWANPolicyRevisions discards any unpublished IF, WF, and WAN
// revisions left by policy acceptance tests. This prevents one package's failed
// publish/delete path from blocking later reorderPolicy tests.
func CleanupFirewallAndWANPolicyRevisions(t *testing.T) {
	t.Helper()
	if accmock.ACCMockActive {
		return
	}

	client := GetClient(t)
	if err := cleanupFirewallAndWANPolicyRevisions(client); err != nil {
		t.Errorf("cleanup firewall/WAN policy revisions: %v", err)
	}
}

func cleanupFirewallAndWANPolicyRevisions(client *cato.Client) error {
	var cleanupErr error

	policy, err := client.Policy(ctx, nil, nil, CatoAccountID)
	if err != nil {
		return fmt.Errorf("list firewall revisions: %w", err)
	}

	for _, rev := range policy.GetPolicy().GetInternetFirewall().GetRevisionsInternetFirewallPolicyQueries().GetRevision() {
		if rev.GetID() == "" {
			continue
		}
		if err := discardInternetFirewallPolicyRevision(client, rev.GetID()); err != nil {
			cleanupErr = errors.Join(cleanupErr, err)
		}
	}
	for _, rev := range policy.GetPolicy().GetWanFirewall().GetRevisionsWanFirewallPolicyQueries().GetRevision() {
		if rev.GetID() == "" {
			continue
		}
		if err := discardWanFirewallPolicyRevision(client, rev.GetID()); err != nil {
			cleanupErr = errors.Join(cleanupErr, err)
		}
	}

	wanNetwork, err := client.WanNetworkPolicy(ctx, CatoAccountID, nil)
	if err != nil {
		cleanupErr = errors.Join(cleanupErr, fmt.Errorf("list WAN network revision: %w", err))
	} else if rev := wanNetwork.GetPolicy().GetWanNetwork().GetPolicy().GetRevision(); rev != nil && rev.GetID() != "" {
		if err := discardWanNetworkPolicyRevision(client, rev.GetID()); err != nil {
			cleanupErr = errors.Join(cleanupErr, err)
		}
	}

	return cleanupErr
}

func discardInternetFirewallPolicyRevision(client *cato.Client, revisionID string) error {
	resp, err := client.PolicyInternetFirewallDiscardPolicyRevision(ctx,
		&cato_models.InternetFirewallPolicyMutationInput{
			Revision: &cato_models.PolicyMutationRevisionInput{ID: &revisionID},
		},
		&cato_models.PolicyDiscardRevisionInput{ID: &revisionID},
		CatoAccountID,
	)
	if err != nil {
		return fmt.Errorf("discard internet firewall revision %s: %w", revisionID, err)
	}
	if err := cleanupMutationErrors(resp.GetPolicy().GetInternetFirewall().GetDiscardPolicyRevision().GetErrors()); err != nil {
		return fmt.Errorf("discard internet firewall revision %s: %w", revisionID, err)
	}
	return nil
}

func discardWanFirewallPolicyRevision(client *cato.Client, revisionID string) error {
	resp, err := client.PolicyWanFirewallDiscardPolicyRevision(ctx,
		&cato_models.PolicyDiscardRevisionInput{ID: &revisionID},
		CatoAccountID,
		&cato_models.WanFirewallPolicyMutationInput{
			Revision: &cato_models.PolicyMutationRevisionInput{ID: &revisionID},
		},
	)
	if err != nil {
		return fmt.Errorf("discard WAN firewall revision %s: %w", revisionID, err)
	}
	if err := cleanupMutationErrors(resp.GetPolicy().GetWanFirewall().GetDiscardPolicyRevision().GetErrors()); err != nil {
		return fmt.Errorf("discard WAN firewall revision %s: %w", revisionID, err)
	}
	return nil
}

func discardWanNetworkPolicyRevision(client *cato.Client, revisionID string) error {
	resp, err := client.PolicyWanNetworkDiscardPolicyRevision(
		ctx,
		CatoAccountID,
		&cato_models.PolicyDiscardRevisionInput{ID: &revisionID},
		&cato_models.WanNetworkPolicyMutationInput{
			Revision: &cato_models.PolicyMutationRevisionInput{ID: &revisionID},
		},
	)
	if err != nil {
		return fmt.Errorf("discard WAN network revision %s: %w", revisionID, err)
	}
	if err := cleanupMutationErrors(resp.GetPolicy().GetWanNetwork().GetDiscardPolicyRevision().GetErrors()); err != nil {
		return fmt.Errorf("discard WAN network revision %s: %w", revisionID, err)
	}
	return nil
}

// A revision may already have been published/discarded between list and cleanup.
// Keep every other API error visible, including mixed missing/failed responses.
func cleanupMutationErrors[T interface {
	GetErrorCode() *string
	GetErrorMessage() *string
}](items []T) error {
	var failures []string
	for _, item := range items {
		code, message := "", ""
		if value := item.GetErrorCode(); value != nil {
			code = *value
		}
		if value := item.GetErrorMessage(); value != nil {
			message = *value
		}
		if code == "PolicyRevisionNotFound" {
			continue
		}
		if code == "" && message == "" {
			message = "API returned a policy error without details"
		}
		failures = append(failures, fmt.Sprintf("%s: %s", code, message))
	}
	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}
