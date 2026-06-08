package validators

import (
	"context"
	"errors"
	"fmt"
	"slices"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// Connection types where it is allowed to specify interface index
var connTypesWithInterfaceIndex = []cato_models.SiteConnectionTypeEnum{
	cato_models.SiteConnectionTypeEnumSocketX1500,
	cato_models.SiteConnectionTypeEnumSocketX1600,
	cato_models.SiteConnectionTypeEnumSocketX1600Lte,
	cato_models.SiteConnectionTypeEnumSocketX1700,
}

// ErrConfig generic configuration error
var ErrConfig = errors.New("configuration error")

func GetNativeRangeValidator() NativeRangeValidator {
	return NativeRangeValidator{}
}

// NativeRangeValidator validates the NativeRange settings
type NativeRangeValidator struct{}

func (v NativeRangeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	var nativeRange tf.NativeRange
	var connectionType types.String

	if !utils.HasValue(req.ConfigValue) {
		return
	}

	// get native range
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.As(ctx, &nativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	if v.isUnknown(nativeRange) {
		return
	}

	// get connection type
	if utils.CheckErr(&resp.Diagnostics, req.Config.GetAttribute(ctx, path.Root("connection_type"), &connectionType)) {
		return
	}
	if connectionType.IsUnknown() {
		return
	}

	// Validate that connection_type allows interface_index to be specified
	if v.checkInterfaceIndex(&resp.Diagnostics, nativeRange, cato_models.SiteConnectionTypeEnum(connectionType.ValueString())) != nil {
		return
	}

	// Validate that local_ip is within native_network_range
	if checkLocalIP(&resp.Diagnostics, nativeRange.LocalIP, nativeRange.NativeNetworkRange) != nil {
		return
	}

	// Validate Lag based on interface dest type
	destType := cato_models.SocketInterfaceDestTypeLan // Default: "LAN"
	if utils.HasValue(nativeRange.InterfaceDestType) {
		destType = cato_models.SocketInterfaceDestType(nativeRange.InterfaceDestType.ValueString())
	}
	if v.checkLag(&resp.Diagnostics, destType, nativeRange.LagMinLinks) != nil {
		return
	}

	// Validate DHCP settings
	if DHCPChecker.Check(ctx, &resp.Diagnostics, nativeRange.DhcpSettings) != nil {
		return
	}
}

func (v NativeRangeValidator) Description(_ context.Context) string {
	return "interface_index can only be specified for " +
		"SOCKET_X1500, SOCKET_X1600, SOCKET_X1600_LTE, or SOCKET_X1700; " +
		"local_ip must be within the native_network_range subnet"
}

func (v NativeRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// checkInterfaceIndex validates that non-default interface-index can be set only for specific connection types
// On error update diags and return error
func (v NativeRangeValidator) checkInterfaceIndex(diags *diag.Diagnostics, nativeRange tf.NativeRange,
	connectionType cato_models.SiteConnectionTypeEnum,
) error {
	if !utils.HasValue(nativeRange.InterfaceIndex) {
		return nil
	}

	// If the specified interface index is the same as the default, we are good.
	defaultIfaceIndex := tf.InterfaceByConnType[connectionType]
	ifaceIndex := cato_models.SocketInterfaceIDEnum(nativeRange.InterfaceIndex.ValueString())
	if ifaceIndex == defaultIfaceIndex {
		return nil
	}

	// Non-default interface index specified - it is allowed only for some con. types
	if slices.Contains(connTypesWithInterfaceIndex, connectionType) {
		return nil
	}

	diags.AddError("Invalid Configuration",
		fmt.Sprintf("interface_index can only be specified when connection_type is one of: %v",
			connTypesWithInterfaceIndex),
	)
	return ErrConfig
}

// checkLag checks lag settings on the native range
// - If interface_dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, lag_min_links must have a value
// - If lag_min_links has a value, interface_dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
// On error update diags and return error
func (v NativeRangeValidator) checkLag(diags *diag.Diagnostics, destType cato_models.SocketInterfaceDestType,
	lagMinLinks types.Int64,
) error {
	var lagMasterDestTypes = []cato_models.SocketInterfaceDestType{
		cato_models.SocketInterfaceDestTypeLanLagMaster,
		cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp,
	}

	hasLagMinLinks := utils.HasValue(lagMinLinks)
	destIsLagMaster := slices.Contains(lagMasterDestTypes, destType)

	// If interface_dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, lag_min_links must have a value
	if destIsLagMaster && !hasLagMinLinks {
		diags.AddError("Invalid LAG Configuration",
			fmt.Sprintf("When interface_dest_type is %s, lag_min_links must be specified.", destType))
		return ErrConfig
	}

	// If lag_min_links has a value, interface_dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
	if hasLagMinLinks && !destIsLagMaster {
		diags.AddError("Invalid LAG Configuration",
			fmt.Sprintf("lag_min_links can only be configured when interface_dest_type is LAN_LAG_MASTER or "+
				"LAN_LAG_MASTER_AND_VRRP, but interface_dest_type is %s.", destType),
		)
		return ErrConfig
	}

	return nil
}

func (v NativeRangeValidator) isUnknown(nativeRange tf.NativeRange) bool { //nolint:gocyclo
	return nativeRange.InterfaceIndex.IsUnknown() ||
		nativeRange.InterfaceID.IsUnknown() ||
		nativeRange.InterfaceName.IsUnknown() ||
		nativeRange.NativeNetworkLanInterfaceID.IsUnknown() ||
		nativeRange.NativeNetworkRange.IsUnknown() ||
		nativeRange.NativeNetworkRangeID.IsUnknown() ||
		nativeRange.RangeName.IsUnknown() ||
		nativeRange.RangeID.IsUnknown() ||
		nativeRange.LocalIP.IsUnknown() ||
		nativeRange.TranslatedSubnet.IsUnknown() ||
		nativeRange.Gateway.IsUnknown() ||
		nativeRange.RangeType.IsUnknown() ||
		nativeRange.DhcpSettings.IsUnknown() ||
		nativeRange.Vlan.IsUnknown() ||
		nativeRange.MdnsReflector.IsUnknown() ||
		nativeRange.LagMinLinks.IsUnknown() ||
		nativeRange.InterfaceDestType.IsUnknown()
}
