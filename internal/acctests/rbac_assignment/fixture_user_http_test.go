//go:build acctest

package rbacassignment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestAssignmentTestUserLifecycle(t *testing.T) {
	operations := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			OperationName string                     `json:"operationName"`
			Variables     map[string]json.RawMessage `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "fixture-key", r.Header.Get("x-api-key"))
		require.Equal(t, `"test-account"`, string(body.Variables["accountId"]))
		select {
		case operations <- body.OperationName:
		default:
			t.Error("unexpected extra API call")
		}
		w.Header().Set("Content-Type", "application/json")
		switch body.OperationName {
		case "userCreateUser":
			require.JSONEq(t, `{"firstName":"Terraform","lastName":"unique-test","email":"unique-test@example.invalid"}`, string(body.Variables["createUserInput"]))
			_, _ = w.Write([]byte(`{"data":{"user":{"createUser":{"user":{"id":"123"}}}}}`))
		case "userDeleteUser":
			require.JSONEq(t, `{"userId":[123]}`, string(body.Variables["deleteUserInput"]))
			_, _ = w.Write([]byte(`{"data":{"user":{"deleteUser":{"status":"SUCCESS"}}}}`))
		default:
			t.Errorf("unexpected operation %s", body.OperationName)
		}
	}))
	defer server.Close()
	client, err := sdk.New(server.URL, "fixture-key", "test-account", server.Client(), nil)
	require.NoError(t, err)
	t.Run("fixture", func(t *testing.T) {
		require.Equal(t, "123", createAssignmentTestUser(t, client, "test-account", "unique-test"))
	})
	require.Len(t, operations, 2)
	require.Equal(t, "userCreateUser", <-operations)
	require.Equal(t, "userDeleteUser", <-operations)
}
