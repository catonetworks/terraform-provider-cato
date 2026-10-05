package utils

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

type testAPIError struct {
	code    *string
	message *string
}

func (e testAPIError) GetErrorCode() *string    { return e.code }
func (e testAPIError) GetErrorMessage() *string { return e.message }

func TestCheckAPIErrors(t *testing.T) {
	t.Parallel()

	t.Run("transport error", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics
		require.True(t, CheckAPIErrors(errors.New("transport failed"), []testAPIError(nil), "mutation failed", &diags))
		require.True(t, diags.HasError())
	})

	t.Run("message-less API error uses code", func(t *testing.T) {
		t.Parallel()

		code := "MutationRejected"
		var diags diag.Diagnostics
		require.True(t, CheckAPIErrors(nil, []testAPIError{{code: &code}}, "mutation failed", &diags))
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), code)
	})

	t.Run("message-less and code-less API error uses fallback", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics
		require.True(t, CheckAPIErrors(nil, []testAPIError{{}}, "mutation failed", &diags))
		require.True(t, diags.HasError())
		require.NotEmpty(t, diags.Errors()[0].Detail())
	})
}

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
