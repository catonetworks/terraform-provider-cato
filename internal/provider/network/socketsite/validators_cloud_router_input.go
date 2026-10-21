package socketsite

import (
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *socketSiteResource) validateCloudRouterInput(plan *SocketSite, diags *diag.Diagnostics) {
	if !utils.HasValue(plan.IsCloudRouter) {
		return
	}

	if plan.ConnectionType.ValueString() != string(cato_models.SiteConnectionTypeEnumSocketGCP1500) {
		diags.AddAttributeError(
			path.Root("is_cloud_router"),
			"Unsupported Cloud Router configuration",
			"`is_cloud_router` is supported only for GCP vSocket HA sites.",
		)
	}
}
