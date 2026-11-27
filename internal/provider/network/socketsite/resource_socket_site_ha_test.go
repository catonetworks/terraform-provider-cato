package socketsite

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestSocketSiteCreateSecondaryVSocket(t *testing.T) {
	t.Parallel()

	const updateHA = "siteUpdateHa"
	tests := map[string]struct {
		connection, model, cloud string
		secondary                map[string]string
		input                    map[string]any
		cloudRouter              types.Bool
		failOperation            string
		invalid                  bool
	}{
		"legacy single socket": {connection: "SOCKET_GCP1500", model: "GCP"},
		"legacy Cloud Router":  {connection: "SOCKET_GCP1500", model: "GCP", cloudRouter: types.BoolValue(true)},
		"AWS": {
			connection: "SOCKET_AWS1500", model: "AWS", cloud: "aws",
			secondary: map[string]string{"eni_ip_address": "10.51.0.132", "eni_ip_subnet": "10.51.0.128/25", "route_table_id": "rtb-test"},
			input:     map[string]any{"eniIpAddress": "10.51.0.132", "eniIpSubnet": "10.51.0.128/25", "routeTableId": "rtb-test"},
		},
		"Azure": {
			connection: "SOCKET_AZ1500", model: "AZURE", cloud: "azure",
			secondary: map[string]string{"interface_ip": "10.51.0.132", "floating_ip": "10.51.0.133"},
			input:     map[string]any{"interfaceIp": "10.51.0.132", "floatingIp": "10.51.0.133"},
		},
		"GCP Cloud Router": {
			connection: "SOCKET_GCP1500", model: "GCP", cloud: "gcp",
			secondary:   map[string]string{"interface_ip": "10.51.0.132", "load_balancer_ip": "10.51.0.133"},
			input:       map[string]any{"interfaceIp": "10.51.0.132", "loadBalancerIp": "10.51.0.133"},
			cloudRouter: types.BoolValue(true),
		},
		"site API failure": {
			connection: "SOCKET_GCP1500", cloud: "gcp", failOperation: "siteAddSocketSite",
			secondary:   map[string]string{"interface_ip": "10.51.0.132", "load_balancer_ip": "10.51.0.133"},
			cloudRouter: types.BoolValue(true),
		},
		"Cloud Router API failure": {
			connection: "SOCKET_GCP1500", cloud: "gcp", failOperation: updateHA,
			secondary:   map[string]string{"interface_ip": "10.51.0.132", "load_balancer_ip": "10.51.0.133"},
			cloudRouter: types.BoolValue(true),
		},
		"invalid cloud fields": {
			connection: "SOCKET_GCP1500",
			secondary:  map[string]string{"interface_ip": "10.51.0.132", "floating_ip": "10.51.0.133"},
			invalid:    true,
		},
		"AWS ENI outside subnet": {
			connection: "SOCKET_AWS1500",
			secondary:  map[string]string{"eni_ip_address": "10.51.0.132", "eni_ip_subnet": "10.51.1.0/24", "route_table_id": "rtb-test"},
			invalid:    true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			client, received := newSocketSiteHATestClient(t, tc.model, tc.cloud != "", tc.failOperation)
			r := &socketSiteResource{client: client}
			var schemaResp resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			model := newSocketSitePlanWithDestType(ctx, t, types.StringNull(), types.Int64Null())
			fields := model.NativeRange.Attributes()
			fields["local_ip"] = types.StringValue("10.51.0.129")
			model.NativeRange = types.ObjectValueMust(SiteNativeRangeResourceAttrTypes, fields)
			model.Name = types.StringValue("aws-site-01")
			model.ConnectionType = types.StringValue(tc.connection)
			model.SiteType = types.StringValue("CLOUD_DC")
			model.Sockets = types.SetUnknown(types.ObjectType{AttrTypes: SocketTypes})
			model.SecondaryVSocket = types.ObjectNull(secondaryVSocketAttrTypes)
			if tc.secondary != nil {
				values := make(map[string]attr.Value, len(secondaryVSocketAttrTypes))
				for key := range secondaryVSocketAttrTypes {
					values[key] = types.StringNull()
					if value, ok := tc.secondary[key]; ok {
						values[key] = types.StringValue(value)
					}
				}
				model.SecondaryVSocket = types.ObjectValueMust(secondaryVSocketAttrTypes, values)
			}

			config := tfsdk.Config{Schema: schemaResp.Schema}
			plan := tfsdk.Plan{Schema: schemaResp.Schema}
			model.IsCloudRouter = tc.cloudRouter
			require.False(t, plan.Set(ctx, model).HasError())
			config.Raw = plan.Raw
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			r.Create(ctx, resource.CreateRequest{Config: config, Plan: plan}, &resp)

			requests := make([]socketSiteGraphQLRequest, 0, len(received))
			for len(received) > 0 {
				requests = append(requests, <-received)
			}
			if tc.invalid {
				require.True(t, resp.Diagnostics.HasError())
				require.Empty(t, requests)
				return
			}
			operations := []string{"siteAddSocketSite"}
			if tc.failOperation == "siteAddSocketSite" {
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, operations, socketSiteOperationNames(requests))
				return
			}
			if tc.cloud != "" {
				operations = append(operations, "siteSocketConfiguration")
			} else if !tc.cloudRouter.IsNull() {
				operations = append(operations, updateHA)
			}
			operations = append(operations, "networkRangeList", "siteUpdateNetworkRange")
			if tc.cloud == "" {
				operations = append(operations, "siteUpdateSocketInterface")
			} else {
				operations = append(operations, updateHA)
			}
			if tc.failOperation != "" {
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, append(operations, "siteRemoveSite"), socketSiteOperationNames(requests))
				return
			}

			operations = append(operations, "siteGeneralDetails", "siteSocketConfiguration", "networkRangeList", "entityLookup")
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			require.Equal(t, operations, socketSiteOperationNames(requests))
			createInput, ok := requests[0].Variables["addSocketSiteInput"].(map[string]any)
			require.True(t, ok)
			require.Equal(t, "account-1", requests[0].Variables["accountId"])
			if tc.cloud == "" {
				require.NotContains(t, createInput, "secondaryVSocket")
				if !tc.cloudRouter.IsNull() {
					require.Equal(t, map[string]any{"isCloudRouter": true}, requests[1].Variables["updateHaInput"])
				}
			} else {
				require.Equal(t, map[string]any{tc.cloud: tc.input}, createInput["secondaryVSocket"])
				networkInput, ok := requests[3].Variables["updateNetworkRangeInput"].(map[string]any)
				require.True(t, ok)
				require.NotContains(t, networkInput, "localIp")
				updateInput := map[string]any{"primaryManagementIp": "10.51.0.129"}
				if !tc.cloudRouter.IsNull() {
					updateInput["isCloudRouter"] = tc.cloudRouter.ValueBool()
				}
				require.Equal(t, updateInput, requests[4].Variables["updateHaInput"])
			}
			assertCreatedVSocketSite(t, resp.State, model.SecondaryVSocket, tc.cloud != "")
		})
	}
}

