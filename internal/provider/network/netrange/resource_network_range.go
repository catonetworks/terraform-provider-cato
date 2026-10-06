package netrange

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	errorsapp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr/application"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/client"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/utils"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/netrange/adapter"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/netrange/application"
)

var (
	_ resource.Resource                = &networkRangeResource{}
	_ resource.ResourceWithConfigure   = &networkRangeResource{}
	_ resource.ResourceWithImportState = &networkRangeResource{}
	_ resource.ResourceWithModifyPlan  = &networkRangeResource{}
)

func NewNetworkRangeResource() resource.Resource {
	return &networkRangeResource{}
}

type NetworkRangeClient = adapter.NetworkRangeClient
type networkRangeResource struct {
	client             *client.CatoClientData
	networkRangeClient NetworkRangeClient
}

func (r *networkRangeResource) getNetworkRangeClient() NetworkRangeClient {
	return adapter.ResolveClient(r.networkRangeClient, r.client)
}
func (r *networkRangeResource) sdk() adapter.SDK {
	return adapter.SDK{
		Client:    r.getNetworkRangeClient(),
		AccountID: r.client.AccountId,
	}
}
func (r *networkRangeResource) service() application.Service {
	return application.Service{
		Port: r.sdk(),
	}
}
func (r *networkRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.resourceSchema()
}
func (r *networkRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_range"
}

func (r *networkRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.CatoClientData)
}

// ModifyPlan ensures that exactly one of interface_id or interface_index is set,
// and if one changes in the config, the other one is marked as unknown.

func (r *networkRangeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, cfg *NetworkRange
	state := &NetworkRange{} // avoid nil pointer dereference
	stateDefined := !req.State.Raw.IsNull()

	if req.Plan.Raw.IsNull() { // resource destruction
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if stateDefined {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Validate config - ensure there is exactly one interface_index or interface_id.
	nrValidator := NetworkRangeValidator{}
	nrValidator.ValidateNetworkRangeWithPriorState(ctx, cfg, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rangeType := plan.RangeType.ValueString()

	// mdns reflector is only relevant for native, direct and vlan ranges
	if rangeType == "Routed" {
		plan.MdnsReflector = types.BoolNull()
		if utils.HasValue(cfg.MdnsReflector) {
			resp.Diagnostics.AddError("Invalid network range configuration",
				"mdns_reflector cannot be used when rangeType is 'Routed'")
			return
		}
	}

	// set interfaceIndex and interfaceID
	r.planInterfaceIDIndex(cfg, plan, state, stateDefined)

	// Gateway is only relevant for Routed range type
	plan.Gateway = types.StringNull()
	if rangeType == "Routed" {
		plan.Gateway = defaultPlanValue(cfg.Gateway, state.Gateway, stateDefined)
	}

	// Local IP is only relevant for Direct, Native and VLAN range types
	plan.LocalIP = types.StringNull()
	if rangeType != "Routed" {
		plan.LocalIP = defaultPlanValue(cfg.LocalIP, state.LocalIP, stateDefined)
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *networkRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Hydrate the state from the API
	var state NetworkRange
	state.ID = types.StringValue(req.ID)

	hydratedState, rangeExists := r.hydrateNetworkRangeState(ctx, nil, &state, req.ID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !rangeExists {
		resp.Diagnostics.AddError(
			"Network Range Not Found",
			fmt.Sprintf("Network range with ID %q not found during import", req.ID),
		)
		return
	}

	// Set the hydrated state
	diags := resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
}

// Create the network range resource
func (r *networkRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var cfg, plan *NetworkRange
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := r.prepareInput(ctx, cfg, plan, &resp.Diagnostics)
	read := r.readInput(ctx, cfg, plan, plan.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.service().Create(ctx, input, read)
	appendNetworkRangeError(&resp.Diagnostics, err)
	if resp.Diagnostics.HasError() {
		return
	}
	if !result.Found {
		resp.State.RemoveResource(ctx)
		return
	}
	state := r.projectState(ctx, cfg, plan, result.Snapshot, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var cfg, plan *NetworkRange
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := r.prepareInput(ctx, cfg, plan, &resp.Diagnostics)
	read := r.readInput(ctx, cfg, plan, plan.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.service().Update(ctx, input, read)
	appendNetworkRangeError(&resp.Diagnostics, err)
	if resp.Diagnostics.HasError() {
		return
	}
	if !result.Found {
		resp.State.RemoveResource(ctx)
		return
	}
	state := r.projectState(ctx, cfg, plan, result.Snapshot, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *NetworkRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, rangeExists := r.hydrateNetworkRangeState(ctx, nil, state, state.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !rangeExists {
		tflog.Warn(ctx, "siteRange not found, siteRange resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete removes the network range, treating an absent range as success.
func (r *networkRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkRange
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	appendNetworkRangeError(&resp.Diagnostics, r.service().Delete(ctx, state.ID.ValueString()))
}
func (r *networkRangeResource) hydrateNetworkRangeState(
	ctx context.Context,
	cfg, state *NetworkRange,
	id string,
	diags *diag.Diagnostics,
) (
	NetworkRange,
	bool,
) {
	in := r.readInput(ctx, cfg, state, id, diags)
	result, err := r.service().Read(ctx, in)
	appendNetworkRangeError(diags, err)
	if diags.HasError() || !result.Found {
		return NetworkRange{}, false
	}
	return r.projectState(ctx, cfg, state, result.Snapshot, diags), true
}
func (r *networkRangeResource) getSiteIDFromNetworkRange(ctx context.Context, id string) (siteID, interfaceName string, err error) {
	return r.sdk().SiteID(ctx, id)
}
func appendNetworkRangeError(diags *diag.Diagnostics, err error) {
	if err == nil {
		return
	}
	var operation *errorsapp.Error
	if errors.As(err, &operation) {
		diags.AddError(operation.Summary, err.Error())
		return
	}
	diags.AddError("Cato API error", err.Error())
}
