package dnsserverset

import (
	"context"
	"fmt"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = (*dnsSetResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsSetResource)(nil)
	_ resource.ResourceWithImportState = (*dnsSetResource)(nil)
)

type dnsSetResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the DNS server-set resource.
func NewResource() resource.Resource { return &dnsSetResource{} }
func (r *dnsSetResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *dnsSetResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *dnsSetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_server_set"
}
func (r *dnsSetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *dnsSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *dnsSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreate(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().NetworkConfigDNSCreateServerSet(ctx, r.client.AccountId, input)
	items := result.GetNetworkConfig().GetDNS().GetCreateServerSet().GetDNSServerSet()
	id := firstID(items)
	if id != "" {
		plan.ID = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create DNS server set", err.Error())
		return
	}
	if len(items) != 1 || !validID(types.StringValue(id), &resp.Diagnostics) {
		resp.Diagnostics.AddError("Invalid API response", "Creation must return exactly one DNS set with an ID.")
		return
	}
	current := parseRecord(ctx, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *dnsSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read DNS server set", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseRecord(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *dnsSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	input := buildUpdate(ctx, plan, state.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().NetworkConfigDNSUpdateServerSet(ctx, r.client.AccountId, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update DNS server set", err.Error())
		return
	}
	items := result.GetNetworkConfig().GetDNS().GetUpdateServerSet().GetDNSServerSet()
	if len(items) != 1 || items[0].GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update must return exactly the requested DNS set.")
		return
	}
	current := parseRecord(ctx, items[0], &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *dnsSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read DNS server set before deletion", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().NetworkConfigDNSDeleteServerSet(ctx, r.client.AccountId,
		m.NetworkConfigDNSDeleteServerSetInput{ID: []string{state.ID.ValueString()}})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete DNS server set", err.Error())
		return
	}
	items := result.GetNetworkConfig().GetDNS().GetDeleteServerSet().GetDNSServerSet()
	if len(items) != 1 || items[0].GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Deletion did not confirm exactly the requested DNS set.")
		return
	}
	resp.State.RemoveResource(ctx)
}
