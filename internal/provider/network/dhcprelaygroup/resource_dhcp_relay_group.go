package dhcprelaygroup

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type relayGroupResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewDHCPRelayGroupResource returns the standalone DHCP relay group resource.
func NewDHCPRelayGroupResource() resource.Resource { return &relayGroupResource{} }

func (r *relayGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_relay_group"
}

func (r *relayGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *relayGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RelayGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpCreateRelayGroup(ctx, r.client.AccountId,
		models.NetworkConfigDhcpCreateRelayGroupInput{RelayGroup: []*models.NetworkConfigDhcpRelayGroupCreateInput{&input}},
	)
	if id := createdID(result); id != "" {
		retained := plan
		retained.ID = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &retained)...)
	}

	if err != nil {
		nc.APIError(&resp.Diagnostics, "createRelayGroup", err)
		return
	}
	if !nc.RequirePayload(result != nil &&
		result.NetworkConfig != nil &&
		result.NetworkConfig.Dhcp.CreateRelayGroup != nil, &resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.Dhcp.CreateRelayGroup.RelayGroup
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].ID != "", &resp.Diagnostics) {
		return
	}
	parseRelayGroup(ctx, &plan, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *relayGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RelayGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpRelayGroup(ctx, r.client.AccountId,
		models.NetworkConfigDhcpRelayGroupRefInput{By: models.ObjectRefByID, Input: state.ID.ValueString()},
	)
	if err != nil {
		if nc.IsNotFound(err, "relayGroup") {
			resp.State.RemoveResource(ctx)
		} else {
			nc.APIError(&resp.Diagnostics, "relayGroup", err)
		}
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil, &resp.Diagnostics) {
		return
	}
	item := result.NetworkConfig.Dhcp.RelayGroup
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if !nc.RequirePayload(item.ID == state.ID.ValueString(), &resp.Diagnostics) {
		return
	}
	parseRelayGroup(ctx, &state, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *relayGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RelayGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpUpdateRelayGroup(ctx, r.client.AccountId,
		models.NetworkConfigDhcpUpdateRelayGroupInput{RelayGroup: []*models.NetworkConfigDhcpRelayGroupUpdateInput{&input}},
	)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateRelayGroup", err)
		return
	}
	if !nc.RequirePayload(result != nil &&
		result.NetworkConfig != nil &&
		result.NetworkConfig.Dhcp.UpdateRelayGroup != nil, &resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.Dhcp.UpdateRelayGroup.RelayGroup
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].ID == plan.ID.ValueString(), &resp.Diagnostics) {
		return
	}
	parseRelayGroup(ctx, &plan, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *relayGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RelayGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpDeleteRelayGroup(ctx, r.client.AccountId,
		models.NetworkConfigDhcpDeleteRelayGroupInput{ID: []string{state.ID.ValueString()}},
	)
	if err != nil {
		if !nc.IsNotFound(err, "deleteRelayGroup") && !r.confirmDeleted(ctx, state.ID.ValueString(), err) {
			nc.APIError(&resp.Diagnostics, "deleteRelayGroup", err)
		}
		return
	}
	nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.DeleteRelayGroup != nil, &resp.Diagnostics)
}

func (r *relayGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}
