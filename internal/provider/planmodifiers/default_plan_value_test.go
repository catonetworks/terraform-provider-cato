package planmodifiers

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestDefaultPlanValue(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name          string
		config, state types.String
		hasState      bool
		want          types.String
	}{
		{"configured value wins", types.StringValue("configured"), types.StringValue("previous"), true, types.StringValue("configured")},
		{"retain state", types.StringNull(), types.StringValue("previous"), true, types.StringValue("previous")},
		{"unknown config uses state", types.StringUnknown(), types.StringValue("previous"), true, types.StringValue("previous")},
		{"no state", types.StringNull(), types.StringValue("previous"), false, types.StringUnknown()},
		{"null state", types.StringNull(), types.StringNull(), true, types.StringUnknown()},
		{"unknown state", types.StringNull(), types.StringUnknown(), true, types.StringUnknown()},
	} {
		t.Run(tc.name, func(t *testing.T) { require.Equal(t, tc.want, DefaultPlanValue(tc.config, tc.state, tc.hasState)) })
	}
}
