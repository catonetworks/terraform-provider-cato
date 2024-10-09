package provider

// import (
// 	"context"

// 	"github.com/cato-networks/terraform-provider-cato/internal/utils"
// 	"github.com/hashicorp/terraform-plugin-framework/resource"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
// 	"github.com/hashicorp/terraform-plugin-framework/types"
// 	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
// 	"github.com/hashicorp/terraform-plugin-log/tflog"
// 	cato_models "github.com/cato-networks/cato-go-sdk/models"
// )

// var (
// 	_ resource.Resource              = &siteIpsecTunnelResource{}
// 	_ resource.ResourceWithConfigure = &siteIpsecTunnelResource{}
// )

// func NewSiteIpsecTunnelResource() resource.Resource {
// 	return &siteIpsecTunnelResource{}
// }

// type siteIpsecTunnelResource struct {
// 	client *catoClientData
// }

// type AddIpsecIkeV2SiteTunnelsInput struct {
// 	SiteId    types.String `tfsdk:"site_id"`
// 	Primary   types.Object `tfsdk:"primary"`   //AddIpsecIkeV2TunnelsInput
// 	Secondary types.Object `tfsdk:"secondary"` //AddIpsecIkeV2TunnelsInput
// }

// type AddIpsecIkeV2TunnelsInput struct {
// 	DestinationType types.String `tfsdk:"destination_type"`
// 	PublicCatoIPID  types.String `tfsdk:"public_cato_ip_id"`
// 	PopLocationID   types.String `tfsdk:"pop_location_id"`
// 	Tunnels         types.List   `tfsdk:"tunnels"` //[]*AddIpsecIkeV2TunnelInput
// }

// type AddIpsecIkeV2TunnelInput struct {
// 	PublicSiteIP  types.String `tfsdk:"public_site_ip"`
// 	PrivateCatoIP types.String `tfsdk:"private_cato_ip"`
// 	PrivateSiteIP types.String `tfsdk:"private_site_ip"`
// 	LastMileBw    types.Object `tfsdk:"last_mile_bw"` //*LastMileBwInput
// 	Psk           types.String `tfsdk:"psk"`
// }

// type LastMileBwInput struct {
// 	Downstream              types.Int64   `tfsdk:"downstream"`
// 	Upstream                types.Int64   `tfsdk:"upstream"`
// 	DownstreamMbpsPrecision types.Float64 `tfsdk:"downstream_mbps_precision"`
// 	UpstreamMbpsPrecision   types.Float64 `tfsdk:"upstream_mbps_precision"`
// }

// func (r *siteIpsecTunnelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
// 	resp.TypeName = req.ProviderTypeName + "_ipsec_tunnel"
// }

