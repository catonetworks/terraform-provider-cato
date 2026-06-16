package provider

import (
	"context"
	"errors"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

const ifwTerraformDraftRevisionName = "terraform-ifw-bulk-move"

func ensureIfwDraftMutationInput(
	ctx context.Context,
	client IfwRulesIndexClient,
	accountID string,
) (*cato_models.InternetFirewallPolicyMutationInput, error) {
	if client == nil {
		return nil, errors.New("ifw rules index client is not configured")
	}

	policy, err := client.PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, accountID)
	if err != nil {
		return nil, fmt.Errorf("policy internet firewall query failed: %w", err)
	}

	if revisions := policy.GetPolicy().GetInternetFirewall().GetRevisionsInternetFirewallPolicyQueries(); revisions != nil {
		for _, revision := range revisions.GetRevision() {
			if revision == nil {
				continue
			}
			revisionID := revision.GetID()
			if revisionID == "" {
				continue
			}
			return ifwMutationInputForRevision(revisionID), nil
		}
	}

	createResp, err := client.PolicyInternetFirewallCreatePolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		cato_models.PolicyCreateRevisionInput{
			Name:        ifwTerraformDraftRevisionName,
			Description: "terraform provider bulk IF move rule",
		},
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("create policy revision failed: %w", err)
	}

	createPayload := createResp.GetPolicy().GetInternetFirewall().GetCreatePolicyRevision()
	if createPayload.GetStatus() != nil && *createPayload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		if apiErrors := createPayload.GetErrors(); len(apiErrors) > 0 && apiErrors[0].GetErrorMessage() != nil {
			return nil, errors.New(*apiErrors[0].GetErrorMessage())
		}
		return nil, errors.New("create policy revision failed")
	}

	revisionID := createPayload.GetPolicy().GetRevisionInternetFirewallPolicy().GetID()
	if revisionID == "" {
		return nil, errors.New("create policy revision returned empty revision id")
	}

	return ifwMutationInputForRevision(revisionID), nil
}

func ifwMutationInputForRevision(revisionID string) *cato_models.InternetFirewallPolicyMutationInput {
	return &cato_models.InternetFirewallPolicyMutationInput{
		Revision: &cato_models.PolicyMutationRevisionInput{
			ID: &revisionID,
		},
	}
}
