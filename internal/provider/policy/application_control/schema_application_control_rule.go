package appcontrol

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

func (r *applicationControlRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	typed := applicationControlTypedRuleSchemaAttributes()
	resp.Schema = schema.Schema{
		Description: "Manages a rule in the Cato Application Control (App & Data Inline Protection) policy. " +
			"Underlying GraphQL is marked @beta; behavior and fields may change.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Rule ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Where to insert the rule",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to policy, section, or another rule",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(
								"AFTER_RULE", "BEFORE_RULE", "FIRST_IN_POLICY", "FIRST_IN_SECTION",
								"LAST_IN_POLICY", "LAST_IN_SECTION",
							),
						},
					},
					"ref": schema.StringAttribute{
						Description: "Reference rule or section ID when required by position",
						Optional:    true,
					},
				},
			},
			"rule": schema.SingleNestedAttribute{
				Description: "Rule definition",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Rule ID",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Rule name",
						Required:    true,
					},
					"description": schema.StringAttribute{
						Description: "Rule description",
						Optional:    true,
					},
					"enabled": schema.BoolAttribute{
						Description: "Whether the rule is enabled",
						Required:    true,
					},
					"rule_type": schema.StringAttribute{
						Description: "Which nested rule block is active",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(cato_models.ApplicationControlRuleTypeApplication),
								string(cato_models.ApplicationControlRuleTypeData),
								string(cato_models.ApplicationControlRuleTypeFile),
							),
						},
					},
					"application_rule": schema.SingleNestedAttribute{
						Description: "Settings when rule_type is APPLICATION",
						Optional:    true,
						Attributes:  typed,
					},
					"data_rule": schema.SingleNestedAttribute{
						Description: "Settings when rule_type is DATA",
						Optional:    true,
						Attributes:  typed,
					},
					"file_rule": schema.SingleNestedAttribute{
						Description: "Settings when rule_type is FILE",
						Optional:    true,
						Attributes:  typed,
					},
				},
			},
		},
	}
}
