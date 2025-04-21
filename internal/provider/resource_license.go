package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
)

var (
	_ resource.Resource                = &licenseResource{}
	_ resource.ResourceWithConfigure   = &licenseResource{}
	_ resource.ResourceWithImportState = &licenseResource{}
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

func (r *licenseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_lan_interface` resource contains the configuration parameters necessary to add a lan interface to a socket. ([physical socket physical socket](https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). Documentation for the underlying API used in this resource can be found at [mutation.updateSocketInterface()](https://api.catonetworks.com/documentation/#mutation-site.updateSocketInterface).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "License ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					// stringplanmodifier.RequiresReplace(),
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
					objectplanmodifier.UseStateForUnknown(), // Avoid drift
				},
				Attributes: map[string]schema.Attribute{
					"sku": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
					},
					"plan": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
					},
					"status": schema.StringAttribute{
						Computed: true,
						Required: false,
						Optional: true,
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
					},
					"site_license_type": schema.StringAttribute{
						Computed: true,
						Optional: true,
						Required: false,
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

func (r *licenseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}

	// Match current license by ID from API response
	license := &cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses{}
	curSiteLicenseId, allocatedBw, siteIsAssigned := getCurrentAssignedLicenseBySiteId(ctx, state.ID.ValueString(), licensingInfoResponse)
	if allocatedBw != nil {
		state.BW = types.Int64Value(*allocatedBw)
	}

	if siteIsAssigned {
		licenses := licensingInfoResponse.GetLicensing().GetLicensingInfo().GetLicenses()
		for _, curLicense := range licenses {
			if curLicense.ID != nil && *curLicense.ID == curSiteLicenseId.ValueString() {
				license = curLicense
			}
		}
	}

	// Check for valid license and hydrate state
	licenseInfo, diagstmp := hydrateLicenseState(ctx, state.ID.ValueString(), license)
	diags = append(diags, diagstmp...)
	if diags.HasError() {
		return
	}
	licenseInfoObject, diags := types.ObjectValueFrom(ctx, LicenseInfoResourceAttrTypes, licenseInfo)
	state.LicenseInfo = licenseInfoObject
	state.SiteID = types.StringValue(state.ID.ValueString())
	state.LicenseID = curSiteLicenseId
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
	siteRef.By = "ID"
	siteRef.Input = state.SiteID.ValueString()
	input.Site = siteRef

	tflog.Debug(ctx, "license remove", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	_, err := r.client.catov2.RemoveSiteBwLicense(ctx, r.client.AccountId, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API RemoveSiteBwLicense error",
			err.Error(),
		)
		return
	}
}

func hydrateLicenseState(ctx context.Context, licenseId string, curLicense *cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses) (basetypes.ObjectValue, diag.Diagnostics) {
	diags := make(diag.Diagnostics, 0)
	licenseInfoAttrs := map[string]attr.Value{
		"allocated_bandwidth": types.Int64Value(0),
		"expiration_date":     types.StringValue(string(curLicense.ExpirationDate)),
		"last_updated":        types.StringNull(),
		"plan":                types.StringValue(string(curLicense.Plan)),
		"site_license_group":  types.StringNull(),
		"site_license_type":   types.StringNull(),
		"sku":                 types.StringValue(string(curLicense.Sku)),
		"start_date":          types.StringNull(),
		"status":              types.StringValue(string(curLicense.Status)),
		"total":               types.Int64Value(0),
	}

	if curLicense.LastUpdated != nil {
		licenseInfoAttrs["last_updated"] = types.StringValue(string(*curLicense.LastUpdated))
	}
	if curLicense.StartDate != nil {
		licenseInfoAttrs["start_date"] = types.StringValue(string(*curLicense.StartDate))
	}

	licenseInfo, diagstmp := types.ObjectValue(LicenseInfoResourceAttrTypes, licenseInfoAttrs)
	diags = append(diags, diagstmp...)
	if curLicense.Sku == "CATO_PB" || curLicense.Sku == "CATO_SITE" {
		if curLicense.Status != "ACTIVE" {
			diags.AddError(
				"INACTIVE LICENSE STATUS",
				"Site License ID '"+licenseId+"' is not active. Must be 'ACTIVE' status.",
			)
		}
	} else {
		diags.AddError(
			"INVALID LICENSE ID",
			"Site License ID '"+licenseId+"' is not a valid site license. Must be 'CATO_PB' or 'CATO_SITE' license type.",
		)
	}
	return licenseInfo, diags
}

func getLicenseByID(ctx context.Context, curLicenseId string, licensingInfoResponse *cato_go_sdk.Licensing) (*cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses, bool) {
	// Match current license by ID from API response
	licenseExists := false
	license := &cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses{}
	licenses := licensingInfoResponse.Licensing.LicensingInfo.Licenses
	for _, curLicense := range licenses {
		licenseID := ""
		if curLicense.ID != nil {
			licenseID = *curLicense.ID
		}
		if licenseID != "" && licenseID == curLicenseId {
			tflog.Warn(ctx, "Found license ID! "+licenseID)
			licenseExists = true
			license = curLicense
			break
		}
	}
	return license, licenseExists
}

func getCurrentAssignedLicenseBySiteId(ctx context.Context, curSiteId string, licensingInfoResponse *cato_go_sdk.Licensing) (types.String, *int64, bool) {
	isAssigned := false
	var allocatedBw *int64 = nil
	var curLicenseId types.String
	licenses := licensingInfoResponse.Licensing.LicensingInfo.Licenses
	for _, curLicense := range licenses {
		if curLicense.Sku == "CATO_PB" || curLicense.Sku == "CATO_PB_SSE" {
			if len(curLicense.PooledBandwidthLicense.Sites) > 0 {
				for _, site := range curLicense.PooledBandwidthLicense.Sites {
					tflog.Warn(ctx, "getCurrentAssignedLicenseBySiteId() - Checking site IDs, input='"+fmt.Sprintf("%v", curSiteId)+"', currentItem='"+site.SitePooledBandwidthLicenseSite.ID+"'")
					if site.SitePooledBandwidthLicenseSite.ID == curSiteId {
						tflog.Warn(ctx, "getCurrentAssignedLicenseBySiteId() - Site ID matched! "+site.SitePooledBandwidthLicenseSite.ID)
						isAssigned = true
						allocatedBw = &site.AllocatedBandwidth
						curLicenseId = types.StringValue(*curLicense.ID)
					}
				}
			}
		} else if curLicense.Sku == "CATO_SITE" || curLicense.Sku == "CATO_SSE_SITE" {
			if curLicense.SiteLicense.Site != nil {
				if curLicense.SiteLicense.Site.ID == curSiteId {
					tflog.Warn(ctx, "getCurrentAssignedLicenseBySiteId() - Site ID matched! "+curLicense.SiteLicense.Site.ID)
					isAssigned = true
					curLicenseId = types.StringValue(*curLicense.ID)
				}
			}
		}
	}
	return curLicenseId, allocatedBw, isAssigned
}

func checkStaticLicenseForAssignment(ctx context.Context, licenseId string, licensingInfoResponse *cato_go_sdk.Licensing) (types.String, bool) {
	isAssigned := false
	var curSiteId types.String
	licenses := licensingInfoResponse.Licensing.LicensingInfo.Licenses
	for _, curLicense := range licenses {
		if curLicense.ID != nil {
			tflog.Debug(ctx, "Calling checkStaticLicenseForAssignment()", map[string]interface{}{
				"curLicense.ID": fmt.Sprintf("%v", curLicense.ID),
			})
			if *curLicense.ID == licenseId && (curLicense.Sku == "CATO_SITE" || curLicense.Sku == "CATO_SSE_SITE") {
				if curLicense.SiteLicense.Site != nil {
					tflog.Debug(ctx, "Calling checkStaticLicenseForAssignment() curLicense.SiteLicense.Site", map[string]interface{}{
						"curLicense.SiteLicense.Site.ID":                  fmt.Sprintf("%v", curLicense.SiteLicense.Site.ID),
						"curLicense.ID == curLicense.SiteLicense.Site.ID": (*curLicense.ID == licenseId),
					})
					isAssigned = true
					curSiteId = types.StringValue(curLicense.SiteLicense.Site.ID)
				}
			}
		}
	}
	return curSiteId, isAssigned
}

// General purpose functions
type customInt64Validator struct{}

func (v customInt64Validator) Description(ctx context.Context) string {
	return "Ensures the value is a multiple of 10"
}
func (v customInt64Validator) MarkdownDescription(ctx context.Context) string {
	return "Ensures the value is a multiple of 10"
}
func (v customInt64Validator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
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
