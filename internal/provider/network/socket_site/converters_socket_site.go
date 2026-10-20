package socketsite

import (
	"context"
	"slices"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tf_model"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// prepareSiteLocation constructs the API input (location part) for SiteAddSocketSite() from the Terraform plan data.
func prepareSiteLocation(ctx context.Context, location types.Object, diags *diag.Diagnostics) *cato_models.AddSiteLocationInput {
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
	input := &cato_models.UpdateNetworkRangeInput{
		Subnet:           convert.KnownStringPointer(planNativeRange.NativeNetworkRange),
		TranslatedSubnet: convert.TranslatedSubnetForAPIInput(configTranslatedSubnet, planNativeRange.TranslatedSubnet),
		MdnsReflector:    convert.KnownBoolPointer(planNativeRange.MdnsReflector),
		Vlan:             convert.KnownInt64Pointer(planNativeRange.Vlan),
		DhcpSettings:     dhcp.PrepareDHCPSettings(ctx, client, cato_models.SubnetTypeNative, planNativeRange.DhcpSettings, diags),
		// AzureFloatingIP must remain omitted until Azure HA support is implemented.
	}

	if !isHA { // for HA scenario, local IP is not allowed to be modified
		input.LocalIP = convert.KnownStringPointer(planNativeRange.LocalIP)
	}

	return input
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
