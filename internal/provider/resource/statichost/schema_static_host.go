package statichost

import (
	"context"

	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *staticHostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_static_host` resource contains the configuration parameters necessary to add a static host. " +
			"Documentation for the underlying API used in this resource can be found at " +
			"[mutation.addStaticHost()](https://api.catonetworks.com/documentation/#mutation-site.addStaticHost).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Host ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID (Host's parent). Exactly one of `site_id` or `site_name` must be set. " +
					"When `site_name` is used, `site_id` is resolved via entityLookup and written to state (Computed).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"site_name": schema.StringAttribute{
				Description: "Site name (Host's parent), an alternative to `site_id`. Resolved to the site's ID " +
					"via entityLookup at apply (site names are unique). Exactly one of `site_id` or `site_name` must be set.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("site_id"), path.MatchRoot("site_name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "Host name",
				Required:    true,
			},
			"ip": schema.StringAttribute{
				Description: "Host IP address",
				Required:    true,
			},
			"mac_address": schema.StringAttribute{
				Description: "Host MAC address (for DHCP reservervation)",
				Optional:    true,
			},
		},
	}
}
