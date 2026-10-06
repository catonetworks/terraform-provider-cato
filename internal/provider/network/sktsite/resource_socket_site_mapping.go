package sktsite

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/parse"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/pops"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/siteloc"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/utils"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/application"
)

func (r *socketSiteResource) isDhcpSettingsExplicit(ctx context.Context, cfg *SocketSite,
	stateNativeRange types.Object, diags *diag.Diagnostics,
) bool {
	if cfg != nil {
		var cfgNativeRange NativeRange
		if apperr.CheckErr(diags, cfg.NativeRange.As(ctx, &cfgNativeRange, basetypes.ObjectAsOptions{})) {
			return false
		}
		return utils.HasValue(cfgNativeRange.DhcpSettings) || cfgNativeRange.DhcpSettings.IsUnknown()
	}
	// cfg is nil → Read(); check state
	if len(stateNativeRange.AttributeTypes(ctx)) == 0 || !utils.HasValue(stateNativeRange) {
		return false
	}
	var tfStateNativeRange NativeRange
	if apperr.CheckErr(diags, stateNativeRange.As(ctx, &tfStateNativeRange, basetypes.ObjectAsOptions{})) {
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
		if apperr.CheckErr(diags, cfg.NativeRange.As(ctx, &cfgNativeRange, basetypes.ObjectAsOptions{})) {
			return false
		}
		if cfgNativeRange.DhcpSettings.IsNull() {
			return true
		}

		var cfgDhcpSettings dhcp.Settings
		if apperr.CheckErr(
			diags,
			cfgNativeRange.DhcpSettings.As(ctx, &cfgDhcpSettings, basetypes.ObjectAsOptions{}),
		) {
			return false
		}
		if utils.HasValue(cfgDhcpSettings.DhcpType) &&
			(cfgDhcpSettings.DhcpType.ValueString() == "ACCOUNT_DEFAULT") {
			return true
		}
		return false
	}

	// cfg is nil -> called from Read(); check the state
	if len(stateNativeRange.AttributeTypes(ctx)) == 0 || !utils.HasValue(stateNativeRange) {
		return false
	}
	var tfStateNativeRange NativeRange
	if apperr.CheckErr(diags, stateNativeRange.As(ctx, &tfStateNativeRange, basetypes.ObjectAsOptions{})) {
		return false
	}
	if tfStateNativeRange.DhcpSettings.IsNull() {
		return true
	}
	var stateDhcpSettings dhcp.Settings
	if apperr.CheckErr(
		diags,
		tfStateNativeRange.DhcpSettings.As(ctx, &stateDhcpSettings, basetypes.ObjectAsOptions{}),
	) {
		return false
	}
	if utils.HasValue(stateDhcpSettings.DhcpType) &&
		(stateDhcpSettings.DhcpType.ValueString() == "ACCOUNT_DEFAULT") {
		return true
	}
	return false
}

// parseDhcpSettingsObj resolves the dhcp_settings object for a native range.
// Three cases:
//  1. User configured a specific type (not ACCOUNT_DEFAULT) → parse from API response.
//  2. User explicitly set dhcp_type = ACCOUNT_DEFAULT → store the default sentinel object.
//  3. User did not configure dhcp_settings at all → keep null to match the plan.

func (r *socketSiteResource) parseDhcpSettingsObj(
	ctx context.Context,
	cfg *SocketSite,
	networkRange *application.Range,
	stateNativeRange types.Object,
	diags *diag.Diagnostics,
) types.Object {
	isDhcpDefault := r.checkDhcpSettingsDefault(ctx, cfg, stateNativeRange, diags)
	if networkRange.DhcpSettings != nil && !isDhcpDefault {
		return dhcp.ProjectSettings(ctx, networkRange.DhcpSettings, diags)
	}
	if isDhcpDefault && r.isDhcpSettingsExplicit(ctx, cfg, stateNativeRange, diags) {
		return dhcp.SettingsDefault(ctx, diags)
	}
	return types.ObjectNull(dhcp.SettingsAttrTypes)
}

