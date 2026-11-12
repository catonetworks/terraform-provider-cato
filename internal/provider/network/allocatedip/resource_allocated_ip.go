package allocatedip

import (
	"context"
	"fmt"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &allocatedIPResource{}
	_ resource.ResourceWithConfigure   = &allocatedIPResource{}
	_ resource.ResourceWithImportState = &allocatedIPResource{}
)

func NewAllocatedIPResource() resource.Resource {
	return &allocatedIPResource{}
}

type allocatedIPResource struct {
	client *catoclient.Service
	// apiClient is the injectable SDK facade (see client_allocated_ip.go);
	// nil in production, replaced in unit tests.
	apiClient ResourceClient
}

func (r *allocatedIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocated_ip"
}

func (r *allocatedIPResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

// ImportState imports by the allocated IP ADDRESS (stable and unique within an
// account), hydrating the rest of the attributes in the subsequent Read.
func (r *allocatedIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), req.ID)...)
	// id is also required non-null in state before Read; Read resolves by ip
	// first, so seed it with the ip and let Read replace it with the real id.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// resolvePopLocationID resolves a PoP location to its id, matching the
// configured value against BOTH the pop's internal name and its displayName.
// The two differ for many pops (e.g. displayName "Paris_DC2" vs name "Paris2"),
// and the data surfaces are inconsistent: allocatedIpList returns the
// DISPLAY name in popLocation.name, while the addAllocatedIp mutation's
// {by: NAME} ref resolves only the INTERNAL name -- so a name captured from an
// export can be unresolvable by name. Mutating by ID avoids the mismatch
// entirely (pop ids are global Cato constants).
func (r *allocatedIPResource) resolvePopLocationID(ctx context.Context, popName string) (string, error) {
	result, err := r.resourceAPI().PopLocationList(ctx, r.client.AccountId, nil)
	if err != nil {
		return "", err
	}
	items := result.GetPopLocations().GetPopLocationList().GetItems()
	for _, item := range items {
		if item == nil {
			continue
		}
		if item.Name == popName || item.DisplayName == popName {
			return item.ID, nil
		}
	}
	return "", fmt.Errorf("pop location %q not found by name or displayName (%d pop locations listed)",
		popName, len(items))
}

// findAllocation looks an allocation up via the list query, preferring the IP
// (stable, also the import key) and falling back to the id.
func (r *allocatedIPResource) findAllocation(ctx context.Context, ip, id string,
) (*cato.PopLocationsAllocatedIpList_PopLocations_AllocatedIPList_AllocatedIP, error) {
	filter := &cato_models.PopLocationAllocatedIPFilterInput{}
	if ip != "" {
		filter.AllocatedIP = &cato_models.StringFilterInput{Eq: &ip}
	} else if id != "" {
		filter.ID = &cato_models.StringFilterInput{Eq: &id}
	}
	result, err := r.resourceAPI().AllocatedIPList(ctx, r.client.AccountId,
		&cato_models.PopLocationAllocatedIPInput{Filter: filter})
	tflog.Debug(ctx, "Read.PopLocationsAllocatedIPList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		return nil, err
	}
	list := result.GetPopLocations().GetAllocatedIPList()
	if list == nil {
		return nil, nil
	}
	for _, item := range list.AllocatedIP {
		if item == nil {
			continue
		}
		if (ip != "" && item.GetIP() == ip) || (ip == "" && id != "" && item.GetID() == id) {
			return item, nil
		}
	}
	return nil, nil
}

// findByIdentity returns a live allocation matching (pop_location_name,
// description) -- the cross-tenant identity pair -- or nil.
func (r *allocatedIPResource) findByIdentity(ctx context.Context, popLocation, description string,
) (*cato.PopLocationsAllocatedIpList_PopLocations_AllocatedIPList_AllocatedIP, error) {
	result, err := r.resourceAPI().AllocatedIPList(ctx, r.client.AccountId,
		&cato_models.PopLocationAllocatedIPInput{})
	if err != nil {
		return nil, err
	}
	list := result.GetPopLocations().GetAllocatedIPList()
	if list == nil {
		return nil, nil
	}
	for _, item := range list.AllocatedIP {
		if item == nil {
			continue
		}
		itemDesc := ""
		if item.GetDescription() != nil {
			itemDesc = *item.GetDescription()
		}
		if item.GetPopLocation().GetName() == popLocation && itemDesc == description {
			return item, nil
		}
	}
	return nil, nil
}

