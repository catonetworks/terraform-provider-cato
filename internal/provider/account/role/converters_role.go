package role

import (
	"context"
	"fmt"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

func createInput(ctx context.Context, plan Model) (m.RoleManagementCreateRoleInput, diag.Diagnostics) {
	permissions, d := iam.PermissionsInput(ctx, plan.Permissions)
	if plan.Name.IsNull() || plan.Name.IsUnknown() || plan.Name.ValueString() == "" || plan.Description.IsUnknown() {
		d.AddError("Invalid role input", "Name and description must be known; name cannot be empty.")
	}
	description := plan.Description.ValueString()
	return m.RoleManagementCreateRoleInput{Name: plan.Name.ValueString(), Description: &description, Permission: permissions}, d
}

func mutableRole(value *iam.RoleRecord) error {
	if value == nil {
		return fmt.Errorf("role no longer exists")
	}
	if value.Predefined || value.External {
		return fmt.Errorf("predefined roles and roles used by external access cannot be managed")
	}
	return nil
}

func sameMutableRole(actual, expected *iam.RoleRecord) bool {
	if actual == nil || actual.ID != expected.ID || actual.Name != expected.Name || actual.Permissions == nil {
		return false
	}
	description := func(value *string) string {
		if value == nil {
			return ""
		}
		return *value
	}
	if description(actual.Description) != description(expected.Description) || len(actual.Permissions) != len(expected.Permissions) {
		return false
	}
	permissions := make(map[string]string, len(expected.Permissions))
	for _, permission := range expected.Permissions {
		permissions[permission.Resource] = permission.Action
	}
	for _, permission := range actual.Permissions {
		if action, ok := permissions[permission.Resource]; !ok || action != permission.Action {
			return false
		}
		delete(permissions, permission.Resource)
	}
	return len(permissions) == 0
}
