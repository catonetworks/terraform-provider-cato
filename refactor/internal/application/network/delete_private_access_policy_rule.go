package network

import (
	"context"
	"errors"
	"strings"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
)

type DeletePrivateAccessPolicyRuleAdapter interface {
	RemovePrivateAccessPolicyRule(ctx context.Context, accountID, ruleID string) error
	PublishPrivateAccessPolicyRevision(ctx context.Context, accountID string) error
	QueryPrivateAccessPolicy(ctx context.Context, accountID string, revision entities.PolicyRevision) (entities.PrivateAccessPolicy, error)
}

// RevisionCoordinator serializes participating workflows for an account's
// private-access policy. The caller must release a successfully acquired lock.
type RevisionCoordinator interface {
	Acquire(ctx context.Context, accountID string) (release func(), err error)
}

// DeletePrivateAccessPolicyRuleUseCase coordinates one deletion through publication and verification.
// Its only mutable state belongs to the injected, shared coordinator.
type DeletePrivateAccessPolicyRuleUseCase struct {
	adapter     DeletePrivateAccessPolicyRuleAdapter
	coordinator RevisionCoordinator
	retry       RetryPolicy
}

// DeletePrivateAccessPolicyRuleCommand identifies a rule in the account's private-access policy.
type DeletePrivateAccessPolicyRuleCommand struct {
	AccountID string
	RuleID    string
}

func NewDeletePrivateAccessPolicyRuleUseCase(
	adapter DeletePrivateAccessPolicyRuleAdapter,
	coordinator RevisionCoordinator, retry RetryPolicy,
) (*DeletePrivateAccessPolicyRuleUseCase, error) {
	if adapter == nil || coordinator == nil {
		return nil, errors.New("all deletion dependencies must be configured")
	}
	if retry.Attempts < 1 || retry.Delay <= 0 {
		return nil, errors.New("retry attempts and delay must be positive")
	}
	return &DeletePrivateAccessPolicyRuleUseCase{adapter: adapter, coordinator: coordinator, retry: retry}, nil
}

func (u *DeletePrivateAccessPolicyRuleUseCase) Execute(ctx context.Context, command DeletePrivateAccessPolicyRuleCommand) error {
	if strings.TrimSpace(command.AccountID) == "" {
		return &ValidationError{Field: "account_id"}
	}
	if strings.TrimSpace(command.RuleID) == "" {
		return &ValidationError{Field: "id"}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	release, err := u.coordinator.Acquire(ctx, command.AccountID)
	if err != nil {
		return &DeleteError{Stage: StageCoordinate, Cause: err}
	}
	defer release()

	err = u.retry.conflict(ctx, func() error {
		return classifyPrivateAccessPolicyError(u.adapter.RemovePrivateAccessPolicyRule(ctx, command.AccountID, command.RuleID))
	})
	if err != nil && !errors.Is(err, ErrRuleNotFound) {
		return &DeleteError{Stage: StageRemove, Cause: err}
	}

	publishErr := u.retry.conflict(ctx, func() error {
		return classifyPrivateAccessPolicyError(u.adapter.PublishPrivateAccessPolicyRevision(ctx, command.AccountID))
	})
	if publishErr != nil && !canObservePublication(publishErr) {
		return &DeleteError{Stage: StagePublish, RemovalCompleted: true, Cause: publishErr}
	}
	// A missing draft or lost response does not prove completion. Only the public
	// policy view can establish that the rule is no longer published.
	if err := u.verify(ctx, command); err != nil {
		return &DeleteError{Stage: StageVerify, RemovalCompleted: true, Cause: errors.Join(publishErr, err)}
	}
	return nil
}

func canObservePublication(err error) bool {
	return errors.Is(err, ErrNoDraft) || errors.Is(err, ErrUnavailable) || errors.Is(err, ErrInvalidResponse)
}

func (u *DeletePrivateAccessPolicyRuleUseCase) verify(ctx context.Context, command DeletePrivateAccessPolicyRuleCommand) error {
	for attempt := 0; attempt < u.retry.Attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		policy, err := u.adapter.QueryPrivateAccessPolicy(ctx, command.AccountID, entities.PublicPolicyRevision)
		if err != nil {
			return classifyPrivateAccessPolicyError(err)
		}
		if !policy.HasRule(command.RuleID) {
			return nil
		}
		if attempt+1 < u.retry.Attempts {
			if err := u.retry.wait(ctx); err != nil {
				return err
			}
		}
	}
	return ErrStillPublished
}
