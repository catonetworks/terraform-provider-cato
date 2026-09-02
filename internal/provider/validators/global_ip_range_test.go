package validators

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	schemaValidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
)

func TestGlobalIPRangeValidatorAcceptsSupportedFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ipRange  string
		wantErr  bool
		errorMsg string
	}{
		{name: "single IPv4 address", ipRange: "192.0.2.10"},
		{name: "IPv4 CIDR block", ipRange: "192.0.2.0/24"},
		{name: "single IPv6 address", ipRange: "2001:db8::10"},
		{name: "IPv6 CIDR block", ipRange: "2001:db8::/32"},
		{name: "IPv4 range", ipRange: "172.20.102.165-172.20.102.183"},
		{name: "range with whitespace", ipRange: "172.20.102.165 - 172.20.102.183"},
		{name: "IPv6 range", ipRange: "2001:db8::10-2001:db8::20"},
		{
			name:     "invalid range",
			ipRange:  "172.20.102.183-172.20.102.165",
			wantErr:  true,
			errorMsg: "must be a valid IP address",
		},
		{
			name:     "invalid IP address",
			ipRange:  "172.20.102.999",
			wantErr:  true,
			errorMsg: "must be a valid IP address",
		},
		{
			name:     "mixed IP versions",
			ipRange:  "172.20.102.165-2001:db8::20",
			wantErr:  true,
			errorMsg: "must be a valid IP address",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			GetGlobalIPRangeValidator().ValidateGlobalIPRange([]tf.GlobalIPRange{{
				IPRange: types.StringValue(test.ipRange),
				Name:    types.StringValue("test-range"),
			}}, &diags)

			if test.wantErr {
				if !diags.HasError() {
					t.Fatalf("expected validation error for %q", test.ipRange)
				}
				if !strings.Contains(diags[0].Detail(), test.errorMsg) {
					t.Fatalf("expected diagnostic containing %q, got %q", test.errorMsg, diags[0].Detail())
				}
				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected validation error for %q: %v", test.ipRange, diags)
			}
		})
	}
}

func TestGlobalIPRangeValidatorRejectsEmptyAndDuplicateValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ranges  []tf.GlobalIPRange
		wantErr string
	}{
		{
			name: "empty range",
			ranges: []tf.GlobalIPRange{{
				IPRange: types.StringValue(""),
				Name:    types.StringValue("test-range"),
			}},
			wantErr: "ip_range cannot be empty",
		},
		{
			name: "duplicate names",
			ranges: []tf.GlobalIPRange{
				{IPRange: types.StringValue("192.0.2.1"), Name: types.StringValue("test-range")},
				{IPRange: types.StringValue("192.0.2.2"), Name: types.StringValue("test-range")},
			},
			wantErr: "duplicate ip range name",
		},
		{
			name: "duplicate values",
			ranges: []tf.GlobalIPRange{
				{IPRange: types.StringValue("192.0.2.1"), Name: types.StringValue("test-range-1")},
				{IPRange: types.StringValue("192.0.2.1"), Name: types.StringValue("test-range-2")},
			},
			wantErr: "duplicate ip range",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			GetGlobalIPRangeValidator().ValidateGlobalIPRange(test.ranges, &diags)
			if !diags.HasError() {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(diags[0].Detail(), test.wantErr) {
				t.Fatalf("expected diagnostic containing %q, got %q", test.wantErr, diags[0].Detail())
			}
		})
	}
}

func TestGlobalIPRangeValidatorSkipsNullAndUnknownSets(t *testing.T) {
	t.Parallel()

	globalIPRangeValidator := GetGlobalIPRangeValidator()
	for _, value := range []types.Set{types.SetNull(types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes}), types.SetUnknown(types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes})} {
		var resp schemaValidator.SetResponse
		globalIPRangeValidator.ValidateSet(
			context.Background(),
			schemaValidator.SetRequest{ConfigValue: value},
			&resp,
		)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected validation error: %v", resp.Diagnostics)
		}
	}
}
