package socketsite

import (
	"context"
	"slices"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	scalars "github.com/catonetworks/cato-go-sdk/scalars"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
	pops "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/pops"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// prepareSiteLocation constructs the API input (location part) for SiteAddSocketSite() from the Terraform plan data.
func prepareSiteLocation(
	ctx context.Context,
	location types.Object,
	diags *diag.Diagnostics,
) *cato_models.AddSiteLocationInput {
	if !utils.HasValue(location) {
		return nil
	}

	var tfLocation tf.SiteLocation
	if utils.CheckErr(diags, location.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.AddSiteLocationInput{
		Address:     convert.KnownStringPointer(tfLocation.Address),
		City:        convert.KnownStringPointer(tfLocation.City),
		CountryCode: tfLocation.CountryCode.ValueString(),
		StateCode:   convert.KnownStringPointer(tfLocation.StateCode),
		Timezone:    tfLocation.Timezone.ValueString(),
	}
}

// prepareSocketSiteInput constructs the API input for SiteAddSocketSite() from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.
func prepareSocketSiteInput(ctx context.Context, plan *SocketSite, diags *diag.Diagnostics) *cato_models.AddSocketSiteInput {
	var tfNativeRange NativeRange
	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.AddSocketSiteInput{
		ConnectionType:     cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString()),
		Description:        convert.KnownStringPointer(plan.Description),
		Name:               plan.Name.ValueString(),
		NativeNetworkRange: tfNativeRange.NativeNetworkRange.ValueString(),
		SiteLocation:       prepareSiteLocation(ctx, plan.SiteLocation, diags),
		SiteType:           cato_models.SiteType(plan.SiteType.ValueString()),
		TranslatedSubnet:   convert.StringPointerForOptionalInput(tfNativeRange.TranslatedSubnet),
		Vlan:               (*scalars.Vlan)(convert.KnownInt64Pointer(tfNativeRange.Vlan)),
	}
}

// vrrpDestTypeSuffix is the suffix the Cato API appends to a socket interface destination
// type once a second socket is physically connected and the HA pair forms. It is never
// accepted as an input value by siteUpdateSocketInterface.
const vrrpDestTypeSuffix = "_AND_VRRP"

// stripVrrpForAPI reduces a "*_AND_VRRP" destination type to its settable base type.
// The API appends the _AND_VRRP suffix automatically after HA forms; siteUpdateSocketInterface
// rejects *_AND_VRRP as an input value on both create and update (only LAN, LAN_LAG_MEMBER and
// LAN_LAG_MASTER are settable), so LAN_LAG_MASTER_AND_VRRP reduces to LAN_LAG_MASTER and any
// other *_AND_VRRP variant reduces to LAN. The configured *_AND_VRRP value is kept in Terraform
// state (see parsedInterfaceDestType); the promotion itself happens server-side once the second
// socket actually exists.
func stripVrrpForAPI(destType cato_models.SocketInterfaceDestType) cato_models.SocketInterfaceDestType {
	if !strings.HasSuffix(string(destType), vrrpDestTypeSuffix) {
		return destType
	}
	if destType == cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp {
		return cato_models.SocketInterfaceDestTypeLanLagMaster
	}
	return cato_models.SocketInterfaceDestTypeLan
}

// parsedInterfaceDestType decides which interface_dest_type value to store in state after an
// API read-back.
//
// On the apply paths (Create/Update, isApplyPath=true) "prior" carries the PLANNED value and it
// wins whenever the user set one: the API is only ever sent the stripped base type (see
// stripVrrpForAPI), so the read-back reports the base type while Terraform requires the applied
// state to equal the planned value exactly.
//
// On the Read path (isApplyPath=false) "prior" carries the previous state and is kept only for
// the single-socket drift case: prior holds "X_AND_VRRP" while the API still reports the base
// "X" because the HA pair has not formed yet (never downgrade). Anything else — including the
// server-side *_AND_VRRP promotion once the second socket connects — stores the API value as-is.
func parsedInterfaceDestType(prior types.String, apiDestType *string, isApplyPath bool) types.String {
	if !utils.HasValue(prior) {
		return types.StringPointerValue(apiDestType)
	}
	if isApplyPath {
		return prior
	}
	if apiDestType != nil &&
		stripVrrpForAPI(cato_models.SocketInterfaceDestType(prior.ValueString())) == cato_models.SocketInterfaceDestType(*apiDestType) {
		return prior
	}
	return types.StringPointerValue(apiDestType)
}

// prepareSocketInterfaceInput prepares inputs for SiteUpdateSocketInterface API, on error updates diags.
func prepareSocketInterfaceInput(
	ctx context.Context,
	cfg, plan *SocketSite,
	isHA bool,
	diags *diag.Diagnostics,
) (*cato_models.UpdateSocketInterfaceInput, cato_models.SocketInterfaceIDEnum) {
	var (
		cfgNativeRange    NativeRange
		planNativeRange   NativeRange
		lagLinksDestTypes = []cato_models.SocketInterfaceDestType{
			cato_models.SocketInterfaceDestTypeLanLagMaster,
			cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp,
		}
		lanDestTypes = []cato_models.SocketInterfaceDestType{
			cato_models.SocketInterfaceDestTypeLan,
			cato_models.SocketInterfaceDestTypeLanAndHa,
			cato_models.SocketInterfaceDestTypeVrrpAndLan,
			cato_models.SocketInterfaceDestTypeLanLagMaster,
			cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp,
		}
	)

	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &planNativeRange, basetypes.ObjectAsOptions{})) {
		return nil, ""
	}
	configTranslatedSubnet := types.StringNull()
	if cfg != nil && utils.HasValue(cfg.NativeRange) {
		if utils.CheckErr(diags, cfg.NativeRange.As(ctx, &cfgNativeRange, basetypes.ObjectAsOptions{})) {
			return nil, ""
		}
		configTranslatedSubnet = cfgNativeRange.TranslatedSubnet
	}

	interfaceDestType := cato_models.SocketInterfaceDestType(planNativeRange.InterfaceDestType.ValueString())
	if interfaceDestType == "" {
		interfaceDestType = cato_models.SocketInterfaceDestTypeLan
	}
	// *_AND_VRRP is never settable via the API (create or update) — send the base type. The
	// configured *_AND_VRRP value stays in state (see parsedInterfaceDestType). Stripping happens
	// before the Lag/Lan checks below so a reduced LAN_LAG_MASTER still carries lag.minLinks.
	interfaceDestType = stripVrrpForAPI(interfaceDestType)

	ifaceName := convert.KnownStringPointer(planNativeRange.InterfaceName)
	if ifaceName == nil {
		ifaceName = convert.KnownStringPointer(planNativeRange.InterfaceIndex)
	}
	if ifaceName == nil {
		ifaceName = new(string(InterfaceByConnType[cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())]))
	}

	input := &cato_models.UpdateSocketInterfaceInput{
		DestType: interfaceDestType,
		Name:     ifaceName,
	}

	if utils.HasValue(planNativeRange.LagMinLinks) && slices.Contains(lagLinksDestTypes, interfaceDestType) {
		input.Lag = &cato_models.SocketInterfaceLagInput{MinLinks: planNativeRange.LagMinLinks.ValueInt64()}
	}

	if slices.Contains(lanDestTypes, interfaceDestType) && !isHA {
		input.Lan = &cato_models.SocketInterfaceLanInput{
			LocalIP:          planNativeRange.LocalIP.ValueString(),
			Subnet:           planNativeRange.NativeNetworkRange.ValueString(),
			TranslatedSubnet: convert.TranslatedSubnetForAPIInput(configTranslatedSubnet, planNativeRange.TranslatedSubnet),
		}
	}

	interfaceIndex := cato_models.SocketInterfaceIDEnum(planNativeRange.InterfaceIndex.ValueString())
	if interfaceIndex == "" {
		interfaceIndex = InterfaceByConnType[cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())]
	}

	return input, interfaceIndex
}

