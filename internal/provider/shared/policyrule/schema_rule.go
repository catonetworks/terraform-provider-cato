package policyrule

import (
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// SchemaPosition returns the schema for PolicyRulePositionInput, matching PositionObjectType.
func SchemaPosition() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Position of the new rule in the policy",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"position": schema.StringAttribute{
				Description: "Insertion position (AFTER_RULE, BEFORE_RULE, FIRST_IN_POLICY, FIRST_IN_SECTION, " +
					"LAST_IN_POLICY, or LAST_IN_SECTION)",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicyRulePositionEnum)...),
				},
			},
			"ref": schema.StringAttribute{
				Description:   "ID of the rule or section relative to which the new rule is positioned",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
			},
		},
	}
}
