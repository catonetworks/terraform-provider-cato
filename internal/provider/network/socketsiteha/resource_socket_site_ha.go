package socketsiteha

import (
	"context"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = &socketSiteHAResource{}
	_ resource.ResourceWithConfigure   = &socketSiteHAResource{}
	_ resource.ResourceWithImportState = &socketSiteHAResource{}
)

type SocketSiteHAClient interface {
	SiteUpdateHa(ctx context.Context, accountID string, siteID string, updateHaInput cato_models.UpdateHaInput,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.SiteUpdateHa, error)
}

type socketSiteHAResource struct {
	client             *catoclient.Service
	socketSiteHAClient SocketSiteHAClient
}

func NewSocketSiteHAResource() resource.Resource {
	return &socketSiteHAResource{}
}

func (r *socketSiteHAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_site_ha"
}

func (r *socketSiteHAResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *socketSiteHAResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), req.ID)...)
}

func (r *socketSiteHAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SocketSiteHA
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateCloudRouter(ctx, plan.SiteID.ValueString(), plan.IsCloudRouter.ValueBool(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.SiteID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *socketSiteHAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocketSiteHA
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// is_cloud_router is write-only in the API, so preserve the configured value.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *socketSiteHAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SocketSiteHA
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateCloudRouter(ctx, plan.SiteID.ValueString(), plan.IsCloudRouter.ValueBool(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.SiteID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *socketSiteHAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SocketSiteHA
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || state.IsCloudRouter.IsNull() || state.IsCloudRouter.IsUnknown() ||
		!state.IsCloudRouter.ValueBool() {
		return
	}

	r.updateCloudRouter(ctx, state.SiteID.ValueString(), false, &resp.Diagnostics)
}

func (r *socketSiteHAResource) getSocketSiteHAClient() SocketSiteHAClient {
	if r.socketSiteHAClient != nil {
		return r.socketSiteHAClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.Catov2
}

func (r *socketSiteHAResource) updateCloudRouter(ctx context.Context, siteID string, enabled bool, diags *diag.Diagnostics) {
	input := cato_models.UpdateHaInput{IsCloudRouter: types.BoolValue(enabled).ValueBoolPointer()}
	_, err := r.getSocketSiteHAClient().SiteUpdateHa(ctx, r.client.AccountId, siteID, input)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateHa error", err.Error())
	}
}
