package dns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func checkReadFailure(t *testing.T, res resource.Resource, model any, leaf string, removable bool) {
	t.Helper()
	cases := []struct {
		name, body string
		missing    bool
	}{
		{"denied", `{"errors":[{"message":"forbidden","path":["networkConfig"],"extensions":{"code":"FORBIDDEN"}}],"data":null}`, false},
		{"null namespace", `{"data":{"networkConfig":null}}`, false},
		{"null domain", `{"data":{"networkConfig":{"dns":null}}}`, false},
		{"omitted domain", `{"data":{"networkConfig":{}}}`, false},
		{"omitted leaf", `{"data":{"networkConfig":{"dns":{}}}}`, false},
		{"missing leaf", `{"data":{"networkConfig":{"dns":{"` + leaf + `":null}}}}`, removable},
		{"not found wrong path", `{"errors":[{"message":"missing account","path":["networkConfig"],"extensions":{"code":"NOT_FOUND"}}],"data":null}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(tc.body))
				require.NoError(t, err)
			}))
			defer server.Close()
			sdk, err := cato.New(server.URL, "test-key", "account-123", server.Client(), nil)
			require.NoError(t, err)
			res.(resource.ResourceWithConfigure).Configure(context.Background(), resource.ConfigureRequest{ProviderData: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}, &resource.ConfigureResponse{})
			schema := &resource.SchemaResponse{}
			res.Schema(context.Background(), resource.SchemaRequest{}, schema)
			state := tfsdk.State{Schema: schema.Schema}
			require.False(t, state.Set(context.Background(), model).HasError())
			response := &resource.ReadResponse{State: state}
			res.Read(context.Background(), resource.ReadRequest{State: state}, response)
			if tc.missing {
				require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
				require.True(t, response.State.Raw.IsNull())
			} else {
				require.True(t, response.Diagnostics.HasError())
				require.Equal(t, state.Raw, response.State.Raw)
			}
		})
	}
}

func TestSettingsReadFailures(t *testing.T) {
	checkReadFailure(t, NewSettingsResource(), SettingsModel{ID: types.StringValue("account-123"), PrimaryServer: types.StringNull(), SecondaryServer: types.StringNull(), Suffix: types.ListValueMust(types.StringType, nil), AcceptDNSRequestsOnLanInterfaceIP: types.BoolValue(false)}, "settings", false)
}

func TestSiteSettingsReadFailures(t *testing.T) {
	checkReadFailure(t, NewSiteSettingsResource(), SiteSettingsModel{ID: types.StringValue("site-123"), SiteID: types.StringValue("site-123"), PrimaryServer: types.StringNull(), SecondaryServer: types.StringNull(), Suffix: types.ListValueMust(types.StringType, nil)}, "siteSettings", true)
}

func TestForwardingRuleReadFailures(t *testing.T) {
	checkReadFailure(t, NewForwardingRuleResource(), ForwardingRuleModel{ID: types.StringValue("opaque-123"), Domain: types.StringValue("terraform.test"), Server: types.ListValueMust(types.StringType, nil)}, "forwardingRule", true)
}

func TestSettingsImportRejectsDifferentAccount(t *testing.T) {
	res := NewSettingsResource()
	res.(resource.ResourceWithConfigure).Configure(context.Background(), resource.ConfigureRequest{ProviderData: &catoclient.Service{AccountId: "account-123"}}, &resource.ConfigureResponse{})
	response := &resource.ImportStateResponse{}
	res.(resource.ResourceWithImportState).ImportState(context.Background(), resource.ImportStateRequest{ID: "other-account"}, response)
	require.True(t, response.Diagnostics.HasError())
}
