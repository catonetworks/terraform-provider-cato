package wifissid

import (
	"net/netip"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var interfacePattern = regexp.MustCompile(`^WBR[1-4]$`)

var namePattern = regexp.MustCompile(`^[A-Za-z0-9 _\-.@#()&+]+$`)

func validID(id types.String, d *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		d.AddError("Invalid SSID ID", "A known non-empty SSID ID is required.")
		return false
	}
	return true
}
func validateFields(p resourceModel, d *diag.Diagnostics) {
	for key, v := range map[string]types.String{"name": p.Name,
		"band":                 p.Band,
		"category":             p.Category,
		"local_ip":             p.LocalIP,
		"subnet":               p.Subnet,
		"network_interface_id": p.NetworkInterfaceID} {
		if v.IsNull() || v.IsUnknown() || strings.TrimSpace(v.ValueString()) == "" {
			d.AddError("Invalid "+key, "A known non-empty value is required.")
		}
	}
	if !namePattern.MatchString(p.Name.ValueString()) || len(p.Name.ValueString()) > 32 {
		d.AddError("Invalid SSID name", "Use 1 to 32 supported ASCII letters, digits, spaces or punctuation: _-.@#()&+.")
	}
	if !m.WifiRadioBand(p.Band.ValueString()).IsValid() {
		d.AddError("Invalid band", "Use a supported Wi-Fi radio band.")
	}
	if p.Category.ValueString() != "GUEST" {
		d.AddError("Unsupported category", "The API currently supports GUEST SSIDs only.")
	}
	if !interfacePattern.MatchString(p.NetworkInterfaceID.ValueString()) {
		d.AddError("Invalid interface", "Use WBR1 through WBR4.")
	}
	validateFlags(p, d)
	validateNetwork(p, d)
	validateDHCP(p, d)
}
func validateFlags(p resourceModel, d *diag.Diagnostics) {
	for key, v := range map[string]types.Bool{"enabled": p.Enabled,
		"visible":       p.Visible,
		"dhcp_enabled":  p.DhcpEnabled,
		"internet_only": p.InternetOnly} {
		if v.IsUnknown() {
			d.AddError("Unknown "+key, "A known boolean is required at apply time.")
		}
	}
	for key, v := range map[string]types.Bool{"mdns_enabled": p.MdnsEnabled, "microsegmentation_enabled": p.MicrosegmentationEnabled} {
		if v.IsNull() || v.IsUnknown() {
			d.AddError("Invalid "+key, "A known boolean is required.")
		}
	}
	if p.DhcpEnabled.ValueBool() {
		d.AddError("Unsupported dhcp_enabled", "The API deprecated this flag; use dhcp instead and leave dhcp_enabled false.")
	}
}
func validateNetwork(p resourceModel, d *diag.Diagnostics) {
	prefix, err := netip.ParsePrefix(p.Subnet.ValueString())
	if err != nil || !prefix.Addr().Is4() || prefix != prefix.Masked() {
		d.AddError("Invalid subnet", "Use a canonical IPv4 network prefix.")
	} else {
		ip, e := netip.ParseAddr(p.LocalIP.ValueString())
		if e != nil || !ip.Is4() || !prefix.Contains(ip) {
			d.AddError("Invalid local_ip", "Use an IPv4 address within subnet.")
		}
	}
}
func validateDHCP(p resourceModel, d *diag.Diagnostics) {
	if p.Dhcp.IsUnknown() {
		d.AddError("Unknown DHCP settings", "DHCP settings must be known at apply time.")
		return
	}
	if p.Dhcp.IsNull() {
		return
	}
	value := p.Dhcp.Attributes()["dhcp_subnet"].(types.String)
	if value.IsUnknown() {
		d.AddError("Unknown DHCP subnet", "DHCP subnet must be known at apply time.")
		return
	}
	if value.ValueString() == "" {
		return
	}
	network, e1 := netip.ParsePrefix(p.Subnet.ValueString())
	dhcp, e2 := netip.ParsePrefix(value.ValueString())
	if e1 != nil || e2 != nil || !dhcp.Addr().Is4() || dhcp != dhcp.Masked() || dhcp.Bits() < network.Bits() ||
		!network.Contains(dhcp.Addr()) {
		d.AddError("Invalid DHCP subnet", "Use a canonical IPv4 prefix within subnet.")
	}
}

func validateSecurity(p resourceModel, d *diag.Diagnostics) {
	if p.Security.IsNull() || p.Security.IsUnknown() {
		d.AddError("Invalid security", "A known non-null security object is required.")
		return
	}
	attrs := p.Security.Attributes()
	mode := attrs["mode"].(types.String)
	protocol := attrs["auth_protocol"].(types.String)
	if mode.IsNull() || mode.IsUnknown() || !m.WifiSecurityMode(mode.ValueString()).IsValid() {
		d.AddError("Invalid security mode", "Use OPEN or PSK.")
	}
	if protocol.IsNull() || protocol.IsUnknown() || !m.WifiAuthProtocol(protocol.ValueString()).IsValid() {
		d.AddError("Invalid authentication protocol", "Use a supported Wi-Fi authentication protocol.")
	}
	if attrs["track_authentication"].IsUnknown() {
		d.AddError("Unknown authentication tracking", "A known boolean is required.")
	}
	validatePsk(mode, attrs["psk"].(types.Object), d)
}
func validatePsk(mode types.String, psk types.Object, d *diag.Diagnostics) {
	if mode.ValueString() == "OPEN" && !psk.IsNull() {
		d.AddError("Unexpected PSK", "Omit psk when using OPEN security.")
	}
	if mode.ValueString() != "PSK" {
		return
	}
	if psk.IsNull() || psk.IsUnknown() {
		d.AddError("Missing PSK", "Configure a known PSK for PSK security. Imported resources cannot recover the passkey.")
		return
	}
	passkey := psk.Attributes()["passkey"].(types.String)
	value := passkey.ValueString()
	length := len(utf16.Encode([]rune(value)))
	if passkey.IsNull() || passkey.IsUnknown() || length < 8 || length > 63 || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		d.AddError("Invalid passkey", "Use a known passkey of 8 to 63 UTF-16 code units without whitespace.")
	}
}
