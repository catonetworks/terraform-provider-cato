package iam

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Role is the Terraform representation shared by role resources and queries.
type Role struct {
	ID                     types.String `tfsdk:"id"`
	AccountID              types.String `tfsdk:"account_id"`
	RoleID                 types.String `tfsdk:"role_id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	Permissions            types.Set    `tfsdk:"permissions"`
	Predefined             types.Bool   `tfsdk:"predefined"`
	ApplicableAccountType  types.Set    `tfsdk:"applicable_account_type"`
	IsUsedOnExternalAccess types.Bool   `tfsdk:"is_used_on_external_access"`
}

// PermissionAttrTypes defines the nested permission representation.
var PermissionAttrTypes = map[string]attr.Type{"resource": types.StringType, "action": types.StringType}

// PermissionObjectType represents a permission in Terraform collections.
var PermissionObjectType = types.ObjectType{AttrTypes: PermissionAttrTypes}

// RoleAttrTypes defines role list entries.
var RoleAttrTypes = map[string]attr.Type{
	"id": types.StringType, "account_id": types.StringType, "role_id": types.StringType, "name": types.StringType,
	"description": types.StringType, "permissions": types.SetType{ElemType: PermissionObjectType},
	"predefined": types.BoolType, "applicable_account_type": types.SetType{ElemType: types.StringType},
	"is_used_on_external_access": types.BoolType,
}

// RoleObjectType represents a complete role in a result list.
var RoleObjectType = types.ObjectType{AttrTypes: RoleAttrTypes}

// RoleRecord is a normalized API role projection used by the shared converters.
type RoleRecord struct {
	ID, Name             string
	Description          *string
	Predefined, External bool
	AccountTypes         []string
	Permissions          []PermissionRecord
}

// PermissionRecord is an API resource/action pair.
type PermissionRecord struct{ Resource, Action string }