func prepareUpdateHaInput(plan *SocketSite) cato_models.UpdateHaInput {
	return cato_models.UpdateHaInput{IsCloudRouter: plan.IsCloudRouter.ValueBoolPointer()}
}

func prepareUpdateSocketSiteGeneralDetailsInput(
	ctx context.Context,
	plan *SocketSite,
	diags *diag.Diagnostics,
) cato_models.UpdateSiteGeneralDetailsInput {
	input := cato_models.UpdateSiteGeneralDetailsInput{
		Description: convert.KnownStringPointer(plan.Description),
		Name:        plan.Name.ValueStringPointer(),
		SiteType:    (*cato_models.SiteType)(plan.SiteType.ValueStringPointer()),
	}
	if !utils.HasValue(plan.SiteLocation) {
		return input
	}

	var tfLocation tf.SiteLocation
	if utils.CheckErr(diags, plan.SiteLocation.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})) {
		return input
	}
	input.SiteLocation = &cato_models.UpdateSiteLocationInput{
		Address:     convert.KnownStringPointer(tfLocation.Address),
		CityName:    convert.KnownStringPointer(tfLocation.City),
		CountryCode: tfLocation.CountryCode.ValueStringPointer(),
		StateCode:   convert.KnownStringPointer(tfLocation.StateCode),
		Timezone:    tfLocation.Timezone.ValueStringPointer(),
	}

	return input
}

