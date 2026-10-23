package bgppeer

import (
	"context"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &bgpPeerResource{}
	_ resource.ResourceWithConfigure   = &bgpPeerResource{}
	_ resource.ResourceWithImportState = &bgpPeerResource{}
)

func NewBgpPeerResource() resource.Resource {
	return &bgpPeerResource{}
}

type bgpPeerResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *bgpPeerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bgp_peer"
}

func (r *bgpPeerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *bgpPeerResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *bgpPeerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BgpPeer
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Create.SiteAddBgpPeer.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	addBgpPeerPayload, err := r.resourceAPI().SiteAddBgpPeer(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddBgpPeer.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(addBgpPeerPayload),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteAddBgpPeer",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bgpPeer := addBgpPeerPayload.GetSite().GetAddBgpPeer().GetBgpPeer()
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), bgpPeer.GetID())
}

func (r *bgpPeerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BgpPeer
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bgpPeerRefInput := convertReadState(state)
	tflog.Debug(ctx, "Read.SiteBgpPeer.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(bgpPeerRefInput),
	})
	result, err := r.resourceAPI().SiteBgpPeer(ctx, bgpPeerRefInput, r.client.AccountId)
	tflog.Debug(ctx, "Read.SiteAddBgpPeer.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteBgpPeer",
			err.Error(),
		)
		return
	}

	if result == nil || result.GetSite() == nil || result.GetSite().GetBgpPeer() == nil {
		tflog.Warn(ctx, "BGP peer resource wasn't found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	bgpPeer := result.GetSite().GetBgpPeer()
	bgpPeerInput := ConvertToBgpPeer(*bgpPeer)

	diags = resp.State.Set(ctx, &bgpPeerInput)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bgpPeerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BgpPeer
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertUpdateInput(ctx, plan, &diags, &resp.Diagnostics)

	tflog.Debug(ctx, "Update.SiteUpdateBgpPeer.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	SiteUpdateBgpPeerResponse, err := r.resourceAPI().SiteUpdateBgpPeer(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteAddBgpPeer.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(SiteUpdateBgpPeerResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteUpdateBgpPeer",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bgpPeerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BgpPeer
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeBgpPeerInput := convertDeleteInput(state)

	tflog.Debug(ctx, "Delete.SiteUpdateBgpPeer.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeBgpPeerInput),
	})
	SiteRemoveBgpPeerResponse, err := r.resourceAPI().SiteRemoveBgpPeer(ctx, removeBgpPeerInput, r.client.AccountId)
	tflog.Debug(ctx, "Delete.SiteUpdateBgpPeer.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(SiteRemoveBgpPeerResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteRemoveBgpPeer",
			err.Error(),
		)
		return
	}
}
