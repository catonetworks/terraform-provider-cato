package globalipranges

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *globalIPRangesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_global_ip_ranges` resource contains the configuration parameters for " +
			"global IP ranges in the Cato platform. " +
			"There should be only one resource of this type in the TF config and it manages all the global IP ranges for a given account.",

		Attributes: map[string]schema.Attribute{
			"ranges": schema.SetNestedAttribute{
				Description: "List of global IP ranges. Values can be a single IP, an inclusive start-end range, or a CIDR block.",
				Required:    true,
				Validators:  []validator.Set{GetGlobalIPRangeValidator()},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Description: "Global IP range description",
							Optional:    true,
						},
						"id": schema.StringAttribute{
							Description: "Global IP range ID",
							Computed:    true,
							// PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"ip_range": schema.StringAttribute{
							Description: "Global IP range as a single IP, inclusive start-end range, or CIDR block.",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Global IP range name",
							Required:    true,
						},
					},
				},
			},
		},
	}
}
