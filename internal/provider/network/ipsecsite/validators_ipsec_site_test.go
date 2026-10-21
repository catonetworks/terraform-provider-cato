package ipsecsite

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestIPSecIdentificationTypeValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, tc := range []struct {
		name                 string
		identification, mode types.String
		wantError            bool
	}{
		{"responder", types.StringValue("FQDN"), types.StringValue("RESPONDER_ONLY"), false},
		{"initiator", types.StringValue("FQDN"), types.StringValue("INITIATOR_ONLY"), true},
		{"missing mode", types.StringValue("FQDN"), types.StringNull(), true},
		{"unknown mode", types.StringValue("FQDN"), types.StringUnknown(), true},
		{"null identification", types.StringNull(), types.StringValue("INITIATOR_ONLY"), false},
		{"unknown identification", types.StringUnknown(), types.StringValue("INITIATOR_ONLY"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan := tfsdk.Plan{Schema: schema.Schema{Attributes: map[string]schema.Attribute{
				"ipsec": schema.SingleNestedAttribute{Optional: true, Attributes: map[string]schema.Attribute{
					"connection_mode": schema.StringAttribute{Optional: true},
				}},
			}}}
			diags := plan.Set(ctx, struct {
				IPsec types.Object `tfsdk:"ipsec"`
			}{IPsec: types.ObjectValueMust(map[string]attr.Type{"connection_mode": types.StringType}, map[string]attr.Value{"connection_mode": tc.mode})})
			require.False(t, diags.HasError(), "%v", diags)
			resp := &planmodifier.StringResponse{PlanValue: tc.identification}
			IdentificationTypeValidator().PlanModifyString(ctx, planmodifier.StringRequest{Plan: plan, PlanValue: tc.identification, Path: path.Root("ipsec").AtName("identification_type")}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			require.Equal(t, tc.identification, resp.PlanValue)
		})
	}
}
