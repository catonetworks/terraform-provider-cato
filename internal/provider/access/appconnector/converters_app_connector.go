package appconnector

import (
	"context"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	pops "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/pops"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

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

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	r *appConnectorResource,
	plan Model,
	originalDiags *diag.Diagnostics,
) (resultInput cato_models.AddZtnaAppConnectorInput) {
	diags := *originalDiags
	defer func() { *originalDiags = diags }()

	resultInput = cato_models.AddZtnaAppConnectorInput{
		Description:               convert.KnownStringPointer(plan.Description),
		GroupName:                 plan.GroupName.ValueString(),
		Location:                  r.prepareLocation(ctx, plan.Location, &diags),
		Name:                      plan.Name.ValueString(),
		PooledBandwidthAllocation: r.prepareBwAllocation(ctx, plan.PooledBandwidthAllocation, &diags),
		PreferredPopLocation:      r.preparePopLocation(ctx, plan.PreferredPopLocation, &diags),
		Type:                      cato_models.ZtnaAppConnectorType(plan.Type.ValueString()),
	}
	return resultInput
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	r *appConnectorResource,
	plan Model,
	originalDiags *diag.Diagnostics,
	state Model,
	diagnostics *diag.Diagnostics,
) (resultID string, resultInput cato_models.UpdateZtnaAppConnectorInput) {
	diags := *originalDiags
	defer func() { *originalDiags = diags }()

	resultID = strings.Trim(state.ID.String(), `"`)
	if resultID == "" {
		diagnostics.AddError("AppConnectorUpdateConnector: ID is unknown", "AppConnector ID is not set in TF state")
		return resultID, resultInput
	}
	resultInput = cato_models.UpdateZtnaAppConnectorInput{
		Description:          convert.KnownStringPointer(plan.Description),
		GroupName:            convert.KnownStringPointer(plan.GroupName),
		ID:                   resultID,
		Location:             r.prepareLocation(ctx, plan.Location, &diags),
		Name:                 convert.KnownStringPointer(plan.Name),
		PreferredPopLocation: r.preparePopLocation(ctx, plan.PreferredPopLocation, &diags),
	}
	return resultID, resultInput
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state Model) (resultInput cato_models.RemoveZtnaAppConnectorInput) {
	resultInput = cato_models.RemoveZtnaAppConnectorInput{
		ZtnaAppConnector: &cato_models.ZtnaAppConnectorRefInput{
			By:    cato_models.ObjectRefByID,
			Input: state.ID.ValueString(),
		},
	}
	return resultInput
}

// convertHydrateAppConnectorStateValues converts between the Terraform model and SDK values.
func convertHydrateAppConnectorStateValues(
	appConnectorID string,
) (resultInput cato_models.ZtnaAppConnectorRefInput) {
	resultInput = cato_models.ZtnaAppConnectorRefInput{
		By:    cato_models.ObjectRefByID,
		Input: appConnectorID,
	}
	return resultInput
}

// convertHydrateAppConnectorStateValues2 converts between the Terraform model and SDK values.
func convertHydrateAppConnectorStateValues2(
	ctx context.Context,
	r *appConnectorResource,
	originalDiags *diag.Diagnostics,
	con *cato_go_sdk.AppConnectorReadConnector_ZtnaAppConnector_ZtnaAppConnector,
) (resultState *Model) {
	diags := *originalDiags
	defer func() { *originalDiags = diags }()

	resultState = &Model{
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
	return resultState
}

// convertRmBwLicensesValues converts between the Terraform model and SDK values.
func convertRmBwLicensesValues(allocID string) (resultInput cato_models.RemoveZtnaAppConnectorBwLicenseInput) {
	resultInput = cato_models.RemoveZtnaAppConnectorBwLicenseInput{AllocationID: allocID}
	return resultInput
}

// convertAddBwLicensesValues converts between the Terraform model and SDK values.
func convertAddBwLicensesValues(appConID string, k licenseKey) (resultInput cato_models.AddZtnaAppConnectorBwLicenseInput) {
	resultInput = cato_models.AddZtnaAppConnectorBwLicenseInput{
		Bw:        k.bw,
		LicenseID: k.lic,
		ZtnaAppConnector: &cato_models.ZtnaAppConnectorRefInput{
			By:    cato_models.ObjectRefByID,
			Input: appConID,
		},
	}
	return resultInput
}

// convertUpdateBwLicensesValues converts between the Terraform model and SDK values.
func convertUpdateBandwidthInput(allocID string, bw int64) (resultInput cato_models.UpdateZtnaAppConnectorBwLicenseInput) {
	resultInput = cato_models.UpdateZtnaAppConnectorBwLicenseInput{
		AllocationID: allocID,
		Bw:           bw,
	}
	return resultInput
}

// convertUpdateBwLicensesValues converts between the Terraform model and SDK values.
func convertUpdateBwLicensesValues(
	ctx context.Context,
	stateLcs types.Set,
	planLcs types.Set,
	diags *diag.Diagnostics,
) (resultTfStateLicenses []BandwidthAllocation, resultTfPlanLicenses []BandwidthAllocation) {
	diags.Append(stateLcs.ElementsAs(ctx, &resultTfStateLicenses, false)...)
	diags.Append(planLcs.ElementsAs(ctx, &resultTfPlanLicenses, false)...)
	return resultTfStateLicenses, resultTfPlanLicenses
}
