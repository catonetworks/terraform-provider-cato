package group

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Group struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Members     types.Set    `tfsdk:"members"`
}

type Members struct {
	ID        types.String `tfsdk:"id"`
	GroupName types.String `tfsdk:"group_name"`
	Members   types.Set    `tfsdk:"members"`
}

type Member struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

var MemberAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"id":   types.StringType,
	"type": types.StringType,
}

var MemberObjectType = types.ObjectType{AttrTypes: MemberAttrTypes}

// Lookup is the type for the groups data source with filters
type Lookup struct {
	IDFilter   types.List `tfsdk:"id_filter"`
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

// ItemAttrTypes defines the attribute types for a group item in the lookup results
var ItemAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"members":     types.SetType{ElemType: MemberObjectType},
}
