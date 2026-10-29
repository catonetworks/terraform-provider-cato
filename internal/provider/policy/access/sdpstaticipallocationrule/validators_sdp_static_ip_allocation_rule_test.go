package sdpstaticipallocationrule

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
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

func TestSourceReferenceAndRequiredValues(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			model := testModel(t)
			var diagnostics diag.Diagnostics
			var encoded []byte
			var err error
			if operation == "create" {
				encoded, err = json.Marshal(buildCreateObject(ctx, model.Rule, &diagnostics))
			} else {
				encoded, err = json.Marshal(buildUpdateObject(ctx, model.Rule, &diagnostics))
			}
			require.NoError(t, err)
			require.False(t, diagnostics.HasError(), "%v", diagnostics)
			var payload map[string]any
			require.NoError(t, json.Unmarshal(encoded, &payload))
			require.Equal(t, "192.0.2.241", payload["ip"])
			require.Equal(t, map[string]any{"user": map[string]any{"by": "ID", "input": "item-1"}}, payload["source"])
		})
	}
	for _, field := range []string{"description", "enabled", "ip", "name", "source"} {
		for _, unknown := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s_unknown_%t", field, unknown), func(t *testing.T) {
				t.Parallel()
				fields := testModel(t).Rule.Attributes()
				typ := ruleAttrTypes[field]
				native := tftypes.NewValue(typ.TerraformType(ctx), nil)
				if unknown {
					native = tftypes.NewValue(typ.TerraformType(ctx), tftypes.UnknownValue)
				}
				value, err := typ.ValueFromTerraform(ctx, native)
				require.NoError(t, err)
				fields[field] = value
				var diagnostics diag.Diagnostics
				buildCreateObject(ctx, types.ObjectValueMust(ruleAttrTypes, fields), &diagnostics)
				require.True(t, diagnostics.HasError())
			})
		}
	}
}

func TestSourceRequiresKnownUser(t *testing.T) {
	t.Parallel()
	for _, user := range []types.Object{types.ObjectNull(idname.RefAttrTypes), types.ObjectUnknown(idname.RefAttrTypes)} {
		source := types.ObjectValueMust(ruleSourceAttrTypes, map[string]attr.Value{"user": user})
		for _, update := range []bool{false, true} {
			var diagnostics diag.Diagnostics
			if update {
				buildUpdateObjectSource(context.Background(), source, &diagnostics)
			} else {
				buildCreateObjectSource(context.Background(), source, &diagnostics)
			}
			require.True(t, diagnostics.HasError())
		}
	}
}
