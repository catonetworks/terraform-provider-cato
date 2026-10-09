package laninterface

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//nolint:lll
func (r *lanInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_lan_interface` resource contains the configuration parameters necessary to add a lan interface to a socket. ([physical socket physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.updateSocketInterface()](https://api.catonetworks.com/documentation/#mutation-site.updateSocketInterface).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network Interface ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "LAN interface name",
				Required:    true,
			},
			"interface_id": schema.StringAttribute{
				Description: "SocketInterface available ids, INT_# stands for 1,2,3...12 supported ids (https://api.catonetworks.com/documentation/#definition-SocketInterfaceIDEnum)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"INT_1", "INT_2", "INT_3", "INT_4", "INT_5", "INT_6", "INT_7", "INT_8", "INT_9", "INT_10", "INT_11", "INT_12", "LAN1", "LAN2",
					),
				},
			},
			"dest_type": schema.StringAttribute{
				Description: "SocketInterface destination type (https://api.catonetworks.com/documentation/#definition-SocketInterfaceDestType)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"INTERFACE_DISABLED", "LAN", "LAN_AND_HA", lanLagMasterDestType, lanLagMasterAndVrrpDestType, "VRRP", "VRRP_AND_LAN",
					),
				},
			},
			"local_ip": schema.StringAttribute{
				Description: "Local IP address of the LAN interface",
				Required:    false,
				Optional:    true,
			},
			"lag_min_links": schema.Int64Attribute{
				Description: "Number of interfaces to include in the link aggregagtion, only relevant for LAN_LAG_MASTER, LAN_LAG_MASTER_AND_VRRP",
				Required:    false,
				Optional:    true,
			},
			"subnet": schema.StringAttribute{
				Description: "Subnet of the LAN interface in CIDR notation",
				Required:    false,
				Optional:    true,
			},
			"translated_subnet": schema.StringAttribute{
				Description: "Translated NAT subnet configuration",
				Required:    false,
				Optional:    true,
			},
			"vrrp_type": schema.StringAttribute{
				Description: "VRRP Type (https://api.catonetworks.com/documentation/#definition-VrrpType)",
				Required:    false,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("DIRECT_LINK", "VIA_SWITCH"),
				},
			},
		},
	}
}
