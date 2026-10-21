package appconnector

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/pops"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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

func (r *appConnectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	input := cato_models.AddZtnaAppConnectorInput{
		Description:               convert.KnownStringPointer(plan.Description),
		GroupName:                 plan.GroupName.ValueString(),
		Location:                  r.prepareLocation(ctx, plan.Location, &diags),
		Name:                      plan.Name.ValueString(),
		PooledBandwidthAllocation: r.prepareBwAllocation(ctx, plan.PooledBandwidthAllocation, &diags),
		PreferredPopLocation:      r.preparePopLocation(ctx, plan.PreferredPopLocation, &diags),
		Type:                      cato_models.ZtnaAppConnectorType(plan.Type.ValueString()),
	}

	// Call Cato API to create a new connector
	tflog.Debug(ctx, "AppConnectorCreateConnector", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.Catov2.AppConnectorCreateConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorCreateConnector", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppConnectorCreateConnector error", err.Error())
		return
	}

	// Set the ID from the response
	plan.ID = types.StringValue(result.GetZtnaAppConnector().GetAddZtnaAppConnector().GetZtnaAppConnector().GetID())

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydrateAppConnectorState(ctx, plan.ID.ValueString(), plan)
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.AddError("Error hydrating appConnector state", hydrateErr.Error())
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
	id := strings.Trim(state.ID.String(), `"`)
	if id == "" {
		resp.Diagnostics.AddError("AppConnectorUpdateConnector: ID is unknown", "AppConnector ID is not set in TF state")
		return
	}
	input := cato_models.UpdateZtnaAppConnectorInput{
		Description:          convert.KnownStringPointer(plan.Description),
		GroupName:            convert.KnownStringPointer(plan.GroupName),
		ID:                   id,
		Location:             r.prepareLocation(ctx, plan.Location, &diags),
		Name:                 convert.KnownStringPointer(plan.Name),
		PreferredPopLocation: r.preparePopLocation(ctx, plan.PreferredPopLocation, &diags),
	}

	tflog.Debug(ctx, "AppConnectorUpdateConnector", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.Catov2.AppConnectorUpdateConnector(ctx, r.client.AccountId, input)
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

func (r *appConnectorResource) UpdateBwLicenses(ctx context.Context, stateLcs, planLcs types.Set,
	appConID string, diags *diag.Diagnostics,
) {
	var tfStateLicenses, tfPlanLicenses []BandwidthAllocation
	diags.Append(stateLcs.ElementsAs(ctx, &tfStateLicenses, false)...)
	diags.Append(planLcs.ElementsAs(ctx, &tfPlanLicenses, false)...)
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

// updateBwLicenses - Call API to update licenses
func (r *appConnectorResource) updateBwLicenses(ctx context.Context, toUpdate map[string]int64, diags *diag.Diagnostics) {
	for allocID, bw := range toUpdate {
		input := cato_models.UpdateZtnaAppConnectorBwLicenseInput{
			AllocationID: allocID,
			Bw:           bw,
		}
		_, err := r.client.Catov2.ZtnaAppConnectorUpdateZtnaAppConnectorBwLicense(ctx, r.client.AccountId, input)
		if err != nil {
			diags.AddError(fmt.Sprintf("%s [bw=%d, allocation='%s']",
				"failed to update AppConnector Pooled bandwidth license ", bw, allocID), err.Error())
			return
		}
	}
}

// addBwLicenses - Call API to add licenses
func (r *appConnectorResource) addBwLicenses(ctx context.Context, toAdd map[licenseKey]string, appConID string,
	diags *diag.Diagnostics,
) {
	for k := range toAdd {
		input := cato_models.AddZtnaAppConnectorBwLicenseInput{
			Bw:        k.bw,
			LicenseID: k.lic,
			ZtnaAppConnector: &cato_models.ZtnaAppConnectorRefInput{
				By:    cato_models.ObjectRefByID,
				Input: appConID,
			},
		}
		_, err := r.client.Catov2.ZtnaAppConnectorAddZtnaAppConnectorBwLicense(ctx, r.client.AccountId, input)
		if err != nil {
			diags.AddError(fmt.Sprintf("%s [bw=%d, license='%s']",
				"failed to add AppConnector Pooled bandwidth license ", k.bw, k.lic), err.Error())
			return
		}
	}
}

// rmBwLicenses - Call API to remove licenses
func (r *appConnectorResource) rmBwLicenses(ctx context.Context, toRemove map[licenseKey]string, diags *diag.Diagnostics) {
	for _, allocID := range toRemove {
		input := cato_models.RemoveZtnaAppConnectorBwLicenseInput{AllocationID: allocID}
		_, err := r.client.Catov2.ZtnaAppConnectorRemoveZtnaAppConnectorBwLicense(ctx, r.client.AccountId, input)
		if err != nil {
			diags.AddError("failed to remove AppConnector Pooled bandwidth license "+allocID, err.Error())
			return
		}
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
	input := cato_models.RemoveZtnaAppConnectorInput{
		ZtnaAppConnector: &cato_models.ZtnaAppConnectorRefInput{
			By:    cato_models.ObjectRefByID,
			Input: state.ID.ValueString(),
		},
	}

	// Call Cato API to delete a connector
	tflog.Debug(ctx, "AppConnectorDeleteConnector", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.Catov2.AppConnectorDeleteConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorDeleteConnector", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppConnectorDeleteConnector error", err.Error())
		return
	}
}

func (r *appConnectorResource) prepareLocation(
	ctx context.Context,
	addr types.Object,
	diags *diag.Diagnostics,
) *cato_models.ZtnaAppConnectorLocationInput {
	if !utils.HasValue(addr) {
		return nil
	}

	var tfLocation Location
	diags.Append(addr.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	sdkLocation := cato_models.ZtnaAppConnectorLocationInput{
		Address:     convert.KnownStringPointer(tfLocation.Address),
		City:        tfLocation.CityName.ValueString(),
		CountryCode: tfLocation.CountryCode.ValueString(),
		StateCode:   convert.KnownStringPointer(tfLocation.StateCode),
		Timezone:    tfLocation.Timezone.ValueString(),
	}

	return &sdkLocation
}

func (r *appConnectorResource) prepareBwAllocation(ctx context.Context, bwAlloc types.Set, diags *diag.Diagnostics,
) []*cato_models.ZtnaAppConnectorPooledBandwidthAllocationInput {
	if !utils.HasValue(bwAlloc) {
		return []*cato_models.ZtnaAppConnectorPooledBandwidthAllocationInput{}
	}

	var tfBwAllocations []BandwidthAllocation
	diags.Append(bwAlloc.ElementsAs(ctx, &tfBwAllocations, false)...)
	if diags.HasError() {
		return nil
	}

	bwAllocations := make([]*cato_models.ZtnaAppConnectorPooledBandwidthAllocationInput, 0, len(tfBwAllocations))
	for _, tfBwAlloc := range tfBwAllocations {
		bwAllocations = append(bwAllocations,
			&cato_models.ZtnaAppConnectorPooledBandwidthAllocationInput{
				Bw:        tfBwAlloc.Bandwidth.ValueInt64(),
				LicenseID: tfBwAlloc.LicenseID.ValueString(),
			})
	}

	return bwAllocations
}

func (r *appConnectorResource) preparePopLocation(ctx context.Context, loc types.Object, diags *diag.Diagnostics,
) *cato_models.ZtnaAppConnectorPreferredPopLocationInput {
	if !utils.HasValue(loc) {
		return nil
	}

	var tfLocation pops.PreferredPopLocationModel
	diags.Append(loc.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	sdkLocation := cato_models.ZtnaAppConnectorPreferredPopLocationInput{
		PreferredOnly: tfLocation.PreferredOnly.ValueBool(),
		Automatic:     tfLocation.Automatic.ValueBool(),
		Primary:       idname.PrepareIDName[cato_models.PopLocationRefInput](ctx, tfLocation.Primary, diags),
		Secondary:     idname.PrepareIDName[cato_models.PopLocationRefInput](ctx, tfLocation.Secondary, diags),
	}

	return &sdkLocation
}

func (r *appConnectorResource) parseLocation(ctx context.Context, addr appConnectorLocation,
	diags *diag.Diagnostics,
) types.Object {
	// Prepare AppConnectorLocation object
	tfLocation := Location{
		Address:     types.StringPointerValue(addr.Address),
		CityName:    types.StringValue(addr.CityName),
		CountryCode: types.StringValue(addr.CountryCode),
		StateCode:   types.StringPointerValue(addr.StateCode),
		Timezone:    types.StringValue(addr.Timezone),
	}

	locObj, objDiags := types.ObjectValueFrom(ctx, LocationTypes, tfLocation)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(LocationTypes)
	}

	return locObj
}

func (r *appConnectorResource) parseBwAllocations(ctx context.Context,
	bwAllocs []*cato_go_sdk.AppConnectorReadConnector_ZtnaAppConnector_ZtnaAppConnector_PooledBandwidthAllocation,
	diags *diag.Diagnostics,
) types.Set {
	var objDiags diag.Diagnostics

	// Prepare BandwidthAllocation set
	var bwAllocSlice []attr.Value
	for _, bwa := range bwAllocs {
		tfBwAlloc := BandwidthAllocation{
			Bandwidth:    types.Int64Value(bwa.Bw),
			AllocationID: types.StringValue(bwa.AllocationID),
			LicenseID:    types.StringValue(bwa.LicenseID),
		}
		bwAllocObj, objDiags := types.ObjectValueFrom(ctx, BandwidthAllocationTypes, tfBwAlloc)
		diags.Append(objDiags...)
		bwAllocSlice = append(bwAllocSlice, bwAllocObj)
	}
	bwAllocSet, objDiags := types.SetValue(types.ObjectType{AttrTypes: BandwidthAllocationTypes}, bwAllocSlice)
	diags.Append(objDiags...)

	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: BandwidthAllocationTypes})
	}

	return bwAllocSet
}

