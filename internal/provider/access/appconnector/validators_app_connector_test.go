package appconnector

import (
	"context"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAppConnectorTypeValidation(t *testing.T) {
	t.Parallel()
	v := AppConTypeValidator{}
	for _, tc := range []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid", types.StringValue(string(cato_models.AllZtnaAppConnectorType[0])), false},
		{"invalid", types.StringValue("INVALID"), true},
		{"null", types.StringNull(), true},
		{"unknown", types.StringUnknown(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), validator.StringRequest{Path: path.Root("value"), ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}
