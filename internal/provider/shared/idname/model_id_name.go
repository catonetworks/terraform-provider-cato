package idname

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var ModelTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}
