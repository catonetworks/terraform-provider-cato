package terminalserverrule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a terminal server rule. Changes are automatically published. Omitting at on update " +
			"preserves placement.",
		Attributes: schemaResource(),
	}
}

func schemaResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"at": schema.SingleNestedAttribute{
			Description: "Position of the rule in the policy",
			Optional:    true,
			Attributes:  schemaResourceAt(),
		},
		"rule": schema.SingleNestedAttribute{
			Description: "Parameters for the rule you are adding",
			Required:    true,
			Attributes:  schemaResourceRule(),
		},
	}
}

func schemaResourceAt() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"position": schema.StringAttribute{
			Description: "Position relative to a policy, a section or another rule",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicyRulePositionEnum)...)},
		},
		"ref": schema.StringAttribute{
			Description: "The identifier of the object (e.g. a rule, a section) relative to which the position of the added " +
				"rule is defined",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
		},
	}
}

func schemaResourceRule() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"allowed_host_ip": schema.SingleNestedAttribute{
			Description:   "Allowed Host IP range. all the other IPs will be blocked by the pop. Globally defined IP range",
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
			Description: "Exclude traffic IP Range. all traffic to those Ips will be excluded from the GRE tunnel. Globally " +
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
			Attributes:  schemaResourceRuleSection(),
		},
	}
}

func schemaResourceRuleSection() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description:   "Section ID",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": schema.StringAttribute{
			Description: "Section name",
			Computed:    true,
		},
	}
}