// connectionTypeFromSocketConfiguration translates the socket model into the resource's connection type.
func connectionTypeFromSocketConfiguration(
	socketConfiguration *cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration,
) types.String {
	if socketConfiguration == nil {
		return types.StringNull()
	}
	model := socketConfiguration.GetPrimarySocketConfiguration().GetSocketInfo().GetModel()
	if model == nil {
		return types.StringNull()
	}

	connectionTypes := map[cato_models.SocketModel]cato_models.SiteConnectionTypeEnum{
		cato_models.SocketModelAWS:      cato_models.SiteConnectionTypeEnumSocketAWS1500,
		cato_models.SocketModelAzure:    cato_models.SiteConnectionTypeEnumSocketAz1500,
		cato_models.SocketModelEsx:      cato_models.SiteConnectionTypeEnumSocketEsx1500,
		cato_models.SocketModelGCP:      cato_models.SiteConnectionTypeEnumSocketGCP1500,
		cato_models.SocketModelX1500:    cato_models.SiteConnectionTypeEnumSocketX1500,
		cato_models.SocketModelX1600:    cato_models.SiteConnectionTypeEnumSocketX1600,
		cato_models.SocketModelX1600Lte: cato_models.SiteConnectionTypeEnumSocketX1600Lte,
		cato_models.SocketModelX1700:    cato_models.SiteConnectionTypeEnumSocketX1700,
	}
	connectionType, ok := connectionTypes[*model]
	if !ok {
		return types.StringNull()
	}
	return types.StringValue(string(connectionType))
}

// parseSiteLocation converts API site location data to the types.Object or tf.SiteLocation
func (r *socketSiteResource) parseSiteLocation(ctx context.Context, siteGenDetails *cato_go_sdk.SiteGeneralDetails,
	siteLocation types.Object, diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics
	var planLocation tf.SiteLocation

	if siteGenDetails == nil || siteGenDetails.Site.SiteGeneralDetails == nil {
		return types.ObjectNull(tf.SiteLocationResourceAttrTypes)
	}
	siteLoc := siteGenDetails.Site.SiteGeneralDetails.SiteLocation

	// Prepare site location object
	state := siteLoc.GetStateCode()
	if state != nil && *state == "" {
		state = nil
	}
	tfLocation := tf.SiteLocation{
		CountryCode: types.StringValue(siteLoc.GetCountryCode()),
		StateCode:   types.StringPointerValue(state),
		Timezone:    types.StringValue(siteLoc.GetTimezone()),
		Address:     types.StringPointerValue(siteLoc.GetAddress()),
		City:        types.StringPointerValue(siteLoc.GetCityName()),
	}

	if utils.HasValue(siteLocation) {
		if utils.CheckErr(diags, siteLocation.As(ctx, &planLocation, basetypes.ObjectAsOptions{})) {
			return types.ObjectNull(tf.SiteLocationResourceAttrTypes)
		}
		preservePlannedSiteLocationValues(planLocation, siteLoc.GetCityName(), &tfLocation)
	}

	locObj, objDiags := types.ObjectValueFrom(ctx, tf.SiteLocationResourceAttrTypes, tfLocation)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(pops.PreferredPopLocationModelTypes)
	}

	return locObj
}

// preservePlannedSiteLocationValues keeps plan-held values the API reports differently:
// the API sometimes returns empty string and sometimes null for city, and it resolves a
// concrete state code (e.g. "US-CA") even when config omits state_code - both would
// otherwise produce an inconsistent result after apply.
func preservePlannedSiteLocationValues(planLocation tf.SiteLocation, apiCity *string, tfLocation *tf.SiteLocation) {
	if (apiCity == nil || *apiCity == "") &&
		(planLocation.City.IsNull() || (utils.HasValue(planLocation.City) && planLocation.City.ValueString() == "")) {
		tfLocation.City = planLocation.City
	}
	if planLocation.StateCode.IsNull() {
		tfLocation.StateCode = types.StringNull()
	}
}

