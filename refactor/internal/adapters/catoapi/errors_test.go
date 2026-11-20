package catoapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
)

func TestCatoAPIErrorExposesContractAndPreservesCause(t *testing.T) {
	t.Parallel()
	cause := fmt.Errorf("sensitive backend details: %w", context.Canceled)
	failure := &CatoAPIError{
		Operation: "remove private-access policy rule",
		Kind:      network.ErrRejected,
		Codes:     []string{"RuleNotFound"},
		Cause:     cause,
	}
	var contract network.CatoAPIAdapterError = failure
	require.Equal(t, network.ErrRejected, contract.ErrorKind())
	require.Equal(t, []string{"RuleNotFound"}, contract.ErrorCodes())
	require.ErrorIs(t, failure, network.ErrRejected)
	require.ErrorIs(t, failure, context.Canceled)
	require.ErrorIs(t, failure, cause)
	require.NotErrorIs(t, failure, network.ErrRuleNotFound)
	require.NotContains(t, contract.Error(), "sensitive")
}
