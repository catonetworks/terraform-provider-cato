package ssiarule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *sdpStaticIPAllocationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule in the Cato SDP static IP allocation policy. " +
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
					"ip": schema.StringAttribute{
						Description: "Static IP address allocated to the user",
						Required:    true,
					},
					"source": schema.SingleNestedAttribute{
						Description: "User the static IP is allocated to",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"user": schema.SingleNestedAttribute{
								Description: "User reference",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"name": schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
