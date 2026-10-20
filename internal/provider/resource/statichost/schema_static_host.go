package statichost

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
				Description: "Site ID (Host's parent)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
