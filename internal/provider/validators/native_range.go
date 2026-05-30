package validators

import (
	"context"
	"errors"
	"fmt"
	"net"
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

var errConfig = errors.New("configuration error")

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
	localIP := nativeRange.LocalIP.ValueString()
	subnet := nativeRange.NativeNetworkRange.ValueString()
	if v.checkLocalIP(&resp.Diagnostics, localIP, subnet) != nil {
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
	if v.checkDHCP(ctx, &resp.Diagnostics, nativeRange.DhcpSettings) != nil {
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
	return errConfig
}

// checkLocalIP checks if local_ip is within native_network_range
// On error update diags and return error
func (v NativeRangeValidator) checkLocalIP(diags *diag.Diagnostics, localIP, subnet string) error {
	// Parse the local IP
	ip := net.ParseIP(localIP)
	if ip == nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("local_ip '%s' is not a valid IP address", localIP),
		)
		return errConfig
	}

	// Parse the subnet CIDR
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("native_network_range '%s' is not a valid CIDR notation", subnet),
		)
		return errConfig
	}

	// Check if the IP is within the subnet
	if !ipNet.Contains(ip) {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("Local IP must be within the Native Range IP. "+
				"local_ip '%s' is not within native_network_range '%s'", localIP, subnet),
		)
		return errConfig
	}

	return nil
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
		return errConfig
	}

	// If lag_min_links has a value, interface_dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
	if hasLagMinLinks && !destIsLagMaster {
		diags.AddError("Invalid LAG Configuration",
			fmt.Sprintf("lag_min_links can only be configured when interface_dest_type is LAN_LAG_MASTER or "+
				"LAN_LAG_MASTER_AND_VRRP, but interface_dest_type is %s.", destType),
		)
		return errConfig
	}

	return nil
}

func (v NativeRangeValidator) checkDHCP(ctx context.Context, diags *diag.Diagnostics, dhcp types.Object) error {
	var dhcpSettings *tf.DhcpSettings

	if !utils.HasValue(dhcp) {
		return nil
	}

	// get dhcp config
	if utils.CheckErr(diags, dhcp.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})) {
		return errConfig
	}
	if dhcpSettings == nil {
		return nil
	}

	// validate dhcp-relay group setting
	if relayErr := v.checkDHCPRelay(diags, *dhcpSettings); relayErr != nil {
		return relayErr
	}

	// validate dhcp ip range setting
	if rangeErr := v.checkDHCPRange(diags, *dhcpSettings); rangeErr != nil {
		return rangeErr
	}

	return nil
}

// checkDHCPRelay validates the consistency of DHCP relay settings
// if DHCP type is DHCP_RELAY, exactly one of relay_group_name or relay_group_id must be set,
// otherwise, neither can be set.
func (v NativeRangeValidator) checkDHCPRelay(diags *diag.Diagnostics, dhcpSettings tf.DhcpSettings) error {
	relayGroupNameSet := utils.HasValue(dhcpSettings.RelayGroupName)
	relayGroupIDSet := utils.HasValue(dhcpSettings.RelayGroupID)
	dhcpType := cato_models.DhcpType(dhcpSettings.DhcpType.ValueString())

	// If DHCP type is not DHCP_RELAY, relay group name/id must not be configured
	if dhcpType != cato_models.DhcpTypeDhcpRelay {
		if relayGroupIDSet || relayGroupNameSet {
			diags.AddError("Invalid DHCP Configuration",
				fmt.Sprintf("relay_group_id or relay_group_name can only be configured when DHCP type is 'DHCP_RELAY' (have %q)", dhcpType))
			return errConfig
		}
		return nil
	}

	// If DHCP type is DHCP_RELAY, exactly on of relay_group_name relay_group_id must not be configured
	if !relayGroupNameSet && !relayGroupIDSet {
		diags.AddError("Invalid DHCP Configuration",
			"either relay_group_id or relay_group_name must be configured when DHCP type is 'DHCP_RELAY'")
		return errConfig
	}
	if relayGroupNameSet && relayGroupIDSet {
		diags.AddError("Invalid DHCP Configuration",
			"only one of relay_group_id or relay_group_name can be configured when DHCP type is 'DHCP_RELAY'")
		return errConfig
	}
	return nil
}

// checkDHCPRange validates the consistency of DHCP range settings
// - if DHCP type is DHCP_RANGE, ip_range must be set, otherwise it must not be set.
// - if DHCP type is not  DHCP_RANGE, dhcp_microsegmentation must not be set to true
func (v NativeRangeValidator) checkDHCPRange(diags *diag.Diagnostics, dhcpSettings tf.DhcpSettings) error {
	dhcpType := cato_models.DhcpType(dhcpSettings.DhcpType.ValueString())
	ipRangeSet := utils.HasValue(dhcpSettings.IPRange)

	if dhcpType == cato_models.DhcpTypeDhcpRange {
		if !ipRangeSet {
			diags.AddError("Invalid DHCP Configuration",
				"ip_range must be configured when DHCP type is 'DHCP_RANGE'")
			return errConfig
		}
		return nil
	}

	// dhcpType != DHCP_RANGE
	if ipRangeSet {
		diags.AddError("Invalid DHCP Configuration",
			"ip_range can only be configured when DHCP type is 'DHCP_RANGE'")
		return errConfig
	}

	if utils.HasValue(dhcpSettings.DhcpMicrosegmentation) && dhcpSettings.DhcpMicrosegmentation.ValueBool() {
		diags.AddError("Invalid DHCP Configuration",
			"dhcp_microsegmentation can only be seto to true, when DHCP type is 'DHCP_RANGE'")
		return errConfig
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
