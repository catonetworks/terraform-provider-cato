package provider

// import (
// 	"context"
// 	"strings"

// 	"github.com/catonetworks/terraform-provider-cato/internal/utils"
// 	"github.com/hashicorp/terraform-plugin-framework/path"
// 	"github.com/hashicorp/terraform-plugin-framework/resource"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
// 	"github.com/hashicorp/terraform-plugin-framework/types"
// 	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
// 	"github.com/hashicorp/terraform-plugin-log/tflog"
// 	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
// 	cato_models "github.com/catonetworks/cato-go-sdk/models"
// )

// var (
// 	_ resource.Resource              = &siteIpsecResource{}
// 	_ resource.ResourceWithConfigure = &siteIpsecResource{}
// )

// func NewSiteIpsecResource() resource.Resource {
// 	return &siteIpsecResource{}
// }

// type siteIpsecResource struct {
// 	client *catoClientData
// }

// type SiteIpsecIkeV2 struct {
// 	ID                   types.String `tfsdk:"id"`
// 	Name                 types.String `tfsdk:"name"`
// 	SiteType             types.String `tfsdk:"site_type"`
// 	Description          types.String `tfsdk:"description"`
// 	NativeNetworkRange   types.String `tfsdk:"native_network_range"`
// 	NativeNetworkRangeId types.String `tfsdk:"native_network_range_id"`
// 	SiteLocation         types.Object `tfsdk:"site_location"`
// }

// type AddIpsecSiteLocationInput struct {
// 	CountryCode types.String `tfsdk:"country_code"`
// 	StateCode   types.String `tfsdk:"state_code"`
// 	Timezone    types.String `tfsdk:"timezone"`
// 	Address     types.String `tfsdk:"address"`
// 	// City        types.String `tfsdk:"city"`
// }

// func (r *siteIpsecResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
// 	resp.TypeName = req.ProviderTypeName + "_ipsec_site"
// }

// func (r *siteIpsecResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
// 	resp.Schema = schema.Schema{
// 		Attributes: map[string]schema.Attribute{
// 			"id": schema.StringAttribute{
// 				Description: "Identifier for Ipsec Site",
// 				Computed:    true,
// 				PlanModifiers: []planmodifier.String{
// 					stringplanmodifier.UseStateForUnknown(),
// 				},
// 			},
// 			"name": schema.StringAttribute{
// 				Description: "Ipsec Site Name",
// 				Required:    true,
// 			},
// 			"site_type": schema.StringAttribute{
// 				Description: "Valid values are: BRANCH, HEADQUARTERS, CLOUD_DC, and DATACENTER.",
// 				Required:    true,
// 			},
// 			"description": schema.StringAttribute{
// 				Description: "description",
// 				Required:    true,
// 			},
// 			"native_network_range": schema.StringAttribute{
// 				Description: "NativeNetworkRange",
// 				Required:    true,
// 			},
// 			"native_network_range_id": schema.StringAttribute{
// 				Description: "Site native IP range ID (for update purpose)",
// 				Optional:    true,
// 				Computed:    true,
// 				PlanModifiers: []planmodifier.String{
// 					stringplanmodifier.UseStateForUnknown(),
// 				},
// 			},
// 			"site_location": schema.SingleNestedAttribute{
// 				Description: "SiteLocation",
// 				Required:    true,
// 				Attributes: map[string]schema.Attribute{
// 					"country_code": schema.StringAttribute{
// 						Description: "Country Code",
// 						Required:    true,
// 					},
// 					"state_code": schema.StringAttribute{
// 						Description: "State Code",
// 						Required:    false,
// 						Optional:    true,
// 					},
// 					"timezone": schema.StringAttribute{
// 						Description: "Timezone",
// 						Required:    true,
// 					},
// 					"address": schema.StringAttribute{
// 						Description: "Address",
// 						Required:    false,
// 						Optional:    true,
// 					},
// 					// "city": schema.StringAttribute{
// 					// 	Description: "City",
// 					// 	Required:    false,
// 					// 	Optional:    true,
// 					// },
// 				},
// 			},
// 		},
// 	}
// }

// func (r *siteIpsecResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
// 	if req.ProviderData == nil {
// 		return
// 	}

// 	r.client = req.ProviderData.(*catoClientData)
// }

// func (r *siteIpsecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

// 	var plan SiteIpsecIkeV2
// 	diags := req.Plan.Get(ctx, &plan)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	// setting input
// 	input := cato_models.AddIpsecIkeV2SiteInput{}

// 	// setting input site location
// 	if !plan.SiteLocation.IsNull() {
// 		input.SiteLocation = &cato_models.AddSiteLocationInput{}
// 		siteLocationInput := AddIpsecSiteLocationInput{}
// 		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
// 		resp.Diagnostics.Append(diags...)

// 		input.SiteLocation.Address = siteLocationInput.Address.ValueStringPointer()
// 		// input.SiteLocation.City = siteLocationInput.City.ValueStringPointer()
// 		input.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueString()
// 		input.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()
// 		input.SiteLocation.Timezone = siteLocationInput.Timezone.ValueString()
// 	}

// 	// setting input other attributes
// 	input.Name = plan.Name.ValueString()
// 	input.SiteType = (cato_models.SiteType)(plan.SiteType.ValueString())
// 	input.NativeNetworkRange = plan.NativeNetworkRange.ValueString()
// 	input.Description = plan.Description.ValueStringPointer()

