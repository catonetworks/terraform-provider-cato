package iam

import sdk "github.com/catonetworks/cato-go-sdk"

// CreateRoleRecord converts the selected SDK role fields without discarding absent collections.
func CreateRoleRecord(value *sdk.RbacRoleManagementCreateRole_Rbac_RoleManagement_CreateRole_Role) *RoleRecord {
	if value == nil {
		return nil
	}
	result := &RoleRecord{ID: value.ID,
		Name:        value.Name,
		Description: value.Description,
		Predefined:  value.Predefined,
		External:    value.IsUsedOnExternalAccess}
	if value.ApplicableAccountType != nil {
		result.AccountTypes = make([]string, 0, len(value.ApplicableAccountType))
		for _, v := range value.ApplicableAccountType {
			result.AccountTypes = append(result.AccountTypes, string(v))
		}
	}
	if value.Permission != nil {
		result.Permissions = make([]PermissionRecord, 0, len(value.Permission))
		for _, v := range value.Permission {
			if v == nil {
				return nil
			}
			result.Permissions = append(result.Permissions, PermissionRecord{Resource: v.Resource, Action: string(v.Action)})
		}
	}
	return result
}

// UpdateRoleRecord converts the selected SDK role fields without discarding absent collections.
func UpdateRoleRecord(value *sdk.RbacRoleManagementUpdateRole_Rbac_RoleManagement_UpdateRole_Role) *RoleRecord {
	if value == nil {
		return nil
	}
	result := &RoleRecord{ID: value.ID,
		Name:        value.Name,
		Description: value.Description,
		Predefined:  value.Predefined,
		External:    value.IsUsedOnExternalAccess}
	if value.ApplicableAccountType != nil {
		result.AccountTypes = make([]string, 0, len(value.ApplicableAccountType))
		for _, v := range value.ApplicableAccountType {
			result.AccountTypes = append(result.AccountTypes, string(v))
		}
	}
	if value.Permission != nil {
		result.Permissions = make([]PermissionRecord, 0, len(value.Permission))
		for _, v := range value.Permission {
			if v == nil {
				return nil
			}
			result.Permissions = append(result.Permissions, PermissionRecord{Resource: v.Resource, Action: string(v.Action)})
		}
	}
	return result
}

// ReadRoleRecord converts the selected SDK role fields without discarding absent collections.
func ReadRoleRecord(value *sdk.RbacRoleManagementReadRole_Rbac_RoleManagement_Role) *RoleRecord {
	if value == nil {
		return nil
	}
	result := &RoleRecord{ID: value.ID,
		Name:        value.Name,
		Description: value.Description,
		Predefined:  value.Predefined,
		External:    value.IsUsedOnExternalAccess}
	if value.ApplicableAccountType != nil {
		result.AccountTypes = make([]string, 0, len(value.ApplicableAccountType))
		for _, v := range value.ApplicableAccountType {
			result.AccountTypes = append(result.AccountTypes, string(v))
		}
	}
	if value.Permission != nil {
		result.Permissions = make([]PermissionRecord, 0, len(value.Permission))
		for _, v := range value.Permission {
			if v == nil {
				return nil
			}
			result.Permissions = append(result.Permissions, PermissionRecord{Resource: v.Resource, Action: string(v.Action)})
		}
	}
	return result
}

// ListRoleRecord converts the selected SDK role fields without discarding absent collections.
func ListRoleRecord(value *sdk.RbacRoleManagementListRoles_Rbac_RoleManagement_RoleList_Items) *RoleRecord {
	if value == nil {
		return nil
	}
	result := &RoleRecord{ID: value.ID,
		Name:        value.Name,
		Description: value.Description,
		Predefined:  value.Predefined,
		External:    value.IsUsedOnExternalAccess}
	if value.ApplicableAccountType != nil {
		result.AccountTypes = make([]string, 0, len(value.ApplicableAccountType))
		for _, v := range value.ApplicableAccountType {
			result.AccountTypes = append(result.AccountTypes, string(v))
		}
	}
	if value.Permission != nil {
		result.Permissions = make([]PermissionRecord, 0, len(value.Permission))
		for _, v := range value.Permission {
			if v == nil {
				return nil
			}
			result.Permissions = append(result.Permissions, PermissionRecord{Resource: v.Resource, Action: string(v.Action)})
		}
	}
	return result
}
