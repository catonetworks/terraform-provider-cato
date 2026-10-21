package license

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestLicenseBandwidthValidation(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		value     types.Int64
		wantError bool
	}{
		{"multiple of ten", types.Int64Value(100), false},
		{"fractional increment", types.Int64Value(105), true},
		{"null", types.Int64Null(), false},
		{"unknown", types.Int64Unknown(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.Int64Response{}
			customInt64Validator{}.ValidateInt64(context.Background(), validator.Int64Request{Path: path.Root("bandwidth"), ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError())
			if tc.wantError {
				require.Equal(t, path.Root("bandwidth"), resp.Diagnostics[0].(diag.DiagnosticWithPath).Path())
			}
		})
	}
}