// parseSockets converts API site socket data to the types.Set of Socket
func (r *socketSiteResource) parseSockets(ctx context.Context,
	socketConfiguration *cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration,
	diags *diag.Diagnostics,
) types.Set {
	var objDiags diag.Diagnostics
	setNull := types.SetNull(types.ObjectType{AttrTypes: SocketTypes})

	if socketConfiguration == nil {
		return setNull
	}

	primary := socketConfiguration.GetPrimarySocketConfiguration()
	socketObjects := make([]types.Object, 0, 2)
	socketObjects = appendSocketConfiguration(ctx, socketObjects, primary.GetSerial(),
		primary.GetSocketInfo().GetIsPrimary(), primary.GetSocketInfo().GetPlatform(), diags)

	if secondary := socketConfiguration.GetSecondarySocketConfiguration(); secondary != nil {
		socketObjects = appendSocketConfiguration(ctx, socketObjects, secondary.GetSerial(),
			secondary.GetSocketInfo().GetIsPrimary(), secondary.GetSocketInfo().GetPlatform(), diags)
	}
	if diags.HasError() {
		return setNull
	}

	// convert slice to types.Set
	tfSocketSet, objDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: SocketTypes}, socketObjects)
	diags.Append(objDiags...)
	if diags.HasError() {
		return setNull
	}

	return tfSocketSet
}

func appendSocketConfiguration(ctx context.Context, sockets []types.Object, serial *string, isPrimary bool,
	platform *string, diags *diag.Diagnostics,
) []types.Object {
	tfSocketObj, objDiags := types.ObjectValueFrom(ctx, SocketTypes, Socket{
		ID:           types.StringNull(), // SiteSocketConfiguration does not expose the socket entity ID.
		SerialNumber: types.StringPointerValue(serial),
		IsPrimary:    types.BoolValue(isPrimary),
		Platform:     types.StringPointerValue(platform),
	})
	diags.Append(objDiags...)
	if diags.HasError() {
		return sockets
	}
	return append(sockets, tfSocketObj)
}

// isDhcpSettingsExplicit returns true when dhcp_settings was explicitly set in config or state
// (i.e. the object is non-null), even if the value is ACCOUNT_DEFAULT.
// Returns false when dhcp_settings is null (user never configured it).
func (r *socketSiteResource) isDhcpSettingsExplicit(ctx context.Context, cfg *SocketSite,
	stateNativeRange types.Object, diags *diag.Diagnostics,
) bool {
	if cfg != nil {
		var cfgNativeRange NativeRange
		if utils.CheckErr(diags, cfg.NativeRange.As(ctx, &cfgNativeRange, basetypes.ObjectAsOptions{})) {
			return false
		}
		return utils.HasValue(cfgNativeRange.DhcpSettings) || cfgNativeRange.DhcpSettings.IsUnknown()
	}
	// cfg is nil → Read(); check state
	if !utils.HasValue(stateNativeRange) {
		return false
	}
	var tfStateNativeRange NativeRange
	if utils.CheckErr(diags, stateNativeRange.As(ctx, &tfStateNativeRange, basetypes.ObjectAsOptions{})) {
		return false
	}
	return !tfStateNativeRange.DhcpSettings.IsNull()
}

// checkDhcpSettingsDefault checks DHCP settings config or state,
// If config is provided,
// return true if it is not defined or if dhcp_type is set to ACCOUNT_DEFAULT, false otherwise
// If config is not provided (nil), check the state with the same logic to determine if DHCP settings are default
//
// Reason: API does not return ACCOUNT_DEFAULT, but some other value based on CMA account configuration
func (r *socketSiteResource) checkDhcpSettingsDefault(ctx context.Context, cfg *SocketSite,
	stateNativeRange types.Object, diags *diag.Diagnostics,
) (isDhcpSettingsDefault bool) {
	if cfg != nil {
		var cfgNativeRange NativeRange
		if utils.CheckErr(diags, cfg.NativeRange.As(ctx, &cfgNativeRange, basetypes.ObjectAsOptions{})) {
			return false
		}
		if cfgNativeRange.DhcpSettings.IsNull() {
			return true
		}

		var cfgDhcpSettings dhcp.Settings
		if utils.CheckErr(diags, cfgNativeRange.DhcpSettings.As(ctx, &cfgDhcpSettings, basetypes.ObjectAsOptions{})) {
			return false
		}
		if utils.HasValue(cfgDhcpSettings.DhcpType) &&
			(cfgDhcpSettings.DhcpType.ValueString() == string(cato_models.DhcpTypeAccountDefault)) {
			return true
		}
		return false
	}

	// cfg is nil -> called from Read(); check the state
	if !utils.HasValue(stateNativeRange) {
		return false
	}
	var tfStateNativeRange NativeRange
	if utils.CheckErr(diags, stateNativeRange.As(ctx, &tfStateNativeRange, basetypes.ObjectAsOptions{})) {
		return false
	}
	if tfStateNativeRange.DhcpSettings.IsNull() {
		return true
	}
	var stateDhcpSettings dhcp.Settings
	if utils.CheckErr(diags, tfStateNativeRange.DhcpSettings.As(ctx, &stateDhcpSettings, basetypes.ObjectAsOptions{})) {
		return false
	}
	if utils.HasValue(stateDhcpSettings.DhcpType) &&
		(stateDhcpSettings.DhcpType.ValueString() == string(cato_models.DhcpTypeAccountDefault)) {
		return true
	}
	return false
}

