package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
					stringplanmodifier.UseStateForUnknown(),
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
				Computed:    true,
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

	// Hydrate the state from the API
	var state NetworkRange
	state.Id = types.StringValue(req.ID)

	hydratedState, rangeExists, hydrateErr := r.hydrateNetworkRangeState(ctx, state, req.ID)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating network range state during import",
			hydrateErr.Error(),
		)
		return
	}

	if !rangeExists {
		resp.Diagnostics.AddError(
			"Network Range Not Found",
			fmt.Sprintf("Network range with ID %q not found during import", req.ID),
		)
		return
	}

	// Set the hydrated state
	diags := resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
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
			// Track if user actually specified dhcp_settings (for later hydration)
			dhcpSettingsWasSpecified := !plan.DhcpSettings.IsNull() && !plan.DhcpSettings.IsUnknown()

			if plan.DhcpSettings.IsNull() {
				// Set default DHCP settings for API when null for VLAN ranges
				// NOTE: Do NOT modify plan.DhcpSettings - keep it null to match Terraform's expected state
				dhcpSettings.DhcpType = types.StringValue("DHCP_DISABLED")
				dhcpSettings.DhcpMicrosegmentation = types.BoolNull()
				dhcpSettings.IpRange = types.StringNull()
				dhcpSettings.RelayGroupId = types.StringNull()
				dhcpSettings.RelayGroupName = types.StringNull()
			} else {
				diags = plan.DhcpSettings.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}

			// Only send DHCP settings to API - either defaults or user-specified
			if dhcpSettingsWasSpecified {
				input.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{}
				var dhcpSettingsInput DhcpSettings
				diags = plan.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				input.DhcpSettings.DhcpType = (cato_models.DhcpType)(dhcpSettingsInput.DhcpType.ValueString())
				input.DhcpSettings.IPRange = dhcpSettingsInput.IpRange.ValueStringPointer()
				// Note: RelayGroupID is only set for DHCP_RELAY type below
				// Validate that dhcp_microsegmentation is only set to true when dhcp_type is DHCP_RANGE or DHCP_RELAY
				if !dhcpSettingsInput.DhcpMicrosegmentation.IsNull() && !dhcpSettingsInput.DhcpMicrosegmentation.IsUnknown() {
					dhcpType := dhcpSettingsInput.DhcpType.ValueString()
					if dhcpSettingsInput.DhcpMicrosegmentation.ValueBool() == true && dhcpType != "DHCP_RANGE" && dhcpType != "DHCP_RELAY" {
						resp.Diagnostics.AddError(
							"Invalid DHCP Microsegmentation Configuration",
							"dhcp_microsegmentation can only be configured when dhcp_type is set to DHCP_RANGE or DHCP_RELAY",
						)
						return
					}
				}

				// Set dhcpMicrosegmentation for DHCP_RANGE or DHCP_RELAY types
				dhcpType := dhcpSettingsInput.DhcpType.ValueString()
				if (dhcpType == "DHCP_RANGE" || dhcpType == "DHCP_RELAY") && (!dhcpSettingsInput.DhcpMicrosegmentation.IsUnknown()) {
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

	// Validate and set vlan value
	var vlanValue *int64
	if plan.RangeType == types.StringValue("VLAN") {
		// Only set vlan if range type is VLAN
		vlanValue = plan.Vlan.ValueInt64Pointer()
	} else {
		// For non-VLAN types, vlan must not be set
		if !plan.Vlan.IsNull() && !plan.Vlan.IsUnknown() {
			resp.Diagnostics.AddError(
				"Invalid configuration",
				fmt.Sprintf("vlan can only be configured for VLAN range types, but range_type is %s", plan.RangeType.ValueString()),
			)
			return
		}
		vlanValue = nil
	}

	// setting input
	input := cato_models.UpdateNetworkRangeInput{
		Name:             plan.Name.ValueStringPointer(),
		RangeType:        (*cato_models.SubnetType)(plan.RangeType.ValueStringPointer()),
		Subnet:           plan.Subnet.ValueStringPointer(),
		LocalIP:          plan.LocalIp.ValueStringPointer(),
		TranslatedSubnet: plan.TranslatedSubnet.ValueStringPointer(),
		Gateway:          plan.Gateway.ValueStringPointer(),
		Vlan:             vlanValue,
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

	// Set Gateway or LocalIP based on range type
	if plan.RangeType == types.StringValue("Routed") {
		if !plan.LocalIp.IsNull() && !plan.LocalIp.IsUnknown() && plan.LocalIp.ValueString() != "" {
			resp.Diagnostics.AddError(
				"Invalid configuration",
				"Configuring LocalIp is only supported for VLAN, Native and Direct range types.",
			)
			return
		}
		input.Gateway = plan.Gateway.ValueStringPointer()
		// Explicitly clear LocalIP for routed ranges
		input.LocalIP = nil
	} else if plan.RangeType == types.StringValue("VLAN") || plan.RangeType == types.StringValue("Direct") {
		if !plan.Gateway.IsNull() && !plan.Gateway.IsUnknown() && plan.Gateway.ValueString() != "" {
			resp.Diagnostics.AddError(
				"Invalid configuration",
				"Configuring gateway is only supported for Routed range types.",
			)
			return
		}
		input.LocalIP = plan.LocalIp.ValueStringPointer()
		// Explicitly clear Gateway for non-routed ranges
		input.Gateway = nil
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
				// Validate that dhcp_microsegmentation is only set to true when dhcp_type is DHCP_RANGE or DHCP_RELAY
				if !dhcpSettingsInput.DhcpMicrosegmentation.IsNull() && !dhcpSettingsInput.DhcpMicrosegmentation.IsUnknown() {
					dhcpType := dhcpSettingsInput.DhcpType.ValueString()
					if dhcpSettingsInput.DhcpMicrosegmentation.ValueBool() == true && dhcpType != "DHCP_RANGE" && dhcpType != "DHCP_RELAY" {
						resp.Diagnostics.AddError(
							"Invalid DHCP Microsegmentation Configuration",
							"dhcp_microsegmentation can only be configured when dhcp_type is set to DHCP_RANGE or DHCP_RELAY",
						)
						return
					}
				}

				// Set dhcpMicrosegmentation for DHCP_RANGE or DHCP_RELAY types
				dhcpType := dhcpSettingsInput.DhcpType.ValueString()
				if (dhcpType == "DHCP_RANGE" || dhcpType == "DHCP_RELAY") && (!dhcpSettingsInput.DhcpMicrosegmentation.IsUnknown()) {
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
	const notFoundMsg = "Invalid network range id: "

	queryRangeResult, err := r.client.catov2.NetworkRange(ctx, r.client.AccountId, networkRangeID)
	if err != nil {
		// Check if error is not found error, if so return (nil, false, nil) to indicate resource should be removed from state without error
		var gqlError *clientv2.ErrorResponse
		if errors.As(err, &gqlError) {
			if (gqlError.GqlErrors != nil) && (len(*gqlError.GqlErrors) > 0) && strings.Contains((*gqlError.GqlErrors)[0].Message, notFoundMsg) {
				return NetworkRange{}, false, nil
			}
		}
		return NetworkRange{}, false, fmt.Errorf("catov2 API NetworkRange error: %w", err)
	}
	if queryRangeResult == nil || queryRangeResult.Site.NetworkRange == nil {
		return NetworkRange{}, false, nil
	}
	responseRange := queryRangeResult.Site.NetworkRange

	// parse DHCP settings - handle both import (when state is null) and refresh cases
	// For Routed and Direct range types, dhcp_settings should always be null
	// The API returns DHCP values for these but they're not valid to submit
	if responseRange.DhcpSettings == nil ||
		responseRange.RangeType == cato_models.SubnetTypeRouted ||
		responseRange.RangeType == cato_models.SubnetTypeDirect {
		state.DhcpSettings = types.ObjectNull(DhcpSettingsAttrTypes)
	} else {
		// Get existing state values if available for preserving computed values
		var stateDhcpSettings DhcpSettings
		hasExistingState := false
		if !state.DhcpSettings.IsNull() && !state.DhcpSettings.IsUnknown() {
			if state.DhcpSettings.As(ctx, &stateDhcpSettings, basetypes.ObjectAsOptions{}).HasError() {
				return NetworkRange{}, false, fmt.Errorf("failed to parse existing dhcpSettings from state for network rangeID '%s'", networkRangeID)
			}
			hasExistingState = true
		}

		dhcpSettings := DhcpSettings{
			DhcpType:              types.StringValue(string(responseRange.DhcpSettings.DhcpType)),
			IpRange:               types.StringNull(),
			RelayGroupId:          types.StringNull(),
			RelayGroupName:        types.StringNull(),
			DhcpMicrosegmentation: types.BoolNull(),
		}

		// Preserve DhcpMicrosegmentation from state if available and not unknown
		if hasExistingState && !stateDhcpSettings.DhcpMicrosegmentation.IsUnknown() {
			dhcpSettings.DhcpMicrosegmentation = stateDhcpSettings.DhcpMicrosegmentation
		}

		// Always set DhcpMicrosegmentation from API response for ALL DHCP types
		// This ensures proper hydration during import when there's no existing state
		dhcpSettings.DhcpMicrosegmentation = types.BoolValue(responseRange.DhcpSettings.DhcpMicrosegmentation)

		// Handle type-specific fields
		switch responseRange.DhcpSettings.DhcpType {
		case cato_models.DhcpTypeDhcpRange:
			dhcpSettings.IpRange = types.StringPointerValue(responseRange.DhcpSettings.IPRange)
		case cato_models.DhcpTypeDhcpRelay:
			if responseRange.DhcpSettings.RelayGroupID == nil { // for DHCP_RELAY, groupID must be set
				return NetworkRange{}, false, fmt.Errorf("dhcpSettings.RelayGroupID not returned by NetworkRange API for rangeID '%s'", networkRangeID)
			}
			// Set BOTH relay_group_id and relay_group_name in state
			// This ensures no drift regardless of which field the user specified in config
			// ConflictsWith validator only applies to configuration, not state
			dhcpSettings.RelayGroupId = types.StringValue(*responseRange.DhcpSettings.RelayGroupID)
			relayGroupName, success, err := checkForDhcpRelayGroup(ctx, r.client, "", *responseRange.DhcpSettings.RelayGroupID)
			if err != nil {
				return NetworkRange{}, false, fmt.Errorf("failed to get dhcpSettings RelayGroup name for network rangeID '%s': %w", networkRangeID, err)
			}
			if !success {
				return NetworkRange{}, false, fmt.Errorf("failed to find dhcpSettings RelayGroup name for network rangeID '%s'", networkRangeID)
			}
			dhcpSettings.RelayGroupName = types.StringValue(relayGroupName)
		}
		// For DHCP_DISABLED and ACCOUNT_DEFAULT, relay fields stay as StringNull() which is correct

		dhcpSettingsObj, diag := types.ObjectValueFrom(ctx, DhcpSettingsAttrTypes, dhcpSettings)
		if diag.HasError() {
			return NetworkRange{}, false, fmt.Errorf("failed to create dhcpSettings object for network rangeID '%s'", networkRangeID)
		}
		state.DhcpSettings = dhcpSettingsObj
	}

	state.Gateway = types.StringPointerValue(responseRange.Gateway)
	state.InternetOnly = types.BoolValue(responseRange.InternetOnly)
	state.MdnsReflector = types.BoolValue(responseRange.MdnsReflector)
	state.LocalIp = types.StringPointerValue(responseRange.LocalIP)
	state.Name = types.StringValue(responseRange.Name)
	state.RangeType = types.StringValue(string(responseRange.RangeType))
	state.Subnet = types.StringValue(responseRange.Subnet)
	state.TranslatedSubnet = types.StringPointerValue(responseRange.TranslatedSubnet)
	state.Vlan = types.Int64PointerValue(responseRange.Vlan)
	if responseRange.RangeType != cato_models.SubnetTypeVlan {
		state.Vlan = types.Int64Null()
	}

	// If SiteId or InterfaceId is not already set (e.g., during import), look it up via entityLookup
	if state.SiteId.IsNull() || state.SiteId.IsUnknown() || state.InterfaceId.IsNull() || state.InterfaceId.IsUnknown() {
		siteId, interfaceName, lookupErr := r.getSiteIdFromNetworkRange(ctx, networkRangeID)
		if lookupErr != nil {
			tflog.Warn(ctx, "Failed to lookup site_id for network range, keeping existing state value", map[string]interface{}{
				"networkRangeID": networkRangeID,
				"error":          lookupErr.Error(),
			})
		} else {
			// Set SiteId if not already set
			if state.SiteId.IsNull() || state.SiteId.IsUnknown() {
				state.SiteId = types.StringValue(siteId)
			}

			// Set InterfaceId and InterfaceIndex if not already set and we have interfaceName
			if (state.InterfaceId.IsNull() || state.InterfaceId.IsUnknown()) && interfaceName != "" {
				// Look up the interface_id using the site_id and interface_name
				interfaceId, interfaceIndex, interfaceErr := getSiteNetworkInterface(ctx, r.client, siteId, "", "", interfaceName)
				if interfaceErr == nil {
					state.InterfaceId = types.StringValue(interfaceId)
					state.InterfaceIndex = types.StringValue(interfaceIndex)
				} else {
					tflog.Warn(ctx, "Failed to lookup interface_id for network range", map[string]interface{}{
						"networkRangeID": networkRangeID,
						"interfaceName":  interfaceName,
						"error":          interfaceErr.Error(),
					})
				}
			}
		}
	}

	return state, true, nil
}

// getSiteIdFromNetworkRange retrieves the site_id and interface info for a network range using entityLookup
func (r *networkRangeResource) getSiteIdFromNetworkRange(ctx context.Context, networkRangeID string) (siteId string, interfaceName string, err error) {
	result, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"), nil, nil, nil, nil, []string{networkRangeID}, nil, nil, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to lookup network range site: %w", err)
	}

	if result == nil || len(result.EntityLookup.GetItems()) == 0 {
		return "", "", fmt.Errorf("network range %s not found in entityLookup", networkRangeID)
	}

	item := result.EntityLookup.GetItems()[0]
	helperFields := item.GetHelperFields()
	if helperFields == nil {
		return "", "", fmt.Errorf("no helperFields returned for network range %s", networkRangeID)
	}

	// Extract siteId from helperFields
	siteId = cast.ToString(helperFields["siteId"])
	if siteId == "" {
		return "", "", fmt.Errorf("siteId not found in helperFields for network range %s", networkRangeID)
	}

	// Extract interfaceName from helperFields
	interfaceName = cast.ToString(helperFields["interfaceName"])

	return siteId, interfaceName, nil
}