func assertCreatedVSocketSite(t *testing.T, tfState tfsdk.State, secondaryConfig types.Object, ha bool) {
	t.Helper()
	ctx := t.Context()
	var state SocketSite
	require.False(t, tfState.Get(ctx, &state).HasError())
	require.Equal(t, secondaryConfig, state.SecondaryVSocket)
	require.Equal(t, types.StringValue("10.51.0.129"), state.NativeRange.Attributes()["local_ip"])
	var sockets []Socket
	require.False(t, state.Sockets.ElementsAs(ctx, &sockets, false).HasError())
	serials := []string{"primary-serial"}
	if ha {
		serials = append(serials, "secondary-serial")
	}
	var actualSerials []string
	for _, socket := range sockets {
		actualSerials = append(actualSerials, socket.SerialNumber.ValueString())
	}
	require.ElementsMatch(t, serials, actualSerials)
}

func TestSocketSiteWaitForHARequiresSecondarySerial(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	client := mocks.NewSocketSiteClient(t)
	configuration := socketConfigurationForTest(models.SocketModelGCP)
	configuration.PrimarySocketConfiguration.Serial = new("primary-serial")
	client.EXPECT().SiteSocketConfiguration(mock.Anything, mock.Anything, "account-1").
		Run(func(context.Context, models.SiteSocketConfigurationInput, string, ...clientv2.RequestInterceptor) {
			cancel()
		}).
		Return(&cato.SiteSocketConfiguration{
			Site: cato.SiteSocketConfiguration_Site{SiteSocketConfiguration: configuration},
		}, nil).Once()
	r := &socketSiteResource{client: &catoclient.Service{AccountId: "account-1"}, socketSiteClient: client}
	var diags diag.Diagnostics
	r.waitForHASerials(ctx, "site-1", &diags)
	require.True(t, diags.HasError())
}

