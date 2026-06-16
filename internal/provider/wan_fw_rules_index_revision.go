package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

const wanTerraformDraftRevisionName = "terraform-wf-bulk-move"

func isActiveRevisionConflict(message string) bool {
	return strings.Contains(message, "other active revisions exist")
}

func publishWanStaleDraftRevisions(
	ctx context.Context,
	client WanRulesIndexClient,
	accountID string,
) error {
	if client == nil {
		return errors.New("wan rules index client is not configured")
	}

	policy, err := client.PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, accountID)
	if err != nil {
		return fmt.Errorf("policy wan firewall query failed: %w", err)
	}

	revisions := policy.GetPolicy().GetWanFirewall().GetRevisionsWanFirewallPolicyQueries()
	if revisions == nil || len(revisions.GetRevision()) == 0 {
		return nil
	}

	_, err = client.PolicyWanFirewallPublishPolicyRevision(ctx, &cato_models.PolicyPublishRevisionInput{}, accountID)
	if err != nil {
		return fmt.Errorf("publish policy revision failed: %w", err)
	}

	return nil
}

func ensureWanDraftMutationInput(
	ctx context.Context,
	client WanRulesIndexClient,
	accountID string,
) (*cato_models.WanFirewallPolicyMutationInput, error) {
	if client == nil {
		return nil, errors.New("wan rules index client is not configured")
	}

	policy, err := client.PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, accountID)
	if err != nil {
		return nil, fmt.Errorf("policy wan firewall query failed: %w", err)
	}

	if revisions := policy.GetPolicy().GetWanFirewall().GetRevisionsWanFirewallPolicyQueries(); revisions != nil {
		for _, revision := range revisions.GetRevision() {
			if revision == nil {
				continue
			}
			revisionID := revision.GetID()
			if revisionID == "" {
				continue
			}
			return wanMutationInputForRevision(revisionID), nil
		}
	}

	createResp, err := client.PolicyWanFirewallCreatePolicyRevision(
		ctx,
		cato_models.PolicyCreateRevisionInput{
			Name:        wanTerraformDraftRevisionName,
			Description: "terraform provider bulk WF move rule",
		},
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("create policy revision failed: %w", err)
	}

	createPayload := createResp.GetPolicy().GetWanFirewall().GetCreatePolicyRevision()
	if createPayload.GetStatus() != nil && *createPayload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		if apiErrors := createPayload.GetErrors(); len(apiErrors) > 0 && apiErrors[0].GetErrorMessage() != nil {
			return nil, errors.New(*apiErrors[0].GetErrorMessage())
		}
		return nil, errors.New("create policy revision failed")
	}

	revisionID := createPayload.GetPolicy().GetRevisionWanFirewallPolicy().GetID()
	if revisionID == "" {
		return nil, errors.New("create policy revision returned empty revision id")
	}

	return wanMutationInputForRevision(revisionID), nil
}

func wanMutationInputForRevision(revisionID string) *cato_models.WanFirewallPolicyMutationInput {
	return &cato_models.WanFirewallPolicyMutationInput{
		Revision: &cato_models.PolicyMutationRevisionInput{
			ID: &revisionID,
		},
	}
}

func wanMoveSectionError(
	moveResp *cato_go_sdk.PolicyWanFirewallMoveSection,
	err error,
) error {
	if err != nil {
		return err
	}
	if moveResp == nil || moveResp.GetPolicy() == nil || moveResp.GetPolicy().GetWanFirewall() == nil {
		return nil
	}

	payload := moveResp.GetPolicy().GetWanFirewall().GetMoveSection()
	if payload == nil {
		return nil
	}
	if payload.GetStatus() != nil && *payload.GetStatus() == cato_models.PolicyMutationStatusSuccess {
		return nil
	}

	apiErrors := payload.GetErrors()
	if len(apiErrors) > 0 && apiErrors[0].GetErrorMessage() != nil {
		return errors.New(*apiErrors[0].GetErrorMessage())
	}

	return errors.New("move section failed")
}
