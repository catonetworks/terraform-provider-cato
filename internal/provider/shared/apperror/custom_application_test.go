package apperror

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/stretchr/testify/require"
)

func TestMissingCustomApplication(t *testing.T) {
	for _, tc := range []struct {
		name, id, message string
		want              bool
	}{
		{"numeric", "42", "Custom application not found by ID: CustomApp_123_42", true},
		{"prefixed numeric", "42", "GQL-User-Error: Custom application not found by ID: CustomApp_123_42", true},
		{"opaque", "app-1", "Custom application not found by ID: app-1", true},
		{"prefixed opaque", "app-1", "GQL-User-Error: Custom application not found by ID: app-1", true},
		{"different account", "42", "Custom application not found by ID: CustomApp_456_42", false},
		{"different identity", "42", "Custom application not found by ID: CustomApp_123_43", false},
		{"denied", "42", "Access denied", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var response clientv2.ErrorResponse
			body := fmt.Sprintf(`{"graphqlErrors":[{"message":%q}]}`, tc.message)
			require.NoError(t, json.Unmarshal([]byte(body), &response))
			require.Equal(t, tc.want, MissingCustomApplication(fmt.Errorf("wrapped: %w", &response), tc.id, "123"))
		})
	}
	var response clientv2.ErrorResponse
	require.NoError(t, json.Unmarshal([]byte(`{"graphqlErrors":[{"message":"Custom application not found by ID: app-1"},{"message":"Access denied"}]}`), &response))
	require.False(t, MissingCustomApplication(&response, "app-1", "123"))
}
