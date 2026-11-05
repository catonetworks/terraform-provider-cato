package terminalserverrule

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
		Description: "Manages a terminal server rule. Changes are automatically published. " +
			"Omitting at on update " +
			"preserves placement.",
		Attributes: schemaResource(),
	}
}

func schemaResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"at": schema.SingleNestedAttribute{
			Description: "Position of the rule in the policy",
			Optional:    true,
			Attributes: policyshared.RulePositionAttributes("Position relative to a policy, a section or another rule",
				"The identifier of the object (e.g. a rule, a section) relative to which "+
					"the position of the added rule is defined"),
		},
		"rule": schema.SingleNestedAttribute{
			Description: "Parameters for the rule you are adding",
			Required:    true,
			Attributes:  schemaResourceRule(),
		},
	}
}

func schemaResourceRule() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"allowed_host_ip": schema.SingleNestedAttribute{
			Description: "Allowed Host IP range. all the other IPs will be blocked by the pop. " +
				"Globally defined IP range",
			Required:      true,
			Attributes:    idname.SchemaNameID("Allowed host ip"),
			PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
		},
		"description": schema.StringAttribute{
			Description: "Description",
			Required:    true,
		},
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Required:    true,
		},
		"exclude_traffic": schema.SetNestedAttribute{
			Description: "Exclude traffic IP Range. all traffic to those Ips will be excluded from " +
				"the GRE tunnel. Globally " +
				"defined IP range",
			Required: true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Exclude traffic"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
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
	}
}
