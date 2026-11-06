package socketlanfirewallrule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSocketLanFirewallRulePositionValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var response resource.SchemaResponse
	(&socketLanFirewallRuleResource{}).Schema(ctx, resource.SchemaRequest{}, &response)
	at := response.Schema.Attributes["at"].(schema.SingleNestedAttribute)
	position := at.Attributes["position"].(schema.StringAttribute)

	for _, tc := range []struct {
		value     types.String
		wantError bool
	}{
		{types.StringValue("AFTER_SUB_RULE"), false},
		{types.StringValue("BEFORE_SUB_RULE"), false},
		{types.StringValue("FIRST_IN_RULE"), false},
		{types.StringValue("LAST_IN_RULE"), false},
		{types.StringValue("AFTER_RULE"), true},
		{types.StringValue("BEFORE_RULE"), true},
		{types.StringValue("INVALID"), true},
		{types.StringNull(), false},
		{types.StringUnknown(), false},
	} {
		t.Run(tc.value.String(), func(t *testing.T) {
			var result validator.StringResponse
			for _, v := range position.Validators {
				v.ValidateString(ctx, validator.StringRequest{
					Path: path.Root("at").AtName("position"), ConfigValue: tc.value,
				}, &result)
			}
			require.Equal(t, tc.wantError, result.Diagnostics.HasError(), "%v", result.Diagnostics)
		})
	}
}
