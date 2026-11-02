package wifissid

import (
	"context"
	"encoding/json"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreate(ctx context.Context, p resourceModel, d *diag.Diagnostics) m.CreateWifiSsidInput {
	validateFields(p, d)
	validateSecurity(p, d)
	in := m.CreateWifiSsidInput{Name: p.Name.ValueString(),
		Site:               idname.PrepareIDName[m.SiteRefInput](ctx, p.Site, d),
		NetworkInterfaceID: p.NetworkInterfaceID.ValueString(),
		Band:               m.WifiRadioBand(p.Band.ValueString()),
		Category:           m.WifiSsidCategory(p.Category.ValueString()),
		Subnet:             p.Subnet.ValueString(),
		LocalIP:            p.LocalIP.ValueString(),
		MdnsEnabled:        p.MdnsEnabled.ValueBool(), MicrosegmentationEnabled: p.MicrosegmentationEnabled.ValueBool(),
		Enabled: boolPointer(p.Enabled, true), Visible: boolPointer(p.Visible, true), InternetOnly: boolPointer(p.InternetOnly, false)}
	if in.Site == nil {
		d.AddError("Missing site", "A known non-null site reference is required.")
	}
	if d.HasError() {
		return in
	}
	attrs := p.Security.Attributes()
	in.Security = &m.SecurityConfigInput{Mode: m.WifiSecurityMode(attrs["mode"].(types.String).ValueString()),
		AuthProtocol:        m.WifiAuthProtocol(attrs["auth_protocol"].(types.String).ValueString()),
		TrackAuthentication: boolPointer(attrs["track_authentication"].(types.Bool), false)}
	if in.Security.Mode == m.WifiSecurityModePsk {
		in.Security.Psk = &m.PskSecurityInput{Passkey: secretValue(p)}
	}
	if !p.Dhcp.IsNull() {
		in.Dhcp = &m.WifiSsidDhcpSettingsInput{}
		value := p.Dhcp.Attributes()["dhcp_subnet"].(types.String).ValueString()
		if value != "" {
			in.Dhcp.DhcpSubnet = &value
		}
	}
	// dhcpEnabled and vlanId are rejected by the current API, including explicit false/zero.
	return in
}
func boolPointer(value types.Bool, fallback bool) *bool {
	if value.IsNull() {
		return &fallback
	}
	return value.ValueBoolPointer()
}
func buildUpdate(in m.CreateWifiSsidInput, id string) m.UpdateWifiSsidInput {
	dhcp := in.Dhcp
	if dhcp == nil {
		dhcp = &m.WifiSsidDhcpSettingsInput{}
	}
	return m.UpdateWifiSsidInput{ID: id, Name: &in.Name, Band: &in.Band, Category: &in.Category, Subnet: &in.Subnet, LocalIP: &in.LocalIP,
		MdnsEnabled: &in.MdnsEnabled, MicrosegmentationEnabled: &in.MicrosegmentationEnabled, Enabled: in.Enabled, Visible: in.Visible,
		InternetOnly: in.InternetOnly, Security: in.Security, Dhcp: dhcp}
}
func identityState(p resourceModel, id string) resourceModel {
	p.ID = types.StringValue(id)
	attrs := p.Site.Attributes()
	for key, value := range attrs {
		if value.IsUnknown() {
			attrs[key] = types.StringValue("")
		}
	}
	p.Site = types.ObjectValueMust(idname.RefAttrTypes, attrs)
	return p
}
func secretValue(p resourceModel) string {
	if p.Security.IsNull() || p.Security.IsUnknown() {
		return ""
	}
	psk := p.Security.Attributes()["psk"].(types.Object)
	if psk.IsNull() || psk.IsUnknown() {
		return ""
	}
	return psk.Attributes()["passkey"].(types.String).ValueString()
}

// Mutation payloads have the same public fields but distinct generated Go types.
func mutationRecord[T any](value *T, d *diag.Diagnostics) *ssidRecord {
	if value == nil {
		d.AddError("Invalid API response", "API returned no SSID.")
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		d.AddError("Invalid API response", "Could not decode the SSID response.")
		return nil
	}
	item := new(ssidRecord)
	if json.Unmarshal(raw, item) != nil {
		d.AddError("Invalid API response", "Could not decode the SSID response.")
		return nil
	}
	return item
}
func parseSSID(ctx context.Context, item *ssidRecord, prior resourceModel, d *diag.Diagnostics) resourceModel {
	if item.GetBand() == nil || item.GetCategory() == nil {
		d.AddError("Invalid API response", "Missing Wi-Fi band/category.")
		return prior
	}
	p := resourceModel{ID: types.StringValue(item.GetID()),
		Site:                     idname.ParseIDName(ctx, *item.GetSite(), d),
		NetworkInterfaceID:       types.StringValue(item.GetNetworkInterfaceID()),
		Name:                     types.StringValue(item.GetName()),
		Band:                     types.StringValue(string(*item.GetBand())),
		Category:                 types.StringValue(string(*item.GetCategory())),
		Subnet:                   types.StringValue(item.GetSubnet()),
		LocalIP:                  types.StringValue(item.GetLocalIP()),
		Enabled:                  types.BoolValue(item.GetEnabled()),
		Visible:                  types.BoolValue(item.GetVisible()),
		DhcpEnabled:              types.BoolValue(item.GetDhcpEnabled()),
		InternetOnly:             types.BoolValue(item.GetInternetOnly()),
		MicrosegmentationEnabled: types.BoolValue(item.GetMicrosegmentationEnabled()),
		MdnsEnabled:              types.BoolValue(item.GetMdnsEnabled()),
		Dhcp:                     types.ObjectNull(dhcpAttrTypes)}
	value := ""
	if item.GetDhcp().GetDhcpSubnet() != nil {
		value = *item.GetDhcp().GetDhcpSubnet()
	}
	if value != "" || (!prior.Dhcp.IsNull() && !prior.Dhcp.IsUnknown()) {
		p.Dhcp = types.ObjectValueMust(dhcpAttrTypes, map[string]attr.Value{"dhcp_subnet": types.StringValue(value)})
	}
	p.Security = parseSecurity(item.GetSecurity(), prior, d)
	validateFields(p, d)
	return p
}
func parseSecurity(security *cato.SiteWifiSsid_Site_WifiSsid_Security, prior resourceModel, d *diag.Diagnostics) types.Object {
	psk := types.ObjectNull(pskAttrTypes)
	if security == nil ||
		security.GetMode() == nil ||
		security.GetAuthProtocol() == nil ||
		!security.GetMode().IsValid() ||
		!security.GetAuthProtocol().IsValid() {
		d.AddError("Invalid API response", "API returned unsupported Wi-Fi security settings.")
		return types.ObjectNull(securityAttrTypes)
	}
	if *security.GetMode() == m.WifiSecurityModePsk && !prior.Security.IsNull() && !prior.Security.IsUnknown() {
		priorAttrs := prior.Security.Attributes()
		if priorAttrs["mode"].(types.String).ValueString() == "PSK" {
			candidate := priorAttrs["psk"].(types.Object)
			if !candidate.IsUnknown() {
				psk = candidate
			}
		}
	}
	return types.ObjectValueMust(securityAttrTypes, map[string]attr.Value{"mode": types.StringValue(string(*security.GetMode())),
		"auth_protocol":        types.StringValue(string(*security.GetAuthProtocol())),
		"track_authentication": types.BoolValue(security.GetTrackAuthentication()),
		"psk":                  psk})
}
func matchesReference(ref *m.SiteRefInput, id, name string) bool {
	return ref != nil && ((ref.By == m.ObjectRefByID && ref.Input == id) || (ref.By == m.ObjectRefByName && ref.Input == name))
}

func removedIdentityMatches(item *ssidRecord, p resourceModel) bool {
	if item == nil || item.GetID() != p.ID.ValueString() || item.GetSite().GetID() == "" || item.GetNetworkInterfaceID() == "" {
		return false
	}
	if knownString(p.NetworkInterfaceID) && p.NetworkInterfaceID.ValueString() != item.GetNetworkInterfaceID() {
		return false
	}
	if !p.Site.IsNull() && !p.Site.IsUnknown() {
		attrs := p.Site.Attributes()
		id := attrs["id"].(types.String)
		name := attrs["name"].(types.String)
		if knownString(id) {
			return id.ValueString() == item.GetSite().GetID()
		}
		if knownString(name) {
			return name.ValueString() == item.GetSite().GetName()
		}
	}
	return true
}

func knownString(v types.String) bool { return !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" }
