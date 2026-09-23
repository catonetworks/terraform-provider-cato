//go:build acctest

package acc

import (
	"errors"
	"fmt"
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

	wanNetwork, err := client.WanNetworkPolicy(ctx, CatoAccountID)
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
	if resp == nil {
		return fmt.Errorf("discard internet firewall revision %s: missing response", revisionID)
	}
	payload := resp.GetPolicy().GetInternetFirewall().GetDiscardPolicyRevision()
	payloadErrors := payload.GetErrors()
	return validateCleanupMutation(
		fmt.Sprintf("discard internet firewall revision %s", revisionID),
		cleanupMutationResult{
			status:            payload.GetStatus(),
			payloadErrors:     payloadErrors,
			payloadErrorCount: len(payloadErrors),
		},
	)
}

func discardWanFirewallPolicyRevision(client *cato.Client, revisionID string) error {
	resp, err := client.PolicyWanFirewallDiscardPolicyRevision(ctx,
		&cato_models.PolicyDiscardRevisionInput{ID: &revisionID},
		CatoAccountID,
	)
	if err != nil {
		return fmt.Errorf("discard WAN firewall revision %s: %w", revisionID, err)
	}
	if resp == nil {
		return fmt.Errorf("discard WAN firewall revision %s: missing response", revisionID)
	}
	payload := resp.GetPolicy().GetWanFirewall().GetDiscardPolicyRevision()
	payloadErrors := payload.GetErrors()
	return validateCleanupMutation(
		fmt.Sprintf("discard WAN firewall revision %s", revisionID),
		cleanupMutationResult{
			status:            payload.GetStatus(),
			payloadErrors:     payloadErrors,
			payloadErrorCount: len(payloadErrors),
		},
	)
}

func discardWanNetworkPolicyRevision(client *cato.Client, revisionID string) error {
	resp, err := client.PolicyWanNetworkDiscardPolicyRevision(ctx, CatoAccountID)
	if err != nil {
		return fmt.Errorf("discard WAN network revision %s: %w", revisionID, err)
	}
	if resp == nil {
		return fmt.Errorf("discard WAN network revision %s: missing response", revisionID)
	}
	payload := resp.GetPolicy().GetWanNetwork().GetDiscardPolicyRevision()
	payloadErrors := payload.GetErrors()
	return validateCleanupMutation(
		fmt.Sprintf("discard WAN network revision %s", revisionID),
		cleanupMutationResult{
			status:            payload.GetStatus(),
			payloadErrors:     payloadErrors,
			payloadErrorCount: len(payloadErrors),
		},
	)
}
