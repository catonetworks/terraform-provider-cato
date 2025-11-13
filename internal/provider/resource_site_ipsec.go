package provider

import (
	"context"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &siteIpsecResource{}
	_ resource.ResourceWithConfigure   = &siteIpsecResource{}
	_ resource.ResourceWithImportState = &siteIpsecResource{}
)

func NewSiteIpsecResource() resource.Resource {
	return &siteIpsecResource{}
}

type siteIpsecResource struct {
	client *catoClientData
}

func (r *siteIpsecResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_site"
}

func (r *siteIpsecResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for Ipsec Site",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Ipsec Site Name",
				Required:    true,
			},
			"site_type": schema.StringAttribute{
				Description: "Valid values are: BRANCH, HEADQUARTERS, CLOUD_DC, and DATACENTER.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "description",
				Required:    true,
			},
			"native_network_range": schema.StringAttribute{
				Description: "NativeNetworkRange",
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
			"interface_id": schema.StringAttribute{
				Description: "IPSec interface ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_location": schema.SingleNestedAttribute{
				Description: "SiteLocation",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"country_code": schema.StringAttribute{
						Description: "Country Code",
						Required:    true,
					},
					"state_code": schema.StringAttribute{
						Description: "State Code",
						Required:    false,
						Optional:    true,
					},
					"timezone": schema.StringAttribute{
						Description: "Timezone",
						Required:    true,
					},
					"address": schema.StringAttribute{
						Description: "Address",
						Required:    false,
						Optional:    true,
					},
					"city": schema.StringAttribute{
						Description: "City",
						Required:    false,
						Optional:    true,
					},
				},
			},
			"ipsec": schema.SingleNestedAttribute{
				Description: "IPSec Configuration",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"site_id": schema.StringAttribute{
						Description: "Site Identifier for Ipsec Site",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"primary": schema.SingleNestedAttribute{
						Description: "primary",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"destination_type": schema.StringAttribute{
								Description: "destinationtype",
								Required:    false,
								Optional:    true,
							},
							"public_cato_ip_id": schema.StringAttribute{
								Description: "publiccatoipid",
								Required:    false,
								Optional:    true,
							},
							"pop_location_id": schema.StringAttribute{
								Description: "poplocationid",
								Required:    false,
								Optional:    true,
							},
							"tunnels": schema.ListNestedAttribute{
								Description: "tunnels",
								Required:    false,
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"tunnel_id": schema.StringAttribute{
											Description: "tunnel ID",
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"public_site_ip": schema.StringAttribute{
											Description: "publicsiteip",
											Required:    false,
											Optional:    true,
										},
										"private_cato_ip": schema.StringAttribute{
											Description: "privatecatoip",
											Optional:    true,
										},
										"private_site_ip": schema.StringAttribute{
											Description: "privatesiteip",
											Optional:    true,
										},
										"psk": schema.StringAttribute{
											Description: "psk",
											Required:    true,
										},
										"last_mile_bw": schema.SingleNestedAttribute{
											Description: "lastmilebw",
											Required:    false,
											Optional:    true,
											Attributes: map[string]schema.Attribute{
												"downstream": schema.Int64Attribute{
													Description: "Downstream",
													Required:    true,
												},
												"upstream": schema.Int64Attribute{
													Description: "upstream",
													Required:    true,
												},
												"downstream_mbps_precision": schema.Float64Attribute{
													Description: "downstreamMbpsPrecision",
													Required:    false,
													Optional:    true,
												},
												"upstream_mbps_precision": schema.Float64Attribute{
													Description: "upstreamMbpsPrecision",
													Required:    false,
													Optional:    true,
												},
											},
										},
									},
								},
							},
						},
					},
					"secondary": schema.SingleNestedAttribute{
						Description: "secondary",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"destination_type": schema.StringAttribute{
								Description: "destinationtype",
								Required:    false,
								Optional:    true,
							},
							"public_cato_ip_id": schema.StringAttribute{
								Description: "publiccatoipid",
								Required:    false,
								Optional:    true,
							},
							"pop_location_id": schema.StringAttribute{
								Description: "poplocationid",
								Required:    false,
								Optional:    true,
							},
							"tunnels": schema.ListNestedAttribute{
								Description: "tunnels",
								Required:    false,
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"tunnel_id": schema.StringAttribute{
											Description: "tunnel ID",
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"public_site_ip": schema.StringAttribute{
											Description: "publicsiteip",
											Required:    false,
											Optional:    true,
										},
										"private_cato_ip": schema.StringAttribute{
											Description: "privatecatoip",
											Optional:    true,
										},
										"private_site_ip": schema.StringAttribute{
											Description: "privatesiteip",
											Optional:    true,
										},
										"psk": schema.StringAttribute{
											Description: "psk",
											Required:    true,
										},
										"last_mile_bw": schema.SingleNestedAttribute{
											Description: "lastmilebw",
											Required:    false,
											Optional:    true,
											Attributes: map[string]schema.Attribute{
												"downstream": schema.Int64Attribute{
													Description: "Downstream",
													Required:    true,
												},
												"upstream": schema.Int64Attribute{
													Description: "upstream",
													Required:    true,
												},
												"downstream_mbps_precision": schema.Float64Attribute{
													Description: "downstreamMbpsPrecision",
													Required:    false,
													Optional:    true,
												},
												"upstream_mbps_precision": schema.Float64Attribute{
													Description: "upstreamMbpsPrecision",
													Required:    false,
													Optional:    true,
												},
											},
										},
									},
								},
							},
						},
					},
					"connection_mode": schema.StringAttribute{
						Description: "Connection mode for IPSec tunnel. Valid values: RESPONDER_ONLY, BIDIRECTIONAL",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("RESPONDER_ONLY", "BIDIRECTIONAL"),
						},
					},
					"identification_type": schema.StringAttribute{
						Description: "Identification type for IPSec. Only applicable when connection_mode is RESPONDER_ONLY. Valid values: IPV4, FQDN, EMAIL, KEY_ID",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("IPV4", "FQDN", "EMAIL", "KEY_ID"),
						},
						PlanModifiers: []planmodifier.String{
							planmodifiers.IdentificationTypeValidator(),
						},
					},
					"init_message": schema.SingleNestedAttribute{
						Description: "IKE initialization message configuration",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"cipher": schema.StringAttribute{
								Description: "Cipher algorithm. Valid values: NONE, AUTOMATIC, AES_CBC_128, AES_CBC_256, AES_GCM_128, AES_GCM_256, DES3_CBC",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "AUTOMATIC", "AES_CBC_128", "AES_CBC_256", "AES_GCM_128", "AES_GCM_256", "DES3_CBC"),
								},
								Default: stringdefault.StaticString("AUTOMATIC"),
							},
							"dh_group": schema.StringAttribute{
								Description: "Diffie-Hellman group. Valid values: NONE, AUTOMATIC, DH_2_MODP1024, DH_5_MODP1536, DH_14_MODP2048, DH_15_MODP3072, DH_16_MODP4096, DH_19_ECP256, DH_20_ECP384, DH_21_ECP521",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "AUTOMATIC", "DH_2_MODP1024", "DH_5_MODP1536", "DH_14_MODP2048", "DH_15_MODP3072", "DH_16_MODP4096", "DH_19_ECP256", "DH_20_ECP384", "DH_21_ECP521"),
								},
								Default: stringdefault.StaticString("AUTOMATIC"),
							},
							"integrity": schema.StringAttribute{
								Description: "Integrity algorithm. Valid values: NONE, AUTOMATIC, MD5, SHA1, SHA256, SHA384, SHA512",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "AUTOMATIC", "MD5", "SHA1", "SHA256", "SHA384", "SHA512"),
								},
								Default: stringdefault.StaticString("AUTOMATIC"),
							},
							"prf": schema.StringAttribute{
								Description: "Pseudo-Random Function. Valid values: NONE, AUTOMATIC, MD5, SHA1, SHA256, SHA384, SHA512",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "AUTOMATIC", "MD5", "SHA1", "SHA256", "SHA384", "SHA512"),
								},
								Default: stringdefault.StaticString("AUTOMATIC"),
							},
						},
					},
					"auth_message": schema.SingleNestedAttribute{
						Description: "IKE authentication message configuration",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"cipher": schema.StringAttribute{
								Description: "Cipher algorithm. Valid values: NONE, AUTOMATIC, AES_CBC_128, AES_CBC_256, AES_GCM_128, AES_GCM_256, DES3_CBC",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "AUTOMATIC", "AES_CBC_128", "AES_CBC_256", "AES_GCM_128", "AES_GCM_256", "DES3_CBC"),
								},
								Default: stringdefault.StaticString("AUTOMATIC"),
							},
							"dh_group": schema.StringAttribute{
								Description: "Diffie-Hellman group. Valid values: NONE, AUTOMATIC, DH_2_MODP1024, DH_5_MODP1536, DH_14_MODP2048, DH_15_MODP3072, DH_16_MODP4096, DH_19_ECP256, DH_20_ECP384, DH_21_ECP521",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "AUTOMATIC", "DH_2_MODP1024", "DH_5_MODP1536", "DH_14_MODP2048", "DH_15_MODP3072", "DH_16_MODP4096", "DH_19_ECP256", "DH_20_ECP384", "DH_21_ECP521"),
								},
								Default: stringdefault.StaticString("AUTOMATIC"),
							},
							"integrity": schema.StringAttribute{
								Description: "Integrity algorithm. Valid values: NONE, AUTOMATIC, MD5, SHA1, SHA256, SHA384, SHA512",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "AUTOMATIC", "MD5", "SHA1", "SHA256", "SHA384", "SHA512"),
								},
								Default: stringdefault.StaticString("AUTOMATIC"),
							},
							"prf": schema.StringAttribute{
								Description: "Pseudo-Random Function. Valid values: NONE, AUTOMATIC, MD5, SHA1, SHA256, SHA384, SHA512",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("NONE", "AUTOMATIC", "MD5", "SHA1", "SHA256", "SHA384", "SHA512"),
								},
								Default: stringdefault.StaticString("AUTOMATIC"),
							},
						},
					},
					"network_ranges": schema.ListAttribute{
						Description: "List of network ranges (e.g., ['servers:192.168.11.0/24', 'desktops:192.169.11.0/24'])",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (r *siteIpsecResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *siteIpsecResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *siteIpsecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan SiteIpsecIkeV2
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate API input for site creation
	input, diags := hydrateAddIpsecIkeV2Site(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2Site.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	ipsecSite, err := r.client.catov2.SiteAddIpsecIkeV2Site(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2Site.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ipsecSite),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error",
			err.Error(),
		)
		return
	}
	// overiding state with socket site id
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), types.StringValue(ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID()))

	// retrieving native-network range ID to update native range
	entityParent := cato_models.EntityInput{
		ID:   ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID(),
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

	siteID := ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID()

	// Hydrate API input for IPSec general details
	inputIpsecGeneralDetails, diags := hydrateUpdateIpsecIkeV2SiteGeneralDetails(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.SiteUpdateIpsecIkeV2SiteGeneralDetails.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputIpsecGeneralDetails),
	})
	ipsecGeneralDetailsResponse, err := r.client.catov2.SiteUpdateIpsecIkeV2SiteGeneralDetails(ctx, siteID, inputIpsecGeneralDetails, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteUpdateIpsecIkeV2SiteGeneralDetails.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ipsecGeneralDetailsResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateIpsecIkeV2SiteGeneralDetails error",
			err.Error(),
		)
		return
	}

	// Hydrate API input for tunnels
	tunnelInputs, diags := hydrateAddIpsecIkeV2SiteTunnels(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2SiteTunnels.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(tunnelInputs.add),
	})
	tunnelData, err_ipsec := r.client.catov2.SiteAddIpsecIkeV2SiteTunnels(ctx, siteID, tunnelInputs.add, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2SiteTunnels.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(tunnelData),
	})
	if err_ipsec != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteAddIpsecIkeV2SiteTunnels",
			err_ipsec.Error(),
		)
		return
	}

	// create types to support multiple primary and secondary tunnels
	tunnelsPrimaryData := []*cato_go_sdk.SiteAddIpsecIkeV2SiteTunnels_Site_AddIpsecIkeV2SiteTunnels_PrimaryAddIpsecIkeV2SiteTunnelsPayload_Tunnels{}
	tunnelsSecondaryData := []*cato_go_sdk.SiteAddIpsecIkeV2SiteTunnels_Site_AddIpsecIkeV2SiteTunnels_SecondaryAddIpsecIkeV2SiteTunnelsPayload_Tunnels{}

	if len(tunnelData.Site.GetAddIpsecIkeV2SiteTunnels().PrimaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()) > 0 {
		tunnelsPrimaryData = tunnelData.Site.GetAddIpsecIkeV2SiteTunnels().PrimaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()
	}

	if len(tunnelData.Site.GetAddIpsecIkeV2SiteTunnels().SecondaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()) > 0 {
		tunnelsSecondaryData = tunnelData.Site.GetAddIpsecIkeV2SiteTunnels().SecondaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()
	}

	// Hydrate the state with API data to ensure consistency
	hydratedState, _, hydrateErr := r.hydrateIpsecSiteState(ctx, plan, siteID)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating IPSec site state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// supports multiple primary ipsec tunnels
	if tunnelInputs.add.Primary != nil {
		for x := 0; x < len(tunnelsPrimaryData); x++ {
			resp.State.SetAttribute(ctx, path.Root("ipsec").AtName("primary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id"), tunnelsPrimaryData[x].GetTunnelIDAddIpsecIkeV2SiteTunnelPayload().String())
		}
	}

	// supports multiple secondary ipsec tunnels
	if tunnelInputs.add.Secondary != nil {
		for x := 0; x < len(tunnelsSecondaryData); x++ {
			resp.State.SetAttribute(ctx, path.Root("ipsec").AtName("secondary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id"), tunnelsSecondaryData[x].GetTunnelIDAddIpsecIkeV2SiteTunnelPayload().String())
		}
	}

	// Override computed fields that hydrate might not get from AccountSnapshot
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), types.StringValue(siteID))
	resp.State.SetAttribute(ctx, path.Empty().AtName("native_network_range_id"), networkRangeEntity.ID)
	resp.State.SetAttribute(ctx, path.Root("ipsec").AtName("site_id"), siteID)
}

