package rbacassignment

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model represents one user or group assignment.
type Model struct {
	ID                           types.String `tfsdk:"id"`
	AccountID                    types.String `tfsdk:"account_id"`
	AssignmentID                 types.String `tfsdk:"assignment_id"`
	UserID                       types.String `tfsdk:"user_id"`
	GroupID                      types.String `tfsdk:"group_id"`
	RoleIDs                      types.Set    `tfsdk:"role_ids"`
	AllowedEntities              types.Set    `tfsdk:"allowed_entities"`
	Description                  types.String `tfsdk:"description"`
	AllowViewingUnobfuscatedData types.Bool   `tfsdk:"allow_viewing_unobfuscated_data"`
	AllowAllSites                types.String `tfsdk:"allow_all_sites"`
	AllowAllUsers                types.String `tfsdk:"allow_all_users"`
	AssignmentType               types.String `tfsdk:"assignment_type"`
	DataObfuscationStatus        types.String `tfsdk:"data_obfuscation_status"`
}

var entityAttrTypes = map[string]attr.Type{"id": types.StringType, "type": types.StringType, "action": types.StringType}
var entityObjectType = types.ObjectType{AttrTypes: entityAttrTypes}

type entityRecord struct{ ID, Type, Action string }
type assignmentRecord struct {
	ID, Account, User, Group, Kind, AllSites, AllUsers, Obfuscation string
	Description                                                     *string
	Roles                                                           []string
	Entities                                                        []entityRecord
}
