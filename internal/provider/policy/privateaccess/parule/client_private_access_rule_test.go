package parule

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLookupPrivateAccessRule(t *testing.T) {
	for _, wait := range []bool{false, true} {
		t.Run(map[bool]string{false: "refresh missing", true: "create read lag"}[wait], func(t *testing.T) {
			calls := 0
			r := newPrivateAccessRuleTestResource(t, make(chan string, 3), func(w http.ResponseWriter, _ string) {
				calls++
				if calls == 1 {
					_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"policy":{"rules":[]}}}}}`))
				} else {
					_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"policy":{"rules":[{"rule":{"id":"rule-123"}}]}}}}}`))
				}
			})
			rule, err := lookupPrivateAccessRule(context.Background(), r.resourceAPI(), "account-123", "rule-123", wait)
			if wait {
				require.NoError(t, err)
				require.Equal(t, "rule-123", rule.Rule.ID)
				require.Equal(t, 2, calls)
			} else {
				require.ErrorIs(t, err, ErrPrivateAcccessRuleNotFound)
				require.Nil(t, rule)
				require.Equal(t, 1, calls)
			}
		})
	}
}

func TestLookupPrivateAccessRuleCancellation(t *testing.T) {
	r := newPrivateAccessRuleTestResource(t, make(chan string, 1), func(w http.ResponseWriter, _ string) {
		_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"policy":{"rules":[]}}}}}`))
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := lookupPrivateAccessRule(ctx, r.resourceAPI(), "account-123", "rule-123", true)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestLookupPrivateAccessRuleErrors(t *testing.T) {
	for _, response := range []string{`{"data":{"policy":{"privateAccess":null}}}`, `{"data":{"policy":{"privateAccess":{"policy":{"rules":[null]}}}}}`, `{"errors":[{"message":"permission denied"}]}`} {
		t.Run(response, func(t *testing.T) {
			calls := 0
			r := newPrivateAccessRuleTestResource(t, make(chan string, 1), func(w http.ResponseWriter, _ string) { calls++; _, _ = w.Write([]byte(response)) })
			_, err := lookupPrivateAccessRule(context.Background(), r.resourceAPI(), "account-123", "rule-123", true)
			require.Error(t, err)
			require.Equal(t, 1, calls)
		})
	}
}
