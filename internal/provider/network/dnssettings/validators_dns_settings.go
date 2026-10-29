package dnssettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

func (r *settingsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config SettingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if !resp.Diagnostics.HasError() {
		nc.ValidateDNS(config.PrimaryServer, config.SecondaryServer, config.Suffix, &resp.Diagnostics)
	}
}
