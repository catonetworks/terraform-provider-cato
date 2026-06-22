package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolicyErrLooksLikeConcurrentRevisionBlock(t *testing.T) {
	t.Parallel()
	require.True(t, policyErrLooksLikeConcurrentRevisionBlock("reorderPolicyBlockedByActiveSessions: x"))
	require.True(t, policyErrLooksLikeConcurrentRevisionBlock("Cannot reorder policy while other active revisions exist"))
	require.False(t, policyErrLooksLikeConcurrentRevisionBlock(""))
	require.False(t, policyErrLooksLikeConcurrentRevisionBlock("some other API failure"))
}
