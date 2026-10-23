package rbacassignment

import sdk "github.com/catonetworks/cato-go-sdk"

func createdRecord(value *sdk.RbacAddRbacAssignments_Rbac_AddRbacAssignments_Items) *assignmentRecord {
	if value == nil {
		return nil
	}
	result := &assignmentRecord{ID: value.ID,
		Account:     value.Account.GetID(),
		User:        value.User.GetID(),
		Group:       value.Group.GetID(),
		Kind:        string(value.AssignmentType),
		AllSites:    string(value.AllowAllSites),
		AllUsers:    string(value.AllowAllUsers),
		Obfuscation: string(value.DataObfuscationStatus),
		Description: value.Description}
	if value.Roles != nil {
		result.Roles = make([]string, 0, len(value.Roles))
		for _, v := range value.Roles {
			if v == nil {
				return nil
			}
			result.Roles = append(result.Roles, v.ID)
		}
	}
	if value.AllowedEntities != nil {
		result.Entities = make([]entityRecord, 0, len(value.AllowedEntities))
		for _, v := range value.AllowedEntities {
			if v == nil {
				return nil
			}
			result.Entities = append(result.Entities, entityRecord{ID: v.ID, Type: string(v.Type), Action: string(v.Action)})
		}
	}
	return result
}
func updatedRecord(value *sdk.RbacUpdateRbacAssignment_Rbac_UpdateRbacAssignment_Assignment) *assignmentRecord {
	if value == nil {
		return nil
	}
	result := &assignmentRecord{ID: value.ID,
		Account:     value.Account.GetID(),
		User:        value.User.GetID(),
		Group:       value.Group.GetID(),
		Kind:        string(value.AssignmentType),
		AllSites:    string(value.AllowAllSites),
		AllUsers:    string(value.AllowAllUsers),
		Obfuscation: string(value.DataObfuscationStatus),
		Description: value.Description}
	if value.Roles != nil {
		result.Roles = make([]string, 0, len(value.Roles))
		for _, v := range value.Roles {
			if v == nil {
				return nil
			}
			result.Roles = append(result.Roles, v.ID)
		}
	}
	if value.AllowedEntities != nil {
		result.Entities = make([]entityRecord, 0, len(value.AllowedEntities))
		for _, v := range value.AllowedEntities {
			if v == nil {
				return nil
			}
			result.Entities = append(result.Entities, entityRecord{ID: v.ID, Type: string(v.Type), Action: string(v.Action)})
		}
	}
	return result
}
func listedRecord(value *sdk.RbacRbacAssignmentList_Rbac_RbacAssignmentList_Items) *assignmentRecord {
	if value == nil {
		return nil
	}
	result := &assignmentRecord{ID: value.ID,
		Account:     value.Account.GetID(),
		User:        value.User.GetID(),
		Group:       value.Group.GetID(),
		Kind:        string(value.AssignmentType),
		AllSites:    string(value.AllowAllSites),
		AllUsers:    string(value.AllowAllUsers),
		Obfuscation: string(value.DataObfuscationStatus),
		Description: value.Description}
	if value.Roles != nil {
		result.Roles = make([]string, 0, len(value.Roles))
		for _, v := range value.Roles {
			if v == nil {
				return nil
			}
			result.Roles = append(result.Roles, v.ID)
		}
	}
	if value.AllowedEntities != nil {
		result.Entities = make([]entityRecord, 0, len(value.AllowedEntities))
		for _, v := range value.AllowedEntities {
			if v == nil {
				return nil
			}
			result.Entities = append(result.Entities, entityRecord{ID: v.ID, Type: string(v.Type), Action: string(v.Action)})
		}
	}
	return result
}
