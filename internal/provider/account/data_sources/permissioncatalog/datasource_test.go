package permissioncatalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

func TestRead(t *testing.T) {
	for _, scenario := range []string{"success", "malformed", "api error"} {
		t.Run(scenario, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				body := `{"data":{"rbac":{"roleManagement":{"permissionCatalog":{"resource":[{"resource":"Sites","supportedAction":["VIEW","EDIT"]}]}}}}}`
				if scenario == "malformed" {
					body = `{"data":{"rbac":{}}}`
				}
				if scenario == "api error" {
					body = `{"errors":[{"message":"denied"}]}`
				}
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()
			client, err := sdk.New(server.URL, "fixture-key", "123", server.Client(), nil)
			require.NoError(t, err)
			d := &catalogDataSource{DataConfig: iam.DataConfig{Service: &catoclient.Service{AccountId: "123", Catov2: catoclient.NewProviderSDKClient(client)}}}
			var schema datasource.SchemaResponse
			d.Schema(context.Background(), datasource.SchemaRequest{}, &schema)
			config := tfsdk.Config{Schema: schema.Schema}
			model := Model{AccountID: types.StringValue("123"), Resources: types.SetNull(catalogObjectType)}
			stateSeed := tfsdk.State{Schema: schema.Schema}
			require.False(t, stateSeed.Set(context.Background(), model).HasError())
			config.Raw = stateSeed.Raw
			resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
			d.Read(context.Background(), datasource.ReadRequest{Config: config}, &resp)
			if scenario != "success" {
				require.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			var state Model
			require.False(t, resp.State.Get(context.Background(), &state).HasError())
			require.Len(t, state.Resources.Elements(), 1)
		})
	}
}
