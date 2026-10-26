package rbacassignment

import (
	"context"
	"fmt"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createInput(ctx context.Context, plan Model) (m.AddRbacAssignmentsInput, diag.Diagnostics) {
	var d diag.Diagnostics
	description := plan.Description.ValueString()
	result := m.AddRbacAssignmentsInput{Users: []*m.UserRefInput{},
		Groups:          []*m.UsersGroupRefInput{},
		Roles:           []*m.RBACRoleRefInput{},
		AllowedEntities: []*m.AuthorizedEntityRefInput{},
		AllowAllSites:   m.RBACActionNone,
		AllowAllUsers:   m.RBACActionNone,
		Description:     &description}
	d.Append(principalInput(plan, &result)...)
	if plan.Description.IsUnknown() || plan.AllowViewingUnobfuscatedData.IsUnknown() {
		d.AddError("Invalid assignment settings", "Description and obfuscation preference must be known.")
	}
	if !plan.AllowViewingUnobfuscatedData.IsNull() {
		v := plan.AllowViewingUnobfuscatedData.ValueBool()
		result.AllowViewingUnobfuscatedData = &v
	}
	roles, roleDiags := rolesInput(ctx, plan.RoleIDs)
	d.Append(roleDiags...)
	result.Roles = roles
	entities, diags := entitiesInput(plan.AllowedEntities)
	d.Append(diags...)
	result.AllowedEntities = entities
	return result, d
}

func entitiesInput(value types.Set) ([]*m.AuthorizedEntityRefInput, diag.Diagnostics) {
	var d diag.Diagnostics
	result := make([]*m.AuthorizedEntityRefInput, 0)
	seen := map[string]bool{}
	if value.IsNull() || value.IsUnknown() {
		d.AddError("Invalid entity scope", "Configure a known allowed_entities set; use [] for empty scope.")
		return result, d
	}
	for _, entry := range value.Elements() {
		object, ok := entry.(types.Object)
		if !ok || object.IsNull() || object.IsUnknown() {
			d.AddError("Invalid entity", "Each entity must be a known object.")
			continue
		}
		a := object.Attributes()
		id := a["id"].(types.String)
		kind := a["type"].(types.String)
		action := a["action"].(types.String)
		key := kind.ValueString() + ":" + id.ValueString()
		if id.IsNull() || id.IsUnknown() || id.ValueString() == "" || !entityTypeValid(kind.ValueString()) ||
			!grantValid(action.ValueString()) || seen[key] {
			d.AddError("Invalid entity", "Each type/ID pair must occur once with a VIEW or EDIT action.")
			continue
		}
		seen[key] = true
		result = append(result,
			&m.AuthorizedEntityRefInput{Type: m.AuthorizedEntityType(kind.ValueString()),
				Action: m.RBACAction(action.ValueString()),
				Ref: &m.ObjectRefInput{By: m.ObjectRefByID,
					Input: id.ValueString()}})
	}
	return result, d
}

func grantValid(value string) bool { return value == "VIEW" || value == "EDIT" }
func entityTypeValid(value string) bool {
	switch value {
	case "SITE", "USERSGROUP", "GROUP", "SYSTEMGROUP", "WAN_FIREWALL_POLICY", "INTERNET_FIREWALL_POLICY":
		return true
	default:
		return false
	}
}
func validateRecord(account string, value *assignmentRecord) error {
	if value == nil || value.ID == "" || value.Account != account || value.Roles == nil || value.Entities == nil {
		return fmt.Errorf("API returned an incomplete or cross-account assignment")
	}
	if err := validatePrincipal(value); err != nil {
		return err
	}
	if !effectiveGrantValid(value.AllSites) || !effectiveGrantValid(value.AllUsers) {
		return fmt.Errorf("API returned an invalid effective scope action")
	}
	return nil
}

func assignmentState(ctx context.Context, account string, value *assignmentRecord, previous *Model) (Model, diag.Diagnostics) {
	var result Model
	var d diag.Diagnostics
	if err := validateRecord(account, value); err != nil {
		d.AddError("Invalid assignment response", err.Error())
		return result, d
	}
	result.ID = types.StringValue(account + ":" + value.ID)
	result.AccountID = types.StringValue(account)
	result.AssignmentID = types.StringValue(value.ID)
	result.UserID = types.StringNull()
	result.GroupID = types.StringNull()
	if value.User != "" {
		result.UserID = types.StringValue(value.User)
	}
	if value.Group != "" {
		result.GroupID = types.StringValue(value.Group)
	}
	if previous != nil && (!previous.UserID.IsNull() || !previous.GroupID.IsNull()) &&
		(previous.UserID.ValueString() != value.User || previous.GroupID.ValueString() != value.Group) {
		d.AddError("Invalid assignment response", "API returned a different principal.")
	}
	result.Description = types.StringValue("")
	if value.Description != nil {
		result.Description = types.StringValue(*value.Description)
	}
	result.AssignmentType = types.StringValue(value.Kind)
	result.AllowAllSites = types.StringValue(value.AllSites)
	result.AllowAllUsers = types.StringValue(value.AllUsers)
	result.DataObfuscationStatus = types.StringValue(value.Obfuscation)
	preference, preferenceDiags := obfuscationState(value.Obfuscation, previous)
	result.AllowViewingUnobfuscatedData = preference
	d.Append(preferenceDiags...)
	for _, id := range value.Roles {
		if id == "" {
			d.AddError("Invalid assignment response", "API returned an empty role ID.")
		}
	}
	var diags diag.Diagnostics
	result.RoleIDs, diags = types.SetValueFrom(ctx, types.StringType, value.Roles)
	d.Append(diags...)
	result.AllowedEntities, diags = entitiesState(value.Entities)
	d.Append(diags...)
	return result, d
}

func principalInput(plan Model, result *m.AddRbacAssignmentsInput) diag.Diagnostics {
	var d diag.Diagnostics
	user, group := plan.UserID.ValueString(), plan.GroupID.ValueString()
	if plan.UserID.IsUnknown() || plan.GroupID.IsUnknown() || (user == "") == (group == "") {
		d.AddError("Invalid assignment principal", "Specify exactly one known user_id or group_id.")
	}
	if user != "" {
		result.Users = append(result.Users, &m.UserRefInput{By: m.ObjectRefByID, Input: user})
	}
	if group != "" {
		result.Groups = append(result.Groups, &m.UsersGroupRefInput{By: m.ObjectRefByID, Input: group})
	}

	return d
}

func rolesInput(ctx context.Context, value types.Set) ([]*m.RBACRoleRefInput, diag.Diagnostics) {
	var d diag.Diagnostics
	result := make([]*m.RBACRoleRefInput, 0)
	if value.IsNull() || value.IsUnknown() || len(value.Elements()) == 0 {
		d.AddError("Invalid assignment roles", "Configure a nonempty set of known role IDs.")
	}
	var roles []types.String
	d.Append(value.ElementsAs(ctx, &roles, false)...)
	for _, role := range roles {
		if role.IsNull() || role.IsUnknown() || role.ValueString() == "" {
			d.AddError("Invalid role ID", "Role IDs must be known and nonempty.")
			continue
		}
		result = append(result, &m.RBACRoleRefInput{By: m.ObjectRefByID, Input: role.ValueString()})
	}

	return result, d
}

func validatePrincipal(value *assignmentRecord) error {
	if (value.Kind != "ADMIN_USER" || value.User == "" || value.Group != "") &&
		(value.Kind != "GROUP" || value.Group == "" || value.User != "") {
		return fmt.Errorf("only ADMIN_USER and GROUP assignments with unambiguous principals can be managed")
	}

	return nil
}

func effectiveGrantValid(value string) bool { return value == "NONE" || grantValid(value) }

func obfuscationState(status string, previous *Model) (types.Bool, diag.Diagnostics) {
	var result types.Bool
	var d diag.Diagnostics
	result = types.BoolNull()
	switch status {
	case "DISABLED":
		if previous != nil {
			result = previous.AllowViewingUnobfuscatedData
		}
	case "OBFUSCATED", "UNOBFUSCATED":
		if previous != nil && !previous.AllowViewingUnobfuscatedData.IsNull() {
			result = types.BoolValue(status == "UNOBFUSCATED")
		}
	default:
		d.AddError("Invalid assignment response", "API returned an unknown obfuscation status.")
	}

	return result, d
}

func entitiesState(values []entityRecord) (types.Set, diag.Diagnostics) {
	var d diag.Diagnostics
	entities := make([]attr.Value, 0, len(values))
	seen := map[string]bool{}
	for _, v := range values {
		key := v.Type + ":" + v.ID
		if v.ID == "" || !entityTypeValid(v.Type) || !grantValid(v.Action) || seen[key] {
			d.AddError("Invalid assignment response", "API returned an invalid or duplicate entity.")
			continue
		}
		seen[key] = true
		object,
			diags := types.ObjectValue(entityAttrTypes,
			map[string]attr.Value{"id": types.StringValue(v.ID),
				"type":   types.StringValue(v.Type),
				"action": types.StringValue(v.Action)})
		d.Append(diags...)
		entities = append(entities, object)
	}
	result, diags := types.SetValue(entityObjectType, entities)
	d.Append(diags...)

	return result, d
}
