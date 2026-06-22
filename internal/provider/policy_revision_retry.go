package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const policyRevisionConflictMaxAttempts = 8

type (
	internetFirewallDiscardError = cato_go_sdk.PolicyInternetFirewallDiscardPolicyRevision_Policy_InternetFirewall_DiscardPolicyRevision_Errors
	wanFirewallDiscardError      = cato_go_sdk.PolicyWanFirewallDiscardPolicyRevision_Policy_WanFirewall_DiscardPolicyRevision_Errors
	wanNetworkDiscardError       = cato_go_sdk.PolicyWanNetworkDiscardPolicyRevision_Policy_WanNetwork_DiscardPolicyRevision_Errors
)

func policyErrLooksLikeConcurrentRevisionBlock(msg string) bool {
	if msg == "" {
		return false
	}
	return strings.Contains(msg, "reorderPolicyBlockedByActiveSessions") ||
		strings.Contains(msg, "Cannot reorder policy while other active revisions exist")
}

func sleepPolicyRevisionRetry(ctx context.Context, failedAttempt int) error {
	if failedAttempt < 1 {
		return nil
	}
	d := time.Duration(5*failedAttempt) * time.Second
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// withPolicyRevisionConflictRetry re-runs fn when the API reports concurrent policy revisions.
func withPolicyRevisionConflictRetry(ctx context.Context, opLabel string, fn func() error) error {
	var last error
	for attempt := 1; attempt <= policyRevisionConflictMaxAttempts; attempt++ {
		last = fn()
		if last == nil {
			return nil
		}
		if !policyErrLooksLikeConcurrentRevisionBlock(last.Error()) {
			return last
		}
		if attempt == policyRevisionConflictMaxAttempts {
			break
		}
		tflog.Warn(ctx, "policy mutation blocked by concurrent revisions; retrying", map[string]any{
			"operation": opLabel,
			"attempt":   attempt,
			"max":       policyRevisionConflictMaxAttempts,
		})
		if err := sleepPolicyRevisionRetry(ctx, attempt); err != nil {
			return err
		}
	}
	return last
}

// withAcctestPolicyRevisionCleanupRetryOnce discards policy drafts and retries once
// when WAN firewall reorder is blocked by active revisions during acceptance tests.
// Outside acceptance tests it preserves the normal retry-only behavior.
func withAcctestPolicyRevisionCleanupRetryOnce(ctx context.Context, opLabel string, cleanup func() error, fn func() error) error {
	if os.Getenv("TF_ACC") != "1" {
		return withPolicyRevisionConflictRetry(ctx, opLabel, fn)
	}

	err := fn()
	if err == nil || !policyErrLooksLikeConcurrentRevisionBlock(err.Error()) {
		return err
	}

	tflog.Warn(ctx, "policy mutation blocked by concurrent revisions; discarding firewall revisions before one retry", map[string]any{
		"operation": opLabel,
	})
	if cleanupErr := cleanup(); cleanupErr != nil {
		return fmt.Errorf("%s cleanup after concurrent policy revision block: %w; original error: %v", opLabel, cleanupErr, err)
	}

	return fn()
}

func discardFirewallAndWANPolicyRevisions(ctx context.Context, client *cato_go_sdk.Client, accountID string) error {
	var discardErr error

	policy, err := client.Policy(ctx, nil, nil, accountID)
	if err != nil {
		return fmt.Errorf("list firewall revisions: %w", err)
	}

	for _, rev := range policy.GetPolicy().GetInternetFirewall().GetRevisionsInternetFirewallPolicyQueries().GetRevision() {
		if rev.GetID() == "" {
			continue
		}
		if err := discardInternetFirewallPolicyRevision(ctx, client, accountID, rev.GetID()); err != nil {
			discardErr = errors.Join(discardErr, err)
		}
	}
	for _, rev := range policy.GetPolicy().GetWanFirewall().GetRevisionsWanFirewallPolicyQueries().GetRevision() {
		if rev.GetID() == "" {
			continue
		}
		if err := discardWanFirewallPolicyRevision(ctx, client, accountID, rev.GetID()); err != nil {
			discardErr = errors.Join(discardErr, err)
		}
	}

	wanNetwork, err := client.WanNetworkPolicy(ctx, accountID)
	if err != nil {
		discardErr = errors.Join(discardErr, fmt.Errorf("list WAN network revision: %w", err))
	} else if rev := wanNetwork.GetPolicy().GetWanNetwork().GetPolicy().GetRevision(); rev != nil && rev.GetID() != "" {
		if err := discardWanNetworkPolicyRevision(ctx, client, accountID, rev.GetID()); err != nil {
			discardErr = errors.Join(discardErr, err)
		}
	}

	return discardErr
}

func discardInternetFirewallPolicyRevision(ctx context.Context, client *cato_go_sdk.Client, accountID string, revisionID string) error {
	resp, err := client.PolicyInternetFirewallDiscardPolicyRevision(ctx,
		&cato_models.InternetFirewallPolicyMutationInput{
			Revision: &cato_models.PolicyMutationRevisionInput{ID: &revisionID},
		},
		&cato_models.PolicyDiscardRevisionInput{ID: &revisionID},
		accountID,
	)
	if err != nil {
		return fmt.Errorf("discard internet firewall revision %s: %w", revisionID, err)
	}
	errs := resp.GetPolicy().GetInternetFirewall().GetDiscardPolicyRevision().GetErrors()
	if len(errs) > 0 {
		return fmt.Errorf("discard internet firewall revision %s: %s", revisionID, formatInternetFirewallDiscardErrors(errs))
	}
	tflog.Info(ctx, "discarded internet firewall policy revision", map[string]any{"revision_id": revisionID})
	return nil
}

func discardWanFirewallPolicyRevision(ctx context.Context, client *cato_go_sdk.Client, accountID string, revisionID string) error {
	resp, err := client.PolicyWanFirewallDiscardPolicyRevision(ctx,
		&cato_models.PolicyDiscardRevisionInput{ID: &revisionID},
		accountID,
	)
	if err != nil {
		return fmt.Errorf("discard WAN firewall revision %s: %w", revisionID, err)
	}
	errs := resp.GetPolicy().GetWanFirewall().GetDiscardPolicyRevision().GetErrors()
	if len(errs) > 0 {
		return fmt.Errorf("discard WAN firewall revision %s: %s", revisionID, formatWanFirewallDiscardErrors(errs))
	}
	tflog.Info(ctx, "discarded WAN firewall policy revision", map[string]any{"revision_id": revisionID})
	return nil
}

func discardWanNetworkPolicyRevision(ctx context.Context, client *cato_go_sdk.Client, accountID string, revisionID string) error {
	resp, err := client.PolicyWanNetworkDiscardPolicyRevision(ctx, accountID)
	if err != nil {
		return fmt.Errorf("discard WAN network revision %s: %w", revisionID, err)
	}
	errs := resp.GetPolicy().GetWanNetwork().GetDiscardPolicyRevision().GetErrors()
	if len(errs) > 0 {
		return fmt.Errorf("discard WAN network revision %s: %s", revisionID, formatWanNetworkDiscardErrors(errs))
	}
	tflog.Info(ctx, "discarded WAN network policy revision", map[string]any{"revision_id": revisionID})
	return nil
}

func formatInternetFirewallDiscardErrors(errs []*internetFirewallDiscardError) string {
	var b strings.Builder
	for _, e := range errs {
		if e == nil {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("; ")
		}
		_, _ = fmt.Fprintf(&b, "%s: %s", gqlOptionalStr(e.GetErrorCode()), gqlOptionalStr(e.GetErrorMessage()))
	}
	return b.String()
}

func formatWanFirewallDiscardErrors(errs []*wanFirewallDiscardError) string {
	var b strings.Builder
	for _, e := range errs {
		if e == nil {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("; ")
		}
		_, _ = fmt.Fprintf(&b, "%s: %s", gqlOptionalStr(e.GetErrorCode()), gqlOptionalStr(e.GetErrorMessage()))
	}
	return b.String()
}

func formatWanNetworkDiscardErrors(errs []*wanNetworkDiscardError) string {
	var b strings.Builder
	for _, e := range errs {
		if e == nil {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("; ")
		}
		_, _ = fmt.Fprintf(&b, "%s: %s", gqlOptionalStr(e.GetErrorCode()), gqlOptionalStr(e.GetErrorMessage()))
	}
	return b.String()
}

func internetFirewallMoveSectionError(resp *cato_go_sdk.PolicyInternetFirewallMoveSection, callErr error) error {
	if callErr != nil {
		return callErr
	}
	if resp == nil || resp.GetPolicy() == nil || resp.GetPolicy().GetInternetFirewall() == nil {
		return nil
	}
	ms := resp.GetPolicy().GetInternetFirewall().GetMoveSection()
	if ms == nil {
		return nil
	}
	if errs := ms.GetErrors(); len(errs) > 0 {
		var b strings.Builder
		for _, e := range errs {
			if e == nil {
				continue
			}
			if b.Len() > 0 {
				b.WriteString("; ")
			}
			_, _ = fmt.Fprintf(&b, "%s: %s", gqlOptionalStr(e.GetErrorCode()), gqlOptionalStr(e.GetErrorMessage()))
		}
		return fmt.Errorf("internet firewall moveSection errors: %s", b.String())
	}
	if st := ms.GetStatus(); st != nil && *st != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("internet firewall moveSection status %q", string(*st))
	}
	return nil
}

func wanFirewallMoveSectionError(resp *cato_go_sdk.PolicyWanFirewallMoveSection, callErr error) error {
	if callErr != nil {
		return callErr
	}
	if resp == nil || resp.GetPolicy() == nil || resp.GetPolicy().GetWanFirewall() == nil {
		return nil
	}
	ms := resp.GetPolicy().GetWanFirewall().GetMoveSection()
	if ms == nil {
		return nil
	}
	if errs := ms.GetErrors(); len(errs) > 0 {
		var b strings.Builder
		for _, e := range errs {
			if e == nil {
				continue
			}
			if b.Len() > 0 {
				b.WriteString("; ")
			}
			_, _ = fmt.Fprintf(&b, "%s: %s", gqlOptionalStr(e.GetErrorCode()), gqlOptionalStr(e.GetErrorMessage()))
		}
		return fmt.Errorf("WAN firewall moveSection errors: %s", b.String())
	}
	if st := ms.GetStatus(); st != nil && *st != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("WAN firewall moveSection status %q", string(*st))
	}
	return nil
}
