package license

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	licensing "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/licensing"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	Licensing(ctx context.Context, accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.Licensing, error)
	RemoveSiteBwLicense(
		ctx context.Context,
		accountID string,
		input cato_models.RemoveSiteBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.RemoveSiteBwLicense, error)
	EntityLookup(
		ctx context.Context,
		accountID string,
		typeArg cato_models.EntityType,
		limit *int64,
		from *int64,
		parent *cato_models.EntityInput,
		search *string,
		entityIDs []string,
		sort []*cato_models.SortInput,
		filters []*cato_models.LookupFilterInput,
		helperFields []string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.EntityLookup, error)
	UpdateSiteBwLicense(
		ctx context.Context,
		accountID string,
		input cato_models.UpdateSiteBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.UpdateSiteBwLicense, error)
	ReplaceSiteBwLicense(
		ctx context.Context,
		accountID string,
		input cato_models.ReplaceSiteBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ReplaceSiteBwLicense, error)
	AssignSiteBwLicense(
		ctx context.Context,
		accountID string,
		input cato_models.AssignSiteBwLicenseInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AssignSiteBwLicense, error)
}

func (r *licenseResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) Licensing(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.Licensing, error) {
	return c.resource.client.Catov2.Licensing(ctx, accountID, interceptors...)
}

