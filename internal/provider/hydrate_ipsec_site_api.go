package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// hydrateAddIpsecIkeV2Site takes the plan and returns AddIpsecIkeV2SiteInput for site creation
func hydrateAddIpsecIkeV2Site(ctx context.Context, plan SiteIpsecIkeV2) (cato_models.AddIpsecIkeV2SiteInput, diag.Diagnostics) {
	diags := []diag.Diagnostic{}
	input := cato_models.AddIpsecIkeV2SiteInput{}

	// Set site location
	if !plan.SiteLocation.IsNull() {
		input.SiteLocation = &cato_models.AddSiteLocationInput{}
		siteLocationInput := AddIpsecSiteLocationInput{}
		diags = append(diags, plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})...)

		input.SiteLocation.Address = siteLocationInput.Address.ValueStringPointer()
		input.SiteLocation.City = siteLocationInput.City.ValueStringPointer()
		input.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueString()
		input.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()
		input.SiteLocation.Timezone = siteLocationInput.Timezone.ValueString()
	}

	// Set other site attributes
	input.Name = plan.Name.ValueString()
	input.SiteType = cato_models.SiteType(plan.SiteType.ValueString())
	input.NativeNetworkRange = plan.NativeNetworkRange.ValueString()
	input.Description = plan.Description.ValueStringPointer()

	return input, diags
}

// hydrateIpsecTunnelsAPITypes holds both Add and Update tunnel inputs
type hydrateIpsecTunnelsAPITypes struct {
	add    cato_models.AddIpsecIkeV2SiteTunnelsInput
	update cato_models.UpdateIpsecIkeV2SiteTunnelsInput
}

