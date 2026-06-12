package parse

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestKnownStringPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    types.String
		wantNil  bool
		wantText string
	}{
		"null_returns_nil": {
			input:   types.StringNull(),
			wantNil: true,
		},
		"unknown_returns_nil": {
			input:   types.StringUnknown(),
			wantNil: true,
		},
		"known_value_returns_pointer": {
			input:    types.StringValue("hello"),
			wantText: "hello",
		},
		"known_empty_string_returns_pointer": {
			input:    types.StringValue(""),
			wantText: "",
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := KnownStringPointer(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil pointer, got %q", *got)
				}
				return
			}

			if got == nil {
				t.Fatalf("expected pointer to %q, got nil", tt.wantText)
			}
			if *got != tt.wantText {
				t.Fatalf("expected %q, got %q", tt.wantText, *got)
			}
		})
	}
}
