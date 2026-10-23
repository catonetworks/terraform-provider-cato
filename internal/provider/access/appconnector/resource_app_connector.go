package appconnector

import (
	"context"
	"errors"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	apperror "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &appConnectorResource{}
	_ resource.ResourceWithConfigure   = &appConnectorResource{}
	_ resource.ResourceWithImportState = &appConnectorResource{}

	ErrAppConnectorNotFound = errors.New("app-connector not found")
)

func NewAppConnectorResource() resource.Resource {
	return &appConnectorResource{}
}

type appConnectorResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *appConnectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_connector"
}

func (r *appConnectorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *appConnectorResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a new app connector
func (r *appConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(
		ctx,
		r,
		plan,
		&diags,
	)

	// Call Cato API to create a new connector
	tflog.Debug(ctx, "AppConnectorCreateConnector", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().AppConnectorCreateConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorCreateConnector", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppConnectorCreateConnector error", err.Error())
		return
	}

	// Set the ID from the response
	hydratedState := convertCreateInput2(
		ctx,
		r,
		&plan,
		&diags,
		result,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read app connector data from Cato API
func (r *appConnectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, diags, hydrateErr := r.hydrateAppConnectorState(ctx, state.ID.ValueString(), state)
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		// Check if app-connector was found
		if errors.Is(hydrateErr, ErrAppConnectorNotFound) {
			tflog.Warn(ctx, "app_connector not found, resource removed")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error hydrating group state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update app connector configuration
func (r *appConnectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Model
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, input := convertUpdateInput(
		ctx,
		r,
		plan,
		&diags,
		state,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "AppConnectorUpdateConnector", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().AppConnectorUpdateConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorUpdateConnector", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})

	if err != nil {
		resp.Diagnostics.AddError("Cato API AppConnectorUpdateConnector error", err.Error())
		return
	}

	r.UpdateBwLicenses(ctx, state.PooledBandwidthAllocation, plan.PooledBandwidthAllocation, id, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydrateAppConnectorState(ctx, id, plan)
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.AddError("Error hydrating app-connector state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete app connector
func (r *appConnectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := convertDeleteInput(state)

	// Call Cato API to delete a connector
	tflog.Debug(ctx, "AppConnectorDeleteConnector", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.resourceAPI().AppConnectorDeleteConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorDeleteConnector", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppConnectorDeleteConnector error", err.Error())
		return
	}
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	ctx context.Context,
	r *appConnectorResource,
	priorPlan *Model,
	priorDiags *diag.Diagnostics,
	result *cato_go_sdk.AppConnectorCreateConnector,
	diagnostics *diag.Diagnostics,
) (resultHydratedState *Model) {
	plan := *priorPlan
	defer func() { *priorPlan = plan }()
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	plan.ID = types.StringValue(result.GetZtnaAppConnector().GetAddZtnaAppConnector().GetZtnaAppConnector().GetID())

	// Hydrate state from API
	resultHydratedState, diags, hydrateErr := r.hydrateAppConnectorState(ctx, plan.ID.ValueString(), plan)
	if hydrateErr != nil {
		diagnostics.Append(diags...)
		diagnostics.AddError("Error hydrating appConnector state", hydrateErr.Error())
		return resultHydratedState
	}
	return resultHydratedState
}

// hydrateAppConnectorState fetches the current state of an appConnector from the API.
func (r *appConnectorResource) hydrateAppConnectorState(
	ctx context.Context,
	appConnectorID string,
	_ Model,
) (*Model, diag.Diagnostics, error) {
	var diags diag.Diagnostics
	input := convertHydrateAppConnectorStateValues(appConnectorID)

	// Call Cato API to get a connector
	tflog.Debug(ctx, "AppConnectorReadConnector", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().AppConnectorReadConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorReadConnector", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return nil, diags, err
	}

	// Map API response to AppConnectorModel
	con := result.GetZtnaAppConnector().GetZtnaAppConnector()
	if con == nil {
		return nil, diags, ErrAppConnectorNotFound
	}

	state := convertHydrateAppConnectorStateValues2(
		ctx,
		r,
		&diags,
		con,
	)

	if diags.HasError() {
		return nil, diags, apperror.ErrAPIResponseParse
	}

	return state, nil, nil
}

// rmBwLicenses - Call API to remove licenses
func (r *appConnectorResource) rmBwLicenses(ctx context.Context, toRemove map[licenseKey]string, diags *diag.Diagnostics) {
	for _, allocID := range toRemove {
		input := convertRmBwLicensesValues(allocID)
		_, err := r.resourceAPI().ZtnaAppConnectorRemoveZtnaAppConnectorBwLicense(ctx, r.client.AccountId, input)
		if err != nil {
			diags.AddError("failed to remove AppConnector Pooled bandwidth license "+allocID, err.Error())
			return
		}
	}
}

// addBwLicenses - Call API to add licenses
func (r *appConnectorResource) addBwLicenses(ctx context.Context, toAdd map[licenseKey]string, appConID string,
	diags *diag.Diagnostics,
) {
	for k := range toAdd {
		input := convertAddBwLicensesValues(appConID, k)
		_, err := r.resourceAPI().ZtnaAppConnectorAddZtnaAppConnectorBwLicense(ctx, r.client.AccountId, input)
		if err != nil {
			diags.AddError(fmt.Sprintf("%s [bw=%d, license='%s']",
				"failed to add AppConnector Pooled bandwidth license ", k.bw, k.lic), err.Error())
			return
		}
	}
}

// updateBwLicenses - Call API to update licenses
func (r *appConnectorResource) updateBwLicenses(ctx context.Context, toUpdate map[string]int64, diags *diag.Diagnostics) {
	for allocID, bw := range toUpdate {
		input := convertUpdateBandwidthInput(allocID, bw)
		_, err := r.resourceAPI().ZtnaAppConnectorUpdateZtnaAppConnectorBwLicense(ctx, r.client.AccountId, input)
		if err != nil {
			diags.AddError(fmt.Sprintf("%s [bw=%d, allocation='%s']",
				"failed to update AppConnector Pooled bandwidth license ", bw, allocID), err.Error())
			return
		}
	}
}

func (r *appConnectorResource) UpdateBwLicenses(ctx context.Context, stateLcs, planLcs types.Set,
	appConID string, diags *diag.Diagnostics,
) {
	tfStateLicenses, tfPlanLicenses := convertUpdateBwLicensesValues(ctx, stateLcs, planLcs, diags)
	if diags.HasError() {
		return
	}

	// Map of currently used licenses
	stateMap := make(map[licenseKey]string) // [license+bandwidth]=>allocationID
	for _, l := range tfStateLicenses {
		k := licenseKey{lic: l.LicenseID.ValueString(), bw: l.Bandwidth.ValueInt64()}
		stateMap[k] = l.AllocationID.ValueString()
	}
	// Map of planned licenses
	planMap := make(map[licenseKey]string) // [license+bandwidth]=>allocationID
	for _, l := range tfPlanLicenses {
		k := licenseKey{lic: l.LicenseID.ValueString(), bw: l.Bandwidth.ValueInt64()}
		planMap[k] = l.AllocationID.ValueString()
	}

	// Find licenses to remove
	var toRemove = make(map[licenseKey]string) // [license+bandwidth]=>allocationID
	for k := range stateMap {
		if _, ok := planMap[k]; !ok { // we have it now, but we don't plan to keep it
			toRemove[k] = stateMap[k]
		}
	}

	// Find licenses to add
	var toAdd = make(map[licenseKey]string) // [license+bandwidth]=>allocationID
	for k := range planMap {
		if _, ok := stateMap[k]; !ok { // we don't have it now, but we plan it
			toAdd[k] = stateMap[k]
		}
	}

	// Find licenses that can be updated
	var toUpdate = make(map[string]int64) // [allocationID]=>bandwidth
	for addKey := range toAdd {
		for rmKey, allocID := range toRemove {
			if addKey.lic == rmKey.lic {
				toUpdate[allocID] = addKey.bw
				delete(toRemove, rmKey)
				delete(toAdd, addKey)
				break
			}
		}
	}

	// Call APIs
	r.updateBwLicenses(ctx, toUpdate, diags)
	r.addBwLicenses(ctx, toAdd, appConID, diags)
	r.rmBwLicenses(ctx, toRemove, diags)
}
