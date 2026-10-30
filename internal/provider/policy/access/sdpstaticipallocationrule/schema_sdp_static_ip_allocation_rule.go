package sdpstaticipallocationrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a sdp static ip allocation rule. Changes are automatically " +
			"published. Omitting at on update " +
			"preserves placement.",
		Attributes: schemaResource(),
	}
}

func schemaResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"at": schema.SingleNestedAttribute{
			Description: "At",
			Optional:    true,
			Attributes:  policyshared.RulePositionAttributes("Position", "Ref"),
		},
		"rule": schema.SingleNestedAttribute{
			Description: "Rule",
			Required:    true,
			Attributes:  schemaResourceRule(),
		},
	}
}

func schemaResourceRule() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Description: "Description",
			Required:    true,
		},
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Required:    true,
		},
		"id": schema.StringAttribute{
			Description:   "Rule ID",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"index": schema.Int64Attribute{
			Description: "Current rule index",
			Computed:    true,
		},
		"ip": schema.StringAttribute{
			Description: "Ip",
			Required:    true,
		},
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"section": schema.SingleNestedAttribute{
			Description: "Policy section containing the rule",
			Computed:    true,
			Attributes:  policyshared.RuleSectionAttributes(),
		},
		"source": schema.SingleNestedAttribute{
			Description: "Source",
			Required:    true,
			Attributes:  schemaResourceRuleSource(),
		},
	}
}

func schemaResourceRuleSource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"user": schema.SingleNestedAttribute{
			Description:   "User",
			Required:      true,
			Attributes:    idname.SchemaNameID("User"),
			PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
		},
	}
}
