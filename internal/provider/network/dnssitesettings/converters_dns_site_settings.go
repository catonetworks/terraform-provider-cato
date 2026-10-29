package dnssitesettings

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

func buildSiteSettings(ctx context.Context, plan SiteSettingsModel, diags *diag.Diagnostics) (*models.SiteRefInput, nc.DNSSettingsInput) {
	ref := idname.PrepareIDName[models.SiteRefInput](ctx, plan.Site, diags)
	input := nc.BuildDNSInput(ctx, plan.PrimaryServer, plan.SecondaryServer, plan.Suffix, diags)
	return ref, input
}
