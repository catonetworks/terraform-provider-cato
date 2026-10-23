package account

import (
	"context"
	json "encoding/json"

	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &accountResource{}
	_ resource.ResourceWithConfigure   = &accountResource{}
	_ resource.ResourceWithImportState = &accountResource{}
)

func NewAccountResource() resource.Resource {
	return &accountResource{}
}

type accountResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *accountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (r *accountResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *accountResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *accountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Account
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := convertCreateInput(plan)

	// publishing new section
	tflog.Info(ctx, "Create.account")
	tflog.Debug(ctx, "Create.Account.request", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	addAccountResult, err := r.resourceAPI().AccountManagementAddAccount(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.Account.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(addAccountResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			err.Error(),
		)
		return
	}

	// Validate account creation result
	if addAccountResult.AccountManagement.AddAccount == nil || addAccountResult.AccountManagement.AddAccount.ID == "" {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			"Account creation failed - no account data returned",
		)
		return
	}

	// Use the account data directly from the create response
	createdAccount := addAccountResult.AccountManagement.AddAccount
	tflog.Debug(ctx, "Account created successfully", map[string]interface{}{
		"account_id":   createdAccount.ID,
		"account_data": utils.InterfaceToJSONString(createdAccount),
	})

	// Update the state with the complete account information from the create response
	convertCreateInput2(&plan, createdAccount)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *accountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Account
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readAccountResponse, err := r.resourceAPI().AccountManagement(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Read.Account.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(readAccountResponse),
	})

	convertReadState(&state, readAccountResponse)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *accountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Account
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Account
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if any immutable fields have changed
	if !plan.Name.Equal(state.Name) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Account name cannot be changed after creation. Current: "+state.Name.ValueString()+", Requested: "+plan.Name.ValueString(),
		)
		return
	}

	if !plan.Tenancy.Equal(state.Tenancy) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Account tenancy cannot be changed after creation. Current: "+state.Tenancy.ValueString()+", Requested: "+plan.Tenancy.ValueString(),
		)
		return
	}

	if !plan.Timezone.Equal(state.Timezone) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Account timezone cannot be changed after creation. Current: "+state.Timezone.ValueString()+", Requested: "+plan.Timezone.ValueString(),
		)
		return
	}

	if !plan.Type.Equal(state.Type) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Account type cannot be changed after creation. Current: "+state.Type.ValueString()+", Requested: "+plan.Type.ValueString(),
		)
		return
	}

	// Check if description has actually changed
	if plan.Description.Equal(state.Description) {
		// No changes needed, just return the current state
		diags = resp.State.Set(ctx, state)
		resp.Diagnostics.Append(diags...)
		return
	}

	input := convertUpdateInput(plan)

	// publishing new section
	tflog.Info(ctx, "Update.account")
	tflog.Debug(ctx, "Update.Account.request", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	updateAccountResult, err := r.resourceAPI().AccountManagementUpdateAccount(ctx, input, plan.ID.ValueString())
	tflog.Debug(ctx, "Update.Account.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateAccountResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			err.Error(),
		)
		return
	}

	// Validate account update result
	if updateAccountResult.AccountManagement.UpdateAccount == nil {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			"Account update failed - no account data returned",
		)
		return
	}

	// Use the account data directly from the update response
	updatedAccount := updateAccountResult.AccountManagement.UpdateAccount
	tflog.Debug(ctx, "Account updated successfully", map[string]interface{}{
		"account_id":   updatedAccount.ID,
		"account_data": utils.InterfaceToJSONString(updatedAccount),
	})

	// Update the state with the complete account information from the update response
	convertUpdateInput2(&plan, updatedAccount)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *accountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Account
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Delete.AccountManagementRemoveAccount.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(state.ID),
	})
	removedAccountResult, err := r.resourceAPI().AccountManagementRemoveAccount(ctx, state.ID.ValueString(), r.client.AccountId)
	tflog.Debug(ctx, "Delete.AccountManagementRemoveAccount.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removedAccountResult),
	})
	if err != nil {
		var apiError struct {
			NetworkErrors interface{} `json:"networkErrors"`
			GraphqlErrors []struct {
				Message string   `json:"message"`
				Path    []string `json:"path"`
			} `json:"graphqlErrors"`
		}
		isDisabled := false
		if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphqlErrors) > 0 {
			if apiError.GraphqlErrors[0].Message == "Please disable account before deleting" {
				isDisabled = true
			}
		}
		tflog.Debug(ctx, "Checking for account not ready for deletion, if not expired yet, gracefully failing deleting resource from state.")
		if isDisabled {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
			return
		}
	}
}
