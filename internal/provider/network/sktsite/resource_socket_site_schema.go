package sktsite

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/socket"
)

func (r *socketSiteResource) resourceSchema() schema.Schema {
	return schema.Schema{
		Description: "The `cato_socket_site` resource contains the configuration parameters necessary to add a socket site " +
			"to the Cato cloud ([virtual socket in AWS/Azure, or physical socket]" +
			"(https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). " +
			"Documentation for the underlying API used in this resource can be found at " +
			"[mutation.addSocketSite()](https://api.catonetworks.com/documentation/#mutation-site.addSocketSite). \n\n" +
			"**Note**: For AWS deployments, please accept the [EULA for the Cato Networks AWS Marketplace product]" +
			"(https://aws.amazon.com/marketplace/pp?sku=dvfhly9fuuu67tw59c7lt5t3c).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Site ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Site name",
				Required:    true,
			},
			"connection_type": schema.StringAttribute{
				Description:   "Connection type for the site (SOCKET_X1500, SOCKET_AWS1500, SOCKET_AZ1500, ...)",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{SiteConnectionTypeValidator{}},
			},
			"site_type": schema.StringAttribute{
				Description: "Site type (https://api.catonetworks.com/documentation/#definition-SiteType)",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Site description",
				Optional:    true,
			},
			"native_range":  r.schemaNativeRange(),
			"site_location": r.schemaSiteLocation(),
			"sockets":       r.schemaSockets(),
		},
	}
}

func (r *socketSiteResource) schemaNativeRange() schema.SingleNestedAttribute { //nolint:funlen
	return schema.SingleNestedAttribute{
		Description: "Site lan native range settings",
		Required:    true,
		Validators:  []validator.Object{GetNativeRangeValidator()},
		Attributes: map[string]schema.Attribute{
			"interface_index": schema.StringAttribute{
				Description: "LAN native range interface index, default is LAN1 for SOCKET_X1500 models, " +
					"INT_5 for SOCKET_X1600 and SOCKET_X1600_LTE, and INT_3 for SOCKET_X1700 models",
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{socket.InterfaceIndexValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"interface_id": schema.StringAttribute{
				Description:   "LAN native range interface id",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"native_network_range": schema.StringAttribute{
				Description: "Site native IP range (CIDR)",
				Required:    true,
			},
			"native_network_lan_interface_id": schema.StringAttribute{
				Description:   "ID of native range LAN interface (for additional network range update purposes)",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"native_network_range_id": schema.StringAttribute{
				Description:   "Site native IP range ID (for update purpose)",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"interface_name": schema.StringAttribute{
				Description:   "LAN native range interface name (e.g., 'LAN 01')",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"range_name": schema.StringAttribute{
				Description:   "Native range name (typically 'Native Range')",
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"range_id": schema.StringAttribute{
				Description:   "Native range ID (base64 encoded identifier)",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"gateway": schema.StringAttribute{
				Description:   "Gateway IP address for the native range",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"vlan": schema.Int64Attribute{
				Description: "VLAN ID for the site native range (optional)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"mdns_reflector": schema.BoolAttribute{
				Description: "Site native range mDNS reflector. When enabled, the Socket functions as an mDNS gateway, " +
					"it relays mDNS requests and response between all enabled subnets.",
				Optional:      true,
				Computed:      true,
				Default:       booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"local_ip": schema.StringAttribute{
				Description: "Site native range local ip",
				Required:    true,
			},
			"primary_management_ip": schema.StringAttribute{
				Description:   "Site native range primary management IP",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"secondary_management_ip": schema.StringAttribute{
				Description:   "Site native range secondary management IP",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"range_type": schema.StringAttribute{
				Description:   "NATIVE",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"translated_subnet": schema.StringAttribute{
				Description:   "Site translated native IP range (CIDR)",
				Computed:      true,
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"lag_min_links": schema.Int64Attribute{
				Description: "Number of interfaces to include in the link aggregation, " +
					"only relevant for LAN_LAG_MASTER and LAN_LAG_MASTER_AND_VRRP interface destination types",
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(1)},
			},
			"interface_dest_type": schema.StringAttribute{
				Description: "Socket interface destination type for the native interface, " +
					"example values: LAN, LAN_LAG_MASTER, LAN_LAG_MASTER_AND_VRRP, LAN_AND_HA, VRRP, VRRP_AND_LAN",
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{SocketInterfaceDestTypeValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dhcp_settings": dhcp.SchemaDhcpSettings(false),
		},
	}
}

func (r *socketSiteResource) schemaSiteLocation() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Site location",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"country_code": schema.StringAttribute{
				Description: "Site country code (can be retrieve from entityLookup)",
				Required:    true,
			},
			"state_code": schema.StringAttribute{
				Description: "Optionnal site state code(can be retrieve from entityLookup)",
				Optional:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "Site timezone (can be retrieve from entityLookup)",
				Required:    true,
			},
			"city": schema.StringAttribute{
				Description:   "Optionnal city",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"address": schema.StringAttribute{
				Description:   "Optionnal address",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *socketSiteResource) schemaSockets() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description:   "Socket information",
		Computed:      true,
		PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description:   "Socket ID",
					Computed:      true,
					PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				},
				"serial_number": schema.StringAttribute{
					Description:   "Socket serial number",
					Computed:      true,
					PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				},
				"is_primary": schema.BoolAttribute{
					Description:   "Indicates if the socket is primary",
					Computed:      true,
					PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
				},
				"platform": schema.StringAttribute{
					Description:   "Socket platform",
					Computed:      true,
					PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				},
			},
		},
	}
}
