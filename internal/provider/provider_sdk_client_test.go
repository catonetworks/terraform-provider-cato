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
		Query         string         `json:"query"`
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
	require.Equal(t, map[string]any{
		"paging": map[string]any{"from": float64(0), "limit": float64(generatedGroupMembersLimit)},
		"sort":   map[string]any{},
	}, request.Variables["groupMembersListInput"])
	require.Equal(t, "test-group", nestedRequestString(t, request.Variables, "createGroupInput", "name"))
}

func TestProviderSDKClientInternetFirewallAddSubPolicyUsesMinimalResponse(t *testing.T) {
	t.Parallel()

	var request struct {
		OperationName string         `json:"operationName"`
		Query         string         `json:"query"`
		Variables     map[string]any `json:"variables"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { require.NoError(t, r.Body.Close()) }()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"data":{"policy":{"internetFirewall":{"addSubPolicy":{"status":"SUCCESS","errors":[]}}}}}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := cato.New(server.URL, "test-token", "account-1", nil, nil)
	require.NoError(t, err)
	result, err := newProviderSDKClient(client).PolicyInternetFirewallAddSubPolicy(
		context.Background(),
		&cato_models.InternetFirewallPolicyMutationInput{},
		cato_models.InternetFirewallAddSubPolicyInput{},
		"account-1",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "policyInternetFirewallAddSubPolicy", request.OperationName)
	require.Equal(t, "account-1", request.Variables["accountId"])
	require.Contains(t, request.Query, "status")
	require.NotContains(t, request.Query, "rules")
}

func TestProviderSDKClientWanFirewallAddSubPolicyUsesMinimalResponse(t *testing.T) {
	t.Parallel()

	var request struct {
		OperationName string         `json:"operationName"`
		Query         string         `json:"query"`
		Variables     map[string]any `json:"variables"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { require.NoError(t, r.Body.Close()) }()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"data":{"policy":{"wanFirewall":{"addSubPolicy":{"status":"SUCCESS","errors":[]}}}}}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := cato.New(server.URL, "test-token", "account-1", nil, nil)
	require.NoError(t, err)
	result, err := newProviderSDKClient(client).PolicyWanFirewallAddSubPolicy(
		context.Background(),
		cato_models.WanFirewallAddSubPolicyInput{},
		"account-1",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "policyWanFirewallAddSubPolicy", request.OperationName)
	require.Equal(t, "account-1", request.Variables["accountId"])
	require.Contains(t, request.Query, "status")
	require.NotContains(t, request.Query, "rules")
}
