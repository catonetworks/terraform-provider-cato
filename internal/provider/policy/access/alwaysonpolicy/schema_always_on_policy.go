package alwaysonpolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func (r *alwaysOnPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages whether the account's Always On policy is enabled.",
		Attributes: map[string]schema.Attribute{
			"published_time": schema.StringAttribute{Description: "Time of the last policy publication.", Computed: true},
			"published_by":   schema.StringAttribute{Description: "Administrator who last published the policy.", Computed: true},
			"revision_id":    schema.StringAttribute{Description: "Active draft revision ID, or null when no draft exists.", Computed: true},
			"revision_name":  schema.StringAttribute{Description: "Active draft revision name, or null when no draft exists.", Computed: true},
			"enabled": schema.BoolAttribute{
				Description: "Whether the Always On policy is enabled",
				Required:    true,
			},
		},
	}
}
