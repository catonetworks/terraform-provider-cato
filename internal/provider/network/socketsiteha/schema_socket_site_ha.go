package socketsiteha

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *socketSiteHAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configures Cloud Router mode on an existing Cato socket site. " +
			"Use this resource after the secondary vSocket has been registered.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Site ID used as the resource ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Cato socket site ID. The secondary vSocket must already be registered.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_cloud_router": schema.BoolAttribute{
				Description: "Enable or disable Cloud Router mode for the HA socket site.",
				Required:    true,
			},
		},
	}
}
