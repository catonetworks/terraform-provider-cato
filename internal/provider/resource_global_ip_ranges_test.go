package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
)

func TestGlobalIPRangesComputePlanDeletesMissingRanges(t *testing.T) {
	t.Parallel()

	state := []tf.GlobalIPRange{
		testGlobalIPRange("range-1", "managed-range", "192.0.2.0/24"),
		testGlobalIPRange("range-2", "existing-range", "192.0.2.1-192.0.2.10"),
	}
	config := []tf.GlobalIPRange{state[0]}

	_, details := (&globalIPRangesResource{}).computePlan(
		context.Background(),
		state,
		config,
		&diag.Diagnostics{},
	)

	require.Equal(t, []tf.GlobalIPRange{state[1]}, details.toDelete)
}

func testGlobalIPRange(id, name, ipRange string) tf.GlobalIPRange {
	return tf.GlobalIPRange{
		Description: types.StringValue("description"),
		ID:          types.StringValue(id),
		IPRange:     types.StringValue(ipRange),
		Name:        types.StringValue(name),
	}
}
