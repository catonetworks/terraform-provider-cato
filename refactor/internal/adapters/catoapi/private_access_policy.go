package catoapi

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"

	privateaccess "github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
)

var _ privateaccess.ReadPrivateAccessPolicyRuleAdapter = (*Adapter)(nil)

// The SDK's generated read has no revision selector. Use the same authenticated
// SDK transport, selecting only rule IDs and passing all inputs as variables.
const policyQuery = `query refactorPrivateAccessPolicy($accountID: ID!, $revision: PolicyRevisionType!) {
  policy(accountId: $accountID) {
    privateAccess {
      policy(input: {revision: {type: $revision}}) {
        rules { rule { id } }
      }
    }
  }
}`

type privateAccessPolicyResponse struct {
	Policy *struct {
		PrivateAccess *struct {
			Policy *struct {
				Rules []*struct {
					Rule *struct {
						ID *string `json:"id"`
					} `json:"rule"`
				} `json:"rules"`
			} `json:"policy"`
		} `json:"privateAccess"`
	} `json:"policy"`
}

func (a *Adapter) QueryPrivateAccessPolicy(
	ctx context.Context, accountID string, revision entities.PolicyRevision,
) (entities.PrivateAccessPolicy, error) {
	const operation = "query private-access policy"
	var result privateAccessPolicyResponse
	err := a.client.Client.Post(ctx, "refactorPrivateAccessPolicy", policyQuery, &result,
		map[string]any{"accountID": accountID, "revision": string(revision)})
	if err != nil {
		return entities.PrivateAccessPolicy{}, requestError(operation, err)
	}
	if result.Policy == nil || result.Policy.PrivateAccess == nil || result.Policy.PrivateAccess.Policy == nil {
		return entities.PrivateAccessPolicy{}, invalidResponse(operation)
	}
	rules := result.Policy.PrivateAccess.Policy.Rules
	if rules == nil {
		return entities.PrivateAccessPolicy{}, invalidResponse(operation)
	}
	policy := entities.PrivateAccessPolicy{Rules: make([]entities.PrivateAccessPolicyRule, 0, len(rules))}
	for _, row := range rules {
		if row == nil || row.Rule == nil || row.Rule.ID == nil || *row.Rule.ID == "" {
			return entities.PrivateAccessPolicy{}, invalidResponse(operation)
		}
		policy.Rules = append(policy.Rules, entities.PrivateAccessPolicyRule{ID: *row.Rule.ID})
	}
	return policy, nil
}

func (a *Adapter) PublishPrivateAccessPolicyRevision(ctx context.Context, accountID string) error {
	const operation = "publish private-access policy revision"
	result, err := a.client.PolicyPrivateAccessPublishRevision(ctx, accountID)
	if err != nil {
		return requestError(operation, err)
	}
	if result == nil || result.Policy.PrivateAccess == nil {
		return invalidResponse(operation)
	}
	mutation := result.Policy.PrivateAccess.PublishPolicyRevision
	codes := make([]string, len(mutation.Errors))
	for index, failure := range mutation.Errors {
		if failure != nil && failure.ErrorCode != nil {
			codes[index] = *failure.ErrorCode
		}
	}
	return privateAccessPolicyMutationError(operation, mutation.GetStatus(), codes)
}

// privateAccessPolicyMutationError validates the policy mutation envelope.
// API codes retain their original meaning for the application to interpret.
func privateAccessPolicyMutationError(operation string, status *cato_models.PolicyMutationStatus, codes []string) error {
	if status == nil {
		return invalidResponse(operation)
	}
	if *status == cato_models.PolicyMutationStatusSuccess {
		if len(codes) != 0 {
			return invalidResponse(operation)
		}
		return nil
	}
	if *status != cato_models.PolicyMutationStatusFailure || len(codes) == 0 {
		return invalidResponse(operation)
	}
	for _, code := range codes {
		if code == "" {
			return invalidResponse(operation)
		}
	}
	return &CatoAPIError{Operation: operation, Kind: privateaccess.ErrRejected, Codes: codes}
}
