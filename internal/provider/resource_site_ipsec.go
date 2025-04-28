package provider

import (
	"context"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
					// "city": schema.StringAttribute{
					// 	Description: "City",
					// 	Required:    false,
					// 	Optional:    true,
					// },
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

	// setting input
	input := cato_models.AddIpsecIkeV2SiteInput{}
	input_ipsec := &cato_models.AddIpsecIkeV2SiteTunnelsInput{}
	varSiteId := ""

	// setting input site location
	if !plan.SiteLocation.IsNull() {
		input.SiteLocation = &cato_models.AddSiteLocationInput{}
		siteLocationInput := AddIpsecSiteLocationInput{}
		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.SiteLocation.Address = siteLocationInput.Address.ValueStringPointer()
		// input.SiteLocation.City = siteLocationInput.City.ValueStringPointer()
		input.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueString()
		input.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()
		input.SiteLocation.Timezone = siteLocationInput.Timezone.ValueString()
	}

	// setting input other attributes
	input.Name = plan.Name.ValueString()
	input.SiteType = (cato_models.SiteType)(plan.SiteType.ValueString())
	input.NativeNetworkRange = plan.NativeNetworkRange.ValueString()
	input.Description = plan.Description.ValueStringPointer()

	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2Site", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input_ipsec),
	})

	ipsecSite, err := r.client.catov2.SiteAddIpsecIkeV2Site(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2Site", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ipsecSite),
	})

	// retrieving native-network range ID to update native range
	entityParent := cato_models.EntityInput{
		ID:   ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID(),
		Type: (cato_models.EntityType)("site"),
	}

	siteRangeEntities, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"), nil, nil, &entityParent, nil, nil, nil, nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Create.EntityLookup", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteRangeEntities),
	})

	var networkRangeEntity cato_go_sdk.EntityLookup_EntityLookup_Items_Entity
	for _, item := range siteRangeEntities.EntityLookup.Items {
		splitName := strings.Split(*item.Entity.Name, " \\ ")
		if splitName[2] == "Native Range" {
			networkRangeEntity = item.Entity
		}
	}

	if !plan.IPSec.IsNull() {
		planIPSec := AddIpsecIkeV2SiteTunnelsInput{}
		diags = plan.IPSec.As(ctx, &planIPSec, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// setting primary
		if !planIPSec.Primary.IsNull() {
			input_ipsec.Primary = &cato_models.AddIpsecIkeV2TunnelsInput{}
			primaryInput := &AddIpsecIkeV2TunnelsInput{}
			diags = planIPSec.Primary.As(ctx, &primaryInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			input_ipsec.Primary.DestinationType = (*cato_models.DestinationType)(primaryInput.DestinationType.ValueStringPointer())
			input_ipsec.Primary.PopLocationID = primaryInput.PopLocationID.ValueStringPointer()
			input_ipsec.Primary.PublicCatoIPID = primaryInput.PublicCatoIPID.ValueStringPointer()

			// setting tunnels
			if !primaryInput.Tunnels.IsNull() {
				elementsTunnels := make([]types.Object, 0, len(primaryInput.Tunnels.Elements()))
				diags = primaryInput.Tunnels.ElementsAs(ctx, &elementsTunnels, false)
				resp.Diagnostics.Append(diags...)

				var itemTunnels AddIpsecIkeV2TunnelInput
				for _, item := range elementsTunnels {
					diags = item.As(ctx, &itemTunnels, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// setting lastMileBw
					var itemTunnelLastMileBw LastMileBwInput
					diags = itemTunnels.LastMileBw.As(ctx, &itemTunnelLastMileBw, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// append tunnels
					input_ipsec.Primary.Tunnels = append(input_ipsec.Primary.Tunnels, &cato_models.AddIpsecIkeV2TunnelInput{
						LastMileBw: &cato_models.LastMileBwInput{
							Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
							Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
						},
						PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
						PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
						Psk:           itemTunnels.Psk.ValueString(),
						PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
					})
				}
			}

			// setting secondary
			if !planIPSec.Secondary.IsNull() {
				input_ipsec.Secondary = &cato_models.AddIpsecIkeV2TunnelsInput{}
				secondaryInput := &AddIpsecIkeV2TunnelsInput{}
				diags = planIPSec.Secondary.As(ctx, &secondaryInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				input_ipsec.Secondary.DestinationType = (*cato_models.DestinationType)(secondaryInput.DestinationType.ValueStringPointer())
				input_ipsec.Secondary.PopLocationID = secondaryInput.PopLocationID.ValueStringPointer()
				input_ipsec.Secondary.PublicCatoIPID = secondaryInput.PublicCatoIPID.ValueStringPointer()

				// setting tunnels
				if !secondaryInput.Tunnels.IsNull() {
					elementsTunnels := make([]types.Object, 0, len(secondaryInput.Tunnels.Elements()))
					diags = secondaryInput.Tunnels.ElementsAs(ctx, &elementsTunnels, false)
					resp.Diagnostics.Append(diags...)

					var itemTunnels AddIpsecIkeV2TunnelInput
					for _, item := range elementsTunnels {
						diags = item.As(ctx, &itemTunnels, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						// setting lastMileBw
						var itemTunnelLastMileBw LastMileBwInput
						diags = itemTunnels.LastMileBw.As(ctx, &itemTunnelLastMileBw, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						// append tunnels
						input_ipsec.Secondary.Tunnels = append(input_ipsec.Secondary.Tunnels, &cato_models.AddIpsecIkeV2TunnelInput{
							LastMileBw: &cato_models.LastMileBwInput{
								Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
								Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
							},
							PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
							PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
							Psk:           itemTunnels.Psk.ValueString(),
							PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
						})
					}
				}

			}
		}

		diags = plan.IPSec.As(ctx, &planIPSec, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
	}

	tflog.Debug(ctx, "site_ipsec_tunnel create", map[string]interface{}{
		"input_ipsec": utils.InterfaceToJSONString(input_ipsec),
	})

	if resp.Diagnostics.HasError() {
		return
	}

	varSiteId = ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID()

	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2SiteTunnels", map[string]interface{}{
		"input_ipsec": utils.InterfaceToJSONString(input_ipsec),
	})

	tunnelData, err_ipsec := r.client.catov2.SiteAddIpsecIkeV2SiteTunnels(ctx, varSiteId, *input_ipsec, r.client.AccountId)
	if err_ipsec != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteAddIpsecIkeV2SiteTunnels",
			err_ipsec.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2SiteTunnels", map[string]interface{}{
		"response": utils.InterfaceToJSONString(tunnelData),
	})

	// create types to support multiple primary and secondary tunnels
	tunnelsPrimaryData := []*cato_go_sdk.SiteAddIpsecIkeV2SiteTunnels_Site_AddIpsecIkeV2SiteTunnels_PrimaryAddIpsecIkeV2SiteTunnelsPayload_Tunnels{}
	tunnelsSecondaryData := []*cato_go_sdk.SiteAddIpsecIkeV2SiteTunnels_Site_AddIpsecIkeV2SiteTunnels_SecondaryAddIpsecIkeV2SiteTunnelsPayload_Tunnels{}

	if len(tunnelData.Site.GetAddIpsecIkeV2SiteTunnels().PrimaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()) > 0 {
		tunnelsPrimaryData = tunnelData.Site.GetAddIpsecIkeV2SiteTunnels().PrimaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()
	}

	if len(tunnelData.Site.GetAddIpsecIkeV2SiteTunnels().SecondaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()) > 0 {
		tunnelsSecondaryData = tunnelData.Site.GetAddIpsecIkeV2SiteTunnels().SecondaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// supports multiple primary ipsec tunnels
	if input_ipsec.Primary != nil {
		for x := 0; x < len(tunnelsPrimaryData); x++ {
			resp.State.SetAttribute(ctx, path.Root("ipsec").AtName("primary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id"), tunnelsPrimaryData[x].GetTunnelIDAddIpsecIkeV2SiteTunnelPayload().String())
		}
	}

	// supports multiple secondary ipsec tunnels
	if input_ipsec.Secondary != nil {
		for x := 0; x < len(tunnelsSecondaryData); x++ {
			resp.State.SetAttribute(ctx, path.Root("ipsec").AtName("secondary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id"), tunnelsSecondaryData[x].GetTunnelIDAddIpsecIkeV2SiteTunnelPayload().String())
		}
	}

	// overiding state with socket site id
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), types.StringValue(ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID()))
	// overiding state with native network range id
	resp.State.SetAttribute(ctx, path.Empty().AtName("native_network_range_id"), networkRangeEntity.ID)
	resp.State.SetAttribute(ctx, path.Root("ipsec").AtName("site_id"), varSiteId)
}

func (r *siteIpsecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state SiteIpsecIkeV2
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if site exist, else remove resource
	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Read.EntityLookup", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})

	// read in the ipsec site entries
	for _, v := range querySiteResult.EntityLookup.Items {
		if v.Entity.ID == state.ID.ValueString() {
			resp.State.SetAttribute(
				ctx,
				path.Root("id"),
				v.Entity.ID,
			)
		}
	}

	// check if site exist before refreshing
	if len(querySiteResult.EntityLookup.GetItems()) != 1 {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
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

	input_ipsec := cato_models.UpdateIpsecIkeV2SiteTunnelsInput{}

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

	tflog.Debug(ctx, "ipsec site update", map[string]interface{}{
		"input-ipsecsite":    utils.InterfaceToJSONString(inputSiteGeneral),
		"input-networkRange": utils.InterfaceToJSONString(inputUpdateNetworkRange),
	})

	tflog.Debug(ctx, "Update.SiteUpdateSiteGeneralDetails", map[string]interface{}{
		"input": utils.InterfaceToJSONString(inputSiteGeneral),
	})

	inputSiteGeneralResponse, err := r.client.catov2.SiteUpdateSiteGeneralDetails(ctx, plan.ID.ValueString(), inputSiteGeneral, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateSiteGeneralDetails error",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Update.SiteUpdateSiteGeneralDetails", map[string]interface{}{
		"response": utils.InterfaceToJSONString(inputSiteGeneralResponse),
	})

	tflog.Debug(ctx, "Update.SiteUpdateNetworkRange", map[string]interface{}{
		"input": utils.InterfaceToJSONString(inputUpdateNetworkRange),
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

	planIPSec := AddIpsecIkeV2SiteTunnelsInput{}
	diags = plan.IPSec.As(ctx, &planIPSec, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	varSiteId := planIPSec.SiteId.ValueString()

	if !plan.IPSec.IsNull() {
		if !planIPSec.Primary.IsNull() {
			input_ipsec.Primary = &cato_models.UpdateIpsecIkeV2TunnelsInput{}
			planIPSecPrimary := AddIpsecIkeV2TunnelsInput{}
			diags = planIPSec.Primary.As(ctx, &planIPSecPrimary, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			input_ipsec.Primary.DestinationType = (*cato_models.DestinationType)(planIPSecPrimary.DestinationType.ValueStringPointer())
			input_ipsec.Primary.PopLocationID = planIPSecPrimary.PopLocationID.ValueStringPointer()
			input_ipsec.Primary.PublicCatoIPID = planIPSecPrimary.PublicCatoIPID.ValueStringPointer()

			if !planIPSecPrimary.Tunnels.IsNull() {
				elementsTunnels := make([]types.Object, 0, len(planIPSecPrimary.Tunnels.Elements()))
				diags = planIPSecPrimary.Tunnels.ElementsAs(ctx, &elementsTunnels, false)
				resp.Diagnostics.Append(diags...)

				var itemTunnels AddIpsecIkeV2TunnelInput
				for _, item := range elementsTunnels {
					diags = item.As(ctx, &itemTunnels, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// setting lastMileBw
					var itemTunnelLastMileBw LastMileBwInput
					diags = itemTunnels.LastMileBw.As(ctx, &itemTunnelLastMileBw, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// append tunnels
					input_ipsec.Primary.Tunnels = append(input_ipsec.Primary.Tunnels, &cato_models.UpdateIpsecIkeV2TunnelInput{
						LastMileBw: &cato_models.LastMileBwInput{
							Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
							Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
						},
						PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
						PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
						Psk:           itemTunnels.Psk.ValueStringPointer(),
						PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
						TunnelID:      cato_models.IPSecV2InterfaceID(itemTunnels.TunnelID.ValueString()),
					})
				}
			}
		}

		if !planIPSec.Secondary.IsNull() {
			input_ipsec.Secondary = &cato_models.UpdateIpsecIkeV2TunnelsInput{}
			planIPSecSecondary := AddIpsecIkeV2TunnelsInput{}
			diags = planIPSec.Secondary.As(ctx, &planIPSecSecondary, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			input_ipsec.Secondary.DestinationType = (*cato_models.DestinationType)(planIPSecSecondary.DestinationType.ValueStringPointer())
			input_ipsec.Secondary.PopLocationID = planIPSecSecondary.PopLocationID.ValueStringPointer()
			input_ipsec.Secondary.PublicCatoIPID = planIPSecSecondary.PublicCatoIPID.ValueStringPointer()

			if !planIPSecSecondary.Tunnels.IsNull() {
				elementsTunnels := make([]types.Object, 0, len(planIPSecSecondary.Tunnels.Elements()))
				diags = planIPSecSecondary.Tunnels.ElementsAs(ctx, &elementsTunnels, false)
				resp.Diagnostics.Append(diags...)

				var itemTunnels AddIpsecIkeV2TunnelInput
				for _, item := range elementsTunnels {
					diags = item.As(ctx, &itemTunnels, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// setting lastMileBw
					var itemTunnelLastMileBw LastMileBwInput
					diags = itemTunnels.LastMileBw.As(ctx, &itemTunnelLastMileBw, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// append tunnels
					input_ipsec.Secondary.Tunnels = append(input_ipsec.Secondary.Tunnels, &cato_models.UpdateIpsecIkeV2TunnelInput{
						LastMileBw: &cato_models.LastMileBwInput{
							Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
							Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
						},
						PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
						PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
						Psk:           itemTunnels.Psk.ValueStringPointer(),
						PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
						TunnelID:      cato_models.IPSecV2InterfaceID(itemTunnels.TunnelID.ValueString()),
					})
				}
			}
		}

		tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteTunnels", map[string]interface{}{
			"input": utils.InterfaceToJSONString(input_ipsec),
		})

		tunnelData, err_ipsec := r.client.catov2.SiteUpdateIpsecIkeV2SiteTunnels(ctx, varSiteId, input_ipsec, r.client.AccountId)
		if err_ipsec != nil {
			resp.Diagnostics.AddError(
				"Cato API error in SiteAddIpsecIkeV2SiteTunnels",
				err_ipsec.Error(),
			)
			return
		}

		tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteTunnels", map[string]interface{}{
			"response": utils.InterfaceToJSONString(tunnelData),
		})

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

	diags = resp.State.Set(ctx, plan)
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
