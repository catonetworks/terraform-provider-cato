package provider

import (
	"context"
	"net"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
				Description: "Site native range settings",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"native_network_range": schema.StringAttribute{
						Description: "Site native IP range (CIDR)",
						Required:    true,
					},
					"native_network_range_id": schema.StringAttribute{
						Description: "Site native IP range ID (for update purpose)",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"vlan": schema.Int64Attribute{
						Description: "VLAN ID for the site native range (optional)",
						Optional:    true,
					},
					"mdns_reflector": schema.BoolAttribute{
						Description: "Site native range mDNS reflector. When enabled, the Socket functions as an mDNS gateway, it relays mDNS requests and response between all enabled subnets.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"local_ip": schema.StringAttribute{
						Description: "Site native range local ip",
						Required:    true,
					},
					"translated_subnet": schema.StringAttribute{
						Description: "Site translated native IP range (CIDR)",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
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
							},
							"dhcp_microsegmentation": schema.BoolAttribute{
								Description: "DHCP Microsegmentation. When enabled, the DHCP server will allocate /32 subnet mask. Make sure to enable the proper Firewall rules and enable it with caution, as it is not supported on all operating systems; monitor the network closely after activation.",
								Optional:    true,
								Computed:    true,
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
					},
					"address": schema.StringAttribute{
						Description: "Optionnal address",
						Optional:    true,
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

	// setting input & input to update network range
	input := cato_models.AddSocketSiteInput{}
	inputUpdateNetworkRange := cato_models.UpdateNetworkRangeInput{}

	// setting input site location
	if !plan.SiteLocation.IsNull() && !plan.SiteLocation.IsUnknown() {
		input.SiteLocation = &cato_models.AddSiteLocationInput{}
		siteLocationInput := SiteLocation{}
		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.SiteLocation.Address = siteLocationInput.Address.ValueStringPointer()
		input.SiteLocation.City = siteLocationInput.City.ValueStringPointer()
		input.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueString()
		input.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()
		input.SiteLocation.Timezone = siteLocationInput.Timezone.ValueString()
	}

	// setting input native range
	if !plan.NativeRange.IsNull() && !plan.NativeRange.IsUnknown() {
		nativeRangeInput := NativeRange{}
		diags = plan.NativeRange.As(ctx, &nativeRangeInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.NativeNetworkRange = nativeRangeInput.NativeNetworkRange.ValueString()
		input.TranslatedSubnet = nativeRangeInput.TranslatedSubnet.ValueStringPointer()

		inputUpdateNetworkRange.Subnet = nativeRangeInput.NativeNetworkRange.ValueStringPointer()
		inputUpdateNetworkRange.TranslatedSubnet = nativeRangeInput.TranslatedSubnet.ValueStringPointer()
		inputUpdateNetworkRange.LocalIP = nativeRangeInput.LocalIp.ValueStringPointer()
		inputUpdateNetworkRange.MdnsReflector = nativeRangeInput.MdnsReflector.ValueBoolPointer()
		inputUpdateNetworkRange.Vlan = nativeRangeInput.Vlan.ValueInt64Pointer()

		// setting input native range DHCP settings
		if !nativeRangeInput.DhcpSettings.IsNull() && !nativeRangeInput.DhcpSettings.IsUnknown() {
			inputUpdateNetworkRange.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{}
			dhcpSettingsInput := DhcpSettings{}
			diags = nativeRangeInput.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			inputUpdateNetworkRange.DhcpSettings.DhcpType = (cato_models.DhcpType)(dhcpSettingsInput.DhcpType.ValueString())
			inputUpdateNetworkRange.DhcpSettings.IPRange = dhcpSettingsInput.IpRange.ValueStringPointer()
			inputUpdateNetworkRange.DhcpSettings.RelayGroupID = dhcpSettingsInput.RelayGroupId.ValueStringPointer()
			inputUpdateNetworkRange.DhcpSettings.DhcpMicrosegmentation = dhcpSettingsInput.DhcpMicrosegmentation.ValueBoolPointer()
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

	// retrieving native-network range ID to update native range
	entityParent := cato_models.EntityInput{
		ID:   socketSite.Site.AddSocketSite.GetSiteID(),
		Type: (cato_models.EntityType)("site"),
	}

	siteRangeEntities, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"), nil, nil, &entityParent, nil, nil, nil, nil, nil)
	tflog.Debug(ctx, "Create.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteRangeEntities),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	var networkRangeEntity cato_go_sdk.EntityLookup_EntityLookup_Items_Entity
	for _, item := range siteRangeEntities.EntityLookup.Items {
		splitName := strings.Split(*item.Entity.Name, " \\ ")
		if splitName[2] == "Native Range" {
			networkRangeEntity = item.Entity
		}
	}

	tflog.Debug(ctx, "Create.SiteUpdateNetworkRange.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateNetworkRange),
	})
	_, err = r.client.catov2.SiteUpdateNetworkRange(ctx, networkRangeEntity.GetID(), inputUpdateNetworkRange, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateNetworkRange error",
			err.Error(),
		)
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, plan)
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
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), types.StringValue(socketSite.Site.AddSocketSite.GetSiteID()))
	// overiding state with native network range id
	resp.State.SetAttribute(ctx, path.Root("native_range").AtName("native_network_range_id"), networkRangeEntity.ID)
}

func (r *socketSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, state)
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

	// setting input & input to update network range
	inputSiteGeneral := cato_models.UpdateSiteGeneralDetailsInput{
		SiteLocation: &cato_models.UpdateSiteLocationInput{},
	}

	inputUpdateNetworkRange := cato_models.UpdateNetworkRangeInput{
		DhcpSettings: &cato_models.NetworkDhcpSettingsInput{
			DhcpType: (cato_models.DhcpType)("DHCP_DISABLED"),
		},
	}

	// setting input site location
	if !plan.SiteLocation.IsNull() && !plan.SiteLocation.IsUnknown() {
		inputSiteGeneral.SiteLocation = &cato_models.UpdateSiteLocationInput{}
		siteLocationInput := SiteLocation{}
		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputSiteGeneral.SiteLocation.CityName = siteLocationInput.City.ValueStringPointer()
		inputSiteGeneral.SiteLocation.Address = siteLocationInput.Address.ValueStringPointer()
		inputSiteGeneral.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueStringPointer()
		inputSiteGeneral.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()
		inputSiteGeneral.SiteLocation.Timezone = siteLocationInput.Timezone.ValueStringPointer()
	}

	// setting input native range
	if !plan.NativeRange.IsNull() && !plan.NativeRange.IsUnknown() {
		nativeRangeInput := NativeRange{}
		diags = plan.NativeRange.As(ctx, &nativeRangeInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputUpdateNetworkRange.TranslatedSubnet = nativeRangeInput.TranslatedSubnet.ValueStringPointer()
		inputUpdateNetworkRange.Subnet = nativeRangeInput.NativeNetworkRange.ValueStringPointer()
		inputUpdateNetworkRange.TranslatedSubnet = nativeRangeInput.TranslatedSubnet.ValueStringPointer()
		inputUpdateNetworkRange.LocalIP = nativeRangeInput.LocalIp.ValueStringPointer()
		inputUpdateNetworkRange.MdnsReflector = nativeRangeInput.MdnsReflector.ValueBoolPointer()
		inputUpdateNetworkRange.Vlan = nativeRangeInput.Vlan.ValueInt64Pointer()

		// setting input native range DHCP settings
		if !nativeRangeInput.DhcpSettings.IsNull() && !nativeRangeInput.DhcpSettings.IsUnknown() {
			inputUpdateNetworkRange.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{}
			dhcpSettingsInput := DhcpSettings{}
			diags = nativeRangeInput.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			inputUpdateNetworkRange.DhcpSettings.DhcpType = (cato_models.DhcpType)(dhcpSettingsInput.DhcpType.ValueString())
			inputUpdateNetworkRange.DhcpSettings.IPRange = dhcpSettingsInput.IpRange.ValueStringPointer()
			inputUpdateNetworkRange.DhcpSettings.RelayGroupID = dhcpSettingsInput.RelayGroupId.ValueStringPointer()
			inputUpdateNetworkRange.DhcpSettings.DhcpMicrosegmentation = dhcpSettingsInput.DhcpMicrosegmentation.ValueBoolPointer()
		}
	}

	// setting input other attributes
	inputUpdateNetworkRange.Name = plan.Name.ValueStringPointer()
	inputSiteGeneral.Name = plan.Name.ValueStringPointer()
	inputSiteGeneral.SiteType = (*cato_models.SiteType)(plan.SiteType.ValueStringPointer())
	inputSiteGeneral.Description = plan.Description.ValueStringPointer()

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
	diags = plan.NativeRange.As(ctx, &nativeRange, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateNetworkRange),
	})
	_, err = r.client.catov2.SiteUpdateNetworkRange(ctx, nativeRange.NativeNetworkRangeId.ValueString(), inputUpdateNetworkRange, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateNetworkRange error",
			err.Error(),
		)
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists, hydrateErr := r.hydrateSocketSiteState(ctx, plan)
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

// hydrateSocketSiteState populates the SocketSite state with data from API responses
func (r *socketSiteResource) hydrateSocketSiteState(ctx context.Context, state SocketSite) (SocketSite, bool, error) {
	// check if site exist, else remove resource
	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.Id.ValueString()}, nil, nil, nil)
	tflog.Warn(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		return state, false, err
	}

	siteAccountSnapshotApiData, err := r.client.catov2.AccountSnapshot(ctx, []string{state.Id.ValueString()}, nil, &r.client.AccountId)
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
		if v.Entity.ID == state.Id.ValueString() {
			var stateSiteLocation types.Object
			if len(siteAccountSnapshotApiData.GetAccountSnapshot().GetSites()) > 0 {
				thisSiteAccountSnapshot := siteAccountSnapshotApiData.GetAccountSnapshot().GetSites()[0]
				connTypeVal := ""
				if val := siteAccountSnapshotApiData.GetAccountSnapshot().GetSites()[0].InfoSiteSnapshot.GetConnType(); val != nil {
					connTypeVal = val.String()
				}
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

				siteType := ""
				if val, containsKey := v.GetHelperFields()["type"]; containsKey {
					siteType = val.(string)
				}

				// check if site exist, else remove resource
				siteEntity := &cato_models.EntityInput{Type: "site", ID: state.Id.ValueString()}
				querySiteRangeResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"), nil, nil, siteEntity, nil, nil, nil, nil, nil)
				tflog.Warn(ctx, "Read.EntityLookupSiteRangeResult.response", map[string]interface{}{
					"response": utils.InterfaceToJSONString(querySiteRangeResult),
				})
				if err != nil {
					return state, false, err
				}

				siteNetRangeApiData := make(map[string]any)
				for _, v := range querySiteRangeResult.GetEntityLookup().GetItems() {
					if v.Entity.Name != nil {
						nameVal := *v.Entity.Name
						splitName := strings.Split(nameVal, " \\ ")
						tflog.Debug(ctx, "Read.querySiteRangeResult", map[string]interface{}{
							"nameVal":     utils.InterfaceToJSONString(nameVal),
							"splitName":   utils.InterfaceToJSONString(splitName),
							"rangeName":   utils.InterfaceToJSONString(splitName[len(splitName)-1]),
							"rangeName==": utils.InterfaceToJSONString(splitName[len(splitName)-1] == "Native Range"),
						})
						if len(splitName) > 0 {
							rangeName := splitName[len(splitName)-1]
							if rangeName == "Native Range" {
								siteNetRangeApiData = v.GetHelperFields()
								// Pull ID from entity attributes
								siteNetRangeApiData["native_network_range_id"] = v.Entity.GetID()
							}
						}
					}
				}
				tflog.Debug(ctx, "Read.siteNetRangeApiData", map[string]interface{}{
					"siteNetRangeApiData": utils.InterfaceToJSONString(siteNetRangeApiData),
				})
				state.Id = types.StringValue(v.Entity.GetID())
				state.Name = types.StringValue(*v.GetEntity().Name)
				// ConnectionType is already set above in the switch statement
				state.SiteType = types.StringValue(siteType)
				state.Description = types.StringValue(v.GetHelperFields()["description"].(string))

				var fromStateNativeRange NativeRange
				if !state.NativeRange.IsNull() && !state.NativeRange.IsUnknown() {
					state.NativeRange.As(ctx, &fromStateNativeRange, basetypes.ObjectAsOptions{})
				}

				var stateNativeRange types.Object
				subnet := ""
				if val, ok := siteNetRangeApiData["subnet"].(string); ok {
					subnet = val
				}
				mdnsReflector := false
				if val, ok := siteNetRangeApiData["mdns_reflector"].(bool); ok {
					mdnsReflector = val
				}
				microsegmentation := false
				if val, ok := siteNetRangeApiData["microsegmentation"].(bool); ok {
					microsegmentation = val
				}
				var vlan attr.Value = types.Int64Null()
				if val, ok := siteNetRangeApiData["vlanTag"].(string); ok {
					if vlanInt, err := cast.ToInt64E(val); err == nil {
						vlan = types.Int64Value(vlanInt)
					}
				}
				dhcpTyleValue := types.StringNull()
				ipRangeVal := types.StringNull()
				relayGroupIdVal := types.StringNull()
				if !fromStateNativeRange.DhcpSettings.IsNull() && !fromStateNativeRange.DhcpSettings.IsUnknown() {
					// Configuration has dhcp_settings, so preserve them
					var dhcpSettings DhcpSettings
					fromStateNativeRange.DhcpSettings.As(ctx, &dhcpSettings, basetypes.ObjectAsOptions{})
					if dhcpSettings.DhcpType.ValueString() != "" {
						dhcpTyleValue = types.StringValue(dhcpSettings.DhcpType.ValueString())
					}
					if dhcpSettings.IpRange.ValueString() != "" {
						ipRangeVal = types.StringValue(dhcpSettings.IpRange.ValueString())
					}
					if dhcpSettings.RelayGroupId.ValueString() != "" {
						relayGroupIdVal = types.StringValue(dhcpSettings.RelayGroupId.ValueString())
					}
				}

				// Handle dhcp_settings properly - only create if they exist in config
				var dhcpSettingsValue attr.Value
				dhcpSettingsValue, _ = types.ObjectValue(
					SiteNativeRangeDhcpResourceAttrTypes,
					map[string]attr.Value{
						"dhcp_type":              dhcpTyleValue,
						"ip_range":               ipRangeVal,
						"relay_group_id":         relayGroupIdVal,
						"dhcp_microsegmentation": types.BoolValue(microsegmentation),
					},
				)

				// Calculate local IP based on connection type and subnet
				calculatedLocalIP := calculateLocalIP(ctx, subnet, connTypeVal)
				localIPValue := types.StringValue(calculatedLocalIP)
				if calculatedLocalIP == "" {
					// Fallback to state value if calculation fails
					localIPValue = fromStateNativeRange.LocalIp
				}

				stateNativeRange, _ = types.ObjectValue(
					SiteNativeRangeResourceAttrTypes,
					map[string]attr.Value{
						"native_network_range": types.StringValue(subnet),
						"native_network_range_id": func() attr.Value {
							if val, ok := siteNetRangeApiData["native_network_range_id"].(string); ok {
								return types.StringValue(val)
							}
							return types.StringNull()
						}(),
						"local_ip":          localIPValue,
						"translated_subnet": fromStateNativeRange.TranslatedSubnet,
						"vlan":              vlan,
						"mdns_reflector":    types.BoolValue(mdnsReflector),
						"dhcp_settings":     dhcpSettingsValue,
					},
				)
				// } else {
				// 	// Create a null object if no data is available
				// 	stateNativeRange = types.ObjectNull(SiteNativeRangeResourceAttrTypes)
				// }
				state.NativeRange = stateNativeRange

				var fromStateSiteLocation SiteLocation
				if !state.SiteLocation.IsNull() && !state.SiteLocation.IsUnknown() {
					state.SiteLocation.As(ctx, &fromStateSiteLocation, basetypes.ObjectAsOptions{})
				}

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

				// If we resolved a timezone and there's no timezone in state, use the resolved one
				timezoneValue := fromStateSiteLocation.Timezone
				if resolvedLocation.Timezone != "" && (fromStateSiteLocation.Timezone.IsNull() || fromStateSiteLocation.Timezone.ValueString() == "") {
					timezoneValue = types.StringValue(resolvedLocation.Timezone)
				}
				// If we resolved a state code and there's no state code in state, use the resolved one
				srtateCodeValue := fromStateSiteLocation.StateCode
				if resolvedLocation.StateCode != "" && (fromStateSiteLocation.StateCode.IsNull() || fromStateSiteLocation.StateCode.ValueString() == "") {
					srtateCodeValue = types.StringValue(resolvedLocation.StateCode)
				}

				stateSiteLocation, _ = types.ObjectValue(
					SiteLocationResourceAttrTypes,
					map[string]attr.Value{
						"country_code": types.StringValue(*thisSiteAccountSnapshot.GetInfoSiteSnapshot().CountryCode),
						"state_code":   srtateCodeValue,
						"timezone":     timezoneValue,
						"address": func() types.String {
							if thisSiteAccountSnapshot.InfoSiteSnapshot.Address != nil && *thisSiteAccountSnapshot.InfoSiteSnapshot.Address != "" {
								return types.StringValue(*thisSiteAccountSnapshot.InfoSiteSnapshot.Address)
							}
							return types.StringNull()
						}(),
						"city": types.StringValue(*thisSiteAccountSnapshot.InfoSiteSnapshot.CityName),
					},
				)
			} else {
				// Create a null object if no data is available
				stateSiteLocation = types.ObjectNull(SiteLocationResourceAttrTypes)
			}
			state.SiteLocation = stateSiteLocation
		}
	}
	return state, true, nil
}
