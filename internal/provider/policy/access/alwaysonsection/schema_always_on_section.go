package alwaysonsection

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *alwaysOnSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a section in the Cato Always On policy.",
		Attributes: map[string]schema.Attribute{
			"at": schema.SingleNestedAttribute{
				Description: "Position of the section in the Always On policy",
				Required:    true,
				Validators:  []validator.Object{sectionPositionValidator{}},
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Insertion position (AFTER_SECTION, BEFORE_SECTION, or LAST_IN_POLICY)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicySectionPositionEnum)...),
						},
					},
					"ref": schema.StringAttribute{
						Description:   "Reference section ID; required for AFTER_SECTION and BEFORE_SECTION",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
					},
				},
			},
			"section": schema.SingleNestedAttribute{
				Description: "Parameters for the Always On policy section",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description:   "Always On policy section ID",
						Computed:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
					},
					"name": schema.StringAttribute{Description: "Always On policy section name", Required: true},
				},
			},
		},
	}
}
