package networkconfig

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dnsSettingsAPI interface {
	GetPrimaryServer() *string
	GetSecondaryServer() *string
	GetSuffix() []string
}

// BuildDNSInput preserves explicit clear operations when converting Terraform DNS settings.
func BuildDNSInput(ctx context.Context,
	primary,
	secondary types.String,
	suffix types.List,
	diags *diag.Diagnostics) DNSSettingsInput {
	if primary.IsUnknown() || secondary.IsUnknown() || suffix.IsUnknown() {
		diags.AddError("Unknown DNS settings", "DNS server and suffix values must be known at apply.")
		return DNSSettingsInput{}
	}
	values := Strings(ctx, suffix, diags)
	if primary.ValueString() == "" && (secondary.ValueString() != "" || len(values) > 0) {
		diags.AddError("Invalid DNS settings", "A secondary DNS server or suffix requires a primary DNS server.")
	}
	return DNSSettingsInput{PrimaryServer: dnsServerPointer(primary),
		SecondaryServer: dnsServerPointer(secondary),
		Suffix:          values}
}

// ParseDNS normalizes API DNS settings into Terraform values.
func ParseDNS(ctx context.Context, item dnsSettingsAPI, primary, secondary *types.String, suffix *types.List, diags *diag.Diagnostics) {
	*primary = types.StringValue(stringValue(item.GetPrimaryServer()))
	*secondary = types.StringValue(stringValue(item.GetSecondaryServer()))
	*suffix = List(ctx, item.GetSuffix(), diags)
}
func dnsServerPointer(value types.String) *string {
	if value.ValueString() == "" {
		return nil
	}
	return value.ValueStringPointer()
}
func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
