package groupmembers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ParseGroupMemberListWithConfig converts API group member items to a Terraform set
// matching with config/plan to preserve user's choice of specifying ID vs name
//
//nolint:gocyclo
func ParseGroupMemberListWithConfig[T any](ctx context.Context, items []T, configMembers types.Set, attrName string) types.Set {
	tflog.Debug(ctx, "parseGroupMemberListWithConfig() "+attrName)

	// Handle empty API list
	if len(items) == 0 {
		tflog.Debug(ctx, "parseGroupMemberListWithConfig() - empty API list")
		return types.SetValueMust(types.ObjectType{AttrTypes: MemberAttrTypes}, []attr.Value{})
	}

	// Parse config members to create a matching map
	// Key: "type:id:value" or "type:name:value"
	var configMembersList []Member
	configMap := make(map[string]Member)

	if !configMembers.IsNull() && !configMembers.IsUnknown() {
		if diags := configMembers.ElementsAs(ctx, &configMembersList, false); !diags.HasError() {
			for _, configMember := range configMembersList {
				memberType := configMember.Type.ValueString()
				// Create lookup keys based on what was specified in config
				if !configMember.ID.IsNull() && !configMember.ID.IsUnknown() && configMember.ID.ValueString() != "" {
					key := memberType + ":id:" + configMember.ID.ValueString()
					configMap[key] = configMember
				}
				if !configMember.Name.IsNull() && !configMember.Name.IsUnknown() && configMember.Name.ValueString() != "" {
					key := memberType + ":name:" + configMember.Name.ValueString()
					configMap[key] = configMember
				}
			}
		}
	}

	// Process API items and match with config
	memberValues := make([]attr.Value, 0, len(items))
	for _, item := range items {
		apiMemberObj := ParseGroupMember(ctx, item, attrName)
		if apiMemberObj.IsNull() {
			continue
		}

		// Extract API member fields
		apiAttrs := apiMemberObj.Attributes()
		apiType := ""
		apiID := ""
		apiName := ""

		if typeVal, ok := apiAttrs["type"].(types.String); ok && !typeVal.IsNull() {
			apiType = typeVal.ValueString()
		}
		if idVal, ok := apiAttrs["id"].(types.String); ok && !idVal.IsNull() {
			apiID = idVal.ValueString()
		}
		if nameVal, ok := apiAttrs["name"].(types.String); ok && !nameVal.IsNull() {
			apiName = nameVal.ValueString()
		}

		// Try to find matching config member.
		matchedConfig := false
		if apiID != "" && apiType != "" {
			key := apiType + ":id:" + apiID
			if _, exists := configMap[key]; exists {
				matchedConfig = true
			}
		}
		if !matchedConfig && apiName != "" && apiType != "" {
			key := apiType + ":name:" + apiName
			if _, exists := configMap[key]; exists {
				matchedConfig = true
			}
		}

		// Build the result member object
		// - Always write both ID and name to state (from API)
		// - This allows Terraform to track changes properly
		resultAttrs := map[string]attr.Value{
			"id":   apiAttrs["id"],
			"name": apiAttrs["name"],
			"type": apiAttrs["type"],
		}

		resultObj, _ := types.ObjectValue(MemberAttrTypes, resultAttrs)
		memberValues = append(memberValues, resultObj)
	}

	// Convert to types.Set
	membersSet, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: MemberAttrTypes}, memberValues)
	return membersSet
}

// ParseGroupMemberList converts API group member items to a Terraform set
// Similar to oldidname.ParseNameIDList but includes the type field
func ParseGroupMemberList[T any](ctx context.Context, items []T, attrName string) types.Set {
	tflog.Debug(ctx, "parseGroupMemberList() "+attrName)

	// Handle empty list - return empty set (not null) to match schema default
	if len(items) == 0 {
		tflog.Debug(ctx, "parseGroupMemberList() - empty input list, returning empty set")
		return types.SetValueMust(types.ObjectType{AttrTypes: MemberAttrTypes}, []attr.Value{})
	}

	// Process each item into an attr.Value
	memberValues := make([]attr.Value, 0, len(items))
	for _, item := range items {
		memberObj := ParseGroupMember(ctx, item, attrName)
		if !memberObj.IsNull() {
			memberValues = append(memberValues, memberObj)
		}
	}

	// Convert to types.Set
	membersSet, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: MemberAttrTypes}, memberValues)
	return membersSet
}

// ParseGroupMember converts a single API group member item to a Terraform object
//
//nolint:gocyclo
func ParseGroupMember(_ context.Context, item interface{}, _ string) types.Object {
	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(MemberAttrTypes)
			}
		} else {
			return types.ObjectNull(MemberAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Pointer {
		if itemValue.IsNil() {
			return types.ObjectNull(MemberAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name, ID, and Type fields
	nameField := itemValue.FieldByName("Name")
	idField := itemValue.FieldByName("ID")
	typeField := itemValue.FieldByName("TypeGroupMemberRefTyped")

	if !nameField.IsValid() || !idField.IsValid() || !typeField.IsValid() {
		return types.ObjectNull(MemberAttrTypes)
	}

	// Extract string values
	var nameValue, idValue, typeValue types.String

	// Handle ID field
	if idField.Kind() == reflect.Pointer {
		if idField.IsNil() || idField.Elem().String() == "" {
			idValue = types.StringNull()
		} else {
			idValue = types.StringValue(idField.Elem().String())
		}
	} else {
		val := idField.String()
		if val == "" {
			idValue = types.StringNull()
		} else {
			idValue = types.StringValue(val)
		}
	}

	// Handle name field
	if nameField.Kind() == reflect.Pointer {
		if nameField.IsNil() || nameField.Elem().String() == "" {
			nameValue = types.StringNull()
		} else {
			nameValue = types.StringValue(nameField.Elem().String())
		}
	} else {
		val := nameField.String()
		if val == "" {
			nameValue = types.StringNull()
		} else {
			nameValue = types.StringValue(val)
		}
	}

	// Handle type field (this is an enum, convert to string)
	if typeField.Kind() == reflect.Pointer {
		if typeField.IsNil() {
			typeValue = types.StringNull()
		} else {
			typeValue = types.StringValue(fmt.Sprintf("%v", typeField.Elem().Interface()))
		}
	} else {
		typeValue = types.StringValue(fmt.Sprintf("%v", typeField.Interface()))
	}

	// Create object value
	memberObj, _ := types.ObjectValue(
		MemberAttrTypes,
		map[string]attr.Value{
			"name": nameValue,
			"id":   idValue,
			"type": typeValue,
		},
	)
	return memberObj
}
