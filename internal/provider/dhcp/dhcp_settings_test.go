package dhcp

import (
	"context"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
)

type stubV2 struct {
	lookup *cato_go_sdk.EntityLookup
	err    error
}

func (s *stubV2) EntityLookupMinimal(_ context.Context, _ string, _ cato_models.EntityType,
	_ *int64, _ *int64, _ *cato_models.EntityInput, _ []*cato_models.SortInput, _ []*cato_models.LookupFilterInput,
	_ ...clientv2.RequestInterceptor,
) (*cato_go_sdk.EntityLookup, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.lookup, nil
}

type stubSdk struct {
	v2        *stubV2
	accountID string
}

func (s *stubSdk) V2() V2EntityLookup { return s.v2 }

func (s *stubSdk) AccountID() string { return s.accountID }

func relayLookupWithGroups(namesToIDs map[string]string) *cato_go_sdk.EntityLookup {
	items := make([]*cato_go_sdk.EntityLookup_EntityLookup_Items, 0, len(namesToIDs))
	for name, id := range namesToIDs {
		n := name
		items = append(items, &cato_go_sdk.EntityLookup_EntityLookup_Items{
			Entity: cato_go_sdk.EntityLookup_EntityLookup_Items_Entity{
				ID:   id,
				Name: &n,
				Type: cato_models.EntityTypeDhcpRelayGroup,
			},
		})
	}
	return &cato_go_sdk.EntityLookup{
		EntityLookup: cato_go_sdk.EntityLookup_EntityLookup{
			Items: items,
		},
	}
}

func TestPrepareDHCPSettingsDhcpRelayRelayGroupID(t *testing.T) {
	ctx := context.Background()

	buildDhcpRelay := func(id, relayName types.String) types.Object {
		t.Helper()
		obj, d := types.ObjectValueFrom(ctx, tf.DhcpSettingsAttrTypes, tf.DhcpSettings{
			DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
			IPRange:               types.StringNull(),
			RelayGroupID:          id,
			RelayGroupName:        relayName,
			DhcpMicrosegmentation: types.BoolNull(),
		})
		require.Empty(t, d)
		return obj
	}

	t.Run("relay_name_resolves_even_when_stale_relay_group_id_present", func(t *testing.T) {
		stub := &stubV2{lookup: relayLookupWithGroups(map[string]string{
			"DC-AD01": "id-dc",
			"NZ-AD01": "id-nz",
		})}
		client := &stubSdk{v2: stub, accountID: "acc"}
		var diags diag.Diagnostics
		dhcpObj := buildDhcpRelay(types.StringValue("id-dc"), types.StringValue("NZ-AD01"))

		out := PrepareDHCPSettings(ctx, client, cato_models.SubnetTypeNative, dhcpObj, &diags)

		require.False(t, diags.HasError())
		require.NotNil(t, out)
		require.NotNil(t, out.RelayGroupID)
		require.Equal(t, "id-nz", *out.RelayGroupID)
	})

	t.Run("relay_group_id_only_when_name_absent", func(t *testing.T) {
		stub := &stubV2{}
		client := &stubSdk{v2: stub, accountID: "acc"}
		var diags diag.Diagnostics
		obj, d := types.ObjectValueFrom(ctx, tf.DhcpSettingsAttrTypes, tf.DhcpSettings{
			DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
			IPRange:               types.StringNull(),
			RelayGroupID:          types.StringValue("id-only"),
			RelayGroupName:        types.StringNull(),
			DhcpMicrosegmentation: types.BoolNull(),
		})
		require.Empty(t, d)

		out := PrepareDHCPSettings(ctx, client, cato_models.SubnetTypeNative, obj, &diags)

		require.False(t, diags.HasError())
		require.NotNil(t, out)
		require.NotNil(t, out.RelayGroupID)
		require.Equal(t, "id-only", *out.RelayGroupID)
	})

	t.Run("relay_group_name_only_calls_lookup", func(t *testing.T) {
		stub := &stubV2{lookup: relayLookupWithGroups(map[string]string{"Single": "id-single"})}
		client := &stubSdk{v2: stub, accountID: "acc"}
		var diags diag.Diagnostics
		obj, d := types.ObjectValueFrom(ctx, tf.DhcpSettingsAttrTypes, tf.DhcpSettings{
			DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
			IPRange:               types.StringNull(),
			RelayGroupID:          types.StringNull(),
			RelayGroupName:        types.StringValue("Single"),
			DhcpMicrosegmentation: types.BoolNull(),
		})
		require.Empty(t, d)

		out := PrepareDHCPSettings(ctx, client, cato_models.SubnetTypeVlan, obj, &diags)

		require.False(t, diags.HasError())
		require.NotNil(t, out)
		require.Equal(t, "id-single", *out.RelayGroupID)
	})

	t.Run("relay_group_requires_name_or_id", func(t *testing.T) {
		stub := &stubV2{lookup: relayLookupWithGroups(map[string]string{})}
		client := &stubSdk{v2: stub, accountID: "acc"}
		var diags diag.Diagnostics
		obj, d := types.ObjectValueFrom(ctx, tf.DhcpSettingsAttrTypes, tf.DhcpSettings{
			DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
			IPRange:               types.StringNull(),
			RelayGroupID:          types.StringNull(),
			RelayGroupName:        types.StringNull(),
			DhcpMicrosegmentation: types.BoolNull(),
		})
		require.Empty(t, d)

		out := PrepareDHCPSettings(ctx, client, cato_models.SubnetTypeNative, obj, &diags)

		require.Nil(t, out)
		require.True(t, diags.HasError())
	})
}
