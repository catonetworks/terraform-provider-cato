package networkconfig

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
)

// SchemaDNSAttributes defines shared account and site DNS attributes.
func SchemaDNSAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": IDAttribute(),
		"primary_server": schema.StringAttribute{
			Optional:      true,
			Computed:      true,
			Description:   "Primary DNS server. Omission clears it; secondary server and suffixes require a primary.",
			Validators:    []validator.String{IPValidator{AllowEmpty: true}},
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
		},
		"secondary_server": schema.StringAttribute{
			Optional:      true,
			Computed:      true,
			Description:   "Secondary DNS server. Omission clears it.",
			Validators:    []validator.String{IPValidator{AllowEmpty: true}},
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
		},
		"suffix": schema.ListAttribute{
			Optional:    true,
			Computed:    true,
			ElementType: types.StringType,
			Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, nil)),
			Description: "Ordered DNS search suffixes. Omission or an empty list clears the list.",
		},
	}
}
