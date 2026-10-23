package static_host_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	statichost "github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resources/static_host"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resourcestest"
)

func TestStaticHostCreateThroughProvider(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, mac := range []types.String{types.StringNull(), types.StringValue("00:11:22:33:44:55")} {
		calls := 0
		r, resourceSchema := resourcestest.ConfiguredResource(t, "cato_static_host", func(request resourcestest.GraphQLRequest) string {
			calls++
			require.Equal(t, "siteAddStaticHost", request.OperationName)
			require.JSONEq(t, `"site-123"`, string(request.Variables["siteId"]))
			require.JSONEq(t, `"account-123"`, string(request.Variables["accountId"]))
			return `{"data":{"site":{"addStaticHost":{"hostId":"host-123"}}}}`
		})
		model := statichost.Model{ID: types.StringUnknown(), SiteID: types.StringValue("site-123"),
			Name: types.StringValue("printer"), IP: types.StringValue("192.0.2.10"), MacAddress: mac}
		plan := tfsdk.Plan{Schema: resourceSchema}
		require.False(t, plan.Set(ctx, model).HasError())
		response := resource.CreateResponse{State: tfsdk.State{Schema: resourceSchema}}
		r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
		require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
		var got statichost.Model
		require.False(t, response.State.Get(ctx, &got).HasError())
		model.ID = types.StringValue("host-123")
		require.Equal(t, model, got)
		require.Equal(t, 1, calls)
	}
}

func TestStaticHostCreateRejectsUnknownValuesBeforeAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r, resourceSchema := resourcestest.ConfiguredResource(t, "cato_static_host", func(resourcestest.GraphQLRequest) string {
		t.Fatal("unexpected API call")
		return ""
	})
	plan := tfsdk.Plan{Schema: resourceSchema}
	require.False(t, plan.Set(ctx, statichost.Model{ID: types.StringUnknown(), SiteID: types.StringUnknown(),
		Name: types.StringValue("printer"), IP: types.StringValue("192.0.2.10"), MacAddress: types.StringNull()}).HasError())
	response := resource.CreateResponse{State: tfsdk.State{Schema: resourceSchema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
	require.True(t, response.Diagnostics.HasError())
}
