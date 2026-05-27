package provider

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

const (
	socketConnectionTypeAWS1500  = "SOCKET_AWS1500"
	socketConnectionTypeAZ1500   = "SOCKET_AZ1500"
	socketConnectionTypeESX1500  = "SOCKET_ESX1500"
	socketConnectionTypeX1600    = "SOCKET_X1600"
	socketConnectionTypeX1600LTE = "SOCKET_X1600_LTE"
	socketConnectionTypeX1700    = "SOCKET_X1700"
	socketConnectionTypeVGXAWS   = "VSOCKET_VGX_AWS"
	socketConnectionTypeVGXAzure = "VSOCKET_VGX_AZURE"
	socketConnectionTypeVGXESX   = "VSOCKET_VGX_ESX"
	socketInterfaceDestTypeLAN   = "LAN"
)

var (
	_ resource.Resource                = &socketSiteResource{}
	_ resource.ResourceWithConfigure   = &socketSiteResource{}
	_ resource.ResourceWithImportState = &socketSiteResource{}
)

func NewSocketSiteResource() resource.Resource {
	return &socketSiteResource{}
}

func stringPointerForOptionalInput(value types.String) *string {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return nil
	}
	return value.ValueStringPointer()
}

type socketSiteResource struct {
	client           *catoClientData
	socketSiteClient SocketSiteClient
}

type SocketSiteClient interface {
	SiteAddSocketSite(ctx context.Context, addSocketSiteInput cato_models.AddSocketSiteInput, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.SiteAddSocketSite, error)
}

func (r *socketSiteResource) getSocketSiteClient() SocketSiteClient {
	if r.socketSiteClient != nil {
		return r.socketSiteClient
	}

	if r.client == nil {
		return nil
	}

	return r.client.catov2
}

func (r *socketSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_site"
}

func (r *socketSiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				Validators:    []validator.String{validators.SiteConnectionTypeValidator{}},
			},
			"site_type": schema.StringAttribute{
				Description: "Site type (https:// api.catonetworks.com/documentation/#definition-SiteType)",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Site description",
				Optional:    true,
			},
			"native_range":  r.schemaNativeRange(),
			"site_location": r.schemaSiteLocation(),
		},
	}
}

func (r *socketSiteResource) schemaNativeRange() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Site lan native range settings",
		Required:    true,
		Validators:  []validator.Object{validators.GetNativeRangeValidator()},
		Attributes: map[string]schema.Attribute{
			"interface_index": schema.StringAttribute{
				Description: "LAN native range interface index, default is LAN1 for SOCKET_X1500 models, " +
					"INT_5 for SOCKET_X1600 and SOCKET_X1600_LTE, and INT_3 for SOCKET_X1700 models",
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{validators.SocketInterfaceIndexValidator{}},
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
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(cato_models.SocketInterfaceDestTypeLan)),
				Validators: []validator.String{validators.SocketInterfaceDestTypeValidator{}},
			},
			"dhcp_settings": r.schemaDhcpSettings(),
		},
	}
}

