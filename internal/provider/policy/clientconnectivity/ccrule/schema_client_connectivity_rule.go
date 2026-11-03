package ccrule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func refSetAttribute(description string) schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description: description,
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"name": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

//nolint:funlen
func (r *clientConnectivityRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule in the Cato client connectivity policy. " +
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
					"action": schema.StringAttribute{
						Description: "Rule action",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(cato_models.ClientConnectivityActionEnumAllow),
								string(cato_models.ClientConnectivityActionEnumAllowInternet),
								string(cato_models.ClientConnectivityActionEnumBlock),
							),
						},
					},
					"confidence_level": schema.StringAttribute{
						Description: "Device posture confidence level the rule requires",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ANY", "HIGH", "LOW"),
						},
					},
					"connection_origin": schema.ListAttribute{
						Description: "Connection origins (ANY, REMOTE, REMOTE_EXTENSION)",
						Optional:    true,
						ElementType: types.StringType,
					},
					"country": refSetAttribute("Countries the rule applies to"),
					"device":  refSetAttribute("Device posture profiles the rule applies to"),
					"platform": schema.ListAttribute{
						Description: "Operating systems the rule applies to",
						Optional:    true,
						ElementType: types.StringType,
					},
					"source": schema.SingleNestedAttribute{
						Description: "Users and user groups the rule applies to",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"user":        refSetAttribute("User references"),
							"users_group": refSetAttribute("User group references"),
						},
					},
					"source_range": schema.ListNestedAttribute{
						Description: "Source global IP ranges the rule applies to",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"global_ip_range": schema.SingleNestedAttribute{
									Description: "Global IP range reference",
									Required:    true,
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Optional: true,
											Computed: true,
										},
										"name": schema.StringAttribute{
											Optional: true,
											Computed: true,
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
