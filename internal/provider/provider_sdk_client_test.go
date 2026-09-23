package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/require"
)

func TestProviderSDKClientGroupsCreateGroupUsesUnfilteredMembersProjection(t *testing.T) {
	t.Parallel()

	var request struct {
		OperationName string         `json:"operationName"`
		Variables     map[string]any `json:"variables"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { require.NoError(t, r.Body.Close()) }()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"data":{"groups":{"createGroup":null}}}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := cato.New(server.URL, "test-token", "account-1", nil, nil)
	require.NoError(t, err)
	_, err = newProviderSDKClient(client).GroupsCreateGroup(
		context.Background(),
		cato_models.CreateGroupInput{Name: "test-group"},
		"account-1",
	)
	require.NoError(t, err)
	require.Equal(t, "groupsCreateGroup", request.OperationName)
	require.Equal(t, "account-1", request.Variables["accountId"])
	require.Equal(t, map[string]any{"paging": nil, "sort": nil}, request.Variables["groupMembersListInput"])
	require.Equal(t, "test-group", nestedRequestString(t, request.Variables, "createGroupInput", "name"))
}
