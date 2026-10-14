package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	rbacapi "github.com/catonetworks/terraform-provider-cato/internal/provider/rbacassignmentapi"
)

type RBACAssignment struct {
	ID                           types.String `tfsdk:"id"`
	AssignmentID                 types.String `tfsdk:"assignment_id"`
	AccountID                    types.String `tfsdk:"account_id"`
	UserID                       types.String `tfsdk:"user_id"`
	GroupID                      types.String `tfsdk:"group_id"`
	RoleIDs                      types.Set    `tfsdk:"role_ids"`
	AllowedEntities              types.Set    `tfsdk:"allowed_entities"`
	AllowAllSites                types.String `tfsdk:"allow_all_sites"`
	AllowAllUsers                types.String `tfsdk:"allow_all_users"`
	AllowViewingUnobfuscatedData types.Bool   `tfsdk:"allow_viewing_unobfuscated_data"`
	Description                  types.String `tfsdk:"description"`
	AssignmentType               types.String `tfsdk:"assignment_type"`
	DataObfuscationStatus        types.String `tfsdk:"data_obfuscation_status"`
}

var assignmentEntityAttrTypes = map[string]attr.Type{"id": types.StringType, "type": types.StringType, "action": types.StringType}
var assignmentEntityObjectType = types.ObjectType{AttrTypes: assignmentEntityAttrTypes}

func assignmentActionValid(action string) bool {
	return action == "NONE" || action == "VIEW" || action == "EDIT"
}
func assignmentEntityTypeValid(entityType string) bool {
	switch entityType {
	case "SITE", "USERSGROUP", "GROUP", "SYSTEMGROUP", "WAN_FIREWALL_POLICY", "INTERNET_FIREWALL_POLICY":
		return true
	default:
		return false
	}
}

func assignmentInput(ctx context.Context, model RBACAssignment, create bool) (rbacapi.RBACAssignmentInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	input := rbacapi.RBACAssignmentInput{
		Users: []rbacapi.RBACReference{}, Groups: []rbacapi.RBACReference{},
		// The API derives the effective all-sites/users grants from SYSTEMGROUP
		// entries. NONE avoids injecting a second representation of the same scope.
		AllowAllSites: "NONE", AllowAllUsers: "NONE",
		Description: model.Description.ValueString(),
	}
	if model.Description.IsUnknown() || model.AllowViewingUnobfuscatedData.IsUnknown() {
		diags.AddError("Invalid assignment scope", "Scope actions, description and obfuscation settings must be known.")
	}
	if !model.AllowViewingUnobfuscatedData.IsNull() {
		v := model.AllowViewingUnobfuscatedData.ValueBool()
		input.AllowViewingUnobfuscatedData = &v
	}
	if create {
		diags.Append(assignmentPrincipalInput(model, &input)...)
	} else {
		input.AssignmentID = model.AssignmentID.ValueString()
		if input.AssignmentID == "" {
			diags.AddError("Invalid assignment identifier", "assignment_id must be known and nonempty.")
		}
	}
	var d diag.Diagnostics
	input.Roles, d = assignmentRolesInput(ctx, model.RoleIDs)
	diags.Append(d...)
	input.AllowedEntities, d = assignmentEntitiesInput(model.AllowedEntities)
	diags.Append(d...)
	return input, diags
}

func assignmentPrincipalInput(model RBACAssignment, input *rbacapi.RBACAssignmentInput) diag.Diagnostics {
	var diags diag.Diagnostics
	user, group := model.UserID.ValueString(), model.GroupID.ValueString()
	if model.UserID.IsUnknown() || model.GroupID.IsUnknown() || (user == "") == (group == "") {
		diags.AddError("Invalid assignment principal", "Specify exactly one known, nonempty user_id or group_id.")
		return diags
	}
	if user != "" {
		input.Users = append(input.Users, rbacapi.RBACReference{By: "ID", Input: user})
	}
	if group != "" {
		input.Groups = append(input.Groups, rbacapi.RBACReference{By: "ID", Input: group})
	}
	return diags
}

