package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPrepareIpsecNetworkRangeUpdate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		planRange  string
		stateRange string
		wantUpdate bool
	}{
		"unchanged range skips update": {
			planRange:  "10.0.0.0/24",
			stateRange: "10.0.0.0/24",
			wantUpdate: false,
		},
		"changed range prepares update": {
			planRange:  "10.0.1.0/24",
			stateRange: "10.0.0.0/24",
			wantUpdate: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			input, update := prepareIpsecNetworkRangeUpdate(
				SiteIpsecIkeV2{NativeNetworkRange: types.StringValue(tt.planRange)},
				SiteIpsecIkeV2{NativeNetworkRange: types.StringValue(tt.stateRange)},
			)

			require.Equal(t, tt.wantUpdate, update)
			if !tt.wantUpdate {
				require.Nil(t, input)
				return
			}

			require.NotNil(t, input)
			require.Equal(t, tt.planRange, *input.Subnet)
			require.Nil(t, input.TranslatedSubnet)
		})
	}
}
