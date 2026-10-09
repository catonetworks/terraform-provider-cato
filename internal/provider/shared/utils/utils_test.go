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
