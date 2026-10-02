package utils

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr"
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
		require.True(t, apperr.CheckAPIErrors(errors.New("transport failed"), []testAPIError(nil), "mutation failed", &diags))
		require.True(t, diags.HasError())
	})

	t.Run("message-less API error uses code", func(t *testing.T) {
		t.Parallel()

		code := "MutationRejected"
		var diags diag.Diagnostics
		require.True(t, apperr.CheckAPIErrors(nil, []testAPIError{{code: &code}}, "mutation failed", &diags))
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), code)
	})

	t.Run("message-less and code-less API error uses fallback", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics
		require.True(t, apperr.CheckAPIErrors(nil, []testAPIError{{}}, "mutation failed", &diags))
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

func TestNormalizeDateTime(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no fractional seconds",
			in:   "2024-02-03T04:05:06",
			want: "2024-02-03T04:05:06Z",
		},
		{
			name: "with fractional seconds",
			in:   "2024-02-03T04:05:06.987654",
			want: "2024-02-03T04:05:06Z",
		},
		{
			name: "with timezone suffix still normalizes prefix",
			in:   "2024-02-03T04:05:06+02:00",
			want: "2024-02-03T04:05:06Z",
		},
		{
			name: "non matching text returns unchanged",
			in:   "not-a-datetime",
			want: "not-a-datetime",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeDateTime(tc.in); got != tc.want {
				t.Fatalf("unexpected normalized value\nwant: %q\ngot:  %q", tc.want, got)
			}
		})
	}
}

func TestNormalizeDateTimePtr(t *testing.T) {
	t.Run("nil stays nil", func(t *testing.T) {
		if got := NormalizeDateTimePtr(nil); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("value is normalized and returned", func(t *testing.T) {
		in := "2024-02-03T04:05:06.123"
		got := NormalizeDateTimePtr(&in)
		if got == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *got != "2024-02-03T04:05:06Z" {
			t.Fatalf("unexpected normalized pointer value: %q", *got)
		}
	})
}
