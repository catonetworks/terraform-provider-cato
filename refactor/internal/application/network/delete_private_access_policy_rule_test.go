package network

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
)

type deletionBackend struct {
	calls         []string
	removeErrors  []error
	publishErrors []error
	exists        []bool
	readError     error
}

func nextError(values *[]error) error {
	if len(*values) == 0 {
		return nil
	}
	err := (*values)[0]
	*values = (*values)[1:]
	return err
}

func (b *deletionBackend) RemovePrivateAccessPolicyRule(_ context.Context, _, _ string) error {
	b.calls = append(b.calls, "remove")
	return nextError(&b.removeErrors)
}

func (b *deletionBackend) PublishPrivateAccessPolicyRevision(_ context.Context, _ string) error {
	b.calls = append(b.calls, "publish")
	return nextError(&b.publishErrors)
}

func (b *deletionBackend) QueryPrivateAccessPolicy(
	_ context.Context, _ string, revision entities.PolicyRevision,
) (entities.PrivateAccessPolicy, error) {
	if revision != entities.PublicPolicyRevision {
		return entities.PrivateAccessPolicy{}, errors.New("expected PUBLIC revision")
	}
	b.calls = append(b.calls, "observe")
	policy := entities.PrivateAccessPolicy{}
	if len(b.exists) > 0 {
		if b.exists[0] {
			policy.Rules = []entities.PrivateAccessPolicyRule{{ID: "rule-456"}}
		}
		b.exists = b.exists[1:]
	}
	return policy, b.readError
}

type testCoordinator struct{ calls *[]string }

func (c testCoordinator) Acquire(ctx context.Context, _ string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	*c.calls = append(*c.calls, "acquire")
	return func() { *c.calls = append(*c.calls, "release") }, nil
}

func newTestDeletion(t *testing.T, backend *deletionBackend) *DeletePrivateAccessPolicyRuleUseCase {
	t.Helper()
	useCase, err := NewDeletePrivateAccessPolicyRuleUseCase(backend, testCoordinator{&backend.calls}, RetryPolicy{Attempts: 3, Delay: time.Nanosecond})
	require.NoError(t, err)
	return useCase
}

func TestDeleteRule(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		backend deletionBackend
		want    []string
		stage   Stage
		removed bool
	}{
		{"success", deletionBackend{}, []string{"acquire", "remove", "publish", "observe", "release"}, "", false},
		{"already removed", deletionBackend{removeErrors: []error{rejectedCatoAPIError("ruleNotExist")}}, []string{"acquire", "remove", "publish", "observe", "release"}, "", false},
		{"no draft still verifies", deletionBackend{publishErrors: []error{rejectedCatoAPIError("PolicyRevisionNotFound")}}, []string{"acquire", "remove", "publish", "observe", "release"}, "", false},
		{"remove rejected", deletionBackend{removeErrors: []error{rejectedCatoAPIError("PermissionDenied")}}, []string{"acquire", "remove", "release"}, StageRemove, false},
		{"publish rejected", deletionBackend{publishErrors: []error{rejectedCatoAPIError("PermissionDenied")}}, []string{"acquire", "remove", "publish", "release"}, StagePublish, true},
		{"read failure", deletionBackend{readError: catoAPIErrorStub{kind: ErrUnavailable}}, []string{"acquire", "remove", "publish", "observe", "release"}, StageVerify, true},
		{"rule remains active", deletionBackend{exists: []bool{true, true, true}}, []string{"acquire", "remove", "publish", "observe", "observe", "observe", "release"}, StageVerify, true},
		{"eventual visibility", deletionBackend{exists: []bool{true, false}}, []string{"acquire", "remove", "publish", "observe", "observe", "release"}, "", false},
		{"remove conflict", deletionBackend{removeErrors: []error{rejectedCatoAPIError("reorderPolicyBlockedByActiveSessions")}}, []string{"acquire", "remove", "remove", "publish", "observe", "release"}, "", false},
		{"publish conflict does not remove again", deletionBackend{publishErrors: []error{rejectedCatoAPIError("reorderPolicyBlockedByActiveSessions")}}, []string{"acquire", "remove", "publish", "publish", "observe", "release"}, "", false},
		{"uncertain publish resolved by read", deletionBackend{publishErrors: []error{catoAPIErrorStub{kind: ErrUnavailable}}}, []string{"acquire", "remove", "publish", "observe", "release"}, "", false},
		{"conflicts exhausted", deletionBackend{removeErrors: []error{
			rejectedCatoAPIError("reorderPolicyBlockedByActiveSessions"),
			rejectedCatoAPIError("reorderPolicyBlockedByActiveSessions"),
			rejectedCatoAPIError("reorderPolicyBlockedByActiveSessions"),
		}}, []string{"acquire", "remove", "remove", "remove", "release"}, StageRemove, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := newTestDeletion(t, &tc.backend).Execute(context.Background(), DeletePrivateAccessPolicyRuleCommand{AccountID: "account-123", RuleID: "rule-456"})
			if tc.stage == "" {
				require.NoError(t, err)
			} else {
				var failure *DeleteError
				require.ErrorAs(t, err, &failure)
				require.Equal(t, tc.stage, failure.Stage)
				require.Equal(t, tc.removed, failure.RemovalCompleted)
			}
			require.Equal(t, tc.want, tc.backend.calls)
		})
	}
}