func assignmentRolesInput(ctx context.Context, roles types.Set) ([]rbacapi.RBACReference, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := []rbacapi.RBACReference{}
	if roles.IsNull() || roles.IsUnknown() || len(roles.Elements()) == 0 {
		diags.AddError("Invalid assignment roles", "Specify a nonempty set of known role IDs.")
		return result, diags
	}
	var ids []types.String
	diags.Append(roles.ElementsAs(ctx, &ids, false)...)
	for _, id := range ids {
		if id.IsNull() || id.IsUnknown() || id.ValueString() == "" {
			diags.AddError("Invalid assignment role", "Role IDs must be known and nonempty.")
			continue
		}
		result = append(result, rbacapi.RBACReference{By: "ID", Input: id.ValueString()})
	}
	return result, diags
}

func assignmentEntitiesInput(entities types.Set) ([]rbacapi.RBACEntityInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := []rbacapi.RBACEntityInput{}
	if entities.IsNull() || entities.IsUnknown() {
		diags.AddError("Invalid assignment entities", "allowed_entities must be known; use [] for no individual entities.")
		return result, diags
	}
	seen := map[string]bool{}
	for _, v := range entities.Elements() {
		entity, err := assignmentEntityInput(v)
		if err != nil {
			diags.AddError("Invalid assignment entity", err.Error())
			continue
		}
		key := entity.Type + ":" + entity.Ref.Input
		if seen[key] {
			diags.AddError("Duplicate assignment entity", "Specify each entity once with its intended action.")
			continue
		}
		seen[key] = true
		result = append(result, entity)
	}
	return result, diags
}

func assignmentEntityInput(v attr.Value) (rbacapi.RBACEntityInput, error) {
	object, ok := v.(types.Object)
	if !ok || object.IsNull() || object.IsUnknown() {
		return rbacapi.RBACEntityInput{}, fmt.Errorf("each entity must be known")
	}
	attrs := object.Attributes()
	id := attrs["id"].(types.String)
	entityType := attrs["type"].(types.String).ValueString()
	action := attrs["action"].(types.String).ValueString()
	if id.IsUnknown() || id.IsNull() || id.ValueString() == "" ||
		!assignmentEntityTypeValid(entityType) || (action != "VIEW" && action != "EDIT") {
		return rbacapi.RBACEntityInput{}, fmt.Errorf("each entity needs an ID, a supported type, and VIEW or EDIT action")
	}
	return rbacapi.RBACEntityInput{
		Type: entityType, Ref: rbacapi.RBACReference{By: "ID", Input: id.ValueString()}, Action: action,
	}, nil
}

func assignmentFromAPI(account string, value *rbacapi.RBACAssignmentAPI, previous *RBACAssignment) (RBACAssignment, diag.Diagnostics) {
	model := RBACAssignment{
		UserID: types.StringNull(), GroupID: types.StringNull(), Description: types.StringValue(""),
		AllowViewingUnobfuscatedData: types.BoolNull(),
	}
	var diags diag.Diagnostics
	if err := validateAssignmentAPI(account, value); err != nil {
		diags.AddError("Invalid assignment response", err.Error())
		return model, diags
	}
	model.ID = types.StringValue(account + ":" + value.ID)
	model.AssignmentID = types.StringValue(value.ID)
	model.AccountID = types.StringValue(account)
	if value.User != nil {
		model.UserID = types.StringValue(value.User.ID)
	}
	if value.Group != nil {
		model.GroupID = types.StringValue(value.Group.ID)
	}
	if previous != nil && (!previous.UserID.IsNull() || !previous.GroupID.IsNull()) &&
		(previous.UserID.ValueString() != model.UserID.ValueString() || previous.GroupID.ValueString() != model.GroupID.ValueString()) {
		diags.AddError("Invalid assignment response", "The API returned an assignment for a different principal.")
	}
	model.AssignmentType = types.StringValue(value.AssignmentType)
	model.AllowAllSites = types.StringValue(value.AllowAllSites)
	model.AllowAllUsers = types.StringValue(value.AllowAllUsers)
	model.DataObfuscationStatus = types.StringValue(value.DataObfuscationStatus)
	diags.Append(assignmentObfuscationState(&model, value.DataObfuscationStatus, previous)...)
	if value.Description != nil {
		model.Description = types.StringValue(*value.Description)
	}
	var d diag.Diagnostics
	model.RoleIDs, d = assignmentRolesState(value.Roles)
	diags.Append(d...)
	model.AllowedEntities, d = assignmentEntitiesState(value.AllowedEntities)
	diags.Append(d...)
	return model, diags
}

