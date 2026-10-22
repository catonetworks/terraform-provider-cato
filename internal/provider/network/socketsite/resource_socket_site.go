package socketsite

import (
	"context"
	"fmt"
	"regexp"
	"time"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

const (
	socketConnectionTypeAWS1500  = "SOCKET_AWS1500"
	socketConnectionTypeAZ1500   = "SOCKET_AZ1500"
	socketConnectionTypeESX1500  = "SOCKET_ESX1500"
	socketConnectionTypeX1600    = "SOCKET_X1600"
	socketConnectionTypeX1600LTE = "SOCKET_X1600_LTE"
	socketConnectionTypeX1700    = "SOCKET_X1700"
	socketConnectionTypeVGXAWS   = "VSOCKET_VGX_AWS"
	socketConnectionTypeVGXAzure = "VSOCKET_VGX_AZURE"
	socketConnectionTypeVGXESX   = "VSOCKET_VGX_ESX"
	socketInterfaceDestTypeLAN   = "LAN"
	socketCreateHydrationRetries = 6
)

var (
	_ resource.Resource                = &socketSiteResource{}
	_ resource.ResourceWithConfigure   = &socketSiteResource{}
	_ resource.ResourceWithImportState = &socketSiteResource{}
)

func NewSocketSiteResource() resource.Resource {
	return &socketSiteResource{}
}

type socketSiteResource struct {
	client           *catoclient.Service
	socketSiteClient Client

	apiClient ResourceClient
}

var numberRE = regexp.MustCompile(`^\d+$`)

func (r *socketSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_site"
}

func (r *socketSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *socketSiteResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// TODO: check if this is really necessary
	// Call Read to hydrate the full state from the API
	readReq := resource.ReadRequest{State: resp.State}
	readResp := resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, &readResp)

	// Copy diagnostics and state back to the import response
	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}

