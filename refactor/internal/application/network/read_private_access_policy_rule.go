package network

import (
	"context"
	"errors"
	"strings"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
)

type ReadPrivateAccessPolicyRuleAdapter interface {
	QueryPrivateAccessPolicy(ctx context.Context, accountID string, revision entities.PolicyRevision) (entities.PrivateAccessPolicy, error)
}

type ReadPrivateAccessPolicyRuleUseCase struct {
	adapter ReadPrivateAccessPolicyRuleAdapter
}

type ReadPrivateAccessPolicyRuleCommand struct {
	AccountID string
	RuleID    string
}

func NewReadPrivateAccessPolicyRuleUseCase(adapter ReadPrivateAccessPolicyRuleAdapter) (*ReadPrivateAccessPolicyRuleUseCase, error) {
	if adapter == nil {
		return nil, errors.New("private-access policy adapter is required")
	}
	return &ReadPrivateAccessPolicyRuleUseCase{adapter: adapter}, nil
}

func (r *ReadPrivateAccessPolicyRuleUseCase) Execute(ctx context.Context, command ReadPrivateAccessPolicyRuleCommand) (bool, error) {
	if strings.TrimSpace(command.AccountID) == "" {
		return false, &ValidationError{Field: "account_id"}
	}
	if strings.TrimSpace(command.RuleID) == "" {
		return false, &ValidationError{Field: "id"}
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	policy, err := r.adapter.QueryPrivateAccessPolicy(ctx, command.AccountID, entities.PublicPolicyRevision)
	if err != nil {
		return false, classifyPrivateAccessPolicyError(err)
	}
	return policy.HasRule(command.RuleID), nil
}
