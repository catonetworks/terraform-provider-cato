package ifrule

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestInternetFirewallActionSchema(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	var schemaResponse resource.SchemaResponse
	NewInternetFwRuleResource().Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	require.False(t, schemaResponse.Diagnostics.HasError(), "%v", schemaResponse.Diagnostics)
	actionPath := path.Root("rule").AtName("action")
	attribute, diags := schemaResponse.Schema.AttributeAtPath(ctx, actionPath)
	require.False(t, diags.HasError(), "%v", diags)
	action := attribute.(schema.StringAttribute)
	require.NotEmpty(t, action.Validators)
	for _, tc := range []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{name: "ALLOW", value: types.StringValue("ALLOW")},
		{name: "BLOCK", value: types.StringValue("BLOCK")},
		{name: "PROMPT", value: types.StringValue("PROMPT")},
		{name: "RBI", value: types.StringValue("RBI")},
		{name: "CAPTIVE_PORTAL", value: types.StringValue("CAPTIVE_PORTAL")},
		{name: "SUB_POLICY", value: types.StringValue("SUB_POLICY"), wantError: true},
		{name: "invalid", value: types.StringValue("INVALID"), wantError: true},
		{name: "empty", value: types.StringValue(""), wantError: true},
		{name: "null", value: types.StringNull()},
		{name: "unknown", value: types.StringUnknown()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var response validator.StringResponse
			for _, v := range action.Validators {
				v.ValidateString(t.Context(), validator.StringRequest{
					Path: actionPath, ConfigValue: tc.value,
				}, &response)
			}
			require.Equal(t, tc.wantError, response.Diagnostics.HasError(), "%v", response.Diagnostics)
		})
	}
}
