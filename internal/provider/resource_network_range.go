package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

var (
	_ resource.Resource                = &networkRangeResource{}
	_ resource.ResourceWithConfigure   = &networkRangeResource{}
	_ resource.ResourceWithImportState = &networkRangeResource{}
)

func NewNetworkRangeResource() resource.Resource {
	return &networkRangeResource{}
}

type networkRangeResource struct {
	client *catoClientData
}

func (r *networkRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_range"
}

func (r *networkRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_network_range` resource contains the configuration parameters necessary to add a network range to a cato site. ([virtual socket in AWS/Azure, or physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.addNetworkRange()](https://api.catonetworks.com/documentation/#mutation-site.addNetworkRange).",
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
			},
			"interface_id": schema.StringAttribute{
				Description: "Network Interface ID",
				Required:    false,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"interface_index": schema.StringAttribute{
				Description: "Network Interface Index",
				Required:    false,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"WAN1",
						"WAN2",
						"LAN1",
						"LAN2",
						"LTE",
						"USB1",
						"USB2",
						"INT_1",
						"INT_2",
						"INT_3",
						"INT_4",
						"INT_5",
						"INT_6",
						"INT_7",
						"INT_8",
						"INT_9",
						"INT_10",
						"INT_11",
						"INT_12",
						"INT_13",
						"INT_14",
						"INT_15",
						"INT_16",
					),
				},
			},
			"internet_only": schema.BoolAttribute{
				Description: "Internet only network range (Only releveant for Routed range_type)",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"local_ip": schema.StringAttribute{
				Description: "Network range local ip",
				Optional:    true,
			},
			"mdns_reflector": schema.BoolAttribute{
				Description: "Site native range mDNS reflector. When enabled, the Socket functions as an mDNS gateway, it relays mDNS requests and response between all enabled subnets.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Network range name",
				Required:    true,
			},
			"range_type": schema.StringAttribute{
				Description: "Network range type (https://api.catonetworks.com/documentation/#definition-SubnetType)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Direct",
						"Native",
						"Routed",
						"SecondaryNative",
						"VLAN",
					),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet": schema.StringAttribute{
				Description: "Network range (CIDR)",
				Required:    true,
			},
			"translated_subnet": schema.StringAttribute{
				Description: "Network range translated native IP range (CIDR)",
				Optional:    true,
			},
			"dhcp_settings": schema.SingleNestedAttribute{
				Description: "Site native range DHCP settings (Only releveant for NATIVE and VLAN range_type)",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"dhcp_type": schema.StringAttribute{
						Description: "Network range dhcp type (https://api.catonetworks.com/documentation/#definition-DhcpType)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"ACCOUNT_DEFAULT",
								"DHCP_DISABLED",
								"DHCP_RANGE",
								"DHCP_RELAY",
							),
						},
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
			"vlan": schema.Int64Attribute{
				Description: "Network range VLAN ID (Only releveant for VLAN range_type)",
				Optional:    true,
			},
		},
	}
}

func (r *networkRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *networkRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *networkRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan NetworkRange
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that interface_id and interface_index cannot be set simultaneously
	tflog.Debug(ctx, "plan.InterfaceId.sample", map[string]interface{}{
		"plan.InterfaceId.IsNull()":    utils.InterfaceToJSONString(plan.InterfaceId.IsNull()),
		"plan.InterfaceIndex.IsNull()": utils.InterfaceToJSONString(plan.InterfaceIndex.IsNull()),
	})

	interfaceIdSet := !plan.InterfaceId.IsNull() && !plan.InterfaceId.IsUnknown()
	interfaceIndexSet := !plan.InterfaceIndex.IsNull() && !plan.InterfaceIndex.IsUnknown()
	if interfaceIdSet && interfaceIndexSet {
		resp.Diagnostics.AddError(
			"Conflicting Configuration",
			fmt.Sprintf("Both interface_id (%q) and interface_index (%q) are specified. "+
				"Only one interface specification method is allowed per network range.",
				plan.InterfaceId.ValueString(), plan.InterfaceIndex.ValueString()),
		)
		return
	}

	// Validate that the site ID exists
	if !plan.SiteId.IsNull() && !plan.SiteId.IsUnknown() {
		// If interface_id is set, use it to validate the site network interface
		var err error
		var curInterfaceId, curInterfaceIndex string

		if !plan.InterfaceId.IsNull() && !plan.InterfaceId.IsUnknown() {
			// When interface_id is provided, get the corresponding interface_index
			_, curInterfaceIndex, err = getSiteNetworkInterface(ctx, r.client, plan.SiteId.ValueString(), plan.InterfaceId.ValueString(), "", "")
			if err == nil {
				// Set the interface_index in the plan based on the lookup
				plan.InterfaceIndex = types.StringValue(curInterfaceIndex)
			}
		} else if !plan.InterfaceIndex.IsNull() && !plan.InterfaceIndex.IsUnknown() {
			// When interface_index is provided, get the corresponding interface_id
			curInterfaceId, _, err = getSiteNetworkInterface(ctx, r.client, plan.SiteId.ValueString(), "", plan.InterfaceIndex.ValueString(), "")
			if err == nil {
				// Set the interface_id in the plan based on the lookup
				plan.InterfaceId = types.StringValue(curInterfaceId)
			}
		}

		if err != nil {
			resp.Diagnostics.AddError(
				"Site Interface Validation Error",
				err.Error(),
			)
			return
		}
	}

	// Validate that InternetOnly and MdnsReflector cannot be set simultaneously
	if !plan.InternetOnly.IsNull() && !plan.MdnsReflector.IsNull() &&
		plan.InternetOnly.ValueBool() == true && plan.MdnsReflector.ValueBool() == true {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"mDNS and Internet Only cannot be set simultaneously",
		)
		return
	}

	// mDNS not supported for rangeType Routed, set to null
	if plan.RangeType == types.StringValue("Routed") && !plan.MdnsReflector.IsNull() &&
		!plan.MdnsReflector.IsNull() && plan.MdnsReflector.ValueBool() == true {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"mDNS reflector is not a supported configuration for routed subnets",
		)
		return
	}

	// mDNS not supported for rangeType Routed, set to null
	curMdnsReflector := plan.MdnsReflector.ValueBoolPointer()
	if plan.RangeType == types.StringValue("Routed") {
		curMdnsReflector = nil
	}

	// setting input
	input := cato_models.AddNetworkRangeInput{
		Name:             plan.Name.ValueString(),
		RangeType:        (cato_models.SubnetType)(plan.RangeType.ValueString()),
		Subnet:           plan.Subnet.ValueString(),
		LocalIP:          plan.LocalIp.ValueStringPointer(),
		TranslatedSubnet: plan.TranslatedSubnet.ValueStringPointer(),
		Gateway:          plan.Gateway.ValueStringPointer(),
		Vlan:             plan.Vlan.ValueInt64Pointer(),
		InternetOnly:     plan.InternetOnly.ValueBoolPointer(),
		MdnsReflector:    curMdnsReflector,
	}

	// get planned DHCP settings Object value, or set default value if null (for VLAN Type)
	var dhcpSettings DhcpSettings
	if !plan.DhcpSettings.IsNull() && plan.RangeType != types.StringValue("VLAN") && plan.RangeType != types.StringValue("NATIVE") {
		resp.Diagnostics.AddError(
			"Invalid dhcpSettings configuration",
			"Configuring dhcpSettings is allowed only for Native or VLAN network range types.",
		)
		return
	}

	// Validate DHCP settings configuration based on dhcp_type
	if !plan.DhcpSettings.IsNull() && !plan.DhcpSettings.IsUnknown() {
		diags = plan.DhcpSettings.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		dhcpType := dhcpSettings.DhcpType.ValueString()

		// Validate that interface_id and interface_index cannot be set simultaneously
		tflog.Debug(ctx, "networkRange.create.dhcpSettings", map[string]interface{}{
			"name":                                      utils.InterfaceToJSONString(plan.Name.ValueString()),
			"dhcpSettings.IpRange.IsNull()":             utils.InterfaceToJSONString(dhcpSettings.IpRange.IsNull()),
			"dhcpSettings.IpRange.IsUnknown()":          utils.InterfaceToJSONString(dhcpSettings.IpRange.IsUnknown()),
			"dhcpSettings.IpRange.ValueString()":        utils.InterfaceToJSONString(dhcpSettings.IpRange.ValueString()),
			"dhcpSettings.RelayGroupId.IsNull()":        utils.InterfaceToJSONString(dhcpSettings.RelayGroupId.IsNull()),
			"dhcpSettings.RelayGroupId.IsUnknown()":     utils.InterfaceToJSONString(dhcpSettings.RelayGroupId.IsUnknown()),
			"dhcpSettings.RelayGroupId.ValueString()":   utils.InterfaceToJSONString(dhcpSettings.RelayGroupId.ValueString()),
			"dhcpSettings.RelayGroupName.IsNull()":      utils.InterfaceToJSONString(dhcpSettings.RelayGroupName.IsNull()),
			"dhcpSettings.RelayGroupName.IsUnknown()":   utils.InterfaceToJSONString(dhcpSettings.RelayGroupName.IsUnknown()),
			"dhcpSettings.RelayGroupName.ValueString()": utils.InterfaceToJSONString(dhcpSettings.RelayGroupName.ValueString()),
		})

		// Validate DHCP_DISABLED configuration
		if dhcpType == "DHCP_DISABLED" {
			if (!dhcpSettings.IpRange.IsNull() && !dhcpSettings.IpRange.IsUnknown() && dhcpSettings.IpRange.ValueString() != "") ||
				(!dhcpSettings.RelayGroupId.IsNull() && !dhcpSettings.RelayGroupId.IsUnknown() && dhcpSettings.RelayGroupId.ValueString() != "") ||
				(!dhcpSettings.RelayGroupName.IsNull() && !dhcpSettings.RelayGroupName.IsUnknown() && dhcpSettings.RelayGroupName.ValueString() != "") {
				resp.Diagnostics.AddError(
					"Invalid DHCP Configuration",
					"When dhcp_type is DHCP_DISABLED, dhcp_ip_range, dhcp_relay_group_id, and dhcp_relay_group_name must be null, unset, or empty strings.",
				)
				return
			}
		}

		// Validate DHCP_RANGE configuration
		if dhcpType == "DHCP_RANGE" {
			if (dhcpSettings.IpRange.IsNull() || dhcpSettings.IpRange.IsUnknown() || dhcpSettings.IpRange.ValueString() == "") ||
				(!dhcpSettings.RelayGroupId.IsNull() && !dhcpSettings.RelayGroupId.IsUnknown() && dhcpSettings.RelayGroupId.ValueString() != "") ||
				(!dhcpSettings.RelayGroupName.IsNull() && !dhcpSettings.RelayGroupName.IsUnknown() && dhcpSettings.RelayGroupName.ValueString() != "") {
				resp.Diagnostics.AddError(
					"Invalid DHCP Configuration",
					"When dhcp_type is DHCP_RANGE, dhcp_ip_range must be provided (not null, unset, or empty string), and dhcp_relay_group_id and dhcp_relay_group_name must be null, unset, or empty strings.",
				)
				return
			}
		}
	}

	if plan.RangeType == types.StringValue("Routed") {
		if !plan.LocalIp.IsNull() && !plan.LocalIp.IsUnknown() && plan.LocalIp.ValueString() != "" {
			resp.Diagnostics.AddError(
				"Invalid configuration",
				"Configuring LocalIp is only supported for VLAN, Native and Direct range types.",
			)
			return
		}
		input.Gateway = plan.Gateway.ValueStringPointer()
	} else if plan.RangeType == types.StringValue("VLAN") || plan.RangeType == types.StringValue("Direct") {
		if !plan.Gateway.IsNull() && !plan.Gateway.IsUnknown() && plan.Gateway.ValueString() != "" {
			resp.Diagnostics.AddError(
				"Invalid configuration",
				"Configuring gateway is only supported for Routed range types.",
			)
			return
		}
		input.LocalIP = plan.LocalIp.ValueStringPointer()

		if plan.RangeType == types.StringValue("VLAN") {
			if plan.DhcpSettings.IsNull() {
				dhcpSettings.DhcpType = types.StringValue("DHCP_DISABLED")
			} else {
				diags = plan.DhcpSettings.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}

			if !plan.DhcpSettings.IsNull() && !plan.DhcpSettings.IsUnknown() {
				input.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{}
				var dhcpSettingsInput DhcpSettings
				diags = plan.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				input.DhcpSettings.DhcpType = (cato_models.DhcpType)(dhcpSettingsInput.DhcpType.ValueString())
				input.DhcpSettings.IPRange = dhcpSettingsInput.IpRange.ValueStringPointer()
				// Note: RelayGroupID is only set for DHCP_RELAY type below
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
					input.DhcpSettings.DhcpMicrosegmentation = dhcpSettingsInput.DhcpMicrosegmentation.ValueBoolPointer()
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
					input.DhcpSettings.RelayGroupID = &resolvedRelayGroupId

					// Update the plan's DhcpSettings with the resolved relay group ID
					// This ensures the value is known after apply
					dhcpSettingsInput.RelayGroupId = types.StringValue(resolvedRelayGroupId)
					plan.DhcpSettings, _ = types.ObjectValueFrom(ctx, DhcpSettingsAttrTypes, dhcpSettingsInput)
				}
			}
		}
	}

	tflog.Debug(ctx, "Create.SiteAddNetworkRange.request", map[string]interface{}{
		"request":     utils.InterfaceToJSONString(input),
		"interfaceId": utils.InterfaceToJSONString(plan.InterfaceId.ValueString()),
	})
	networkRange, err := r.client.catov2.SiteAddNetworkRange(ctx, plan.InterfaceId.ValueString(), input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddNetworkRange.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(networkRange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API SiteAddNetworkRange error",
			err.Error(),
		)
		return
	}

	// hydrate the state with API data
	hydratedState, rangeExists, hydrateErr := r.hydrateNetworkRangeState(ctx, plan, networkRange.Site.AddNetworkRange.NetworkRangeID)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating socket site state",
			hydrateErr.Error(),
		)
		return
	}
	if !rangeExists {
		tflog.Warn(ctx, "siteRange not found, siteRange resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	hydratedState.InterfaceId = types.StringValue(plan.InterfaceId.ValueString())
	hydratedState.Id = types.StringValue(networkRange.Site.AddNetworkRange.NetworkRangeID)

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *networkRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, rangeExists, hydrateErr := r.hydrateNetworkRangeState(ctx, state, state.Id.ValueString())
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating socket site state",
			hydrateErr.Error(),
		)
		return
	}

	if !rangeExists {
		tflog.Warn(ctx, "siteRange not found, siteRange resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *networkRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan NetworkRange
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// // Validate that interface_id and interface_index cannot be set simultaneously
	// if !plan.InterfaceId.IsNull() && !plan.InterfaceIndex.IsNull() {
	// 	resp.Diagnostics.AddError(
	// 		"Invalid Configuration",
	// 		"Interface ID and Interface Index cannot be set simultaneously",
	// 	)
	// 	return
	// }

	// Validate that the site ID exists
	if !plan.SiteId.IsNull() && !plan.SiteId.IsUnknown() {
		// If interface_id is set, use it to validate the site network interface
		var err error
		tflog.Info(ctx, "networkRangeUpdate.plan.InterfaceAttr", map[string]interface{}{
			"plan.InterfaceId":                utils.InterfaceToJSONString(plan.InterfaceId),
			"plan.InterfaceId.IsUnknown()":    plan.InterfaceId.IsUnknown(),
			"plan.InterfaceId.IsNull()":       plan.InterfaceId.IsNull(),
			"plan.InterfaceIndex":             utils.InterfaceToJSONString(plan.InterfaceIndex),
			"plan.InterfaceIndex.IsUnknown()": plan.InterfaceIndex.IsUnknown(),
			"plan.InterfaceIndex.IsNull()":    plan.InterfaceIndex.IsNull(),
		})
		if !plan.InterfaceId.IsNull() && !plan.InterfaceId.IsUnknown() {
			tflog.Info(ctx, "networkRangeUpdate.plan.InterfaceId.IsNull", map[string]interface{}{})
			_, _, err = getSiteNetworkInterface(ctx, r.client, plan.SiteId.ValueString(), plan.InterfaceId.ValueString(), "", "")
		} else if !plan.InterfaceIndex.IsNull() && !plan.InterfaceIndex.IsUnknown() {
			tflog.Info(ctx, "networkRangeUpdate.plan.InterfaceIndex.IsNull", map[string]interface{}{})
			_, _, err = getSiteNetworkInterface(ctx, r.client, plan.SiteId.ValueString(), "", plan.InterfaceIndex.ValueString(), "")
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Site Validation Error",
				err.Error(),
			)
			return
		}
	}

	// Validate that InternetOnly and MdnsReflector cannot be set simultaneously
	if !plan.InternetOnly.IsNull() && !plan.MdnsReflector.IsNull() &&
		plan.InternetOnly.ValueBool() == true && plan.MdnsReflector.ValueBool() == true {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"mDNS and Internet Only cannot be set simultaneously",
		)
		return
	}

	// mDNS not supported for rangeType Routed, set to null
	curMdnsReflector := plan.MdnsReflector.ValueBoolPointer()
	if plan.RangeType == types.StringValue("Routed") {
		curMdnsReflector = nil
	}

	// setting input
	input := cato_models.UpdateNetworkRangeInput{
		Name:             plan.Name.ValueStringPointer(),
		RangeType:        (*cato_models.SubnetType)(plan.RangeType.ValueStringPointer()),
		Subnet:           plan.Subnet.ValueStringPointer(),
		LocalIP:          plan.LocalIp.ValueStringPointer(),
		TranslatedSubnet: plan.TranslatedSubnet.ValueStringPointer(),
		Gateway:          plan.Gateway.ValueStringPointer(),
		Vlan:             plan.Vlan.ValueInt64Pointer(),
		InternetOnly:     plan.InternetOnly.ValueBoolPointer(),
		MdnsReflector:    curMdnsReflector,
	}

	// get planned DHCP settings Object value, or set default value if null (for VLAN Type)
	var dhcpSettings DhcpSettings
	if !plan.DhcpSettings.IsNull() && plan.RangeType != types.StringValue("VLAN") && plan.RangeType != types.StringValue("NATIVE") {
		resp.Diagnostics.AddError(
			"Invalid dhcpSettings configuration",
			"Configuring dhcpSettings is allowed only for Native or VLAN network range types.",
		)
		return
	}

	// Validate that interface_id and interface_index cannot be set simultaneously
	tflog.Debug(ctx, "networkRange.update.dhcpSettings", map[string]interface{}{
		"name":                                      utils.InterfaceToJSONString(plan.Name.ValueStringPointer()),
		"dhcpSettings.IpRange.IsNull()":             utils.InterfaceToJSONString(dhcpSettings.IpRange.IsNull()),
		"dhcpSettings.IpRange.IsUnknown()":          utils.InterfaceToJSONString(dhcpSettings.IpRange.IsUnknown()),
		"dhcpSettings.IpRange.ValueString()":        utils.InterfaceToJSONString(dhcpSettings.IpRange.ValueString()),
		"dhcpSettings.RelayGroupId.IsNull()":        utils.InterfaceToJSONString(dhcpSettings.RelayGroupId.IsNull()),
		"dhcpSettings.RelayGroupId.IsUnknown()":     utils.InterfaceToJSONString(dhcpSettings.RelayGroupId.IsUnknown()),
		"dhcpSettings.RelayGroupId.ValueString()":   utils.InterfaceToJSONString(dhcpSettings.RelayGroupId.ValueString()),
		"dhcpSettings.RelayGroupName.IsNull()":      utils.InterfaceToJSONString(dhcpSettings.RelayGroupName.IsNull()),
		"dhcpSettings.RelayGroupName.IsUnknown()":   utils.InterfaceToJSONString(dhcpSettings.RelayGroupName.IsUnknown()),
		"dhcpSettings.RelayGroupName.ValueString()": utils.InterfaceToJSONString(dhcpSettings.RelayGroupName.ValueString()),
	})

	// Validate DHCP settings configuration based on dhcp_type
	if !plan.DhcpSettings.IsNull() && !plan.DhcpSettings.IsUnknown() {
		diags = plan.DhcpSettings.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		dhcpType := dhcpSettings.DhcpType.ValueString()

		// Validate DHCP_DISABLED configuration
		if dhcpType == "DHCP_DISABLED" {
			if (!dhcpSettings.IpRange.IsNull() && !dhcpSettings.IpRange.IsUnknown() && dhcpSettings.IpRange.ValueString() != "") ||
				(!dhcpSettings.RelayGroupId.IsNull() && !dhcpSettings.RelayGroupId.IsUnknown() && dhcpSettings.RelayGroupId.ValueString() != "") ||
				(!dhcpSettings.RelayGroupName.IsNull() && !dhcpSettings.RelayGroupName.IsUnknown() && dhcpSettings.RelayGroupName.ValueString() != "") {
				resp.Diagnostics.AddError(
					"Invalid DHCP Configuration",
					"When dhcp_type is DHCP_DISABLED, dhcp_ip_range, dhcp_relay_group_id, and dhcp_relay_group_name must be null, unset, or empty strings.",
				)
				return
			}
		}

		// Validate DHCP_RANGE configuration
		if dhcpType == "DHCP_RANGE" {
			if (dhcpSettings.IpRange.IsNull() || dhcpSettings.IpRange.IsUnknown() || dhcpSettings.IpRange.ValueString() == "") ||
				(!dhcpSettings.RelayGroupId.IsNull() && !dhcpSettings.RelayGroupId.IsUnknown() && dhcpSettings.RelayGroupId.ValueString() != "") ||
				(!dhcpSettings.RelayGroupName.IsNull() && !dhcpSettings.RelayGroupName.IsUnknown() && dhcpSettings.RelayGroupName.ValueString() != "") {
				resp.Diagnostics.AddError(
					"Invalid DHCP Configuration",
					"When dhcp_type is DHCP_RANGE, dhcp_ip_range must be provided (not null, unset, or empty string), and dhcp_relay_group_id and dhcp_relay_group_name must be null, unset, or empty strings.",
				)
				return
			}
		}
	}

	if plan.RangeType == types.StringValue("Routed") {
		if !plan.LocalIp.IsNull() && !plan.LocalIp.IsUnknown() && plan.LocalIp.ValueString() != "" {
			resp.Diagnostics.AddError(
				"Invalid configuration",
				"Configuring LocalIp is only supported for VLAN, Native and Direct range types.",
			)
			return
		}
		input.Gateway = plan.Gateway.ValueStringPointer()
	} else if plan.RangeType == types.StringValue("VLAN") || plan.RangeType == types.StringValue("Direct") {
		if !plan.Gateway.IsNull() && !plan.Gateway.IsUnknown() && plan.Gateway.ValueString() != "" {
			resp.Diagnostics.AddError(
				"Invalid configuration",
				"Configuring gateway is only supported for Routed range types.",
			)
			return
		}
		input.LocalIP = plan.LocalIp.ValueStringPointer()

		if plan.RangeType == types.StringValue("VLAN") {
			if plan.DhcpSettings.IsNull() {
				dhcpSettings.DhcpType = types.StringValue("DHCP_DISABLED")
			} else {
				diags = plan.DhcpSettings.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}

			if !plan.DhcpSettings.IsNull() && !plan.DhcpSettings.IsUnknown() {
				input.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{}
				var dhcpSettingsInput DhcpSettings
				diags = plan.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				input.DhcpSettings.DhcpType = (cato_models.DhcpType)(dhcpSettingsInput.DhcpType.ValueString())
				input.DhcpSettings.IPRange = dhcpSettingsInput.IpRange.ValueStringPointer()
				// Note: RelayGroupID is only set for DHCP_RELAY type below
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
					input.DhcpSettings.DhcpMicrosegmentation = dhcpSettingsInput.DhcpMicrosegmentation.ValueBoolPointer()
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
					input.DhcpSettings.RelayGroupID = &resolvedRelayGroupId

					// Update the plan's DhcpSettings with the resolved relay group ID
					// This ensures the value is known after apply
					dhcpSettingsInput.RelayGroupId = types.StringValue(resolvedRelayGroupId)
					plan.DhcpSettings, _ = types.ObjectValueFrom(ctx, DhcpSettingsAttrTypes, dhcpSettingsInput)
				}
			}
		}
	}

	tflog.Debug(ctx, "network range update", map[string]interface{}{
		"input":          utils.InterfaceToJSONString(input),
		"lanInterfaceID": plan.Id.ValueString(),
	})

	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.request", map[string]interface{}{
		"lanInterfaceID": plan.Id.ValueString(),
		"input":          utils.InterfaceToJSONString(input),
	})
	siteUpdateNetworkRangeResponse, err := r.client.catov2.SiteUpdateNetworkRange(ctx, plan.Id.ValueString(), input, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteUpdateNetworkRangeResponse),
	})
	if err != nil {
		var apiError struct {
			NetworkErrors interface{} `json:"networkErrors"`
			GraphqlErrors []struct {
				Message string   `json:"message"`
				Path    []string `json:"path"`
			} `json:"graphqlErrors"`
		}
		interfaceNotPresent := false
		if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphqlErrors) > 0 {
			msg := apiError.GraphqlErrors[0].Message
			if strings.Contains(msg, "Network range with id: ") && strings.Contains(msg, "is not found") {
				interfaceNotPresent = true
			}
		}
		if !interfaceNotPresent {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
			return
		}
		// If the network range is not present, delete the resource and recreate it
		tflog.Warn(ctx, "Network range not found during update, recreating resource")
		// Remove the resource from state
		resp.State.RemoveResource(ctx)
		return
	}

	// hydrate the state with API data
	hydratedState, rangeExists, hydrateErr := r.hydrateNetworkRangeState(ctx, plan, plan.Id.ValueString())
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating socket site state",
			hydrateErr.Error(),
		)
		return
	}
	if !rangeExists {
		tflog.Warn(ctx, "siteRange not found, siteRange resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	hydratedState.InterfaceId = types.StringValue(plan.InterfaceId.ValueString())
	hydratedState.Id = types.StringValue(siteUpdateNetworkRangeResponse.Site.UpdateNetworkRange.NetworkRangeID)

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *networkRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state NetworkRange
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if interface is already removed and fail gracefully
	//	if len(querySiteResult.EntityLookup.GetItems()) == 1 {
	_, err := r.client.catov2.SiteRemoveNetworkRange(ctx, state.Id.ValueString(), r.client.AccountId)
	if err != nil {
		var apiError struct {
			NetworkErrors interface{} `json:"networkErrors"`
			GraphqlErrors []struct {
				Message string   `json:"message"`
				Path    []string `json:"path"`
			} `json:"graphqlErrors"`
		}
		interfaceNotPresent := false
		if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphqlErrors) > 0 {
			msg := apiError.GraphqlErrors[0].Message
			if strings.Contains(msg, "Network range with id: ") && strings.Contains(msg, "is not found") {
				interfaceNotPresent = true
			}
		}
		if !interfaceNotPresent {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
			return
		}
	}

}

// hydrateNetworkRangeState populates the NetworkRange state with data from API responses
func (r *networkRangeResource) hydrateNetworkRangeState(ctx context.Context, state NetworkRange, networkRangeID string) (NetworkRange, bool, error) {
	// check if site exist, else remove resource
	querySiteRangeResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityTypeSiteRange, nil, nil, nil, nil, []string{networkRangeID}, nil, nil, nil)
	tflog.Debug(ctx, "hydrateNetworkRangeState.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteRangeResult),
	})
	if err != nil {
		return NetworkRange{}, false, fmt.Errorf("catov2 API EntityLookup error: %w", err)
	}

	// check if site exist before refreshing
	// we should only have one entry since we are filtering on site ID
	if len(querySiteRangeResult.EntityLookup.GetItems()) != 1 {
		return state, false, nil
	}

	for _, curRange := range querySiteRangeResult.EntityLookup.Items {
		// find the siteRange entry we need
		// Compare against networkRangeID (the parameter), not state.Id, because during Create
		// state.Id is unknown/empty - we need to match using the ID returned from the API
		if curRange.Entity.ID == networkRangeID {
			state.Id = types.StringValue(curRange.Entity.ID)
			helperFields := curRange.GetHelperFields()
			siteIdStr := cast.ToString(helperFields["siteId"])
			interfaceNameStr := cast.ToString(helperFields["interfaceName"])

			// Preserve existing interface_index if it's already set and valid
			if !state.InterfaceIndex.IsNull() && !state.InterfaceIndex.IsUnknown() {
				tflog.Debug(ctx, "hydrateNetworkRangeState.EntityLookup.response.hasInterfaceIndex", map[string]interface{}{
					"curRange":                           utils.InterfaceToJSONString(curRange),
					"helperFields":                       utils.InterfaceToJSONString(helperFields),
					"state.InterfaceIndex.ValueString()": utils.InterfaceToJSONString(state.InterfaceIndex.ValueString()),
				})
				// Validate that the existing interface_index is still valid
				existingInterfaceId, _, err := getSiteNetworkInterface(ctx, r.client, siteIdStr, "", state.InterfaceIndex.ValueString(), interfaceNameStr)
				if err != nil {
					// If validation fails, log and fall back to lookup by interface name
					tflog.Warn(ctx, "Existing interface_index validation failed, performing lookup by interface name", map[string]interface{}{
						"existing_interface_index": state.InterfaceIndex.ValueString(),
						"error":                    err.Error(),
					})
					interfaceId, curInterfaceIndex, err := getSiteNetworkInterface(ctx, r.client, siteIdStr, "", "", interfaceNameStr)
					if err != nil {
						return state, false, fmt.Errorf("Site Interface Validation Error: %w", err)
					}
					state.InterfaceId = types.StringValue(interfaceId)
					state.InterfaceIndex = types.StringValue(curInterfaceIndex)
				} else {
					// Existing interface_index is valid, just update interface_id if needed
					state.InterfaceId = types.StringValue(existingInterfaceId)
					// Keep existing interface_index unchanged
				}
			} else {
				tflog.Debug(ctx, "hydrateNetworkRangeState.EntityLookup.response.noInterfaceIndex", map[string]interface{}{
					"curRange":         utils.InterfaceToJSONString(curRange),
					"helperFields":     utils.InterfaceToJSONString(helperFields),
					"interfaceNameStr": utils.InterfaceToJSONString(interfaceNameStr),
				})

				// No existing interface_index, perform normal lookup
				interfaceId, curInterfaceIndex, err := getSiteNetworkInterface(ctx, r.client, siteIdStr, "", "", interfaceNameStr)
				tflog.Debug(ctx, "getSiteNetworkInterface.return", map[string]interface{}{
					"interfaceId":       utils.InterfaceToJSONString(helperFields),
					"curInterfaceIndex": utils.InterfaceToJSONString(curInterfaceIndex),
				})
				if err != nil {
					return state, false, fmt.Errorf("Site Interface Validation Error: %w", err)
				}
				state.InterfaceId = types.StringValue(interfaceId)
				state.InterfaceIndex = types.StringValue(curInterfaceIndex)
			}

			if internetOnlyVal, ok := curRange.HelperFields["internetOnly"]; ok {
				state.InternetOnly = types.BoolValue(cast.ToBool(internetOnlyVal))
			}

			// MdnsReflector is not supported for routed subnets, and the rangeType is not available or known via API response
			// Preserve mdnsReflector value from plan/state, only set to null if not present in state
			if state.MdnsReflector.IsNull() || state.MdnsReflector.IsUnknown() {
				if mdnsReflectorVal, ok := curRange.HelperFields["mdnsReflector"]; ok {
					state.MdnsReflector = types.BoolValue(cast.ToBool(mdnsReflectorVal))
				} else {
					state.MdnsReflector = types.BoolNull()
				}
			}

			if curRange.GetEntity() != nil && curRange.GetEntity().Name != nil {
				nameParts := strings.Split(*curRange.GetEntity().Name, " \\ ")
				state.Name = types.StringValue(nameParts[len(nameParts)-1])
			}

			// Retrieve current rangeType from the API response, and translate for what API is expecting for updates
			curRangeTypeVal := "VLAN"
			if rangeTypeVal, ok := curRange.HelperFields["rangeType"]; ok {
				curRangeTypeVal = cast.ToString(rangeTypeVal)
			}

			// Parse gateway value, used for both localIp and gateway
			gatewayLocalIpVal := types.StringNull()
			if gatewayVal, ok := curRange.HelperFields["gateway"]; ok {
				gatewayLocalIpVal = types.StringValue(cast.ToString(gatewayVal))
			}

			// Assign gatewayLocalIpVal to gateway or localIp
			switch curRangeTypeVal {
			case "ROUTED_ROUTE":
				state.RangeType = types.StringValue("Routed")
				state.Gateway = gatewayLocalIpVal
			case "DIRECT_ROUTE":
				state.RangeType = types.StringValue("Direct")
				state.LocalIp = gatewayLocalIpVal
			case "VLAN":
				state.RangeType = types.StringValue("VLAN")
				state.LocalIp = gatewayLocalIpVal
			default:
				state.RangeType = types.StringValue(curRangeTypeVal)
			}

			if siteIdVal, ok := curRange.HelperFields["siteId"]; ok {
				state.SiteId = types.StringValue(cast.ToString(siteIdVal))
			}
			if subnetVal, ok := curRange.HelperFields["subnet"]; ok {
				state.Subnet = types.StringValue(cast.ToString(subnetVal))
			}

			if translatedSubnetVal, ok := curRange.HelperFields["translated_subnet"]; ok {
				translatedSubnetStr := cast.ToString(translatedSubnetVal)
				if translatedSubnetStr != "" {
					state.TranslatedSubnet = types.StringValue(translatedSubnetStr)
				} else {
					state.TranslatedSubnet = types.StringNull()
				}
			}

			// Always populate VLAN from API if available
			// API returns this as 'vlanTag' not 'vlan'
			if vlanVal, ok := curRange.HelperFields["vlanTag"]; ok {
				vlanInt64 := cast.ToInt64(vlanVal)
				if vlanInt64 > 0 {
					state.Vlan = types.Int64Value(vlanInt64)
				} else {
					state.Vlan = types.Int64Null()
				}
			} else {
				state.Vlan = types.Int64Null()
			}

			// Always populate DHCP settings from API for VLAN and Native range types
			// Check if DHCP settings exist in API response
			_, hasDhcpType := curRange.HelperFields["dhcpType"]
			
			// Populate DHCP settings if:
			// 1. They were already in state (preserve user config)
			// 2. OR the API returns dhcpType (meaning DHCP is configured)
			if (!state.DhcpSettings.IsNull() && !state.DhcpSettings.IsUnknown()) || hasDhcpType {
				// Start with null values as defaults - unknown values MUST be resolved to known values
				dhcpType := types.StringNull()
				ipRange := types.StringNull()
				relayGroupId := types.StringNull()
				relayGroupName := types.StringNull()
				dhcpMicrosegmentation := types.BoolValue(false)

				// Extract current DHCP settings from state if present (not null/unknown)
				if !state.DhcpSettings.IsNull() && !state.DhcpSettings.IsUnknown() {
					var currentDhcpSettings DhcpSettings
					diags := state.DhcpSettings.As(ctx, &currentDhcpSettings, basetypes.ObjectAsOptions{})
					if diags.HasError() {
						return state, false, fmt.Errorf("failed to read current dhcp settings: %v", diags)
					}

					// Preserve concrete values from state (not null/unknown)
					if !currentDhcpSettings.DhcpType.IsNull() && !currentDhcpSettings.DhcpType.IsUnknown() {
						dhcpType = currentDhcpSettings.DhcpType
					}
					if !currentDhcpSettings.IpRange.IsNull() && !currentDhcpSettings.IpRange.IsUnknown() {
						ipRange = currentDhcpSettings.IpRange
					}
					if !currentDhcpSettings.RelayGroupId.IsNull() && !currentDhcpSettings.RelayGroupId.IsUnknown() {
						relayGroupId = currentDhcpSettings.RelayGroupId
					}
					if !currentDhcpSettings.RelayGroupName.IsNull() && !currentDhcpSettings.RelayGroupName.IsUnknown() {
						relayGroupName = currentDhcpSettings.RelayGroupName
					}
					if !currentDhcpSettings.DhcpMicrosegmentation.IsNull() && !currentDhcpSettings.DhcpMicrosegmentation.IsUnknown() {
						dhcpMicrosegmentation = currentDhcpSettings.DhcpMicrosegmentation
					}
				}

				// Override with API values if available (API values take precedence when present)
				if dhcpTypeVal, ok := curRange.HelperFields["dhcpType"]; ok {
					dhcpType = types.StringValue(cast.ToString(dhcpTypeVal))
				}

				if dhcpRangeVal, ok := curRange.HelperFields["dhcpRange"]; ok {
					ipRange = types.StringValue(cast.ToString(dhcpRangeVal))
				}

				if dhcpRelayGroupIdVal, ok := curRange.HelperFields["dhcpRelayGroupId"]; ok {
					relayGroupId = types.StringValue(cast.ToString(dhcpRelayGroupIdVal))
				}

				if dhcpRelayGroupNameVal, ok := curRange.HelperFields["dhcpRelayGroupName"]; ok {
					relayGroupName = types.StringValue(cast.ToString(dhcpRelayGroupNameVal))
				}

				if microsegmentationVal, ok := curRange.HelperFields["microsegmentation"]; ok {
					dhcpMicrosegmentation = types.BoolValue(cast.ToBool(microsegmentationVal))
				}

				// Manually construct Object to ensure no Unknown values persist
				dhcpSettingsObject, dErr := types.ObjectValue(
					DhcpSettingsAttrTypes,
					map[string]attr.Value{
						"dhcp_type":              dhcpType,
						"ip_range":               ipRange,
						"relay_group_id":         relayGroupId,
						"relay_group_name":       relayGroupName,
						"dhcp_microsegmentation": dhcpMicrosegmentation,
					},
				)
				if dErr.HasError() {
					return state, false, fmt.Errorf("failed to convert dhcp settings to object: %v", dErr)
				}

				state.DhcpSettings = dhcpSettingsObject
			} else {
				// No DHCP settings in API or state, set to null
				state.DhcpSettings = types.ObjectNull(DhcpSettingsAttrTypes)
			}
		}
	}
	return state, true, nil
}
