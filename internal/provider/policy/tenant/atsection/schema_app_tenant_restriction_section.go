package atsection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *appTenantRestrictionSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom section in the Cato app tenant restriction (App & Data Inline Protection) policy. " +
			"Adding a section can reassign existing rules into that section and create a draft revision; see provider documentation. " +
			"Underlying GraphQL is marked @beta.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Section ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Where to insert the section",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to policy or another section",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("AFTER_SECTION", "BEFORE_SECTION", "LAST_IN_POLICY"),
						},
					},
					"ref": schema.StringAttribute{
						Description: "Reference section ID when position is AFTER_SECTION or BEFORE_SECTION",
						Optional:    true,
					},
				},
			},
			"section": schema.SingleNestedAttribute{
				Description: "Section name and computed ID",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Section ID",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Section name",
						Required:    true,
					},
				},
			},
		},
	}
}
