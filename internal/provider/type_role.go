package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Role is the shared Terraform representation of an account-scoped IAM role.
type Role struct {
	ID                     types.String `tfsdk:"id"`
	AccountID              types.String `tfsdk:"account_id"`
	RoleID                 types.String `tfsdk:"role_id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	Permissions            types.Set    `tfsdk:"permissions"`
	Predefined             types.Bool   `tfsdk:"predefined"`
	AccountType            types.String `tfsdk:"account_type"`
	IsUsedOnExternalAccess types.Bool   `tfsdk:"is_used_on_external_access"`
}

var rolePermissionAttrTypes = map[string]attr.Type{"resource": types.StringType, "action": types.StringType}
var rolePermissionObjectType = types.ObjectType{AttrTypes: rolePermissionAttrTypes}
var roleAttrTypes = map[string]attr.Type{
	"id": types.StringType, "account_id": types.StringType, "role_id": types.StringType,
	"name": types.StringType, "description": types.StringType,
	"permissions": types.SetType{ElemType: rolePermissionObjectType},
	"predefined":  types.BoolType, "account_type": types.StringType, "is_used_on_external_access": types.BoolType,
}
var roleObjectType = types.ObjectType{AttrTypes: roleAttrTypes}

func rolePermissionsInput(_ context.Context, permissions types.Set) ([]*cato_models.RoleManagementPermissionInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	if permissions.IsNull() || permissions.IsUnknown() {
		diags.AddError("Invalid role permissions", "permissions must be known. Use an explicit empty set when granting no permissions.")
		return nil, diags
	}
	result := make([]*cato_models.RoleManagementPermissionInput, 0, len(permissions.Elements()))
	seen := make(map[string]struct{})
	for _, value := range permissions.Elements() {
		object, ok := value.(types.Object)
		if !ok || object.IsNull() || object.IsUnknown() {
			diags.AddError("Invalid role permission", "Each permission must be a known object.")
			continue
		}
		attrs := object.Attributes()
		resourceName := attrs["resource"].(types.String)
		action := attrs["action"].(types.String)
		if resourceName.IsNull() || resourceName.IsUnknown() || resourceName.ValueString() == "" ||
			action.IsNull() || action.IsUnknown() || (action.ValueString() != "VIEW" && action.ValueString() != "EDIT") {
			diags.AddError("Invalid role permission", "Each permission requires a resource and a VIEW or EDIT action.")
			continue
		}
		if _, duplicate := seen[resourceName.ValueString()]; duplicate {
			diags.AddError("Duplicate role permission", "Each resource may appear only once in permissions.")
			continue
		}
		seen[resourceName.ValueString()] = struct{}{}
		result = append(result, &cato_models.RoleManagementPermissionInput{
			Resource: resourceName.ValueString(),
			Action:   cato_models.RoleManagementGrantAction(action.ValueString()),
		})
	}
	return result, diags
}

type rolePermissionResponse interface {
	GetResource() string
	GetAction() *cato_models.RBACAction
}
type roleResponse[P rolePermissionResponse] interface {
	GetID() string
	GetName() string
	GetDescription() *string
	GetPredefined() bool
	GetAccountType() *cato_models.RoleManagementAccountType
	GetIsUsedOnExternalAccess() bool
	GetPermission() []P
}

func roleFromAPI[P rolePermissionResponse](_ context.Context, accountID string, value roleResponse[P]) (Role, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := Role{
		ID:                     types.StringValue(accountID + ":" + value.GetID()),
		AccountID:              types.StringValue(accountID),
		RoleID:                 types.StringValue(value.GetID()),
		Name:                   types.StringValue(value.GetName()),
		Description:            types.StringValue(""),
		Predefined:             types.BoolValue(value.GetPredefined()),
		IsUsedOnExternalAccess: types.BoolValue(value.GetIsUsedOnExternalAccess()),
	}
	if value.GetID() == "" || value.GetAccountType() == nil || *value.GetAccountType() == "" {
		diags.AddError("Invalid role response", "The API returned a role without an identifier or account type.")
		return model, diags
	}
	model.AccountType = types.StringValue(string(*value.GetAccountType()))
	if value.GetDescription() != nil {
		model.Description = types.StringValue(*value.GetDescription())
	}
	permissions := make([]attr.Value, 0, len(value.GetPermission()))
	for _, permission := range value.GetPermission() {
		if permission.GetResource() == "" || permission.GetAction() == nil || *permission.GetAction() == "" {
			diags.AddError("Invalid role response", "The API returned an incomplete permission.")
			return model, diags
		}
		object, d := types.ObjectValue(rolePermissionAttrTypes, map[string]attr.Value{
			"resource": types.StringValue(permission.GetResource()),
			"action":   types.StringValue(string(*permission.GetAction())),
		})
		diags.Append(d...)
		permissions = append(permissions, object)
	}
	var d diag.Diagnostics
	model.Permissions, d = types.SetValue(rolePermissionObjectType, permissions)
	diags.Append(d...)
	return model, diags
}

func roleObject(ctx context.Context, model Role) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, roleAttrTypes, model)
}

func effectiveRoleAccount(accountID types.String, client *catoClientData) (string, error) {
	if !accountID.IsNull() && !accountID.IsUnknown() && accountID.ValueString() != "" {
		return accountID.ValueString(), nil
	}
	if client == nil || client.AccountId == "" {
		return "", fmt.Errorf("configure the provider account_id or specify account_id on the role")
	}
	return client.AccountId, nil
}
