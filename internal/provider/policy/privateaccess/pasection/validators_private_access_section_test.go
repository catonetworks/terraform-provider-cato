package pasection

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPositionOmitsEmptyReference(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	positionValue := types.ObjectValueMust(atAttrTypes, map[string]attr.Value{
		"position": types.StringValue("LAST_IN_POLICY"), "ref": types.StringValue(""),
	})
	var diagnostics diag.Diagnostics
	position := buildPosition(ctx, positionValue, &diagnostics)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	require.NotNil(t, position)
	require.Nil(t, position.Ref)
}
