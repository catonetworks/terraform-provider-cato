package iam

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoleDataAttributes describes a complete role returned by data sources.
func RoleDataAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                         schema.StringAttribute{Description: "Composite account and role identifier.", Computed: true},
		"account_id":                 schema.StringAttribute{Description: "Account used for the lookup.", Computed: true},
		"role_id":                    schema.StringAttribute{Description: "API role identifier.", Computed: true},
		"name":                       schema.StringAttribute{Description: "Role name.", Computed: true},
		"description":                schema.StringAttribute{Description: "Role description.", Computed: true},
		"predefined":                 schema.BoolAttribute{Description: "Whether the role is predefined.", Computed: true},
		"is_used_on_external_access": schema.BoolAttribute{Description: "Whether external access uses the role.", Computed: true},
		"applicable_account_type": schema.SetAttribute{Description: "Applicable destination account types.",
			Computed:    true,
			ElementType: types.StringType},

		"permissions": schema.SetNestedAttribute{Description: "Permissions granted by the role.",
			Computed: true,
			NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"resource": schema.StringAttribute{Description: "Resource identifier.", Computed: true},
				"action":   schema.StringAttribute{Description: "Granted action.", Computed: true},
			}}},
	}
}

// DataAccountAttribute describes the optional provider-account override.
func DataAccountAttribute() schema.StringAttribute {
	return schema.StringAttribute{Description: "Account to query. Defaults to the provider account_id.", Optional: true, Computed: true}
}
