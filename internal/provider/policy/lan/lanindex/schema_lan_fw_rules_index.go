package lanindex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

//nolint:funlen // Declarative nested schema remains clearer in one function.
func (r *lanRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ordering of LAN firewall policy items.\n\n" +
			"**LAN Firewall Policy**\n" +
			"- Section\n" +
			"  - Network rule\n" +
			"    - Firewall rule\n" +
			"  - Sub-policy\n" +
			"    - Section\n" +
			"      - Network rule\n" +
			"        - Firewall rule",

		Attributes: map[string]schema.Attribute{
			"section_data": schema.MapNestedAttribute{
				Description: "Map of section indexes keyed by a caller-chosen stable key. " +
					"For backward compatibility, the key is used as section_name when section_name is omitted.",
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:   "Section id",
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
						},
						"section_name": schema.StringAttribute{
							Description: "LAN section name. Defaults to the map key for backward compatibility.",
							Optional:    true,
							Computed:    true,
						},
						"section_index": schema.Int64Attribute{
							Description: "Position of the section in the policy or sub-policy. Starts with 1.",
							Required:    true,
						},
						"sub_policy_name": schema.StringAttribute{
							Description: "Sub-policy name. If not set, the main policy is used.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
						},
					},
				},
			},
			"network_rules": schema.MapNestedAttribute{
				Description: "Map of network rule or sub-policy indexes keyed by a caller-chosen stable key. " +
					"For backward compatibility, the key is used as rule_name when rule_name is omitted.",
				Required: false,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:   "Network rule ID or sub-policy scope rule ID",
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
						},
						"rule_type": schema.StringAttribute{
							Description: "Rule type: POLICY_RULE, SUB_POLICY_SCOPE, SUB_RULE",
							Computed:    true,
						},
						"rule_name": schema.StringAttribute{
							Description: "Network rule or sub-policy name. Defaults to the map key for backward compatibility.",
							Optional:    true,
							Computed:    true,
						},
						"section_name": schema.StringAttribute{
							Description: "LAN section name housing rule",
							Required:    true,
						},
						"section_key": schema.StringAttribute{
							Description: "Map key of the parent section. Required when section_name is ambiguous.",
							Optional:    true,
						},
						"index_in_section": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    true,
						},
					},
				},
			},
			"firewall_rules": schema.MapNestedAttribute{
				Description: "Map of firewall rule indexes keyed by a caller-chosen stable key. " +
					"For backward compatibility, the key is used as firewall_rule_name when firewall_rule_name is omitted.",
				Required: false,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:   "Firewall rule ID",
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
						},
						"firewall_rule_name": schema.StringAttribute{
							Description: "Firewall rule name. Defaults to the map key for backward compatibility.",
							Optional:    true,
							Computed:    true,
						},
						"net_rule_name": schema.StringAttribute{
							Description: "Parent LAN network rule name",
							Required:    true,
						},
						"net_rule_key": schema.StringAttribute{
							Description: "Map key of the parent network rule. Required when net_rule_name is ambiguous.",
							Optional:    true,
						},
						"index_in_rule": schema.Int64Attribute{
							Description: "Index value remapped per network rule",
							Required:    true,
						},
					},
				},
			},
		},
	}
}