func (r *appConnectorResource) parsePopLocation(ctx context.Context, loc *appConnectorPreferredPopLocation,
	diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics

	if loc == nil {
		return types.ObjectNull(pops.PreferredPopLocationModelTypes)
	}

	// Prepare PreferredPopLocationModel object
	tfLocation := pops.PreferredPopLocationModel{
		PreferredOnly: types.BoolValue(loc.PreferredOnly),
		Automatic:     types.BoolValue(loc.Automatic),
		Primary:       types.ObjectNull(idname.RefAttrTypes),
		Secondary:     types.ObjectNull(idname.RefAttrTypes),
	}
	if loc.Primary != nil {
		tfLocation.Primary = idname.ParseIDName(ctx, *loc.Primary, diags)
	}
	if loc.Secondary != nil {
		tfLocation.Secondary = idname.ParseIDName(ctx, *loc.Secondary, diags)
	}

	locObj, objDiags := types.ObjectValueFrom(ctx, pops.PreferredPopLocationModelTypes, tfLocation)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(pops.PreferredPopLocationModelTypes)
	}

	return locObj
}

// hydrateAppConnectorState fetches the current state of an appConnector from the API.
func (r *appConnectorResource) hydrateAppConnectorState(
	ctx context.Context,
	appConnectorID string,
	_ Model,
) (*Model, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	input := cato_models.ZtnaAppConnectorRefInput{
		By:    cato_models.ObjectRefByID,
		Input: appConnectorID,
	}

	// Call Cato API to get a connector
	tflog.Debug(ctx, "AppConnectorReadConnector", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.Catov2.AppConnectorReadConnector(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "AppConnectorReadConnector", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return nil, diags, err
	}

	// Map API response to AppConnectorModel
	con := result.GetZtnaAppConnector().GetZtnaAppConnector()
	if con == nil {
		return nil, diags, ErrAppConnectorNotFound
	}

	state := &Model{
		Description:               types.StringPointerValue(con.Description),
		GroupName:                 types.StringValue(con.GroupName),
		ID:                        types.StringValue(con.ID),
		Location:                  r.parseLocation(ctx, con.Location, &diags),
		Name:                      types.StringValue(con.Name),
		PooledBandwidthAllocation: r.parseBwAllocations(ctx, con.PooledBandwidthAllocation, &diags),
		PreferredPopLocation:      r.parsePopLocation(ctx, con.PreferredPopLocation, &diags),
		PrivateAppRef:             idname.ParseIDNameSet(ctx, con.PrivateAppRef, &diags),
		SerialNumber:              types.StringPointerValue(con.SerialNumber),
		SocketID:                  types.StringPointerValue(con.SocketID),
		SocketModel:               types.StringPointerValue((*string)(con.SocketModel)),
		Type:                      types.StringValue(con.Type.String()),
	}

	if diags.HasError() {
		return nil, diags, apperror.ErrAPIResponseParse
	}

	return state, nil, nil
}
