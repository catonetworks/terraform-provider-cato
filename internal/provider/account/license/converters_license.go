package license

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	licensing "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/licensing"
)

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
	if curLicense.Sku == licensing.SKUCatoPB || curLicense.Sku == licensing.SKUCatoPBSSE {
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
	if curLicense.Sku == licensing.SKUCatoPB || curLicense.Sku == licensing.SKUCatoSite {
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
	case licensing.SKUCatoPB, licensing.SKUCatoPBSSE:
		for _, site := range license.PooledBandwidthLicense.Sites {
			if site.SitePooledBandwidthLicenseSite.ID == siteID {
				return &site.AllocatedBandwidth, true
			}
		}
	case licensing.SKUCatoSite, licensing.SKUCatoSSESite:
		if license.SiteLicense.Site != nil && license.SiteLicense.Site.ID == siteID {
			return nil, true
		}
	}
	return nil, false
}

func getLicenseAssignmentForState(
	ctx context.Context,
	state Resource,
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
		case licensing.SKUCatoPB, licensing.SKUCatoPBSSE:
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
		case licensing.SKUCatoSite, licensing.SKUCatoSSESite:
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
			if *curLicense.ID == licenseID && (curLicense.Sku == licensing.SKUCatoSite || curLicense.Sku == licensing.SKUCatoSSESite) {
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

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	plan *Resource,
	priorDiags *diag.Diagnostics,
	curLicense *cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses,
) (resultLicenseInfo basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	plan.ID = types.StringValue(plan.SiteID.ValueString())
	plan.SiteID = types.StringValue(plan.SiteID.ValueString())
	// Check for valid license and hydrate state
	resultLicenseInfo, diagstmp := hydrateLicenseState(ctx, plan.LicenseID.ValueString(), curLicense)
	diags = append(diags, diagstmp...)
	return resultLicenseInfo
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	ctx context.Context,
	priorDiags *diag.Diagnostics,
	licenseInfo basetypes.ObjectValue,
) (resultLicenseInfoObject basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultLicenseInfoObject, diags = types.ObjectValueFrom(ctx, LicenseInfoResourceAttrTypes, licenseInfo)
	return resultLicenseInfoObject
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	state *Resource,
	license *cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses,
	curSiteLicenseID types.String,
	allocatedBw *int64,
	siteIsAssigned bool,
	diagnostics *diag.Diagnostics,
) {
	if allocatedBw != nil {
		state.BW = types.Int64Value(*allocatedBw)
	}

	if siteIsAssigned {
		if curSiteLicenseID.IsNull() {
			diagnostics.AddError(
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
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan *Resource,
	priorDiags *diag.Diagnostics,
	curLicense *cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses,
) (resultLicenseInfo basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	plan.ID = types.StringValue(plan.SiteID.ValueString())
	plan.SiteID = types.StringValue(plan.SiteID.ValueString())
	// Check for valid license and hydrate state
	resultLicenseInfo, diagstmp := hydrateLicenseState(ctx, plan.LicenseID.ValueString(), curLicense)
	diags = append(diags, diagstmp...)
	return resultLicenseInfo
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	ctx context.Context,
	priorDiags *diag.Diagnostics,
	licenseInfo basetypes.ObjectValue,
) (resultLicenseInfoObject basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultLicenseInfoObject, diags = types.ObjectValueFrom(ctx, LicenseInfoResourceAttrTypes, licenseInfo)
	return resultLicenseInfoObject
}

// convertUpdateInput3 converts between the Terraform model and SDK values.
func convertUpdateInput3(plan *Resource, licenseInfoObject basetypes.ObjectValue) {
	plan.LicenseInfo = licenseInfoObject
	plan.SiteID = types.StringValue(plan.ID.ValueString())
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state Resource) (resultInput cato_models.RemoveSiteBwLicenseInput) {
	resultInput = cato_models.RemoveSiteBwLicenseInput{}
	resultInput.LicenseID = state.LicenseID.ValueString()
	siteRef := &cato_models.SiteRefInput{}
	siteRef.By = licenseSiteRefByID
	siteRef.Input = state.SiteID.ValueString()
	resultInput.Site = siteRef
	return resultInput
}