// hydrateAddIpsecIkeV2SiteTunnels takes the plan and returns both Add and Update inputs for tunnels
//
//nolint:funlen
func hydrateAddIpsecIkeV2SiteTunnels(ctx context.Context, plan SiteIpsecIkeV2) (hydrateIpsecTunnelsAPITypes, diag.Diagnostics) {
	diags := []diag.Diagnostic{}
	result := hydrateIpsecTunnelsAPITypes{
		add:    cato_models.AddIpsecIkeV2SiteTunnelsInput{},
		update: cato_models.UpdateIpsecIkeV2SiteTunnelsInput{},
	}

	if plan.IPSec.IsNull() {
		return result, diags
	}

	planIPSec := AddIpsecIkeV2SiteTunnelsInput{}
	diags = append(diags, plan.IPSec.As(ctx, &planIPSec, basetypes.ObjectAsOptions{})...)

	// Process Primary tunnels
	if !planIPSec.Primary.IsNull() {
		result.add.Primary = &cato_models.AddIpsecIkeV2TunnelsInput{}
		result.update.Primary = &cato_models.UpdateIpsecIkeV2TunnelsInput{}

		planIPSecPrimary := AddIpsecIkeV2TunnelsInput{}
		diags = append(diags, planIPSec.Primary.As(ctx, &planIPSecPrimary, basetypes.ObjectAsOptions{})...)

		// Set destination type, pop location, and public Cato IP
		result.add.Primary.DestinationType = (*cato_models.DestinationType)(planIPSecPrimary.DestinationType.ValueStringPointer())
		result.add.Primary.PopLocationID = planIPSecPrimary.PopLocationID.ValueStringPointer()
		result.add.Primary.PublicCatoIPID = planIPSecPrimary.PublicCatoIPID.ValueStringPointer()

		result.update.Primary.DestinationType = (*cato_models.DestinationType)(planIPSecPrimary.DestinationType.ValueStringPointer())
		result.update.Primary.PopLocationID = planIPSecPrimary.PopLocationID.ValueStringPointer()
		result.update.Primary.PublicCatoIPID = planIPSecPrimary.PublicCatoIPID.ValueStringPointer()

		// Process tunnels
		if !planIPSecPrimary.Tunnels.IsNull() {
			elementsTunnels := make([]basetypes.ObjectValue, 0, len(planIPSecPrimary.Tunnels.Elements()))
			diags = append(diags, planIPSecPrimary.Tunnels.ElementsAs(ctx, &elementsTunnels, false)...)

			for _, item := range elementsTunnels {
				var itemTunnels AddIpsecIkeV2TunnelInput
				diags = append(diags, item.As(ctx, &itemTunnels, basetypes.ObjectAsOptions{})...)

				// Process last mile bandwidth
				var itemTunnelLastMileBw LastMileBwInput
				if !itemTunnels.LastMileBw.IsNull() {
					diags = append(diags, itemTunnels.LastMileBw.As(ctx, &itemTunnelLastMileBw, basetypes.ObjectAsOptions{})...)
				}

				// Append to Add input
				result.add.Primary.Tunnels = append(result.add.Primary.Tunnels, &cato_models.AddIpsecIkeV2TunnelInput{
					LastMileBw: &cato_models.LastMileBwInput{
						Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
						Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
					},
					PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
					PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
					Psk:           itemTunnels.Psk.ValueString(),
					PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
				})

				// Append to Update input
				result.update.Primary.Tunnels = append(result.update.Primary.Tunnels, &cato_models.UpdateIpsecIkeV2TunnelInput{
					LastMileBw: &cato_models.LastMileBwInput{
						Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
						Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
					},
					PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
					PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
					Psk:           itemTunnels.Psk.ValueStringPointer(),
					PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
					TunnelID:      cato_models.IPSecV2InterfaceID(itemTunnels.TunnelID.ValueString()),
				})
			}
		}
	}

	// Process Secondary tunnels
	if !planIPSec.Secondary.IsNull() {
		result.add.Secondary = &cato_models.AddIpsecIkeV2TunnelsInput{}
		result.update.Secondary = &cato_models.UpdateIpsecIkeV2TunnelsInput{}

		planIPSecSecondary := AddIpsecIkeV2TunnelsInput{}
		diags = append(diags, planIPSec.Secondary.As(ctx, &planIPSecSecondary, basetypes.ObjectAsOptions{})...)

		// Set destination type, pop location, and public Cato IP
		result.add.Secondary.DestinationType = (*cato_models.DestinationType)(planIPSecSecondary.DestinationType.ValueStringPointer())
		result.add.Secondary.PopLocationID = planIPSecSecondary.PopLocationID.ValueStringPointer()
		result.add.Secondary.PublicCatoIPID = planIPSecSecondary.PublicCatoIPID.ValueStringPointer()

		result.update.Secondary.DestinationType = (*cato_models.DestinationType)(planIPSecSecondary.DestinationType.ValueStringPointer())
		result.update.Secondary.PopLocationID = planIPSecSecondary.PopLocationID.ValueStringPointer()
		result.update.Secondary.PublicCatoIPID = planIPSecSecondary.PublicCatoIPID.ValueStringPointer()

		// Process tunnels
		if !planIPSecSecondary.Tunnels.IsNull() {
			elementsTunnels := make([]basetypes.ObjectValue, 0, len(planIPSecSecondary.Tunnels.Elements()))
			diags = append(diags, planIPSecSecondary.Tunnels.ElementsAs(ctx, &elementsTunnels, false)...)

			for _, item := range elementsTunnels {
				var itemTunnels AddIpsecIkeV2TunnelInput
				diags = append(diags, item.As(ctx, &itemTunnels, basetypes.ObjectAsOptions{})...)

				// Process last mile bandwidth
				var itemTunnelLastMileBw LastMileBwInput
				if !itemTunnels.LastMileBw.IsNull() {
					diags = append(diags, itemTunnels.LastMileBw.As(ctx, &itemTunnelLastMileBw, basetypes.ObjectAsOptions{})...)
				}

				// Append to Add input
				result.add.Secondary.Tunnels = append(result.add.Secondary.Tunnels, &cato_models.AddIpsecIkeV2TunnelInput{
					LastMileBw: &cato_models.LastMileBwInput{
						Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
						Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
					},
					PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
					PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
					Psk:           itemTunnels.Psk.ValueString(),
					PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
				})

				// Append to Update input
				result.update.Secondary.Tunnels = append(result.update.Secondary.Tunnels, &cato_models.UpdateIpsecIkeV2TunnelInput{
					LastMileBw: &cato_models.LastMileBwInput{
						Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
						Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
					},
					PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
					PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
					Psk:           itemTunnels.Psk.ValueStringPointer(),
					PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
					TunnelID:      cato_models.IPSecV2InterfaceID(itemTunnels.TunnelID.ValueString()),
				})
			}
		}
	}

	return result, diags
}

