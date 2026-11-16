package networkrange

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

func TestUpdatePreservesExplicitIdentityTranslation(t *testing.T) {
	t.Parallel()
	for _, configured := range []bool{false, true} {
		plan := tf.NetworkRange{
			Subnet:           types.StringValue("192.0.2.0/25"),
			TranslatedSubnet: types.StringValue("192.0.2.0/25"),
			RangeType:        types.StringValue("VLAN"),
		}
		cfg := plan
		if !configured {
			cfg.TranslatedSubnet = types.StringNull()
		}
		var diags diag.Diagnostics
		input := hydrateUpdateNetworkRangeAPI(context.Background(), &catoclient.Service{}, &cfg, &plan, &diags)
		require.False(t, diags.HasError(), "%v", diags)
		if configured {
			require.NotNil(t, input.TranslatedSubnet)
			require.Equal(t, "192.0.2.0/25", *input.TranslatedSubnet)
		} else {
			require.Nil(t, input.TranslatedSubnet)
		}
	}
}