// func (r *siteIpsecTunnelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
// 	resp.Schema = schema.Schema{
// 		Attributes: map[string]schema.Attribute{
// 			"site_id": schema.StringAttribute{
// 				Description: "site_id",
// 				Computed:    true,
// 				PlanModifiers: []planmodifier.String{
// 					stringplanmodifier.UseStateForUnknown(),
// 				},
// 			},
// 			"primary": schema.SingleNestedAttribute{
// 				Description: "primary",
// 				Required:    true,
// 				Attributes: map[string]schema.Attribute{
// 					"destination_type": schema.StringAttribute{
// 						Description: "destinationtype",
// 						Required:    false,
// 						Optional:    true,
// 					},
// 					"public_cato_ip_id": schema.StringAttribute{
// 						Description: "publiccatoipid",
// 						Required:    true,
// 					},
// 					"pop_location_id": schema.StringAttribute{
// 						Description: "poplocationid",
// 						Required:    true,
// 					},
// 					"tunnels": schema.ListNestedAttribute{
// 						Description: "tunnels",
// 						Required:    false,
// 						Optional:    true,
// 						NestedObject: schema.NestedAttributeObject{
// 							Attributes: map[string]schema.Attribute{
// 								"public_site_ip": schema.StringAttribute{
// 									Description: "publicsiteip",
// 									Required:    true,
// 								},
// 								"private_cato_ip": schema.StringAttribute{
// 									Description: "privatecatoip",
// 									Optional:    true,
// 								},
// 								"private_site_ip": schema.StringAttribute{
// 									Description: "privatesiteip",
// 									Optional:    true,
// 								},
// 								"psk": schema.StringAttribute{
// 									Description: "psk",
// 									Required:    true,
// 								},
// 								"last_mile_bw": schema.SingleNestedAttribute{
// 									Description: "lastmilebw",
// 									Required:    false,
// 									Optional:    true,
// 									Attributes: map[string]schema.Attribute{
// 										"downstream": schema.Int64Attribute{
// 											Description: "Downstream",
// 											Required:    true,
// 										},
// 										"upstream": schema.Int64Attribute{
// 											Description: "upstream",
// 											Required:    true,
// 										},
// 										"downstream_mbps_precision": schema.Float64Attribute{
// 											Description: "downstreamMbpsPrecision",
// 											Required:    false,
// 											Optional:    true,
// 										},
// 										"upstream_mbps_precision": schema.Float64Attribute{
// 											Description: "upstreamMbpsPrecision",
// 											Required:    false,
// 											Optional:    true,
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			"secondary": schema.SingleNestedAttribute{
// 				Description: "secondary",
// 				Optional:    true,
// 				Attributes: map[string]schema.Attribute{
// 					"destination_type": schema.StringAttribute{
// 						Description: "destinationtype",
// 						Required:    true,
// 					},
// 					"public_cato_ip_id": schema.StringAttribute{
// 						Description: "publiccatoipid",
// 						Required:    true,
// 					},
// 					"pop_location_id": schema.StringAttribute{
// 						Description: "poplocationid",
// 						Required:    true,
// 					},
// 					"tunnels": schema.SingleNestedAttribute{
// 						Description: "tunnels",
// 						Required:    false,
// 						Optional:    true,
// 						Attributes: map[string]schema.Attribute{
// 							"public_site_ip": schema.StringAttribute{
// 								Description: "publicsiteip",
// 								Required:    true,
// 							},
// 							"private_cato_ip": schema.StringAttribute{
// 								Description: "privatecatoip",
// 								Optional:    true,
// 							},
// 							"private_site_ip": schema.StringAttribute{
// 								Description: "privatesiteip",
// 								Optional:    true,
// 							},
// 							"psk": schema.StringAttribute{
// 								Description: "psk",
// 								Required:    true,
// 							},
// 							"last_mile_bw": schema.SingleNestedAttribute{
// 								Description: "lastmilebw",
// 								Required:    false,
// 								Optional:    true,
// 								Attributes: map[string]schema.Attribute{
// 									"downstream": schema.Int64Attribute{
// 										Description: "Downstream",
// 										Required:    true,
// 									},
// 									"upstream": schema.Int64Attribute{
// 										Description: "upstream",
// 										Required:    true,
// 									},
// 									"downstream_mbps_precision": schema.Float64Attribute{
// 										Description: "downstreamMbpsPrecision",
// 										Required:    false,
// 										Optional:    true,
// 									},
// 									"upstream_mbps_precision": schema.Float64Attribute{
// 										Description: "upstreamMbpsPrecision",
// 										Required:    false,
// 										Optional:    true,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// }

// func (r *siteIpsecTunnelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
// 	if req.ProviderData == nil {
// 		return
// 	}

// 	r.client = req.ProviderData.(*catoClientData)
// }

// func (r *siteIpsecTunnelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

// 	var plan AddIpsecIkeV2SiteTunnelsInput
// 	diags := req.Plan.Get(ctx, &plan)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	// setting input
// 	input := &cato_models.AddIpsecIkeV2SiteTunnelsInput{}

// 	// setting primary
// 	if !plan.Primary.IsNull() {
// 		input.Primary = &cato_models.AddIpsecIkeV2TunnelsInput{}
// 		primaryInput := &AddIpsecIkeV2TunnelsInput{}
// 		diags = plan.Primary.As(ctx, &primaryInput, basetypes.ObjectAsOptions{})
// 		resp.Diagnostics.Append(diags...)

// 		input.Primary.DestinationType = (*cato_models.DestinationType)(primaryInput.DestinationType.ValueStringPointer())
// 		input.Primary.PopLocationID = primaryInput.PopLocationID.ValueStringPointer()
// 		input.Primary.PublicCatoIPID = primaryInput.PublicCatoIPID.ValueStringPointer()

// 		// setting tunnels
// 		if !primaryInput.Tunnels.IsNull() {
// 			elementsTunnels := make([]types.Object, 0, len(primaryInput.Tunnels.Elements()))
// 			diags = primaryInput.Tunnels.ElementsAs(ctx, &elementsTunnels, false)
// 			resp.Diagnostics.Append(diags...)

// 			var itemTunnels AddIpsecIkeV2TunnelInput
// 			for _, item := range elementsTunnels {
// 				diags = item.As(ctx, &itemTunnels, basetypes.ObjectAsOptions{})
// 				resp.Diagnostics.Append(diags...)

// 				// setting lastMileBw
// 				var itemTunnelLastMileBw LastMileBwInput
// 				diags = itemTunnels.LastMileBw.As(ctx, &itemTunnelLastMileBw, basetypes.ObjectAsOptions{})
// 				resp.Diagnostics.Append(diags...)

// 				// append tunnels
// 				input.Primary.Tunnels = append(input.Primary.Tunnels, &cato_models.AddIpsecIkeV2TunnelInput{
// 					LastMileBw: &cato_models.LastMileBwInput{
// 						Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
// 						Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
// 					},
// 					PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
// 					PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
// 					Psk:           itemTunnels.Psk.ValueString(),
// 					PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
// 				})
// 			}
// 		}

// 		// setting secondary
// 		if !plan.Primary.IsNull() {
// 			input.Secondary = &cato_models.AddIpsecIkeV2TunnelsInput{}
// 			primaryInput := &AddIpsecIkeV2TunnelsInput{}
// 			diags = plan.Secondary.As(ctx, &primaryInput, basetypes.ObjectAsOptions{})
// 			resp.Diagnostics.Append(diags...)

// 			input.Secondary.DestinationType = (*cato_models.DestinationType)(primaryInput.DestinationType.ValueStringPointer())
// 			input.Secondary.PopLocationID = primaryInput.PopLocationID.ValueStringPointer()
// 			input.Secondary.PublicCatoIPID = primaryInput.PublicCatoIPID.ValueStringPointer()

// 			// setting tunnels
// 			if !primaryInput.Tunnels.IsNull() {
// 				elementsTunnels := make([]types.Object, 0, len(primaryInput.Tunnels.Elements()))
// 				diags = primaryInput.Tunnels.ElementsAs(ctx, &elementsTunnels, false)
// 				resp.Diagnostics.Append(diags...)

// 				var itemTunnels AddIpsecIkeV2TunnelInput
// 				for _, item := range elementsTunnels {
// 					diags = item.As(ctx, &itemTunnels, basetypes.ObjectAsOptions{})
// 					resp.Diagnostics.Append(diags...)

// 					// setting lastMileBw
// 					var itemTunnelLastMileBw LastMileBwInput
// 					diags = itemTunnels.LastMileBw.As(ctx, &itemTunnelLastMileBw, basetypes.ObjectAsOptions{})
// 					resp.Diagnostics.Append(diags...)

// 					// append tunnels
// 					input.Secondary.Tunnels = append(input.Secondary.Tunnels, &cato_models.AddIpsecIkeV2TunnelInput{
// 						LastMileBw: &cato_models.LastMileBwInput{
// 							Downstream: itemTunnelLastMileBw.Downstream.ValueInt64Pointer(),
// 							Upstream:   itemTunnelLastMileBw.Upstream.ValueInt64Pointer(),
// 						},
// 						PrivateCatoIP: itemTunnels.PrivateCatoIP.ValueStringPointer(),
// 						PrivateSiteIP: itemTunnels.PrivateSiteIP.ValueStringPointer(),
// 						Psk:           itemTunnels.Psk.ValueString(),
// 						PublicSiteIP:  itemTunnels.PublicSiteIP.ValueStringPointer(),
// 					})
// 				}
// 			}

// 		}
// 	}

// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	tflog.Debug(ctx, "site_ipsec_tunnel create", map[string]interface{}{
// 		"input": utils.InterfaceToJSONString(input),
// 	})

// 	_, err := r.client.catov2.SiteAddIpsecIkeV2SiteTunnels(ctx, plan.SiteId.ValueString(), *input, r.client.AccountId)
// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Cato API error in SiteAddIpsecIkeV2SiteTunnels",
// 			err.Error(),
// 		)
// 		return
// 	}

// 	diags = resp.State.Set(ctx, plan)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// }

// func (r *siteIpsecTunnelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
// 	// var state SiteIpsecIkeV2
// 	// diags := req.State.Get(ctx, &state)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }

// 	// entityIds := []string{state.ID.ValueString()}

// 	// queryResult, err := r.info.catov2.EntityLookup(ctx, r.info.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, entityIds, nil, nil, nil)
// 	// if err != nil {
// 	// 	tflog.Error(ctx, "Read: did not receive a response in read for EntityLookup")
// 	// 	resp.Diagnostics.AddError(
// 	// 		"Catov2 API EntityLookup error",
// 	// 		err.Error(),
// 	// 	)
// 	// 	return
// 	// }

// 	// if len(queryResult.EntityLookup.GetItems()) == 1 {
// 	// 	entryList := queryResult.EntityLookup.GetItems()
// 	// 	entry := entryList[0]
// 	// 	helperFields := entry.GetHelperFields()
// 	// 	state.ID = types.StringValue(entry.Entity.ID)
// 	// 	state.Name = types.StringValue(*entry.Entity.Name)
// 	// 	state.SiteType = types.StringValue(helperFields["type"].(string))
// 	// 	state.Description = types.StringValue(helperFields["description"].(string))

// 	// } else {
// 	// 	tflog.Error(ctx, "Read: did not receive a response in read for EntityLookup")
// 	// 	resp.Diagnostics.AddError(
// 	// 		"Catov2 API EntityLookup error",
// 	// 		"more than one value returned in EntityLookup for SiteIpsecIkeV2",
// 	// 	)
// 	// 	return
// 	// }

// 	// diags = resp.State.Set(ctx, &state)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }
// }

// func (r *siteIpsecTunnelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

// 	// var plan SiteIpsecIkeV2
// 	// diags := req.Plan.Get(ctx, &plan)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }

// 	// currentId := plan.ID.ValueString()
// 	// siteType := cato_models.SiteType(plan.SiteType.ValueString())

// 	// input := cato_models.UpdateSiteGeneralDetailsInput{
// 	// 	Name:        plan.Name.ValueStringPointer(),
// 	// 	SiteType:    &siteType,
// 	// 	Description: plan.Description.ValueStringPointer(),
// 	// 	SiteLocation: &cato_models.UpdateSiteLocationInput{
// 	// 		CountryCode: plan.SiteLocation.CountryCode.ValueStringPointer(),
// 	// 		StateCode:   plan.SiteLocation.StateCode.ValueStringPointer(),
// 	// 		Timezone:    plan.SiteLocation.Timezone.ValueStringPointer(),
// 	// 		Address:     plan.SiteLocation.Address.ValueStringPointer(),
// 	// 	},
// 	// }

// 	// _, err := r.info.catov2.SiteUpdateSiteGeneralDetails(ctx, currentId, input, r.info.AccountId)
// 	// if err != nil {
// 	// 	resp.Diagnostics.AddError(
// 	// 		"Cato API error",
// 	// 		err.Error(),
// 	// 	)
// 	// 	return
// 	// }

// 	// diags = resp.State.Set(ctx, plan)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }

// 	// diags = resp.State.Set(ctx, plan)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }
// }

// func (r *siteIpsecTunnelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

// 	// var state SiteIpsecIkeV2
// 	// diags := req.State.Get(ctx, &state)
// 	// resp.Diagnostics.Append(diags...)
// 	// if resp.Diagnostics.HasError() {
// 	// 	return
// 	// }

// 	// logIdEntry := "DELETING SITEIPSEC ID: " + state.ID.ValueString()
// 	// tflog.Warn(ctx, logIdEntry)
// 	// _, err := r.info.catov2.SiteRemoveSite(ctx, state.ID.ValueString(), r.info.AccountId)
// 	// if err != nil {
// 	// 	resp.Diagnostics.AddError(
// 	// 		"Cato API v2 error in SiteRemoveSite",
// 	// 		err.Error(),
// 	// 	)
// 	// 	return
// 	// }

// }
