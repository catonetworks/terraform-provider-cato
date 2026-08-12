package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestHydrateFirewallSectionPositionState(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		current      types.Object
		wantPosition string
		wantRef      types.String
	}{
		"missing import state defaults position": {
			current:      types.ObjectNull(PositionAttrTypes),
			wantPosition: ifwLastInPolicyPosition,
			wantRef:      types.StringNull(),
		},
		"configured subpolicy reference is preserved": {
			current: types.ObjectValueMust(
				PositionAttrTypes,
				map[string]attr.Value{
					"position": types.StringValue(ifwLastInPolicyPosition),
					"ref":      types.StringValue("subpolicy-123"),
				},
			),
			wantPosition: ifwLastInPolicyPosition,
			wantRef:      types.StringValue("subpolicy-123"),
		},
		"relative section position is preserved": {
			current: types.ObjectValueMust(
				PositionAttrTypes,
				map[string]attr.Value{
					"position": types.StringValue("BEFORE_SECTION"),
					"ref":      types.StringValue("section-123"),
				},
			),
			wantPosition: "BEFORE_SECTION",
			wantRef:      types.StringValue("section-123"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := hydrateFirewallSectionPositionState(context.Background(), test.current)
			require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)

			attrs := got.Attributes()
			require.Equal(t, types.StringValue(test.wantPosition), attrs["position"])
			require.Equal(t, test.wantRef, attrs["ref"])
		})
	}
}
