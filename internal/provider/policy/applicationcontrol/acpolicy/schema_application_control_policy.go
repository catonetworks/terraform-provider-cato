package acpolicy

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *applicationControlPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Account-level Application Control (App & Data Inline Protection) policy toggles. " +
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
				Description: "Whether the Application Control policy is enabled",
				Required:    true,
			},
			"data_control_enabled": schema.StringAttribute{
				Description: "Data control feature toggle (ENABLED or DISABLED)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(cato_models.PolicyToggleStateEnabled), string(cato_models.PolicyToggleStateDisabled)),
				},
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
