package dhcprelaygroup

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestCreatePreservesPartialIdentity(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewDHCPRelayGroupResourceClient(t)
	r := NewDHCPRelayGroupResource().(*relayGroupResource)
	r.client = &catoclient.Service{AccountId: "account-123"}
	r.apiClient = client
	var result cato.NetworkConfigDhcpCreateRelayGroup
	require.NoError(t, json.Unmarshal([]byte(`{"networkConfig":{"dhcp":{"createRelayGroup":{"relayGroup":[{"id":"opaque-123","name":"test","server":["192.0.2.1"]}]}}}}`), &result))
	client.EXPECT().NetworkConfigDhcpCreateRelayGroup(mock.Anything, "account-123", mock.Anything).Return(&result, errors.New("partial response")).Once()
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	plan := tfsdk.Plan{Schema: schema.Schema}
	require.False(t, plan.Set(ctx, RelayGroupModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Server: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("192.0.2.1")})}).HasError())
	response := resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
	require.True(t, response.Diagnostics.HasError())
	var id types.String
	require.False(t, response.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "opaque-123", id.ValueString())
	require.True(t, response.State.Raw.IsFullyKnown())
}
func TestDeleteRaceRequiresConfirmedAbsence(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(map[bool]string{false: "absent", true: "still present"}[present], func(t *testing.T) {
			calls := []string{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				defer func() { require.NoError(t, req.Body.Close()) }()
				var body struct {
					OperationName string         `json:"operationName"`
					Variables     map[string]any `json:"variables"`
				}
				require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
				calls = append(calls, body.OperationName)
				require.Equal(t, "account-123", body.Variables["accountId"])
				w.Header().Set("Content-Type", "application/json")
				var response any
				if len(calls) == 1 {
					response = map[string]any{"errors": []any{map[string]any{"message": "backend rejected delete", "path": []any{"networkConfig", "dhcp", "deleteRelayGroup"}, "extensions": map[string]any{"code": "CC2_ERROR"}}}, "data": nil}
				} else {
					var value any
					if present {
						require.NoError(t, json.Unmarshal([]byte(`{"id":"opaque-123","name":"test","server":["192.0.2.1"]}`), &value))
					}
					response = map[string]any{"data": map[string]any{"networkConfig": map[string]any{"dhcp": map[string]any{"relayGroup": value}}}}
					input := body.Variables["networkConfigDhcp"+"RelayGroupRefInput"]
					if input != nil {
						require.Equal(t, map[string]any{"by": "ID", "input": "opaque-123"}, input)
					}
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))
			}))
			defer server.Close()
			sdk, err := cato.New(server.URL, "test-key", "account-123", server.Client(), nil)
			require.NoError(t, err)
			r := NewDHCPRelayGroupResource()
			r.(resource.ResourceWithConfigure).Configure(context.Background(), resource.ConfigureRequest{ProviderData: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}, &resource.ConfigureResponse{})
			var schema resource.SchemaResponse
			r.Schema(context.Background(), resource.SchemaRequest{}, &schema)
			value := RelayGroupModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Server: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("192.0.2.1")})}
			value.ID = types.StringValue("opaque-123")
			state := tfsdk.State{Schema: schema.Schema}
			require.False(t, state.Set(context.Background(), value).HasError())
			response := resource.DeleteResponse{State: state}
			r.Delete(context.Background(), resource.DeleteRequest{State: state}, &response)
			require.Equal(t, present, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			require.Len(t, calls, 2)
		})
	}
}

func TestDefaultTransportPreservesCreateIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { require.NoError(t, r.Body.Close()) }()
		require.Equal(t, "test-key", r.Header.Get("x-api-key"))
		require.Equal(t, "account-123", r.Header.Get("x-account-id"))
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"data":{"networkConfig":{"dhcp":{"createRelayGroup":{"relayGroup":[{"id":"opaque-123","name":"test","server":["192.0.2.1"]}]}}}},"errors":[{"message":"partial response"}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	sdk, err := cato.New(server.URL, "test-key", "account-123", server.Client(), nil)
	require.NoError(t, err)
	require.False(t, sdk.Client.ParseDataWhenErrors)
	r := NewDHCPRelayGroupResource()
	r.(resource.ResourceWithConfigure).Configure(context.Background(), resource.ConfigureRequest{ProviderData: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}, &resource.ConfigureResponse{})
	var schema resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schema)
	plan := tfsdk.Plan{Schema: schema.Schema}
	require.False(t, plan.Set(context.Background(), RelayGroupModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Server: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("192.0.2.1")})}).HasError())
	response := resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
	require.True(t, response.Diagnostics.HasError())
	var id types.String
	require.False(t, response.State.GetAttribute(context.Background(), path.Root("id"), &id).HasError())
	require.Equal(t, "opaque-123", id.ValueString())
	require.True(t, response.State.Raw.IsFullyKnown())
	require.False(t, sdk.Client.ParseDataWhenErrors)
}
