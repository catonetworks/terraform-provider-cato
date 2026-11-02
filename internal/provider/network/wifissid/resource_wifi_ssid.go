package wifissid

import (
	"context"
	"errors"
	"fmt"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = (*ssidResource)(nil)
	_ resource.ResourceWithConfigure   = (*ssidResource)(nil)
	_ resource.ResourceWithImportState = (*ssidResource)(nil)
)

type ssidResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the Wi-Fi SSID resource.
func NewResource() resource.Resource { return &ssidResource{} }
func (r *ssidResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *ssidResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *ssidResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wifi_ssid"
}
func (r *ssidResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}
func (r *ssidResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *ssidResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := buildCreate(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	occupied, err := r.slotOccupied(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check Wi-Fi interface", safeAPIError(err, secretValue(p)))
		return
	}
	if occupied {
		resp.Diagnostics.AddError("Wi-Fi interface occupied", "Import the existing SSID or use an unused Wi-Fi interface.")
		return
	}
	result, err := r.getClient().SiteCreateWifiSsid(ctx, r.client.AccountId, in)
	value := result.GetSite().GetCreateWifiSsid().GetWifiSsid()
	id := value.GetID()
	if id != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(p, id))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Wi-Fi SSID", safeMutationError(err, secretValue(p)))
		return
	}
	if !validID(types.StringValue(id), &resp.Diagnostics) {
		return
	}
	item := mutationRecord(value, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !matchesReference(in.Site, item.GetSite().GetID(), item.GetSite().GetName()) || item.GetNetworkInterfaceID() != in.NetworkInterfaceID {
		resp.Diagnostics.AddError("Invalid API response", "Created SSID has inconsistent site/interface ownership.")
		return
	}
	current := parseSSID(ctx, item, p, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *ssidResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() || !validID(p.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, p.ID.ValueString())
	if errors.Is(err, errStaleSSID) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Wi-Fi SSID", safeAPIError(err, secretValue(p)))
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseSSID(ctx, item, p, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *ssidResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var p, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	in := buildCreate(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SSID before update", safeAPIError(err, secretValue(state)))
		return
	}
	if existing == nil {
		resp.Diagnostics.AddError("SSID no longer exists", "Refresh state before applying again.")
		return
	}
	if !matchesReference(in.Site, existing.GetSite().GetID(), existing.GetSite().GetName()) ||
		in.NetworkInterfaceID != existing.GetNetworkInterfaceID() {
		resp.Diagnostics.AddError("Immutable SSID identity", "Site and network_interface_id changes require replacement.")
		return
	}
	result, err := r.getClient().SiteUpdateWifiSsid(ctx, r.client.AccountId, buildUpdate(in, state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Wi-Fi SSID", safeMutationError(err, secretValue(p)))
		return
	}
	item := mutationRecord(result.GetSite().GetUpdateWifiSsid().GetWifiSsid(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if item.GetID() != state.ID.ValueString() ||
		item.GetSite().GetID() != existing.GetSite().GetID() ||
		item.GetNetworkInterfaceID() != existing.GetNetworkInterfaceID() {
		resp.Diagnostics.AddError("Invalid API response", "Update returned a different SSID identity.")
		return
	}
	current := parseSSID(ctx, item, p, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *ssidResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() || !validID(p.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, p.ID.ValueString())
	if err != nil && !errors.Is(err, errStaleSSID) {
		resp.Diagnostics.AddError("Failed to read SSID before deletion", safeAPIError(err, secretValue(p)))
		return
	}
	if existing == nil && !errors.Is(err, errStaleSSID) {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().SiteRemoveWifiSsid(ctx, r.client.AccountId,
		m.RemoveWifiSsidInput{WifiSsid: &m.WifiSsidRefInput{By: m.ObjectRefByID,
			Input: p.ID.ValueString()}})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete Wi-Fi SSID", safeAPIError(err, secretValue(p)))
		return
	}
	item := mutationRecord(result.GetSite().GetRemoveWifiSsid().GetWifiSsid(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !removedIdentityMatches(item, p) {
		resp.Diagnostics.AddError("Invalid API response", "Deletion did not confirm the requested SSID identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