func (r *allocatedIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AllocatedIP
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create-or-adopt: an allocation with the same (pop_location_name,
	// description) identity may already exist -- allocation licenses are a
	// finite pool and Cato assigns the IPs, so re-allocating on every greenfield
	// run would exhaust the license (or fail once it is exhausted). Adopt the
	// existing allocation instead of allocating a duplicate.
	existing, identityErr := r.findByIdentity(ctx, plan.PopLocationName.ValueString(), plan.Description.ValueString())
	if identityErr == nil && existing != nil {
		tflog.Info(ctx, "adopting existing allocated IP with matching (popLocation, description)",
			map[string]interface{}{"ip": existing.GetIP(), "id": existing.GetID()})
		plan.ID = types.StringValue(existing.GetID())
		plan.IP = types.StringValue(existing.GetIP())
		plan.Name = types.StringValue(displayName(existing.GetPopLocation().GetName(), existing.GetIP()))
		plan.Type = types.StringValue(string(existing.AllocationType))
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		return
	}

	// Resolve the pop to its ID (matching name OR displayName) -- the by-NAME
	// mutation ref only resolves internal names, while exports carry display
	// names (e.g. "Paris_DC2" vs internal "Paris2").
	popID, err := r.resolvePopLocationID(ctx, plan.PopLocationName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve pop location", err.Error())
		return
	}

	input := cato_models.PopLocationAddAllocatedIPInput{
		Description: plan.Description.ValueStringPointer(),
		PopLocation: &cato_models.PopLocationRefInput{
			By:    cato_models.ObjectRefByID,
			Input: popID,
		},
		Type: cato_models.PopLocationAllocatedIPType(plan.Type.ValueString()),
	}
	tflog.Debug(ctx, "Create.PopLocationMutationsAddAllocatedIP.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	body, err := r.resourceAPI().AddAllocatedIP(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "Create.PopLocationMutationsAddAllocatedIP.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(body),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}

	allocated := body.GetPopLocationMutations().GetAddAllocatedIP().GetAllocatedIP()
	if allocated == nil || allocated.GetIP() == "" {
		resp.Diagnostics.AddError("Invalid allocate IP response",
			"API did not return the allocated IP")
		return
	}

	plan.ID = types.StringValue(allocated.GetID())
	plan.IP = types.StringValue(allocated.GetIP())
	plan.Name = types.StringValue(displayName(allocated.GetPopLocation().GetName(), allocated.GetIP()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *allocatedIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AllocatedIP
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := r.findAllocation(ctx, state.IP.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	if item == nil {
		// allocation released outside Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	applyAllocationToState(&state, item)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *allocatedIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AllocatedIP
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// pop_location_name and type RequiresReplace; only description updates in place.
	input := cato_models.PopLocationUpdateAllocatedIPDescriptionInput{
		Description: plan.Description.ValueString(),
		IP:          state.IP.ValueString(),
	}
	tflog.Debug(ctx, "Update.PopLocationMutationsUpdateAllocatedIPDescription.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	body, err := r.resourceAPI().UpdateAllocatedIPDescription(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "Update.PopLocationMutationsUpdateAllocatedIPDescription.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(body),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}

	// carry the computed values forward
	plan.ID = state.ID
	plan.IP = state.IP
	plan.Name = state.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *allocatedIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AllocatedIP
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.PopLocationRemoveAllocatedIPInput{
		IP:   state.IP.ValueString(),
		Type: cato_models.PopLocationAllocatedIPType(state.Type.ValueString()),
	}
	tflog.Debug(ctx, "Delete.PopLocationMutationsRemoveAllocatedIP.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	body, err := r.resourceAPI().RemoveAllocatedIP(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "Delete.PopLocationMutationsRemoveAllocatedIP.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(body),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
}
