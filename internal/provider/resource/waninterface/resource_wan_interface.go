package waninterface

import (
	"context"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource              = &wanInterfaceResource{}
	_ resource.ResourceWithConfigure = &wanInterfaceResource{}
)

const (
	wanInterfaceIDParts                = 2
	wanInterfaceActiveNaturalOrder     = 1
	wanInterfacePassiveNaturalOrder    = 2
	wanInterfaceLastResortNaturalOrder = 3
)

func NewWanInterfaceResource() resource.Resource {
	return &wanInterfaceResource{}
}

type wanInterfaceResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *wanInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wan_interface"
}

func (r *wanInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *wanInterfaceResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Call Read to hydrate the full state from the API
	readReq := resource.ReadRequest{State: resp.State}
	readResp := resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, &readResp)

	// Copy diagnostics and state back to the import response
	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}

func (r *wanInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input
	input := convertCreateInput(plan)

	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.resourceAPI().SiteUpdateSocketInterface(
		ctx,
		plan.SiteID.ValueString(),
		cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()),
		input,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	hydratedState := convertCreateInput2(
		ctx,
		r,
		&plan,
		siteUpdateSocketInterfaceResponse,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WanInterface
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Split the ID to extract site_id and interface_id
	parts := strings.Split(state.ID.ValueString(), ":")
	if len(parts) != wanInterfaceIDParts {
		resp.Diagnostics.AddError(
			"Invalid WAN interface ID format",
			"Expected format 'site_id:interface_id', got: "+state.ID.ValueString(),
		)
		return
	}
	siteID := parts[0]
	interfaceID := parts[1]

	// Hydrate the state from the API
	hydratedState, interfaceExists, err := r.hydrateWanInterfaceState(ctx, state, siteID, interfaceID, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating WAN interface state during read",
			err.Error(),
		)
		return
	}

	if !interfaceExists {
		resp.Diagnostics.AddError(
			"WAN Interface Not Found",
			fmt.Sprintf("WAN interface with ID %q not found in site %q", interfaceID, siteID),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WanInterface
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input
	input := convertUpdateInput(plan)

	tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	siteUpdateSocketInterfaceResponse, err := r.resourceAPI().SiteUpdateSocketInterface(
		ctx,
		plan.SiteID.ValueString(),
		cato_models.SocketInterfaceIDEnum(plan.InterfaceID.ValueString()),
		input,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	hydratedState := convertUpdateInput2(
		ctx,
		r,
		&plan,
		siteUpdateSocketInterfaceResponse,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WanInterface
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.resourceAPI().EntityLookup(
		ctx,
		r.client.AccountId,
		cato_models.EntityType("site"),
		nil,
		nil,
		nil,
		nil,
		[]string{state.SiteID.ValueString()},
		nil,
		nil,
		nil,
	)
	tflog.Debug(ctx, "Delete.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	// check if site exist before removing
	if len(querySiteResult.EntityLookup.GetItems()) == 1 {
		// check if there is only one WAN interface & rewrite the input with default one
		accountSnapshotSite, err := r.resourceAPI().AccountSnapshot(ctx, []string{state.SiteID.ValueString()}, nil, true)
		tflog.Debug(ctx, "Delete.AccountSnapshot.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(accountSnapshotSite),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
			return
		}

		input := convertDeleteValues(state, accountSnapshotSite)

		tflog.Debug(ctx, "Delete.SiteUpdateSocketInterface.request", map[string]interface{}{
			"request": utils.InterfaceToJSONString(input),
		})
		_, err = r.resourceAPI().SiteUpdateSocketInterface(
			ctx,
			state.SiteID.ValueString(),
			cato_models.SocketInterfaceIDEnum(state.InterfaceID.ValueString()),
			input,
			r.client.AccountId,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteUpdateSocketInterface error",
				err.Error(),
			)
			return
		}
	}
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	ctx context.Context,
	r *wanInterfaceResource,
	priorPlan *WanInterface,
	siteUpdateSocketInterfaceResponse *cato_go_sdk.SiteUpdateSocketInterface,
	diagnostics *diag.Diagnostics,
) (resultHydratedState WanInterface) {
	plan := *priorPlan
	defer func() { *priorPlan = plan }()
	intID := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID.String())
	siteID := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID)
	plan.ID = types.StringValue(siteID.ValueString() + ":" + intID.ValueString())

	// Hydrate the state from the API to get the latest values including precedence
	resultHydratedState, interfaceExists, hydrateErr := r.hydrateWanInterfaceState(ctx, plan, siteID.ValueString(), intID.ValueString(), true)
	if hydrateErr != nil {
		diagnostics.AddError(
			"Error hydrating WAN interface state after update",
			hydrateErr.Error(),
		)
		return resultHydratedState
	}
	if !interfaceExists {
		diagnostics.AddError(
			"WAN Interface Not Found",
			fmt.Sprintf("WAN interface with ID %q not found after update", intID.ValueString()),
		)
		return resultHydratedState
	}
	return resultHydratedState
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	ctx context.Context,
	r *wanInterfaceResource,
	priorPlan *WanInterface,
	siteUpdateSocketInterfaceResponse *cato_go_sdk.SiteUpdateSocketInterface,
	diagnostics *diag.Diagnostics,
) (resultHydratedState WanInterface) {
	plan := *priorPlan
	defer func() { *priorPlan = plan }()
	intID := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID.String())
	siteID := types.StringValue(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID)
	plan.ID = types.StringValue(siteID.ValueString() + ":" + intID.ValueString())

	// Hydrate the state from the API to get the latest values including precedence
	resultHydratedState, interfaceExists, hydrateErr := r.hydrateWanInterfaceState(ctx, plan, siteID.ValueString(), intID.ValueString(), true)
	if hydrateErr != nil {
		diagnostics.AddError(
			"Error hydrating WAN interface state after create",
			hydrateErr.Error(),
		)
		return resultHydratedState
	}
	if !interfaceExists {
		diagnostics.AddError(
			"WAN Interface Not Found",
			fmt.Sprintf("WAN interface with ID %q not found after create", intID.ValueString()),
		)
		return resultHydratedState
	}
	return resultHydratedState
}

// hydrateWanInterfaceState fetches the current state of a WAN interface from the API
// and populates the state object with the latest values, including precedence mapping
//

func (r *wanInterfaceResource) hydrateWanInterfaceState(
	ctx context.Context,
	state WanInterface,
	siteID string,
	interfaceID string,
	forceRefresh bool,
) (WanInterface, bool, error) {
	// Get accountSnapshot data to check if site exists and has interfaces
	siteAccountSnapshotData, err := r.resourceAPI().AccountSnapshot(ctx, []string{siteID}, nil, forceRefresh)
	tflog.Debug(ctx, "hydrateWanInterfaceState.AccountSnapshot.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteAccountSnapshotData),
	})
	if err != nil {
		return WanInterface{}, false, fmt.Errorf("catov2 API AccountSnapshot error: %w", err)
	}

	// Check if site exists
	if len(siteAccountSnapshotData.AccountSnapshot.Sites) == 0 {
		return state, false, nil
	}

	site := siteAccountSnapshotData.AccountSnapshot.Sites[0]

	// Find the interface in InfoSiteSnapshot.Interfaces
	foundInterfaceIndex := -1
	for i := range site.InfoSiteSnapshot.Interfaces {
		iface := site.InfoSiteSnapshot.Interfaces[i]
		if "INT_"+iface.ID == interfaceID || iface.ID == interfaceID {
			foundInterfaceIndex = i
			break
		}
	}

	if foundInterfaceIndex == -1 {
		return state, false, nil
	}

	foundInterface := convertHydrateWanInterfaceStateValues(&state, siteID, site, foundInterfaceIndex)

	// Map precedence from naturalOrder in devices.interfaces
	// naturalOrder: 1 = ACTIVE, 2 = PASSIVE, 3 = LAST_RESORT
	if len(site.Devices) > 0 {
		for _, device := range site.Devices {
			for _, deviceIface := range device.Interfaces {
				// Match interface by ID - deviceIface.ID is a pointer to string
				deviceIfaceID := ""
				if deviceIface.ID != nil {
					deviceIfaceID = *deviceIface.ID
				}
				if wanInterfaceIDsMatch(foundInterface.ID, deviceIfaceID, "") {
					mappedPrecedence := precedenceFromNaturalOrder(deviceIface.NaturalOrder)
					if !mappedPrecedence.IsNull() && !mappedPrecedence.IsUnknown() {
						state.Precedence = mappedPrecedence
					}
					break
				}
			}
		}
	}

	return state, true, nil
}
