package tlsrule

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

func TestTLSRuleEnumValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var response resource.SchemaResponse
	(&tlsInspectionRuleResource{}).Schema(ctx, resource.SchemaRequest{}, &response)
	rule := response.Schema.Attributes["rule"].(schema.SingleNestedAttribute)
	application := rule.Attributes["application"].(schema.SingleNestedAttribute)
	service := application.Attributes["custom_service"].(schema.SingleNestedAttribute)

	for _, tc := range []struct {
		attribute schema.StringAttribute
		name      string
		value     string
		wantError bool
	}{
		{application.Attributes["tls_inspect_category"].(schema.StringAttribute), "category", "SAFE_TO_INSPECT_DOMAINS", false},
		{application.Attributes["tls_inspect_category"].(schema.StringAttribute), "category", "POPULAR_CLOUD_APPS", false},
		{application.Attributes["tls_inspect_category"].(schema.StringAttribute), "category", "STREAMING_MEDIA", true},
		{rule.Attributes["platform"].(schema.StringAttribute), "platform", "CHROMEOS", false},
		{rule.Attributes["platform"].(schema.StringAttribute), "platform", "INVALID", true},
		{service.Attributes["protocol"].(schema.StringAttribute), "protocol", "TCP_UDP", false},
		{service.Attributes["protocol"].(schema.StringAttribute), "protocol", "INVALID", true},
	} {
		t.Run(tc.name+"/"+tc.value, func(t *testing.T) {
			var result validator.StringResponse
			for _, v := range tc.attribute.Validators {
				v.ValidateString(ctx, validator.StringRequest{
					Path: path.Root(tc.name), ConfigValue: types.StringValue(tc.value),
				}, &result)
			}
			require.Equal(t, tc.wantError, result.Diagnostics.HasError(), "%v", result.Diagnostics)
		})
	}
}
