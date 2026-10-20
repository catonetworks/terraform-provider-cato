package groupmembers

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
