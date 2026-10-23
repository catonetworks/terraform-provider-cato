package parulebulk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func (r *privAccessRuleBulkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ordering and publishng private access policy rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the bulk, always set to 0",
				Computed:    true,
			},
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of private access policy rules keyed by name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Rule id",
							Computed:    true,
						},
						"index": schema.Int64Attribute{
							Description: "Required position of the rule",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Private access rule name",
							Required:    true,
						},
					},
				},
			},
			"publish": schema.Int64Attribute{
				Description: "publish policy revision if there is a change",
				Computed:    true,
			},
		},
	}
}
