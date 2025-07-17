package provider

import (
	"context"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
					"local_ip": schema.StringAttribute{
						Description: "Site native range local ip",
						Required:    true,
					},
					"translated_subnet": schema.StringAttribute{
						Description: "Site translated native IP range (CIDR)",
						Optional:    true,
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

		// setting input native range DHCP settings
		if !nativeRangeInput.DhcpSettings.IsNull() && !nativeRangeInput.DhcpSettings.IsUnknown() {
			inputUpdateNetworkRange.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{}
			dhcpSettingsInput := DhcpSettings{}
			diags = nativeRangeInput.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			inputUpdateNetworkRange.DhcpSettings.DhcpType = (cato_models.DhcpType)(dhcpSettingsInput.DhcpType.ValueString())
			inputUpdateNetworkRange.DhcpSettings.IPRange = dhcpSettingsInput.IpRange.ValueStringPointer()
			inputUpdateNetworkRange.DhcpSettings.RelayGroupID = dhcpSettingsInput.RelayGroupId.ValueStringPointer()
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

	diags = resp.State.Set(ctx, plan)
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

	// check if site exist, else remove resource
	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.Id.ValueString()}, nil, nil, nil)
	tflog.Warn(ctx, "Read.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// check if site exist before refreshing
	// we should only have one entry since we are filtering on site ID
	if len(querySiteResult.EntityLookup.GetItems()) != 1 {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	// read in the socket site entries

	// type SocketSite struct {
	//     Id             types.String `tfsdk:"id"`
	//     Name           types.String `tfsdk:"name"`
	//     ConnectionType types.String `tfsdk:"connection_type"`
	//     SiteType       types.String `tfsdk:"site_type"`
	//     Description    types.String `tfsdk:"description"`
	//     NativeRange    types.Object `tfsdk:"native_range"`
	//     SiteLocation   types.Object `tfsdk:"site_location"`
	// }
	for _, v := range querySiteResult.EntityLookup.Items {
		// find the socket site entry we need
		if v.Entity.ID == state.Id.ValueString() {
			resp.State.SetAttribute(
				ctx,
				path.Root("id"),
				v.Entity.ID,
			)
			connType := ""
			if val, containsKey := v.GetHelperFields()["connectionType"]; containsKey {
				connType = val.(string)
			}
			siteType := ""
			if val, containsKey := v.GetHelperFields()["type"]; containsKey {
				siteType = val.(string)
			}

			primaryLanInterfaceId := "LAN1"

			if connType == "Socket X1500" {
				primaryLanInterfaceId = "LAN1"
			} else if connType == "Socket X1600" {
				primaryLanInterfaceId = "5"
			} else if connType == "Socket X1600 LTE" {
				primaryLanInterfaceId = "5"
			} else if connType == "Socket X1700" {
				primaryLanInterfaceId = "3"
			}

			querySiteNetworkRangeResult, err := r.client.catov2.EntityLookup(
				ctx,
				r.client.AccountId,
				cato_models.EntityType("networkInterface"),
				nil, // limit
				nil, // from
				&cato_models.EntityInput{
					Type: "site",
					ID:   state.Id.ValueString(),
				}, // parent
				nil, // search
				nil, // entityIds
				nil, nil, nil,
			)
			tflog.Warn(ctx, "Read.EntityLookup/querySiteNetworkRangeResult.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(querySiteResult),
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Catov2 API error",
					err.Error(),
				)
				return
			}

			siteNetRangeApiData := make(map[string]any)
			for _, v := range querySiteNetworkRangeResult.GetEntityLookup().GetItems() {
				if val, containsKey := v.HelperFields["interfaceId"]; containsKey {
					if val == primaryLanInterfaceId {
						siteNetRangeApiData = v.GetHelperFields()
					}
				}
			}

			tflog.Warn(ctx, "Read.querySiteNetworkRangeResult/siteAccountSnapshotApiData.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(siteNetRangeApiData),
			})

			state.Id = types.StringValue(v.Entity.GetID())
			state.Name = types.StringValue(*v.GetEntity().Name)
			// state.ConnectionType = state.ConnectionType
			state.SiteType = types.StringValue(siteType)
			// state.Description = types.StringValue(v.GetDescription())
			state.Description = types.StringValue(v.GetHelperFields()["description"].(string))

			var fromStateNativeRange NativeRange
			if !state.NativeRange.IsNull() && !state.NativeRange.IsUnknown() {
				diags = append(diags, state.NativeRange.As(ctx, &fromStateNativeRange, basetypes.ObjectAsOptions{})...)
				resp.Diagnostics.Append(diags...)
			}

			var stateNativeRange types.Object
			if len(siteNetRangeApiData) > 0 {
				subnet := ""
				if val, ok := siteNetRangeApiData["subnet"].(string); ok {
					subnet = val
				}

				// Handle dhcp_settings properly - either use existing or create null
				var dhcpSettingsValue attr.Value
				if !fromStateNativeRange.DhcpSettings.IsNull() && !fromStateNativeRange.DhcpSettings.IsUnknown() {
					dhcpSettingsValue = fromStateNativeRange.DhcpSettings
				} else {
					dhcpSettingsValue = types.ObjectNull(SiteNativeRangeDhcpResourceAttrTypes)
				}

				stateNativeRange, diags = types.ObjectValue(
					SiteNativeRangeResourceAttrTypes,
					map[string]attr.Value{
						"native_network_range":    types.StringValue(subnet),
						"native_network_range_id": fromStateNativeRange.NativeNetworkRangeId,
						"local_ip":                fromStateNativeRange.LocalIp,
						"translated_subnet":       fromStateNativeRange.TranslatedSubnet,
						"dhcp_settings":           dhcpSettingsValue,
					},
				)
				resp.Diagnostics.Append(diags...)
			} else {
				// Create a null object if no data is available
				stateNativeRange = types.ObjectNull(SiteNativeRangeResourceAttrTypes)
			}
			state.NativeRange = stateNativeRange

			siteAccountSnapshotApiData, err := r.client.catov2.AccountSnapshot(ctx, []string{state.Id.ValueString()}, nil, &r.client.AccountId)
			tflog.Warn(ctx, "Read.AccountSnapshot/siteAccountSnapshotApiData.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(siteAccountSnapshotApiData),
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Catov2 API error",
					err.Error(),
				)
				return
			}

			var fromStateSiteLocation SiteLocation
			if !state.SiteLocation.IsNull() && !state.SiteLocation.IsUnknown() {
				diags = append(diags, state.SiteLocation.As(ctx, &fromStateSiteLocation, basetypes.ObjectAsOptions{})...)
				resp.Diagnostics.Append(diags...)
			}

			var stateSiteLocation types.Object
			if len(siteAccountSnapshotApiData.GetAccountSnapshot().GetSites()) > 0 {
				thisSiteAccountSnapshot := siteAccountSnapshotApiData.GetAccountSnapshot().GetSites()[0]
				stateSiteLocation, diags = types.ObjectValue(
					SiteLocationResourceAttrTypes,
					map[string]attr.Value{
						"country_code": types.StringValue(*thisSiteAccountSnapshot.GetInfoSiteSnapshot().CountryCode),
						"state_code":   fromStateSiteLocation.StateCode,
						"timezone":     fromStateSiteLocation.Timezone,
						"address":      types.StringValue(*thisSiteAccountSnapshot.InfoSiteSnapshot.Address),
						"city":         types.StringValue(*thisSiteAccountSnapshot.InfoSiteSnapshot.CityName),
					},
				)
				resp.Diagnostics.Append(diags...)
			} else {
				// Create a null object if no data is available
				stateSiteLocation = types.ObjectNull(SiteLocationResourceAttrTypes)
			}
			state.SiteLocation = stateSiteLocation
		}
	}

	diags = resp.State.Set(ctx, &state)
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

		// setting input native range DHCP settings
		if !nativeRangeInput.DhcpSettings.IsNull() && !nativeRangeInput.DhcpSettings.IsUnknown() {
			inputUpdateNetworkRange.DhcpSettings = &cato_models.NetworkDhcpSettingsInput{}
			dhcpSettingsInput := DhcpSettings{}
			diags = nativeRangeInput.DhcpSettings.As(ctx, &dhcpSettingsInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			inputUpdateNetworkRange.DhcpSettings.DhcpType = (cato_models.DhcpType)(dhcpSettingsInput.DhcpType.ValueString())
			inputUpdateNetworkRange.DhcpSettings.IPRange = dhcpSettingsInput.IpRange.ValueStringPointer()
			inputUpdateNetworkRange.DhcpSettings.RelayGroupID = dhcpSettingsInput.RelayGroupId.ValueStringPointer()
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

	diags = resp.State.Set(ctx, plan)
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
