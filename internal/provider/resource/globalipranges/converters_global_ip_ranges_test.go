package globalipranges

import (
	"context"
	"encoding/json"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

func TestGlobalIPRangeResponseConverters(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"createGlobalIPRangeBulk", "updateGlobalIPRangeBulk"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			for _, description := range []string{`null`, `"office range"`} {
				t.Run(description, func(t *testing.T) {
					t.Parallel()
					payload := []byte(`{"object":{"` + operation + `":{"globalIPRange":[{"id":"range-1","name":"office","ipRange":"192.0.2.0/24","description":` + description + `}]}}}`)
					var ranges []tf.GlobalIPRange
					if operation == "createGlobalIPRangeBulk" {
						var response cato.ObjectCreateGlobalIPRangeBulk
						require.NoError(t, json.Unmarshal(payload, &response))
						ranges = globalIPRangesFromCreateResponse(&response)
					} else {
						var response cato.ObjectUpdateGlobalIPRangeBulk
						require.NoError(t, json.Unmarshal(payload, &response))
						ranges = globalIPRangesFromUpdateResponse(&response)
					}
					require.Len(t, ranges, 1)
					require.Equal(t, "range-1", ranges[0].ID.ValueString())
					require.Equal(t, "office", ranges[0].Name.ValueString())
					require.Equal(t, "192.0.2.0/24", ranges[0].IPRange.ValueString())
					require.Equal(t, description == `null`, ranges[0].Description.IsNull())
					if description != `null` {
						require.Equal(t, "office range", ranges[0].Description.ValueString())
					}
				})
			}
		})
	}
}

func TestGlobalIPRangeObjectConversionDiagnostics(t *testing.T) {
	// A mismatched schema forces ObjectValueFrom to report a conversion error.
	original := tf.GlobalIPRangeTypes
	tf.GlobalIPRangeTypes = map[string]attr.Type{"name": types.BoolType}
	t.Cleanup(func() { tf.GlobalIPRangeTypes = original })
	var diags diag.Diagnostics
	globalIPRangeObjectFromModel(context.Background(), &diags, tf.GlobalIPRange{})
	require.True(t, diags.HasError(), "conversion errors must reach the caller")
}
