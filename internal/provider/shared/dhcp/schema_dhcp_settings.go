package dhcp

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// SchemaDhcpSettings returns the schema for the DHCP settings of a network range resource.
func SchemaDhcpSettings(isRangeResource bool) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:   "Site native range DHCP settings (Only relevant for NATIVE and VLAN range_type)",
		Optional:      true,
		Computed:      true,
		PlanModifiers: []planmodifier.Object{SettingsModifier(isRangeResource)},
		Attributes: map[string]schema.Attribute{
			"dhcp_type": schema.StringAttribute{
				Description: "Network range dhcp type (https://api.catonetworks.com/documentation/#definition-DhcpType)",
				Required:    true,
				Validators:  []validator.String{TypeValidator{}},
			},
			"ip_range": schema.StringAttribute{
				Description: "Network range dhcp range (format \"192.168.1.10-192.168.1.20\")",
				Optional:    true,
			},
			"relay_group_id": schema.StringAttribute{
				Description: "Network range dhcp relay group id",
				Optional:    true,
				Computed:    true,
			},
			"relay_group_name": schema.StringAttribute{
				Description: "Network range dhcp relay group name",
				Optional:    true,
				Computed:    true,
			},
			"dhcp_microsegmentation": schema.BoolAttribute{
				Description: "DHCP Microsegmentation. When enabled, the DHCP server will allocate /32 subnet mask. " +
					"Make sure to enable the proper Firewall rules and enable it with caution, " +
					"as it is not supported on all operating systems; monitor the network closely after activation. " +
					"This setting can only be configured when dhcp_type is set to DHCP_RANGE.",
				Optional: true,
				Computed: true,
			},
		},
	}
}
