package dns

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

type settingsResource struct{ nc.Resource }

func NewSettingsResource() resource.Resource { return &settingsResource{} }
func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_settings"
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

	input := dnsInput(ctx, plan.PrimaryServer, plan.SecondaryServer, plan.Suffix, &resp.Diagnostics)
	input.AcceptDNSRequestsOnLanInterfaceIP = plan.AcceptDNSRequestsOnLanInterfaceIP.ValueBool()
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.DNS().UpdateDNSSettings(ctx, r.Client.AccountId, input)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.DNS.UpdateSettings != nil, &resp.Diagnostics) {
		return
	}
	item := &result.NetworkConfig.DNS.UpdateSettings.Settings
	if !nc.RequirePayload(item.Suffix != nil, &resp.Diagnostics) {
		return
	}
	plan.AcceptDNSRequestsOnLanInterfaceIP = types.BoolValue(item.AcceptDNSRequestsOnLanInterfaceIP)
	plan.ID = types.StringValue(r.Client.AccountId)
	hydrateDNS(ctx, item, &plan.PrimaryServer, &plan.SecondaryServer, &plan.Suffix, &resp.Diagnostics)
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
	input := dnsInput(ctx, plan.PrimaryServer, plan.SecondaryServer, plan.Suffix, &resp.Diagnostics)
	input.AcceptDNSRequestsOnLanInterfaceIP = plan.AcceptDNSRequestsOnLanInterfaceIP.ValueBool()
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.DNS().UpdateDNSSettings(ctx, r.Client.AccountId, input)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.DNS.UpdateSettings != nil, &resp.Diagnostics) {
		return
	}
	item := &result.NetworkConfig.DNS.UpdateSettings.Settings
	if !nc.RequirePayload(item.Suffix != nil, &resp.Diagnostics) {
		return
	}
	plan.AcceptDNSRequestsOnLanInterfaceIP = types.BoolValue(item.AcceptDNSRequestsOnLanInterfaceIP)
	plan.ID = types.StringValue(r.Client.AccountId)
	hydrateDNS(ctx, item, &plan.PrimaryServer, &plan.SecondaryServer, &plan.Suffix, &resp.Diagnostics)
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
	result, err := r.Client.DNS().NetworkConfigDNSSettings(ctx, r.Client.AccountId)
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
	state.AcceptDNSRequestsOnLanInterfaceIP = types.BoolValue(item.AcceptDNSRequestsOnLanInterfaceIP)
	hydrateDNS(ctx, item, &state.PrimaryServer, &state.SecondaryServer, &state.Suffix, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

// Account settings always exist. Destroy relinquishes ownership without resetting them.
func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
