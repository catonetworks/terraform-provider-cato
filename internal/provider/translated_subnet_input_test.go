package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTranslatedSubnetForAPIInput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		cfg, plan types.String
		wantNil   bool
		want      string
	}{
		"config_null_omits_even_when_plan_has_api_value": {
			cfg:     types.StringNull(),
			plan:    types.StringValue("10.12.62.0/24"),
			wantNil: true,
		},
		"config_unknown_omits": {
			cfg:     types.StringUnknown(),
			plan:    types.StringValue("10.12.62.0/24"),
			wantNil: true,
		},
		"config_present_empty_plan_empty": {
			cfg:     types.StringValue(""),
			plan:    types.StringValue(""),
			wantNil: true,
		},
		"config_present_forwards_non_empty_plan": {
			cfg:     types.StringValue("10.0.0.0/24"),
			plan:    types.StringValue("10.1.0.0/24"),
			wantNil: false,
			want:    "10.1.0.0/24",
		},
		"config_present_plan_empty": {
			cfg:     types.StringValue("10.0.0.0/24"),
			plan:    types.StringValue(""),
			wantNil: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := translatedSubnetForAPIInput(tt.cfg, tt.plan)
			assertTranslatedSubnetPointer(t, got, tt.wantNil, tt.want)
		})
	}
}
