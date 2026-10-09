package convert

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestStringPointerValue(t *testing.T) {
	t.Parallel()

	empty := ""
	description := "site description"

	testCases := map[string]struct {
		apiValue *string
		plan     types.String
		want     types.String
	}{
		"nil API value preserves null plan": {
			plan: types.StringNull(),
			want: types.StringNull(),
		},
		"empty API value preserves null plan": {
			apiValue: &empty,
			plan:     types.StringNull(),
			want:     types.StringNull(),
		},
		"nil API value becomes empty string for configured value": {
			plan: types.StringValue(description),
			want: types.StringValue(""),
		},
		"empty API value becomes empty string for configured value": {
			apiValue: &empty,
			plan:     types.StringValue(description),
			want:     types.StringValue(""),
		},
		"nil API value becomes empty string for unknown plan": {
			plan: types.StringUnknown(),
			want: types.StringValue(""),
		},
		"non-empty API value is preserved": {
			apiValue: &description,
			plan:     types.StringNull(),
			want:     types.StringValue(description),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.want, StringPointerValue(testCase.apiValue, testCase.plan))
		})
	}
}
