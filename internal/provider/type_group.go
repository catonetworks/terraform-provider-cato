package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Group struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Members     types.Set    `tfsdk:"members"`
}

type GroupMembers struct {
	Id        types.String `tfsdk:"id"`
	GroupName types.String `tfsdk:"group_name"`
	Members   types.Set    `tfsdk:"members"`
}

type GroupMember struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

var GroupMemberAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"id":   types.StringType,
	"type": types.StringType,
}

var GroupMemberObjectType = types.ObjectType{AttrTypes: GroupMemberAttrTypes}
