package rbacassignment

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAssignmentActionSchemaUsesRBACAction(t *testing.T) {
	t.Parallel()
	action := entityAttributes()["action"].(schema.StringAttribute)
	for _, tc := range []struct {
		value     string
		wantError bool
	}{
		{"VIEW", false},
		{"EDIT", false},
		{"NONE", true},
		{"INVALID", true},
	} {
		t.Run(tc.value, func(t *testing.T) {
			var response validator.StringResponse
			for _, v := range action.Validators {
				v.ValidateString(context.Background(), validator.StringRequest{
					ConfigValue: types.StringValue(tc.value),
				}, &response)
			}
			require.Equal(t, tc.wantError, response.Diagnostics.HasError(), "%v", response.Diagnostics)
		})
	}
}
