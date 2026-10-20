package ifindex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *ifwRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves index values for Internet Firewall Rules.",
		Attributes: map[string]schema.Attribute{
			"section_to_start_after_id": schema.StringAttribute{
				Description: "IFW rule id",
				Required:    false,
				Optional:    true,
				// Computed:    true,
			},
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of IF Rule Policy Indexes keyed by rule_name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "IFW rule id",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"index_in_section": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"section_name": schema.StringAttribute{
							Description: "IFW section name housing rule",
							Required:    false,
							Optional:    true,
						},
						"rule_name": schema.StringAttribute{
							Description: "IFW rule name housing rule",
							Required:    false,
							Optional:    true,
						},
						"description": schema.StringAttribute{
							Description: "IFW rule description",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "IFW rule enabled",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseNonNullStateForUnknown(),
							},
						},
					},
				},
			},
			"section_data": schema.MapNestedAttribute{
				Description: "Map of IFW section Indexes keyed by section_name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "IFW section id housing rule",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"section_index": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    true,
							Optional:    false,
						},
						"section_name": schema.StringAttribute{
							Description: "IFW section name housing rule",
							Required:    true,
							Optional:    false,
						},
					},
				},
			},
		},
	}
}

var IfwRuleIndexResourceObjectTypes = types.ObjectType{AttrTypes: IfwRuleIndexResourceAttrTypes}

var IfwRuleIndexResourceAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"index_in_section": types.Int64Type,
	"section_name":     types.StringType,
	"rule_name":        types.StringType,
	"description":      types.StringType,
	"enabled":          types.BoolType,
}

var IfwSectionIndexResourceObjectTypes = types.ObjectType{AttrTypes: IfwSectionIndexResourceAttrTypes}

var IfwSectionIndexResourceAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"section_name":  types.StringType,
	"section_index": types.Int64Type,
}
