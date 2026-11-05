package user

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = (*userResource)(nil)
	_ resource.ResourceWithConfigure   = (*userResource)(nil)
	_ resource.ResourceWithImportState = (*userResource)(nil)
)

type userResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the local-user resource.
func NewResource() resource.Resource { return &userResource{} }
func (r *userResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *userResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}
func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func mutableUser(value remoteUser, diags *diag.Diagnostics) bool {
	if sourceOf(value) != cato_models.UserImportTypeManual {
		diags.AddError("Unsupported user source", "Only local MANUAL users can be managed. Directory-managed users are not supported.")
		return false
	}
	return true
}
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreate(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().UserCreateUser(ctx, r.client.AccountId, input)
	user := result.GetUser().GetCreateUser().GetUser()
	if err != nil {
		if user.GetID() != "" {
			resp.Diagnostics.Append(resp.State.Set(ctx, identityState(plan, user.GetID()))...)
		}
		resp.Diagnostics.AddError("Failed to create user", err.Error())
		return
	}
	if user.GetID() == "" {
		resp.Diagnostics.AddError("Invalid API response", "Create did not return a user ID.")
		return
	}
	current := parseUser(user, &resp.Diagnostics)
	// Persist the returned identity before any follow-up enable/disable operation.
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	if resp.Diagnostics.HasError() || !validUserID(current.ID, &resp.Diagnostics) || !mutableUser(user, &resp.Diagnostics) {
		return
	}
	if current.Enabled.ValueBool() != plan.Enabled.ValueBool() {
		user, err := r.changeEnabled(ctx, current.ID.ValueString(), plan.Enabled.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Failed to set user enabled state", err.Error())
			return
		}
		current = parseUser(user, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validUserID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	user, err := r.findUser(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read user", err.Error())
		return
	}
	if user == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if !mutableUser(user, &resp.Diagnostics) {
		return
	}
	current := parseUser(user, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validUserID(state.ID, &resp.Diagnostics) {
		return
	}
	input := buildUpdate(plan, state.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.findUser(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read user before update", err.Error())
		return
	}
	if existing == nil {
		resp.Diagnostics.AddError("User no longer exists", "Refresh state before applying again.")
		return
	}
	if !mutableUser(existing, &resp.Diagnostics) {
		return
	}
	result, err := r.getClient().UserUpdateUser(ctx, r.client.AccountId, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update user", err.Error())
		return
	}
	user := result.GetUser().GetUpdateUser().GetUser()
	if user.GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update returned a different user identity.")
		return
	}
	var updated remoteUser = user
	current := parseUser(updated, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if current.Enabled.ValueBool() != plan.Enabled.ValueBool() {
		updated, err = r.changeEnabled(ctx, state.ID.ValueString(), plan.Enabled.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Failed to set user enabled state", err.Error())
			return
		}
		current = parseUser(updated, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() || !validUserID(id, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	user, err := r.findUser(ctx, id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read user before deletion", err.Error())
		return
	}
	if user == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if !mutableUser(user, &resp.Diagnostics) {
		return
	}
	result, err := r.getClient().UserDeleteUser(ctx, r.client.AccountId,
		cato_models.DeleteUserInput{UserID: []scalars.Long{scalars.Long(id.ValueString())}},
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete user", err.Error())
		return
	}
	if status := result.GetUser().GetDeleteUser().GetStatus(); status == nil || *status != cato_models.DeleteUserStatusSuccess {
		resp.Diagnostics.AddError("Failed to delete user", "The API did not confirm successful deletion.")
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *userResource) changeEnabled(ctx context.Context, id string, enabled bool) (remoteUser, error) {
	var users []remoteUser
	ids := []scalars.Long{scalars.Long(id)}
	if enabled {
		result, err := r.getClient().UserEnableUser(ctx, r.client.AccountId, cato_models.EnableUserInput{UserID: ids})
		if err != nil {
			return nil, err
		}
		for _, user := range result.GetUser().GetEnableUser().GetUsers() {
			users = append(users, user)
		}
	} else {
		result, err := r.getClient().UserDisableUser(ctx, r.client.AccountId, cato_models.DisableUserInput{UserID: ids})
		if err != nil {
			return nil, err
		}
		for _, user := range result.GetUser().GetDisableUser().GetUsers() {
			users = append(users, user)
		}
	}
	for _, user := range users {
		if user.GetID() == id {
			if sourceOf(user) != cato_models.UserImportTypeManual {
				return nil, fmt.Errorf("API returned a directory-managed user")
			}
			active := statusOf(user) == cato_models.UserStatusActive || statusOf(user) == cato_models.UserStatusLocked
			if !statusOf(user).IsValid() || active != enabled {
				return nil, fmt.Errorf("API did not confirm the requested user status")
			}
			return user, nil
		}
	}
	return nil, fmt.Errorf("API did not return the requested user identity")
}
