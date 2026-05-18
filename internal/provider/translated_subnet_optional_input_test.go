package provider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTranslatedSubnetUsesOptionalPointerHelper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		file                     string
		mustContain             []string
		mustNotContain          []string
	}{
		{
			file: "resource_network_range.go",
			mustContain: []string{
				"TranslatedSubnet: stringPointerForOptionalInput(plan.TranslatedSubnet)",
			},
			mustNotContain: []string{
				"TranslatedSubnet: plan.TranslatedSubnet.ValueStringPointer()",
			},
		},
		{
			file: "resource_lan_interface.go",
			mustContain: []string{
				"TranslatedSubnet: stringPointerForOptionalInput(plan.TranslatedSubnet)",
			},
			mustNotContain: []string{
				"TranslatedSubnet: plan.TranslatedSubnet.ValueStringPointer()",
			},
		},
		{
			file: "resource_site_socket.go",
			mustContain: []string{
				"input.TranslatedSubnet = stringPointerForOptionalInput(nativeRangeInput.TranslatedSubnet)",
				"inputUpdateNetworkRange.TranslatedSubnet = stringPointerForOptionalInput(nativeRangeInput.TranslatedSubnet)",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.file, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(tt.file)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.file, err)
			}
			source := string(content)

			for _, fragment := range tt.mustContain {
				if !strings.Contains(source, fragment) {
					t.Fatalf("expected %s to contain %q", tt.file, fragment)
				}
			}

			for _, fragment := range tt.mustNotContain {
				if strings.Contains(source, fragment) {
					t.Fatalf("expected %s to not contain %q", tt.file, fragment)
				}
			}
		})
	}
}
