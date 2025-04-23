package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func upsertLicense(ctx context.Context, plan LicenseResource, cc *catoClientData) (*cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses, error) {
	diags := make(diag.Diagnostics, 0)
	// Get all sites, check for valid siteID
	siteExists := false
	siteResponse, err := cc.catov2.EntityLookup(ctx, cc.AccountId, cato_models.EntityTypeSite, nil, nil, nil, nil, nil, nil, nil, nil)

	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic("Catov2 API error", err.Error()))
		return nil, err
	}
	for _, item := range siteResponse.GetEntityLookup().GetItems() {
		tflog.Warn(ctx, "Checking site IDs, input='"+fmt.Sprintf("%v", plan.SiteID.ValueString())+"', currentItem='"+item.GetEntity().GetID()+"'")
		if item.GetEntity().GetID() == plan.SiteID.ValueString() {
			tflog.Warn(ctx, "Site ID matched! "+item.GetEntity().GetID())
			siteExists = true
			break
		}
	}
	if !siteExists {
		message := "Site '" + plan.SiteID.ValueString() + "' not found."
		diags = append(diags, diag.NewErrorDiagnostic("INVALID SITE ID", message))
		return nil, fmt.Errorf("INVALID SITE ID: %s", message)
	}

	// Get all licenses
	licensingInfoResponse, err := cc.catov2.Licensing(ctx, cc.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic("Catov2 API error", err.Error()))
		return nil, err
	}

	// Check if static license currently assigned to site
	curLicenseSiteId, licenseStaticIsAssigned := checkStaticLicenseForAssignment(ctx, plan.LicenseID.ValueString(), licensingInfoResponse)
	if licenseStaticIsAssigned {
		message := "The license ID '" + fmt.Sprintf("%v", plan.LicenseID.ValueString()) + "' is already assigned to site ID " + fmt.Sprintf("%v", curLicenseSiteId)
		diags = append(diags, diag.NewErrorDiagnostic("LICENSE ALREADY ASSIGNED", message))
		return nil, fmt.Errorf("LICENSE ALREADY ASSIGNED: %s", message)
	}

	// Check if the site has a license currently
	curSiteLicenseId, allocatedBw, siteIsAssigned := getCurrentAssignedLicenseBySiteId(ctx, plan.SiteID.ValueString(), licensingInfoResponse)
	if allocatedBw != nil {
		plan.BW = types.Int64Value(*allocatedBw)
	}

	// Get current license objeect by ID
	license, licenseExists := getLicenseByID(ctx, plan.LicenseID.ValueString(), licensingInfoResponse)

	if licenseExists {
		// Check for the correct license type
		tflog.Warn(ctx, "Checking license SKU for CATO_PB, CATO_PB_SSE, CATO_SITE, or CATO_SSE_SITE type, license.Sku='"+string(license.Sku)+"'")
		if string(license.Sku) != "CATO_PB" && string(license.Sku) != "CATO_PB_SSE" && string(license.Sku) != "CATO_SITE" && string(license.Sku) != "CATO_SSE_SITE" {
			message := "Site License ID '" + plan.LicenseID.ValueString() + "' is not a valid site license. Must be 'CATO_PB', 'CATO_PB_SSE', 'CATO_SITE', or 'CATO_SSE_SITE' license type."
			diags = append(diags, diag.NewErrorDiagnostic("INVALID LICENSE TYPE", message))
			return nil, fmt.Errorf("INVALID LICENSE TYPE: %s", message)
		}

		// check if assigned, use replace else use assign
		if siteIsAssigned {
			// if the site is already assigned to this license
			if curSiteLicenseId.ValueString() == plan.LicenseID.ValueString() {
				tflog.Warn(ctx, "Site ID '"+fmt.Sprintf("%v", plan.SiteID.ValueString())+"' is already assigned to license ID '"+fmt.Sprintf("%v", license.ID)+"'")
				if string(license.Sku) == "CATO_PB" || string(license.Sku) == "CATO_PB_SSE" {
					tflog.Warn(ctx, "Checking current license BW '"+fmt.Sprintf("%v", license.PooledBandwidthLicense.GetAllocatedBandwidth())+"' to configured plan BW '"+fmt.Sprintf("%v", plan.BW.ValueInt64())+"'")
					if license.PooledBandwidthLicense.GetAllocatedBandwidth() != plan.BW.ValueInt64() {
						// Checking license SKU type for PB
						tflog.Warn(ctx, "Pooled bandwidth license identfied '"+fmt.Sprintf("%v", license.Sku)+"', assigning BW")
						if plan.BW.IsUnknown() || plan.BW.IsNull() {
							message := "Bandwidth must be set for 'CATO_PB' and 'CATO_PB_SSE' pooled bandwidth license type."
							diags = append(diags, diag.NewErrorDiagnostic("INVALID CONFIGURATION", message))
							return nil, fmt.Errorf("INVALID CONFIGURATION: %s", message)
						}
						input := cato_models.UpdateSiteBwLicenseInput{}
						input.LicenseID = plan.LicenseID.ValueString()
						siteRef := &cato_models.SiteRefInput{}
						siteRef.By = "ID"
						siteRef.Input = plan.SiteID.ValueString()
						input.Site = siteRef
						input.Bw = plan.BW.ValueInt64()
						tflog.Debug(ctx, "Calling UpdateSiteBwLicenseInput()", map[string]interface{}{
							"input": utils.InterfaceToJSONString(input),
						})
						_, err := cc.catov2.UpdateSiteBwLicense(ctx, cc.AccountId, input)
						if err != nil {
							diags = append(diags, diag.NewErrorDiagnostic("Catov2 API error", err.Error()))
							return nil, err
						}
					} else {
						tflog.Warn(ctx, "License ID and BW matches, nothing to update, reading existing license into state.")
					}
				}
			} else {
				// call replace
				tflog.Warn(ctx, "Site ID '"+fmt.Sprintf("%v", plan.SiteID.ValueString())+"' is currently assigned to '"+fmt.Sprintf("%v", curSiteLicenseId)+"'")
				input := cato_models.ReplaceSiteBwLicenseInput{}
				input.LicenseIDToRemove = curSiteLicenseId.ValueString()
				input.LicenseIDToAdd = plan.LicenseID.ValueString()
				siteRef := &cato_models.SiteRefInput{}
				siteRef.By = "ID"
				siteRef.Input = plan.SiteID.ValueString()
				input.Site = siteRef
				// Check for BW if pooled
				tflog.Warn(ctx, "Checking license SKU type for PB")
				if string(license.Sku) == "CATO_PB" || string(license.Sku) == "CATO_PB_SSE" {
					if plan.BW.IsUnknown() || plan.BW.IsNull() {
						message := "Bandwidth must be set for 'CATO_PB' and 'CATO_PB_SSE' pooled bandwidth license type."
						diags = append(diags, diag.NewErrorDiagnostic("INVALID CONFIGURATION", message))
						return nil, fmt.Errorf("INVALID CONFIGURATION: %s", message)
					}
					bw := plan.BW.ValueInt64()
					tflog.Warn(ctx, "License SKU is '"+string(license.Sku)+"', adding bw '"+fmt.Sprintf("%v", bw)+"'")
					input.Bw = &bw
				}
				tflog.Warn(ctx, "Calling ReplaceSiteBwLicense()")
				_, err := cc.catov2.ReplaceSiteBwLicense(ctx, cc.AccountId, input)
				if err != nil {
					diags = append(diags, diag.NewErrorDiagnostic("Catov2 API error", err.Error()))
					return nil, err
				}
			}
		} else {
			tflog.Warn(ctx, "Site ID '"+fmt.Sprintf("%v", plan.SiteID.ValueString())+"' is not currently assigned to a license, ")
			input := cato_models.AssignSiteBwLicenseInput{}
			input.LicenseID = plan.LicenseID.ValueString()
			siteRef := &cato_models.SiteRefInput{}
			siteRef.By = "ID"
			siteRef.Input = plan.SiteID.ValueString()
			input.Site = siteRef
			// Check for BW if pooled
			tflog.Warn(ctx, "Checking license SKU type for PB")
			if string(license.Sku) == "CATO_PB" || string(license.Sku) == "CATO_PB_SSE" {
				if plan.BW.IsUnknown() || plan.BW.IsNull() {
					message := "Bandwidth must be set for 'CATO_PB' and 'CATO_PB_SSE' pooled bandwidth license type."
					diags = append(diags, diag.NewErrorDiagnostic("INVALID CONFIGURATION", message))
					return nil, fmt.Errorf("INVALID CONFIGURATION: %s", message)
				}
				tflog.Warn(ctx, "License SKU is '"+string(license.Sku)+"', adding bw.")
				bw := plan.BW.ValueInt64()
				input.Bw = &bw
			} else {
				// Check for BW not present if not pooled
				if !plan.BW.IsUnknown() && !plan.BW.IsNull() {
					message := "Bandwidth is not supported for 'CATO_SITE' and 'CATO_SSE_SITE' license type only for 'CATO_PB' pooled bandwidth type."
					diags = append(diags, diag.NewErrorDiagnostic("INVALID CONFIGURATION", message))
					return nil, fmt.Errorf("INVALID CONFIGURATION: %s", message)
				}
				tflog.Warn(ctx, "License SKU is '"+string(license.Sku)+"', bw not present.")
			}
			tflog.Warn(ctx, "Calling AssignSiteBwLicense()")
			_, err := cc.catov2.AssignSiteBwLicense(ctx, cc.AccountId, input)
			if err != nil {
				diags = append(diags, diag.NewErrorDiagnostic("Catov2 API error", err.Error()))
				return nil, err
			}
		}
		return license, nil
	} else {
		message := "License ID '" + plan.LicenseID.ValueString() + "' not found. Either the license ID specificed is not valid, or this account is not synced an does not support license updates via API.  Please contact Cato support."
		diags = append(diags, diag.NewErrorDiagnostic("INVALID LICENSE ID", message))
		return nil, fmt.Errorf("INVALID CONFIGURATION: %s", message)
	}
}
