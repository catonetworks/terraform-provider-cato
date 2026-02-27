package parse

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IdNameRefModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var IdNameRefModelTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}
