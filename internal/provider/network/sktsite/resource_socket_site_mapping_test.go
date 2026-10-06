package sktsite

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/application"
)

func TestNativeRangeFallbacks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &socketSiteResource{}
	for _, management := range []*string{nil, new("10.0.0.5")} {
		t.Run(func() string {
			if management == nil {
				return "state_ip"
			}
			return "management_ip"
		}(), func(t *testing.T) {
			prior := NativeRange{LocalIP: types.StringValue("10.0.0.1"), LagMinLinks: types.Int64Value(2), DhcpSettings: types.ObjectNull(dhcp.SettingsAttrTypes)}
			state, ds := types.ObjectValueFrom(ctx, SiteNativeRangeResourceAttrTypes, prior)
			require.False(t, ds.HasError())
			var diags diag.Diagnostics
			got := r.parseNativeRange(ctx, nil, &application.Range{NetworkRangeID: "range", RangeType: "Native", PrimaryManagementIP: management}, &application.Interface{Index: new("5"), ID: new("interface")}, state, &diags)
			require.False(t, diags.HasError(), "%v", diags)
			var native NativeRange
			require.False(t, got.As(ctx, &native, basetypes.ObjectAsOptions{}).HasError())
			require.Equal(t, "INT_5", native.InterfaceIndex.ValueString())
			require.Equal(t, int64(2), native.LagMinLinks.ValueInt64())
			want := "10.0.0.1"
			if management != nil {
				want = *management
			}
			require.Equal(t, want, native.LocalIP.ValueString())
			require.True(t, native.DhcpSettings.IsNull())
		})
	}
}
func TestGeneralLocationUnknowns(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	plan := socketLifecycleModel(ctx, t)
	in := (&socketSiteResource{}).prepareGeneralLocation(ctx, plan.SiteLocation, &diags)
	require.Nil(t, in)
	require.False(t, diags.HasError())
}
