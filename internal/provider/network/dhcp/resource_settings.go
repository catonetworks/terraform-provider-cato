package dhcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

type settingsResource struct{ nc.Resource }

func NewSettingsResource() resource.Resource { return &settingsResource{} }
func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_settings"
}
func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	nc.ImportAccount(ctx, &r.Resource, req, resp)
}
func (r *settingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}
func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !nc.CheckAccount(&r.Resource, plan.ID, &resp.Diagnostics) {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}
func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !nc.CheckAccount(&r.Resource, state.ID, &resp.Diagnostics) {
		return
	}
	result, err := r.Client.DHCP().NetworkConfigDhcpSettings(ctx, r.Client.AccountId)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "settings", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.Settings != nil, &resp.Diagnostics) {
		return
	}
	item := result.NetworkConfig.Dhcp.Settings
	state.LeaseTime = types.Int64Value(item.LeaseTime)
	state.RelayEnabled = types.BoolValue(item.Relay.Enabled)
	state.RelayTimeout = types.Int64PointerValue(item.Relay.Timeout)
	state.RelayGroupID = types.StringNull()
	if item.Relay.Group != nil {
		state.RelayGroupID = types.StringValue(item.Relay.Group.ID)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Account settings always exist; destroying ownership does not reset configuration.
func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
