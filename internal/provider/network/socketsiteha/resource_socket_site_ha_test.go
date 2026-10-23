package socketsiteha

import (
	"context"
	"errors"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

type recordingSocketSiteHAClient struct {
	isCloudRouter *bool
	err           error
}

func (c *recordingSocketSiteHAClient) SiteUpdateHa(
	_ context.Context,
	_ string,
	_ string,
	input cato_models.UpdateHaInput,
	_ ...clientv2.RequestInterceptor,
) (*cato_go_sdk.SiteUpdateHa, error) {
	c.isCloudRouter = input.IsCloudRouter
	return &cato_go_sdk.SiteUpdateHa{}, c.err
}

func TestSocketSiteHASchema(t *testing.T) {
	t.Parallel()

	resp := &resource.SchemaResponse{}
	(&socketSiteHAResource{}).Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Schema.Attributes["site_id"] == nil {
		t.Fatal("expected site_id attribute")
	}
	if resp.Schema.Attributes["is_cloud_router"] == nil {
		t.Fatal("expected is_cloud_router attribute")
	}
}

func TestSocketSiteHAUpdateCloudRouter(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"enabled":  true,
		"disabled": false,
	}

	for name, enabled := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := &recordingSocketSiteHAClient{}
			r := &socketSiteHAResource{
				client:             &catoclient.Service{AccountId: "account-123"},
				socketSiteHAClient: client,
			}
			var diags diag.Diagnostics

			r.updateCloudRouter(context.Background(), "site-123", enabled, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %+v", diags)
			}
			if client.isCloudRouter == nil || *client.isCloudRouter != enabled {
				t.Fatalf("expected is_cloud_router=%t, got %v", enabled, client.isCloudRouter)
			}
		})
	}
}

func TestSocketSiteHAUpdateCloudRouterError(t *testing.T) {
	t.Parallel()

	client := &recordingSocketSiteHAClient{err: errors.New("update ha failed")}
	r := &socketSiteHAResource{
		client:             &catoclient.Service{AccountId: "account-123"},
		socketSiteHAClient: client,
	}
	var diags diag.Diagnostics

	r.updateCloudRouter(context.Background(), "site-123", true, &diags)

	if !diags.HasError() {
		t.Fatal("expected diagnostics")
	}
}

func TestSocketSiteHADeleteDisablesCloudRouter(t *testing.T) {
	t.Parallel()

	client := &recordingSocketSiteHAClient{}
	r := &socketSiteHAResource{
		client:             &catoclient.Service{AccountId: "account-123"},
		socketSiteHAClient: client,
	}
	state := newSocketSiteHAState(context.Background(), t)
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	if client.isCloudRouter == nil || *client.isCloudRouter {
		t.Fatal("expected delete to disable Cloud Router mode")
	}
}

func newSocketSiteHAState(ctx context.Context, t *testing.T) tfsdk.State {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	(&socketSiteHAResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}
	diags := state.Set(ctx, SocketSiteHA{
		ID:            types.StringValue("site-123"),
		SiteID:        types.StringValue("site-123"),
		IsCloudRouter: types.BoolValue(true),
	})
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	return state
}
