package dns

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateDNS(primary, secondary types.String, suffix types.List, diags *diag.Diagnostics) {
	if primary.IsUnknown() || !primary.IsNull() {
		return
	}
	if !secondary.IsNull() && !secondary.IsUnknown() {
		diags.AddAttributeError(path.Root("secondary_server"), "Invalid DNS settings", "A secondary DNS server requires a primary DNS server.")
	}
	if !suffix.IsNull() && !suffix.IsUnknown() && len(suffix.Elements()) > 0 {
		diags.AddAttributeError(path.Root("suffix"), "Invalid DNS settings", "DNS search suffixes require a primary DNS server.")
	}
}
func (r *settingsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config SettingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if !resp.Diagnostics.HasError() {
		validateDNS(config.PrimaryServer, config.SecondaryServer, config.Suffix, &resp.Diagnostics)
	}
}
func (r *siteSettingsResource) ValidateConfig(ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse) {
	var config SiteSettingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if !resp.Diagnostics.HasError() {
		validateDNS(config.PrimaryServer, config.SecondaryServer, config.Suffix, &resp.Diagnostics)
	}
}
