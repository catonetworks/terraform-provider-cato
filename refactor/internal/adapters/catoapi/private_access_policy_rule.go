package catoapi

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
)

var _ network.DeletePrivateAccessPolicyRuleAdapter = (*Adapter)(nil)

func (a *Adapter) RemovePrivateAccessPolicyRule(ctx context.Context, accountID, ruleID string) error {
	const operation = "remove private-access policy rule"
	result, err := a.client.PolicyPrivateAccessDeleteRule(ctx, accountID, cato_models.PrivateAccessRemoveRuleInput{ID: ruleID})
	if err != nil {
		return requestError(operation, err)
	}
	if result == nil || result.Policy.PrivateAccess == nil {
		return invalidResponse(operation)
	}
	mutation := result.Policy.PrivateAccess.RemoveRule
	codes := make([]string, len(mutation.Errors))
	for index, failure := range mutation.Errors {
		if failure != nil && failure.ErrorCode != nil {
			codes[index] = *failure.ErrorCode
		}
	}
	return privateAccessPolicyMutationError(operation, mutation.GetStatus(), codes)
}