// isHA determines if the site is in HA scenario based on the number of sockets
func (r *socketSiteResource) isHA(state *SocketSite) bool {
	if state == nil || !utils.HasValue(state.Sockets) {
		return false
	}
	return len(state.Sockets.Elements()) > 1
}

// convertCreateValues converts between the Terraform model and SDK values.
func convertCreateValues(cfg SocketSite, siteID string) (resultFallbackState SocketSite) {
	resultFallbackState = cfg
	resultFallbackState.ID = types.StringValue(siteID)
	return resultFallbackState
}

// convertFindDefaultInterfaceValues converts between the Terraform model and SDK values.
func convertFindDefaultInterfaceValues(siteID string) (resultSiteEntity *cato_models.EntityInput) {
	resultSiteEntity = &cato_models.EntityInput{Type: cato_models.EntityTypeSite, ID: siteID}
	return resultSiteEntity
}

// convertFindNativeRangeValues converts between the Terraform model and SDK values.
func convertFindNativeRangeValues(siteID string) (resultInput cato_models.NetworkRangeListInput) {
	resultInput = cato_models.NetworkRangeListInput{Site: &cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID}}
	return resultInput
}

// convertParseNativeRangeValues converts between the Terraform model and SDK values.
// On the apply paths (Create/Update) stateNativeRange/tfStateNativeRange carry the PLANNED
// native range; on Read they carry the prior state (isApplyPath=false).
func convertParseNativeRangeValues(
	ctx context.Context,
	networkRange *cato_go_sdk.NetworkRangeList_Site_NetworkRangeList_Items,
	nativeInterface *nativeInterfaceDetails,
	stateNativeRange types.Object,
	diags *diag.Diagnostics,
	priorObjDiags *diag.Diagnostics,
	tfStateNativeRange NativeRange,
	dhcpSettingsObj types.Object,
	isApplyPath bool,
) (resultNetRangeObj basetypes.ObjectValue) {
	objDiags := *priorObjDiags
	defer func() { *priorObjDiags = objDiags }()
	lagMinLinks := types.Int64Null()
	priorInterfaceDestType := types.StringNull()
	if utils.HasValue(stateNativeRange) {
		lagMinLinks = tfStateNativeRange.LagMinLinks
		priorInterfaceDestType = tfStateNativeRange.InterfaceDestType
	}

	// Fix interface index - API sometimes returns 'INT_5', sometimes just 5
	fixedIfaceIndex := nativeInterface.interfaceIndex
	if fixedIfaceIndex != nil && numberRE.MatchString(*fixedIfaceIndex) {
		fixedIfaceIndex = new("INT_" + *fixedIfaceIndex)
	}

	localIP := types.StringPointerValue(networkRange.LocalIP)
	if networkRange.LocalIP == nil { // In HA scenario the API does not return local IP
		if networkRange.PrimaryManagementIP != nil { // AWS: use primary management IP as local IP
			localIP = types.StringPointerValue(networkRange.PrimaryManagementIP)
		} else if utils.HasValue(stateNativeRange) { // try state local IP
			localIP = tfStateNativeRange.LocalIP
		}
	}

	// Prepare native range object
	tfNativeRange := NativeRange{
		InterfaceIndex:              types.StringPointerValue(fixedIfaceIndex),
		InterfaceID:                 types.StringPointerValue(nativeInterface.interfaceID),
		InterfaceName:               types.StringPointerValue(nativeInterface.interfaceName),
		NativeNetworkLanInterfaceID: types.StringPointerValue(nativeInterface.interfaceID),

		NativeNetworkRange:    types.StringValue(networkRange.Subnet),
		NativeNetworkRangeID:  types.StringValue(networkRange.NetworkRangeID),
		RangeName:             types.StringValue(networkRange.Name),
		RangeID:               types.StringValue(networkRange.NetworkRangeID),
		LocalIP:               localIP,
		PrimaryManagementIP:   types.StringPointerValue(networkRange.PrimaryManagementIP),
		SecondaryManagementIP: types.StringPointerValue(networkRange.SecondaryManagementIP),
		TranslatedSubnet:      types.StringPointerValue(networkRange.TranslatedSubnet),
		Gateway:               types.StringPointerValue(networkRange.Gateway),
		RangeType:             types.StringValue(strings.ToUpper(string(networkRange.RangeType))),
		DhcpSettings:          dhcpSettingsObj,
		Vlan:                  types.Int64PointerValue(networkRange.Vlan),
		MdnsReflector:         types.BoolValue(networkRange.MdnsReflector),
		LagMinLinks:           lagMinLinks,
		InterfaceDestType:     parsedInterfaceDestType(priorInterfaceDestType, nativeInterface.destType, isApplyPath),
	}

	// The native range is always a NATIVE range (never Routed), so the API's mdns_reflector
	// value is authoritative and is surfaced as-is (removing the Default(false) on the schema
	// is enough to avoid the inconsistent-result since an omitted value plans as unknown).

	resultNetRangeObj, objDiags = types.ObjectValueFrom(ctx, SiteNativeRangeResourceAttrTypes, tfNativeRange)
	diags.Append(objDiags...)
	return resultNetRangeObj
}

