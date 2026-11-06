package bgppeer

import (
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPreserveOmittedBgpPassword(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name        string
		remote      *string
		prior, want types.String
	}{
		{"omitted empty API default", stringPointer(""), types.StringNull(), types.StringNull()},
		{"explicit empty", stringPointer(""), types.StringValue(""), types.StringValue("")},
		{"nonempty remote", stringPointer("secret"), types.StringNull(), types.StringValue("secret")},
		{"remote cleared", stringPointer(""), types.StringValue("old"), types.StringValue("")},
		{"remote null", nil, types.StringNull(), types.StringNull()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := ConvertToBgpPeer(cato_go_sdk.Site_Site_BgpPeer{Md5AuthKey: tc.remote})
			preserveOmittedBgpPassword(&model, BgpPeer{Md5AuthKey: tc.prior})
			require.Equal(t, tc.want, model.Md5AuthKey)
		})
	}
}
func stringPointer(value string) *string { return &value }
