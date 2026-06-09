package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHydrateLanInterfaceAPITranslatedSubnet(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		translatedSubnet types.String
		wantNil          bool
		wantValue        string
	}{
		"null_omitted": {
			translatedSubnet: types.StringNull(),
			wantNil:          true,
		},
		"empty_omitted": {
			translatedSubnet: types.StringValue(""),
			wantNil:          true,
		},
		"value_set": {
			translatedSubnet: types.StringValue("172.16.30.0/24"),
			wantValue:        "172.16.30.0/24",
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := LanInterface{
				Name:             types.StringValue("lan-1"),
				DestType:         types.StringValue("LAN"),
				LocalIP:          types.StringValue("10.30.30.1"),
				Subnet:           types.StringValue("10.30.30.0/24"),
				TranslatedSubnet: tt.translatedSubnet,
				VrrpType:         types.StringNull(),
			}

			input := hydrateLanInterfaceAPI(context.Background(), plan)
			if input.Lan == nil {
				t.Fatal("expected LAN input")
			}
			assertTranslatedSubnetPointer(t, input.Lan.TranslatedSubnet, tt.wantNil, tt.wantValue)
		})
	}
}

func assertTranslatedSubnetPointer(t *testing.T, got *string, wantNil bool, wantValue string) {
	t.Helper()

	if wantNil {
		if got != nil {
			t.Fatalf("expected translated subnet to be nil, got %q", *got)
		}
		return
	}

	if got == nil {
		t.Fatalf("expected translated subnet %q, got nil", wantValue)
	}
	if *got != wantValue {
		t.Fatalf("expected translated subnet %q, got %q", wantValue, *got)
	}
}
