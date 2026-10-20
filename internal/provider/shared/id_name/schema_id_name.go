package idname

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func SchemaNameID(prefix string) map[string]schema.Attribute {
	if prefix != "" {
		prefix += " "
	}
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: prefix + "name",
			Optional:    true,
			Computed:    true,
		},
		"id": schema.StringAttribute{
			Description: prefix + "ID",
			Optional:    true,
			Computed:    true,
		},
	}
}
