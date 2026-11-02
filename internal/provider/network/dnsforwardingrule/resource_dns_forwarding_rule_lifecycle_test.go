package dnsforwardingrule

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func checkLifecycle(t *testing.T, res resource.Resource, model any, leaf, payload, id string, retain bool) {
	t.Helper()
	ctx := context.Background()
	var operations []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "account-123", body.Variables["accountId"])
		require.Equal(t, "account-123", r.Header.Get("x-account-id"))
		operations = append(operations, body.OperationName)
		require.NotContains(t, strings.ToLower(body.OperationName), "setoption")
		require.NotContains(t, strings.ToLower(body.OperationName), "setrelaygroup")
		require.NotContains(t, strings.ToLower(body.OperationName), "setforwardingrule")
		assertSingleItem(t, body.Variables)
		field := leaf
		value := json.RawMessage(payload)
		if strings.Contains(body.OperationName, "Create") || strings.Contains(body.OperationName, "Update") || strings.Contains(body.OperationName, "Delete") || strings.Contains(body.OperationName, "Remove") {
			for _, operation := range []string{"create", "update", "delete", "remove"} {
				for _, entity := range []string{"Settings", "SiteSettings", "ForwardingRule", "Option", "RelayGroup"} {
					if strings.Contains(strings.ToLower(body.OperationName), strings.ToLower(operation+entity)) {
						field = operation + entity
					}
				}
			}
			writePayload := payload
			if leaf != "settings" && !strings.HasPrefix(payload, "[") {
				writePayload = "[" + payload + "]"
			}
			value = json.RawMessage(`{"` + leaf + `":` + writePayload + `}`)
		}
		data := map[string]any{"data": map[string]any{"networkConfig": map[string]any{"dns": map[string]any{field: value}}}}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(data))
	}))
	defer server.Close()
	sdk, err := cato.New(server.URL, "test-key", "account-123", server.Client(), nil)
	require.NoError(t, err)
	conf := &resource.ConfigureResponse{}
	res.(resource.ResourceWithConfigure).Configure(ctx, resource.ConfigureRequest{ProviderData: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}, conf)
	require.False(t, conf.Diagnostics.HasError(), conf.Diagnostics)
	sch := &resource.SchemaResponse{}
	res.Schema(ctx, resource.SchemaRequest{}, sch)
	plan := tfsdk.Plan{Schema: sch.Schema}
	require.False(t, plan.Set(ctx, model).HasError())
	created := &resource.CreateResponse{State: tfsdk.State{Schema: sch.Schema}}
	res.Create(ctx, resource.CreateRequest{Plan: plan}, created)
	require.False(t, created.Diagnostics.HasError(), created.Diagnostics)
	var gotID string
	require.False(t, created.State.GetAttribute(ctx, path.Root("id"), &gotID).HasError())
	require.Equal(t, id, gotID)
	read := &resource.ReadResponse{State: created.State}
	res.Read(ctx, resource.ReadRequest{State: created.State}, read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	require.Equal(t, created.State.Raw, read.State.Raw)
	updated := &resource.UpdateResponse{State: read.State}
	res.Update(ctx, resource.UpdateRequest{State: read.State, Plan: tfsdk.Plan{Schema: sch.Schema, Raw: read.State.Raw}}, updated)
	require.False(t, updated.Diagnostics.HasError(), updated.Diagnostics)
	imported := &resource.ImportStateResponse{State: created.State}
	res.(resource.ResourceWithImportState).ImportState(ctx, resource.ImportStateRequest{ID: id}, imported)
	require.False(t, imported.Diagnostics.HasError(), imported.Diagnostics)
	removed := &resource.DeleteResponse{State: updated.State}
	res.Delete(ctx, resource.DeleteRequest{State: updated.State}, removed)
	require.False(t, removed.Diagnostics.HasError(), removed.Diagnostics)
	expected := 4
	if retain {
		expected = 3
	}
	require.Len(t, operations, expected)
}

func TestForwardingRuleLifecycle(t *testing.T) {
	checkLifecycle(t, NewDNSForwardingRuleResource(), ForwardingRuleModel{ID: types.StringUnknown(), Domain: types.StringValue("terraform.test"), Server: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("192.0.2.53")})}, "forwardingRule", `{"id":"opaque-123","domain":"terraform.test","server":["192.0.2.53"]}`, "opaque-123", false)
}

func assertSingleItem(t *testing.T, variables map[string]any) {
	t.Helper()
	for name, value := range variables {
		if name == "accountId" {
			continue
		}
		input, ok := value.(map[string]any)
		if !ok {
			continue
		}
		for _, key := range []string{"option", "relayGroup", "forwardingRule", "siteSettings", "site", "id"} {
			if items, ok := input[key].([]any); ok {
				require.Len(t, items, 1)
			}
		}
	}
}
