package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
					},
					"relay_group_name": schema.StringAttribute{
						Description: "Network range dhcp relay group name",
						Optional:    true,
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

	// Validate that the site ID exists
	if !plan.SiteId.IsNull() && !plan.SiteId.IsUnknown() {
		_, err := getSiteNetworkInterfaceById(ctx, r.client, plan.SiteId.ValueString(), plan.InterfaceId.ValueString())
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
	if plan.RangeType == types.StringValue("Routed") && !plan.InternetOnly.IsNull() &&
		!plan.MdnsReflector.IsNull() && plan.InternetOnly.ValueBool() == true {
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
				input.DhcpSettings.RelayGroupID = dhcpSettingsInput.RelayGroupId.ValueStringPointer()
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
				}
			}
		}
	}

	tflog.Debug(ctx, "Create.SiteAddNetworkRange.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
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

	// Validate that the site ID exists
	if !plan.SiteId.IsNull() && !plan.SiteId.IsUnknown() {
		_, err := getSiteNetworkInterfaceById(ctx, r.client, plan.SiteId.ValueString(), plan.InterfaceId.ValueString())
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
				input.DhcpSettings.RelayGroupID = dhcpSettingsInput.RelayGroupId.ValueStringPointer()
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
		resp.Diagnostics.AddError(
			"Cato API error",
			err.Error(),
		)
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
		if curRange.Entity.ID == state.Id.ValueString() {
			state.Id = types.StringValue(curRange.Entity.ID)

			// Field missing, add when supported by API
			// if gatewayVal, ok := curRange.HelperFields["gateway"]; ok {
			// 	state.Gateway = types.StringValue(cast.ToString(gatewayVal))
			// }

			// Field missing, add when supported by API
			// if gatewayVal, ok := curRange.HelperFields["gateway"]; ok {
			// 	state.Gateway = types.StringValue(cast.ToString(gatewayVal))
			// }

			// Field missing, add when supported by API
			// if internetOnlyVal, ok := curRange.HelperFields["internetOnly"]; ok {
			// 	state.InternetOnly = types.BoolValue(cast.ToBool(internetOnlyVal))
			// }

			// Field missing, add when supported by API
			// if localIpVal, ok := curRange.HelperFields["local_ip"]; ok {
			// 	state.LocalIp = types.StringValue(cast.ToString(localIpVal))
			// }

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

			// Field missing, add when supported by API
			// if rangeTypeVal, ok := curRange.HelperFields["range_type"]; ok {
			// 	state.RangeType = types.StringValue(cast.ToString(rangeTypeVal))
			// }

			if siteIdVal, ok := curRange.HelperFields["siteId"]; ok {
				state.SiteId = types.StringValue(cast.ToString(siteIdVal))
			}
			if subnetVal, ok := curRange.HelperFields["subnet"]; ok {
				state.Subnet = types.StringValue(cast.ToString(subnetVal))
			}

			// Field missing, add when supported by API
			// if translatedSubnetVal, ok := curRange.HelperFields["translated_subnet"]; ok {
			// 	state.TranslatedSubnet = types.StringValue(cast.ToString(translatedSubnetVal))
			// }

			// if vlanVal, ok := curRange.HelperFields["vlan"]; ok {
			// 	state.Vlan = types.Int64Value(cast.ToInt64(vlanVal))
			// }

			// Only populate DHCP settings if they were configured in the plan or if this is for VLAN/Native range types
			// This prevents drift detection when DHCP settings are not configured
			if !state.DhcpSettings.IsNull() && !state.DhcpSettings.IsUnknown() {
				// DHCP settings are configured in the plan, so we should populate them
				var dhcpSettings DhcpSettings

				// Get current plan values to preserve them
				diags := state.DhcpSettings.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})
				if diags.HasError() {
					// If we can't read the plan, initialize with defaults
					dhcpSettings.DhcpType = types.StringNull()
					dhcpSettings.IpRange = types.StringNull()
					dhcpSettings.RelayGroupId = types.StringNull()
					dhcpSettings.RelayGroupName = types.StringNull()
					dhcpSettings.DhcpMicrosegmentation = types.BoolValue(false)
				}

				// Update microsegmentation from API if available
				if microsegmentationVal, ok := curRange.HelperFields["microsegmentation"]; ok {
					dhcpSettings.DhcpMicrosegmentation = types.BoolValue(cast.ToBool(microsegmentationVal))
				}

				// Convert DhcpSettings struct to types.Object with all attributes
				dhcpSettingsObject, dErr := types.ObjectValueFrom(ctx, DhcpSettingsAttrTypes, dhcpSettings)
				if dErr != nil {
					return state, false, fmt.Errorf("failed to convert dhcp settings to object: %w", dErr)
				}

				state.DhcpSettings = dhcpSettingsObject
			} else {
				// DHCP settings are not configured in the plan, keep them null to avoid drift
				// This handles both regular reads and import operations
				state.DhcpSettings = types.ObjectNull(DhcpSettingsAttrTypes)
			}
		}
	}
	return state, true, nil
}