func (r *siteIpsecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state SiteIpsecIkeV2
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate the state with API data
	hydratedState, siteExists, err := r.hydrateIpsecSiteState(ctx, state, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating IPSec site state",
			err.Error(),
		)
		return
	}

	// Check if site was found, else remove resource
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

func (r *siteIpsecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan SiteIpsecIkeV2
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input & input to update network range
	inputSiteGeneral := cato_models.UpdateSiteGeneralDetailsInput{
		SiteLocation: &cato_models.UpdateSiteLocationInput{},
	}

	inputUpdateNetworkRange := cato_models.UpdateNetworkRangeInput{}

	// setting input site location
	if !plan.SiteLocation.IsNull() {
		inputSiteGeneral.SiteLocation = &cato_models.UpdateSiteLocationInput{}
		siteLocationInput := SiteLocation{}
		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputSiteGeneral.SiteLocation.Address = siteLocationInput.Address.ValueStringPointer()
		inputSiteGeneral.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueStringPointer()
		inputSiteGeneral.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()
		inputSiteGeneral.SiteLocation.Timezone = siteLocationInput.Timezone.ValueStringPointer()
		// inputSiteGeneral.SiteLocation.City = siteLocationInput.City.ValueStringPointer()
	}

	inputUpdateNetworkRange.Subnet = plan.NativeNetworkRange.ValueStringPointer()
	inputUpdateNetworkRange.TranslatedSubnet = plan.NativeNetworkRange.ValueStringPointer()
	inputSiteGeneral.Name = plan.Name.ValueStringPointer()
	inputSiteGeneral.SiteType = (*cato_models.SiteType)(plan.SiteType.ValueStringPointer())
	inputSiteGeneral.Description = plan.Description.ValueStringPointer()

	tflog.Debug(ctx, "Update.SiteUpdateSiteGeneralDetails.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputSiteGeneral),
	})
	inputSiteGeneralResponse, err := r.client.catov2.SiteUpdateSiteGeneralDetails(ctx, plan.ID.ValueString(), inputSiteGeneral, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteUpdateSiteGeneralDetails", map[string]interface{}{
		"response": utils.InterfaceToJSONString(inputSiteGeneralResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateSiteGeneralDetails error",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputUpdateNetworkRange),
	})
	// TODO, look at why response object does not resolve
	_, err = r.client.catov2.SiteUpdateNetworkRange(ctx, plan.NativeNetworkRangeId.ValueString(), inputUpdateNetworkRange, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateNetworkRange error",
			err.Error(),
		)
		return
	}

	// Hydrate API input for IPSec general details
	inputIpsecGeneralDetails, diags := hydrateUpdateIpsecIkeV2SiteGeneralDetails(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get site ID from plan
	planIPSec := AddIpsecIkeV2SiteTunnelsInput{}
	diags = plan.IPSec.As(ctx, &planIPSec, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	siteID := planIPSec.SiteId.ValueString()

	tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteGeneralDetails.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputIpsecGeneralDetails),
	})
	ipsecGeneralDetailsResponse, err := r.client.catov2.SiteUpdateIpsecIkeV2SiteGeneralDetails(ctx, siteID, inputIpsecGeneralDetails, r.client.AccountId)
	tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteGeneralDetails.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ipsecGeneralDetailsResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateIpsecIkeV2SiteGeneralDetails error",
			err.Error(),
		)
		return
	}

	// Hydrate the state with API data to ensure consistency
	hydratedState, siteExists, hydrateErr := r.hydrateIpsecSiteState(ctx, plan, plan.ID.ValueString())
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating IPSec site state",
			hydrateErr.Error(),
		)
		return
	}

	// Check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found after update, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	// Hydrate API input for tunnels
	tunnelInputs, diags := hydrateAddIpsecIkeV2SiteTunnels(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.IPSec.IsNull() {
		tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteTunnels.request", map[string]interface{}{
			"request": utils.InterfaceToJSONString(tunnelInputs.update),
		})
		tunnelData, err_ipsec := r.client.catov2.SiteUpdateIpsecIkeV2SiteTunnels(ctx, siteID, tunnelInputs.update, r.client.AccountId)
		tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteTunnels.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(tunnelData),
		})
		if err_ipsec != nil {
			resp.Diagnostics.AddError(
				"Cato API error in SiteAddIpsecIkeV2SiteTunnels",
				err_ipsec.Error(),
			)
			return
		}

		// create types to support multiple primary and secondary tunnels
		if len(tunnelData.Site.GetUpdateIpsecIkeV2SiteTunnels().GetPrimaryUpdateIpsecIkeV2SiteTunnelsPayload().GetTunnels()) > 0 {
			tunnelsPrimaryData := tunnelData.Site.GetUpdateIpsecIkeV2SiteTunnels().GetPrimaryUpdateIpsecIkeV2SiteTunnelsPayload().GetTunnels()
			for x := 0; x < len(tunnelsPrimaryData); x++ {
				resp.State.SetAttribute(ctx, path.Root("ipsec").AtName("primary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id"), tunnelsPrimaryData[x].GetTunnelIDUpdateIpsecIkeV2SiteTunnelPayload().String())
			}
		}

		if len(tunnelData.Site.GetUpdateIpsecIkeV2SiteTunnels().GetSecondaryUpdateIpsecIkeV2SiteTunnelsPayload().GetTunnels()) > 0 {
			tunnelsSecondaryData := tunnelData.Site.GetUpdateIpsecIkeV2SiteTunnels().GetSecondaryUpdateIpsecIkeV2SiteTunnelsPayload().GetTunnels()
			for x := 0; x < len(tunnelsSecondaryData); x++ {
				resp.State.SetAttribute(ctx, path.Root("ipsec").AtName("primary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id"), tunnelsSecondaryData[x].GetTunnelIDUpdateIpsecIkeV2SiteTunnelPayload().String())
			}
		}
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *siteIpsecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state SiteIpsecIkeV2
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
	tflog.Debug(ctx, "Delete.EntityLookup.response", map[string]interface{}{
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
		tflog.Debug(ctx, "Delete site")
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
