package validators

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFieldIsExplicit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfgVal   types.String
		stateVal types.String
		want     bool
	}{
		{
			name:     "config null is never explicit",
			cfgVal:   types.StringNull(),
			stateVal: types.StringValue("4456"),
			want:     false,
		},
		{
			name:     "config unknown is never explicit",
			cfgVal:   types.StringUnknown(),
			stateVal: types.StringValue("4456"),
			want:     false,
		},
		{
			name:     "config value with null state is explicit",
			cfgVal:   types.StringValue("4456"),
			stateVal: types.StringNull(),
			want:     true,
		},
		{
			name:     "config value with unknown state is explicit",
			cfgVal:   types.StringValue("4456"),
			stateVal: types.StringUnknown(),
			want:     true,
		},
		{
			name:     "config value equal to state is propagated, not explicit",
			cfgVal:   types.StringValue("4456"),
			stateVal: types.StringValue("4456"),
			want:     false,
		},
		{
			name:     "config value different from state is explicit",
			cfgVal:   types.StringValue("9999"),
			stateVal: types.StringValue("4456"),
			want:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := StringFieldIsExplicit(tt.cfgVal, tt.stateVal); got != tt.want {
				t.Fatalf("StringFieldIsExplicit(%v, %v) = %v, want %v", tt.cfgVal, tt.stateVal, got, tt.want)
			}
		})
	}
}
