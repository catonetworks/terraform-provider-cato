package iam

import (
	"context"
	"fmt"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PermissionsInput preserves explicit empty grants and rejects duplicate resources.
func PermissionsInput(ctx context.Context, value types.Set) ([]*m.RoleManagementPermissionInput, diag.Diagnostics) {
	var d diag.Diagnostics
	result := make([]*m.RoleManagementPermissionInput, 0)
	if value.IsNull() || value.IsUnknown() {
		d.AddError("Invalid permissions", "Configure a known permissions set; use [] for no permissions.")
		return result, d
	}
	seen := map[string]bool{}
	for _, entry := range value.Elements() {
		object, ok := entry.(types.Object)
		if !ok || object.IsNull() || object.IsUnknown() {
			d.AddError("Invalid permission", "Each permission must be a known object.")
			continue
		}
		attrs := object.Attributes()
		resource := attrs["resource"].(types.String)
		action := attrs["action"].(types.String)
		if resource.IsNull() || resource.IsUnknown() || resource.ValueString() == "" || action.IsUnknown() ||
			(action.ValueString() != "VIEW" && action.ValueString() != "EDIT") || seen[resource.ValueString()] {
			d.AddError("Invalid permission", "Each resource must occur once with a known VIEW or EDIT action.")
			continue
		}
		seen[resource.ValueString()] = true
		result = append(result,
			&m.RoleManagementPermissionInput{Resource: resource.ValueString(),
				Action: m.RoleManagementGrantAction(action.ValueString())})
	}
	_ = ctx
	return result, d
}

// RoleState validates and converts a complete API projection.
func RoleState(ctx context.Context, account string, value *RoleRecord) (Role, diag.Diagnostics) {
	var d diag.Diagnostics
	var result Role
	if value == nil || value.ID == "" || value.Name == "" || value.Permissions == nil || value.AccountTypes == nil {
		d.AddError("Invalid role response", "The API returned an incomplete role.")
		return result, d
	}
	result.ID = types.StringValue(account + ":" + value.ID)
	result.AccountID = types.StringValue(account)
	result.RoleID = types.StringValue(value.ID)
	result.Name = types.StringValue(value.Name)
	result.Description = types.StringValue("")
	if value.Description != nil {
		result.Description = types.StringValue(*value.Description)
	}
	result.Predefined = types.BoolValue(value.Predefined)
	result.IsUsedOnExternalAccess = types.BoolValue(value.External)
	permissions, permissionDiags := permissionState(value.Permissions)
	result.Permissions = permissions
	d.Append(permissionDiags...)
	var diags diag.Diagnostics
	for _, v := range value.AccountTypes {
		if v != "CUSTOMER" && v != "PARTNER" {
			d.AddError("Invalid role response", fmt.Sprintf("Unsupported applicable account type %q.", v))
		}
	}
	result.ApplicableAccountType, diags = types.SetValueFrom(ctx, types.StringType, value.AccountTypes)
	d.Append(diags...)
	return result, d
}

func permissionState(values []PermissionRecord) (types.Set, diag.Diagnostics) {
	var d diag.Diagnostics
	permissions := make([]attr.Value, 0, len(values))
	seen := map[string]bool{}
	for _, p := range values {
		if p.Resource == "" || (p.Action != "NONE" && p.Action != "VIEW" && p.Action != "EDIT") || seen[p.Resource] {
			d.AddError("Invalid role response", "The API returned invalid or duplicate permissions.")
			continue
		}
		seen[p.Resource] = true
		obj,
			diags := types.ObjectValue(PermissionAttrTypes,
			map[string]attr.Value{"resource": types.StringValue(p.Resource),
				"action": types.StringValue(p.Action)})
		d.Append(diags...)
		permissions = append(permissions, obj)
	}

	result, diags := types.SetValue(PermissionObjectType, permissions)
	d.Append(diags...)

	return result, d
}
