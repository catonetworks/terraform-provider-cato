package laninterface

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"
	cast "github.com/spf13/cast"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	socketinterface "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/socketinterface"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func hydrateLanInterfaceAPI(ctx context.Context, cfg, plan LanInterface) cato_models.UpdateSocketInterfaceInput {
	tflog.Debug(ctx, "lan_interface update", map[string]interface{}{
		"plan.DestType.String()": utils.InterfaceToJSONString(plan.DestType.ValueString()),
		"plan.Name.String()":     utils.InterfaceToJSONString(plan.Name.ValueStringPointer()),
		"cato_models.SocketInterfaceDestType(plan.DestType.String())": cato_models.SocketInterfaceDestType(plan.DestType.ValueString()),
		"plan": utils.InterfaceToJSONString(plan),
	})

	destType := plan.DestType.ValueString()
	input := cato_models.UpdateSocketInterfaceInput{
		DestType: cato_models.SocketInterfaceDestType(destType),
		Name:     plan.Name.ValueStringPointer(),
	}

	// For LAN_LAG_MEMBER, only send basic interface info - no LAN, LAG, or VRRP config
	if destType != socketinterface.LanLagMemberDestType {
		// Add LAN configuration for non-LAG member interfaces
		input.Lan = &cato_models.SocketInterfaceLanInput{
			Subnet:           plan.Subnet.ValueString(),
			TranslatedSubnet: convert.TranslatedSubnetForAPIInput(cfg.TranslatedSubnet, plan.TranslatedSubnet),
			LocalIP:          plan.LocalIP.ValueString(),
		}

		// Add VRRP configuration if specified and not LAG member
		if !plan.VrrpType.IsNull() {
			input.Vrrp = &cato_models.SocketInterfaceVrrpInput{
				VrrpType: (*cato_models.VrrpType)(plan.VrrpType.ValueStringPointer()),
			}
		}
	}

	// Add LAG configuration only for LAG master types
	if (destType == lanLagMasterDestType || destType == lanLagMasterAndVrrpDestType) &&
		!plan.LagMinLinks.IsNull() &&
		!plan.LagMinLinks.IsUnknown() {
		input.Lag = &cato_models.SocketInterfaceLagInput{
			MinLinks: plan.LagMinLinks.ValueInt64(),
		}
	}

	return input
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(plan LanInterface) (resultSiteEntity *cato_models.EntityInput, resultZeroInt64 int64) {
	resultSiteEntity = &cato_models.EntityInput{Type: "site", ID: plan.SiteID.ValueString()}
	resultZeroInt64 = int64(0)
	return resultSiteEntity, resultZeroInt64
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state LanInterface) (resultInput cato_models.UpdateSocketInterfaceInput) {
	resultInput = cato_models.UpdateSocketInterfaceInput{
		Name:     state.InterfaceID.ValueStringPointer(),
		DestType: "INTERFACE_DISABLED",
	}
	return resultInput
}

// convertCreateValues converts between the Terraform model and SDK values.
func convertCreateValues(plan *LanInterface, item *cato_go_sdk.EntityLookup_EntityLookup_Items) {
	plan.ID = types.StringValue(item.GetEntity().GetID())
}

// convertDeleteValues converts between the Terraform model and SDK values.
func convertDeleteValues(originalCurInterfaceID *string) (resultInput cato_models.UpdateSocketInterfaceInput) {
	curInterfaceID := *originalCurInterfaceID
	defer func() { *originalCurInterfaceID = curInterfaceID }()

	resultInput = cato_models.UpdateSocketInterfaceInput{
		Name:     &curInterfaceID,
		DestType: "INTERFACE_DISABLED",
	}
	return resultInput
}

// convertHydrateLanInterfaceStateValues converts between the Terraform model and SDK values.
func convertHydrateLanInterfaceStateValues(
	state *LanInterface,
	priorIsPresent *bool,
	curIint *cato_go_sdk.EntityLookup_EntityLookup_Items,
) {
	isPresent := *priorIsPresent
	defer func() { *priorIsPresent = isPresent }()
	isPresent = true
	state.ID = types.StringValue(curIint.Entity.ID)
	if siteIDVal, ok := curIint.HelperFields["siteId"]; ok {
		state.SiteID = types.StringValue(cast.ToString(siteIDVal))
	}
}

// convertHydrateLanInterfaceStateValues2 converts between the Terraform model and SDK values.
func convertHydrateLanInterfaceStateValues2(state *LanInterface, curIint *cato_go_sdk.EntityLookup_EntityLookup_Items) {
	if nameVal, ok := curIint.HelperFields["interfaceName"]; ok {
		state.Name = types.StringValue(cast.ToString(nameVal))
	}
	if destTypeVal, ok := curIint.HelperFields["destType"]; ok {
		state.DestType = types.StringValue(cast.ToString(destTypeVal))
	}
	if subnetVal, ok := curIint.HelperFields["subnet"]; ok {
		state.Subnet = types.StringValue(cast.ToString(subnetVal))
	}

	// Populate vrrp_type from networkInterface if available
	if vrrpTypeVal, ok := curIint.HelperFields["vrrpType"]; ok {
		vrrpTypeStr := cast.ToString(vrrpTypeVal)
		if vrrpTypeStr != "" {
			state.VrrpType = types.StringValue(vrrpTypeStr)
		} else {
			state.VrrpType = types.StringNull()
		}
	} else {
		state.VrrpType = types.StringNull()
	}
}

// convertHydrateLanInterfaceStateValues3 converts between the Terraform model and SDK values.
func convertHydrateLanInterfaceStateValues3(state *LanInterface, idxInt int) {
	state.InterfaceID = types.StringValue(fmt.Sprintf("INT_%d", idxInt))
}

// convertHydrateLanInterfaceStateValues4 converts between the Terraform model and SDK values.
func convertHydrateLanInterfaceStateValues4(state *LanInterface, curIint *cato_go_sdk.EntityLookup_EntityLookup_Items) {
	state.InterfaceID = types.StringValue(cast.ToString(curIint.HelperFields["interfaceID"]))
}

// convertHydrateLanInterfaceStateValues5 converts between the Terraform model and SDK values.
func convertHydrateLanInterfaceStateValues5(
	state *LanInterface,
	lagMinLinks int,
	siteAccountSnapshotAPIData *cato_go_sdk.AccountSnapshot,
) {
	for _, site := range siteAccountSnapshotAPIData.AccountSnapshot.GetSites() {
		siteID := site.GetID()
		if siteID != nil && state.SiteID.ValueString() == *siteID {
			for _, iface := range site.InfoSiteSnapshot.Interfaces {
				if iface.DestType != nil && *iface.DestType == socketinterface.LanLagMemberDestType {
					lagMinLinks++
				}
			}
		}
	}
	state.LagMinLinks = types.Int64Value(int64(lagMinLinks))
}

// convertHydrateLanInterfaceStateValues6 converts between the Terraform model and SDK values.
func convertHydrateLanInterfaceStateValues6(
	state *LanInterface,
) (resultSiteEntity *cato_models.EntityInput, resultZeroInt64 int64, resultThouInt64 int64) {
	resultSiteEntity = &cato_models.EntityInput{Type: "site", ID: state.SiteID.ValueString()}
	resultZeroInt64 = int64(0)
	resultThouInt64 = int64(1000)
	return resultSiteEntity, resultZeroInt64, resultThouInt64
}

// convertHydrateLanInterfaceStateValues7 converts between the Terraform model and SDK values.
func convertHydrateLanInterfaceStateValues7(state *LanInterface, v *cato_go_sdk.EntityLookup_EntityLookup_Items) {
	if translatedSubnetVal, ok := v.HelperFields["translatedSubnet"]; ok {
		translatedSubnetStr := cast.ToString(translatedSubnetVal)
		if translatedSubnetStr != "" && translatedSubnetStr != state.Subnet.ValueString() {
			state.TranslatedSubnet = types.StringValue(translatedSubnetStr)
		} else {
			state.TranslatedSubnet = types.StringNull()
		}
	} else {
		state.TranslatedSubnet = types.StringNull()
	}

	// Populate local_ip from gateway field in siteRange
	// For LAN interfaces, localIp is the same as gateway
	if gatewayVal, ok := v.HelperFields["gateway"]; ok && gatewayVal != nil {
		gatewayStr := cast.ToString(gatewayVal)
		if gatewayStr != "" {
			state.LocalIP = types.StringValue(gatewayStr)
		} else {
			state.LocalIP = types.StringNull()
		}
	} else {
		state.LocalIP = types.StringNull()
	}
}