func (c *sdkClient) RemoveSiteBwLicense(
	ctx context.Context,
	accountID string,
	input cato_models.RemoveSiteBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.RemoveSiteBwLicense, error) {
	return c.resource.client.Catov2.RemoveSiteBwLicense(ctx, accountID, input, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *licenseResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) EntityLookup(
	ctx context.Context,
	accountID string,
	typeArg cato_models.EntityType,
	limit *int64,
	from *int64,
	parent *cato_models.EntityInput,
	search *string,
	entityIDs []string,
	sort []*cato_models.SortInput,
	filters []*cato_models.LookupFilterInput,
	helperFields []string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.EntityLookup, error) {
	return c.resource.client.Catov2.EntityLookup(
		ctx,
		accountID,
		typeArg,
		limit,
		from,
		parent,
		search,
		entityIDs,
		sort,
		filters,
		helperFields,
		interceptors...,
	)
}
func (c *sdkClient) UpdateSiteBwLicense(
	ctx context.Context,
	accountID string,
	input cato_models.UpdateSiteBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.UpdateSiteBwLicense, error) {
	return c.resource.client.Catov2.UpdateSiteBwLicense(ctx, accountID, input, interceptors...)
}
func (c *sdkClient) ReplaceSiteBwLicense(
	ctx context.Context,
	accountID string,
	input cato_models.ReplaceSiteBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ReplaceSiteBwLicense, error) {
	return c.resource.client.Catov2.ReplaceSiteBwLicense(ctx, accountID, input, interceptors...)
}
func (c *sdkClient) AssignSiteBwLicense(
	ctx context.Context,
	accountID string,
	input cato_models.AssignSiteBwLicenseInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AssignSiteBwLicense, error) {
	return c.resource.client.Catov2.AssignSiteBwLicense(ctx, accountID, input, interceptors...)
}

//nolint:gocyclo,funlen,ineffassign,lll,staticcheck
func upsertLicense(
	ctx context.Context,
	plan Resource,
	cc *catoclient.Service,
	clients ...ResourceClient,
) (*cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses, error) {
	var api ResourceClient = cc.Catov2
	if len(clients) > 0 {
		api = clients[0]
	}

	diags := make(diag.Diagnostics, 0)
	// Get all sites, check for valid siteID
	siteExists := false
	thousandInt64 := int64(1000)
	tenThousandInt64 := int64(10000) // to cap the lookup and avoid an infinite loop at current scale.
	fromInt64 := int64(0)
	processedItems := int64(0)
	for !siteExists && processedItems < tenThousandInt64 {
		siteResponse, err := api.EntityLookup(
			ctx,
			cc.AccountId,
			cato_models.EntityTypeSite,
			&thousandInt64,
			&fromInt64,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		)
		tflog.Warn(ctx, "upsertLicense().EntityLookup.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(siteResponse),
		})

		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic("Catov2 API error", err.Error()))
			return nil, err
		}

		items := siteResponse.GetEntityLookup().GetItems()
		for _, item := range items {
			tflog.Warn(ctx, "Checking site IDs, input='"+plan.SiteID.ValueString()+"', currentItem='"+item.GetEntity().GetID()+"'")
			if item.GetEntity().GetID() == plan.SiteID.ValueString() {
				tflog.Warn(ctx, "Site ID matched! "+item.GetEntity().GetID())
				siteExists = true
				break
			}
		}

		processedItems += int64(len(items))
		if len(items) < int(thousandInt64) {
			break
		}

		fromInt64 += thousandInt64
	}
	if !siteExists {
		message := "Site '" + plan.SiteID.ValueString() + "' not found."
		diags = append(diags, diag.NewErrorDiagnostic("INVALID SITE ID", message))
		return nil, fmt.Errorf("INVALID SITE ID: %s", message)
	}

	// Get all licenses
	licensingInfoResponse, err := api.Licensing(ctx, cc.AccountId)
	tflog.Warn(ctx, "upsertLicense().Licensing.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(licensingInfoResponse),
	})

	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic("Catov2 API error", err.Error()))
		return nil, err
	}

	licenses := licensingInfoResponse.Licensing.LicensingInfo.Licenses
	if len(licenses) == 0 {
		message := "No licenses were found on this account."
		diags = append(diags, diag.NewErrorDiagnostic("LICENSE API ERROR", message))
		return nil, fmt.Errorf("LICENSE API ERROR: %s", message)
	}

	// Check if the site has a license currently
	curSiteLicenseID, allocatedBw, siteIsAssigned := getCurrentAssignedLicenseBySiteID(ctx, plan.SiteID.ValueString(), licensingInfoResponse)

	// Get current license objeect by ID
	license, licenseExists := getLicenseByID(ctx, plan.LicenseID.ValueString(), licensingInfoResponse)

	// Check if site license is currently assigned
	curLicenseSiteID, siteLicenseCurrentlyAssigned := checkStaticLicenseForAssignment(ctx, plan.LicenseID.ValueString(), licensingInfoResponse)
	tflog.Debug(ctx, "checkStaticLicenseForAssignment().result", map[string]interface{}{
		"curLicenseSiteId":                   curLicenseSiteID.String(),
		"siteLicenseCurrentlyAssigned":       fmt.Sprintf("%v", siteLicenseCurrentlyAssigned),
		"siteLicenseCurrentlyAssigned==true": fmt.Sprintf("%t", siteLicenseCurrentlyAssigned),
	})
	if siteLicenseCurrentlyAssigned {
		if plan.SiteID.ValueString() == curLicenseSiteID.ValueString() {
			tflog.Warn(
				ctx,
				"License '"+fmt.Sprintf(
					"%v",
					license.ID,
				)+"', already assigned to site ID '"+plan.SiteID.ValueString()+"', returning license successfully applied.",
			)
			return license, nil
		}
		message := "The license ID '" + plan.LicenseID.ValueString() + "' is already assigned to site ID " + curLicenseSiteID.String()
		diags = append(diags, diag.NewErrorDiagnostic("LICENSE ALREADY ASSIGNED", message))
		return nil, fmt.Errorf("LICENSE ALREADY ASSIGNED: %s", message)
	}
	if licenseExists {
		// Check for the correct license type
		tflog.Warn(ctx, "Checking license SKU for CATO_PB, CATO_PB_SSE, CATO_SITE, or CATO_SSE_SITE type, license.Sku='"+string(license.Sku)+"'")
		if string(license.Sku) != licensing.SKUCatoPB &&
			string(license.Sku) != licensing.SKUCatoPBSSE &&
			string(license.Sku) != licensing.SKUCatoSite &&
			string(license.Sku) != licensing.SKUCatoSSESite {
			message := "Site License ID '" + plan.LicenseID.ValueString() + "' is not a valid site license. Must be 'CATO_PB', 'CATO_PB_SSE', 'CATO_SITE', or 'CATO_SSE_SITE' license type."
			diags = append(diags, diag.NewErrorDiagnostic("INVALID LICENSE TYPE", message))
			return nil, fmt.Errorf("INVALID LICENSE TYPE: %s", message)
		}

		// A site can have allocations from multiple pooled licenses. When the
		// selected pool is already assigned, scope the update and comparison to
		// that allocation instead of whichever site allocation was listed last.
		if string(license.Sku) == licensing.SKUCatoPB || string(license.Sku) == licensing.SKUCatoPBSSE {
			selectedAllocatedBw, selectedPoolIsAssigned := getSiteAssignmentForLicense(plan.SiteID.ValueString(), license)
			if selectedPoolIsAssigned {
				allocatedBw = selectedAllocatedBw
				siteIsAssigned = true
				curSiteLicenseID = plan.LicenseID
			}
		}

		if siteIsAssigned && curSiteLicenseID.IsNull() {
			message := "License ID not found: This could be due to the license or the account being set as trial where the license does not have an ID, or if there is a China license on the account.  If a trial license is assigned, please unassign the trial license and try to reapply."
			diags = append(diags, diag.NewErrorDiagnostic("LICENSE API ERROR", message))
			return nil, fmt.Errorf("LICENSE API ERROR: %s", message)
		}

		// check if assigned, use replace else use assign
		if siteIsAssigned {
			// if the site is already assigned to this license
			if curSiteLicenseID.ValueString() == plan.LicenseID.ValueString() {
				tflog.Warn(ctx, "Site ID '"+plan.SiteID.ValueString()+"' is already assigned to license ID '"+fmt.Sprintf("%v", license.ID)+"'")
				if string(license.Sku) == licensing.SKUCatoPB || string(license.Sku) == licensing.SKUCatoPBSSE {
					if plan.BW.ValueInt64() != *allocatedBw {
						// Checking license SKU type for PB
						tflog.Warn(ctx, "Pooled bandwidth license identfied '"+license.Sku.String()+"', assigning BW")
						if plan.BW.IsUnknown() || plan.BW.IsNull() {
							message := bandwidthRequiredError
							diags = append(diags, diag.NewErrorDiagnostic("INVALID CONFIGURATION", message))
							return nil, fmt.Errorf("INVALID CONFIGURATION: %s", message)
						}
						input := cato_models.UpdateSiteBwLicenseInput{}
						input.LicenseID = plan.LicenseID.ValueString()
						siteRef := &cato_models.SiteRefInput{}
						siteRef.By = licenseSiteRefByID
						siteRef.Input = plan.SiteID.ValueString()
						input.Site = siteRef
						input.Bw = plan.BW.ValueInt64()
						tflog.Warn(ctx, "upsertLicense().UpdateSiteBwLicense.request", map[string]interface{}{
							"request": utils.InterfaceToJSONString(input),
						})
						UpdateSiteBwLicenseResponse, err := api.UpdateSiteBwLicense(ctx, cc.AccountId, input)
						tflog.Warn(ctx, "upsertLicense().UpdateSiteBwLicense.response", map[string]interface{}{
							"response": utils.InterfaceToJSONString(UpdateSiteBwLicenseResponse),
						})
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
				tflog.Warn(ctx, "Site ID '"+plan.SiteID.ValueString()+"' is currently assigned to '"+curSiteLicenseID.String()+"'")
				input := cato_models.ReplaceSiteBwLicenseInput{}
				input.LicenseIDToRemove = curSiteLicenseID.ValueString()
				input.LicenseIDToAdd = plan.LicenseID.ValueString()
				siteRef := &cato_models.SiteRefInput{}
				siteRef.By = licenseSiteRefByID
				siteRef.Input = plan.SiteID.ValueString()
				input.Site = siteRef
				// Check for BW if pooled
				tflog.Warn(ctx, "Checking license SKU type for PB")
				if string(license.Sku) == licensing.SKUCatoPB || string(license.Sku) == licensing.SKUCatoPBSSE {
					if plan.BW.IsUnknown() || plan.BW.IsNull() {
						message := bandwidthRequiredError
						diags = append(diags, diag.NewErrorDiagnostic("INVALID CONFIGURATION", message))
						return nil, fmt.Errorf("INVALID CONFIGURATION: %s", message)
					}
					bw := plan.BW.ValueInt64()
					tflog.Warn(ctx, "License SKU is '"+string(license.Sku)+"', adding bw '"+fmt.Sprintf("%v", bw)+"'")
					input.Bw = &bw
				}
				tflog.Warn(ctx, "upsertLicense().ReplaceSiteBwLicense.request", map[string]interface{}{
					"request": utils.InterfaceToJSONString(input),
				})
				replaceSiteBwLicenseResponse, err := api.ReplaceSiteBwLicense(ctx, cc.AccountId, input)
				tflog.Warn(ctx, "upsertLicense().ReplaceSiteBwLicense.response", map[string]interface{}{
					"response": utils.InterfaceToJSONString(replaceSiteBwLicenseResponse),
				})
				if err != nil {
					diags = append(diags, diag.NewErrorDiagnostic("Catov2 API error", err.Error()))
					return nil, err
				}
			}
		} else {
			tflog.Warn(ctx, "Site ID '"+plan.SiteID.ValueString()+"' is not currently assigned to a license, ")
			input := cato_models.AssignSiteBwLicenseInput{}
			input.LicenseID = plan.LicenseID.ValueString()
			siteRef := &cato_models.SiteRefInput{}
			siteRef.By = licenseSiteRefByID
			siteRef.Input = plan.SiteID.ValueString()
			input.Site = siteRef
			// Check for BW if pooled
			tflog.Warn(ctx, "Checking license SKU type for PB")
			if string(license.Sku) == licensing.SKUCatoPB || string(license.Sku) == licensing.SKUCatoPBSSE {
				if plan.BW.IsUnknown() || plan.BW.IsNull() {
					message := bandwidthRequiredError
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
			tflog.Warn(ctx, "upsertLicense().AssignSiteBwLicense.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(input),
			})
			assignSiteBwLicenseResponse, err := api.AssignSiteBwLicense(ctx, cc.AccountId, input)
			tflog.Warn(ctx, "upsertLicense().AssignSiteBwLicense.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(assignSiteBwLicenseResponse),
			})
			if err != nil {
				diags = append(diags, diag.NewErrorDiagnostic("Catov2 API error", err.Error()))
				return nil, err
			}
		}
		return license, nil
	}
	message := "License ID '" + plan.LicenseID.ValueString() + "' not found. Either the license ID specificed is not valid, or this account is not synced an does not support license updates via API.  Please contact Cato support."
	diags = append(diags, diag.NewErrorDiagnostic("INVALID LICENSE ID", message))
	return nil, fmt.Errorf("INVALID CONFIGURATION: %s", message)
}
