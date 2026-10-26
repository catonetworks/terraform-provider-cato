package dhcpsettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type settingsResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewDHCPSettingsResource returns the account DHCP settings resource.
func NewDHCPSettingsResource() resource.Resource { return &settingsResource{} }

func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_settings"
}

func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	nc.ImportAccount(ctx, r.client.AccountId, req, resp)
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
	if !nc.CheckAccount(r.client.AccountId, plan.ID, &resp.Diagnostics) {
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
	if !nc.CheckAccount(r.client.AccountId, state.ID, &resp.Diagnostics) {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpSettings(ctx, r.client.AccountId)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "settings", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.Dhcp.Settings != nil, &resp.Diagnostics) {
		return
	}
	item := result.NetworkConfig.Dhcp.Settings
	state.LeaseTime = types.Int64Value(item.LeaseTime)
	id, name := "", ""
	if item.Relay.Group != nil {
		id = item.Relay.Group.ID
		name = item.Relay.Group.Name
	}
	state.Relay = parseRelay(ctx, item.Relay.Enabled, item.Relay.Timeout, id, name, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *settingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *settingsResource) write(ctx context.Context, plan *SettingsModel, diags *diag.Diagnostics) {
	input := buildSettings(ctx, *plan, diags)
	if diags.HasError() {
		return
	}
	if input.Relay != nil && input.Relay.Enabled != nil && *input.Relay.Enabled && !completeRelay(ctx,
		r.resourceAPI(),
		r.client.AccountId,
		input.Relay,
		diags) {
		return
	}
	result, err := r.resourceAPI().NetworkConfigDhcpUpdateSettings(ctx, r.client.AccountId, input)
	if err != nil {
		nc.APIError(diags, "updateSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil &&
		result.NetworkConfig != nil &&
		result.NetworkConfig.Dhcp.UpdateSettings != nil &&
		result.NetworkConfig.Dhcp.UpdateSettings.Settings.LeaseTime >= 60,
		diags) {
		return
	}
	item := result.NetworkConfig.Dhcp.UpdateSettings.Settings
	plan.ID = types.StringValue(r.client.AccountId)
	plan.LeaseTime = types.Int64Value(item.LeaseTime)
	id, name := "", ""
	if item.Relay.Group != nil {
		id = item.Relay.Group.ID
		name = item.Relay.Group.Name
	}
	plan.Relay = parseRelay(ctx, item.Relay.Enabled, item.Relay.Timeout, id, name, diags)
}
