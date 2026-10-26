package roles

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

func stringFilterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{Description: "API filter operators.", Optional: true, Attributes: map[string]schema.Attribute{
		"eq": schema.StringAttribute{Description: "Equal to.",
			Optional: true},
		"neq": schema.StringAttribute{Description: "Not equal to.",
			Optional: true},

		"in":  schema.SetAttribute{Description: "Match any value.", Optional: true, ElementType: types.StringType},
		"nin": schema.SetAttribute{Description: "Exclude these values.", Optional: true, ElementType: types.StringType},
	}}
}
func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Lists matching custom and predefined roles, fetching all pages and preserving API name order.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Description: "Account identifier for this query.", Computed: true},
			"account_id": iam.DataAccountAttribute(),
			"sort_direction": schema.StringAttribute{Description: "Name sort direction, ASC (default) or DESC.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{stringvalidator.OneOf("ASC",
					"DESC")}},

			"filter": schema.SingleNestedAttribute{Description: "Optional role filters.", Optional: true, Attributes: map[string]schema.Attribute{
				"id": stringFilterSchema(), "name": stringFilterSchema(),
				"predefined": schema.SingleNestedAttribute{Description: "Predefined-status filter.",
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"eq": schema.BoolAttribute{Description: "Equal to.",
							Optional: true},
						"neq": schema.BoolAttribute{Description: "Not equal to.",
							Optional: true},
					}},
			}},
			"items": schema.ListNestedAttribute{Description: "All matching roles in API sort order.",
				Computed:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: iam.RoleDataAttributes()}},

			"total": schema.Int64Attribute{Description: "Number of matching roles.", Computed: true},
		}}
}
