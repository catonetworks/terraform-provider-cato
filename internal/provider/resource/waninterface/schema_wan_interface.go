//nolint:lll
package waninterface

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *wanInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_wan_interface` resource contains the configuration parameters necessary to add a wan interface to a socket. ([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.updateSocketInterface()](https://api.catonetworks.com/documentation/#mutation-site.updateSocketInterface).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The WAN interface ID, which is a combination of the site ID and the interface ID (e.g., `site_id:interface_id`, 12345:INT_1). This is used to identify the WAN interface resource.",
				Required:    false,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"interface_id": schema.StringAttribute{
				Description: "The interface ID, which is a unique identifier for the WAN interface (e.g., `INT_1`, `INT_2`, etc.). This is used to identify the specific WAN interface resource.",
				Required:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "WAN interface name",
				Required:    true,
			},
			"upstream_bandwidth": schema.Int64Attribute{
				Description: "WAN interface upstream bandwidth",
				Required:    true,
			},
			"downstream_bandwidth": schema.Int64Attribute{
				Description: "WAN interface downstream bandwidth",
				Required:    true,
			},
			"role": schema.StringAttribute{
				Description: "WAN interface role (https://api.catonetworks.com/documentation/#definition-SocketInterfaceRole)",
				Required:    true,
			},
			"precedence": schema.StringAttribute{
				Description: "WAN interface precedence (https://api.catonetworks.com/documentation/#definition-SocketInterfacePrecedenceEnum)",
				Required:    true,
			},
			// "off_cloud": schema.SingleNestedAttribute{
			// 	Description: "Off Cloud configuration (https://support.catonetworks.com/hc/en-us/articles/4413265642257-Routing-Traffic-to-an-Off-Cloud-Link#heading-1)",
			// 	Required:    true,
			// 	Optional:    false,
			// 	Attributes: map[string]schema.Attribute{
			// 		"enabled": schema.BoolAttribute{
			// 			Description: "Attribute to define off cloud status (enabled or disabled)",
			// 			Required:    true,
			// 			Optional:    false,
			// 		},
			// 		"public_ip": schema.StringAttribute{
			// 			Required:    false,
			// 			Optional:    true,
			// 		},
			// 		"public_port": schema.StringAttribute{
			// 			Required:    false,
			// 			Optional:    true,
			// 		},
			// 	},
			// },
		},
	}
}
