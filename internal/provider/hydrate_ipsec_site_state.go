package provider

import (
	"context"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

func (r *siteIpsecResource) hydrateIpsecSiteState(ctx context.Context, state SiteIpsecIkeV2, siteID string) (SiteIpsecIkeV2, bool, error) {
	// Check if site exists
	tflog.Debug(ctx, "hydrateIpsecSiteState.EntityLookup.request", map[string]interface{}{
		"siteID":     utils.InterfaceToJSONString(siteID),
		"EntityType": utils.InterfaceToJSONString(cato_models.EntityType("site")),
	})
	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{siteID}, nil, nil, nil)
	tflog.Debug(ctx, "hydrateIpsecSiteState.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		return state, false, err
	}

	// Check if site exists
	if len(querySiteResult.EntityLookup.GetItems()) != 1 {
		return state, false, nil
	}

	// Get site entity
	var siteEntity *cato_go_sdk.EntityLookup_EntityLookup_Items
	for _, v := range querySiteResult.EntityLookup.Items {
		if v.Entity.ID == siteID {
			siteEntity = v
			break
		}
	}

	if siteEntity == nil {
		return state, false, nil
	}

	// Get AccountSnapshot for detailed site information
	siteAccountSnapshotApiData, err := r.client.catov2.AccountSnapshot(ctx, []string{siteID}, nil, &r.client.AccountId)
	tflog.Debug(ctx, "hydrateIpsecSiteState.AccountSnapshot.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteAccountSnapshotApiData),
	})
	if err != nil {
		return state, false, err
	}

	if len(siteAccountSnapshotApiData.GetAccountSnapshot().GetSites()) == 0 {
		return state, false, nil
	}

	thisSite := siteAccountSnapshotApiData.GetAccountSnapshot().GetSites()[0]

	// Set basic site information
	state.ID = types.StringValue(siteEntity.Entity.GetID())

	if siteEntity.Entity.Name != nil {
		state.Name = types.StringValue(*siteEntity.Entity.Name)
	}

	if siteType, ok := siteEntity.GetHelperFields()["type"]; ok && siteType != nil {
		state.SiteType = types.StringValue(siteType.(string))
	}

	if description, ok := siteEntity.GetHelperFields()["description"]; ok && description != nil {
		descStr := description.(string)
		if descStr != "" {
			state.Description = types.StringValue(descStr)
		} else {
			state.Description = types.StringNull()
		}
	}

	// Get native network range ID and interface ID
	siteEntityInput := &cato_models.EntityInput{Type: "site", ID: siteID}
	siteRangeEntities, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"), nil, nil, siteEntityInput, nil, nil, nil, nil, nil)
	tflog.Debug(ctx, "hydrateIpsecSiteState.EntityLookup.siteRange.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteRangeEntities),
	})
	if err != nil {
		return state, false, err
	}

	// Find native network range and interface ID
	for _, item := range siteRangeEntities.EntityLookup.Items {
		if item.Entity.Name != nil {
			splitName := strings.Split(*item.Entity.Name, " \\ ")
			if len(splitName) >= 3 && splitName[2] == "Native Range" {
				state.NativeNetworkRangeId = types.StringValue(item.Entity.ID)
				if subnet, ok := item.HelperFields["subnet"]; ok && subnet != nil {
					state.NativeNetworkRange = types.StringValue(subnet.(string))
				}
				break
			}
		}
	}

	// Get native interface ID
	networkInterfaceEntities, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), nil, nil, siteEntityInput, nil, nil, nil, nil, nil)
	tflog.Debug(ctx, "hydrateIpsecSiteState.EntityLookup.networkInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(networkInterfaceEntities),
	})
	if err != nil {
		return state, false, err
	}

	// Find native network interface
	for _, curIint := range networkInterfaceEntities.EntityLookup.Items {
		curSiteId := cast.ToString(curIint.HelperFields["siteId"])
		if curSiteId == siteID {
			state.InterfaceId = types.StringValue(curIint.Entity.ID)
			break
		}
	}

	// Set site location from AccountSnapshot
	if thisSite.InfoSiteSnapshot != nil {
		siteLocationAttrs := map[string]attr.Value{
			"country_code": types.StringNull(),
			"state_code":   types.StringNull(),
			"timezone":     types.StringNull(),
			"address":      types.StringNull(),
			"city":         types.StringNull(),
		}

		if thisSite.InfoSiteSnapshot.CountryCode != nil && *thisSite.InfoSiteSnapshot.CountryCode != "" {
			siteLocationAttrs["country_code"] = types.StringValue(*thisSite.InfoSiteSnapshot.CountryCode)
		}
		if thisSite.InfoSiteSnapshot.CountryStateName != nil && *thisSite.InfoSiteSnapshot.CountryStateName != "" {
			// Map state name back to state code if needed - for now just use the state code from config
			// The AccountSnapshot returns the full state name, not the code
			// We'll preserve whatever is in the current state
			if !state.SiteLocation.IsNull() {
				var currentLocation SiteLocation
				state.SiteLocation.As(ctx, &currentLocation, basetypes.ObjectAsOptions{})
				if !currentLocation.StateCode.IsNull() {
					siteLocationAttrs["state_code"] = currentLocation.StateCode
				}
			}
		}
		// Note: AccountSnapshot doesn't have timezone, address separately
		// We preserve from state if they exist
		if !state.SiteLocation.IsNull() {
			var currentLocation SiteLocation
			state.SiteLocation.As(ctx, &currentLocation, basetypes.ObjectAsOptions{})
			if !currentLocation.Timezone.IsNull() {
				siteLocationAttrs["timezone"] = currentLocation.Timezone
			}
			if !currentLocation.Address.IsNull() {
				siteLocationAttrs["address"] = currentLocation.Address
			}
		}
		if thisSite.InfoSiteSnapshot.CityName != nil && *thisSite.InfoSiteSnapshot.CityName != "" {
			siteLocationAttrs["city"] = types.StringValue(*thisSite.InfoSiteSnapshot.CityName)
		}

		siteLocationObj, diags := types.ObjectValue(SiteLocationResourceAttrTypes, siteLocationAttrs)
		if diags.HasError() {
			return state, false, nil
		}
		state.SiteLocation = siteLocationObj
	}

	// Set IPSec configuration from AccountSnapshot
	// Note: AccountSnapshot has limited IPSec info (tunnels, IPs)
	// All IPSec config fields (tunnels, init_message, auth_message, connection_mode, etc.) are preserved from state
	// since the API does not return these values in AccountSnapshot
	ipsecAttrs := map[string]attr.Value{
		"site_id":             types.StringValue(siteID),
		"primary":             types.ObjectNull(IpsecTunnelsResourceAttrTypes),
		"secondary":           types.ObjectNull(IpsecTunnelsResourceAttrTypes),
		"connection_mode":     types.StringNull(),
		"identification_type": types.StringNull(),
		"init_message":        types.ObjectNull(IpsecMessageResourceAttrTypes),
		"auth_message":        types.ObjectNull(IpsecMessageResourceAttrTypes),
		"network_ranges":      types.ListNull(types.StringType),
	}

	// Preserve all IPSec fields from state (not returned by AccountSnapshot API)
	if !state.IPSec.IsNull() {
		var currentIpsec AddIpsecIkeV2SiteTunnelsInput
		state.IPSec.As(ctx, &currentIpsec, basetypes.ObjectAsOptions{})

		// Preserve primary tunnels
		if !currentIpsec.Primary.IsNull() {
			ipsecAttrs["primary"] = currentIpsec.Primary
		}

		// Preserve secondary tunnels
		if !currentIpsec.Secondary.IsNull() {
			ipsecAttrs["secondary"] = currentIpsec.Secondary
		}

		// Preserve connection mode
		if !currentIpsec.ConnectionMode.IsNull() {
			ipsecAttrs["connection_mode"] = currentIpsec.ConnectionMode
		}

		// Preserve identification type
		if !currentIpsec.IdentificationType.IsNull() {
			ipsecAttrs["identification_type"] = currentIpsec.IdentificationType
		}

		// Preserve init message
		if !currentIpsec.InitMessage.IsNull() {
			ipsecAttrs["init_message"] = currentIpsec.InitMessage
		}

		// Preserve auth message
		if !currentIpsec.AuthMessage.IsNull() {
			ipsecAttrs["auth_message"] = currentIpsec.AuthMessage
		}

		// Preserve network ranges
		if !currentIpsec.NetworkRanges.IsNull() {
			ipsecAttrs["network_ranges"] = currentIpsec.NetworkRanges
		}
	}

	// Create IPSec object
	ipsecObj, diags := types.ObjectValue(IpsecResourceAttrTypes, ipsecAttrs)
	if diags.HasError() {
		return state, false, nil
	}
	state.IPSec = ipsecObj

	return state, true, nil
}
