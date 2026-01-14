package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

const defaultInterfaceByConnType = `{
	"SOCKET_AWS1500":   "LAN1",
	"SOCKET_AZ1500":    "LAN1",
	"SOCKET_ESX1500":   "LAN1",
	"SOCKET_GCP1500":   "LAN1",
	"SOCKET_X1500":     "LAN1",
	"SOCKET_X1600":     "INT_5",
	"SOCKET_X1600_LTE": "INT_5",
	"SOCKET_X1700":     "INT_3"
}`

var (
	_ resource.Resource                = &socketSiteResource{}
	_ resource.ResourceWithConfigure   = &socketSiteResource{}
	_ resource.ResourceWithImportState = &socketSiteResource{}
)

func NewSocketSiteResource() resource.Resource {
	return &socketSiteResource{}
}

type socketSiteResource struct {
	client *catoClientData
}

func (r *socketSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_site"
}

func (r *socketSiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_socket_site` resource contains the configuration parameters necessary to add a socket site to the Cato cloud ([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.addSocketSite()](https://api.catonetworks.com/documentation/#mutation-site.addSocketSite). \n\n **Note**: For AWS deployments, please accept the [EULA for the Cato Networks AWS Marketplace product](https://aws.amazon.com/marketplace/pp?sku=dvfhly9fuuu67tw59c7lt5t3c).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Site ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Site name",
				Required:    true,
			},
			"connection_type": schema.StringAttribute{
				Description: "Connection type for the site (SOCKET_X1500, SOCKET_AWS1500, SOCKET_AZ1500, ...)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"SOCKET_AWS1500",
						"SOCKET_AZ1500",
						"SOCKET_ESX1500",
						"SOCKET_GCP1500",
						"SOCKET_X1500",
						"SOCKET_X1600",
						"SOCKET_X1600_LTE",
						"SOCKET_X1700",
					),
				},
			},
			"site_type": schema.StringAttribute{
				Description: "Site type (https://api.catonetworks.com/documentation/#definition-SiteType)",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Site description",
				Optional:    true,
			},
			"native_range": schema.SingleNestedAttribute{
				Description: "Site lan native range settings",
				Required:    true,
				Validators: []validator.Object{
					socketSiteNativeRangeValidator{},
				},
				Attributes: map[string]schema.Attribute{
					"interface_index": schema.StringAttribute{
						Description: "LAN native range interface index, default is LAN1 for SOCKET_X1500 models, INT_5 for SOCKET_X1600 and SOCKET_X1600_LTE, and INT_3 for SOCKET_X1700 models",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("LAN1", "LAN2", "INT_1", "INT_2", "INT_3", "INT_4", "INT_5", "INT_6", "INT_7", "INT_8", "INT_9", "INT_10", "INT_11", "INT_12"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"interface_id": schema.StringAttribute{
						Description: "LAN native range interface id",
						Computed:    true,
					},
					"native_network_range": schema.StringAttribute{
						Description: "Site native IP range (CIDR)",
						Required:    true,
					},
					"native_network_lan_interface_id": schema.StringAttribute{
						Description: "ID of native range LAN interface (for additional network range update purposes)",
						Optional:    true,
						Computed:    true,
					},
					"native_network_range_id": schema.StringAttribute{
						Description: "Site native IP range ID (for update purpose)",
						Optional:    true,
						Computed:    true,
					},
					"interface_name": schema.StringAttribute{
						Description: "LAN native range interface name (e.g., 'LAN 01')",
						Optional:    true,
						Computed:    true,
						// PlanModifiers: []planmodifier.String{
						// 	stringplanmodifier.UseStateForUnknown(),
						// },
					},
					"range_name": schema.StringAttribute{
						Description: "Native range name (typically 'Native Range')",
						Computed:    true,
						Optional:    false,
					},
					"range_id": schema.StringAttribute{
						Description: "Native range ID (base64 encoded identifier)",
						Computed:    true,
						// PlanModifiers: []planmodifier.String{
						// 	stringplanmodifier.UseStateForUnknown(),
						// },
					},
					"gateway": schema.StringAttribute{
						Description: "Gateway IP address for the native range",
						Computed:    true,
					},
					"vlan": schema.Int64Attribute{
						Description: "VLAN ID for the site native range (optional)",
						Optional:    true,
						Validators: []validator.Int64{
							int64validator.AtLeast(1),
						},
					},
					// "internet_only": schema.BoolAttribute{
					// 	Description: "Internet only network range (Only releveant for Routed range_type)",
					// 	Computed:    true,
					// 	Optional:    true,
					// 	Default:     booldefault.StaticBool(false),
					// 	PlanModifiers: []planmodifier.Bool{
					// 		boolplanmodifier.UseStateForUnknown(),
					// 	},
					// },
					"mdns_reflector": schema.BoolAttribute{
						Description: "Site native range mDNS reflector. When enabled, the Socket functions as an mDNS gateway, it relays mDNS requests and response between all enabled subnets.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"local_ip": schema.StringAttribute{
						Description: "Site native range local ip",
						Required:    true,
					},
					"range_type": schema.StringAttribute{
						Description: "NATIVE",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"translated_subnet": schema.StringAttribute{
						Description: "Site translated native IP range (CIDR)",
						Computed:    true,
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"lag_min_links": schema.Int64Attribute{
						Description: "Number of interfaces to include in the link aggregation, only relevant for LAN_LAG_MASTER and LAN_LAG_MASTER_AND_VRRP interface destination types",
						Optional:    true,
						Validators: []validator.Int64{
							int64validator.AtLeast(1),
						},
					},
					"interface_dest_type": schema.StringAttribute{
						Description: "Socket interface destination type for the native interface, example values: LAN, LAN_LAG_MASTER, LAN_LAG_MASTER_AND_VRRP, LAN_AND_HA, VRRP, VRRP_AND_LAN",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("LAN"),
						Validators: []validator.String{
							stringvalidator.OneOf(
								"LAN",
								"LAN_LAG_MASTER",
								"LAN_LAG_MASTER_AND_VRRP",
								"LAN_AND_HA",
								"VRRP",
								"VRRP_AND_LAN",
							),
						},
					},
					"dhcp_settings": schema.SingleNestedAttribute{
						Description: "Site native range DHCP settings (Only releveant for NATIVE and VLAN range_type)",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"dhcp_type": schema.StringAttribute{
								Description: "Network range dhcp type (https://api.catonetworks.com/documentation/#definition-DhcpType)",
								Required:    true,
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
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.Expressions{
										path.MatchRelative().AtParent().AtName("relay_group_id"),
									}...),
								},
							},
							"dhcp_microsegmentation": schema.BoolAttribute{
								Description: "DHCP Microsegmentation. When enabled, the DHCP server will allocate /32 subnet mask. Make sure to enable the proper Firewall rules and enable it with caution, as it is not supported on all operating systems; monitor the network closely after activation. This setting can only be configured when dhcp_type is set to DHCP_RANGE.",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
			},
			"site_location": schema.SingleNestedAttribute{
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
						Description: "Optionnal city",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"address": schema.StringAttribute{
						Description: "Optionnal address",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
				},
			},
		},
	}
}

func (r *socketSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *socketSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *socketSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan SocketSite
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.plan", map[string]interface{}{
		"plan": utils.InterfaceToJSONString(plan),
	})

	var interfaceByConnType map[string]string
	err := json.Unmarshal([]byte(defaultInterfaceByConnType), &interfaceByConnType)
	if err != nil {
		panic(err)
	}

	// setting input & input to update network range
	input := cato_models.AddSocketSiteInput{}
	inputUpdateNetworkRange := cato_models.UpdateNetworkRangeInput{}
	inputUpdateSocketInterface := cato_models.UpdateSocketInterfaceInput{}

	// setting input site location
	if !plan.SiteLocation.IsNull() && !plan.SiteLocation.IsUnknown() {
		input.SiteLocation = &cato_models.AddSiteLocationInput{}
		siteLocationInput := SiteLocation{}
		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// Normalize address: if not provided or empty, send nil (to avoid empty string -> null drift)
		addrPtr := siteLocationInput.Address.ValueStringPointer()
		if addrPtr != nil && *addrPtr == "" {
			addrPtr = nil
		}
		input.SiteLocation.Address = addrPtr
		// Normalize city: if not provided or empty, send nil (to avoid empty string -> null drift)
		cityPtr := siteLocationInput.City.ValueStringPointer()
		if cityPtr != nil && *cityPtr == "" {
			cityPtr = nil
		}
		input.SiteLocation.City = cityPtr
		input.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueString()
		input.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()
		input.SiteLocation.Timezone = siteLocationInput.Timezone.ValueString()
	}

	// setting input native range
	tflog.Debug(ctx, "Create.plan.NativeRange check", map[string]interface{}{
		"IsNull":    plan.NativeRange.IsNull(),
		"IsUnknown": plan.NativeRange.IsUnknown(),
		"Value":     utils.InterfaceToJSONString(plan.NativeRange),
	})

	if !plan.NativeRange.IsNull() && !plan.NativeRange.IsUnknown() {
		nativeRangeInput := NativeRange{}
		diags = plan.NativeRange.As(ctx, &nativeRangeInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "Error parsing native_range from plan", map[string]interface{}{
				"diagnostics": utils.InterfaceToJSONString(resp.Diagnostics),
			})
			return
		}

		// Validate that local_ip is within native_network_range
		if !nativeRangeInput.LocalIp.IsNull() && !nativeRangeInput.LocalIp.IsUnknown() &&
			!nativeRangeInput.NativeNetworkRange.IsNull() && !nativeRangeInput.NativeNetworkRange.IsUnknown() {
			localIPStr := nativeRangeInput.LocalIp.ValueString()
			subnetStr := nativeRangeInput.NativeNetworkRange.ValueString()

			// Parse the local IP
			ip := net.ParseIP(localIPStr)
			if ip == nil {
				resp.Diagnostics.AddError(
					"Invalid Local IP",
					fmt.Sprintf("local_ip '%s' is not a valid IP address", localIPStr),
				)
				return
			}

			// Parse the subnet CIDR
			_, ipNet, err := net.ParseCIDR(subnetStr)
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid Native Network Range",
					fmt.Sprintf("native_network_range '%s' is not a valid CIDR notation", subnetStr),
				)
				return
			}

			// Check if the IP is within the subnet
			if !ipNet.Contains(ip) {
				resp.Diagnostics.AddError(
					"Local IP Not in Native Range",
					fmt.Sprintf("Local IP must be within the Native Range IP. local_ip '%s' is not within native_network_range '%s'", localIPStr, subnetStr),
				)
				return
			}
		}

		tflog.Debug(ctx, "Create.nativeRangeInput.InterfaceDestType", map[string]interface{}{
			"nativeRangeInput.InterfaceDestType": utils.InterfaceToJSONString(nativeRangeInput.InterfaceDestType.ValueString()),
		})
		interfaceDestType := nativeRangeInput.InterfaceDestType.ValueString()
		if interfaceDestType == "" {
			interfaceDestType = "LAN" // Use default if not specified
		}
		hasLagMinLinks := !nativeRangeInput.LagMinLinks.IsNull() && !nativeRangeInput.LagMinLinks.IsUnknown()

		// Rule 1: If interface_dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, lag_min_links must have a value
		if (interfaceDestType == "LAN_LAG_MASTER" || interfaceDestType == "LAN_LAG_MASTER_AND_VRRP") && !hasLagMinLinks {
			resp.Diagnostics.AddError(
				"Invalid LAG Configuration",
				fmt.Sprintf("When interface_dest_type is %s, lag_min_links must be specified.", interfaceDestType),
			)
			return
		}

		// Rule 2: If lag_min_links has a value, interface_dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
		if hasLagMinLinks && interfaceDestType != "LAN_LAG_MASTER" && interfaceDestType != "LAN_LAG_MASTER_AND_VRRP" {
			resp.Diagnostics.AddError(
				"Invalid LAG Configuration",
				fmt.Sprintf("lag_min_links can only be configured when interface_dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, but interface_dest_type is %s.", interfaceDestType),
			)
			return
		}

		// Validate that interface_index can only be set for specific connection types
		// Only validate if user is setting a non-default value
		hasInterfaceIndex := !nativeRangeInput.InterfaceIndex.IsNull() && !nativeRangeInput.InterfaceIndex.IsUnknown() && nativeRangeInput.InterfaceIndex.ValueString() != ""
		if hasInterfaceIndex {
			connType := plan.ConnectionType.ValueString()
			interfaceIndexValue := nativeRangeInput.InterfaceIndex.ValueString()
			defaultInterface, hasDefault := interfaceByConnType[connType]

			// Only validate if trying to set a non-default interface_index for non-X1600/X1700 types
			if connType != "SOCKET_X1600" && connType != "SOCKET_X1600_LTE" && connType != "SOCKET_X1700" {
				// Allow if it's the default interface for this connection type
				if !hasDefault || interfaceIndexValue != defaultInterface {
					resp.Diagnostics.AddError(
						"Invalid Interface Index Configuration",
						fmt.Sprintf("interface_index can only be explicitly configured for SOCKET_X1600, SOCKET_X1600_LTE, or SOCKET_X1700 connection types, but connection_type is %s", connType),
					)
					return
				}
			}
		}

		tflog.Debug(ctx, "Create.nativeRangeInput", map[string]interface{}{
			"nativeRangeInput":                    utils.InterfaceToJSONString(nativeRangeInput),
			"nativeRangeInput.NativeNetworkRange": utils.InterfaceToJSONString(nativeRangeInput.NativeNetworkRange.ValueString()),
			"nativeRangeInput.RangeName":          utils.InterfaceToJSONString(nativeRangeInput.RangeName.ValueStringPointer()),
		})
		input.NativeNetworkRange = nativeRangeInput.NativeNetworkRange.ValueString()
		input.TranslatedSubnet = nativeRangeInput.TranslatedSubnet.ValueStringPointer()

		// inputUpdateNetworkRange.Name = nativeRangeInput.RangeName.ValueStringPointer() // The API does not update this attribute for native ranges
		inputUpdateNetworkRange.Subnet = nativeRangeInput.NativeNetworkRange.ValueStringPointer()
		inputUpdateNetworkRange.LocalIP = nativeRangeInput.LocalIp.ValueStringPointer()
		inputUpdateNetworkRange.MdnsReflector = nativeRangeInput.MdnsReflector.ValueBoolPointer()
		inputUpdateNetworkRange.TranslatedSubnet = nativeRangeInput.TranslatedSubnet.ValueStringPointer()
		inputUpdateNetworkRange.Vlan = nativeRangeInput.Vlan.ValueInt64Pointer()
		inputUpdateSocketInterface.DestType = cato_models.SocketInterfaceDestType(interfaceDestType)
		inputUpdateSocketInterface.Name = nativeRangeInput.InterfaceName.ValueStringPointer()
		if (interfaceDestType == "LAN_LAG_MASTER" || interfaceDestType == "LAN_LAG_MASTER_AND_VRRP") && hasLagMinLinks {
			lagConfig := cato_models.SocketInterfaceLagInput{
				MinLinks: nativeRangeInput.LagMinLinks.ValueInt64(),
			}
			inputUpdateSocketInterface.Lag = &lagConfig
		}
		socketInterfaceLanInput := cato_models.SocketInterfaceLanInput{}
		if localIP := nativeRangeInput.LocalIp.ValueStringPointer(); localIP != nil {
			socketInterfaceLanInput.LocalIP = *localIP // string
		}
		if subnet := nativeRangeInput.NativeNetworkRange.ValueStringPointer(); subnet != nil {
			socketInterfaceLanInput.Subnet = *subnet // string
		}
		inputUpdateSocketInterface.Lan = &socketInterfaceLanInput

		// setting input native range DHCP settings
		if !nativeRangeInput.DhcpSettings.IsNull() && !nativeRangeInput.DhcpSettings.IsUnknown() {
			inputUpdateNetworkRange.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{}
			dhcpSettingsInput := DhcpSettings{}
			diags = nativeRangeInput.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			inputUpdateNetworkRange.DhcpSettings.DhcpType = (cato_models.DhcpType)(dhcpSettingsInput.DhcpType.ValueString())
			inputUpdateNetworkRange.DhcpSettings.IPRange = dhcpSettingsInput.IpRange.ValueStringPointer()

			// Validate that relay_group_name or relay_group_id are only set when dhcp_type is DHCP_RELAY
			if dhcpSettingsInput.DhcpType.ValueString() != "DHCP_RELAY" {
				hasRelayGroupName := !dhcpSettingsInput.RelayGroupName.IsNull() && !dhcpSettingsInput.RelayGroupName.IsUnknown() && dhcpSettingsInput.RelayGroupName.ValueString() != ""
				hasRelayGroupId := !dhcpSettingsInput.RelayGroupId.IsNull() && !dhcpSettingsInput.RelayGroupId.IsUnknown() && dhcpSettingsInput.RelayGroupId.ValueString() != ""
				if hasRelayGroupName || hasRelayGroupId {
					resp.Diagnostics.AddError(
						"Invalid DHCP Relay Configuration",
						fmt.Sprintf("relay_group_name and relay_group_id can only be configured when dhcp_type is DHCP_RELAY, but dhcp_type is %s", dhcpSettingsInput.DhcpType.ValueString()),
					)
					return
				}
			}

			// Validate that ip_range is only set when dhcp_type is DHCP_RANGE
			if dhcpSettingsInput.DhcpType.ValueString() != "DHCP_RANGE" {
				hasIpRange := !dhcpSettingsInput.IpRange.IsNull() && !dhcpSettingsInput.IpRange.IsUnknown() && dhcpSettingsInput.IpRange.ValueString() != ""
				if hasIpRange {
					resp.Diagnostics.AddError(
						"Invalid DHCP Range Configuration",
						fmt.Sprintf("ip_range can only be configured when dhcp_type is DHCP_RANGE, but dhcp_type is %s", dhcpSettingsInput.DhcpType.ValueString()),
					)
					return
				}
			}

			// Validate that dhcp_microsegmentation is only set to true when dhcp_type is DHCP_RANGE
			if !dhcpSettingsInput.DhcpMicrosegmentation.IsNull() && !dhcpSettingsInput.DhcpMicrosegmentation.IsUnknown() {
				// set to false if dhcp_microsegmentation is not set for DHCP_RANGE use case
				fmt.Println("dhcpSettingsInput.DhcpMicrosegmentation.ValueBool() " + fmt.Sprintf("%v", dhcpSettingsInput.DhcpMicrosegmentation.ValueBool()))
				if dhcpSettingsInput.DhcpMicrosegmentation.ValueBool() == true && dhcpSettingsInput.DhcpType.ValueString() != "DHCP_RANGE" {
					resp.Diagnostics.AddError(
						"Invalid DHCP Microsegmentation Configuration",
						"dhcp_microsegmentation can only be configured when dhcp_type is set to DHCP_RANGE",
					)
					return
				}
			}

			// Only set dhcpMicrosegmentation for DHCP_RANGE type
			if dhcpSettingsInput.DhcpType.ValueString() == "DHCP_RANGE" {
				inputUpdateNetworkRange.DhcpSettings.DhcpMicrosegmentation = dhcpSettingsInput.DhcpMicrosegmentation.ValueBoolPointer()
			}

			// Validate DHCP relay group configuration when dhcp_type is DHCP_RELAY
			if dhcpSettingsInput.DhcpType.ValueString() == "DHCP_RELAY" {
				relayGroupName := ""
				relayGroupId := ""

				if !dhcpSettingsInput.RelayGroupName.IsNull() && !dhcpSettingsInput.RelayGroupName.IsUnknown() {
					relayGroupName = dhcpSettingsInput.RelayGroupName.ValueString()
				}
				if !dhcpSettingsInput.RelayGroupId.IsNull() && !dhcpSettingsInput.RelayGroupId.IsUnknown() {
					relayGroupId = dhcpSettingsInput.RelayGroupId.ValueString()
				}

				resolvedRelayGroupId, success, err := checkForDhcpRelayGroup(ctx, r.client, relayGroupName, relayGroupId)
				if err != nil {
					resp.Diagnostics.AddError(
						"DHCP Relay Configuration Error",
						err.Error(),
					)
					return
				}
				if !success {
					resp.Diagnostics.AddError(
						"DHCP Relay Group Validation Failed",
						"Failed to validate DHCP relay group configuration.",
					)
					return
				}

				// Set the resolved relay group ID
				inputUpdateNetworkRange.DhcpSettings.RelayGroupID = &resolvedRelayGroupId
			}
		}
	}

	// setting input other attributes
	input.Name = plan.Name.ValueString()
	input.ConnectionType = (cato_models.SiteConnectionTypeEnum)(plan.ConnectionType.ValueString())
	input.SiteType = (cato_models.SiteType)(plan.SiteType.ValueString())
	input.Description = plan.Description.ValueStringPointer()

	tflog.Debug(ctx, "Create.SiteAddSocketSite.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	socketSite, err := r.client.catov2.SiteAddSocketSite(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddSocketSite.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(socketSite),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteAddSocketSite error",
			err.Error(),
		)
		return
	}
	siteID := socketSite.Site.AddSocketSite.GetSiteID()

	// Added section to handle isDefault where socket site is not able to define custom index for lan
	// Logic to move the interface by creating a secondary lan interface disabling the default after
	// retrieving native-network range ID to update native range
	// Lookup default lan index by socket connection type
	var defaultLanInterfaceIndex *string
	if val, ok := interfaceByConnType[plan.ConnectionType.ValueString()]; ok {
		defaultLanInterfaceIndex = &val
	}

	if !plan.NativeRange.IsNull() && !plan.NativeRange.IsUnknown() {
		nativeRangeCheck := NativeRange{}
		diags = plan.NativeRange.As(ctx, &nativeRangeCheck, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if defaultLanInterfaceIndex != nil && !nativeRangeCheck.InterfaceIndex.IsNull() && !nativeRangeCheck.InterfaceIndex.IsUnknown() &&
			*defaultLanInterfaceIndex != nativeRangeCheck.InterfaceIndex.ValueString() {
			// Move the default interface by creating a new interface, disabling the original
			// then updating the new interface with the desired range
			//cato_models.SocketInterfaceIDEnum
			interfaceIndex := nativeRangeCheck.InterfaceIndex.ValueString()
			interfaceName := nativeRangeCheck.InterfaceName.ValueString()
			if interfaceName == "" {
				interfaceName = interfaceIndex
			}
			localIp := nativeRangeCheck.LocalIp.ValueString()
			nativeNetworkRange := nativeRangeCheck.NativeNetworkRange.ValueString()
			interfaceDestType := nativeRangeCheck.InterfaceDestType.ValueString()
			lagMinLinks := nativeRangeCheck.LagMinLinks.ValueInt64()
			translatedSubnet := nativeRangeCheck.TranslatedSubnet.ValueString()
			tflog.Debug(ctx, "Create.nativeRangeCheck", map[string]interface{}{
				"interfaceIndex":     utils.InterfaceToJSONString(interfaceIndex),
				"interfaceName":      utils.InterfaceToJSONString(interfaceName),
				"localIp":            utils.InterfaceToJSONString(localIp),
				"nativeNetworkRange": utils.InterfaceToJSONString(nativeNetworkRange),
				"interfaceDestType":  utils.InterfaceToJSONString(interfaceDestType),
				"lagMinLinks":        utils.InterfaceToJSONString(lagMinLinks),
				"translatedSubnet":   utils.InterfaceToJSONString(translatedSubnet),
			})
			isDefaultPresent, err := r.attemptReassignNativeRangeIndex(ctx, cato_models.SocketInterfaceIDEnum(interfaceIndex), interfaceName, localIp, nativeNetworkRange, siteID, interfaceDestType, lagMinLinks, translatedSubnet)
			if err != nil {
				resp.Diagnostics.AddError(
					"Catov2 API SiteAddSocketSite error - create.attemptReassignNativeRangeIndex()",
					err.Error(),
				)
				return
			}
			if isDefaultPresent == nil || !*isDefaultPresent {
				resp.Diagnostics.AddError(
					"Configuration Error",
					"Invalid interface configuration, the api to support moving the lan interface index to '"+interfaceIndex+"' from the default index of '"+*defaultLanInterfaceIndex+"' for the socket connectionType of "+plan.ConnectionType.ValueString()+"' is not yet supported.  This API is being gradually rolled out.",
				)
				return
			}
		}
	}

	// Get native interface and subnet information
	nativeInterfaceAndSubnet, err := r.getNativeInterfaceAndSubnet(ctx, string(input.ConnectionType), siteID, plan, interfaceByConnType)
	if err != nil {
		return
	}
	nativeNetworkRangeId := nativeInterfaceAndSubnet.NativeNetworkRangeId

	tflog.Debug(ctx, "Create.SiteUpdateNetworkRange.request", map[string]interface{}{
		"nativeNetworkRangeId": utils.InterfaceToJSONString(nativeNetworkRangeId),
		"request":              utils.InterfaceToJSONString(inputUpdateNetworkRange),
	})

	// Update native network range
	_, err = r.client.catov2.SiteUpdateNetworkRange(ctx, nativeNetworkRangeId, inputUpdateNetworkRange, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateNetworkRange error",
			err.Error(),
		)
		return
	}

	// Update native socket interface
	if inputUpdateSocketInterface.Name == nil || (inputUpdateSocketInterface.Name != nil && *inputUpdateSocketInterface.Name == "") {
		inputUpdateSocketInterface.Name = &nativeInterfaceAndSubnet.InterfaceIndex
	}
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateSocketInterface),
		"nativeInterfaceAndSubnet.InterfaceIndex": utils.InterfaceToJSONString(nativeInterfaceAndSubnet.InterfaceIndex),
		"nativeInterfaceAndSubnet.InterfaceName":  utils.InterfaceToJSONString(nativeInterfaceAndSubnet.InterfaceName),
		"inputUpdateSocketInterface.Name":         utils.InterfaceToJSONString(inputUpdateSocketInterface.Name),
		"siteID":                                  utils.InterfaceToJSONString(siteID),
	})
	siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, siteID, cato_models.SocketInterfaceIDEnum(nativeInterfaceAndSubnet.InterfaceIndex), inputUpdateSocketInterface, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateNetworkRange error",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Create.SiteUpdateSocketInterface.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
	})

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, plan, siteID)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating socket site state",
			hydrateErr.Error(),
		)
		return
	}

	// check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overiding state with socket site id
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), siteID)
	// overiding state with native network range id
	resp.State.SetAttribute(ctx, path.Root("native_range").AtName("native_network_range_id"), nativeNetworkRangeId)
}

func (r *socketSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, state, state.Id.ValueString())
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating socket site state",
			hydrateErr.Error(),
		)
		return
	}

	// check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *socketSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan SocketSite
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state to preserve computed values
	var state SocketSite
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var interfaceByConnType map[string]string
	err := json.Unmarshal([]byte(defaultInterfaceByConnType), &interfaceByConnType)
	if err != nil {
		panic(err)
	}

	// setting input & input to update network range
	inputSiteGeneral := cato_models.UpdateSiteGeneralDetailsInput{
		SiteLocation: &cato_models.UpdateSiteLocationInput{},
	}

	inputUpdateNetworkRange := cato_models.UpdateNetworkRangeInput{
		DhcpSettings: &cato_models.NetworkDhcpSettingsInput{
			DhcpType: (cato_models.DhcpType)("DHCP_DISABLED"),
		},
	}
	inputUpdateSocketInterface := cato_models.UpdateSocketInterfaceInput{}

	// setting input site location
	if !plan.SiteLocation.IsNull() && !plan.SiteLocation.IsUnknown() {
		inputSiteGeneral.SiteLocation = &cato_models.UpdateSiteLocationInput{}
		siteLocationInput := SiteLocation{}
		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// Get state location for comparison
		var stateLocationInput SiteLocation
		if !state.SiteLocation.IsNull() && !state.SiteLocation.IsUnknown() {
			diags = state.SiteLocation.As(ctx, &stateLocationInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
		}

		// Handle city field
		cityPtr := siteLocationInput.City.ValueStringPointer()
		if (cityPtr == nil || (cityPtr != nil && *cityPtr == "")) &&
			!stateLocationInput.City.IsNull() && !stateLocationInput.City.IsUnknown() && stateLocationInput.City.ValueString() != "" {
			// API bug workaround: if city had a value in state and is now blank/null in plan,
			// send " " (single space) to clear the field
			spaceCityValue := " "
			cityPtr = &spaceCityValue
		} else if cityPtr != nil && *cityPtr == "" {
			// Normal case: empty string becomes nil
			cityPtr = nil
		}
		inputSiteGeneral.SiteLocation.CityName = cityPtr
		inputSiteGeneral.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueStringPointer()
		inputSiteGeneral.SiteLocation.Timezone = siteLocationInput.Timezone.ValueStringPointer()
		inputSiteGeneral.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()

		// Handle address field
		addrPtr := siteLocationInput.Address.ValueStringPointer()
		if (addrPtr == nil || (addrPtr != nil && *addrPtr == "")) &&
			!stateLocationInput.Address.IsNull() && !stateLocationInput.Address.IsUnknown() && stateLocationInput.Address.ValueString() != "" {
			// API bug workaround: if address had a value in state and is now blank/null in plan,
			// send "" (empty string) to clear the field
			emptyAddrValue := ""
			addrPtr = &emptyAddrValue
		} else if addrPtr != nil && *addrPtr == "" {
			// Normal case: empty string becomes nil
			addrPtr = nil
		}
		inputSiteGeneral.SiteLocation.Address = addrPtr
	}

	// setting input native range
	var nativeRangeState NativeRange
	diags = state.NativeRange.As(ctx, &nativeRangeState, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	if !plan.NativeRange.IsNull() && !plan.NativeRange.IsUnknown() {

		nativeRangeInput := NativeRange{}
		diags = plan.NativeRange.As(ctx, &nativeRangeInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// Validate that local_ip is within native_network_range
		if !nativeRangeInput.LocalIp.IsNull() && !nativeRangeInput.LocalIp.IsUnknown() &&
			!nativeRangeInput.NativeNetworkRange.IsNull() && !nativeRangeInput.NativeNetworkRange.IsUnknown() {
			localIPStr := nativeRangeInput.LocalIp.ValueString()
			subnetStr := nativeRangeInput.NativeNetworkRange.ValueString()

			// Parse the local IP
			ip := net.ParseIP(localIPStr)
			if ip == nil {
				resp.Diagnostics.AddError(
					"Invalid Local IP",
					fmt.Sprintf("local_ip '%s' is not a valid IP address", localIPStr),
				)
				return
			}

			// Parse the subnet CIDR
			_, ipNet, err := net.ParseCIDR(subnetStr)
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid Native Network Range",
					fmt.Sprintf("native_network_range '%s' is not a valid CIDR notation", subnetStr),
				)
				return
			}

			// Check if the IP is within the subnet
			if !ipNet.Contains(ip) {
				resp.Diagnostics.AddError(
					"Local IP Not in Native Range",
					fmt.Sprintf("Local IP must be within the Native Range IP. local_ip '%s' is not within native_network_range '%s'", localIPStr, subnetStr),
				)
				return
			}
		}

		// Validate LAG configuration
		interfaceDestType := nativeRangeInput.InterfaceDestType.ValueString()
		if interfaceDestType == "" {
			interfaceDestType = "LAN" // Use default if not specified
		}
		hasLagMinLinks := !nativeRangeInput.LagMinLinks.IsNull() && !nativeRangeInput.LagMinLinks.IsUnknown()

		// Rule 1: If interface_dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, lag_min_links must have a value
		if (interfaceDestType == "LAN_LAG_MASTER" || interfaceDestType == "LAN_LAG_MASTER_AND_VRRP") && !hasLagMinLinks {
			resp.Diagnostics.AddError(
				"Invalid LAG Configuration",
				fmt.Sprintf("When interface_dest_type is %s, lag_min_links must be specified.", interfaceDestType),
			)
			return
		}

		// Rule 2: If lag_min_links has a value, interface_dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
		if hasLagMinLinks && interfaceDestType != "LAN_LAG_MASTER" && interfaceDestType != "LAN_LAG_MASTER_AND_VRRP" {
			resp.Diagnostics.AddError(
				"Invalid LAG Configuration",
				fmt.Sprintf("lag_min_links can only be configured when interface_dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, but interface_dest_type is %s.", interfaceDestType),
			)
			return
		}

		// Validate that interface_index can only be set for specific connection types
		// Only validate if user is setting a non-default value
		hasInterfaceIndex := !nativeRangeInput.InterfaceIndex.IsNull() && !nativeRangeInput.InterfaceIndex.IsUnknown() && nativeRangeInput.InterfaceIndex.ValueString() != ""
		if hasInterfaceIndex {
			connType := plan.ConnectionType.ValueString()
			interfaceIndexValue := nativeRangeInput.InterfaceIndex.ValueString()
			defaultInterface, hasDefault := interfaceByConnType[connType]

			// Only validate if trying to set a non-default interface_index for non-X1600/X1700 types
			if connType != "SOCKET_X1600" && connType != "SOCKET_X1600_LTE" && connType != "SOCKET_X1700" {
				// Allow if it's the default interface for this connection type
				if !hasDefault || interfaceIndexValue != defaultInterface {
					resp.Diagnostics.AddError(
						"Invalid Interface Index Configuration",
						fmt.Sprintf("interface_index can only be explicitly configured for SOCKET_X1600, SOCKET_X1600_LTE, or SOCKET_X1700 connection types, but connection_type is %s", connType),
					)
					return
				}
			}
		}

		// Check for interface_name removal from config
		tflog.Info(ctx, "nativeRangeState.InterfaceName.check", map[string]interface{}{
			"nativeRangeState.InterfaceName":                   utils.InterfaceToJSONString(nativeRangeState.InterfaceName),
			"nativeRangeState.InterfaceName.IsNull()":          nativeRangeState.InterfaceName.IsNull(),
			"nativeRangeState.InterfaceName.IsUnknown()":       nativeRangeState.InterfaceName.IsUnknown(),
			"nativeRangeState.InterfaceName.ValueString()!=''": nativeRangeState.InterfaceName.ValueString() != "",
			"nativeRangeState.InterfaceName.ValueString()":     nativeRangeState.InterfaceName.ValueString(),
		})
		interfaceNameRemovedFromConfig := false
		if !nativeRangeState.InterfaceName.IsNull() && !nativeRangeState.InterfaceName.IsUnknown() &&
			nativeRangeState.InterfaceName.ValueString() != "" {
			// Interface name exists in state
			tflog.Info(ctx, "nativeRangeInput.InterfaceName.check", map[string]interface{}{
				"nativeRangeInput.InterfaceName":                   utils.InterfaceToJSONString(nativeRangeInput.InterfaceName),
				"nativeRangeInput.InterfaceName.IsNull()":          nativeRangeInput.InterfaceName.IsNull(),
				"nativeRangeInput.InterfaceName.IsUnknown()":       nativeRangeInput.InterfaceName.IsUnknown(),
				"nativeRangeInput.InterfaceName.ValueString()!=''": nativeRangeInput.InterfaceName.ValueString() != "",
				"nativeRangeInput.InterfaceName.ValueString()":     nativeRangeInput.InterfaceName.ValueString(),
			})

			if nativeRangeInput.InterfaceName.IsNull() || nativeRangeInput.InterfaceName.IsUnknown() ||
				nativeRangeInput.InterfaceName.ValueString() == "" {
				// But doesn't exist (or is empty) in plan - it was removed from config
				interfaceNameRemovedFromConfig = true
				tflog.Info(ctx, "Detected interface_name removal from configuration", map[string]interface{}{
					"state_interface_name": nativeRangeState.InterfaceName.ValueString(),
					"plan_interface_name":  nativeRangeInput.InterfaceName.ValueString(),
				})
			}
		}

		inputUpdateNetworkRange.Subnet = nativeRangeInput.NativeNetworkRange.ValueStringPointer()
		inputUpdateNetworkRange.LocalIP = nativeRangeInput.LocalIp.ValueStringPointer()
		inputUpdateNetworkRange.MdnsReflector = nativeRangeInput.MdnsReflector.ValueBoolPointer()
		inputUpdateNetworkRange.Vlan = nativeRangeInput.Vlan.ValueInt64Pointer()

		// Handle interface name changes/removals
		if interfaceNameRemovedFromConfig {
			// Removed from local config to reset todefault value of interface index
			inputUpdateSocketInterface.Name = nativeRangeInput.InterfaceIndex.ValueStringPointer()
			tflog.Info(ctx, "inputUpdateSocketInterface.Name=Removed from local config to reset to default value of interface index", map[string]interface{}{
				"inputUpdateSocketInterface.Name": utils.InterfaceToJSONString(inputUpdateSocketInterface.Name),
			})
		} else if !nativeRangeInput.InterfaceName.IsNull() && !nativeRangeInput.InterfaceName.IsUnknown() {
			// Interface name exists in plan - use it
			inputUpdateSocketInterface.Name = nativeRangeInput.InterfaceName.ValueStringPointer()
			tflog.Info(ctx, "inputUpdateSocketInterface.Name=Interface name exists in plan - use it", map[string]interface{}{
				"inputUpdateSocketInterface.Name": utils.InterfaceToJSONString(inputUpdateSocketInterface.Name),
			})
		} else {
			// No interface name in plan, use what's in state if available
			inputUpdateSocketInterface.Name = nativeRangeState.InterfaceName.ValueStringPointer()
			tflog.Info(ctx, "inputUpdateSocketInterface.Name=No interface name in plan, use what's in state if available", map[string]interface{}{
				"inputUpdateSocketInterface.Name": utils.InterfaceToJSONString(inputUpdateSocketInterface.Name),
			})
		}

		// Use the interfaceDestType string variable for the check, not the cast result
		inputUpdateSocketInterface.DestType = cato_models.SocketInterfaceDestType(interfaceDestType)

		// Add LAG configuration if needed
		if (interfaceDestType == "LAN_LAG_MASTER" || interfaceDestType == "LAN_LAG_MASTER_AND_VRRP") && hasLagMinLinks {
			lagConfig := cato_models.SocketInterfaceLagInput{
				MinLinks: nativeRangeInput.LagMinLinks.ValueInt64(),
			}
			inputUpdateSocketInterface.Lag = &lagConfig
		}

		socketInterfaceLanInput := cato_models.SocketInterfaceLanInput{}
		// Use plan values (nativeRangeInput) to ensure consistency with network range update
		if localIP := nativeRangeInput.LocalIp.ValueStringPointer(); localIP != nil {
			socketInterfaceLanInput.LocalIP = *localIP // string
		}
		if subnet := nativeRangeInput.NativeNetworkRange.ValueStringPointer(); subnet != nil {
			socketInterfaceLanInput.Subnet = *subnet // string
		}
		inputUpdateSocketInterface.Lan = &socketInterfaceLanInput

		// setting input native range DHCP settings
		if !nativeRangeInput.DhcpSettings.IsNull() && !nativeRangeInput.DhcpSettings.IsUnknown() {
			// Configuration has dhcp_settings block - use it
			inputUpdateNetworkRange.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{}
			dhcpSettingsState := DhcpSettings{}
			diags = nativeRangeInput.DhcpSettings.As(ctx, &dhcpSettingsState, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			dhcpSettingsInput := DhcpSettings{}
			diags = nativeRangeInput.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			inputUpdateNetworkRange.DhcpSettings.DhcpType = (cato_models.DhcpType)(dhcpSettingsInput.DhcpType.ValueString())
			inputUpdateNetworkRange.DhcpSettings.IPRange = dhcpSettingsInput.IpRange.ValueStringPointer()

			// Validate that relay_group_name or relay_group_id are only set when dhcp_type is DHCP_RELAY
			if dhcpSettingsInput.DhcpType.ValueString() != "DHCP_RELAY" {
				hasRelayGroupName := !dhcpSettingsInput.RelayGroupName.IsNull() && !dhcpSettingsInput.RelayGroupName.IsUnknown() && dhcpSettingsInput.RelayGroupName.ValueString() != ""
				hasRelayGroupId := !dhcpSettingsInput.RelayGroupId.IsNull() && !dhcpSettingsInput.RelayGroupId.IsUnknown() && dhcpSettingsInput.RelayGroupId.ValueString() != ""
				if hasRelayGroupName || hasRelayGroupId {
					resp.Diagnostics.AddError(
						"Invalid DHCP Relay Configuration",
						fmt.Sprintf("relay_group_name and relay_group_id can only be configured when dhcp_type is DHCP_RELAY, but dhcp_type is %s", dhcpSettingsInput.DhcpType.ValueString()),
					)
					return
				}
			}

			// Validate that ip_range is only set when dhcp_type is DHCP_RANGE
			if dhcpSettingsInput.DhcpType.ValueString() != "DHCP_RANGE" {
				hasIpRange := !dhcpSettingsInput.IpRange.IsNull() && !dhcpSettingsInput.IpRange.IsUnknown() && dhcpSettingsInput.IpRange.ValueString() != ""
				if hasIpRange {
					resp.Diagnostics.AddError(
						"Invalid DHCP Range Configuration",
						fmt.Sprintf("ip_range can only be configured when dhcp_type is DHCP_RANGE, but dhcp_type is %s", dhcpSettingsInput.DhcpType.ValueString()),
					)
					return
				}
			}

			// Validate that dhcp_microsegmentation is only set to true when dhcp_type is DHCP_RANGE
			if !dhcpSettingsInput.DhcpMicrosegmentation.IsNull() && !dhcpSettingsInput.DhcpMicrosegmentation.IsUnknown() {
				if dhcpSettingsInput.DhcpMicrosegmentation.ValueBool() == true && dhcpSettingsInput.DhcpType.ValueString() != "DHCP_RANGE" {
					resp.Diagnostics.AddError(
						"Invalid DHCP Microsegmentation Configuration",
						"dhcp_microsegmentation can only be configured when dhcp_type is set to DHCP_RANGE",
					)
					return
				}
			}

			// Only set dhcpMicrosegmentation for DHCP_RANGE type
			if dhcpSettingsInput.DhcpType.ValueString() == "DHCP_RANGE" {
				if !dhcpSettingsInput.DhcpMicrosegmentation.IsNull() && !dhcpSettingsInput.DhcpMicrosegmentation.IsUnknown() {
					inputUpdateNetworkRange.DhcpSettings.DhcpMicrosegmentation = dhcpSettingsInput.DhcpMicrosegmentation.ValueBoolPointer()
				}
			}

			// Validate DHCP relay group configuration when dhcp_type is DHCP_RELAY
			if dhcpSettingsInput.DhcpType.ValueString() == "DHCP_RELAY" {
				relayGroupName := ""
				relayGroupId := ""

				if !dhcpSettingsInput.RelayGroupName.IsNull() && !dhcpSettingsInput.RelayGroupName.IsUnknown() {
					relayGroupName = dhcpSettingsInput.RelayGroupName.ValueString()
				}
				if !dhcpSettingsInput.RelayGroupId.IsNull() && !dhcpSettingsInput.RelayGroupId.IsUnknown() {
					relayGroupId = dhcpSettingsInput.RelayGroupId.ValueString()
				}

				resolvedRelayGroupId, success, err := checkForDhcpRelayGroup(ctx, r.client, relayGroupName, relayGroupId)
				if err != nil {
					resp.Diagnostics.AddError(
						"DHCP Relay Configuration Error",
						err.Error(),
					)
					return
				}
				if !success {
					resp.Diagnostics.AddError(
						"DHCP Relay Group Validation Failed",
						"Failed to validate DHCP relay group configuration.",
					)
					return
				}
				// Set the resolved relay group ID
				inputUpdateNetworkRange.DhcpSettings.RelayGroupID = &resolvedRelayGroupId
			}
		} else {
			// Configuration has no dhcp_settings block - preserve dhcp_microsegmentation from state if it exists
			if !state.NativeRange.IsNull() && !state.NativeRange.IsUnknown() {
				var stateNativeRange NativeRange
				diags = state.NativeRange.As(ctx, &stateNativeRange, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				if !stateNativeRange.DhcpSettings.IsNull() && !stateNativeRange.DhcpSettings.IsUnknown() {
					var stateDhcpSettings DhcpSettings
					diags = stateNativeRange.DhcpSettings.As(ctx, &stateDhcpSettings, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// Only preserve dhcp_microsegmentation, don't send other DHCP settings to API
					if !stateDhcpSettings.DhcpMicrosegmentation.IsNull() && !stateDhcpSettings.DhcpMicrosegmentation.IsUnknown() {
						// We don't actually want to send DHCP settings to the API when config omits them
						// The preservation will happen during state hydration
						tflog.Debug(ctx, "Preserving dhcp_microsegmentation from state during update", map[string]interface{}{
							"dhcp_microsegmentation": stateDhcpSettings.DhcpMicrosegmentation.ValueBool(),
						})
					}
				}
			}
		}
	}

	// setting input other attributes
	inputSiteGeneral.Name = plan.Name.ValueStringPointer()
	inputSiteGeneral.SiteType = (*cato_models.SiteType)(plan.SiteType.ValueStringPointer())
	// Subnet empty string due to bug in API where value does not clear when null
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		emptyDesc := ""
		inputSiteGeneral.Description = &emptyDesc
	} else {
		inputSiteGeneral.Description = plan.Description.ValueStringPointer()
	}

	tflog.Debug(ctx, "Update.SiteUpdateSiteGeneralDetails.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputSiteGeneral),
	})
	siteUpdateSiteGeneralDetailsResponse, err := r.client.catov2.SiteUpdateSiteGeneralDetails(ctx, plan.Id.ValueString(), inputSiteGeneral, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteUpdateSiteGeneralDetails.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateSiteGeneralDetailsResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateSiteGeneralDetails error",
			err.Error(),
		)
		return
	}

	//retrieve native range ID
	nativeRange := NativeRange{}
	diags = state.NativeRange.As(ctx, &nativeRange, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.request", map[string]interface{}{
		"request":                          utils.InterfaceToJSONString(inputUpdateNetworkRange),
		"nativeRange.NativeNetworkRangeId": utils.InterfaceToJSONString(nativeRange.NativeNetworkRangeId.ValueString()),
		"r.client.AccountId":               utils.InterfaceToJSONString(r.client.AccountId),
	})
	siteUpdateNetworkRangeResponse, err := r.client.catov2.SiteUpdateNetworkRange(ctx, nativeRange.NativeNetworkRangeId.ValueString(), inputUpdateNetworkRange, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateNetworkRangeResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateNetworkRange error",
			err.Error(),
		)
		return
	}

	// Added section to handle isDefault where socket site is not able to define custom index for lan
	// if the interface identified as isDefault=true is different from the plan.NativeRange.InterfaceIndex
	// Move the interface by creating a secondary lan interface disabling the default after
	// retrieving native-network range ID to update native range
	// Lookup default lan index by socket connection type
	if !plan.NativeRange.IsNull() && !plan.NativeRange.IsUnknown() {
		nativeRangeCheck := NativeRange{}
		diags = plan.NativeRange.As(ctx, &nativeRangeCheck, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// // Lookup default lan index by socket connection type
		// var defaultLanInterfaceIndex *string
		// if val, ok := interfaceByConnType[plan.ConnectionType.ValueString()]; ok {
		// 	defaultLanInterfaceIndex = &val
		// }

		// // If the interface index is not specified, reset to default.
		// if nativeRangeCheck.InterfaceIndex.IsNull() && nativeRangeCheck.InterfaceIndex.IsUnknown() {
		// 	nativeRangeCheck.InterfaceIndex = types.StringValue(*defaultLanInterfaceIndex)
		// }

		tflog.Info(ctx, "Update: interface_index", map[string]interface{}{
			"state_value": nativeRangeState.InterfaceIndex.ValueString(),
			"plan_value":  nativeRangeCheck.InterfaceIndex.ValueString(),
		})

		// Check if interface_index changed from state to plan
		interfaceIndexChanged := false
		if !nativeRangeState.InterfaceIndex.Equal(nativeRangeCheck.InterfaceIndex) {
			interfaceIndexChanged = true
			tflog.Info(ctx, "Update: interface_index changed", map[string]interface{}{
				"state_value": nativeRangeState.InterfaceIndex.ValueString(),
				"plan_value":  nativeRangeCheck.InterfaceIndex.ValueString(),
			})
		}

		// Only perform interface reassignment if interface_index actually changed
		if interfaceIndexChanged {
			// Interface index changed - need to reassign it via create/disable/update sequence
			// Move the default interface by creating a new interface, disabling the original
			// then updating the new interface with the desired range
			//cato_models.SocketInterfaceIDEnum
			interfaceIndex := nativeRangeCheck.InterfaceIndex.ValueString()
			interfaceName := nativeRangeCheck.InterfaceName.ValueString()
			if interfaceName == "" {
				interfaceName = interfaceIndex
			}
			localIp := nativeRangeCheck.LocalIp.ValueString()
			nativeNetworkRange := nativeRangeCheck.NativeNetworkRange.ValueString()
			interfaceDestType := nativeRangeCheck.InterfaceDestType.ValueString()
			lagMinLinks := nativeRangeCheck.LagMinLinks.ValueInt64()
			translatedSubnet := ""
			if !nativeRangeCheck.TranslatedSubnet.IsNull() && !nativeRangeCheck.TranslatedSubnet.IsUnknown() {
				translatedSubnet = nativeRangeCheck.TranslatedSubnet.ValueString()
			}
			tflog.Debug(ctx, "Update.nativeRangeCheck", map[string]interface{}{
				"interfaceIndex":     utils.InterfaceToJSONString(interfaceIndex),
				"interfaceName":      utils.InterfaceToJSONString(interfaceName),
				"localIp":            utils.InterfaceToJSONString(localIp),
				"nativeNetworkRange": utils.InterfaceToJSONString(nativeNetworkRange),
				"interfaceDestType":  utils.InterfaceToJSONString(interfaceDestType),
				"lagMinLinks":        utils.InterfaceToJSONString(lagMinLinks),
				"translatedSubnet":   utils.InterfaceToJSONString(translatedSubnet),
			})
			isDefaultPresent, err := r.attemptReassignNativeRangeIndex(ctx, cato_models.SocketInterfaceIDEnum(interfaceIndex), interfaceName, localIp, nativeNetworkRange, plan.Id.ValueString(), interfaceDestType, lagMinLinks, translatedSubnet)
			if err != nil {
				resp.Diagnostics.AddError(
					"Catov2 API SiteAddSocketSite error",
					err.Error(),
				)
				return
			}
			if isDefaultPresent == nil || !*isDefaultPresent {
				resp.Diagnostics.AddError(
					"Configuration Error",
					"Invalid interface configuration, the API to support moving the lan interface index to '"+interfaceIndex+"' from the previous index is not yet supported.  This API is being gradually rolled out.",
				)
				return
			}
			// Interface was reassigned - need to retrieve new native range ID and reapply DHCP settings
			// Get the updated native interface and subnet information
			nativeInterfaceAndSubnet, err := r.getNativeInterfaceAndSubnet(ctx, plan.ConnectionType.ValueString(), plan.Id.ValueString(), plan, interfaceByConnType)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error retrieving native interface after reassignment",
					err.Error(),
				)
				return
			}

			// Re-apply DHCP settings to the new interface
			// The interface reassignment resets DHCP to API defaults, so we must reapply config
			tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.request (after interface reassignment)", map[string]interface{}{
				"request":              utils.InterfaceToJSONString(inputUpdateNetworkRange),
				"nativeNetworkRangeId": utils.InterfaceToJSONString(nativeInterfaceAndSubnet.NativeNetworkRangeId),
				"r.client.AccountId":   utils.InterfaceToJSONString(r.client.AccountId),
			})
			siteUpdateNetworkRangeResponseAfter, err := r.client.catov2.SiteUpdateNetworkRange(ctx, nativeInterfaceAndSubnet.NativeNetworkRangeId, inputUpdateNetworkRange, r.client.AccountId)
			tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.response (after interface reassignment)", map[string]interface{}{
				"response": utils.InterfaceToJSONString(siteUpdateNetworkRangeResponseAfter),
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Catov2 API SiteUpdateNetworkRange error (after interface reassignment)",
					err.Error(),
				)
				return
			}

			// Clear computed IDs from plan so they can be refreshed during hydration
			var nativeRangePlan NativeRange
			diags := plan.NativeRange.As(ctx, &nativeRangePlan, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			nativeRangePlan.InterfaceId = types.StringNull()
			nativeRangePlan.NativeNetworkLanInterfaceId = types.StringNull()
			nativeRangePlan.NativeNetworkRangeId = types.StringNull()
			plan.NativeRange, diags = types.ObjectValueFrom(ctx, plan.NativeRange.AttributeTypes(ctx), nativeRangePlan)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			// interface_index didn't change, just update the interface normally
			tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.request (no index change)", map[string]interface{}{
				"request": utils.InterfaceToJSONString(inputUpdateSocketInterface),
			})
			_, err = r.client.catov2.SiteUpdateSocketInterface(ctx, plan.Id.ValueString(), cato_models.SocketInterfaceIDEnum(nativeRangeState.InterfaceIndex.ValueString()), inputUpdateSocketInterface, r.client.AccountId)
			if err != nil {
				resp.Diagnostics.AddError(
					"Catov2 API SiteUpdateSocketInterface error",
					err.Error(),
				)
				return
			}
		}
	}

	// if err != nil {
	// 	var apiError struct {
	// 		NetworkErrors interface{} `json:"networkErrors"`
	// 		GraphqlErrors []struct {
	// 			Message string   `json:"message"`
	// 			Path    []string `json:"path"`
	// 		} `json:"graphqlErrors"`
	// 	}
	// 	interfaceNotPresent := false
	// 	if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphqlErrors) > 0 {
	// 		msg := apiError.GraphqlErrors[0].Message
	// 		if strings.Contains(msg, "Network range with id: ") && strings.Contains(msg, "is not found") {
	// 			interfaceNotPresent = true
	// 		}
	// 	}
	// 	if !interfaceNotPresent {
	// 		resp.Diagnostics.AddError(
	// 			"Catov2 API error",
	// 			err.Error(),
	// 		)
	// 		return
	// 	}
	// 	// If the network range is not present, delete the resource and recreate it
	// 	tflog.Warn(ctx, "Network range not found during update, recreating resource")
	// 	// Remove the resource from state
	// 	resp.State.RemoveResource(ctx)
	// 	return
	// }

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, plan, plan.Id.ValueString())
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating socket site state",
			hydrateErr.Error(),
		)
		return
	}

	// check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *socketSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.Id.ValueString()}, nil, nil, nil)
	tflog.Debug(ctx, "Create.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// check if site exist before removing
	if len(querySiteResult.EntityLookup.GetItems()) == 1 {

		_, err := r.client.catov2.SiteRemoveSite(ctx, state.Id.ValueString(), r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteRemoveSite error",
				err.Error(),
			)
			return
		}
	}
}

// calculateLocalIP calculates the local IP based on connection type and subnet
// For SOCKET_GCP1500, VSOCKET_VGX_AWS, VSOCKET_VGX_AZURE: use 4th IP (.4)
// For all others: use first available IP (.1)
func calculateLocalIP(ctx context.Context, subnet, connType string) string {
	if subnet == "" {
		return ""
	}

	// Parse the CIDR
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return ""
	}

	// Get the network address
	networkIP := ipNet.IP
	if networkIP == nil {
		return ""
	}

	// Convert to 4-byte representation
	ip := networkIP.To4()
	if ip == nil {
		return ""
	}

	// Determine the offset based on connection type
	var offset int
	switch connType {
	case "SOCKET_GCP1500", "SOCKET_AWS1500", "SOCKET_AZ1500":
		offset = 4 // Use 5th IP (.4)
	default:
		offset = 1 // Use first available IP (.1)
	}
	tflog.Warn(ctx, "calculateLocalIP.connType", map[string]interface{}{
		"connType": utils.InterfaceToJSONString(connType),
		"offset":   utils.InterfaceToJSONString(offset),
	})

	// Calculate the local IP by adding the offset to the network address
	localIP := make(net.IP, 4)
	copy(localIP, ip)
	localIP[3] += byte(offset)

	return localIP.String()
}

// NativeInterfaceAndSubnetResult contains all the data returned by getNativeInterfaceAndSubnet
type NativeInterfaceAndSubnetResult struct {
	Subnet               string
	NativeNetworkRangeId string
	InterfaceIndex       string
	InterfaceId          string
	InterfaceName        string
	SiteNetRangeApiData  map[string]any
	NativeRangeObj       NativeRange
}

// getNativeInterfaceAndSubnet retrieves native interface and subnet information
// Returns: NativeInterfaceAndSubnetResult, error
func (r *socketSiteResource) getNativeInterfaceAndSubnet(ctx context.Context, connType string, siteID string, state SocketSite, interfaceByConnType map[string]string) (*NativeInterfaceAndSubnetResult, error) {
	siteEntity := &cato_models.EntityInput{Type: "site", ID: siteID}
	zeroInt64 := int64(0)
	if connType != "" {
		// Translate VSOCKET_VGX_* values to SOCKET_* equivalents
		switch connType {
		case "VSOCKET_VGX_AWS":
			connType = "SOCKET_AWS1500"
		case "VSOCKET_VGX_AZURE":
			connType = "SOCKET_AZ1500"
		case "VSOCKET_VGX_ESX":
			connType = "SOCKET_ESX1500"
		}
	} else {
		return nil, fmt.Errorf("connection type is empty")
	}

	// Only assign interfaceIndex if it does not already exist in state
	if _, ok := interfaceByConnType[connType]; !ok {
		return nil, fmt.Errorf("connection type %s not found in interfaceByConnType", connType)
	}
	var nativeRangeObj NativeRange
	if !state.NativeRange.IsNull() && !state.NativeRange.IsUnknown() {
		state.NativeRange.As(ctx, &nativeRangeObj, basetypes.ObjectAsOptions{})
	}
	// if nativeRangeObj.InterfaceIndex.IsNull() || nativeRangeObj.InterfaceIndex.ValueString() == "" {
	defaultInterfaceIndexByConnType, ok := interfaceByConnType[connType]
	if !ok {
		return nil, fmt.Errorf("connection type %s not found in interfaceByConnType", connType)
	}
	// Retrieve default interface range attributes
	queryInterfaceResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), &zeroInt64, nil, siteEntity, nil, nil, nil, nil, nil)
	tflog.Warn(ctx, "Read.EntityLookupInterfaceRangeResult.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryInterfaceResult),
	})
	if err != nil {
		return nil, err
	}
	isPresent := false
	// Check for isDefault flag to be present first
	for _, curIint := range queryInterfaceResult.EntityLookup.Items {
		curSiteId := cast.ToString(curIint.HelperFields["siteId"])
		if curSiteId == siteID {
			curInterfaceId := curIint.HelperFields["interfaceId"]
			// curInterfaceName := curIint.HelperFields["interfaceName"]
			// Try to parse the interfaceId as int, otherwise prefix with "INT_"
			if idxInt, err := cast.ToIntE(curInterfaceId); err == nil {
				curInterfaceIdStr := fmt.Sprintf("INT_%d", idxInt)
				curInterfaceId = curInterfaceIdStr
			}
			isDefault := false
			if v, ok := curIint.HelperFields["isDefault"]; ok && v != nil {
				if b, err := cast.ToBoolE(v); err == nil {
					isDefault = b
				}
			}
			if isDefault {
				isPresent = true
				nativeRangeObj.InterfaceIndex = types.StringValue(cast.ToString(curInterfaceId))
				nativeRangeObj.InterfaceId = types.StringValue(curIint.Entity.ID)
				nativeRangeObj.InterfaceName = types.StringValue(curIint.HelperFields["interfaceName"].(string))
				nativeRangeObj.NativeNetworkRange = types.StringValue(curIint.HelperFields["subnet"].(string))
				if destType, ok := curIint.HelperFields["destType"]; ok && destType != nil {
					nativeRangeObj.InterfaceDestType = types.StringValue(cast.ToString(destType))
				} else {
					nativeRangeObj.InterfaceDestType = types.StringNull()
				}
			}
		}
	}
	// If isDefault flag not found, look for default by index
	// This is due to bug/fix getting gradually pushed out where this default flag may not be present
	// and the following can be purged after this is rolled out
	if !isPresent {
		for _, curIint := range queryInterfaceResult.EntityLookup.Items {
			// find the socket site entry we need
			curSiteId := cast.ToString(curIint.HelperFields["siteId"])
			tflog.Warn(ctx, "for.queryInterfaceResult.EntityLookup.Items | siteID==siteID", map[string]interface{}{
				"siteID":                          siteID,
				"curSiteId":                       curSiteId,
				"defaultInterfaceIndexByConnType": defaultInterfaceIndexByConnType,
				"curInterfaceId":                  curIint.HelperFields["interfaceId"],
				"curInterfaceName":                curIint.HelperFields["interfaceName"],
			})
			if curSiteId == siteID {
				// get current interfaceId from the API and use to map to interface index
				curInterfaceId := curIint.HelperFields["interfaceId"]
				curInterfaceName := curIint.HelperFields["interfaceName"]
				// Try to parse the interfaceId as int, otherwise prefix with "INT_"
				if idxInt, err := cast.ToIntE(curInterfaceId); err == nil {
					curInterfaceIdStr := fmt.Sprintf("INT_%d", idxInt)
					curInterfaceId = curInterfaceIdStr
				}
				tflog.Warn(ctx, "defaultInterfaceIndexByConnType==curInterfaceId", map[string]interface{}{
					"defaultInterfaceIndexByConnType": cast.ToString(defaultInterfaceIndexByConnType),
					"curInterfaceId":                  curInterfaceId,
					"curInterfaceName":                curInterfaceName,
				})
				if cast.ToString(defaultInterfaceIndexByConnType) == curInterfaceId {
					isPresent = true
					if _, err := cast.ToIntE(curInterfaceId); err == nil {
						nativeRangeObj.InterfaceIndex = types.StringValue(cast.ToString(curInterfaceId))
					} else {
						nativeRangeObj.InterfaceIndex = types.StringValue(cast.ToString(curInterfaceId))
					}
					nativeRangeObj.InterfaceIndex = types.StringValue(defaultInterfaceIndexByConnType)
					nativeRangeObj.InterfaceId = types.StringValue(curIint.Entity.ID)
					nativeRangeObj.InterfaceName = types.StringValue(curIint.HelperFields["interfaceName"].(string))
					nativeRangeObj.NativeNetworkRange = types.StringValue(curIint.HelperFields["subnet"].(string))
					if destType, ok := curIint.HelperFields["destType"]; ok && destType != nil {
						nativeRangeObj.InterfaceDestType = types.StringValue(cast.ToString(destType))
					} else {
						nativeRangeObj.InterfaceDestType = types.StringNull()
					}
				} else {
					tflog.Warn(ctx, "Skipping interface by connection type", map[string]interface{}{
						"defaultInterfaceIndexByConnType": defaultInterfaceIndexByConnType,
						"curInterfaceId":                  curInterfaceId,
						"curInterfaceName":                curInterfaceName,
					})
				}
			}
		}
	}
	if !isPresent {
		return nil, fmt.Errorf("Site does not contain configuration for default LAN interface index %s for connection type %s, please update this site configuratation once either in the cato management application or via API, and the correct interface should be marked as default resolving this issue.", defaultInterfaceIndexByConnType, connType)
	}

	// Retrieve default site range attributes
	querySiteRangeResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"), &zeroInt64, nil, siteEntity, nil, nil, nil, nil, nil)
	tflog.Warn(ctx, "Read.EntityLookupSiteRangeResult.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteRangeResult),
	})
	if err != nil {
		return nil, err
	}

	siteNetRangeApiData := make(map[string]any)
	for _, v := range querySiteRangeResult.GetEntityLookup().GetItems() {
		curSubnet := v.GetHelperFields()["subnet"].(string)
		if curSubnet == nativeRangeObj.NativeNetworkRange.ValueString() {
			if v.GetEntity() != nil && v.GetEntity().Name != nil {
				nameParts := strings.Split(*v.GetEntity().Name, " \\ ")
				siteNetRangeApiData["rangeName"] = nameParts[len(nameParts)-1] // Store as string, not types.StringValue
			}
			// siteNetRangeApiData = v.GetHelperFields()
			// Pull ID from entity attributes
			siteNetRangeApiData["native_network_range_id"] = v.Entity.GetID()
			if vlanTag, ok := v.HelperFields["vlanTag"]; ok && vlanTag != nil {
				if vlanInt, err := cast.ToInt64E(vlanTag); err == nil {
					nativeRangeObj.Vlan = types.Int64Value(vlanInt)
				} else {
					nativeRangeObj.Vlan = types.Int64Null()
				}
			} else {
				nativeRangeObj.Vlan = types.Int64Null()
			}
			// Set gateway from API
			if gatewayVal, ok := v.HelperFields["gateway"]; ok && gatewayVal != nil {
				nativeRangeObj.Gateway = types.StringValue(cast.ToString(gatewayVal))
			} else {
				nativeRangeObj.Gateway = types.StringNull()
			}
			// Set LocalIp to gateway value (they are the same in the API)
			nativeRangeObj.LocalIp = nativeRangeObj.Gateway
			// Set range_type from API
			if rangeTypeVal, ok := v.HelperFields["rangeType"]; ok && rangeTypeVal != nil {
				nativeRangeObj.RangeType = types.StringValue(cast.ToString(rangeTypeVal))
			} else {
				nativeRangeObj.RangeType = types.StringNull()
			}
			// Set mdns_reflector from API
			if mdnsReflectorVal, ok := v.HelperFields["mdnsReflector"]; ok {
				nativeRangeObj.MdnsReflector = types.BoolValue(cast.ToBool(mdnsReflectorVal))
			} else {
				nativeRangeObj.MdnsReflector = types.BoolValue(false)
			}
			// Set translated_subnet from API
			if translatedSubnetVal, ok := v.HelperFields["translatedSubnet"]; ok && translatedSubnetVal != nil {
				nativeRangeObj.TranslatedSubnet = types.StringValue(cast.ToString(translatedSubnetVal))
			} else {
				nativeRangeObj.TranslatedSubnet = types.StringNull()
			}
			// lag_min_links is not available in API - preserve from state if present
			// (already set when nativeRangeObj was extracted from state at the top of this function)

			// Hydrate DHCP settings ONLY if they were present in the state/plan
			// Check if dhcp_settings was in the original state/plan
			dhcpSettingsInState := !nativeRangeObj.DhcpSettings.IsNull() && !nativeRangeObj.DhcpSettings.IsUnknown()

			// Check if API returned a non-default dhcp type
			dhcpTypeVal, hasDhcpType := v.HelperFields["dhcpType"]
			hasNonDefaultDhcp := hasDhcpType && dhcpTypeVal != nil && cast.ToString(dhcpTypeVal) != "DHCP_DISABLED"

			// Only populate DHCP settings if they were already in state OR the API returns a non-default dhcp type
			// This prevents inconsistency when config doesn't include dhcp_settings and API returns default DHCP_DISABLED
			if dhcpSettingsInState || hasNonDefaultDhcp {
				// Hydrate DHCP settings - resolve unknown values to known values
				// Start with null values as defaults - unknown values MUST be resolved to known values
				dhcpType := types.StringNull()
				ipRange := types.StringNull()
				relayGroupId := types.StringNull()
				relayGroupName := types.StringNull()
				dhcpMicrosegmentation := types.BoolValue(false)

				// Hydrate from API values (API is source of truth)
				if dhcpTypeVal, ok := v.HelperFields["dhcpType"]; ok && dhcpTypeVal != nil {
					dhcpType = types.StringValue(cast.ToString(dhcpTypeVal))
				}

				// Only hydrate relay group fields if dhcp_type is DHCP_RELAY
				if dhcpType.ValueString() == "DHCP_RELAY" {
					if dhcpRelayGroupIdVal, ok := v.HelperFields["dhcpRelayGroupId"]; ok && dhcpRelayGroupIdVal != nil {
						relayGroupId = types.StringValue(cast.ToString(dhcpRelayGroupIdVal))
					}
					if dhcpRelayGroupNameVal, ok := v.HelperFields["dhcpRelayGroupName"]; ok && dhcpRelayGroupNameVal != nil {
						relayGroupName = types.StringValue(cast.ToString(dhcpRelayGroupNameVal))
					}
				} else {
					// If dhcp_type is not DHCP_RELAY, explicitly set relay fields to null
					// This prevents drift when config doesn't specify them but they exist in state
					relayGroupId = types.StringNull()
					relayGroupName = types.StringNull()
				}
				// Only hydrate ip_range if dhcp_type is DHCP_RANGE
				if dhcpType.ValueString() == "DHCP_RANGE" {
					if dhcpRangeVal, ok := v.HelperFields["dhcpRange"]; ok && dhcpRangeVal != nil {
						ipRange = types.StringValue(cast.ToString(dhcpRangeVal))
					}
					if microsegmentationVal, ok := v.HelperFields["microsegmentation"]; ok && microsegmentationVal != nil {
						dhcpMicrosegmentation = types.BoolValue(cast.ToBool(microsegmentationVal))
					}
				}

				// Manually construct Object to ensure no Unknown values persist
				dhcpSettingsObject, dErr := types.ObjectValue(
					SiteNativeRangeDhcpResourceAttrTypes,
					map[string]attr.Value{
						"dhcp_type":              dhcpType,
						"ip_range":               ipRange,
						"relay_group_id":         relayGroupId,
						"relay_group_name":       relayGroupName,
						"dhcp_microsegmentation": dhcpMicrosegmentation,
					},
				)
				if dErr.HasError() {
					return nil, fmt.Errorf("failed to convert dhcp settings to object: %v", dErr)
				}
				nativeRangeObj.DhcpSettings = dhcpSettingsObject
			} else {
				// dhcp_settings was not in state and API didn't return it - keep it null
				nativeRangeObj.DhcpSettings = types.ObjectNull(SiteNativeRangeDhcpResourceAttrTypes)
			}

			break
		}
	}

	tflog.Debug(ctx, "Read.siteNetRangeApiData", map[string]interface{}{
		"siteNetRangeApiData": utils.InterfaceToJSONString(siteNetRangeApiData),
	})

	nativeNetworkRangeId := ""
	if val, ok := siteNetRangeApiData["native_network_range_id"].(string); ok {
		nativeNetworkRangeId = val
	}

	interfaceIndex := ""
	if !nativeRangeObj.InterfaceIndex.IsNull() && !nativeRangeObj.InterfaceIndex.IsUnknown() {
		interfaceIndex = nativeRangeObj.InterfaceIndex.ValueString()
	}

	interfaceId := ""
	if !nativeRangeObj.InterfaceId.IsNull() && !nativeRangeObj.InterfaceId.IsUnknown() {
		interfaceId = nativeRangeObj.InterfaceId.ValueString()
	}

	return &NativeInterfaceAndSubnetResult{
		Subnet:               nativeRangeObj.NativeNetworkRange.ValueString(),
		NativeNetworkRangeId: nativeNetworkRangeId,
		InterfaceIndex:       interfaceIndex,
		InterfaceName:        nativeRangeObj.InterfaceName.ValueString(),
		InterfaceId:          interfaceId,
		SiteNetRangeApiData:  siteNetRangeApiData,
		NativeRangeObj:       nativeRangeObj,
	}, nil
}

func (r *socketSiteResource) attemptReassignNativeRangeIndex(ctx context.Context, interfaceIndex cato_models.SocketInterfaceIDEnum, name string, localIp string, subnet string, siteID string, interfaceDestType string, lagMinLinks int64, translatedSubnet string) (*bool, error) {
	// Attempt to create a lan interface on non default index
	// checking first for the isDefault flag to be present, if not present return false for isDefaultPresent
	// return isValid or error if isDefault flag is not present on entityLookup interface query
	isDefaultPresent := false
	siteEntity := &cato_models.EntityInput{Type: "site", ID: siteID}
	zeroInt64 := int64(0)
	tflog.Warn(ctx, "attemptReassignNativeRangeIndex.EntityLookupInterfaceRange", map[string]interface{}{
		"siteEntity": utils.InterfaceToJSONString(siteEntity),
	})
	queryInterfaceResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("networkInterface"), &zeroInt64, nil, siteEntity, nil, nil, nil, nil, nil)
	tflog.Warn(ctx, "attemptReassignNativeRangeIndex.EntityLookupInterfaceRange.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryInterfaceResult),
	})
	if err != nil {
		return nil, err
	}

	// Lookup current interface index from what is returned in entityLookup
	// Check for an interface already present on that desired index
	// If present update that interface with the native range config from the config
	// if there is no interface there, attempt to create and disable the default
	// curInterfaceId := nil
	var curInterfaceIndex *string
	for _, curIint := range queryInterfaceResult.EntityLookup.Items {
		tflog.Warn(ctx, "Read.EntityLookupInterfaceRangeResult.curIint.interator", map[string]interface{}{
			"interfaceIndex":    utils.InterfaceToJSONString(interfaceIndex),
			"curInterfaceId":    utils.InterfaceToJSONString(curIint.Entity.ID),
			"curInterfaceIndex": utils.InterfaceToJSONString(curIint.HelperFields["interfaceId"]),
		})
		if isDefaultVal, ok := curIint.HelperFields["isDefault"]; ok && isDefaultVal != nil {
			isDefault := false
			if b, err := cast.ToBoolE(isDefaultVal); err == nil {
				isDefault = b
			}
			if isDefault {
				curInterfaceIndexStr := cast.ToString(curIint.HelperFields["interfaceId"])
				if idxInt, err := cast.ToIntE(curInterfaceIndexStr); err == nil {
					curInterfaceIndexStr = fmt.Sprintf("INT_%d", idxInt)
				}
				curInterfaceIndex = &curInterfaceIndexStr
				tflog.Warn(ctx, "Read.EntityLookupInterfaceRangeResult.curIint.isDefault", map[string]interface{}{
					"isDefault":         utils.InterfaceToJSONString(isDefault),
					"interfaceIndex":    utils.InterfaceToJSONString(interfaceIndex),
					"curInterfaceId":    utils.InterfaceToJSONString(curIint.Entity.ID),
					"curInterfaceIndex": utils.InterfaceToJSONString(curInterfaceIndex),
				})
				isDefaultPresent = true
			}
		}
	}
	if isDefaultPresent {
		// Create placeholder interface
		tmpInputUpdateSocketInterface := cato_models.UpdateSocketInterfaceInput{}
		tmpName := name + "_tmp"
		tmpInputUpdateSocketInterface.Name = &tmpName
		tmpInputUpdateSocketInterface.DestType = cato_models.SocketInterfaceDestType("LAN")
		tmpSocketInterfaceLanInput := cato_models.SocketInterfaceLanInput{}
		tmpSocketInterfaceLanInput.LocalIP = "127.111.111.1"
		tmpSocketInterfaceLanInput.Subnet = "127.111.111.0/24"
		if translatedSubnet != "" {
			tmpSocketInterfaceLanInput.TranslatedSubnet = &translatedSubnet
		}
		tmpInputUpdateSocketInterface.Lan = &tmpSocketInterfaceLanInput
		tflog.Debug(ctx, "attemptReassignNativeRangeIndex.SiteUpdateSocketInterface.tmp.request", map[string]interface{}{
			"tmpRequest":        utils.InterfaceToJSONString(tmpInputUpdateSocketInterface),
			"tmpInterfaceIndex": utils.InterfaceToJSONString(interfaceIndex),
		})
		tmpSiteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, siteID, interfaceIndex, tmpInputUpdateSocketInterface, r.client.AccountId)
		if err != nil {
			return nil, err
		}
		tflog.Debug(ctx, "attemptReassignNativeRangeIndex.tmpSiteUpdateSocketInterface.tmp.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(tmpSiteUpdateSocketInterfaceResponse),
		})

		// Disable original native range interface
		// First check to see if exisitng default interface is configured as LAG, disable all members before disabling the interface directly
		if (interfaceDestType == "LAN_LAG_MASTER" || interfaceDestType == "LAN_LAG_MASTER_AND_VRRP") && lagMinLinks != 0 {
			// Get the site's accountSnapshot to find the LAG master
			siteAccountSnapshotApiData, err := r.client.catov2.AccountSnapshot(ctx, []string{siteID}, nil, &r.client.AccountId)
			tflog.Debug(ctx, "Create.AccountSnapshot.response looking for LAN_LAG_MEMBERs", map[string]interface{}{
				"response": utils.InterfaceToJSONString(siteAccountSnapshotApiData),
			})
			for _, site := range siteAccountSnapshotApiData.AccountSnapshot.GetSites() {
				siteSiteID := site.GetID()
				if siteSiteID != nil && *siteSiteID == siteID {
					for _, iface := range site.InfoSiteSnapshot.Interfaces {
						if iface.DestType != nil && *iface.DestType == "LAN_LAG_MEMBER" {
							tflog.Debug(ctx, "Create.AccountSnapshot.response found LAN_LAG_MEMBER", map[string]interface{}{
								"response": utils.InterfaceToJSONString(iface),
							})
							curInterfaceId := iface.ID
							if _, err := cast.ToIntE(curInterfaceId); err == nil {
								curInterfaceId = fmt.Sprintf("INT_%v", curInterfaceId)
							}
							input := cato_models.UpdateSocketInterfaceInput{
								Name:     &curInterfaceId,
								DestType: "INTERFACE_DISABLED",
							}

							tflog.Debug(ctx, "Delete.SiteUpdateSocketInterface.request LAN_LAG_MEMBER", map[string]interface{}{
								"request": utils.InterfaceToJSONString(input),
							})
							_, err := r.client.catov2.SiteUpdateSocketInterface(ctx, siteID, cato_models.SocketInterfaceIDEnum(curInterfaceId), input, r.client.AccountId)
							if err != nil {
								return nil, err
							}
						}
					}
				}
			}
			if err != nil {
				return nil, err
			}
		}

		disabledInputUpdateSocketInterface := cato_models.UpdateSocketInterfaceInput{
			Name:     curInterfaceIndex,
			DestType: "INTERFACE_DISABLED",
		}
		tflog.Debug(ctx, "attemptReassignNativeRangeIndex.DisableDefaultInterface.request", map[string]interface{}{
			"disabledInputUpdateSocketInterface": utils.InterfaceToJSONString(disabledInputUpdateSocketInterface),
		})
		siteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, siteID, cato_models.SocketInterfaceIDEnum(*curInterfaceIndex), disabledInputUpdateSocketInterface, r.client.AccountId)
		if err != nil {
			return nil, err
		}
		tflog.Debug(ctx, "attemptReassignNativeRangeIndex.DisableDefaultInterface.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(siteUpdateSocketInterfaceResponse),
		})

		// Update new lan interface with the correct native range subnet
		curInputUpdateSocketInterface := cato_models.UpdateSocketInterfaceInput{}
		curInputUpdateSocketInterface.Name = &name
		curInputUpdateSocketInterface.DestType = cato_models.SocketInterfaceDestType(interfaceDestType)
		if (interfaceDestType == "LAN_LAG_MASTER" || interfaceDestType == "LAN_LAG_MASTER_AND_VRRP") && lagMinLinks != 0 {
			tmpLagConfig := cato_models.SocketInterfaceLagInput{
				MinLinks: lagMinLinks,
			}
			curInputUpdateSocketInterface.Lag = &tmpLagConfig
		}
		curSocketInterfaceLanInput := cato_models.SocketInterfaceLanInput{}
		curSocketInterfaceLanInput.LocalIP = localIp
		curSocketInterfaceLanInput.Subnet = subnet
		if translatedSubnet != "" {
			curSocketInterfaceLanInput.TranslatedSubnet = &translatedSubnet
		}
		tflog.Debug(ctx, "attemptReassignNativeRangeIndex.SiteUpdateSocketInterface.current.request", map[string]interface{}{
			"tmpRequest":        utils.InterfaceToJSONString(curInputUpdateSocketInterface),
			"tmpInterfaceIndex": utils.InterfaceToJSONString(interfaceIndex),
		})
		curInputUpdateSocketInterface.Lan = &curSocketInterfaceLanInput
		curSiteUpdateSocketInterfaceResponse, err := r.client.catov2.SiteUpdateSocketInterface(ctx, siteID, interfaceIndex, curInputUpdateSocketInterface, r.client.AccountId)
		if err != nil {
			return nil, err
		}
		tflog.Debug(ctx, "attemptReassignNativeRangeIndex.tmpSiteUpdateSocketInterface.current.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(curSiteUpdateSocketInterfaceResponse),
		})
	}
	return &isDefaultPresent, err
}

// hydrateSocketSiteState populates the SocketSite state with data from API responses
func (r *socketSiteResource) hydrateSocketSiteState(ctx context.Context, state SocketSite, siteID string) (SocketSite, bool, error) {
	// Unmarshal socketMapping into a map
	var interfaceByConnType map[string]string
	err := json.Unmarshal([]byte(defaultInterfaceByConnType), &interfaceByConnType)
	if err != nil {
		panic(err)
	}

	// check if site exist, else remove resource
	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{siteID}, nil, nil, nil)
	tflog.Warn(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		return state, false, err
	}

	// // Get native interface and subnet information (remove unused variable)
	// _, err = r.getNativeInterfaceAndSubnet(ctx, state.ConnectionType.ValueString(), siteID, state, interfaceByConnType)
	// if err != nil {
	// 	return state, false, err
	// }

	siteAccountSnapshotApiData, err := r.client.catov2.AccountSnapshot(ctx, []string{siteID}, nil, &r.client.AccountId)
	tflog.Warn(ctx, "Read.AccountSnapshot/siteAccountSnapshotApiData.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteAccountSnapshotApiData),
	})
	if err != nil {
		return state, false, err
	}

	// check if site exist before refreshing
	// we should only have one entry since we are filtering on site ID
	if len(querySiteResult.EntityLookup.GetItems()) != 1 {
		return state, false, nil
	}
	for _, v := range querySiteResult.EntityLookup.Items {
		// find the socket site entry we need
		if v.Entity.ID == siteID {
			var stateSiteLocation types.Object
			// Find the correct site in the AccountSnapshot by matching siteID
			var thisSiteAccountSnapshot *cato_go_sdk.AccountSnapshot_AccountSnapshot_Sites
			for _, site := range siteAccountSnapshotApiData.GetAccountSnapshot().GetSites() {
				if site.ID != nil && *site.ID == siteID {
					thisSiteAccountSnapshot = site
					break
				}
			}

			if thisSiteAccountSnapshot != nil {
				// Get connection type and set state
				connTypeVal := ""
				if val := thisSiteAccountSnapshot.InfoSiteSnapshot.GetConnType(); val != nil {
					connTypeVal = val.String()
				}

				tflog.Debug(ctx, "Read.connTypeVal", map[string]interface{}{
					"connTypeVal":             connTypeVal,
					"thisSiteConnType":        thisSiteAccountSnapshot.InfoSiteSnapshot.GetConnType(),
					"thisSiteAccountSnapshot": utils.InterfaceToJSONString(thisSiteAccountSnapshot),
				})

				if connTypeVal != "" {
					// Translate VSOCKET_VGX_* values to SOCKET_* equivalents
					switch connTypeVal {
					case "VSOCKET_VGX_AWS":
						connTypeVal = "SOCKET_AWS1500"
					case "VSOCKET_VGX_AZURE":
						connTypeVal = "SOCKET_AZ1500"
					case "VSOCKET_VGX_ESX":
						connTypeVal = "SOCKET_ESX1500"
					}
					state.ConnectionType = types.StringValue(connTypeVal)
				} else {
					state.ConnectionType = types.StringNull()
				}

				// Get native interface and subnet information
				nativeInterfaceAndSubnet, err := r.getNativeInterfaceAndSubnet(ctx, connTypeVal, siteID, state, interfaceByConnType)
				if err != nil {
					return state, false, err
				}

				tflog.Debug(ctx, "Read.getNativeInterfaceAndSubnet.result", map[string]interface{}{
					"result": utils.InterfaceToJSONString(nativeInterfaceAndSubnet),
				})

				// Extract values from result struct
				subnet := nativeInterfaceAndSubnet.Subnet
				resultNativeNetworkRangeId := nativeInterfaceAndSubnet.NativeNetworkRangeId
				// interfaceIndex := nativeInterfaceAndSubnet.InterfaceIndex
				// interfaceId := nativeInterfaceAndSubnet.InterfaceId
				siteNetRangeApiData := nativeInterfaceAndSubnet.SiteNetRangeApiData
				nativeRangeObj := nativeInterfaceAndSubnet.NativeRangeObj

				siteType := ""
				if val, containsKey := v.GetHelperFields()["type"]; containsKey {
					siteType = val.(string)
				}
				state.Id = types.StringValue(v.Entity.GetID())
				state.Name = types.StringValue(*v.GetEntity().Name)
				// ConnectionType is already set above in the switch statement
				state.SiteType = types.StringValue(siteType)
				descriptionStr := v.GetHelperFields()["description"].(string)
				if descriptionStr != "" {
					state.Description = types.StringValue(descriptionStr)
				}

				// Determine the translated_subnet value to use in state
				// If plan has translated_subnet different from native_network_range, use plan value
				// If translated_subnet equals native_network_range, return null
				// Otherwise, use API value
				translatedSubnetValue := nativeRangeObj.TranslatedSubnet
				if !state.NativeRange.IsNull() && !state.NativeRange.IsUnknown() {
					var planNativeRange NativeRange
					diags := state.NativeRange.As(ctx, &planNativeRange, basetypes.ObjectAsOptions{})
					if diags == nil || !diags.HasError() {
						// Check if plan has translated_subnet
						if !planNativeRange.TranslatedSubnet.IsNull() && !planNativeRange.TranslatedSubnet.IsUnknown() &&
							!planNativeRange.NativeNetworkRange.IsNull() && !planNativeRange.NativeNetworkRange.IsUnknown() {
							planTranslated := planNativeRange.TranslatedSubnet.ValueString()
							planNative := planNativeRange.NativeNetworkRange.ValueString()
							if planTranslated != "" && planTranslated != planNative {
								// Plan has a different translated_subnet - preserve it
								translatedSubnetValue = types.StringValue(planTranslated)
								tflog.Debug(ctx, "Preserving plan translated_subnet value", map[string]interface{}{
									"plan_translated":   planTranslated,
									"plan_native":       planNative,
									"api_translated":    nativeRangeObj.TranslatedSubnet.ValueString(),
									"api_native_subnet": subnet,
								})
							} else if planTranslated == planNative {
								// translated_subnet equals native_network_range - return null
								translatedSubnetValue = types.StringNull()
								tflog.Debug(ctx, "translated_subnet equals native_network_range, returning null", map[string]interface{}{
									"plan_translated": planTranslated,
									"plan_native":     planNative,
								})
							}
						}
					}
				}
				// Additionally check if API returned translated_subnet equals the native subnet
				if !translatedSubnetValue.IsNull() && !translatedSubnetValue.IsUnknown() &&
					translatedSubnetValue.ValueString() == subnet {
					// API translated_subnet equals native subnet - return null
					translatedSubnetValue = types.StringNull()
					tflog.Debug(ctx, "API translated_subnet equals native subnet, returning null", map[string]interface{}{
						"api_translated": translatedSubnetValue.ValueString(),
						"native_subnet":  subnet,
					})
				}

				// All values from API via nativeRangeObj
				var stateNativeRange types.Object
				stateNativeRange, _ = types.ObjectValue(
					SiteNativeRangeResourceAttrTypes,
					map[string]attr.Value{
						"interface_index":                 nativeRangeObj.InterfaceIndex,
						"interface_id":                    nativeRangeObj.InterfaceId,
						"interface_name":                  nativeRangeObj.InterfaceName,
						"native_network_lan_interface_id": nativeRangeObj.InterfaceId,
						"native_network_range":            types.StringValue(subnet),
						"native_network_range_id":         types.StringValue(resultNativeNetworkRangeId),
						"range_name": func() attr.Value {
							if val, ok := siteNetRangeApiData["rangeName"].(string); ok && val != "" {
								return types.StringValue(val)
							}
							return types.StringValue("Native Range")
						}(),
						"range_id": func() attr.Value {
							if val, ok := siteNetRangeApiData["rangeId"].(string); ok && val != "" {
								return types.StringValue(val)
							}
							return types.StringNull()
						}(),
						"local_ip":            nativeRangeObj.LocalIp,
						"translated_subnet":   translatedSubnetValue,
						"gateway":             nativeRangeObj.Gateway,
						"range_type":          nativeRangeObj.RangeType,
						"vlan":                nativeRangeObj.Vlan,
						"mdns_reflector":      nativeRangeObj.MdnsReflector,
						"lag_min_links":       nativeRangeObj.LagMinLinks,
						"interface_dest_type": nativeRangeObj.InterfaceDestType,
						"dhcp_settings":       nativeRangeObj.DhcpSettings,
					},
				)
				state.NativeRange = stateNativeRange

				// Extract location data from API response
				countryName := ""
				if thisSiteAccountSnapshot.InfoSiteSnapshot.CountryName != nil {
					countryName = *thisSiteAccountSnapshot.InfoSiteSnapshot.CountryName
				}
				stateName := ""
				if thisSiteAccountSnapshot.InfoSiteSnapshot.CountryStateName != nil {
					stateName = *thisSiteAccountSnapshot.InfoSiteSnapshot.CountryStateName
				}
				cityName := ""
				if thisSiteAccountSnapshot.InfoSiteSnapshot.CityName != nil {
					cityName = *thisSiteAccountSnapshot.InfoSiteSnapshot.CityName
				}

				// Resolve location data using the new function
				resolvedLocation := populateSiteLocationData(countryName, stateName, cityName)

				tflog.Debug(ctx, "Read.populateSiteLocationData.input", map[string]interface{}{
					"countryName": countryName,
					"stateName":   stateName,
					"cityName":    cityName,
				})

				tflog.Debug(ctx, "Read.populateSiteLocationData.result", map[string]interface{}{
					"resolvedLocation": utils.InterfaceToJSONString(resolvedLocation),
				})

				// Resolve timezone locally from location data
				timezoneValue := types.StringNull()
				if resolvedLocation.Timezone != "" {
					timezoneValue = types.StringValue(resolvedLocation.Timezone)
				}

				// Resolve state_code locally from location data
				stateCodeValue := types.StringNull()
				if resolvedLocation.StateCode != "" {
					stateCodeValue = types.StringValue(resolvedLocation.StateCode)
				}

				// All location values from API, timezone and state_code resolved locally
				stateSiteLocation, _ = types.ObjectValue(
					SiteLocationResourceAttrTypes,
					map[string]attr.Value{
						"country_code": types.StringValue(*thisSiteAccountSnapshot.GetInfoSiteSnapshot().CountryCode),
						"state_code":   stateCodeValue,
						"timezone":     timezoneValue,
						"address": func() types.String {
							if thisSiteAccountSnapshot.InfoSiteSnapshot.Address != nil && *thisSiteAccountSnapshot.InfoSiteSnapshot.Address != "" {
								return types.StringValue(*thisSiteAccountSnapshot.InfoSiteSnapshot.Address)
							}
							return types.StringNull()
						}(),
						"city": func() types.String {
							if thisSiteAccountSnapshot.InfoSiteSnapshot.CityName != nil && *thisSiteAccountSnapshot.InfoSiteSnapshot.CityName != "" {
								return types.StringValue(*thisSiteAccountSnapshot.InfoSiteSnapshot.CityName)
							}
							return types.StringNull()
						}(),
					},
				)
				state.SiteLocation = stateSiteLocation
			}
		}
	}
	return state, true, nil
}

// socketSiteNativeRangeValidator validates that interface_index can only be specified
// for certain connection types and that local_ip is within the native_network_range subnet
type socketSiteNativeRangeValidator struct{}

func (v socketSiteNativeRangeValidator) Description(ctx context.Context) string {
	return "interface_index can only be specified for SOCKET_X1500, SOCKET_X1600, SOCKET_X1600_LTE, or SOCKET_X1700; local_ip must be within the native_network_range subnet"
}

func (v socketSiteNativeRangeValidator) MarkdownDescription(ctx context.Context) string {
	return "interface_index can only be specified for SOCKET_X1500, SOCKET_X1600, SOCKET_X1600_LTE, or SOCKET_X1700; local_ip must be within the native_network_range subnet"
}

func (v socketSiteNativeRangeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// Get the connection_type value from the config
	var connectionType types.String
	diags := req.Config.GetAttribute(ctx, path.Root("connection_type"), &connectionType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the native_range object
	var nativeRange types.Object
	diags = req.Config.GetAttribute(ctx, path.Root("native_range"), &nativeRange)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || nativeRange.IsNull() {
		return
	}

	// Get interface_index from native_range
	var interfaceIndex types.String
	diags = req.Config.GetAttribute(ctx, path.Root("native_range").AtName("interface_index"), &interfaceIndex)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || interfaceIndex.IsNull() || interfaceIndex.IsUnknown() {
		return
	}

	// Check if connection_type allows interface_index to be specified
	allowedTypes := map[string]bool{
		"SOCKET_X1500":     true,
		"SOCKET_X1600":     true,
		"SOCKET_X1600_LTE": true,
		"SOCKET_X1700":     true,
	}

	if !allowedTypes[connectionType.ValueString()] {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("interface_index"),
			"Invalid Configuration",
			fmt.Sprintf("interface_index can only be specified when connection_type is one of: SOCKET_X1500, SOCKET_X1600, SOCKET_X1600_LTE, SOCKET_X1700. Current connection_type: %s", connectionType.ValueString()),
		)
	}

	// Get local_ip from native_range
	var localIP types.String
	diags = req.Config.GetAttribute(ctx, path.Root("native_range").AtName("local_ip"), &localIP)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || localIP.IsNull() || localIP.IsUnknown() {
		return
	}

	// Get native_network_range from native_range
	var nativeNetworkRange types.String
	diags = req.Config.GetAttribute(ctx, path.Root("native_range").AtName("native_network_range"), &nativeNetworkRange)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || nativeNetworkRange.IsNull() || nativeNetworkRange.IsUnknown() {
		return
	}

	// Validate that local_ip is within native_network_range
	localIPStr := localIP.ValueString()
	subnetStr := nativeNetworkRange.ValueString()

	// Parse the local IP
	ip := net.ParseIP(localIPStr)
	if ip == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("local_ip"),
			"Invalid Local IP",
			fmt.Sprintf("local_ip '%s' is not a valid IP address", localIPStr),
		)
		return
	}

	// Parse the subnet CIDR
	_, ipNet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("native_network_range"),
			"Invalid Native Network Range",
			fmt.Sprintf("native_network_range '%s' is not a valid CIDR notation", subnetStr),
		)
		return
	}

	// Check if the IP is within the subnet
	if !ipNet.Contains(ip) {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("local_ip"),
			"Local IP Not in Native Range",
			fmt.Sprintf("Local IP must be within the Native Range IP. local_ip '%s' is not within native_network_range '%s'", localIPStr, subnetStr),
		)
	}
}
