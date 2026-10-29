package dnssitesettings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestParentSiteAbsence(t *testing.T) {
	cases := []struct {
		name, lookup, code, errorLeaf string
		absent                        bool
		lookupCalls                   int
	}{
		{name: "downstream deleted parent", code: "DOWNSTREAM_SERVICE_ERROR", lookup: `{"data":{"entityLookup":{"items":[],"total":0}}}`, absent: true, lookupCalls: 1},
		{name: "deleted parent", lookup: `{"data":{"entityLookup":{"items":[],"total":0}}}`, absent: true, lookupCalls: 1},
		{name: "existing parent", lookup: `{"data":{"entityLookup":{"items":[{"entity":{"id":"site-123"}}],"total":1}}}`, lookupCalls: 1},
		{name: "lookup permission error", lookup: `{"errors":[{"message":"denied","extensions":{"code":"FORBIDDEN"}}],"data":null}`, lookupCalls: 1},
		{name: "lookup transient error", lookup: `{"errors":[{"message":"unavailable","extensions":{"code":"INTERNAL_SERVER_ERROR"}}],"data":null}`, lookupCalls: 1},
		{name: "lookup null namespace", lookup: `{"data":{"entityLookup":null}}`, lookupCalls: 1},
		{name: "lookup omitted namespace", lookup: `{"data":{}}`, lookupCalls: 1},
		{name: "lookup null items", lookup: `{"data":{"entityLookup":{"items":null,"total":0}}}`, lookupCalls: 1},
		{name: "lookup omitted total", lookup: `{"data":{"entityLookup":{"items":[]}}}`, lookupCalls: 1},
		{name: "lookup inconsistent total", lookup: `{"data":{"entityLookup":{"items":[],"total":1}}}`, lookupCalls: 1},
		{name: "lookup wrong site", lookup: `{"data":{"entityLookup":{"items":[{"entity":{"id":"other-site"}}],"total":1}}}`, lookupCalls: 1},
		{name: "DNS permission error", code: "FORBIDDEN"},
		{name: "DNS transient error", code: "INTERNAL_SERVER_ERROR"},
		{name: "different error path", errorLeaf: "networkConfig"},
	}
	for _, action := range []string{"read", "delete"} {
		for _, tc := range cases {
			t.Run(action+"/"+tc.name, func(t *testing.T) {
				ctx := context.Background()
				calls := 0
				leaf := "siteSettings"
				if action == "delete" {
					leaf = "removeSiteSettings"
				}
				if tc.errorLeaf != "" {
					leaf = tc.errorLeaf
				}
				code := tc.code
				if code == "" {
					code = "CC2_ERROR"
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var body struct {
						OperationName string         `json:"operationName"`
						Variables     map[string]any `json:"variables"`
					}
					require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					w.Header().Set("Content-Type", "application/json")
					if body.OperationName == "entityLookup" {
						calls++
						require.Equal(t, "account-123", body.Variables["accountID"])
						require.Equal(t, "site", body.Variables["type"])
						require.Equal(t, []any{"site-123"}, body.Variables["entityIDs"])
						_, err := w.Write([]byte(tc.lookup))
						require.NoError(t, err)
						return
					}
					_, err := fmt.Fprintf(w, `{"errors":[{"message":"backend rejected request","path":["networkConfig","dns",%q],"extensions":{"code":%q}}],"data":null}`, leaf, code)
					require.NoError(t, err)
				}))
				defer server.Close()
				sdk, err := cato.New(server.URL, "test-key", "account-123", server.Client(), nil)
				require.NoError(t, err)
				r := &siteSettingsResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
				var schema resource.SchemaResponse
				r.Schema(ctx, resource.SchemaRequest{}, &schema)
				state := tfsdk.State{Schema: schema.Schema}
				require.False(t, state.Set(ctx, SiteSettingsModel{ID: types.StringValue("site-123"), Site: types.ObjectNull(idname.RefAttrTypes), PrimaryServer: types.StringValue("1.1.1.1"), SecondaryServer: types.StringValue(""), Suffix: types.ListValueMust(types.StringType, nil)}).HasError())
				if action == "read" {
					response := &resource.ReadResponse{State: state}
					r.Read(ctx, resource.ReadRequest{State: state}, response)
					require.Equal(t, !tc.absent, response.Diagnostics.HasError(), response.Diagnostics)
					if tc.absent {
						require.True(t, response.State.Raw.IsNull())
					} else {
						require.Equal(t, state.Raw, response.State.Raw)
					}
				} else {
					response := &resource.DeleteResponse{}
					r.Delete(ctx, resource.DeleteRequest{State: state}, response)
					require.Equal(t, !tc.absent, response.Diagnostics.HasError(), response.Diagnostics)
				}
				require.Equal(t, tc.lookupCalls, calls)
			})
		}
	}
}
