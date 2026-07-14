package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource                = &licenseResource{}
	_ resource.ResourceWithConfigure   = &licenseResource{}
	_ resource.ResourceWithImportState = &licenseResource{}
)

const (
	licenseSiteRefByID                = "ID"
	licenseStatusActive               = "ACTIVE"
	multipleOfTenValidatorDescription = "Ensures the value is a multiple of 10"
)

func NewLicenseResource() resource.Resource {
	return &licenseResource{}
}

type licenseResource struct {
	client *catoClientData
}

func (r *licenseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

//nolint:funlen // Terraform schemas are declarative and lengthy by nature.
func (r *licenseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_license` resource contains the configuration parameters necessary to assign and " +
			"replace licenses for sites. When creating a new license resource, the `site_id` and `license_id` " +
			"attributes are required. The `license_info` attribute is optional and will be populated with the " +
			"license information after the resource is created. If the site has an existing license assigned, " +
			"the license resource will call the mutation.ReplaceSiteBwLicense() operation replacing the existing " +
			"license with the new license_id ensuring there is no interruption in servive of reassigning a license." +
			"\n\n**NOTE** License assignment does not work for Trial accounts, or for accounts that are \"not synced\".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "License ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "Site ID",
				Required:    true,
				Optional:    false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"license_id": schema.StringAttribute{
				Description: "License ID",
				Required:    false,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bw": schema.Int64Attribute{
				Description: "Bandwidth to allocate to site (only used for pooled license model)",
				Required:    false,
				Optional:    true,
				Validators: []validator.Int64{
					customInt64Validator{},
				},
			},
			"license_info": schema.SingleNestedAttribute{
				Description: "",
				Required:    false,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(), // avoid drift on refresh
				},
				Attributes: map[string]schema.Attribute{
					"sku": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"plan": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"status": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"expiration_date": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"start_date": schema.StringAttribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"last_updated": schema.StringAttribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"total": schema.Int64Attribute{
						Computed: true,
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"site_license_group": schema.StringAttribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"site_license_type": schema.StringAttribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"allocated_bandwidth": schema.Int64Attribute{
						Computed: true,
						Optional: true,
						Required: false,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}

func (r *licenseResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *licenseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *licenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LicenseResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	curLicense, err := upsertLicense(ctx, plan, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error updating license", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.SiteID.ValueString())
	plan.SiteID = types.StringValue(plan.SiteID.ValueString())
	// Check for valid license and hydrate state
	licenseInfo, diagstmp := hydrateLicenseState(ctx, plan.LicenseID.ValueString(), curLicense)
	diags = append(diags, diagstmp...)
	if diags.HasError() {
		return
	}
	licenseInfoObject, diags := types.ObjectValueFrom(ctx, LicenseInfoResourceAttrTypes, licenseInfo)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.LicenseInfo = licenseInfoObject
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
}

func (r *licenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LicenseResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get all licenses
	licensingInfoResponse, err := r.client.catov2.Licensing(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.Licensing.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(licensingInfoResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}

	// A site can use multiple pooled licenses. For configured resources, read
	// only the allocation represented by this resource's license_id.
	license, curSiteLicenseID, allocatedBw, siteIsAssigned, remoteAssignmentExists := getLicenseAssignmentForState(
		ctx,
		state,
		licensingInfoResponse,
	)
	if !remoteAssignmentExists {
		resp.State.RemoveResource(ctx)
		return
	}
	if allocatedBw != nil {
		state.BW = types.Int64Value(*allocatedBw)
	}

	if siteIsAssigned {
		if curSiteLicenseID.IsNull() {
			resp.Diagnostics.AddError(
				"License ID not found",
				"This could be due to the license or the account being set as trial where the license does not have an ID, "+
					"or if there is a China license on the account. If a trial license is assigned, "+
					"please unassign the trial license and try to reapply.",
			)
			return
		}
		// Check for valid license and hydrate state
		licenseInfo, diagstmp := hydrateLicenseState(ctx, state.ID.ValueString(), license)
		licenseInfoObject, diags := types.ObjectValueFrom(ctx, LicenseInfoResourceAttrTypes, licenseInfo)
		diags = append(diags, diagstmp...)
		if diags.HasError() {
			return
		}
		state.LicenseInfo = licenseInfoObject
	}

	state.SiteID = types.StringValue(state.ID.ValueString())
	state.LicenseID = curSiteLicenseID
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
}

func (r *licenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LicenseResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	curLicense, err := upsertLicense(ctx, plan, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error updating license", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.SiteID.ValueString())
	plan.SiteID = types.StringValue(plan.SiteID.ValueString())
	// Check for valid license and hydrate state
	licenseInfo, diagstmp := hydrateLicenseState(ctx, plan.LicenseID.ValueString(), curLicense)
	diags = append(diags, diagstmp...)
	if diags.HasError() {
		return
	}
	licenseInfoObject, diags := types.ObjectValueFrom(ctx, LicenseInfoResourceAttrTypes, licenseInfo)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.LicenseInfo = licenseInfoObject
	plan.SiteID = types.StringValue(plan.ID.ValueString())
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
}

func (r *licenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LicenseResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Disabled interface to "remove" an interface
	input := cato_models.RemoveSiteBwLicenseInput{}
	input.LicenseID = state.LicenseID.ValueString()
	siteRef := &cato_models.SiteRefInput{}
	siteRef.By = licenseSiteRefByID
	siteRef.Input = state.SiteID.ValueString()
	input.Site = siteRef

	tflog.Debug(ctx, "Delete.RemoveSiteBwLicense.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	RemoveSiteBwLicenseResponse, err := r.client.catov2.RemoveSiteBwLicense(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "Delete.RemoveSiteBwLicense.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(RemoveSiteBwLicenseResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API RemoveSiteBwLicense error",
			err.Error(),
		)
		return
	}
}

func hydrateLicenseState(
	_ context.Context,
	licenseID string,
	curLicense *cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses,
) (basetypes.ObjectValue, diag.Diagnostics) {
	diags := make(diag.Diagnostics, 0)
	licenseInfoAttrs := map[string]attr.Value{
		"allocated_bandwidth": types.Int64Null(),
		"expiration_date":     types.StringValue(curLicense.ExpirationDate),
		"last_updated":        types.StringValue(*curLicense.LastUpdated),
		"plan":                types.StringValue(string(curLicense.Plan)),
		"site_license_group":  types.StringNull(),
		"site_license_type":   types.StringNull(),
		"sku":                 types.StringValue(string(curLicense.Sku)),
		"start_date":          types.StringValue(*curLicense.StartDate),
		"status":              types.StringValue(string(curLicense.Status)),
		"total":               types.Int64Value(curLicense.SiteLicense.Total),
	}
	if curLicense.Sku == licenseSkuCatoPB || curLicense.Sku == licenseSkuCatoPBSSE {
		licenseInfoAttrs = map[string]attr.Value{
			"allocated_bandwidth": types.Int64Value(curLicense.PooledBandwidthLicense.AllocatedBandwidth),
			"expiration_date":     types.StringValue(curLicense.ExpirationDate),
			"last_updated":        types.StringValue(*curLicense.LastUpdated),
			"plan":                types.StringValue(string(curLicense.Plan)),
			"site_license_group":  types.StringNull(),
			"site_license_type":   types.StringNull(),
			"sku":                 types.StringValue(string(curLicense.Sku)),
			"start_date":          types.StringValue(*curLicense.StartDate),
			"status":              types.StringValue(string(curLicense.Status)),
			"total":               types.Int64Value(curLicense.SiteLicense.Total),
		}
	}

	licenseInfo, diagstmp := types.ObjectValue(LicenseInfoResourceAttrTypes, licenseInfoAttrs)
	diags = append(diags, diagstmp...)
	if curLicense.Sku == licenseSkuCatoPB || curLicense.Sku == licenseSkuCatoSite {
		if curLicense.Status != licenseStatusActive {
			diags.AddError(
				"INACTIVE LICENSE STATUS",
				"Site License ID '"+licenseID+"' is not active. Must be 'ACTIVE' status.",
			)
		}
	} else {
		diags.AddError(
			"INVALID LICENSE ID",
			"Site License ID '"+licenseID+"' is not a valid site license. Must be 'CATO_PB' or 'CATO_SITE' license type.",
		)
	}
	return licenseInfo, diags
}

func getLicenseByID(
	ctx context.Context,
	curLicenseID string,
	licensingInfoResponse *cato_go_sdk.Licensing,
) (*cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses, bool) {
	// Match current license by ID from API response
	licenseExists := false
	license := &cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses{}
	licenses := licensingInfoResponse.Licensing.LicensingInfo.Licenses
	for _, curLicense := range licenses {
		licenseID := ""
		if curLicense.ID != nil {
			licenseID = *curLicense.ID
		}
		if licenseID != "" && licenseID == curLicenseID {
			tflog.Warn(ctx, "Found license ID! "+licenseID)
			licenseExists = true
			license = curLicense
			break
		}
	}
	return license, licenseExists
}

func getSiteAssignmentForLicense(
	siteID string,
	license *cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses,
) (allocatedBw *int64, isAssigned bool) {
	switch license.Sku {
	case licenseSkuCatoPB, licenseSkuCatoPBSSE:
		for _, site := range license.PooledBandwidthLicense.Sites {
			if site.SitePooledBandwidthLicenseSite.ID == siteID {
				return &site.AllocatedBandwidth, true
			}
		}
	case licenseSkuCatoSite, licenseSkuCatoSSESite:
		if license.SiteLicense.Site != nil && license.SiteLicense.Site.ID == siteID {
			return nil, true
		}
	}
	return nil, false
}

func getLicenseAssignmentForState(
	ctx context.Context,
	state LicenseResource,
	licensingInfoResponse *cato_go_sdk.Licensing,
) (
	license *cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses,
	licenseID types.String,
	allocatedBw *int64,
	isAssigned bool,
	exists bool,
) {
	configuredLicenseID := state.LicenseID
	if !configuredLicenseID.IsNull() && !configuredLicenseID.IsUnknown() && configuredLicenseID.ValueString() != "" {
		license, licenseExists := getLicenseByID(ctx, configuredLicenseID.ValueString(), licensingInfoResponse)
		if !licenseExists {
			return license, configuredLicenseID, nil, false, false
		}
		allocatedBw, isAssigned := getSiteAssignmentForLicense(state.SiteID.ValueString(), license)
		return license, configuredLicenseID, allocatedBw, isAssigned, isAssigned
	}

	licenseID, allocatedBw, isAssigned = getCurrentAssignedLicenseBySiteID(ctx, state.SiteID.ValueString(), licensingInfoResponse)
	license = &cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses{}
	if isAssigned && !licenseID.IsNull() {
		license, _ = getLicenseByID(ctx, licenseID.ValueString(), licensingInfoResponse)
	}
	return license, licenseID, allocatedBw, isAssigned, true
}

func getCurrentAssignedLicenseBySiteID(
	ctx context.Context,
	curSiteID string,
	licensingInfoResponse *cato_go_sdk.Licensing,
) (curLicenseID types.String, allocatedBw *int64, isAssigned bool) {
	licenses := licensingInfoResponse.Licensing.LicensingInfo.Licenses
	for _, curLicense := range licenses {
		switch curLicense.Sku {
		case licenseSkuCatoPB, licenseSkuCatoPBSSE:
			if len(curLicense.PooledBandwidthLicense.Sites) > 0 {
				for _, site := range curLicense.PooledBandwidthLicense.Sites {
					if site.SitePooledBandwidthLicenseSite.ID != "" {
						tflog.Warn(ctx,
							"getCurrentAssignedLicenseBySiteID() - Checking site IDs, input='"+curSiteID+
								"', currentItem='"+site.SitePooledBandwidthLicenseSite.ID+"'",
						)
						if site.SitePooledBandwidthLicenseSite.ID == curSiteID {
							tflog.Warn(ctx, "getCurrentAssignedLicenseBySiteID() - Site ID matched! "+site.SitePooledBandwidthLicenseSite.ID)
							isAssigned = true
							allocatedBw = &site.AllocatedBandwidth
							if curLicense.ID != nil {
								curLicenseID = types.StringValue(*curLicense.ID)
							} else {
								curLicenseID = types.StringNull()
							}
						}
					}
				}
			}
		case licenseSkuCatoSite, licenseSkuCatoSSESite:
			if curLicense.SiteLicense.Site != nil {
				if curLicense.SiteLicense.Site.ID == curSiteID {
					tflog.Warn(ctx, "getCurrentAssignedLicenseBySiteID() - Site ID matched! "+curLicense.SiteLicense.Site.ID)
					isAssigned = true
					if curLicense.ID != nil {
						curLicenseID = types.StringValue(*curLicense.ID)
					} else {
						curLicenseID = types.StringNull()
					}
				}
			}
		}
	}
	return curLicenseID, allocatedBw, isAssigned
}

func checkStaticLicenseForAssignment(
	ctx context.Context,
	licenseID string,
	licensingInfoResponse *cato_go_sdk.Licensing,
) (types.String, bool) {
	isAssigned := false
	var curSiteID types.String
	licenses := licensingInfoResponse.Licensing.LicensingInfo.Licenses
	for _, curLicense := range licenses {
		if curLicense.ID != nil {
			tflog.Debug(ctx, "Calling checkStaticLicenseForAssignment()", map[string]interface{}{
				"curLicense.ID": fmt.Sprintf("%v", curLicense.ID),
			})
			if *curLicense.ID == licenseID && (curLicense.Sku == licenseSkuCatoSite || curLicense.Sku == licenseSkuCatoSSESite) {
				if curLicense.SiteLicense.Site != nil {
					tflog.Debug(ctx, "Calling checkStaticLicenseForAssignment() curLicense.SiteLicense.Site", map[string]interface{}{
						"curLicense.SiteLicense.Site.ID":                  curLicense.SiteLicense.Site.ID,
						"curLicense.ID == curLicense.SiteLicense.Site.ID": (*curLicense.ID == licenseID),
					})
					isAssigned = true
					curSiteID = types.StringValue(curLicense.SiteLicense.Site.ID)
				}
			}
		}
	}
	return curSiteID, isAssigned
}

// General purpose functions
type customInt64Validator struct{}

func (v customInt64Validator) Description(_ context.Context) string {
	return multipleOfTenValidatorDescription
}
func (v customInt64Validator) MarkdownDescription(_ context.Context) string {
	return multipleOfTenValidatorDescription
}
func (v customInt64Validator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueInt64()
	if val%10 != 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid bandwidth, must be in increments of 10",
			fmt.Sprintf("The value %d is not a multiple of 10", val),
		)
	}
}
