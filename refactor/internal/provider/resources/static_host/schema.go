package static_host

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	ID         types.String `tfsdk:"id"`
	SiteID     types.String `tfsdk:"site_id"`
	Name       types.String `tfsdk:"name"`
	IP         types.String `tfsdk:"ip"`
	MacAddress types.String `tfsdk:"mac_address"`
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creation-only POC for a static host on an existing site.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, Description: "Created static-host ID."},
			"site_id":     schema.StringAttribute{Required: true, Description: "Existing site ID."},
			"name":        schema.StringAttribute{Required: true, Description: "Static-host name."},
			"ip":          schema.StringAttribute{Required: true, Description: "Static-host IP address."},
			"mac_address": schema.StringAttribute{Optional: true, Description: "Optional MAC address."},
		},
	}
}
