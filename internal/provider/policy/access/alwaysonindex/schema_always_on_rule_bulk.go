package alwaysonindex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *alwaysOnRuleBulkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ordering of existing Always On policy sections and rules.",
		Attributes: map[string]schema.Attribute{
			"section_data": schema.MapNestedAttribute{
				Description: "Map of Always On section positions keyed by section name",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schemaIndexID("Always On policy section ID"),
					"section_index": schema.Int64Attribute{
						Description: "Position of the section within the policy, starting at 1",
						Required:    true,
						Validators:  []validator.Int64{int64validator.AtLeast(1)},
					},
				}},
				Validators: []validator.Map{sectionDataValidator{}},
			},
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of Always On rule positions keyed by rule name",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schemaIndexID("Always On rule ID"),
					"index_in_section": schema.Int64Attribute{
						Description: "Position of the rule within its section, starting at 1",
						Required:    true,
						Validators:  []validator.Int64{int64validator.AtLeast(1)},
					},
					"section_name": schema.StringAttribute{
						Description: "Name of the parent Always On section, matching a section_data map key",
						Required:    true,
					},
				}},
				Validators: []validator.Map{ruleDataValidator{}},
			},
			"changes": schema.Int64Attribute{
				Description:   "Internal flag",
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{newModChanges()},
			},
		},
	}
}

func schemaIndexID(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Description:   description,
		Computed:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
	}
}
