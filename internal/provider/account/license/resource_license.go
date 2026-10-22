package license

import (
	"context"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
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