// parseNativeRange projects a plain snapshot while preserving values unavailable from the API.
func (r *socketSiteResource) parseNativeRange(ctx context.Context, cfg *SocketSite,
	networkRange *application.Range,
	nativeInterface *application.Interface, stateNativeRange types.Object, diags *diag.Diagnostics,
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
		nativeInterface = &application.Interface{}
	}

	// decode status NativeRange
	if utils.HasValue(stateNativeRange) {
		if apperr.CheckErr(diags, stateNativeRange.As(ctx, &tfStateNativeRange, basetypes.ObjectAsOptions{})) {
			return types.ObjectNull(SiteNativeRangeResourceAttrTypes)
		}
	}

	// try to get LagMinLinks from state; not available in API; TODO: add to API
	lagMinLinks := types.Int64Null()
	if utils.HasValue(stateNativeRange) {
		lagMinLinks = tfStateNativeRange.LagMinLinks
	}

	// Fix interface index - API sometimes returns 'INT_5', sometimes just 5
	fixedIfaceIndex := nativeInterface.Index
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
		InterfaceID:                 types.StringPointerValue(nativeInterface.ID),
		InterfaceName:               types.StringPointerValue(nativeInterface.Name),
		NativeNetworkLanInterfaceID: types.StringPointerValue(nativeInterface.ID),
		NativeNetworkRange:          types.StringValue(networkRange.Subnet),
		NativeNetworkRangeID:        types.StringValue(networkRange.NetworkRangeID),
		RangeName:                   types.StringValue(networkRange.Name),
		RangeID:                     types.StringValue(networkRange.NetworkRangeID),
		LocalIP:                     localIP,
		PrimaryManagementIP:         types.StringPointerValue(networkRange.PrimaryManagementIP),
		SecondaryManagementIP:       types.StringPointerValue(networkRange.SecondaryManagementIP),
		TranslatedSubnet:            types.StringPointerValue(networkRange.TranslatedSubnet),
		Gateway:                     types.StringPointerValue(networkRange.Gateway),
		RangeType:                   types.StringValue(strings.ToUpper(networkRange.RangeType)),
		DhcpSettings:                dhcpSettingsObj,
		Vlan:                        types.Int64PointerValue(networkRange.Vlan),
		MdnsReflector:               types.BoolValue(networkRange.MdnsReflector),
		LagMinLinks:                 lagMinLinks,
		InterfaceDestType:           types.StringPointerValue(nativeInterface.DestType),
	}

	netRangeObj, objDiags := types.ObjectValueFrom(ctx, SiteNativeRangeResourceAttrTypes, tfNativeRange)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(SiteNativeRangeResourceAttrTypes)
	}

	return netRangeObj
}

// prepareSiteLocation constructs the application input (location part) for SiteAddSocketSite() from the Terraform plan data.

func (r *socketSiteResource) prepareSiteLocation(ctx context.Context, location types.Object, diags *diag.Diagnostics,
) *application.Location {
	if !utils.HasValue(location) {
		return nil
	}

	var tfLocation siteloc.SiteLocation
	if apperr.CheckErr(diags, location.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &application.Location{

		Address:     parse.KnownStringPointer(tfLocation.Address),
		City:        parse.KnownStringPointer(tfLocation.City),
		CountryCode: tfLocation.CountryCode.ValueString(),
		StateCode:   parse.KnownStringPointer(tfLocation.StateCode),
		Timezone:    tfLocation.Timezone.ValueString(),
	}
}

// prepareSocketSiteInput constructs the application input for SiteAddSocketSite() from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.

func (r *socketSiteResource) prepareSocketSiteInput(ctx context.Context, plan *SocketSite, diags *diag.Diagnostics,
) *application.AddInput {
	var tfNativeRange NativeRange
	if apperr.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return nil
	}

	input := &application.AddInput{

		ConnectionType:     plan.ConnectionType.ValueString(),
		Description:        parse.KnownStringPointer(plan.Description),
		Name:               plan.Name.ValueString(),
		NativeNetworkRange: tfNativeRange.NativeNetworkRange.ValueString(),
		SiteLocation:       r.prepareSiteLocation(ctx, plan.SiteLocation, diags),
		SiteType:           plan.SiteType.ValueString(),
		TranslatedSubnet:   parse.StringPointerForOptionalInput(tfNativeRange.TranslatedSubnet),
		Vlan:               parse.KnownInt64Pointer(tfNativeRange.Vlan),
	}
	return input
}

// prepareNetworkRangeInput constructs the application input for SiteUpdateNetworkRange() from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.