// 	tflog.Debug(ctx, "ipsec site create", map[string]interface{}{
// 		"input-ipsecsite": utils.InterfaceToJSONString(input),
// 	})

// 	ipsecSite, err := r.client.catov2.SiteAddIpsecIkeV2Site(ctx, input, r.client.AccountId)
// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Cato API error",
// 			err.Error(),
// 		)
// 		return
// 	}

// 	// retrieving native-network range ID to update native range
// 	entityParent := cato_models.EntityInput{
// 		ID:   ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID(),
// 		Type: (cato_models.EntityType)("site"),
// 	}

// 	siteRangeEntities, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("siteRange"), nil, nil, &entityParent, nil, nil, nil, nil, nil)
// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Catov2 API EntityLookup error",
// 			err.Error(),
// 		)
// 		return
// 	}

// 	var networkRangeEntity cato_go_sdk.EntityLookup_EntityLookup_Items_Entity
// 	for _, item := range siteRangeEntities.EntityLookup.Items {
// 		splitName := strings.Split(*item.Entity.Name, " \\ ")
// 		if splitName[2] == "Native Range" {
// 			networkRangeEntity = item.Entity
// 		}
// 	}

// 	diags = resp.State.Set(ctx, plan)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	// overiding state with socket site id
// 	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), types.StringValue(ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID()))
// 	// overiding state with native network range id
// 	resp.State.SetAttribute(ctx, path.Empty().AtName("native_network_range_id"), networkRangeEntity.ID)
// }

// func (r *siteIpsecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

// 	var state SiteIpsecIkeV2
// 	diags := req.State.Get(ctx, &state)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	// check if site exist, else remove resource
// 	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Catov2 API error",
// 			err.Error(),
// 		)
// 		return
// 	}

// 	// check if site exist before refreshing
// 	if len(querySiteResult.EntityLookup.GetItems()) != 1 {
// 		tflog.Warn(ctx, "site not found, site resource removed")
// 		resp.State.RemoveResource(ctx)
// 		return
// 	}

// 	diags = resp.State.Set(ctx, &state)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// }

// func (r *siteIpsecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

// 	var plan SiteIpsecIkeV2
// 	diags := req.Plan.Get(ctx, &plan)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	// setting input & input to update network range
// 	inputSiteGeneral := cato_models.UpdateSiteGeneralDetailsInput{
// 		SiteLocation: &cato_models.UpdateSiteLocationInput{},
// 	}

// 	inputUpdateNetworkRange := cato_models.UpdateNetworkRangeInput{}

// 	// setting input site location
// 	if !plan.SiteLocation.IsNull() {
// 		inputSiteGeneral.SiteLocation = &cato_models.UpdateSiteLocationInput{}
// 		siteLocationInput := SiteLocation{}
// 		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
// 		resp.Diagnostics.Append(diags...)

// 		inputSiteGeneral.SiteLocation.Address = siteLocationInput.Address.ValueStringPointer()
// 		inputSiteGeneral.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueStringPointer()
// 		inputSiteGeneral.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()
// 		inputSiteGeneral.SiteLocation.Timezone = siteLocationInput.Timezone.ValueStringPointer()
// 		// inputSiteGeneral.SiteLocation.City = siteLocationInput.City.ValueStringPointer()
// 	}

// 	inputUpdateNetworkRange.Subnet = plan.NativeNetworkRange.ValueStringPointer()
// 	inputUpdateNetworkRange.TranslatedSubnet = plan.NativeNetworkRange.ValueStringPointer()
// 	inputSiteGeneral.Name = plan.Name.ValueStringPointer()
// 	inputSiteGeneral.SiteType = (*cato_models.SiteType)(plan.SiteType.ValueStringPointer())
// 	inputSiteGeneral.Description = plan.Description.ValueStringPointer()

// 	tflog.Debug(ctx, "ipsec site update", map[string]interface{}{
// 		"input-ipsecsite":    utils.InterfaceToJSONString(inputSiteGeneral),
// 		"input-networkRange": utils.InterfaceToJSONString(inputUpdateNetworkRange),
// 	})

// 	_, err := r.client.catov2.SiteUpdateSiteGeneralDetails(ctx, plan.ID.ValueString(), inputSiteGeneral, r.client.AccountId)
// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Catov2 API SiteUpdateSiteGeneralDetails error",
// 			err.Error(),
// 		)
// 		return
// 	}

// 	_, err = r.client.catov2.SiteUpdateNetworkRange(ctx, plan.NativeNetworkRangeId.ValueString(), inputUpdateNetworkRange, r.client.AccountId)
// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Catov2 API SiteUpdateNetworkRange error",
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

// func (r *siteIpsecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

// 	var state SiteIpsecIkeV2
// 	diags := req.State.Get(ctx, &state)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"), nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Catov2 API error",
// 			err.Error(),
// 		)
// 		return
// 	}

// 	// check if site exist before removing
// 	if len(querySiteResult.EntityLookup.GetItems()) == 1 {

// 		_, err := r.client.catov2.SiteRemoveSite(ctx, state.ID.ValueString(), r.client.AccountId)
// 		if err != nil {
// 			resp.Diagnostics.AddError(
// 				"Catov2 API SiteRemoveSite error",
// 				err.Error(),
// 			)
// 			return
// 		}
// 	}

// }
