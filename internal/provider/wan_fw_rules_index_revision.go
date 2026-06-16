package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const wanTerraformDraftRevisionName = "terraform-wf-bulk-move"

func isActiveRevisionConflict(message string) bool {
	return strings.Contains(message, "other active revisions exist")
}

func discardWanStaleDraftRevisions(
	ctx context.Context,
	client WanRulesIndexClient,
	accountID string,
) error {
	if client == nil {
		return errors.New("wan rules index client is not configured")
	}

	discardResp, err := client.PolicyWanFirewallDiscardPolicyRevision(
		ctx,
		&cato_models.PolicyDiscardRevisionInput{},
		accountID,
	)
	if err != nil {
		return fmt.Errorf("discard policy revision failed: %w", err)
	}

	return wanDiscardPolicyRevisionError(discardResp, nil)
}

func wanDiscardPolicyRevisionError(
	discardResp *cato_go_sdk.PolicyWanFirewallDiscardPolicyRevision,
	err error,
) error {
	if err != nil {
		return err
	}
	if discardResp == nil || discardResp.GetPolicy() == nil || discardResp.GetPolicy().GetWanFirewall() == nil {
		return nil
	}

	payload := discardResp.GetPolicy().GetWanFirewall().GetDiscardPolicyRevision()
	if payload == nil {
		return nil
	}
	if payload.GetStatus() != nil && *payload.GetStatus() == cato_models.PolicyMutationStatusSuccess {
		return nil
	}

	apiErrors := payload.GetErrors()
	if len(apiErrors) > 0 {
		if apiErrors[0].GetErrorCode() != nil && *apiErrors[0].GetErrorCode() == policyRevisionNotFound {
			return nil
		}
		if apiErrors[0].GetErrorMessage() != nil {
			return errors.New(*apiErrors[0].GetErrorMessage())
		}
	}

	return errors.New("discard policy revision failed")
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

func wanReorderPolicyWithRetry(
	ctx context.Context,
	client WanRulesIndexClient,
	accountID string,
	mutationInput *cato_models.WanFirewallPolicyMutationInput,
	reorderInput cato_models.PolicyReorderInput,
) (*cato_go_sdk.PolicyWanFirewallReorderPolicy, *cato_models.WanFirewallPolicyMutationInput, error) {
	reorderResult, err := client.PolicyWanFirewallReorderPolicy(ctx, mutationInput, reorderInput, accountID)
	if err != nil && isActiveRevisionConflict(err.Error()) {
		tflog.Warn(ctx, "Write.PolicyWanFirewallReorderPolicy.active_revision_retry", map[string]interface{}{
			"error": err.Error(),
		})
		_, publishErr := client.PolicyWanFirewallPublishPolicyRevision(
			ctx,
			&cato_models.PolicyPublishRevisionInput{},
			accountID,
		)
		if publishErr != nil {
			tflog.Warn(ctx, "Write.PolicyWanFirewallReorderPolicy.active_revision_retry.publish_error", map[string]interface{}{
				"error": publishErr.Error(),
			})
			return reorderResult, mutationInput, err
		}

		mutationInput, err = ensureWanDraftMutationInput(ctx, client, accountID)
		if err != nil {
			return nil, mutationInput, err
		}
		reorderResult, err = client.PolicyWanFirewallReorderPolicy(ctx, mutationInput, reorderInput, accountID)
	}

	return reorderResult, mutationInput, err
}

func wanReorderPolicyError(
	reorderResp *cato_go_sdk.PolicyWanFirewallReorderPolicy,
	err error,
) error {
	if err != nil {
		return err
	}
	if reorderResp == nil || reorderResp.GetPolicy() == nil || reorderResp.GetPolicy().GetWanFirewall() == nil {
		return nil
	}

	payload := reorderResp.GetPolicy().GetWanFirewall().GetReorderPolicy()
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

	return errors.New("reorder policy failed")
}