// Create cato_socket_site resource
//
//nolint:funlen // create flow composes multiple API calls and hydration checks in sequence.
func (r *socketSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, cfg SocketSite
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.validateCloudRouterInput(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a socket site - API call
	siteID := r.createBasicSocketSite(ctx, &plan, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	r.configureCloudRouterAfterCreate(ctx, &plan, siteID, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Find Network Range ID for the created site and update the NetworkRange details
	networkRange := r.findNativeRange(ctx, siteID, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	r.updateNetworkRange(ctx, &cfg, &plan, networkRange.GetNetworkRangeID(), false, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Move interface to a non-default index if specified in the plan
	defaultInterfaceIndex := InterfaceByConnType[cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())]
	r.assignInterfaceIndex(ctx, defaultInterfaceIndex, &plan, siteID, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// Update Socket Interface
	r.updateSocketInterface(ctx, &cfg, &plan, siteID, false, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// hydrate the state with API data, retrying briefly for eventual consistency
	var (
		hydratedState SocketSite
		siteExists    bool
	)
	for attempt := 0; attempt < socketCreateHydrationRetries; attempt++ {
		attemptDiags := diag.Diagnostics{}
		hydratedState, siteExists = r.hydrateSocketSiteState(ctx, &cfg, plan, siteID, &attemptDiags)
		if attemptDiags.HasError() {
			if attempt == socketCreateHydrationRetries-1 {
				resp.Diagnostics.Append(attemptDiags...)
				return
			}
			time.Sleep(2 * time.Second)
			continue
		}
		if siteExists {
			break
		}
		time.Sleep(2 * time.Second)
	}

	// if the site is still not visible, keep minimal state to avoid losing created resource
	if !siteExists {
		fallbackState := convertCreateValues(cfg, siteID)
		tflog.Warn(ctx, "site not found after create; preserving minimal state until next refresh")
		diags = resp.State.Set(ctx, &fallbackState)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overriding state with socket site id
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), siteID)
}

// Read cato_socket_site resource
func (r *socketSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists := r.hydrateSocketSiteState(ctx, nil, state, state.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update cato_socket_site resource
func (r *socketSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var cfg, plan, state SocketSite
	var planNativeRange, stateNativeRange NativeRange

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := plan.ID.ValueString()
	isHA := r.isHA(&state)

	// Update general socket site details
	r.updateBasicSocketSite(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update network range details
	if utils.CheckErr(&diags, plan.NativeRange.As(ctx, &planNativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	networkRangeID := planNativeRange.NativeNetworkRangeID.ValueString()
	r.updateNetworkRange(ctx, &cfg, &plan, networkRangeID, isHA, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Move interface to another index if needed
	convertUpdateInput(
		ctx,
		r,
		&plan,
		state,
		&stateNativeRange,
		&diags,
		siteID,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Socket Interface
	r.updateSocketInterface(ctx, &cfg, &plan, siteID, isHA, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists := r.hydrateSocketSiteState(ctx, &cfg, plan, siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete cato_socket_site resource
func (r *socketSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.resourceAPI().EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"),
		nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
	tflog.Debug(ctx, "Create.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// check if site exist before removing
	if len(querySiteResult.EntityLookup.GetItems()) == 1 {
		err := removeSocketSiteWithRetry(ctx, r.resourceAPI(), state.ID.ValueString(), r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteRemoveSite error",
				err.Error(),
			)
			return
		}
	}
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	r *socketSiteResource,
	originalPlan *SocketSite,
	state SocketSite,
	originalStateNativeRange *NativeRange,
	originalDiags *diag.Diagnostics,
	siteID string,
	diagnostics *diag.Diagnostics,
) {
	plan := *originalPlan
	defer func() { *originalPlan = plan }()

	stateNativeRange := *originalStateNativeRange
	defer func() { *originalStateNativeRange = stateNativeRange }()

	diags := *originalDiags
	defer func() { *originalDiags = diags }()

	if utils.CheckErr(&diags, state.NativeRange.As(ctx, &stateNativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	currentInterfaceIndex := cato_models.SocketInterfaceIDEnum(stateNativeRange.InterfaceIndex.ValueString())
	r.assignInterfaceIndex(ctx, currentInterfaceIndex, &plan, siteID, diagnostics)
}

// updateSocketInterface updates the socket interface details.
func (r *socketSiteResource) updateSocketInterface(ctx context.Context, cfg, plan *SocketSite, siteID string,
	isHA bool, diags *diag.Diagnostics,
) {
	if isHA && plan.ConnectionType.ValueString() == string(cato_models.SiteConnectionTypeEnumSocketGCP1500) {
		return // update does not work on GCP HA
	}

	socketIfaceInput, interfaceID := prepareSocketInterfaceInput(ctx, cfg, plan, isHA, diags)
	if diags.HasError() {
		return
	}

	// Update socket interface
	_, err := r.resourceAPI().SiteUpdateSocketInterface(ctx, siteID, interfaceID, *socketIfaceInput, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateSocketInterface error", err.Error())
	}
}

// updateNativeRange updates the native network range details by calling SiteUpdateNetworkRange API.
func (r *socketSiteResource) updateNetworkRange(ctx context.Context, cfg, plan *SocketSite, networkRangeID string,
	isHA bool, diags *diag.Diagnostics,
) {
	networkRangeInput := prepareNetworkRangeInput(ctx, r.dhcpClient(), cfg, plan, isHA, diags)
	if diags.HasError() {
		return
	}

	// Update native network range
	_, err := r.resourceAPI().SiteUpdateNetworkRange(ctx, networkRangeID, *networkRangeInput, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateNetworkRange error", err.Error())
	}
}

// assignInterfaceIndex attempts to assign the desired interface index for the native range by calling SiteExchangeSocketPorts API.
func (r *socketSiteResource) assignInterfaceIndex(ctx context.Context, currentInterfaceIndex cato_models.SocketInterfaceIDEnum,
	plan *SocketSite, siteID string, diags *diag.Diagnostics,
) {
	var tfNativeRange NativeRange
	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	if !utils.HasValue(tfNativeRange.InterfaceIndex) {
		return
	}

	desiredInterfaceIndex := cato_models.SocketInterfaceIDEnum(tfNativeRange.InterfaceIndex.ValueString())

	if desiredInterfaceIndex == currentInterfaceIndex {
		return // No need to move the interface if the desired index is the default
	}

	exchangeInput := cato_models.ExchangeSocketPortsInput{
		Site:            &cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID},
		FirstInterface:  &cato_models.SocketInterfaceRefInput{InterfaceID: currentInterfaceIndex},
		SecondInterface: &cato_models.SocketInterfaceRefInput{InterfaceID: desiredInterfaceIndex},
	}

	siteExchangeSocketPortsResponse, err := r.resourceAPI().SiteExchangeSocketPorts(ctx, r.client.AccountId, exchangeInput)
	if err != nil {
		diags.AddError("Catov2 API SiteExchangeSocketPorts error", err.Error())
		return
	}
	tflog.Debug(ctx, "attemptReassignNativeRangeIndex.tmpSiteUpdateSocketInterface.tmp.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteExchangeSocketPortsResponse),
	})
}

// findDefaultInterface looks up the default network interface for the given site ID and returns its details.
func (r *socketSiteResource) findDefaultInterface(ctx context.Context, siteID string,
	nativeRange types.Object, diags *diag.Diagnostics,
) *nativeInterfaceDetails {
	var defaultIface *cato_go_sdk.EntityLookup_EntityLookup_Items
	helperField := func(fieldName string, it *cato_go_sdk.EntityLookup_EntityLookup_Items) *string {
		valAny, ok := it.HelperFields[fieldName]
		if !ok {
			return nil
		}
		if valStr, ok := valAny.(string); ok {
			return &valStr
		}
		return nil
	}

	// Retrieve default interface range attributes
	siteEntity := convertFindDefaultInterfaceValues(siteID)
	result, err := r.resourceAPI().EntityLookup(ctx, r.client.AccountId,
		cato_models.EntityTypeNetworkInterface, new(int64(0)), nil, siteEntity, nil, nil, nil, nil, nil)
	if err != nil {
		diags.AddError("Catov2 API EntityLookup 'networkInterface' error", err.Error())
		return nil
	}

	// find the default interface based on helper field "isDefault" set by API
	for _, item := range result.EntityLookup.GetItems() {
		if isDefaultAny, ok := item.HelperFields["isDefault"]; ok {
			if isDefault, ok := isDefaultAny.(bool); ok && isDefault {
				defaultIface = item
				break
			}
		}
	}

	// Fallback: if API does not return "isDefault" as a helper field, try to match the interface ID from state
	if defaultIface == nil && utils.HasValue(nativeRange) {
		var tfNativeRange NativeRange
		if utils.CheckErr(diags, nativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
			return nil
		}
		statIfaceID := tfNativeRange.InterfaceID.ValueString()
		for _, item := range result.EntityLookup.GetItems() {
			if item.Entity.ID == statIfaceID {
				defaultIface = item
				break
			}
		}
	}

	if defaultIface == nil {
		diags.AddError("Default interface not found", fmt.Sprintf("no default network interface found for site ID %s", siteID))
		return nil
	}

	ifaceDetails := &nativeInterfaceDetails{
		interfaceIndex: helperField("interfaceId", defaultIface), // yes interfaceId contains the index ("LAN1")
		interfaceID:    &defaultIface.Entity.ID,                  // entityID (1234)
		interfaceName:  helperField("interfaceName", defaultIface),
		destType:       helperField("destType", defaultIface),
	}
	return ifaceDetails
}

// findNativeRange looks up the native network range for the given site ID and returns its details.
func (r *socketSiteResource) findNativeRange(ctx context.Context, siteID string, diags *diag.Diagnostics,
) (networkRange *cato_go_sdk.NetworkRangeList_Site_NetworkRangeList_Items) {
	input := convertFindNativeRangeValues(siteID)
	queryNetworkRangeResult, err := r.resourceAPI().NetworkRangeList(ctx, r.client.AccountId, input)
	if err != nil {
		diags.AddError("Catov2 API NetworkRangeList error", err.Error())
		return
	}
	for _, netRange := range queryNetworkRangeResult.GetSite().GetNetworkRangeList().GetItems() {
		if netRange.RangeType == cato_models.SubnetTypeNative {
			return netRange
		}
	}
	diags.AddError("Native network range not found", fmt.Sprintf("no native network range found for site ID %s", siteID))
	return nil
}

// updateBasicSocketSite updates a socket site
// It does NOT update native range or interface details.
func (r *socketSiteResource) updateBasicSocketSite(ctx context.Context, plan *SocketSite, diags *diag.Diagnostics) {
	input := prepareUpdateSocketSiteGeneralDetailsInput(ctx, plan, diags)
	if diags.HasError() {
		return
	}

	_, err := r.resourceAPI().SiteUpdateSiteGeneralDetails(ctx, plan.ID.ValueString(), input, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateSiteGeneralDetails error", err.Error())
		return
	}
}

func (r *socketSiteResource) configureCloudRouterAfterCreate(ctx context.Context, plan *SocketSite, siteID string,
	diags *diag.Diagnostics,
) {
	r.updateCloudRouter(ctx, plan, siteID, diags)
	if diags.HasError() {
		r.cleanupCreatedSocketSite(ctx, siteID, diags)
	}
}

func (r *socketSiteResource) cleanupCreatedSocketSite(ctx context.Context, siteID string, diags *diag.Diagnostics) {
	_, err := r.resourceAPI().SiteRemoveSite(ctx, siteID, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteRemoveSite cleanup error", err.Error())
	}
}

func (r *socketSiteResource) updateCloudRouter(
	ctx context.Context,
	plan *SocketSite,
	siteID string,
	diags *diag.Diagnostics,
) {
	if !utils.HasValue(plan.IsCloudRouter) {
		return
	}

	input := prepareUpdateHaInput(plan)
	_, err := r.resourceAPI().SiteUpdateHa(ctx, r.client.AccountId, siteID, input)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateHa error", err.Error())
	}
}

// createBasicSocketSite creates a socket site and returns the new site ID.
// It does NOT update native range or interface details.
func (r *socketSiteResource) createBasicSocketSite(
	ctx context.Context,
	plan *SocketSite,
	diags *diag.Diagnostics,
) (siteID string) {
	socketSiteInput := prepareSocketSiteInput(ctx, plan, diags)
	if diags.HasError() {
		return ""
	}
	socketSite, err := r.resourceAPI().SiteAddSocketSite(ctx, *socketSiteInput, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteAddSocketSite error", err.Error())
		return ""
	}
	siteID = socketSite.Site.AddSocketSite.GetSiteID()
	if siteID == "" {
		diags.AddError("Catov2 API SiteAddSocketSite error", "empty site ID returned from API")
		return ""
	}
	return siteID
}

func (r *socketSiteResource) parseNativeRange(ctx context.Context, cfg *SocketSite,
	networkRange *cato_go_sdk.NetworkRangeList_Site_NetworkRangeList_Items,
	nativeInterface *nativeInterfaceDetails, stateNativeRange types.Object, diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics
	var tfStateNativeRange NativeRange

	if networkRange == nil {
		return types.ObjectNull(SiteNativeRangeResourceAttrTypes)
	}

	dhcpSettingsObj := r.parseDhcpSettingsObj(ctx, cfg, networkRange, stateNativeRange, diags)
	if diags.HasError() {
		return types.ObjectNull(SiteNativeRangeResourceAttrTypes)
	}

	if nativeInterface == nil {
		nativeInterface = &nativeInterfaceDetails{}
	}

	// decode status NativeRange
	if utils.HasValue(stateNativeRange) {
		if utils.CheckErr(diags, stateNativeRange.As(ctx, &tfStateNativeRange, basetypes.ObjectAsOptions{})) {
			return types.ObjectNull(SiteNativeRangeResourceAttrTypes)
		}
	}

	// try to get LagMinLinks from state; not available in API; TODO: add to API
	netRangeObj := convertParseNativeRangeValues(
		ctx,
		networkRange,
		nativeInterface,
		stateNativeRange,
		diags,
		&objDiags,
		tfStateNativeRange,
		dhcpSettingsObj,
	)
	if diags.HasError() {
		return types.ObjectNull(SiteNativeRangeResourceAttrTypes)
	}

	return netRangeObj
}

// parseNativeRange converts API native range data to the types.Object or NativeRange
// parseDhcpSettingsObj resolves the dhcp_settings object for a native range.
// Three cases:
//  1. User configured a specific type (not ACCOUNT_DEFAULT) → parse from API response.
//  2. User explicitly set dhcp_type = ACCOUNT_DEFAULT → store the default sentinel object.
//  3. User did not configure dhcp_settings at all → keep null to match the plan.
func (r *socketSiteResource) parseDhcpSettingsObj(
	ctx context.Context,
	cfg *SocketSite,
	networkRange *cato_go_sdk.NetworkRangeList_Site_NetworkRangeList_Items,
	stateNativeRange types.Object,
	diags *diag.Diagnostics,
) types.Object {
	isDhcpDefault := r.checkDhcpSettingsDefault(ctx, cfg, stateNativeRange, diags)
	if networkRange.DhcpSettings != nil && !isDhcpDefault {
		return dhcp.ParseSettings(ctx, r.dhcpClient(),
			&cato_go_sdk.NetworkRange_Site_NetworkRange_DhcpSettings{
				DhcpMicrosegmentation: networkRange.DhcpSettings.DhcpMicrosegmentation,
				DhcpType:              networkRange.DhcpSettings.DhcpType,
				IPRange:               networkRange.DhcpSettings.IPRange,
				RelayGroupID:          networkRange.DhcpSettings.RelayGroupID,
			},
			diags)
	}
	if isDhcpDefault && r.isDhcpSettingsExplicit(ctx, cfg, stateNativeRange, diags) {
		return dhcp.SettingsDefault(ctx, diags)
	}
	return types.ObjectNull(dhcp.SettingsAttrTypes)
}

// hydrateSocketSiteState populates the SocketSite state with data from API responses
func (r *socketSiteResource) hydrateSocketSiteState(ctx context.Context, cfg *SocketSite, state SocketSite,
	siteID string, diags *diag.Diagnostics,
) (newState SocketSite, siteExists bool) {
	// Fetch site general details for basic and location data.
	siteDetails, err := r.resourceAPI().SiteGeneralDetails(ctx,
		cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID}, r.client.AccountId)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to fetch SiteGeneralDetails for site '%s'", siteID), err.Error())
		return state, false
	}
	if siteDetails == nil || siteDetails.GetSite().GetSiteGeneralDetails() == nil {
		return state, false
	}
	siteGeneralDetails := siteDetails.GetSite().GetSiteGeneralDetails()

	// Fetch socket-specific data without the expensive AccountSnapshot query.
	siteSocketConfiguration := r.fetchSocketConfiguration(ctx, siteID, diags)
	if diags.HasError() {
		return state, true
	}

	// Fetch site native range and interface details
	networkRange := r.findNativeRange(ctx, siteID, diags)
	if diags.HasError() {
		return state, true
	}

	// Fetch network interface details
	defaultInterface := r.findDefaultInterface(ctx, siteID, state.NativeRange, diags)
	if diags.HasError() {
		return state, true
	}

	convertHydrateSocketSiteStateValues(
		ctx,
		r,
		cfg,
		state,
		siteID,
		diags,
		&newState,
		siteDetails,
		siteGeneralDetails,
		siteSocketConfiguration,
		networkRange,
		defaultInterface,
	)
	if diags.HasError() {
		return state, true
	}

	return newState, true
}

// prepareNetworkRangeInput constructs the API input for SiteUpdateNetworkRange() from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.
func prepareNetworkRangeInput(
	ctx context.Context,
	client dhcp.SdkClienter,
	cfg, plan *SocketSite,
	isHA bool,
	diags *diag.Diagnostics,
) *cato_models.UpdateNetworkRangeInput {
	var (
		cfgNativeRange  NativeRange
		planNativeRange NativeRange
	)
	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &planNativeRange, basetypes.ObjectAsOptions{})) {
		return nil
	}
	configTranslatedSubnet := types.StringNull()
	if cfg != nil && utils.HasValue(cfg.NativeRange) {
		if utils.CheckErr(diags, cfg.NativeRange.As(ctx, &cfgNativeRange, basetypes.ObjectAsOptions{})) {
			return nil
		}
		configTranslatedSubnet = cfgNativeRange.TranslatedSubnet
	}
	input := convertPrepareNetworkRangeInputValues(ctx, client, isHA, diags, planNativeRange, configTranslatedSubnet)

	return input
}
