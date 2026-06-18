package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const policyRevisionConflictMaxAttempts = 8

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
