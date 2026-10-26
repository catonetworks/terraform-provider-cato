package license

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	licensing "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/licensing"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &licenseResource{}
	_ resource.ResourceWithConfigure   = &licenseResource{}
	_ resource.ResourceWithImportState = &licenseResource{}
)

const (
	licenseSiteRefByID  = "ID"
	licenseStatusActive = "ACTIVE"
)

func NewLicenseResource() resource.Resource {
	return &licenseResource{}
}

type licenseResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *licenseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

func (r *licenseResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *licenseResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *licenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Resource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	curLicense, err := upsertLicense(ctx, plan, r.client, r.resourceAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error updating license", err.Error())
		return
	}

	licenseInfo := convertCreateInput(ctx, &plan, &diags, curLicense)
	if diags.HasError() {
		return
	}
	licenseInfoObject := convertCreateInput2(ctx, &diags, licenseInfo)
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
	var state Resource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get all licenses
	licensingInfoResponse, err := r.resourceAPI().Licensing(ctx, r.client.AccountId)
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
	convertReadState(ctx, &state, license, curSiteLicenseID, allocatedBw, siteIsAssigned, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
}

func (r *licenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Resource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	curLicense, err := upsertLicense(ctx, plan, r.client, r.resourceAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error updating license", err.Error())
		return
	}

	licenseInfo := convertUpdateInput(ctx, &plan, &diags, curLicense)
	if diags.HasError() {
		return
	}
	licenseInfoObject := convertUpdateInput2(ctx, &diags, licenseInfo)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	convertUpdateInput3(&plan, licenseInfoObject)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
}

func (r *licenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Resource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Disabled interface to "remove" an interface
	input := convertDeleteInput(state)

	tflog.Debug(ctx, "Delete.RemoveSiteBwLicense.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	RemoveSiteBwLicenseResponse, err := r.resourceAPI().RemoveSiteBwLicense(ctx, r.client.AccountId, input)
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

const (
	bandwidthRequiredError = "Bandwidth must be set for 'CATO_PB' and 'CATO_PB_SSE' pooled bandwidth license type."
)

//nolint:gocyclo,funlen,ineffassign,lll,staticcheck
func upsertLicense(
	ctx context.Context,
	plan Resource,
	cc *catoclient.Service,
	clients ...ResourceClient,
) (*cato_go_sdk.Licensing_Licensing_LicensingInfo_Licenses, error) {
	api := licenseAPI(cc, clients...)

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
