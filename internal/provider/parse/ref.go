package parse

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func SchemaNameID(prefix string) map[string]schema.Attribute {
	if prefix != "" {
		prefix += " "
	}
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: prefix + "name",
			Required:    true,
		},
		"id": schema.StringAttribute{
			Description: prefix + "ID",
			Optional:    false,
			Computed:    true,
		},
	}
}
