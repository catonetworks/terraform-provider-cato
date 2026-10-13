package laninterface

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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
	if destType != lanLagMemberDestType {
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