func TestSocketSiteInlineHAUpdatePrimaryManagementIP(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := mocks.NewSocketSiteClient(t)
	r := &socketSiteResource{client: &catoclient.Service{AccountId: "account-1"}, socketSiteClient: client}
	prior := newSocketSitePlanWithDestType(ctx, t, types.StringNull(), types.Int64Null())
	prior.SecondaryVSocket = types.ObjectValueMust(secondaryVSocketAttrTypes, map[string]attr.Value{
		"eni_ip_address": types.StringNull(), "eni_ip_subnet": types.StringNull(), "route_table_id": types.StringNull(),
		"interface_ip": types.StringValue("10.51.0.132"), "load_balancer_ip": types.StringValue("10.51.0.133"),
		"floating_ip": types.StringNull(),
	})
	prior.IsCloudRouter = types.BoolValue(true)
	prior.Sockets = types.SetNull(types.ObjectType{AttrTypes: SocketTypes})
	plan := *prior
	fields := plan.NativeRange.Attributes()
	fields["local_ip"] = types.StringValue("10.51.0.130")
	fields["primary_management_ip"] = types.StringUnknown()
	plan.NativeRange = types.ObjectValueMust(SiteNativeRangeResourceAttrTypes, fields)

	var diags diag.Diagnostics
	r.updateInlineHA(ctx, prior, prior, "site-1", &diags)
	require.False(t, diags.HasError())
	client.AssertNotCalled(t, "SiteUpdateHa", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	client.EXPECT().SiteUpdateHa(mock.Anything, "account-1", "site-1",
		models.UpdateHaInput{PrimaryManagementIP: new("10.51.0.130")}).
		Return(&cato.SiteUpdateHa{}, nil).Once()
	r.updateInlineHA(ctx, &plan, prior, "site-1", &diags)
	require.False(t, diags.HasError())

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	tfPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	require.False(t, tfPlan.Set(ctx, &plan).HasError())
	require.False(t, tfState.Set(ctx, prior).HasError())
	req := planmodifier.StringRequest{
		Path: path.Root("native_range").AtName("primary_management_ip"), Plan: tfPlan, State: tfState,
		PlanValue: types.StringUnknown(), StateValue: types.StringValue("10.51.0.129"),
	}
	resp := planmodifier.StringResponse{PlanValue: req.PlanValue}
	primaryManagementIPPlanModifier{}.PlanModifyString(ctx, req, &resp)
	require.False(t, resp.Diagnostics.HasError())
	require.True(t, resp.PlanValue.IsUnknown())
}

type socketSiteGraphQLRequest struct {
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

func newSocketSiteHATestClient(t *testing.T, model string, ha bool, failOperation string,
) (client *catoclient.Service, captured <-chan socketSiteGraphQLRequest) {
	t.Helper()

	sockets := map[string]any{
		"primarySocketConfiguration": map[string]any{
			"serial": "primary-serial", "managementIp": "10.51.0.129", "socketInfo": map[string]any{"isPrimary": true, "model": model, "platform": model},
		},
	}
	if ha {
		sockets["secondarySocketConfiguration"] = map[string]any{
			"serial": "secondary-serial", "socketInfo": map[string]any{"isPrimary": false, "model": model, "platform": model},
		}
	}
	socketData, err := json.Marshal(sockets)
	require.NoError(t, err)

	responses := map[string]string{
		"siteAddSocketSite":         `{"site":{"addSocketSite":{"siteId":"site-1"}}}`,
		"networkRangeList":          `{"site":{"networkRangeList":{"items":[{"networkRangeId":"range-1","name":"Native Range","rangeType":"Native","subnet":"10.51.0.128/25","localIp":"10.51.0.129"}]}}}`,
		"siteUpdateNetworkRange":    `{"site":{"updateNetworkRange":{"networkRangeId":"range-1"}}}`,
		"siteUpdateSocketInterface": `{"site":{"updateSocketInterface":{"siteId":"site-1"}}}`,
		"siteUpdateHa":              `{"site":{"updateHa":{"siteId":"site-1"}}}`,
		"siteGeneralDetails":        `{"site":{"siteGeneralDetails":{"site":{"id":"site-1","name":"aws-site-01"},"siteType":"CLOUD_DC","siteLocation":{"countryCode":"US","timezone":"America/Chicago"}}}}`,
		"siteSocketConfiguration":   `{"site":{"siteSocketConfiguration":` + string(socketData) + `}}`,
		"entityLookup":              `{"entityLookup":{"items":[{"entity":{"id":"interface-1","name":"LAN1"},"helperFields":{"isDefault":true,"interfaceIndex":"LAN1","destType":"LAN"}}]}}`,
		"siteRemoveSite":            `{"site":{"removeSite":{"siteId":"site-1"}}}`,
	}
	received := make(chan socketSiteGraphQLRequest, 128)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req socketSiteGraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode GraphQL request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		received <- req
		w.Header().Set("Content-Type", "application/json")
		body, ok := responses[req.OperationName]
		if !ok {
			t.Errorf("unexpected operation %s", req.OperationName)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		payload := `{"data":` + body + `}`
		if req.OperationName == failOperation {
			payload = `{"errors":[{"message":"the operation failed"}]}`
		}
		if _, err := w.Write([]byte(payload)); err != nil {
			t.Errorf("write GraphQL response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	sdkClient, err := cato.New(server.URL, "test-token", "account-1", nil, nil)
	require.NoError(t, err)
	return &catoclient.Service{AccountId: "account-1", Catov2: catoclient.NewProviderSDKClient(sdkClient)}, received
}

func socketSiteOperationNames(requests []socketSiteGraphQLRequest) []string {
	names := make([]string, 0, len(requests))
	for _, request := range requests {
		names = append(names, request.OperationName)
	}
	return names
}
