package dns

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

type dnsSettingsAPI interface {
	GetPrimaryServer() *string
	GetSecondaryServer() *string
	GetSuffix() []string
}

func dnsInput(ctx context.Context,
	primary,
	secondary types.String,
	suffix types.List,
	diags *diag.Diagnostics) catoclient.DNSSettingsInput {
	if primary.IsUnknown() || secondary.IsUnknown() || suffix.IsUnknown() {
		diags.AddError("Unknown DNS settings", "DNS server and suffix values must be known at apply.")
		return catoclient.DNSSettingsInput{}
	}
	values := nc.Strings(ctx, suffix, diags)
	if primary.IsNull() && (!secondary.IsNull() || len(values) > 0) {
		diags.AddError("Invalid DNS settings", "A secondary DNS server or suffix requires a primary DNS server.")
	}
	return catoclient.DNSSettingsInput{PrimaryServer: primary.ValueStringPointer(),
		SecondaryServer: secondary.ValueStringPointer(),
		Suffix:          values}
}

func hydrateDNS(ctx context.Context, item dnsSettingsAPI, primary, secondary *types.String, suffix *types.List, diags *diag.Diagnostics) {
	*primary = types.StringPointerValue(item.GetPrimaryServer())
	*secondary = types.StringPointerValue(item.GetSecondaryServer())
	*suffix = nc.List(ctx, item.GetSuffix(), diags)
}
