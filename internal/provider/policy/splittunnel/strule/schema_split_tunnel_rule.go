package strule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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
func (r *splitTunnelRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule in the Cato split tunnel policy. " +
			"Underlying GraphQL is marked @beta; behavior and fields may change. " +
			"Explicitly configured empty filters are preserved in state when the API omits them.",
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
								string(cato_models.SplitTunnelActionEnumExclude),
								string(cato_models.SplitTunnelActionEnumInclude),
								"ROUTE_ALL_TO_CATO_EXCEPT", "ROUTE_ONLY_SELECTED", "USER_DEFINED",
							),
						},
					},
					"coverage": schema.StringAttribute{
						Description: "Traffic coverage (ALL or WEB_ONLY)",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ALL", "WEB_ONLY"),
						},
					},
					"routing_priority": schema.StringAttribute{
						Description: "Routing priority (LAN or TUNNEL)",
						Required:    true,
					},
					"country": refSetAttribute("Countries the rule applies to"),
					"platform": schema.ListAttribute{
						Description: "Operating systems the rule applies to",
						Optional:    true,
						ElementType: types.StringType,
					},
					"destination": schema.SingleNestedAttribute{
						Description: "Destinations the rule applies to",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"application": refSetAttribute("Application references"),
							"domain": schema.ListAttribute{
								Description: "Domains",
								Optional:    true,
								ElementType: types.StringType,
							},
							"fqdn": schema.ListAttribute{
								Description: "FQDNs",
								Optional:    true,
								ElementType: types.StringType,
							},
							"global_ip_range": refSetAttribute("Global IP range references"),
						},
					},
					"dns_exclusion": schema.SingleNestedAttribute{
						Description: "Domains excluded from DNS tunneling",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"domain": schema.ListAttribute{
								Description: "Domains",
								Optional:    true,
								ElementType: types.StringType,
							},
						},
					},
					"source": schema.SingleNestedAttribute{
						Description: "Users and user groups the rule applies to",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"user":        refSetAttribute("User references"),
							"users_group": refSetAttribute("User group references"),
						},
					},
					"source_network": schema.SingleNestedAttribute{
						Description:   "Source network the rule applies to",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
						Attributes: map[string]schema.Attribute{
							"source_network_type": schema.StringAttribute{
								Description: "Source network type (ANY, ANY_MANAGED_NETWORK, ANY_UNMANAGED_NETWORK)",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("ANY", "ANY_MANAGED_NETWORK", "ANY_UNMANAGED_NETWORK"),
								},
							},
						},
					},
				},
			},
		},
	}
}
