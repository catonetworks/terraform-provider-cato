package waninterface

import (
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

func normalizedInterfaceIDCandidates(rawID string) []string {
	id := strings.TrimSpace(rawID)
	if id == "" {
		return nil
	}

	candidatesMap := map[string]struct{}{
		id: {},
	}

	upper := strings.ToUpper(id)
	if strings.HasPrefix(upper, "INT_") {
		candidatesMap[strings.TrimPrefix(upper, "INT_")] = struct{}{}
	}
	if strings.HasPrefix(upper, "WAN") {
		suffix := strings.TrimPrefix(upper, "WAN")
		if suffix != "" {
			candidatesMap[suffix] = struct{}{}
			candidatesMap["INT_"+suffix] = struct{}{}
		}
	}

	// If the raw ID is numeric (e.g., "1"), include common aliases.
	isNumeric := true
	for _, ch := range id {
		if ch < '0' || ch > '9' {
			isNumeric = false
			break
		}
	}
	if isNumeric {
		candidatesMap["INT_"+id] = struct{}{}
		candidatesMap["WAN"+id] = struct{}{}
	}

	candidates := make([]string, 0, len(candidatesMap))
	for candidate := range candidatesMap {
		candidates = append(candidates, strings.ToUpper(candidate))
	}

	return candidates
}

func wanInterfaceIDsMatch(resourceInterfaceID string, snapshotInterfaceID string, deviceInterfaceID string) bool {
	resourceCandidates := normalizedInterfaceIDCandidates(resourceInterfaceID)
	if len(resourceCandidates) == 0 {
		return false
	}

	otherCandidatesMap := map[string]struct{}{}
	for _, candidate := range normalizedInterfaceIDCandidates(snapshotInterfaceID) {
		otherCandidatesMap[candidate] = struct{}{}
	}
	for _, candidate := range normalizedInterfaceIDCandidates(deviceInterfaceID) {
		otherCandidatesMap[candidate] = struct{}{}
	}

	for _, candidate := range resourceCandidates {
		if _, ok := otherCandidatesMap[candidate]; ok {
			return true
		}
	}

	return false
}

func precedenceFromNaturalOrder(naturalOrder *int64) types.String {
	if naturalOrder == nil {
		return types.StringNull()
	}

	switch *naturalOrder {
	case wanInterfaceActiveNaturalOrder:
		return types.StringValue("ACTIVE")
	case wanInterfacePassiveNaturalOrder:
		return types.StringValue("PASSIVE")
	case wanInterfaceLastResortNaturalOrder:
		return types.StringValue("LAST_RESORT")
	default:
		return types.StringNull()
	}
}

func wanRoleFromInterfaceID(interfaceID string) string {
	switch strings.ToUpper(strings.TrimSpace(interfaceID)) {
	case "1", "WAN1", "INT_1":
		return "wan_1"
	case "2", "WAN2", "INT_2":
		return "wan_2"
	case "3", "WAN3", "INT_3":
		return "wan_3"
	case "4", "WAN4", "INT_4":
		return "wan_4"
	default:
		return ""
	}
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(plan WanInterface) (resultInput cato_models.UpdateSocketInterfaceInput) {
	resultInput = cato_models.UpdateSocketInterfaceInput{
		DestType: "CATO",
		Name:     plan.Name.ValueStringPointer(),
		Bandwidth: &cato_models.SocketInterfaceBandwidthInput{
			UpstreamBandwidth:   plan.UpstreamBandwidth.ValueInt64Pointer(),
			DownstreamBandwidth: plan.DownstreamBandwidth.ValueInt64Pointer(),
		},
		Wan: &cato_models.SocketInterfaceWanInput{
			Role:       cato_models.SocketInterfaceRole(plan.Role.ValueString()),
			Precedence: cato_models.SocketInterfacePrecedenceEnum(plan.Precedence.ValueString()),
		},
	}
	return resultInput
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(plan WanInterface) (resultInput cato_models.UpdateSocketInterfaceInput) {
	resultInput = cato_models.UpdateSocketInterfaceInput{
		DestType: "CATO",
		Name:     plan.Name.ValueStringPointer(),
		Bandwidth: &cato_models.SocketInterfaceBandwidthInput{
			UpstreamBandwidth:   plan.UpstreamBandwidth.ValueInt64Pointer(),
			DownstreamBandwidth: plan.DownstreamBandwidth.ValueInt64Pointer(),
		},
		Wan: &cato_models.SocketInterfaceWanInput{
			Role:       cato_models.SocketInterfaceRole(plan.Role.ValueString()),
			Precedence: cato_models.SocketInterfacePrecedenceEnum(plan.Precedence.ValueString()),
		},
	}
	return resultInput
}

// convertDeleteValues converts between the Terraform model and SDK values.
func convertDeleteValues(
	state WanInterface,
	accountSnapshotSite *cato_go_sdk.AccountSnapshot,
) (resultInput cato_models.UpdateSocketInterfaceInput) {
	var c = 0
	for _, item := range accountSnapshotSite.AccountSnapshot.Sites[0].InfoSiteSnapshot.Interfaces {
		if *item.DestType == "CATO" {
			c++
		}
	}

	if (c >= 1) && (state.Role == types.StringValue("wan_1")) {
		bandwidth := int64(10)
		resultInput = cato_models.UpdateSocketInterfaceInput{
			DestType: "CATO",
			Name:     state.InterfaceID.ValueStringPointer(),
			Bandwidth: &cato_models.SocketInterfaceBandwidthInput{
				UpstreamBandwidth:   &bandwidth,
				DownstreamBandwidth: &bandwidth,
			},
			Wan: &cato_models.SocketInterfaceWanInput{
				Role:       cato_models.SocketInterfaceRole("wan_1"),
				Precedence: cato_models.SocketInterfacePrecedenceEnum("ACTIVE"),
			},
		}
	} else {
		// Disabled interface to "remove" an interface
		resultInput = cato_models.UpdateSocketInterfaceInput{
			Name:     state.InterfaceID.ValueStringPointer(),
			DestType: "INTERFACE_DISABLED",
		}
	}
	return resultInput
}

// convertHydrateWanInterfaceStateValues converts between the Terraform model and SDK values.
func convertHydrateWanInterfaceStateValues(
	state *WanInterface,
	siteID string,
	site *cato_go_sdk.AccountSnapshot_AccountSnapshot_Sites,
	foundInterfaceIndex int,
) (
	resultFoundInterface *cato_go_sdk.AccountSnapshot_AccountSnapshot_Sites_InfoSiteSnapshot_Interfaces,
) {
	resultFoundInterface = site.InfoSiteSnapshot.Interfaces[foundInterfaceIndex]

	// Populate basic interface information
	if strings.HasPrefix(resultFoundInterface.ID, "WAN") ||
		strings.HasPrefix(resultFoundInterface.ID, "LTE") ||
		strings.HasPrefix(resultFoundInterface.ID, "USB") {
		state.InterfaceID = types.StringValue(resultFoundInterface.ID)
	} else {
		state.InterfaceID = types.StringValue("INT_" + resultFoundInterface.ID)
	}
	state.SiteID = types.StringValue(siteID)
	state.ID = types.StringValue(siteID + ":" + state.InterfaceID.ValueString())
	state.Name = types.StringValue(*resultFoundInterface.Name)
	state.UpstreamBandwidth = types.Int64Value(*resultFoundInterface.UpstreamBandwidth)
	state.DownstreamBandwidth = types.Int64Value(*resultFoundInterface.DownstreamBandwidth)
	if derivedRole := wanRoleFromInterfaceID(resultFoundInterface.ID); derivedRole != "" {
		state.Role = types.StringValue(derivedRole)
	}
	return resultFoundInterface
}
