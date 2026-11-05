package cloudinterconnectphysicalconnection

import (
	"net/netip"
	"strconv"
	"strings"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const minVLAN int64 = 1
const maxVLAN int64 = 4094

func validID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		diags.AddError("Invalid connection ID", "A known non-empty connection ID is required.")
		return false
	}
	return true
}
func validateFields(plan resourceModel, diags *diag.Diagnostics) {
	for name, value := range map[string]types.String{
		"encapsulation_method": plan.EncapsulationMethod, "ha_role": plan.HARole, "service_provider_name": plan.ServiceProviderName,
		"subnet": plan.Subnet, "private_cato_ip": plan.PrivateCatoIP, "private_site_ip": plan.PrivateSiteIP,
		"upstream_bw_limit": plan.UpstreamBwLimit, "downstream_bw_limit": plan.DownstreamBwLimit,
	} {
		if value.IsNull() || value.IsUnknown() || strings.TrimSpace(value.ValueString()) == "" {
			diags.AddError("Invalid "+name, "A known non-empty value is required.")
		}
	}
	if !m.HaRole(plan.HARole.ValueString()).IsValid() {
		diags.AddError("Invalid HA role", "Use PRIMARY or SECONDARY.")
	}
	if !m.TaggingMethod(plan.EncapsulationMethod.ValueString()).IsValid() {
		diags.AddError("Invalid encapsulation method", "Use DOT1Q or QINQ.")
	}
	for _, value := range []types.String{plan.UpstreamBwLimit, plan.DownstreamBwLimit} {
		n, err := strconv.ParseInt(value.ValueString(), 10, 64)
		if err != nil || n < 0 || strconv.FormatInt(n, 10) != value.ValueString() {
			diags.AddError("Invalid bandwidth", "Use a canonical non-negative decimal integer.")
		}
	}
	validateSubnet(plan, diags)
	validateVLANs(plan, diags)
}
func validateSubnet(plan resourceModel, diags *diag.Diagnostics) {
	prefix, err := netip.ParsePrefix(plan.Subnet.ValueString())
	if err != nil || !prefix.Addr().Is4() || prefix != prefix.Masked() {
		diags.AddError("Invalid subnet", "Use a canonical IPv4 network prefix.")
		return
	}
	for _, value := range []types.String{plan.PrivateCatoIP, plan.PrivateSiteIP} {
		ip, err := netip.ParseAddr(value.ValueString())
		if err != nil || !ip.Is4() || !prefix.Contains(ip) {
			diags.AddError("Invalid tunnel address", "Both IPv4 addresses must belong to subnet.")
		}
	}
	if plan.PrivateCatoIP.Equal(plan.PrivateSiteIP) {
		diags.AddError("Duplicate tunnel address", "Cato and site addresses must differ.")
	}
}
func validVLAN(value types.Int64) bool {
	return !value.IsNull() && !value.IsUnknown() && value.ValueInt64() >= minVLAN && value.ValueInt64() <= maxVLAN
}
func validateVLANs(plan resourceModel, diags *diag.Diagnostics) {
	if plan.EncapsulationMethod.ValueString() == "DOT1Q" {
		if (!plan.Vlan.IsNull() && !validVLAN(plan.Vlan)) || !plan.QinqVlanConfiguration.IsNull() {
			diags.AddError("Invalid DOT1Q configuration", "Omit vlan for automatic allocation or set 1 through 4094; omit qinq_vlan_configuration.")
		}
		return
	}
	if plan.EncapsulationMethod.ValueString() != "QINQ" {
		return
	}
	if !plan.Vlan.IsNull() || plan.QinqVlanConfiguration.IsUnknown() {
		diags.AddError("Invalid QINQ configuration", "Omit vlan. Omit qinq_vlan_configuration for allocation or supply known valid tags.")
		return
	}
	if plan.QinqVlanConfiguration.IsNull() {
		return
	}
	attrs := plan.QinqVlanConfiguration.Attributes()
	if !validVLAN(attrs["c_vlan"].(types.Int64)) || !validVLAN(attrs["s_vlan"].(types.Int64)) {
		diags.AddError("Invalid QINQ VLAN", "Both c_vlan and s_vlan must be known integers from 1 through 4094.")
	}
}
