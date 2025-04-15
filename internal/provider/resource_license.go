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
					stringplanmodifier.RequiresReplace(),
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
			"license_id_current": schema.StringAttribute{
				Description: "License ID Current",
				Required:    false,
				Optional:    true,
				Computed:    true,
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

	// Get all sites, check for valid siteID
	siteExists := false
	siteResponse, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityTypeSite, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	for _, item := range siteResponse.GetEntityLookup().GetItems() {
		entityFields := item.GetEntity()
		tflog.Warn(ctx, "Checking site IDs, input='"+fmt.Sprintf("%v", plan.SiteID.ValueString())+"', currentItem='"+entityFields.GetID()+"'")
		if entityFields.GetID() == plan.SiteID.ValueString() {
			tflog.Warn(ctx, "Site ID matched! "+entityFields.GetID())
			siteExists = true
			break
		}
	}
	if !siteExists {
		resp.Diagnostics.AddError("INVALID SITE ID", "Site '"+plan.SiteID.ValueString()+"' not found.")
		return
	}

	// Get all licenses
	licensingInfoResponse, err := r.client.catov2.Licensing(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	license, licenseExists := getLicenseByID(ctx, plan.LicenseID.ValueString(), licensingInfoResponse)

	if licenseExists {
		input := cato_models.AssignSiteBwLicenseInput{}
		input.LicenseID = plan.LicenseID.ValueString()
		plan.ID = types.StringValue(plan.SiteID.ValueString())
		plan.LicenseIDCurrent = types.StringValue(plan.LicenseID.ValueString())
		siteRef := &cato_models.SiteRefInput{}
		siteRef.By = "ID"
		siteRef.Input = plan.SiteID.ValueString()
		input.Site = siteRef
		// Check for the correct license type
		tflog.Warn(ctx, "Checking license SKU for CATO_PB or CATO_SITE type, license.Sku='"+string(license.Sku)+"'")
		if string(license.Sku) != "CATO_PB" && string(license.Sku) != "CATO_SITE" {
			resp.Diagnostics.AddError(
				"INVALID LICENSE TYPE",
				"Site License ID '"+plan.LicenseID.ValueString()+"' is not a valid site license. Must be 'CATO_PB' or 'CATO_SITE' license type.",
			)
			return
		}
		// Check for BW if pooled
		if string(license.Sku) == "CATO_PB" {
			if plan.BW.IsUnknown() || plan.BW.IsNull() {
				resp.Diagnostics.AddError("INVALID CONFIGURATION", "Bandwidth must be set for 'CATO_PB' pooled bandwidth license type.")
				return
			}
			bw := plan.BW.ValueInt64()
			input.Bw = &bw
		} else {
			// Check for BW not present if not pooled
			if !plan.BW.IsUnknown() && !plan.BW.IsNull() {
				resp.Diagnostics.AddError("INVALID CONFIGURATION", "Bandwidth is not supported for CATO_SITE license type only for 'CATO_PB' pooled bandwidth type.")
				return
			}
		}

		// TODO Check if the site has a license currently
		// If site assigned, use replace, otherwise use assign

		_, err := r.client.catov2.AssignSiteBwLicense(ctx, r.client.AccountId, input)
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API error", err.Error())
			return
		}
		// Check for valid license and hydrate state
		licenseInfo, diagstmp := hydrateLicenseState(ctx, plan.LicenseID.ValueString(), license)
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
	} else {
		resp.Diagnostics.AddError("INVALID LICENSE ID", "License '"+plan.LicenseID.ValueString()+"' not found.")
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
	licenses := licensingInfoResponse.GetLicensing().GetLicensingInfo().GetLicenses()
	for _, curLicense := range licenses {
		if curLicense.ID != nil && *curLicense.ID == state.ID.ValueString() {
			license = curLicense
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
	state.LicenseIDCurrent = state.LicenseID
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
	tflog.Warn(ctx, "Current1 plan.ID (old license) '"+fmt.Sprintf("%v", plan.LicenseIDCurrent.ValueString())+"' and current plan.LicenseID (new license)='"+fmt.Sprintf("%v", plan.LicenseID.ValueString())+"'")

	// Get all sites, check for valid siteID
	siteExists := false
	siteResponse, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityTypeSite, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	for _, item := range siteResponse.GetEntityLookup().GetItems() {
		entityFields := item.GetEntity()
		tflog.Warn(ctx, "Checking site IDs, input='"+fmt.Sprintf("%v", plan.SiteID.ValueString())+"', currentItem='"+entityFields.GetID()+"'")
		if entityFields.GetID() == plan.SiteID.ValueString() {
			tflog.Warn(ctx, "Site ID matched! "+entityFields.GetID())
			siteExists = true
			break
		}
	}
	if !siteExists {
		resp.Diagnostics.AddError("INVALID SITE ID", "Site '"+plan.SiteID.ValueString()+"' not found.")
		return
	}

	// Get all licenses, check for valid licenseID
	licensingInfoResponse, err := r.client.catov2.Licensing(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API error", err.Error())
		return
	}
	license, licenseExists := getLicenseByID(ctx, plan.LicenseID.ValueString(), licensingInfoResponse)

	if licenseExists {
		tflog.Warn(ctx, "Current plan.ID (old license) '"+fmt.Sprintf("%v", plan.LicenseIDCurrent.ValueString())+"' and current plan.LicenseID (new license)='"+fmt.Sprintf("%v", plan.LicenseID.ValueString())+"'")
		// Check to see if licenseID changed, use replace else use udpate
		if plan.LicenseIDCurrent.ValueString() != plan.LicenseID.ValueString() {
			tflog.Warn(ctx, "License does not match, using ReplaceSiteBwLicenseInput()")
			input := cato_models.ReplaceSiteBwLicenseInput{}
			input.LicenseIDToAdd = plan.LicenseID.ValueString()
			input.LicenseIDToRemove = plan.LicenseIDCurrent.ValueString()
			siteRef := &cato_models.SiteRefInput{}
			siteRef.By = "ID"
			siteRef.Input = plan.SiteID.ValueString()
			input.Site = siteRef
			// Check for BW if pooled
			if license.Sku == "CATO_PB" {
				if plan.BW.IsUnknown() || plan.BW.IsNull() {
					resp.Diagnostics.AddError("INVALID CONFIGURATION", "Bandwidth must be set for 'CATO_PB' pooled bandwidth license type.")
					return
				}
				input.Bw = plan.BW.ValueInt64Pointer()
			} else {
				// Check for BW not present if not pooled
				if !plan.BW.IsUnknown() && !plan.BW.IsNull() {
					resp.Diagnostics.AddError("INVALID CONFIGURATION", "Bandwidth is not supported for CATO_SITE license type only for 'CATO_PB' pooled bandwidth type.")
					return
				}
			}
			_, err := r.client.catov2.ReplaceSiteBwLicense(ctx, r.client.AccountId, input)
			if err != nil {
				resp.Diagnostics.AddError("Catov2 API ReplaceSiteBwLicense error", err.Error())
				return
			}
		} else {
			tflog.Warn(ctx, "License does match, using UpdateSiteBwLicenseInput()")
			input := cato_models.UpdateSiteBwLicenseInput{}
			input.LicenseID = plan.LicenseID.ValueString()
			siteRef := &cato_models.SiteRefInput{}
			siteRef.By = "ID"
			siteRef.Input = plan.SiteID.ValueString()
			input.Site = siteRef
			// Check for BW if pooled
			if license.Sku == "CATO_PB" {
				if plan.BW.IsUnknown() || plan.BW.IsNull() {
					resp.Diagnostics.AddError("INVALID CONFIGURATION", "Bandwidth must be set for 'CATO_PB' pooled bandwidth license type.")
					return
				}
				bw := plan.BW.ValueInt64()
				input.Bw = bw
			} else {
				// Check for BW not present if not pooled
				if !plan.BW.IsUnknown() && !plan.BW.IsNull() {
					resp.Diagnostics.AddError("INVALID CONFIGURATION", "Bandwidth is not supported for CATO_SITE license type only for 'CATO_PB' pooled bandwidth type.")
					return
				}
			}
			_, err := r.client.catov2.UpdateSiteBwLicense(ctx, r.client.AccountId, input)
			if err != nil {
				resp.Diagnostics.AddError("Catov2 API ReplaceSiteBwLicense error", err.Error())
				return
			}
		}
		plan.ID = types.StringValue(plan.SiteID.ValueString())
		plan.LicenseIDCurrent = types.StringValue(plan.LicenseID.ValueString())
		// Check for valid license and hydrate state
		licenseInfo, diagstmp := hydrateLicenseState(ctx, plan.LicenseID.ValueString(), license)
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
	} else {
		resp.Diagnostics.AddError("INVALID LICENSE ID", "License '"+plan.LicenseID.ValueString()+"' not found.")
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
	// licenseInfo := LicenseInfoResource{}
	// if curLicense.Sku == "CATO_PB" || curLicense.Sku == "CATO_SITE" {
	// 	if curLicense.Status == "ACTIVE" {
	// 		licenseInfo.AllocatedBandwidth = types.Int64Value(0)
	// 		licenseInfo.ExpirationDate = types.StringValue(string(curLicense.ExpirationDate))
	// 		licenseInfo.LastUpdated = types.StringValue(string(*curLicense.LastUpdated))
	// 		licenseInfo.Plan = types.StringValue(string(curLicense.Plan))
	// 		licenseInfo.SiteLicenseGroup = types.StringValue(string("*curLicense.SiteLicenseGroup"))
	// 		licenseInfo.SiteLicenseType = types.StringValue(string("*curLicense.SiteLicenseType"))
	// 		licenseInfo.SKU = types.StringValue(string(curLicense.Sku))
	// 		licenseInfo.StartDate = types.StringValue(string(*curLicense.StartDate))
	// 		licenseInfo.Status = types.StringValue(string(curLicense.Status))
	// 		licenseInfo.Total = types.Int64Value(int64(0))
	// 	} else {
	// 		diags.AddError(
	// 			"INACTIVE LICENSE STATUS",
	// 			"Site License ID '"+licenseId+"' is not active. Must be 'ACTIVE' status.",
	// 		)
	// 		return licenseInfo, diags
	// 	}
	// } else {
	// 	diags.AddError(
	// 		"INVALID LICENSE ID",
	// 		"Site License ID '"+licenseId+"' is not a valid site license. Must be 'CATO_PB' or 'CATO_SITE' license type.",
	// 	)
	// 	return licenseInfo, diags
	// }
	tflog.Warn(ctx, "hydrate(1)")
	// licenseInfo, diagstmp := types.ObjectValue(
	// 	LicenseInfoResourceAttrTypes,
	// 	map[string]attr.Value{
	// 		"allocated_bandwidth": types.Int64Value(0),
	// 		"expiration_date":     types.StringValue(string(curLicense.ExpirationDate)),
	// 		"last_updated":        types.StringValue(string(*curLicense.LastUpdated)),
	// 		"plan":                types.StringValue(string(curLicense.Plan)),
	// 		"site_license_group":  types.StringValue(string("*curLicense.SiteLicenseGroup")),
	// 		"site_license_type":   types.StringValue(string("*curLicense.SiteLicenseType")),
	// 		"sku":                 types.StringValue(string(curLicense.Sku)),
	// 		"start_date":          types.StringValue(string(*curLicense.StartDate)),
	// 		"status":              types.StringValue(string(curLicense.Status)),
	// 		"total":               types.Int64Value(int64(0)),
	// 	},
	// )

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
	tflog.Warn(ctx, "hydrate(2)")
	diags = append(diags, diagstmp...)
	if curLicense.Sku == "CATO_PB" || curLicense.Sku == "CATO_SITE" {
		tflog.Warn(ctx, "hydrate(3)")
		if curLicense.Status != "ACTIVE" {
			tflog.Warn(ctx, "hydrate(4)")
			diags.AddError(
				"INACTIVE LICENSE STATUS",
				"Site License ID '"+licenseId+"' is not active. Must be 'ACTIVE' status.",
			)
		}
	} else {
		tflog.Warn(ctx, "hydrate(5)")
		diags.AddError(
			"INVALID LICENSE ID",
			"Site License ID '"+licenseId+"' is not a valid site license. Must be 'CATO_PB' or 'CATO_SITE' license type.",
		)
	}
	tflog.Warn(ctx, "hydrate(6)")
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
		tflog.Warn(ctx, "Checking license IDs, input='"+curLicenseId+"', currentItem='"+licenseID+"'")
		if licenseID == curLicenseId {
			tflog.Warn(ctx, "Found license ID! "+licenseID)
			licenseExists = true
			license = curLicense
			break
		}
	}
	return license, licenseExists
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