// convertHydrateSocketSiteStateValues converts between the Terraform model and SDK values.
func convertHydrateSocketSiteStateValues(
	ctx context.Context,
	r *socketSiteResource,
	cfg *SocketSite,
	state SocketSite,
	siteID string,
	diags *diag.Diagnostics,
	priorNewState *SocketSite,
	siteDetails *cato_go_sdk.SiteGeneralDetails,
	siteGeneralDetails *cato_go_sdk.SiteGeneralDetails_Site_SiteGeneralDetails,
	siteSocketConfiguration *cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration,
	networkRange *cato_go_sdk.NetworkRangeList_Site_NetworkRangeList_Items,
	defaultInterface *nativeInterfaceDetails,
) {
	newState := *priorNewState
	defer func() { *priorNewState = newState }()
	newState = SocketSite{
		ID:             types.StringValue(siteID),
		Name:           types.StringValue(siteGeneralDetails.GetSite().GetName()),
		ConnectionType: connectionTypeFromSocketConfiguration(siteSocketConfiguration),
		SiteType:       types.StringPointerValue((*string)(siteGeneralDetails.GetSiteType())),
		Description:    convert.StringPointerValue(siteGeneralDetails.GetDescription(), state.Description),
		IsCloudRouter:  state.IsCloudRouter,
		NativeRange:    r.parseNativeRange(ctx, cfg, networkRange, defaultInterface, state.NativeRange, diags),
		SiteLocation:   r.parseSiteLocation(ctx, siteDetails, state.SiteLocation, diags),
		Sockets:        r.parseSockets(ctx, siteSocketConfiguration, diags),
	}
}

// convertPrepareNetworkRangeInputValues converts between the Terraform model and SDK values.
func convertPrepareNetworkRangeInputValues(
	ctx context.Context,
	client dhcp.SdkClienter,
	isHA bool,
	diags *diag.Diagnostics,
	planNativeRange NativeRange,
	configTranslatedSubnet basetypes.StringValue,
) (resultInput *cato_models.UpdateNetworkRangeInput) {
	resultInput = &cato_models.UpdateNetworkRangeInput{
		Subnet:           convert.KnownStringPointer(planNativeRange.NativeNetworkRange),
		TranslatedSubnet: convert.TranslatedSubnetForAPIInput(configTranslatedSubnet, planNativeRange.TranslatedSubnet),
		MdnsReflector:    convert.KnownBoolPointer(planNativeRange.MdnsReflector),
		Vlan:             convert.KnownInt64Pointer(planNativeRange.Vlan),
		DhcpSettings:     dhcp.PrepareDHCPSettings(ctx, client, cato_models.SubnetTypeNative, planNativeRange.DhcpSettings, diags),
		// AzureFloatingIP must remain omitted until Azure HA support is implemented.
	}

	if !isHA { // for HA scenario, local IP is not allowed to be modified
		resultInput.LocalIP = convert.KnownStringPointer(planNativeRange.LocalIP)
	}
	return resultInput
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
