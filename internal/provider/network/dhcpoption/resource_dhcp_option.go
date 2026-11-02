package dhcpoption

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type optionResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewDHCPOptionResource returns the standalone DHCP option resource.
func NewDHCPOptionResource() resource.Resource { return &optionResource{} }

func (r *optionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_option"
}

func (r *optionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *optionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpCreateOption(ctx, r.client.AccountId,
		models.NetworkConfigDhcpCreateOptionInput{Option: []*models.NetworkConfigDhcpOptionCreateInput{&input}},
	)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "createOption", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.CreateOption != nil,
		&resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.Dhcp.CreateOption.Option
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].ID != "", &resp.Diagnostics) {
		return
	}
	parseOption(ctx, &plan, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *optionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpOption(ctx, r.client.AccountId,
		models.NetworkConfigDhcpOptionRefInput{By: models.ObjectRefByID, Input: state.ID.ValueString()},
	)
	if err != nil {
		if nc.IsNotFound(err, "option") {
			resp.State.RemoveResource(ctx)
		} else {
			nc.APIError(&resp.Diagnostics, "option", err)
		}
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil, &resp.Diagnostics) {
		return
	}
	item := result.NetworkConfig.Dhcp.Option
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if !nc.RequirePayload(item.ID == state.ID.ValueString(), &resp.Diagnostics) {
		return
	}
	parseOption(ctx, &state, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *optionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpUpdateOption(ctx, r.client.AccountId,
		models.NetworkConfigDhcpUpdateOptionInput{Option: []*models.NetworkConfigDhcpOptionUpdateInput{&input}},
	)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateOption", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.UpdateOption != nil,
		&resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.Dhcp.UpdateOption.Option
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].ID == plan.ID.ValueString(), &resp.Diagnostics) {
		return
	}
	parseOption(ctx, &plan, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *optionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpDeleteOption(ctx, r.client.AccountId,
		models.NetworkConfigDhcpDeleteOptionInput{ID: []string{state.ID.ValueString()}},
	)
	if err != nil {
		if !nc.IsNotFound(err, "deleteOption") {
			nc.APIError(&resp.Diagnostics, "deleteOption", err)
		}
		return
	}
	nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.DeleteOption != nil, &resp.Diagnostics)
}

func (r *optionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}
