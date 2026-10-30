package customapplication

import (
	"net/netip"
	"strconv"
	"strings"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validID(value types.String, d *diag.Diagnostics) bool {
	if !knownString(value) {
		d.AddError("Invalid application ID", "A known non-empty application ID is required.")
		return false
	}
	return true
}
func knownString(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown() && strings.TrimSpace(value.ValueString()) != ""
}
func validateRoot(p resourceModel, d *diag.Diagnostics) {
	if !knownString(p.Name) || strings.TrimSpace(p.Name.ValueString()) != p.Name.ValueString() {
		d.AddError("Invalid name", "A known non-empty name without surrounding whitespace is required.")
	}
	if p.Description.IsUnknown() {
		d.AddError("Unknown description", "Description must be known at apply time.")
	}
	if p.Category.IsUnknown() {
		d.AddError("Unknown categories", "Categories must be known at apply time.")
	}
	for _, value := range p.Category.Elements() {
		if value.IsNull() || value.IsUnknown() {
			d.AddError("Invalid category", "Every category requires a known non-null ID/name reference.")
		}
	}
	if p.Criteria.IsNull() || p.Criteria.IsUnknown() || len(p.Criteria.Elements()) == 0 {
		d.AddError("Invalid criteria", "At least one known non-null criterion is required.")
	}
}
func validateKnown(value attr.Value, d *diag.Diagnostics) {
	if value.IsUnknown() {
		d.AddError("Unknown criterion value", "Criteria values must be known at apply time.")
		return
	}
	if value.IsNull() {
		return
	}
	switch v := value.(type) {
	case types.Object:
		for _, item := range v.Attributes() {
			validateKnown(item, d)
		}
	case types.Set:
		for _, item := range v.Elements() {
			if item.IsNull() {
				d.AddError("Null collection member", "Criterion collections cannot contain null members.")
			} else {
				validateKnown(item, d)
			}
		}
	}
}
func validateCriterion(in *m.CustomApplicationCriteriaInput, d *diag.Diagnostics) {
	if !in.Protocol.IsValid() {
		d.AddError("Invalid protocol", "Use a supported IP protocol.")
	}
	for _, port := range in.Port {
		validatePort(string(port), d)
	}
	for _, pair := range in.PortRange {
		from := validatePort(string(pair.From), d)
		to := validatePort(string(pair.To), d)
		if from > to {
			d.AddError("Invalid port range", "Port range from must not exceed to.")
		}
	}
	if in.Destination != nil {
		validateDestination(in.Destination, d)
	}
	if in.Protocol == m.IPProtocolAny && len(in.Port) == 0 && len(in.PortRange) == 0 && !hasDestination(in.Destination) {
		d.AddError("Empty criterion", "Each criterion must specify a protocol, ports, ranges or destination criteria.")
	}
}
func validatePort(value string, d *diag.Diagnostics) int64 {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n < 0 || n > 65535 || strconv.FormatInt(n, 10) != value {
		d.AddError("Invalid port", "Use a canonical decimal port from 0 through 65535.")
	}
	return n
}
func hasDestination(in *m.CustomApplicationDestinationInput) bool {
	if in == nil {
		return false
	}
	if len(in.Domain)+len(in.Fqdn) > 0 {
		return true
	}
	return in.DestinationIP != nil && (len(in.DestinationIP.IP)+len(in.DestinationIP.Subnet)+len(in.DestinationIP.IPRange) > 0)
}
func validateDestination(in *m.CustomApplicationDestinationInput, d *diag.Diagnostics) {
	for _, value := range append(append([]string{}, in.Domain...), in.Fqdn...) {
		if strings.TrimSpace(value) == "" || strings.ContainsAny(value, " \t\r\n") {
			d.AddError("Invalid destination name", "Domain/FQDN entries must be non-empty without whitespace.")
		}
	}
	if in.DestinationIP == nil {
		return
	}
	validateIPDestination(in.DestinationIP, d)
}
func validateIPDestination(ip *m.CustomApplicationIPInput, d *diag.Diagnostics) {
	if len(ip.IP)+len(ip.Subnet)+len(ip.IPRange) == 0 {
		d.AddError("Empty destination IP",
			"Omit destination_ip when it has no addresses, subnets or ranges.")
		return
	}

	for _, value := range ip.IP {
		parsed, err := netip.ParseAddr(value)
		if err != nil || parsed.String() != value {
			d.AddError("Invalid IP address", "Use a canonical IP address.")
		}
	}
	for _, value := range ip.Subnet {
		parsed, err := netip.ParsePrefix(value)
		if err != nil || parsed != parsed.Masked() || parsed.String() != value {
			d.AddError("Invalid subnet", "Use a canonical network prefix.")
		}
	}
	for _, pair := range ip.IPRange {
		from, e1 := netip.ParseAddr(pair.From)
		to, e2 := netip.ParseAddr(pair.To)
		if e1 != nil || e2 != nil || from.Is4() != to.Is4() || from.Compare(to) > 0 {
			d.AddError("Invalid IP range", "Use ordered IP addresses from the same family.")
		}
	}
}
