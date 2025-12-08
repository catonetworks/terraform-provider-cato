package provider

import (
	"context"
	"fmt"
	"reflect"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	client *catoClientData
}

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_group` resource contains the configuration parameters necessary to manage a group. Groups can contain various member types including sites, hosts, network ranges, and more. Documentation for the underlying API used in this resource can be found at [mutation.groups.createGroup()](https://api.catonetworks.com/documentation/#mutation-groups.createGroup).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Group ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Group name",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Group description",
				Optional:    true,
			},
			"members": schema.SetNestedAttribute{
				Description: "List of group members. Each member has 'name', 'id', and 'type' fields. You can specify either 'name' or 'id' when creating/updating; both will be populated in state. At least one member is required.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Member name (computed if not specified)",
							Optional:    true,
							Computed:    true,
						},
						"id": schema.StringAttribute{
							Description: "Member ID (computed if not specified)",
							Optional:    true,
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Member type: FLOATING_SUBNET, GLOBAL_IP_RANGE, HOST, NETWORK_INTERFACE, SITE, SITE_NETWORK_SUBNET", // USER, USERS_GROUP, SYSTEM_GROUP
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"FLOATING_SUBNET",
									"GLOBAL_IP_RANGE",
									"HOST",
									"NETWORK_INTERFACE",
									"SITE",
									"SITE_NETWORK_SUBNET",
									// "USER",
									// "USERS_GROUP",
									// "SYSTEM_GROUP",
								),
							},
						},
					},
				},
			},
		},
	}
}

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan Group
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert members from Terraform set to SDK input
	var members []*cato_models.GroupMemberRefTypedInput
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var membersList []GroupMember
		diags = plan.Members.ElementsAs(ctx, &membersList, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, member := range membersList {
			// Determine whether to use ID or NAME based on which field is provided
			var by cato_models.ObjectRefBy
			var input string

			if !member.Id.IsNull() && !member.Id.IsUnknown() && member.Id.ValueString() != "" {
				// Use ID if provided
				by = cato_models.ObjectRefByID
				input = member.Id.ValueString()
			} else if !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "" {
				// Fall back to NAME if ID not provided
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			} else {
				resp.Diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return
			}

			members = append(members, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	}

	// setting input
	input := cato_models.CreateGroupInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
		Members:     members,
	}

	tflog.Debug(ctx, "Create.GroupsCreateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.GroupsCreateGroup(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.GroupsCreateGroup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API GroupsCreateGroup error",
			err.Error(),
		)
		return
	}

	// Set the ID from the response
	plan.Id = types.StringValue(result.Groups.CreateGroup.Group.ID)

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydrateGroupState(ctx, plan.Id.ValueString(), plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating group state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan Group
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Group
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert members from Terraform set to SDK input
	var members []*cato_models.GroupMemberRefTypedInput
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var membersList []GroupMember
		diags = plan.Members.ElementsAs(ctx, &membersList, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, member := range membersList {
			// Determine whether to use ID or NAME based on which field is provided
			var by cato_models.ObjectRefBy
			var input string

			if !member.Id.IsNull() && !member.Id.IsUnknown() && member.Id.ValueString() != "" {
				// Use ID if provided
				by = cato_models.ObjectRefByID
				input = member.Id.ValueString()
			} else if !member.Name.IsNull() && !member.Name.IsUnknown() && member.Name.ValueString() != "" {
				// Fall back to NAME if ID not provided
				by = cato_models.ObjectRefByName
				input = member.Name.ValueString()
			} else {
				resp.Diagnostics.AddError(
					"Invalid member configuration",
					"Each member must specify either 'id' or 'name'",
				)
				return
			}

			members = append(members, &cato_models.GroupMemberRefTypedInput{
				By:    by,
				Input: input,
				Type:  cato_models.GroupMemberRefType(member.Type.ValueString()),
			})
		}
	} else {
		members = []*cato_models.GroupMemberRefTypedInput{}
	}

	groupRef := &cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: plan.Id.ValueString(),
	}

	// Only send name if it has changed from state
	var nameToSend *string
	if plan.Name.ValueString() != state.Name.ValueString() {
		nameToSend = plan.Name.ValueStringPointer()
	}

	input := cato_models.UpdateGroupInput{
		Group:       groupRef,
		Name:        nameToSend,
		Description: plan.Description.ValueStringPointer(),
		Members:     members,
	}

	tflog.Debug(ctx, "Update.GroupsUpdateGroup.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.GroupsUpdateGroup(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Update.GroupsUpdateGroup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API GroupsUpdateGroup error",
			err.Error(),
		)
		return
	}

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydrateGroupState(ctx, plan.Id.ValueString(), plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating group state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state Group
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API using current state as plan reference
	hydratedState, hydrateErr := r.hydrateGroupState(ctx, state.Id.ValueString(), state)
	if hydrateErr != nil {
		// Check if group not found
		if hydrateErr.Error() == "group not found" {
			tflog.Warn(ctx, "group not found, resource removed")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error hydrating group state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state Group
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the group
	groupRef := cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: state.Id.ValueString(),
	}

	result, err := r.client.catov2.GroupsDeleteGroup(ctx, groupRef, r.client.AccountId)
	tflog.Debug(ctx, "Delete.GroupsDeleteGroup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		// Check if the error is because the group is already deleted
		resp.Diagnostics.AddError(
			"Cato API GroupsDeleteGroup error",
			fmt.Sprintf("Error deleting group: %s", err.Error()),
		)
		return
	}
}

// hydrateGroupState fetches the current state of a group from the API
// It takes a plan parameter to match config members with API members correctly
func (r *groupResource) hydrateGroupState(ctx context.Context, groupID string, plan Group) (Group, error) {
	var state Group
	state.Id = types.StringValue(groupID)

	// Read group metadata using GroupsList
	groupListInput := &cato_models.GroupListInput{
		Filter: []*cato_models.GroupListFilterInput{
			{
				ID: []*cato_models.IDFilterInput{
					{
						Eq: &groupID,
					},
				},
			},
		},
		Paging: &cato_models.PagingInput{
			Limit: 1,
			From:  0,
		},
		Sort: &cato_models.GroupListSortInput{},
	}

	groupListResult, err := r.client.catov2.GroupsList(ctx, groupListInput, r.client.AccountId)
	tflog.Debug(ctx, "hydrateGroupState.GroupsList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(groupListResult),
	})
	if err != nil {
		return state, fmt.Errorf("GroupsList API error: %w", err)
	}

	// Check if the group was found
	if groupListResult.Groups.GroupList == nil || len(groupListResult.Groups.GroupList.Items) == 0 {
		return state, fmt.Errorf("group not found")
	}

	// Update state with group details from GroupsList
	groupItem := groupListResult.Groups.GroupList.Items[0]
	state.Name = types.StringValue(groupItem.Name)
	if groupItem.Description != nil {
		state.Description = types.StringValue(*groupItem.Description)
	} else {
		state.Description = types.StringNull()
	}

	// Read group members using GroupsMembers
	groupRef := cato_models.GroupRefInput{
		By:    cato_models.ObjectRefByID,
		Input: groupID,
	}

	groupMembersListInput := cato_models.GroupMembersListInput{
		Filter: []*cato_models.GroupMembersListFilterInput{},
		Paging: &cato_models.PagingInput{
			Limit: 1000,
			From:  0,
		},
		Sort: &cato_models.GroupMembersListSortInput{},
	}

	membersResult, err := r.client.catov2.GroupsMembers(ctx, groupRef, groupMembersListInput, r.client.AccountId)
	tflog.Debug(ctx, "hydrateGroupState.GroupsMembers.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(membersResult),
	})
	if err != nil {
		return state, fmt.Errorf("GroupsMembers API error: %w", err)
	}

	// Convert members to Terraform set, matching with plan/config to preserve user's choice of ID vs name
	state.Members = parseGroupMemberListWithConfig(ctx, membersResult.Groups.Group.Members.Items, plan.Members, "group.members")

	return state, nil
}

// parseGroupMemberListWithConfig converts API group member items to a Terraform set
// matching with config/plan to preserve user's choice of specifying ID vs name
func parseGroupMemberListWithConfig[T any](ctx context.Context, items []T, configMembers types.Set, attrName string) types.Set {
	tflog.Debug(ctx, "parseGroupMemberListWithConfig() "+attrName)

	// Handle empty API list
	if len(items) == 0 {
		tflog.Debug(ctx, "parseGroupMemberListWithConfig() - empty API list")
		return types.SetValueMust(types.ObjectType{AttrTypes: GroupMemberAttrTypes}, []attr.Value{})
	}

	// Parse config members to create a matching map
	// Key: "type:id:value" or "type:name:value"
	var configMembersList []GroupMember
	configMap := make(map[string]GroupMember)

	if !configMembers.IsNull() && !configMembers.IsUnknown() {
		if diags := configMembers.ElementsAs(ctx, &configMembersList, false); !diags.HasError() {
			for _, configMember := range configMembersList {
				memberType := configMember.Type.ValueString()
				// Create lookup keys based on what was specified in config
				if !configMember.Id.IsNull() && !configMember.Id.IsUnknown() && configMember.Id.ValueString() != "" {
					key := memberType + ":id:" + configMember.Id.ValueString()
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
		apiMemberObj := parseGroupMember(ctx, item, attrName)
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

		// Try to find matching config member
		var matchedConfig *GroupMember
		if apiID != "" && apiType != "" {
			key := apiType + ":id:" + apiID
			if config, exists := configMap[key]; exists {
				matchedConfig = &config
			}
		}
		if matchedConfig == nil && apiName != "" && apiType != "" {
			key := apiType + ":name:" + apiName
			if config, exists := configMap[key]; exists {
				matchedConfig = &config
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

		resultObj, _ := types.ObjectValue(GroupMemberAttrTypes, resultAttrs)
		memberValues = append(memberValues, resultObj)
	}

	// Convert to types.Set
	membersSet, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: GroupMemberAttrTypes}, memberValues)
	return membersSet
}

// parseGroupMemberList converts API group member items to a Terraform set
// Similar to parseNameIDList but includes the type field
func parseGroupMemberList[T any](ctx context.Context, items []T, attrName string) types.Set {
	tflog.Debug(ctx, "parseGroupMemberList() "+attrName)

	// Handle empty list - return empty set (not null) to match schema default
	if len(items) == 0 {
		tflog.Debug(ctx, "parseGroupMemberList() - empty input list, returning empty set")
		return types.SetValueMust(types.ObjectType{AttrTypes: GroupMemberAttrTypes}, []attr.Value{})
	}

	// Process each item into an attr.Value
	memberValues := make([]attr.Value, 0, len(items))
	for _, item := range items {
		memberObj := parseGroupMember(ctx, item, attrName)
		if !memberObj.IsNull() {
			memberValues = append(memberValues, memberObj)
		}
	}

	// Convert to types.Set
	membersSet, _ := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: GroupMemberAttrTypes}, memberValues)
	return membersSet
}

// parseGroupMember converts a single API group member item to a Terraform object
func parseGroupMember(ctx context.Context, item interface{}, attrName string) types.Object {
	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(GroupMemberAttrTypes)
			}
		} else {
			return types.ObjectNull(GroupMemberAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		if itemValue.IsNil() {
			return types.ObjectNull(GroupMemberAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name, ID, and Type fields
	nameField := itemValue.FieldByName("Name")
	idField := itemValue.FieldByName("ID")
	typeField := itemValue.FieldByName("TypeGroupMemberRefTyped")

	if !nameField.IsValid() || !idField.IsValid() || !typeField.IsValid() {
		return types.ObjectNull(GroupMemberAttrTypes)
	}

	// Extract string values
	var nameValue, idValue, typeValue types.String

	// Handle ID field
	if idField.Kind() == reflect.Ptr {
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
	if nameField.Kind() == reflect.Ptr {
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
	if typeField.Kind() == reflect.Ptr {
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
		GroupMemberAttrTypes,
		map[string]attr.Value{
			"name": nameValue,
			"id":   idValue,
			"type": typeValue,
		},
	)
	return memberObj
}
