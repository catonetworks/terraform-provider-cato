package atpolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *appTenantRestrictionPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Account-level App Tenant Restriction policy toggles. " +
			"Underlying GraphQL is marked @beta.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed identifier for import",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the App Tenant Restriction policy is enabled",
				Required:    true,
			},
			"published_time": schema.StringAttribute{
				Description: "Last published time from audit",
				Computed:    true,
			},
			"published_by": schema.StringAttribute{
				Description: "Last published by from audit",
				Computed:    true,
			},
			"revision_id": schema.StringAttribute{
				Description: "Active draft revision ID when present",
				Computed:    true,
			},
			"revision_name": schema.StringAttribute{
				Description: "Active draft revision name when present",
				Computed:    true,
			},
		},
	}
}
