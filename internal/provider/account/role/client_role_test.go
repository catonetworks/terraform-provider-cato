package role

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestDeleteRoleRetry(t *testing.T) {
	for _, scenario := range []string{"retry", "unavailable status", "already deleted", "business error", "unrelated error", "canceled"} {
		t.Run(scenario, func(t *testing.T) {
			deletes, reads := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					OperationName string                     `json:"operationName"`
					Variables     map[string]json.RawMessage `json:"variables"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, "fixture-key", r.Header.Get("x-api-key"))
				require.Equal(t, `"123"`, string(body.Variables["accountId"]))
				w.Header().Set("Content-Type", "application/json")
				if body.OperationName == "rbacRoleManagementReadRole" {
					reads++
					require.Equal(t, `"r"`, string(body.Variables["id"]))
					role := `{"id":"r","name":"test","permission":[],"applicableAccountType":[]}`
					if scenario == "already deleted" {
						role = "null"
					}
					_, _ = w.Write([]byte(`{"data":{"rbac":{"__typename":"RBACQueries","roleManagement":{"__typename":"RoleManagementQueries","role":` + role + `}}}}`))
					return
				}
				require.Equal(t, "rbacRoleManagementDeleteRole", body.OperationName)
				require.JSONEq(t, `{"id":"r"}`, string(body.Variables["roleManagementDeleteRoleInput"]))
				deletes++
				if deletes > 1 {
					_, _ = w.Write([]byte(`{"data":{"rbac":{"roleManagement":{"deleteRole":{"id":"r"}}}}}`))
					return
				}
				code, detail := "DOWNSTREAM_SERVICE_ERROR", "Role management is temporarily unavailable"
				if scenario == "business error" {
					code = "BAD_USER_INPUT"
				}
				if scenario == "unrelated error" {
					detail = "Role is assigned to admins"
				}
				response := map[string]any{"errors": []any{map[string]any{"message": detail, "path": []string{"rbac", "roleManagement", "deleteRole"}, "extensions": map[string]any{"code": code, "full": detail}}}}
				if scenario == "unavailable status" {
					response["errors"].([]any)[0].(map[string]any)["extensions"].(map[string]any)["errorType"] = "UNAVAILABLE"
					response["errors"].([]any)[0].(map[string]any)["extensions"].(map[string]any)["full"] = "Service unavailable"
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))
			}))
			defer server.Close()
			client, err := sdk.New(server.URL, "fixture-key", "123", server.Client(), nil)
			require.NoError(t, err)
			ctx := context.Background()
			if scenario == "canceled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()
			}
			deleted, err := deleteRoleWithRetry(ctx, client, "123", "r")
			switch scenario {
			case "retry", "unavailable status":
				require.NoError(t, err)
				require.True(t, deleted)
				require.Equal(t, 2, deletes)
				require.Equal(t, 1, reads)
			case "already deleted":
				require.NoError(t, err)
				require.True(t, deleted)
				require.Equal(t, 1, deletes)
				require.Equal(t, 1, reads)
			case "canceled":
				require.ErrorIs(t, err, context.DeadlineExceeded)
				require.False(t, deleted)
				require.Equal(t, 1, deletes)
			default:
				require.Error(t, err)
				require.False(t, deleted)
				require.Equal(t, 1, deletes)
				require.Zero(t, reads)
			}
		})
	}
}