// hydrateUpdateIpsecIkeV2SiteTunnels builds the update input and restores computed
// tunnel IDs from prior state when Terraform marks nested computed values unknown.
func hydrateUpdateIpsecIkeV2SiteTunnels(
	ctx context.Context,
	plan SiteIpsecIkeV2,
	state SiteIpsecIkeV2,
) (cato_models.UpdateIpsecIkeV2SiteTunnelsInput, diag.Diagnostics) {
	result, diags := hydrateAddIpsecIkeV2SiteTunnels(ctx, plan)
	if diags.HasError() {
		return result.update, diags
	}

	stateIPSec := AddIpsecIkeV2SiteTunnelsInput{}
	diags = append(diags, state.IPSec.As(ctx, &stateIPSec, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return result.update, diags
	}

	diags = append(diags, restoreIpsecTunnelIDs(ctx, result.update.Primary, stateIPSec.Primary, "primary")...)
	diags = append(diags, restoreIpsecTunnelIDs(ctx, result.update.Secondary, stateIPSec.Secondary, "secondary")...)

	return result.update, diags
}

func restoreIpsecTunnelIDs(
	ctx context.Context,
	update *cato_models.UpdateIpsecIkeV2TunnelsInput,
	stateGroup basetypes.ObjectValue,
	groupName string,
) diag.Diagnostics {
	var diags diag.Diagnostics
	if update == nil || len(update.Tunnels) == 0 {
		return diags
	}

	stateTunnelIDs, stateDiags := ipsecTunnelIDsFromState(ctx, stateGroup)
	diags = append(diags, stateDiags...)
	if diags.HasError() {
		return diags
	}

	for index, tunnel := range update.Tunnels {
		if tunnel.TunnelID != "" {
			continue
		}

		if index < len(stateTunnelIDs) && stateTunnelIDs[index] != "" {
			tunnel.TunnelID = stateTunnelIDs[index]
			continue
		}

		if derivedID, ok := positionalIpsecTunnelID(groupName, index); ok {
			tunnel.TunnelID = derivedID
			continue
		}

		diags = append(diags, diag.NewErrorDiagnostic(
			"Missing IPSec tunnel ID",
			"The provider could not determine a valid tunnel ID for the "+
				groupName+" tunnel. IPSec supports at most three tunnels per group.",
		))
	}

	return diags
}

func ipsecTunnelIDsFromState(
	ctx context.Context,
	stateGroup basetypes.ObjectValue,
) ([]cato_models.IPSecV2InterfaceID, diag.Diagnostics) {
	var diags diag.Diagnostics
	if stateGroup.IsNull() || stateGroup.IsUnknown() {
		return nil, diags
	}

	group := AddIpsecIkeV2TunnelsInput{}
	diags = append(diags, stateGroup.As(ctx, &group, basetypes.ObjectAsOptions{})...)
	if diags.HasError() || group.Tunnels.IsNull() || group.Tunnels.IsUnknown() {
		return nil, diags
	}

	stateTunnels := make([]basetypes.ObjectValue, 0, len(group.Tunnels.Elements()))
	diags = append(diags, group.Tunnels.ElementsAs(ctx, &stateTunnels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	tunnelIDs := make([]cato_models.IPSecV2InterfaceID, len(stateTunnels))
	for index, stateTunnelValue := range stateTunnels {
		stateTunnel := AddIpsecIkeV2TunnelInput{}
		diags = append(diags, stateTunnelValue.As(ctx, &stateTunnel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		if !stateTunnel.TunnelID.IsNull() && !stateTunnel.TunnelID.IsUnknown() {
			tunnelIDs[index] = cato_models.IPSecV2InterfaceID(stateTunnel.TunnelID.ValueString())
		}
	}

	return tunnelIDs, diags
}

func positionalIpsecTunnelID(groupName string, index int) (cato_models.IPSecV2InterfaceID, bool) {
	tunnelIDs := map[string][]cato_models.IPSecV2InterfaceID{
		"primary":   {"PRIMARY1", "PRIMARY2", "PRIMARY3"},
		"secondary": {"SECONDARY1", "SECONDARY2", "SECONDARY3"},
	}

	groupIDs, ok := tunnelIDs[groupName]
	if !ok || index < 0 || index >= len(groupIDs) {
		return "", false
	}
	return groupIDs[index], true
}

// hydrateUpdateIpsecIkeV2SiteGeneralDetails takes the plan and returns UpdateIpsecIkeV2SiteGeneralDetailsInput
//
//nolint:gocyclo,funlen
func hydrateUpdateIpsecIkeV2SiteGeneralDetails(
	ctx context.Context,
	plan SiteIpsecIkeV2,
) (cato_models.UpdateIpsecIkeV2SiteGeneralDetailsInput, diag.Diagnostics) {
	diags := []diag.Diagnostic{}
	input := cato_models.UpdateIpsecIkeV2SiteGeneralDetailsInput{}

	if plan.IPSec.IsNull() {
		return input, diags
	}

	planIPSec := AddIpsecIkeV2SiteTunnelsInput{}
	diags = append(diags, plan.IPSec.As(ctx, &planIPSec, basetypes.ObjectAsOptions{})...)

	// Set connection mode
	var connectionModeValue string
	if !planIPSec.ConnectionMode.IsNull() {
		connectionModeValue = planIPSec.ConnectionMode.ValueString()
		connectionMode := cato_models.ConnectionMode(connectionModeValue)
		input.ConnectionMode = &connectionMode
	}

	// Set identification type (only applicable when connection_mode is RESPONDER_ONLY)
	if !planIPSec.IdentificationType.IsNull() {
		// Only set if connection_mode is RESPONDER_ONLY
		if connectionModeValue == "RESPONDER_ONLY" {
			identificationType := cato_models.IdentificationType(planIPSec.IdentificationType.ValueString())
			input.IdentificationType = &identificationType
		}
		// If connection_mode is not RESPONDER_ONLY and identification_type is set, it will be ignored
		// This allows users to set it in config but it won't be sent to API unless connection_mode is RESPONDER_ONLY
	}

	// Set init message
	if !planIPSec.InitMessage.IsNull() {
		planInitMessage := IpsecIkeV2MessageInput{}
		diags = append(diags, planIPSec.InitMessage.As(ctx, &planInitMessage, basetypes.ObjectAsOptions{})...)

		input.InitMessage = &cato_models.IpsecIkeV2MessageInput{}

		if !planInitMessage.Cipher.IsNull() {
			cipher := cato_models.IPSecCipher(planInitMessage.Cipher.ValueString())
			input.InitMessage.Cipher = &cipher
		}

		if !planInitMessage.DhGroup.IsNull() {
			dhGroup := cato_models.IPSecDHGroup(planInitMessage.DhGroup.ValueString())
			input.InitMessage.DhGroup = &dhGroup
		}

		if !planInitMessage.Integrity.IsNull() {
			integrity := cato_models.IPSecHash(planInitMessage.Integrity.ValueString())
			input.InitMessage.Integrity = &integrity
		}

		if !planInitMessage.Prf.IsNull() {
			prf := cato_models.IPSecHash(planInitMessage.Prf.ValueString())
			input.InitMessage.Prf = &prf
		}
	}

	// Set auth message
	if !planIPSec.AuthMessage.IsNull() {
		planAuthMessage := IpsecIkeV2MessageInput{}
		diags = append(diags, planIPSec.AuthMessage.As(ctx, &planAuthMessage, basetypes.ObjectAsOptions{})...)

		input.AuthMessage = &cato_models.IpsecIkeV2MessageInput{}

		if !planAuthMessage.Cipher.IsNull() {
			cipher := cato_models.IPSecCipher(planAuthMessage.Cipher.ValueString())
			input.AuthMessage.Cipher = &cipher
		}

		if !planAuthMessage.DhGroup.IsNull() {
			dhGroup := cato_models.IPSecDHGroup(planAuthMessage.DhGroup.ValueString())
			input.AuthMessage.DhGroup = &dhGroup
		}

		if !planAuthMessage.Integrity.IsNull() {
			integrity := cato_models.IPSecHash(planAuthMessage.Integrity.ValueString())
			input.AuthMessage.Integrity = &integrity
		}

		if !planAuthMessage.Prf.IsNull() {
			prf := cato_models.IPSecHash(planAuthMessage.Prf.ValueString())
			input.AuthMessage.Prf = &prf
		}
	}

	// Set network ranges - convert from list of strings
	if !planIPSec.NetworkRanges.IsNull() && !planIPSec.NetworkRanges.IsUnknown() {
		var networkRangesList []string
		diags = append(diags, planIPSec.NetworkRanges.ElementsAs(ctx, &networkRangesList, false)...)

		if len(networkRangesList) > 0 {
			networkRanges := make([]*string, 0, len(networkRangesList))
			for _, r := range networkRangesList {
				if r != "" {
					rangeCopy := r
					networkRanges = append(networkRanges, &rangeCopy)
				}
			}
			if len(networkRanges) > 0 {
				input.NetworkRanges = networkRanges
			}
		}
	}

	return input, diags
}