func (r *socketSiteResource) prepareNetworkRangeInput(ctx context.Context, cfg, plan *SocketSite, isHA bool, diags *diag.Diagnostics,
) *application.RangeInput {
	var (
		cfgNativeRange  NativeRange
		planNativeRange NativeRange
	)
	if apperr.CheckErr(diags, plan.NativeRange.As(ctx, &planNativeRange, basetypes.ObjectAsOptions{})) {
		return nil
	}
	configTranslatedSubnet := types.StringNull()
	if cfg != nil && utils.HasValue(cfg.NativeRange) {
		if apperr.CheckErr(diags, cfg.NativeRange.As(ctx, &cfgNativeRange, basetypes.ObjectAsOptions{})) {
			return nil
		}
		configTranslatedSubnet = cfgNativeRange.TranslatedSubnet
	}
	input := &application.RangeInput{

		Subnet:           parse.KnownStringPointer(planNativeRange.NativeNetworkRange),
		TranslatedSubnet: translatedSubnetForAPIInput(configTranslatedSubnet, planNativeRange.TranslatedSubnet),
		MdnsReflector:    parse.KnownBoolPointer(planNativeRange.MdnsReflector),
		Vlan:             parse.KnownInt64Pointer(planNativeRange.Vlan),
		DhcpSettings:     dhcp.DecodeSettings(ctx, planNativeRange.DhcpSettings, diags),
		//    AzureFloatingIP *string `json:"azureFloatingIp,omitempty"` TODO: implement when AZURE HA support is added
	}

	if !isHA { // for HA scenario, local IP is not allowed to be modified
		input.LocalIP = parse.KnownStringPointer(planNativeRange.LocalIP)
	}

	return input
}

// prepareSocketInterfaceInput prepares plain inputs for the socket-interface adapter,
// on error updates diags

func (r *socketSiteResource) prepareSocketInterfaceInput(ctx context.Context, cfg, plan *SocketSite, isHA bool, diags *diag.Diagnostics,
) (input *application.InterfaceInput, index string) {
	var (
		cfgNativeRange    NativeRange
		planNativeRange   NativeRange
		lagLinksDestTypes = []string{
			// set lag input for these dest types
			"LAN_LAG_MASTER",
			"LAN_LAG_MASTER_AND_VRRP",
		}
		lanDestTypes = []string{
			// set lan input for these dest types
			"LAN",
			"LAN_AND_HA",
			"VRRP_AND_LAN",
			"LAN_LAG_MASTER",
			"LAN_LAG_MASTER_AND_VRRP",
		}
	)

	if apperr.CheckErr(diags, plan.NativeRange.As(ctx, &planNativeRange, basetypes.ObjectAsOptions{})) {
		return nil, ""
	}
	configTranslatedSubnet := types.StringNull()
	if cfg != nil && utils.HasValue(cfg.NativeRange) {
		if apperr.CheckErr(diags, cfg.NativeRange.As(ctx, &cfgNativeRange, basetypes.ObjectAsOptions{})) {
			return nil, ""
		}
		configTranslatedSubnet = cfgNativeRange.TranslatedSubnet
	}

	interfaceDestType := planNativeRange.InterfaceDestType.ValueString()

	if interfaceDestType == "" {
		interfaceDestType = "LAN" // default to LAN if not specified
	}

	// Determine interface ifaceName with the following precedence: InterfaceName > InterfaceIndex > default based on connection type
	ifaceName := parse.KnownStringPointer(planNativeRange.InterfaceName)
	if ifaceName == nil {
		ifaceName = parse.KnownStringPointer(planNativeRange.InterfaceIndex)
	}
	if ifaceName == nil {
		ifaceName = new(application.DefaultInterfaceIndex(plan.ConnectionType.ValueString()))
	}

	input = &application.InterfaceInput{

		DestType: interfaceDestType,
		Name:     ifaceName,
	}

	// MinLinks for LAG
	if utils.HasValue(planNativeRange.LagMinLinks) && slices.Contains(lagLinksDestTypes, interfaceDestType) {
		input.Lag = &application.LagInput{
			MinLinks: planNativeRange.LagMinLinks.ValueInt64(),
		}
	}

	// LAN input
	if slices.Contains(lanDestTypes, interfaceDestType) && !isHA {
		input.Lan = &application.LanInput{

			LocalIP:          planNativeRange.LocalIP.ValueString(),
			Subnet:           planNativeRange.NativeNetworkRange.ValueString(),
			TranslatedSubnet: translatedSubnetForAPIInput(configTranslatedSubnet, planNativeRange.TranslatedSubnet),
		}
	}

	// get interface index if configured, otherwise use a default based on connection type
	interfaceIndex := planNativeRange.InterfaceIndex.ValueString()
	if interfaceIndex == "" {
		interfaceIndex = application.DefaultInterfaceIndex(plan.ConnectionType.ValueString())
	}

	return input, interfaceIndex
}

