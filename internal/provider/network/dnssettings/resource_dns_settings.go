package dnssettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type settingsResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewDNSSettingsResource returns the account DNS settings resource.
func NewDNSSettingsResource() resource.Resource { return &settingsResource{} }

func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_settings"
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

	input := buildSettings(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().UpdateDNSSettings(ctx, r.client.AccountId, input)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.DNS.UpdateSettings != nil,
		&resp.Diagnostics) {
		return
	}
	item := &result.NetworkConfig.DNS.UpdateSettings.Settings
	if !nc.RequirePayload(item.Suffix != nil, &resp.Diagnostics) {
		return
	}
	plan = parseSettings(ctx, item, r.client.AccountId, &resp.Diagnostics)
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
	input := buildSettings(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().UpdateDNSSettings(ctx, r.client.AccountId, input)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.DNS.UpdateSettings != nil,
		&resp.Diagnostics) {
		return
	}
	item := &result.NetworkConfig.DNS.UpdateSettings.Settings
	if !nc.RequirePayload(item.Suffix != nil, &resp.Diagnostics) {
		return
	}
	plan = parseSettings(ctx, item, r.client.AccountId, &resp.Diagnostics)
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
	result, err := r.resourceAPI().NetworkConfigDNSSettings(ctx, r.client.AccountId)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "settings", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil, &resp.Diagnostics) {
		return
	}
	item := result.NetworkConfig.DNS.Settings
	if !nc.RequirePayload(item != nil && item.Suffix != nil, &resp.Diagnostics) {
		return
	}
	state = parseSettings(ctx, item, r.client.AccountId, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *settingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}
