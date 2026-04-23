package accmock

import (
	"strings"
	"testing"
)

func TestGetItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		jsonContent string
		itemPath    string
		want        string
		wantErr     string
	}{
		{
			name:        "nested string",
			jsonContent: `{"variables":{"updateConnector":{"id":"1500000250"}}}`,
			itemPath:    "variables.updateConnector.id",
			want:        "1500000250",
		},
		{
			name:        "number value",
			jsonContent: `{"variables":{"count":5}}`,
			itemPath:    "variables.count",
			want:        "5",
		},
		{
			name:        "bool value",
			jsonContent: `{"variables":{"preferredOnly":false}}`,
			itemPath:    "variables.preferredOnly",
			want:        "false",
		},
		{
			name:        "array returns first scalar item",
			jsonContent: `{"variables":{"entityIDs":["506537","511846"]}}`,
			itemPath:    "variables.entityIDs",
			want:        "506537",
		},
		{
			name:        "array returns first object item as json",
			jsonContent: `{"variables":{"items":[{"id":"2003","name":"Amsterdam Sta"},{"id":"9106"}]}}`,
			itemPath:    "variables.items",
			want:        `{"id":"2003","name":"Amsterdam Sta"}`,
		},
		{
			name:        "object returns compact json",
			jsonContent: `{"variables":{"ref":{"by":"ID","input":"1500000250"}}}`,
			itemPath:    "variables.ref",
			want:        `{"by":"ID","input":"1500000250"}`,
		},
		{
			name:        "empty path",
			jsonContent: `{"variables":{"id":"1500000250"}}`,
			itemPath:    "",
			wantErr:     "item path is empty",
		},
		{
			name:        "invalid json",
			jsonContent: `{"variables":`,
			itemPath:    "variables.id",
			wantErr:     "unexpected end of JSON input",
		},
		{
			name:        "missing key",
			jsonContent: `{"variables":{"id":"1500000250"}}`,
			itemPath:    "variables.missing",
			wantErr:     `missing key "missing"`,
		},
		{
			name:        "cannot descend into scalar",
			jsonContent: `{"variables":{"id":"1500000250"}}`,
			itemPath:    "variables.id.value",
			wantErr:     `cannot descend into string at "value"`,
		},
		{
			name:        "null value",
			jsonContent: `{"variables":{"primary":null}}`,
			itemPath:    "variables.primary",
			wantErr:     `resolved to null`,
		},
		{
			name:        "empty array",
			jsonContent: `{"variables":{"entityIDs":[]}}`,
			itemPath:    "variables.entityIDs",
			wantErr:     `resolved to an empty array`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, _, err := getItem([]byte(tt.jsonContent), tt.itemPath)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}

				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("getItem returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
