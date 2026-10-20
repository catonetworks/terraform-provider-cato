package papolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func (r *privAccessPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_private_access_policy` resource contains the configuration parameters for " +
			"private access policies in the Cato platform.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the policy, always set to 0",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Is the private access policy enabled?",
				Required:    true,
			},
			"audit": schema.SingleNestedAttribute{
				Description:   "Audit log record",
				Computed:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"published_time": schema.StringAttribute{
						Description: "Audit log published time",
						Computed:    true,
					},
					"published_by": schema.StringAttribute{
						Description: "Author of the published log record",
						Computed:    true,
					},
				},
			},
		},
	}
}
