package inetindex

import (
	"context"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy_rule"
)

func TestHydrateFirewallSectionPositionState(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		current      types.Object
		wantPosition string
		wantRef      types.String
	}{
		"missing import state defaults position": {
			current:      types.ObjectNull(policyrule.PositionAttrTypes),
			wantPosition: string(cato_models.PolicySectionPositionEnumLastInPolicy),
			wantRef:      types.StringNull(),
		},
		"configured subpolicy reference is preserved": {
			current: types.ObjectValueMust(
				policyrule.PositionAttrTypes,
				map[string]attr.Value{
					"position": types.StringValue(string(cato_models.PolicySectionPositionEnumLastInPolicy)),
					"ref":      types.StringValue("subpolicy-123"),
				},
			),
			wantPosition: string(cato_models.PolicySectionPositionEnumLastInPolicy),
			wantRef:      types.StringValue("subpolicy-123"),
		},
		"relative section position is preserved": {
			current: types.ObjectValueMust(
				policyrule.PositionAttrTypes,
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

			got, diags := convert.DefaultSectionPosition(context.Background(), test.current)
			require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)

			attrs := got.Attributes()
			require.Equal(t, types.StringValue(test.wantPosition), attrs["position"])
			require.Equal(t, test.wantRef, attrs["ref"])
		})
	}
}