func (r *socketSiteResource) schemaDhcpSettings() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:   "Site native range DHCP settings (Only releveant for NATIVE and VLAN range_type)",
		Optional:      true,
		Computed:      true,
		PlanModifiers: []planmodifier.Object{planmodifiers.DHCPSettingsModifier()},
		Attributes: map[string]schema.Attribute{
			"dhcp_type": schema.StringAttribute{
				Description:   "Network range dhcp type (https://api.catonetworks.com/documentation/#definition-DhcpType)",
				Required:      true,
				Validators:    []validator.String{validators.DHCPTypeValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ip_range": schema.StringAttribute{
				Description:   "Network range dhcp range (format \"192.168.1.10-192.168.1.20\")",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"relay_group_id": schema.StringAttribute{
				Description:   "Network range dhcp relay group id",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"relay_group_name": schema.StringAttribute{
				Description:   "Network range dhcp relay group name",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dhcp_microsegmentation": schema.BoolAttribute{
				Description: "DHCP Microsegmentation. When enabled, the DHCP server will allocate /32 subnet mask. " +
					"Make sure to enable the proper Firewall rules and enable it with caution, " +
					"as it is not supported on all operating systems; monitor the network closely after activation. " +
					"This setting can only be configured when dhcp_type is set to DHCP_RANGE.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
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

func (r *socketSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *socketSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// TODO: check if this is really necessary
	// Call Read to hydrate the full state from the API
	readReq := resource.ReadRequest{State: resp.State}
	readResp := resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, &readResp)

	// Copy diagnostics and state back to the import response
	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}

// Create cato_socket_site resource
func (r *socketSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tf.SocketSite
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a socket site - API call
	siteID := r.createBasicSocketSite(ctx, &plan, &diags)
	if diags.HasError() {
		return
	}

	// Find Network Range ID for the created site and update the NetworkRange details
	networkRangeID := r.findNativeRange(ctx, siteID, &diags)
	if diags.HasError() {
		return
	}
	r.updateNetworkRange(ctx, &plan, networkRangeID, &diags)
	if diags.HasError() {
		return
	}

	// Move interface to a non-default index if specified in the plan
	defaultInterfaceIndex := tf.InterfaceByConnType[cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())]
	r.assignInterfaceIndex(ctx, defaultInterfaceIndex, &plan, siteID, &diags)
	if diags.HasError() {
		return
	}
	// Update Socket Interface
	r.updateSocketInterface(ctx, &plan, siteID, &diags)
	if diags.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, plan, siteID)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating socket site state", hydrateErr.Error())
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
}

func (r *socketSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tf.SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, state, state.ID.ValueString())
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

// TODO: update
// - update site general details
// - update network range
// - interface index change -> special API
// - update interface?
//
//nolint:gocyclo,funlen,gocritic
func (r *socketSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tf.SocketSite
	var planNativeRange, stateNativeRange tf.NativeRange

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := plan.ID.ValueString()

	// Update general socket site details
	r.updateBasicSocketSite(ctx, &plan, &diags)
	if diags.HasError() {
		return
	}

	// Update network range details
	if utils.CheckErr(&diags, plan.NativeRange.As(ctx, &planNativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	networkRangeID := planNativeRange.NativeNetworkRangeID.ValueString()
	r.updateNetworkRange(ctx, &plan, networkRangeID, &diags)
	if diags.HasError() {
		return
	}

	// Move interface to another index if needed
	if utils.CheckErr(&diags, state.NativeRange.As(ctx, &stateNativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	currentInterfaceIndex := cato_models.SocketInterfaceIDEnum(stateNativeRange.InterfaceIndex.ValueString())
	r.assignInterfaceIndex(ctx, currentInterfaceIndex, &plan, siteID, &diags)
	if diags.HasError() {
		return
	}

	// Update Socket Interface
	r.updateSocketInterface(ctx, &plan, siteID, &diags)
	if diags.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, plan, siteID)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating socket site state", hydrateErr.Error())
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

func (r *socketSiteResource) Update2(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tf.SocketSite
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state to preserve computed values
	var state tf.SocketSite
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input & input to update network range
	inputSiteGeneral := cato_models.UpdateSiteGeneralDetailsInput{
		SiteLocation: &cato_models.UpdateSiteLocationInput{},
	}

	inputUpdateNetworkRange := cato_models.UpdateNetworkRangeInput{
		DhcpSettings: &cato_models.NetworkDhcpSettingsInput{
			DhcpType: cato_models.DhcpType(networkRangeDHCPDisabled),
		},
	}
	inputUpdateSocketInterface := cato_models.UpdateSocketInterfaceInput{}

	// setting input site location
	if !plan.SiteLocation.IsNull() && !plan.SiteLocation.IsUnknown() {
		inputSiteGeneral.SiteLocation = &cato_models.UpdateSiteLocationInput{}
		siteLocationInput := tf.SiteLocation{}
		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// Get state location for comparison
		var stateLocationInput tf.SiteLocation
		if !state.SiteLocation.IsNull() && !state.SiteLocation.IsUnknown() {
			diags = state.SiteLocation.As(ctx, &stateLocationInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
		}

		// Handle city field
		cityPtr := siteLocationInput.City.ValueStringPointer()
		if (cityPtr == nil || *cityPtr == "") &&
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

		// Handle state_code field
		stateCodePtr := siteLocationInput.StateCode.ValueStringPointer()
		if (stateCodePtr == nil || *stateCodePtr == "") &&
			!stateLocationInput.StateCode.IsNull() && !stateLocationInput.StateCode.IsUnknown() && stateLocationInput.StateCode.ValueString() != "" {
			// If state_code had a value in state and is now blank/null in plan,
			// send "" (empty string) to clear the field
			stateCodePtr = nil
		} else if stateCodePtr != nil && *stateCodePtr == "" {
			// Normal case: empty string becomes nil
			stateCodePtr = nil
		}
		inputSiteGeneral.SiteLocation.StateCode = stateCodePtr

		// Handle address field
		addrPtr := siteLocationInput.Address.ValueStringPointer()
		if (addrPtr == nil || *addrPtr == "") &&
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
	var nativeRangeState tf.NativeRange
	diags = state.NativeRange.As(ctx, &nativeRangeState, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	if !plan.NativeRange.IsNull() && !plan.NativeRange.IsUnknown() {
		nativeRangeInput := tf.NativeRange{}
		diags = plan.NativeRange.As(ctx, &nativeRangeInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// Validate that local_ip is within native_network_range
		if !nativeRangeInput.LocalIP.IsNull() && !nativeRangeInput.LocalIP.IsUnknown() &&
			!nativeRangeInput.NativeNetworkRange.IsNull() && !nativeRangeInput.NativeNetworkRange.IsUnknown() {
			localIPStr := nativeRangeInput.LocalIP.ValueString()
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
					fmt.Sprintf("Local IP must be within the Native Range IP. local_ip '%s' is not within "+
						"native_network_range '%s'", localIPStr, subnetStr),
				)
				return
			}
		}

		// Validate LAG configuration
		interfaceDestType := nativeRangeInput.InterfaceDestType.ValueString()
		if interfaceDestType == "" {
			interfaceDestType = socketInterfaceDestTypeLAN // Use default if not specified
		}
		hasLagMinLinks := !nativeRangeInput.LagMinLinks.IsNull() && !nativeRangeInput.LagMinLinks.IsUnknown()

		// Rule 1: If interface_dest_type is LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP, lag_min_links must have a value
		if (interfaceDestType == lanLagMasterDestType || interfaceDestType == lanLagMasterAndVrrpDestType) && !hasLagMinLinks {
			resp.Diagnostics.AddError(
				"Invalid LAG Configuration",
				fmt.Sprintf("When interface_dest_type is %s, lag_min_links must be specified.", interfaceDestType),
			)
			return
		}

		// Rule 2: If lag_min_links has a value, interface_dest_type must be LAN_LAG_MASTER or LAN_LAG_MASTER_AND_VRRP
		if hasLagMinLinks && interfaceDestType != lanLagMasterDestType && interfaceDestType != lanLagMasterAndVrrpDestType {
			resp.Diagnostics.AddError(
				"Invalid LAG Configuration",
				fmt.Sprintf("lag_min_links can only be configured when interface_dest_type is LAN_LAG_MASTER or "+
					"LAN_LAG_MASTER_AND_VRRP, but interface_dest_type is %s.", interfaceDestType),
			)
			return
		}

		// Validate that interface_index can only be set for specific connection types
		// Only validate if user is setting a non-default value
		hasInterfaceIndex := !nativeRangeInput.InterfaceIndex.IsNull() && !nativeRangeInput.InterfaceIndex.IsUnknown() &&
			nativeRangeInput.InterfaceIndex.ValueString() != ""
		if hasInterfaceIndex {
			connType := cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())
			interfaceIndexValue := nativeRangeInput.InterfaceIndex.ValueString()
			defaultInterface, hasDefault := tf.InterfaceByConnType[connType]

			// Only validate if trying to set a non-default interface_index for non-X1600/X1700 types
			if connType != socketConnectionTypeX1600 && connType != socketConnectionTypeX1600LTE && connType != socketConnectionTypeX1700 {
				// Allow if it's the default interface for this connection type
				if !hasDefault || interfaceIndexValue != string(defaultInterface) {
					resp.Diagnostics.AddError(
						"Invalid Interface Index Configuration",
						fmt.Sprintf("interface_index can only be explicitly configured for SOCKET_X1600, SOCKET_X1600_LTE, "+
							"or SOCKET_X1700 connection types, but connection_type is %s", connType),
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
		inputUpdateNetworkRange.LocalIP = nativeRangeInput.LocalIP.ValueStringPointer()
		inputUpdateNetworkRange.TranslatedSubnet = stringPointerForOptionalInput(nativeRangeInput.TranslatedSubnet)
		inputUpdateNetworkRange.MdnsReflector = nativeRangeInput.MdnsReflector.ValueBoolPointer()
		inputUpdateNetworkRange.Vlan = nativeRangeInput.Vlan.ValueInt64Pointer()

		// Handle interface name changes/removals
		if interfaceNameRemovedFromConfig {
			// Removed from local config to reset todefault value of interface index
			inputUpdateSocketInterface.Name = nativeRangeInput.InterfaceIndex.ValueStringPointer()
			tflog.Info(ctx, "inputUpdateSocketInterface.Name=Removed from local config to reset to default value of "+
				"interface index", map[string]interface{}{
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
		if (interfaceDestType == lanLagMasterDestType || interfaceDestType == lanLagMasterAndVrrpDestType) && hasLagMinLinks {
			lagConfig := cato_models.SocketInterfaceLagInput{
				MinLinks: nativeRangeInput.LagMinLinks.ValueInt64(),
			}
			inputUpdateSocketInterface.Lag = &lagConfig
		}

		socketInterfaceLanInput := cato_models.SocketInterfaceLanInput{}
		// Use plan values (nativeRangeInput) to ensure consistency with network range update
		if localIP := nativeRangeInput.LocalIP.ValueStringPointer(); localIP != nil {
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
			dhcpSettingsState := tf.DhcpSettings{}
			diags = nativeRangeInput.DhcpSettings.As(ctx, &dhcpSettingsState, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			dhcpSettingsInput := tf.DhcpSettings{}
			diags = nativeRangeInput.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			inputUpdateNetworkRange.DhcpSettings.DhcpType = cato_models.DhcpType(dhcpSettingsInput.DhcpType.ValueString())
			inputUpdateNetworkRange.DhcpSettings.IPRange = dhcpSettingsInput.IPRange.ValueStringPointer()

			// Validate that relay_group_name or relay_group_id are only set when dhcp_type is DHCP_RELAY
			if dhcpSettingsInput.DhcpType.ValueString() != networkRangeDHCPRelay {
				hasRelayGroupName := !dhcpSettingsInput.RelayGroupName.IsNull() &&
					!dhcpSettingsInput.RelayGroupName.IsUnknown() &&
					dhcpSettingsInput.RelayGroupName.ValueString() != ""
				hasRelayGroupID := !dhcpSettingsInput.RelayGroupID.IsNull() && !dhcpSettingsInput.RelayGroupID.IsUnknown() &&
					dhcpSettingsInput.RelayGroupID.ValueString() != ""
				if hasRelayGroupName || hasRelayGroupID {
					resp.Diagnostics.AddError(
						"Invalid DHCP Relay Configuration",
						fmt.Sprintf("relay_group_name and relay_group_id can only be configured when dhcp_type is DHCP_RELAY, "+
							"but dhcp_type is %s", dhcpSettingsInput.DhcpType.ValueString()),
					)
					return
				}
			}

			// Validate that ip_range is only set when dhcp_type is DHCP_RANGE
			if dhcpSettingsInput.DhcpType.ValueString() != networkRangeDHCPRange {
				hasIPRange := !dhcpSettingsInput.IPRange.IsNull() && !dhcpSettingsInput.IPRange.IsUnknown() &&
					dhcpSettingsInput.IPRange.ValueString() != ""
				if hasIPRange {
					resp.Diagnostics.AddError(
						"Invalid DHCP Range Configuration",
						fmt.Sprintf("ip_range can only be configured when dhcp_type is DHCP_RANGE, but dhcp_type is %s",
							dhcpSettingsInput.DhcpType.ValueString()),
					)
					return
				}
			}

			// Validate that dhcp_microsegmentation is only set to true when dhcp_type is DHCP_RANGE
			if !dhcpSettingsInput.DhcpMicrosegmentation.IsNull() && !dhcpSettingsInput.DhcpMicrosegmentation.IsUnknown() {
				if dhcpSettingsInput.DhcpMicrosegmentation.ValueBool() && dhcpSettingsInput.DhcpType.ValueString() != networkRangeDHCPRange {
					resp.Diagnostics.AddError(
						"Invalid DHCP Microsegmentation Configuration",
						"dhcp_microsegmentation can only be configured when dhcp_type is set to DHCP_RANGE",
					)
					return
				}
			}

			// Only set dhcpMicrosegmentation for DHCP_RANGE type
			if dhcpSettingsInput.DhcpType.ValueString() == networkRangeDHCPRange {
				if !dhcpSettingsInput.DhcpMicrosegmentation.IsNull() && !dhcpSettingsInput.DhcpMicrosegmentation.IsUnknown() {
					inputUpdateNetworkRange.DhcpSettings.DhcpMicrosegmentation = dhcpSettingsInput.DhcpMicrosegmentation.ValueBoolPointer()
				}
			}

			// Validate DHCP relay group configuration when dhcp_type is DHCP_RELAY
			if dhcpSettingsInput.DhcpType.ValueString() == networkRangeDHCPRelay {
				relayGroupName := ""
				relayGroupID := ""

				if !dhcpSettingsInput.RelayGroupName.IsNull() && !dhcpSettingsInput.RelayGroupName.IsUnknown() {
					relayGroupName = dhcpSettingsInput.RelayGroupName.ValueString()
				}
				if !dhcpSettingsInput.RelayGroupID.IsNull() && !dhcpSettingsInput.RelayGroupID.IsUnknown() {
					relayGroupID = dhcpSettingsInput.RelayGroupID.ValueString()
				}

				resolvedRelayGroupID, success, err := checkForDhcpRelayGroup(ctx, r.client, relayGroupName, relayGroupID)
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
				inputUpdateNetworkRange.DhcpSettings.RelayGroupID = &resolvedRelayGroupID
			}
		} else {
			// Configuration has no dhcp_settings block - preserve dhcp_microsegmentation from state if it exists
			if !state.NativeRange.IsNull() && !state.NativeRange.IsUnknown() {
				var stateNativeRange tf.NativeRange
				diags = state.NativeRange.As(ctx, &stateNativeRange, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				if !stateNativeRange.DhcpSettings.IsNull() && !stateNativeRange.DhcpSettings.IsUnknown() {
					var stateDhcpSettings tf.DhcpSettings
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
	siteUpdateSiteGeneralDetailsResponse, err := r.client.catov2.SiteUpdateSiteGeneralDetails(ctx, plan.ID.ValueString(),
		inputSiteGeneral, r.client.AccountId)
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

	// retrieve native range ID
	nativeRange := tf.NativeRange{}
	diags = state.NativeRange.As(ctx, &nativeRange, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.request", map[string]interface{}{
		"request":                          utils.InterfaceToJSONString(inputUpdateNetworkRange),
		"nativeRange.NativeNetworkRangeID": utils.InterfaceToJSONString(nativeRange.NativeNetworkRangeID.ValueString()),
		"r.client.AccountId":               utils.InterfaceToJSONString(r.client.AccountId),
	})
	siteUpdateNetworkRangeResponse, err := r.client.catov2.SiteUpdateNetworkRange(ctx,
		nativeRange.NativeNetworkRangeID.ValueString(), inputUpdateNetworkRange, r.client.AccountId)
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
		nativeRangeCheck := tf.NativeRange{}
		diags = plan.NativeRange.As(ctx, &nativeRangeCheck, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// // Lookup default lan index by socket connection type
		// var defaultLanInterfaceIndex *string
		// if val, ok := tf.InterfaceByConnType[plan.ConnectionType.ValueString()]; ok {
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
			// cato_models.SocketInterfaceIDEnum
			interfaceIndex := nativeRangeCheck.InterfaceIndex.ValueString()
			interfaceName := nativeRangeCheck.InterfaceName.ValueString()
			if interfaceName == "" {
				interfaceName = interfaceIndex
			}
			localIP := nativeRangeCheck.LocalIP.ValueString()
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
				"localIP":            utils.InterfaceToJSONString(localIP),
				"nativeNetworkRange": utils.InterfaceToJSONString(nativeNetworkRange),
				"interfaceDestType":  utils.InterfaceToJSONString(interfaceDestType),
				"lagMinLinks":        utils.InterfaceToJSONString(lagMinLinks),
				"translatedSubnet":   utils.InterfaceToJSONString(translatedSubnet),
			})
			isDefaultPresent, err := r.attemptReassignNativeRangeIndex(ctx,
				cato_models.SocketInterfaceIDEnum(interfaceIndex), interfaceName, localIP, nativeNetworkRange,
				plan.ID.ValueString(), interfaceDestType, lagMinLinks, translatedSubnet)
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
					"Invalid interface configuration, the API to support moving the lan interface index to '"+
						interfaceIndex+"' from the previous index is not yet supported. "+
						"This API is being gradually rolled out.",
				)
				return
			}
			// Interface was reassigned - need to retrieve new native range ID and reapply DHCP settings
			// Get the updated native interface and subnet information
			nativeInterfaceAndSubnet, err := r.getNativeInterfaceAndSubnet(ctx,
				cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString()),
				plan.ID.ValueString(), plan)
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
				"nativeNetworkRangeID": utils.InterfaceToJSONString(nativeInterfaceAndSubnet.NativeNetworkRangeID),
				"r.client.AccountId":   utils.InterfaceToJSONString(r.client.AccountId),
			})
			siteUpdateNetworkRangeResponseAfter, err := r.client.catov2.SiteUpdateNetworkRange(ctx,
				nativeInterfaceAndSubnet.NativeNetworkRangeID, inputUpdateNetworkRange, r.client.AccountId)
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

			// Update the socket interface properties on the new interface (destType, name, lag)
			tflog.Debug(ctx, "Update.SiteUpdateSocketInterface.request (after interface reassignment)", map[string]interface{}{
				"request":        utils.InterfaceToJSONString(inputUpdateSocketInterface),
				"interfaceIndex": utils.InterfaceToJSONString(nativeRangeCheck.InterfaceIndex.ValueString()),
			})
			_, err = r.client.catov2.SiteUpdateSocketInterface(ctx, plan.ID.ValueString(),
				cato_models.SocketInterfaceIDEnum(nativeRangeCheck.InterfaceIndex.ValueString()),
				inputUpdateSocketInterface, r.client.AccountId)
			if err != nil {
				resp.Diagnostics.AddError(
					"Catov2 API SiteUpdateSocketInterface error (after interface reassignment)",
					err.Error(),
				)
				return
			}

			// Clear computed IDs from plan so they can be refreshed during hydration
			var nativeRangePlan tf.NativeRange
			diags := plan.NativeRange.As(ctx, &nativeRangePlan, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			nativeRangePlan.InterfaceID = types.StringNull()
			nativeRangePlan.NativeNetworkLanInterfaceID = types.StringNull()
			nativeRangePlan.NativeNetworkRangeID = types.StringNull()
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
			_, err = r.client.catov2.SiteUpdateSocketInterface(ctx, plan.ID.ValueString(),
				cato_models.SocketInterfaceIDEnum(nativeRangeState.InterfaceIndex.ValueString()),
				inputUpdateSocketInterface, r.client.AccountId)
			if err != nil {
				resp.Diagnostics.AddError(
					"Catov2 API SiteUpdateSocketInterface error",
					err.Error(),
				)
				return
			}
		}
	}

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, plan, plan.ID.ValueString())
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
	var state tf.SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"),
		nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
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
		_, err := r.client.catov2.SiteRemoveSite(ctx, state.ID.ValueString(), r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteRemoveSite error",
				err.Error(),
			)
			return
		}
	}
}

// NativeInterfaceAndSubnetResult contains all the data returned by getNativeInterfaceAndSubnet
type NativeInterfaceAndSubnetResult struct {
	Subnet               string
	NativeNetworkRangeID string
	InterfaceIndex       string
	InterfaceID          string
	InterfaceName        string
	SiteNetRangeAPIData  map[string]any
	NativeRangeObj       tf.NativeRange
}

// getNativeInterfaceAndSubnet retrieves native interface and subnet information
// Returns: NativeInterfaceAndSubnetResult, error
//
//nolint:gocyclo,funlen,gocritic
func (r *socketSiteResource) getNativeInterfaceAndSubnet(ctx context.Context, connType cato_models.SiteConnectionTypeEnum,
	siteID string, state tf.SocketSite,
) (*NativeInterfaceAndSubnetResult, error) {
	var relayGroupName attr.Value = types.StringUnknown()
	siteEntity := &cato_models.EntityInput{Type: "site", ID: siteID}
	zeroInt64 := int64(0)
	if connType != "" {
		// Translate VSOCKET_VGX_* values to SOCKET_* equivalents
		switch connType {
		case socketConnectionTypeVGXAWS:
			connType = socketConnectionTypeAWS1500
		case socketConnectionTypeVGXAzure:
			connType = socketConnectionTypeAZ1500
		case socketConnectionTypeVGXESX:
			connType = socketConnectionTypeESX1500
		}
	} else {
		return nil, fmt.Errorf("connection type is empty")
	}

	// Only assign interfaceIndex if it does not already exist in state
	if _, ok := tf.InterfaceByConnType[connType]; !ok {
		return nil, fmt.Errorf("connection type %s not found in tf.InterfaceByConnType", connType)
	}
	isMicrosegmentationKnown := false
	var nativeRangeObj tf.NativeRange
	// Track if dhcp_settings was actually present in the original state/plan (before deserialization)
	dhcpSettingsWasInOriginalState := false
	if !state.NativeRange.IsNull() && !state.NativeRange.IsUnknown() {
		// Check if dhcp_settings attribute exists and has meaningful content (dhcp_type specified)
		if attrs := state.NativeRange.Attributes(); attrs != nil {
			if dhcpAttr, exists := attrs["dhcp_settings"]; exists && !dhcpAttr.IsNull() && !dhcpAttr.IsUnknown() {
				// Additional check: verify dhcp_type is specified (not null/unknown)
				// This handles the case where an empty dhcp_settings object is passed
				if dhcpObj, ok := dhcpAttr.(types.Object); ok && !dhcpObj.IsNull() {
					dhcpAttrs := dhcpObj.Attributes()
					if dhcpTypeAttr, hasType := dhcpAttrs["dhcp_type"]; hasType && !dhcpTypeAttr.IsNull() && !dhcpTypeAttr.IsUnknown() {
						dhcpSettingsWasInOriginalState = true
					}
					if microSeg, ok := dhcpAttrs["dhcp_microsegmentation"]; ok {
						if !microSeg.IsNull() && !microSeg.IsUnknown() {
							isMicrosegmentationKnown = true
						}
					}
					if grpName, ok := dhcpAttrs["relay_group_name"]; ok {
						relayGroupName = grpName
					}
				}
			}
		}
		state.NativeRange.As(ctx, &nativeRangeObj, basetypes.ObjectAsOptions{})
	}
	// if nativeRangeObj.InterfaceIndex.IsNull() || nativeRangeObj.InterfaceIndex.ValueString() == "" {
	defaultInterfaceIndexByConnType, ok := tf.InterfaceByConnType[connType]
	if !ok {
		return nil, fmt.Errorf("connection type %s not found in tf.InterfaceByConnType", connType)
	}
	// Retrieve default interface range attributes
	queryInterfaceResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId,
		cato_models.EntityType("networkInterface"), &zeroInt64, nil, siteEntity, nil, nil, nil, nil, nil)
	tflog.Warn(ctx, "Read.EntityLookupInterfaceRangeResult.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryInterfaceResult),
	})
	if err != nil {
		return nil, err
	}
	isPresent := false

	// If user-specified interface not found or not specified, check for isDefault flag
	if !isPresent {
		for _, curIint := range queryInterfaceResult.EntityLookup.Items {
			curSiteID := cast.ToString(curIint.HelperFields["siteId"])
			if curSiteID == siteID {
				curInterfaceID := curIint.HelperFields["interfaceId"]
				// curInterfaceName := curIint.HelperFields["interfaceName"]
				// Try to parse the interfaceId as int, otherwise prefix with "INT_"
				if idxInt, err := cast.ToIntE(curInterfaceID); err == nil {
					curInterfaceIDStr := fmt.Sprintf("INT_%d", idxInt)
					curInterfaceID = curInterfaceIDStr
				}
				isDefault := false
				if v, ok := curIint.HelperFields["isDefault"]; ok && v != nil {
					if b, err := cast.ToBoolE(v); err == nil {
						isDefault = b
					}
				}
				if isDefault {
					isPresent = true
					nativeRangeObj.InterfaceIndex = types.StringValue(cast.ToString(curInterfaceID))
					nativeRangeObj.InterfaceID = types.StringValue(curIint.Entity.ID)
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
	}
	// If isDefault flag not found, look for default by index
	// This is due to bug/fix getting gradually pushed out where this default flag may not be present
	// and the following can be purged after this is rolled out
	if !isPresent {
		for _, curIint := range queryInterfaceResult.EntityLookup.Items {
			// find the socket site entry we need
			curSiteID := cast.ToString(curIint.HelperFields["siteId"])
			tflog.Warn(ctx, "for.queryInterfaceResult.EntityLookup.Items | siteID==siteID", map[string]interface{}{
				"siteID":                          siteID,
				"curSiteID":                       curSiteID,
				"defaultInterfaceIndexByConnType": defaultInterfaceIndexByConnType,
				"curInterfaceID":                  curIint.HelperFields["interfaceId"],
				"curInterfaceName":                curIint.HelperFields["interfaceName"],
			})
			if curSiteID == siteID {
				// get current interfaceId from the API and use to map to interface index
				curInterfaceID := curIint.HelperFields["interfaceId"]
				curInterfaceName := curIint.HelperFields["interfaceName"]
				// Try to parse the interfaceId as int, otherwise prefix with "INT_"
				if idxInt, err := cast.ToIntE(curInterfaceID); err == nil {
					curInterfaceIDStr := fmt.Sprintf("INT_%d", idxInt)
					curInterfaceID = curInterfaceIDStr
				}
				tflog.Warn(ctx, "defaultInterfaceIndexByConnType==curInterfaceID", map[string]interface{}{
					"defaultInterfaceIndexByConnType": cast.ToString(defaultInterfaceIndexByConnType),
					"curInterfaceID":                  curInterfaceID,
					"curInterfaceName":                curInterfaceName,
				})
				if cast.ToString(defaultInterfaceIndexByConnType) == curInterfaceID {
					isPresent = true
					if _, err := cast.ToIntE(curInterfaceID); err == nil {
						nativeRangeObj.InterfaceIndex = types.StringValue(cast.ToString(curInterfaceID))
					} else {
						nativeRangeObj.InterfaceIndex = types.StringValue(cast.ToString(curInterfaceID))
					}
					nativeRangeObj.InterfaceIndex = types.StringValue(string(defaultInterfaceIndexByConnType))
					nativeRangeObj.InterfaceID = types.StringValue(curIint.Entity.ID)
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
						"curInterfaceID":                  curInterfaceID,
						"curInterfaceName":                curInterfaceName,
					})
				}
			}
		}
	}
	if !isPresent {
		return nil, fmt.Errorf("site does not contain configuration for default LAN interface index %s for connection type %s, "+
			"please update this site configuratation once either in the cato management application or via API, "+
			"and the correct interface should be marked as default resolving this issue",
			defaultInterfaceIndexByConnType, connType)
	}

	// updates *nativeRangeObj
	if err = r.getNativeRange(ctx, siteID, &nativeRangeObj, isMicrosegmentationKnown, dhcpSettingsWasInOriginalState,
		relayGroupName); err != nil {
		return nil, err
	}
	siteNetRangeAPIData := make(map[string]any)
	nameParts := strings.Split(nativeRangeObj.RangeName.ValueString(), " \\ ")
	siteNetRangeAPIData["rangeName"] = nameParts[len(nameParts)-1] // Store as string, not types.StringValue
	siteNetRangeAPIData["native_network_range_id"] = nativeRangeObj.NativeNetworkRangeID.ValueString()

	tflog.Debug(ctx, "Read.siteNetRangeAPIData", map[string]interface{}{
		"siteNetRangeAPIData": utils.InterfaceToJSONString(siteNetRangeAPIData),
	})

	nativeNetworkRangeID := ""
	if val, ok := siteNetRangeAPIData["native_network_range_id"].(string); ok {
		nativeNetworkRangeID = val
	}

	interfaceIndex := ""
	if !nativeRangeObj.InterfaceIndex.IsNull() && !nativeRangeObj.InterfaceIndex.IsUnknown() {
		interfaceIndex = nativeRangeObj.InterfaceIndex.ValueString()
	}

	interfaceID := ""
	if !nativeRangeObj.InterfaceID.IsNull() && !nativeRangeObj.InterfaceID.IsUnknown() {
		interfaceID = nativeRangeObj.InterfaceID.ValueString()
	}

	return &NativeInterfaceAndSubnetResult{
		Subnet:               nativeRangeObj.NativeNetworkRange.ValueString(),
		NativeNetworkRangeID: nativeNetworkRangeID,
		InterfaceIndex:       interfaceIndex,
		InterfaceName:        nativeRangeObj.InterfaceName.ValueString(),
		InterfaceID:          interfaceID,
		SiteNetRangeAPIData:  siteNetRangeAPIData,
		NativeRangeObj:       nativeRangeObj,
	}, nil
}

//nolint:gocyclo,funlen
func (r *socketSiteResource) getNativeRange(ctx context.Context, siteID string, nativeRangeObj *tf.NativeRange,
	_, isDhcpKnown bool, relayGroupName attr.Value,
) error {
	var microsegmentation attr.Value
	var dhcpRelayGroupID attr.Value = types.StringNull()
	var dhcpRelayGroupName attr.Value = types.StringNull()
	var dhcpIPRange attr.Value = types.StringNull()
	var dhcpSettingsObj = types.ObjectNull(tf.SiteNativeRangeDhcpResourceAttrTypes)

	input := cato_models.NetworkRangeListInput{Site: &cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID}}
	queryNetworkRangeResult, err := r.client.catov2.NetworkRangeList(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "getNativeRange.response", map[string]interface{}{"response": utils.InterfaceToJSONString(queryNetworkRangeResult)})
	if err != nil {
		return err
	}
	for _, netRange := range queryNetworkRangeResult.GetSite().GetNetworkRangeList().GetItems() {
		if netRange.RangeType != cato_models.SubnetTypeNative {
			continue
		}
		// Only process the network range matching the default interface's subnet
		// This prevents overwriting with data from other native-type ranges (e.g., LAN2 with is_native_range=TRUE)
		if netRange.Subnet != nativeRangeObj.NativeNetworkRange.ValueString() {
			continue
		}

		rDhcp := netRange.GetDhcpSettings()
		// Always set microsegmentation from the API response to ensure proper hydration during import
		microsegmentation = types.BoolValue(rDhcp.GetDhcpMicrosegmentation())

		// Only populate DHCP settings if they were already in state OR the API returns a non-default dhcp type
		// This prevents inconsistency when config doesn't include dhcp_settings and API returns default DHCP_DISABLED
		dhcpType := rDhcp.GetDhcpType()
		if isDhcpKnown || (dhcpType != nil && *dhcpType != "" && *dhcpType != cato_models.DhcpTypeDhcpDisabled &&
			*dhcpType != cato_models.DhcpTypeAccountDefault) {
			dType := cato_models.DhcpType("")
			if dhcpType != nil {
				dType = *dhcpType
			}
			if dType == cato_models.DhcpTypeDhcpRelay { // Only hydrate relay group fields if dhcp_type is DHCP_RELAY
				gid := rDhcp.GetRelayGroupID()
				dhcpRelayGroupID = types.StringPointerValue(gid)
				dhcpRelayGroupName = relayGroupName // from state
				if gid != nil {
					rgName, success, err := checkForDhcpRelayGroup(ctx, r.client, "", *gid)
					if err != nil {
						return fmt.Errorf("failed to get dhcpSettings RelayGroup name for id '%s': %w", *gid, err)
					}
					if !success {
						return fmt.Errorf("failed to find dhcpSettings RelayGroup name for id '%s'", *gid)
					}
					dhcpRelayGroupName = types.StringValue(rgName) // from response
				}
			}
			if dType != cato_models.DhcpTypeDhcpDisabled { // Only hydrate ip range if dhcp is enabled
				dhcpIPRange = types.StringPointerValue(rDhcp.GetIPRange())
			}

			obj, dErr := types.ObjectValue(
				tf.SiteNativeRangeDhcpResourceAttrTypes,
				map[string]attr.Value{
					"dhcp_type":              types.StringPointerValue((*string)(rDhcp.GetDhcpType())),
					"ip_range":               dhcpIPRange,
					"relay_group_id":         dhcpRelayGroupID,
					"relay_group_name":       dhcpRelayGroupName,
					"dhcp_microsegmentation": microsegmentation,
				},
			)
			if dErr.HasError() {
				return fmt.Errorf("failed to convert dhcp settings to object: %v", dErr)
			}
			dhcpSettingsObj = obj
		}

		nativeRangeObj.Vlan = types.Int64PointerValue(netRange.Vlan)
		nativeRangeObj.DhcpSettings = dhcpSettingsObj
		nativeRangeObj.Gateway = types.StringPointerValue(netRange.Gateway)
		nativeRangeObj.LocalIP = types.StringPointerValue(netRange.LocalIP)
		nativeRangeObj.MdnsReflector = types.BoolValue(netRange.MdnsReflector)
		nativeRangeObj.RangeName = types.StringValue(netRange.Name)
		nativeRangeObj.RangeID = types.StringValue(netRange.NetworkRangeID)
		nativeRangeObj.NativeNetworkRangeID = types.StringValue(netRange.NetworkRangeID)
		nativeRangeObj.RangeType = types.StringValue(strings.ToUpper(string(netRange.RangeType)))
		nativeRangeObj.TranslatedSubnet = types.StringPointerValue(netRange.TranslatedSubnet)
		nativeRangeObj.Vlan = types.Int64PointerValue(netRange.Vlan)
	}
	return nil
}

func (r *socketSiteResource) attemptReassignNativeRangeIndex(ctx context.Context,
	interfaceIndex cato_models.SocketInterfaceIDEnum, _ string, _ string, _ string, siteID string, _ string,
	_ int64, _ string) (*bool, error) {
	// Attempt to create a lan interface on non default index
	// checking first for the isDefault flag to be present, if not present return false for isDefaultPresent
	// return isValid or error if isDefault flag is not present on entityLookup interface query
	isDefaultPresent := false
	siteEntity := &cato_models.EntityInput{Type: "site", ID: siteID}
	zeroInt64 := int64(0)
	tflog.Warn(ctx, "attemptReassignNativeRangeIndex.EntityLookupInterfaceRange", map[string]interface{}{
		"siteEntity": utils.InterfaceToJSONString(siteEntity),
	})
	queryInterfaceResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId,
		cato_models.EntityType("networkInterface"), &zeroInt64, nil, siteEntity, nil, nil, nil, nil, nil)
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
	// curInterfaceID := nil
	var curInterfaceIndex *string
	for _, curIint := range queryInterfaceResult.EntityLookup.Items {
		tflog.Warn(ctx, "Read.EntityLookupInterfaceRangeResult.curIint.interator", map[string]interface{}{
			"interfaceIndex":    utils.InterfaceToJSONString(interfaceIndex),
			"curInterfaceID":    utils.InterfaceToJSONString(curIint.Entity.ID),
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
					"curInterfaceID":    utils.InterfaceToJSONString(curIint.Entity.ID),
					"curInterfaceIndex": utils.InterfaceToJSONString(curInterfaceIndex),
				})
				isDefaultPresent = true
			}
		}
	}
	if isDefaultPresent {
		exchangeInput := cato_models.ExchangeSocketPortsInput{
			Site: &cato_models.SiteRefInput{
				By:    cato_models.ObjectRefByID,
				Input: siteID,
			},
			FirstInterface: &cato_models.SocketInterfaceRefInput{
				InterfaceID: cato_models.SocketInterfaceIDEnum(*curInterfaceIndex),
			},
			SecondInterface: &cato_models.SocketInterfaceRefInput{
				InterfaceID: interfaceIndex,
			},
		}

		siteExchangeSocketPortsResponse, err := r.client.catov2.SiteExchangeSocketPorts(ctx, r.client.AccountId, exchangeInput)
		if err != nil {
			return nil, err
		}
		tflog.Debug(ctx, "attemptReassignNativeRangeIndex.tmpSiteUpdateSocketInterface.tmp.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(siteExchangeSocketPortsResponse),
		})
	}
	return &isDefaultPresent, err
}

// hydrateSocketSiteState populates the tf.SocketSite state with data from API responses
//
//nolint:gocyclo,funlen
func (r *socketSiteResource) hydrateSocketSiteState(ctx context.Context, state tf.SocketSite, siteID string) (tf.SocketSite, bool, error) {
	// check if site exist, else remove resource
	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"),
		nil, nil, nil, nil, []string{siteID}, nil, nil, nil)
	tflog.Warn(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		return state, false, err
	}

	siteAccountSnapshotAPIData, err := r.client.catov2.AccountSnapshot(ctx, []string{siteID}, nil, &r.client.AccountId)
	tflog.Warn(ctx, "Read.AccountSnapshot/siteAccountSnapshotAPIData.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteAccountSnapshotAPIData),
	})
	if err != nil {
		return state, false, err
	}

	// Get site general details for location data
	siteGeneralDetailsData, err := r.client.catov2.SiteGeneralDetails(ctx,
		cato_models.SiteRefInput{By: cato_models.ObjectRefBy("ID"), Input: siteID}, r.client.AccountId)
	tflog.Debug(ctx, "Read.SiteGeneralDetails.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteGeneralDetailsData),
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
			for _, site := range siteAccountSnapshotAPIData.GetAccountSnapshot().GetSites() {
				if site.ID != nil && *site.ID == siteID {
					thisSiteAccountSnapshot = site
					break
				}
			}

			if thisSiteAccountSnapshot != nil {
				// Get connection type and set state
				connTypeVal := cato_models.SiteConnectionTypeEnum("")
				if val := thisSiteAccountSnapshot.InfoSiteSnapshot.GetConnType(); val != nil {
					connTypeVal = cato_models.SiteConnectionTypeEnum(val.String())
				}

				tflog.Debug(ctx, "Read.connTypeVal", map[string]interface{}{
					"connTypeVal":             connTypeVal,
					"thisSiteConnType":        thisSiteAccountSnapshot.InfoSiteSnapshot.GetConnType(),
					"thisSiteAccountSnapshot": utils.InterfaceToJSONString(thisSiteAccountSnapshot),
				})

				if connTypeVal != "" {
					// Translate VSOCKET_VGX_* values to SOCKET_* equivalents
					switch connTypeVal {
					case socketConnectionTypeVGXAWS:
						connTypeVal = socketConnectionTypeAWS1500
					case socketConnectionTypeVGXAzure:
						connTypeVal = socketConnectionTypeAZ1500
					case socketConnectionTypeVGXESX:
						connTypeVal = socketConnectionTypeESX1500
					}
					state.ConnectionType = types.StringValue(string(connTypeVal))
				} else {
					state.ConnectionType = types.StringNull()
				}

				// Get native interface and subnet information
				nativeInterfaceAndSubnet, err := r.getNativeInterfaceAndSubnet(ctx, connTypeVal, siteID, state)
				if err != nil {
					return state, false, err
				}

				tflog.Debug(ctx, "Read.getNativeInterfaceAndSubnet.result", map[string]interface{}{
					"result": utils.InterfaceToJSONString(nativeInterfaceAndSubnet),
				})

				// Extract values from result struct
				subnet := nativeInterfaceAndSubnet.Subnet
				resultNativeNetworkRangeID := nativeInterfaceAndSubnet.NativeNetworkRangeID
				siteNetRangeAPIData := nativeInterfaceAndSubnet.SiteNetRangeAPIData
				nativeRangeObj := nativeInterfaceAndSubnet.NativeRangeObj

				siteType := ""
				if val, containsKey := v.GetHelperFields()["type"]; containsKey {
					siteType = val.(string)
				}
				state.ID = types.StringValue(v.Entity.GetID())
				state.Name = types.StringValue(*v.GetEntity().Name)
				// ConnectionType is already set above in the switch statement
				state.SiteType = types.StringValue(siteType)
				descriptionStr := v.GetHelperFields()["description"].(string)
				if descriptionStr != "" {
					state.Description = types.StringValue(descriptionStr)
				}

				// Use translated_subnet value directly from API
				// Always hydrate from API to ensure proper state during import and refresh
				translatedSubnetValue := nativeRangeObj.TranslatedSubnet
				tflog.Debug(ctx, "Using translated_subnet from API", map[string]interface{}{
					"translated_subnet": translatedSubnetValue.ValueString(),
					"native_subnet":     subnet,
				})

				// All values from API via nativeRangeObj
				var stateNativeRange types.Object
				stateNativeRange, _ = types.ObjectValue(
					tf.SiteNativeRangeResourceAttrTypes,
					map[string]attr.Value{
						"interface_index":                 nativeRangeObj.InterfaceIndex,
						"interface_id":                    nativeRangeObj.InterfaceID,
						"interface_name":                  nativeRangeObj.InterfaceName,
						"native_network_lan_interface_id": nativeRangeObj.InterfaceID,
						"native_network_range":            types.StringValue(subnet),
						"native_network_range_id":         types.StringValue(resultNativeNetworkRangeID),
						"range_name": func() attr.Value {
							if val, ok := siteNetRangeAPIData["rangeName"].(string); ok && val != "" {
								return types.StringValue(val)
							}
							return types.StringValue("Native Range")
						}(),
						"range_id": func() attr.Value {
							if val, ok := siteNetRangeAPIData["rangeId"].(string); ok && val != "" {
								return types.StringValue(val)
							}
							return types.StringNull()
						}(),
						"local_ip":            nativeRangeObj.LocalIP,
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

				// Extract location data from siteGeneralDetails API response
				siteLocation := siteGeneralDetailsData.GetSite().GetSiteGeneralDetails().GetSiteLocation()

				tflog.Debug(ctx, "Read.SiteGeneralDetails.siteLocation", map[string]interface{}{
					"siteLocation": utils.InterfaceToJSONString(siteLocation),
				})

				// Build state location from siteGeneralDetails response
				// Always hydrate from API to ensure proper state during import and refresh
				stateSiteLocation, _ = types.ObjectValue(
					tf.SiteLocationResourceAttrTypes,
					map[string]attr.Value{
						"country_code": types.StringValue(siteLocation.GetCountryCode()),
						"state_code": func() types.String {
							if siteLocation.GetStateCode() != nil && *siteLocation.GetStateCode() != "" {
								return types.StringValue(*siteLocation.GetStateCode())
							}
							return types.StringNull()
						}(),
						"timezone": types.StringValue(siteLocation.GetTimezone()),
						"address": func() types.String {
							if siteLocation.GetAddress() != nil && *siteLocation.GetAddress() != "" {
								return types.StringValue(*siteLocation.GetAddress())
							}
							return types.StringNull()
						}(),
						"city": func() types.String {
							if siteLocation.GetCityName() != nil && *siteLocation.GetCityName() != "" {
								return types.StringValue(*siteLocation.GetCityName())
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

func (r *socketSiteResource) prepareSiteLocation(ctx context.Context, location types.Object, diags *diag.Diagnostics,
) *cato_models.AddSiteLocationInput {
	if !utils.HasValue(location) {
		return nil
	}

	var tfLocation tf.SiteLocation
	if utils.CheckErr(diags, location.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.AddSiteLocationInput{
		Address:     parse.KnownStringPointer(tfLocation.Address),
		City:        parse.KnownStringPointer(tfLocation.City),
		CountryCode: tfLocation.CountryCode.ValueString(),
		StateCode:   parse.KnownStringPointer(tfLocation.StateCode),
		Timezone:    tfLocation.Timezone.ValueString(),
	}
}

// prepareSocketSiteInput constructs the API input for SiteAddSocketSite() from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.
func (r *socketSiteResource) prepareSocketSiteInput(ctx context.Context, plan *tf.SocketSite, diags *diag.Diagnostics,
) *cato_models.AddSocketSiteInput {
	var tfNativeRange tf.NativeRange
	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return nil
	}

	input := &cato_models.AddSocketSiteInput{
		ConnectionType:     cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString()),
		Description:        parse.KnownStringPointer(plan.Description),
		Name:               plan.Name.ValueString(),
		NativeNetworkRange: tfNativeRange.NativeNetworkRange.ValueString(),
		SiteLocation:       r.prepareSiteLocation(ctx, plan.SiteLocation, diags),
		SiteType:           cato_models.SiteType(plan.SiteType.ValueString()),
		TranslatedSubnet:   parse.KnownStringPointer(tfNativeRange.TranslatedSubnet),
		Vlan:               (*scalars.Vlan)(parse.KnownInt64Pointer(tfNativeRange.Vlan)),
	}
	return input
}

// prepareDhcpSettings constructs the NetworkDhcpSettingsInput part of API input for SiteUpdateNetworkRange()
// from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.
func (r *socketSiteResource) prepareDhcpSettings(ctx context.Context, dhcpSettings types.Object, diags *diag.Diagnostics,
) *cato_models.NetworkDhcpSettingsInput {
	var tfDhcpSettings tf.DhcpSettings

	if !utils.HasValue(dhcpSettings) {
		return nil
	}

	if utils.CheckErr(diags, dhcpSettings.As(ctx, &tfDhcpSettings, basetypes.ObjectAsOptions{})) {
		return nil
	}
	if !utils.HasValue(tfDhcpSettings.DhcpType) {
		return nil
	}

	dhcpType := cato_models.DhcpType(tfDhcpSettings.DhcpType.ValueString())

	input := &cato_models.NetworkDhcpSettingsInput{
		DhcpType: dhcpType,
		IPRange:  parse.KnownStringPointer(tfDhcpSettings.IPRange),
	}

	// if dhcp type is DHCP_RELAY and we don't have relayGroupID (just Name), fetch the relayGroupID
	if dhcpType == cato_models.DhcpTypeDhcpRelay && (!utils.HasValue(tfDhcpSettings.RelayGroupID)) {
		if !utils.HasValue(tfDhcpSettings.RelayGroupName) {
			diags.AddError("Missing DHCP relay group name", "DHCP settings of type DHCP_RELAY require a relay group name to be specified.")
			return nil
		}
		relayGroupID := r.getDhcpRelayGroupID(ctx, r.client, tfDhcpSettings.RelayGroupName.ValueString(), diags)
		if diags.HasError() {
			return nil
		}
		input.RelayGroupID = &relayGroupID
	}

	// Microsegmentation is only relevant for DHCP range
	if input.DhcpType == cato_models.DhcpTypeDhcpRange {
		input.DhcpMicrosegmentation = parse.KnownBoolPointer(tfDhcpSettings.DhcpMicrosegmentation)
	}

	return input
}

// prepareNetworkRangeInput constructs the API input for SiteUpdateNetworkRange() from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.
func (r *socketSiteResource) prepareNetworkRangeInput(ctx context.Context, plan *tf.SocketSite, diags *diag.Diagnostics,
) *cato_models.UpdateNetworkRangeInput {
	var tfNativeRange tf.NativeRange
	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return nil
	}
	input := &cato_models.UpdateNetworkRangeInput{
		Subnet:           parse.KnownStringPointer(tfNativeRange.NativeNetworkRange),
		TranslatedSubnet: parse.KnownStringPointer(tfNativeRange.TranslatedSubnet),
		LocalIP:          parse.KnownStringPointer(tfNativeRange.LocalIP),
		MdnsReflector:    parse.KnownBoolPointer(tfNativeRange.MdnsReflector),
		Vlan:             parse.KnownInt64Pointer(tfNativeRange.Vlan),
		DhcpSettings:     r.prepareDhcpSettings(ctx, tfNativeRange.DhcpSettings, diags),
		//    AzureFloatingIP *string `json:"azureFloatingIp,omitempty"` TODO: implement when AZURE HA support is added
	}

	return input
}

// prepareSocketInterfaceInput prepares inputs for SiteUpdateSocketInterface API,
// on error updates diags
func (r *socketSiteResource) prepareSocketInterfaceInput(ctx context.Context, plan *tf.SocketSite, diags *diag.Diagnostics,
) (*cato_models.UpdateSocketInterfaceInput, cato_models.SocketInterfaceIDEnum) {
	var (
		tfNativeRange     tf.NativeRange
		lagLinksDestTypes = []cato_models.SocketInterfaceDestType{ // set lag input for these dest types
			cato_models.SocketInterfaceDestTypeLanLagMaster,
			cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp,
		}
		lanDestTypes = []cato_models.SocketInterfaceDestType{ // set lan input for these dest types
			cato_models.SocketInterfaceDestTypeLan,
			cato_models.SocketInterfaceDestTypeLanAndHa,
			cato_models.SocketInterfaceDestTypeVrrpAndLan,
			cato_models.SocketInterfaceDestTypeLanLagMaster,
			cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp,
		}
	)

	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return nil, ""
	}

	interfaceDestType := cato_models.SocketInterfaceDestType(tfNativeRange.InterfaceDestType.ValueString())

	// Determine interface ifaceName with the following precedence: InterfaceName > InterfaceIndex > default based on connection type
	ifaceName := parse.KnownStringPointer(tfNativeRange.InterfaceName)
	if ifaceName == nil {
		ifaceName = parse.KnownStringPointer(tfNativeRange.InterfaceIndex)
	}
	if ifaceName == nil {
		ifaceName = ptr(string(tf.InterfaceByConnType[cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())]))
	}

	input := &cato_models.UpdateSocketInterfaceInput{
		DestType: interfaceDestType,
		Name:     ifaceName,
	}

	// MinLinks for LAG
	if utils.HasValue(tfNativeRange.LagMinLinks) && slices.Contains(lagLinksDestTypes, interfaceDestType) {
		input.Lag = &cato_models.SocketInterfaceLagInput{MinLinks: tfNativeRange.LagMinLinks.ValueInt64()}
	}

	// LAN input
	if slices.Contains(lanDestTypes, interfaceDestType) {
		input.Lan = &cato_models.SocketInterfaceLanInput{
			LocalIP:          tfNativeRange.LocalIP.ValueString(),
			Subnet:           tfNativeRange.NativeNetworkRange.ValueString(),
			TranslatedSubnet: tfNativeRange.TranslatedSubnet.ValueStringPointer(),
		}
	}

	// get interface index if configured, otherwise use a default based on connection type
	interfaceIndex := cato_models.SocketInterfaceIDEnum(tfNativeRange.InterfaceIndex.ValueString())
	if interfaceIndex == "" {
		interfaceIndex = tf.InterfaceByConnType[cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())]
	}

	return input, interfaceIndex
}

// getDhcpRelayGroupID looks up the DHCP relay group ID based on the provided relay group name.
func (r *socketSiteResource) getDhcpRelayGroupID(ctx context.Context, client *catoClientData,
	relayGroupName string, diags *diag.Diagnostics,
) (groupNameOrID string) {
	// Lookup and validate the DHCP relay group exists
	dhcpRelayGroupResult, err := client.catov2.EntityLookupMinimal(ctx, client.AccountId, cato_models.EntityTypeDhcpRelayGroup,
		nil, nil, nil, nil, nil)
	if err != nil {
		diags.AddError("Failed to lookup DHCP relay group", fmt.Sprintf("An error was encountered when looking up DHCP relay group: %v", err))
		return ""
	}

	// Check if the specified relay group exists
	for _, item := range dhcpRelayGroupResult.EntityLookup.Items {
		if namePtr := item.Entity.GetName(); namePtr != nil && *namePtr == relayGroupName {
			return item.Entity.GetID()
		}
	}

	// Relay group not found
	diags.AddError("Failed to lookup DHCP relay group", fmt.Sprintf("DHCP relay group: '%s' not found", relayGroupName))
	return ""
}

// createBasicSocketSite creates a socket site and returns the new site ID.
// It does NOT update native range or interface details.
func (r *socketSiteResource) createBasicSocketSite(ctx context.Context, plan *tf.SocketSite, diags *diag.Diagnostics) (siteID string) {
	socketSiteInput := r.prepareSocketSiteInput(ctx, plan, diags)
	if diags.HasError() {
		return ""
	}
	socketSite, err := r.getSocketSiteClient().SiteAddSocketSite(ctx, *socketSiteInput, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteAddSocketSite error", err.Error())
		return ""
	}
	siteID = socketSite.Site.AddSocketSite.GetSiteID()
	if siteID == "" {
		diags.AddError("Catov2 API SiteAddSocketSite error", "empty site ID returned from API")
		return ""
	}
	return siteID
}

// updateBasicSocketSite updates a socket site
// It does NOT update native range or interface details.
func (r *socketSiteResource) updateBasicSocketSite(ctx context.Context, plan *tf.SocketSite, diags *diag.Diagnostics) {
	input := cato_models.UpdateSiteGeneralDetailsInput{
		Description: parse.KnownStringPointer(plan.Description),
		Name:        plan.Name.ValueStringPointer(),
		SiteType:    (*cato_models.SiteType)(plan.SiteType.ValueStringPointer()),
	}
	if utils.HasValue(plan.SiteLocation) {
		var tfLocation tf.SiteLocation
		if utils.CheckErr(diags, plan.SiteLocation.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})) {
			return
		}
		siteLocationInput := &cato_models.UpdateSiteLocationInput{
			Address:     parse.KnownStringPointer(tfLocation.Address),
			CityName:    parse.KnownStringPointer(tfLocation.City),
			CountryCode: tfLocation.CountryCode.ValueStringPointer(),
			StateCode:   parse.KnownStringPointer(tfLocation.StateCode),
			Timezone:    tfLocation.Timezone.ValueStringPointer(),
		}
		input.SiteLocation = siteLocationInput
	}

	_, err := r.client.catov2.SiteUpdateSiteGeneralDetails(ctx, plan.ID.ValueString(), input, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateSiteGeneralDetails error", err.Error())
		return
	}
}

// findNativeRange looks up the native network range for the given site ID and returns its ID.
func (r *socketSiteResource) findNativeRange(ctx context.Context, siteID string, diags *diag.Diagnostics) (networkRangeID string) {
	input := cato_models.NetworkRangeListInput{Site: &cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID}}
	queryNetworkRangeResult, err := r.client.catov2.NetworkRangeList(ctx, r.client.AccountId, input)
	if err != nil {
		diags.AddError("Catov2 API NetworkRangeList error", err.Error())
		return
	}
	for _, netRange := range queryNetworkRangeResult.GetSite().GetNetworkRangeList().GetItems() {
		if netRange.RangeType == cato_models.SubnetTypeNative {
			return netRange.GetNetworkRangeID()
		}
	}
	diags.AddError("Native network range not found", fmt.Sprintf("no native network range found for site ID %s", siteID))
	return ""
}

// assignInterfaceIndex attempts to assign the desired interface index for the native range by calling SiteExchangeSocketPorts API.
// TODO: pass oldID and newID, instead of using default
func (r *socketSiteResource) assignInterfaceIndex(ctx context.Context, currentInterfaceIndex cato_models.SocketInterfaceIDEnum,
	plan *tf.SocketSite, siteID string, diags *diag.Diagnostics,
) {
	var tfNativeRange tf.NativeRange
	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	if !utils.HasValue(tfNativeRange.InterfaceIndex) {
		return
	}

	desiredInterfaceIndex := cato_models.SocketInterfaceIDEnum(tfNativeRange.InterfaceIndex.ValueString())

	if desiredInterfaceIndex == currentInterfaceIndex {
		return // No need to move the interface if the desired index is the default
	}

	exchangeInput := cato_models.ExchangeSocketPortsInput{
		Site:            &cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID},
		FirstInterface:  &cato_models.SocketInterfaceRefInput{InterfaceID: currentInterfaceIndex},
		SecondInterface: &cato_models.SocketInterfaceRefInput{InterfaceID: desiredInterfaceIndex},
	}

	siteExchangeSocketPortsResponse, err := r.client.catov2.SiteExchangeSocketPorts(ctx, r.client.AccountId, exchangeInput)
	if err != nil {
		diags.AddError("Catov2 API SiteExchangeSocketPorts error", err.Error())
		return
	}
	tflog.Debug(ctx, "attemptReassignNativeRangeIndex.tmpSiteUpdateSocketInterface.tmp.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteExchangeSocketPortsResponse),
	})
}

// updateNativeRange updates the native network range details by calling SiteUpdateNetworkRange API.
func (r *socketSiteResource) updateNetworkRange(ctx context.Context, plan *tf.SocketSite, networkRangeID string, diags *diag.Diagnostics) {
	networkRangeInput := r.prepareNetworkRangeInput(ctx, plan, diags)
	if diags.HasError() {
		return
	}

	// Update native network range
	_, err := r.client.catov2.SiteUpdateNetworkRange(ctx, networkRangeID, *networkRangeInput, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateNetworkRange error", err.Error())
	}
}

// updateSocketInterface updates the socket interface details.
func (r *socketSiteResource) updateSocketInterface(ctx context.Context, plan *tf.SocketSite, siteID string, diags *diag.Diagnostics) {
	socketIfaceInput, interfaceID := r.prepareSocketInterfaceInput(ctx, plan, diags)
	if diags.HasError() {
		return
	}

	// Update socket interface
	_, err := r.client.catov2.SiteUpdateSocketInterface(ctx, siteID, interfaceID, *socketIfaceInput, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateSocketInterface error", err.Error())
	}
}
