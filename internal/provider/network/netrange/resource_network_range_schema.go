package netrange

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/socket"
)

const (
	networkRangeDescription = "The `cato_network_range` resource contains the configuration parameters necessary to " +
		"add a network range to a cato site. ([virtual socket in AWS/Azure, or physical socket]" +
		"(https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). " +
		"Documentation for the underlying API used in this resource can be found at [mutation.addNetworkRange()]" +
		"(https://api.catonetworks.com/documentation/#mutation-site.addNetworkRange)."
	networkRangeMDNSReflectorDescription = "Site native range mDNS reflector. When enabled, the Socket functions as an " +
		"mDNS gateway, it relays mDNS requests and response between all enabled subnets."
	networkRangeDHCPMicrosegmentationDescription = "DHCP Microsegmentation. When enabled, the DHCP server will allocate " +
		"/32 subnet mask. Make sure to enable the proper Firewall rules and enable it with caution, as it is not supported " +
		"on all operating systems; monitor the network closely after activation. This setting can only be configured when " +
		"dhcp_type is set to DHCP_RANGE."
	networkRangeDHCPDisabledError = "When dhcp_type is DHCP_DISABLED, dhcp_ip_range, dhcp_relay_group_id, and " +
		"dhcp_relay_group_name must be null, unset, or empty strings."
	networkRangeDHCPRangeError = "When dhcp_type is DHCP_RANGE, dhcp_ip_range must be provided (not null, unset, or " +
		"empty string), and dhcp_relay_group_id and dhcp_relay_group_name must be null, unset, or empty strings."
)

func (r *networkRangeResource) resourceSchema() schema.Schema {
	return schema.Schema{
		Description: networkRangeDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network Range ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway": schema.StringAttribute{
				Description: "Network range gateway (Only releveant for Routed range_type)",
				Optional:    true,
				Computed:    true,
			},
			"interface_id": schema.StringAttribute{
				Description: "Network Interface ID",
				Required:    false,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"interface_index": schema.StringAttribute{
				Description:   "Network Interface Index",
				Required:      false,
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{socket.InterfaceIndexValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"internet_only": schema.BoolAttribute{
				Description:   "Internet only network range (Only releveant for Routed range_type)",
				Computed:      true,
				Optional:      true,
				Default:       booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"local_ip": schema.StringAttribute{
				Description: "Network range local ip",
				Optional:    true,
				Computed:    true,
			},
			"mdns_reflector": schema.BoolAttribute{
				Description:   networkRangeMDNSReflectorDescription,
				Optional:      true,
				Computed:      true,
				Default:       booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Network range name",
				Required:    true,
			},
			"range_type": schema.StringAttribute{
				Description:   "Network range type (https://api.catonetworks.com/documentation/#definition-SubnetType)",
				Required:      true,
				Validators:    []validator.String{SubnetTypeValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"site_id": schema.StringAttribute{
				Description:   "Site ID",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"subnet": schema.StringAttribute{
				Description: "Network range (CIDR)",
				Required:    true,
			},
			"translated_subnet": schema.StringAttribute{
				Description:   "Network range translated native IP range (CIDR)",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dhcp_settings": dhcp.SchemaDhcpSettings(true),
			"vlan": schema.Int64Attribute{
				Description: "Network range VLAN ID (Only releveant for VLAN range_type)",
				Optional:    true,
			},
		},
	}
}
