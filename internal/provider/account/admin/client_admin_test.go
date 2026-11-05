package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestManagedRoleLookupAccountScope(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			OperationName string                     `json:"operationName"`
			Variables     map[string]json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		require.Equal(t, "accountRoles", request.OperationName)
		require.Equal(t, `"managed-account"`, string(request.Variables["accountID"]))
		require.Equal(t, "null", string(request.Variables["accountType"]))
		require.Equal(t, "fixture-key", r.Header.Get("x-api-key"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"accountRoles":{"items":[{"id":"2","name":"Viewer"}],"total":1}}}`))
	}))
	defer server.Close()
	httpClient := server.Client()
	httpClient.Timeout = time.Second
	client, err := sdk.New(server.URL, "fixture-key", "provider-account", httpClient, nil)
	require.NoError(t, err)
	r := &adminResource{client: &catoclient.Service{AccountId: "provider-account", Catov2: catoclient.NewProviderSDKClient(client)}}
	id, err := resolveManagedRoleID(context.Background(), r.resourceAPI(), "managed-account", "Viewer")
	require.NoError(t, err)
	require.Equal(t, "2", id)
}
