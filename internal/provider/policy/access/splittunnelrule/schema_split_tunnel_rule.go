package splittunnelrule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a split tunnel rule. Changes are automatically published. Omitting at on update preserves " +
			"placement.",
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
		"action": schema.StringAttribute{
			Description: "The action applied by the split tunnel if the rule is matched",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllSplitTunnelActionEnum)...)},
		},
		"country": schema.SetNestedAttribute{
			Description: "Country traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ " +
				"is applied between criteria sets.",
			Required: true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Country"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"coverage": schema.StringAttribute{
			Description: "Coverage",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllSplitTunnelCoverageEnum)...)},
		},
		"description": schema.StringAttribute{
			Description: "Description",
			Required:    true,
		},
		"destination": schema.SingleNestedAttribute{
			Description: "Destination traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical " +
				"‘AND’ is applied between criteria sets.",
			Required:   true,
			Attributes: schemaResourceRuleDestination(),
		},
		"dns_exclusion": schema.SingleNestedAttribute{
			Description: "Dns exclusion",
			Optional:    true,
			Attributes:  schemaResourceRuleDNSExclusion(),
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
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"platform": schema.SetAttribute{
			Description: "Source device Operating System traffic matching criteria. Logical ‘OR’ is applied within the " +
				"criteria set. Logical ‘AND’ is applied between criteria sets.",
			Required:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(
					stringvalidator.OneOf(utils.ToStrings(cato_models.AllOperatingSystem)...),
				),
			},
		},
		"routing_priority": schema.StringAttribute{
			Description: "Routing priority",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllSplitTunnelRoutingPriorityEnum)...)},
		},
		"section": schema.SingleNestedAttribute{
			Description: "Policy section containing the rule",
			Computed:    true,
			Attributes:  schemaResourceRuleSection(),
		},
		"source": schema.SingleNestedAttribute{
			Description: "Source traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is " +
				"applied between criteria sets.",
			Required:   true,
			Attributes: schemaResourceRuleSource(),
		},
		"source_network": schema.SingleNestedAttribute{
			Description: "Source network",
			Required:    true,
			Attributes:  schemaResourceRuleSourceNetwork(),
		},
	}
}

func schemaResourceRuleDestination() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"application": schema.SetNestedAttribute{
			Description: "Applications for the rule (pre-defined)",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Application"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"domain": schema.SetAttribute{
			Description: "A Second-Level Domain (SLD). It matches all Top-Level Domains (TLD), and subdomains that include " +
				"the Domain. Example: example.com.",
			Required:    true,
			ElementType: types.StringType,
		},
		"fqdn": schema.SetAttribute{
			Description: "An exact match of the fully qualified domain (FQDN). Example: www.my.example.com.",
			Required:    true,
			ElementType: types.StringType,
		},
		"global_ip_range": schema.SetNestedAttribute{
			Description: "Globally defined IP range",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Global ip range"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
	}
}

func schemaResourceRuleDNSExclusion() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"domain": schema.SetAttribute{
			Description: "A Second-Level Domain (SLD). It matches all Top-Level Domains (TLD), and subdomains that include " +
				"the Domain. Example: example.com.",
			Required:    true,
			ElementType: types.StringType,
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

func schemaResourceRuleSource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"user": schema.SetNestedAttribute{
			Description: "Individual users defined for the account",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("User"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"users_group": schema.SetNestedAttribute{
			Description: "Group of users",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Users group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
	}
}

func schemaResourceRuleSourceNetwork() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"source_network_type": schema.StringAttribute{
			Description: "Source network type",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllSplitTunnelSourceNetworkTypeEnum)...)},
		},
	}
}