func (r *socketSiteResource) isHA(state *SocketSite) bool {
	if state == nil || !utils.HasValue(state.Sockets) {
		return false
	}
	return len(state.Sockets.Elements()) > 1
}

// translatedSubnetForAPIInput omits translated_subnet from API payloads when the attribute is not
// explicitly set in Terraform config, even if plan/state still carry an API-hydrated value
// (e.g. equal to subnet when Static Range Translation is disabled).

func translatedSubnetForAPIInput(configValue, planValue types.String) *string {
	if configValue.IsNull() || configValue.IsUnknown() {
		return nil
	}
	return parse.StringPointerForOptionalInput(planValue)
}

var numberRE = regexp.MustCompile(`^\d+$`)

func connectionTypeFromSocketConfiguration(configuration *application.Configuration) types.String {
	if configuration == nil || configuration.Primary.Model == nil {
		return types.StringNull()
	}
	connections := map[string]string{

		"AWS":       "SOCKET_AWS1500",
		"AZURE":     "SOCKET_AZ1500",
		"ESX":       "SOCKET_ESX1500",
		"GCP":       "SOCKET_GCP1500",
		"X1500":     "SOCKET_X1500",
		"X1600":     "SOCKET_X1600",
		"X1600_LTE": "SOCKET_X1600_LTE",
		"X1700":     "SOCKET_X1700",
	}
	if value, ok := connections[*configuration.Primary.Model]; ok {
		return types.StringValue(value)
	}
	return types.StringNull()
}
func (r *socketSiteResource) parseSockets(
	ctx context.Context,
	configuration *application.Configuration,
	diags *diag.Diagnostics,
) types.Set {
	if configuration == nil {
		return types.SetNull(types.ObjectType{
			AttrTypes: SocketTypes,
		})
	}
	objects := make([]types.Object, 0, 2)
	primary := configuration.Primary
	objects = appendSocketConfiguration(ctx, objects, primary.Serial, primary.IsPrimary, primary.Platform, diags)
	if secondary := configuration.Secondary; secondary != nil {
		objects = appendSocketConfiguration(
			ctx,
			objects,
			secondary.Serial,
			secondary.IsPrimary,
			secondary.Platform,
			diags,
		)
	}
	result, ds := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: SocketTypes,
	}, objects)
	diags.Append(ds...)
	return result
}
func appendSocketConfiguration(ctx context.Context, sockets []types.Object, serial *string, isPrimary bool,
	platform *string, diags *diag.Diagnostics,
) []types.Object {
	tfSocketObj, objDiags := types.ObjectValueFrom(ctx, SocketTypes, Socket{

		ID: types.StringNull(),
		// SiteSocketConfiguration does not expose the socket entity ID.
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

func (r *socketSiteResource) parseSiteLocation(ctx context.Context, siteGenDetails *application.Snapshot,
	siteLocation types.Object, diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics
	var planLocation siteloc.SiteLocation

	if siteGenDetails == nil || siteGenDetails.Location == nil {
		return types.ObjectNull(siteloc.SiteLocationResourceAttrTypes)
	}
	siteLoc := siteGenDetails.Location

	// Prepare site location object
	state := siteLoc.StateCode
	if state != nil && *state == "" {
		state = nil
	}
	tfLocation := siteloc.SiteLocation{

		CountryCode: types.StringValue(siteLoc.CountryCode),
		StateCode:   types.StringPointerValue(state),
		Timezone:    types.StringValue(siteLoc.Timezone),
		Address:     types.StringPointerValue(siteLoc.Address),
		City:        types.StringPointerValue(siteLoc.City),
	}

	// API sometimes returns empty string and sometimes null for city - use what is in the plan in that case
	city := siteLoc.City
	if ((city == nil) || (*city == "")) && utils.HasValue(siteLocation) {
		if apperr.CheckErr(diags, siteLocation.As(ctx, &planLocation, basetypes.ObjectAsOptions{})) {
			return types.ObjectNull(siteloc.SiteLocationResourceAttrTypes)
		}
		if planLocation.City.IsNull() || (utils.HasValue(planLocation.City) && (planLocation.City.ValueString() == "")) {
			tfLocation.City = planLocation.City
		}
	}

	locObj, objDiags := types.ObjectValueFrom(ctx, siteloc.SiteLocationResourceAttrTypes, tfLocation)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(pops.PreferredPopLocationModelTypes)
	}

	return locObj
}

func (r *socketSiteResource) projectState(
	ctx context.Context,
	cfg *SocketSite,
	state SocketSite,
	snapshot *application.Snapshot,
	diags *diag.Diagnostics,
) SocketSite {
	return SocketSite{

		ID:             types.StringValue(snapshot.ID),
		Name:           types.StringValue(snapshot.Name),
		ConnectionType: connectionTypeFromSocketConfiguration(snapshot.Configuration),
		SiteType:       types.StringPointerValue(snapshot.SiteType),
		Description:    utils.StringPointerValue(snapshot.Description, state.Description),
		NativeRange:    r.parseNativeRange(ctx, cfg, snapshot.Range, snapshot.Interface, state.NativeRange, diags),
		SiteLocation:   r.parseSiteLocation(ctx, snapshot, state.SiteLocation, diags),
		Sockets:        r.parseSockets(ctx, snapshot.Configuration, diags),
	}
}
func (r *socketSiteResource) readInput(
	ctx context.Context,
	cfg *SocketSite,
	state SocketSite,
	id string,
	diags *diag.Diagnostics,
) application.ReadInput {
	in := application.ReadInput{

		ID:               id,
		ResolveRelayName: !r.checkDhcpSettingsDefault(ctx, cfg, state.NativeRange, diags),
	}
	if len(state.NativeRange.AttributeTypes(ctx)) > 0 && utils.HasValue(state.NativeRange) {
		var native NativeRange
		diags.Append(state.NativeRange.As(ctx, &native, basetypes.ObjectAsOptions{})...)
		in.InterfaceID = native.InterfaceID.ValueString()
	}
	return in
}
func (r *socketSiteResource) prepareInput(
	ctx context.Context,
	cfg, plan *SocketSite,
	state *SocketSite,
	diags *diag.Diagnostics,
) application.Input {
	ha := r.isHA(state)
	add := r.prepareSocketSiteInput(ctx, plan, diags)
	network := r.prepareNetworkRangeInput(ctx, cfg, plan, ha, diags)
	iface, index := r.prepareSocketInterfaceInput(ctx, cfg, plan, ha, diags)
	if diags.HasError() {
		return application.Input{}
	}
	var native NativeRange
	diags.Append(plan.NativeRange.As(ctx, &native, basetypes.ObjectAsOptions{})...)
	in := application.Input{

		ID:             plan.ID.ValueString(),
		RangeID:        native.NativeNetworkRangeID.ValueString(),
		DesiredIndex:   parse.KnownStringPointer(native.InterfaceIndex),
		InterfaceIndex: index,
		ConnectionType: plan.ConnectionType.ValueString(),
		IsHA:           ha,
		Add:            *add,
		Range:          *network,
		Interface:      *iface,
		Read:           r.readInput(ctx, cfg, *plan, plan.ID.ValueString(), diags),
	}
	in.General = application.GeneralInput{

		Name:         plan.Name.ValueStringPointer(),
		Description:  parse.KnownStringPointer(plan.Description),
		SiteType:     plan.SiteType.ValueStringPointer(),
		SiteLocation: r.prepareGeneralLocation(ctx, plan.SiteLocation, diags),
	}
	if state != nil && utils.HasValue(state.NativeRange) {
		var prior NativeRange
		diags.Append(state.NativeRange.As(ctx, &prior, basetypes.ObjectAsOptions{})...)
		in.CurrentIndex = prior.InterfaceIndex.ValueString()
	}
	return in
}

// prepareGeneralLocation preserves optional update fields, including unknown values.
func (r *socketSiteResource) prepareGeneralLocation(
	ctx context.Context,
	value types.Object,
	diags *diag.Diagnostics,
) *application.GeneralLocation {
	if !utils.HasValue(value) {
		return nil
	}
	var location siteloc.SiteLocation
	diags.Append(value.As(ctx, &location, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return &application.GeneralLocation{

		Address:     parse.KnownStringPointer(location.Address),
		City:        parse.KnownStringPointer(location.City),
		StateCode:   parse.KnownStringPointer(location.StateCode),
		CountryCode: location.CountryCode.ValueStringPointer(),
		Timezone:    location.Timezone.ValueStringPointer(),
	}
}