func TestDeleteRuleResumesAfterPublicationFailure(t *testing.T) {
	t.Parallel()
	backend := &deletionBackend{
		removeErrors:  []error{nil, rejectedCatoAPIError("ruleNotExist")},
		publishErrors: []error{rejectedCatoAPIError("PermissionDenied"), nil},
	}
	useCase := newTestDeletion(t, backend)
	command := DeletePrivateAccessPolicyRuleCommand{AccountID: "account-123", RuleID: "rule-456"}
	require.ErrorIs(t, useCase.Execute(context.Background(), command), ErrRejected)
	require.NoError(t, useCase.Execute(context.Background(), command))
	require.Equal(t, []string{"acquire", "remove", "publish", "release", "acquire", "remove", "publish", "observe", "release"}, backend.calls)
}

func TestDeleteRuleRejectsInvalidCommandsBeforeIO(t *testing.T) {
	t.Parallel()
	for _, command := range []DeletePrivateAccessPolicyRuleCommand{{}, {AccountID: "account-123"}, {AccountID: " ", RuleID: "rule-456"}} {
		backend := &deletionBackend{}
		var failure *ValidationError
		require.ErrorAs(t, newTestDeletion(t, backend).Execute(context.Background(), command), &failure)
		require.Empty(t, backend.calls)
	}
}

func TestDeleteRuleCancellation(t *testing.T) {
	t.Parallel()
	backend := &deletionBackend{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, newTestDeletion(t, backend).Execute(ctx, DeletePrivateAccessPolicyRuleCommand{AccountID: "account-123", RuleID: "rule-456"}), context.Canceled)
	require.Empty(t, backend.calls)
}

func TestDeleteRuleRetainsUncertainPublishCause(t *testing.T) {
	t.Parallel()
	backend := &deletionBackend{
		publishErrors: []error{catoAPIErrorStub{kind: ErrUnavailable}},
		readError:     catoAPIErrorStub{kind: ErrInvalidResponse},
	}
	err := newTestDeletion(t, backend).Execute(context.Background(), DeletePrivateAccessPolicyRuleCommand{AccountID: "account-123", RuleID: "rule-456"})
	require.True(t, errors.Is(err, ErrUnavailable))
	require.ErrorIs(t, err, ErrInvalidResponse)
}

func TestDeleteRuleInterpretsAPICodesInApplication(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		codes []string
		want  error
	}{
		{"already absent", []string{"ruleNotExist"}, nil},
		{"conflict retried", []string{"reorderPolicyBlockedByActiveSessions"}, nil},
		{"mixed errors rejected", []string{"RuleNotFound", "PermissionDenied"}, ErrRejected},
		{"unknown code rejected", []string{"PermissionDenied"}, ErrRejected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			backend := &deletionBackend{removeErrors: []error{
				rejectedCatoAPIError(tc.codes...),
			}}
			err := newTestDeletion(t, backend).Execute(context.Background(),
				DeletePrivateAccessPolicyRuleCommand{AccountID: "account-123", RuleID: "rule-456"})
			require.ErrorIs(t, err, tc.want)
			if tc.want != nil {
				require.NotContains(t, backend.calls, "publish")
			}
			if tc.name == "conflict retried" {
				require.Equal(t, []string{"acquire", "remove", "remove", "publish", "observe", "release"}, backend.calls)
			}
		})
	}
}

func TestReadRuleUsesCommandAndEvaluatesPolicy(t *testing.T) {
	t.Parallel()
	for _, exists := range []bool{true, false} {
		backend := &deletionBackend{exists: []bool{exists}}
		useCase, err := NewReadPrivateAccessPolicyRuleUseCase(backend)
		require.NoError(t, err)
		got, err := useCase.Execute(context.Background(),
			ReadPrivateAccessPolicyRuleCommand{AccountID: "account-123", RuleID: "rule-456"})
		require.NoError(t, err)
		require.Equal(t, exists, got)
	}
	backend := &deletionBackend{readError: catoAPIErrorStub{kind: ErrInvalidResponse}}
	useCase, err := NewReadPrivateAccessPolicyRuleUseCase(backend)
	require.NoError(t, err)
	_, err = useCase.Execute(context.Background(), ReadPrivateAccessPolicyRuleCommand{})
	require.Error(t, err)
	require.Empty(t, backend.calls)
	_, err = useCase.Execute(context.Background(),
		ReadPrivateAccessPolicyRuleCommand{AccountID: "account-123", RuleID: "rule-456"})
	require.ErrorIs(t, err, ErrInvalidResponse)
	var failure *UseCaseError
	require.ErrorAs(t, err, &failure)
}
