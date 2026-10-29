package dnssettings

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestDNSSettingsValidation(t *testing.T) {
	for _, tc := range []struct {
		name               string
		primary, secondary types.String
		invalid            bool
	}{
		{"secondary without primary", types.StringNull(), types.StringValue("192.0.2.53"), true},
		{"primary", types.StringValue("192.0.2.53"), types.StringNull(), false},
		{"unknown primary", types.StringUnknown(), types.StringValue("192.0.2.53"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := NewDNSSettingsResource()
			schema := &resource.SchemaResponse{}
			res.Schema(context.Background(), resource.SchemaRequest{}, schema)
			plan := tfsdk.Plan{Schema: schema.Schema}
			model := SettingsModel{ID: types.StringUnknown(), PrimaryServer: tc.primary, SecondaryServer: tc.secondary, Suffix: types.ListValueMust(types.StringType, nil), AcceptDNSRequestsOnLanInterfaceIP: types.BoolValue(false)}
			require.False(t, plan.Set(context.Background(), model).HasError())
			validating, ok := res.(resource.ResourceWithValidateConfig)
			require.True(t, ok, "DNS cross-field constraints must be checked during plan")
			resp := &resource.ValidateConfigResponse{}
			validating.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: tfsdk.Config{Raw: plan.Raw, Schema: schema.Schema}}, resp)
			require.Equal(t, tc.invalid, resp.Diagnostics.HasError(), resp.Diagnostics)
		})
	}
}