func validateAssignmentAPI(account string, value *rbacapi.RBACAssignmentAPI) error {
	if value == nil || value.ID == "" || value.Account == nil || value.Account.ID != account {
		return fmt.Errorf("API returned an incomplete or cross-account assignment")
	}
	if err := validateAssignmentPrincipal(value); err != nil {
		return err
	}

	if !assignmentActionValid(value.AllowAllSites) || !assignmentActionValid(value.AllowAllUsers) ||
		value.Roles == nil || value.AllowedEntities == nil {
		return fmt.Errorf("API returned incomplete roles or scope")
	}
	return nil
}

func assignmentObfuscationState(model *RBACAssignment, status string, previous *RBACAssignment) diag.Diagnostics {
	var diags diag.Diagnostics
	switch status {
	case "OBFUSCATED", "UNOBFUSCATED":
		if previous != nil && !previous.AllowViewingUnobfuscatedData.IsNull() {
			model.AllowViewingUnobfuscatedData = types.BoolValue(status == "UNOBFUSCATED")
		}
	case "DISABLED":
		// Effective status cannot reveal the stored preference while account
		// obfuscation is disabled. Preserve configuration rather than infer a grant.
		if previous != nil {
			model.AllowViewingUnobfuscatedData = previous.AllowViewingUnobfuscatedData
		}
	default:
		diags.AddError("Invalid assignment response", "The API returned an unknown data obfuscation status.")
	}
	return diags
}

func assignmentRolesState(values []*rbacapi.RBACObject) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	roles := []attr.Value{}
	for _, v := range values {
		if v == nil || v.ID == "" {
			diags.AddError("Invalid assignment response", "The API returned an incomplete role.")
			continue
		}
		roles = append(roles, types.StringValue(v.ID))
	}
	result, d := types.SetValue(types.StringType, roles)
	diags.Append(d...)
	return result, diags
}

func assignmentEntitiesState(values []*rbacapi.RBACEntity) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	entities := []attr.Value{}
	for _, v := range values {
		if v == nil || v.ID == "" || !assignmentEntityTypeValid(v.Type) || (v.Action != "VIEW" && v.Action != "EDIT") {
			diags.AddError("Invalid assignment response", "The API returned an unsupported allowed entity.")
			continue
		}
		object, d := types.ObjectValue(assignmentEntityAttrTypes, map[string]attr.Value{
			"id": types.StringValue(v.ID), "type": types.StringValue(v.Type), "action": types.StringValue(v.Action),
		})
		diags.Append(d...)
		entities = append(entities, object)
	}
	result, d := types.SetValue(assignmentEntityObjectType, entities)
	diags.Append(d...)
	return result, diags
}

func validateAssignmentPrincipal(value *rbacapi.RBACAssignmentAPI) error {
	switch value.AssignmentType {
	case "ADMIN_USER":
		if value.User == nil || value.User.ID == "" || value.Group != nil {
			return fmt.Errorf("user assignment has no unambiguous user reference")
		}
	case "GROUP":
		if value.Group == nil || value.Group.ID == "" || value.User != nil {
			return fmt.Errorf("group assignment has no unambiguous group reference")
		}
	default:
		return fmt.Errorf("legacy ADMIN assignments cannot be managed; only ADMIN_USER and GROUP assignments support deletion")
	}
	return nil
}
