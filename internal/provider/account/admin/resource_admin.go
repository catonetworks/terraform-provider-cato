package admin

import (
	"context"
	json "encoding/json"
	"regexp"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &adminResource{}
	_ resource.ResourceWithConfigure   = &adminResource{}
	_ resource.ResourceWithImportState = &adminResource{}
)

const adminImportIDParts = 2

func NewAdminResource() resource.Resource {
	return &adminResource{}
}

type adminResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *adminResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin"
}

func (r *adminResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *adminResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *adminResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Admin
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	curAccountID, input := convertCreateInput(
		ctx,
		r,
		plan,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	} // Always true for admin users

	// Log the request
	tflog.Info(ctx, "Create.admin")
	tflog.Debug(ctx, "Create.Admin.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})

	// Call the API
	addAdminResult, err := r.resourceAPI().AdminAddAdmin(ctx, input, curAccountID)
	tflog.Debug(ctx, "Create.Admin.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(addAdminResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			err.Error(),
		)
		return
	}

	// Validate admin creation result
	if addAdminResult.GetAdmin() == nil || addAdminResult.GetAdmin().AddAdmin == nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			"Admin creation failed - no admin data returned",
		)
		return
	}

	// Get the admin ID from the response
	adminID := addAdminResult.GetAdmin().AddAdmin.GetAdminID()
	if adminID == "" {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			"Admin creation failed - no admin ID returned",
		)
		return
	}

	tflog.Debug(ctx, "Admin created successfully", map[string]interface{}{
		"admin_id": adminID,
	})

	// Update the state with the admin ID
	convertCreateInput2(&plan, curAccountID, adminID)

	// Read the admin data to populate computed fields
	readAdminResult, err := r.resourceAPI().Admin(ctx, curAccountID, adminID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			"Failed to read admin after creation: "+err.Error(),
		)
		return
	}

	convertCreateInput3(ctx, &plan, curAccountID, readAdminResult, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:gocyclo
func (r *adminResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Admin
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := regexp.MustCompile(":").Split(state.ID.ValueString(), 2)
	curAccountID := ""
	curAdminID := ""
	if len(idParts) == adminImportIDParts {
		curAccountID = idParts[0]
		curAdminID = idParts[1]
	}

	// Fail and return if either curAccountID or curAdminID is empty
	if curAccountID == "" || curAdminID == "" {
		resp.Diagnostics.AddError(
			"Invalid Admin ID",
			"Failed to parse admin ID: either account ID or admin ID is empty.",
		)
		return
	}

	// Log the request
	tflog.Info(ctx, "Read.admin")
	tflog.Debug(ctx, "Read.Admin.request", map[string]interface{}{
		"admin_id":   curAdminID,
		"account_id": curAccountID,
	})

	// Call the API to read the admin
	readAdminResult, err := r.resourceAPI().Admins(ctx, curAccountID, nil, nil, nil, nil, []string{curAdminID})
	tflog.Debug(ctx, "Read.Admins.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(readAdminResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			err.Error(),
		)
		return
	}

	var adminData *cato_models.Admin
	if readAdminResult != nil && readAdminResult.Admins != nil && readAdminResult.Admins.Items != nil {
		for _, admin := range readAdminResult.Admins.Items {
			if admin != nil && admin.ID == curAdminID {
				// Convert or map admin (*cato_go_sdk.Admins_Admins_Items) to *cato_models.Admin
				adminBytes, err := json.Marshal(admin)
				if err == nil {
					var mappedAdmin cato_models.Admin
					if err := json.Unmarshal(adminBytes, &mappedAdmin); err == nil {
						adminData = &mappedAdmin
					}
				}
				break
			}
		}
	}

	if adminData != nil {
		matchedAdminJSON, _ := json.MarshalIndent(adminData, "", "  ")
		tflog.Debug(ctx, "Matched admin record", map[string]interface{}{
			"response": string(matchedAdminJSON),
		})

		// Update the state with data from the API response
		convertReadValues(&state, curAccountID, curAdminID, adminData)
	} else {
		resp.Diagnostics.AddError("Admin not found", "The admin with the given ID does not exist.")
		return
	}

	// Handle managed roles
	convertReadState(ctx, &state, adminData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the updated state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *adminResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Admin
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Admin
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := regexp.MustCompile(":").Split(state.ID.ValueString(), 2)
	curAccountID := ""
	curAdminID := ""
	if len(idParts) == adminImportIDParts {
		curAccountID = idParts[0]
		curAdminID = idParts[1]
	}

	// Fail and return if either curAccountID or curAdminID is empty
	if curAccountID == "" || curAdminID == "" {
		resp.Diagnostics.AddError(
			"Invalid Admin ID",
			"Failed to parse admin ID: either account ID or admin ID is empty.",
		)
		return
	}

	// Log the request
	tflog.Info(ctx, "Read.admin")
	tflog.Debug(ctx, "Read.Admin.request", map[string]interface{}{
		"admin_id":   curAdminID,
		"account_id": curAccountID,
	})

	// Check if any immutable fields have changed
	input := convertUpdateInput(ctx, plan, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	} // Always true for admin users

	// Log the request
	tflog.Info(ctx, "Update.admin")
	tflog.Debug(ctx, "Update.Admin.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})

	// Call the API
	updateAdminResult, err := r.resourceAPI().AdminUpdateAdmin(ctx, curAdminID, input, curAccountID)
	tflog.Debug(ctx, "Update.Admin.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateAdminResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			err.Error(),
		)
		return
	}

	// Validate admin update result
	if updateAdminResult.GetAdmin() == nil || updateAdminResult.GetAdmin().UpdateAdmin == nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			"Admin update failed - no admin data returned",
		)
		return
	}

	tflog.Debug(ctx, "Admin updated successfully", map[string]interface{}{
		"admin_id":   curAdminID,
		"account_id": curAccountID,
	})

	// Update the state with the admin ID (should be the same)
	convertUpdateInput2(&plan, curAccountID, curAdminID)

	// Set the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *adminResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Admin
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := regexp.MustCompile(":").Split(state.ID.ValueString(), 2)
	curAccountID := ""
	curAdminID := ""
	if len(idParts) == adminImportIDParts {
		curAccountID = idParts[0]
		curAdminID = idParts[1]
	}

	// Fail and return if either curAccountID or curAdminID is empty
	if curAccountID == "" || curAdminID == "" {
		resp.Diagnostics.AddError(
			"Invalid Admin ID",
			"Failed to parse admin ID: either account ID or admin ID is empty.",
		)
		return
	}

	// Log the request
	tflog.Info(ctx, "Read.admin")
	tflog.Debug(ctx, "Read.Admin.request", map[string]interface{}{
		"admin_id":   curAdminID,
		"account_id": curAccountID,
	})

	// Log the request
	tflog.Info(ctx, "Delete.admin")
	tflog.Debug(ctx, "Delete.Admin.request", map[string]interface{}{
		"admin_id": state.ID.ValueString(),
	})

	// Call the API
	removeAdminResult, err := r.resourceAPI().AdminRemoveAdmin(ctx, curAdminID, curAccountID)
	tflog.Debug(ctx, "Delete.Admin.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeAdminResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Admin deleted successfully", map[string]interface{}{
		"admin_id": state.ID.ValueString(),
	})
}
