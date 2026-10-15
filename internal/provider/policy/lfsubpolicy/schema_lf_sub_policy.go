package lfsubpolicy

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/lannetwork"
)

// Schema defines the Terraform schema for a LAN Firewall sub-policy.
func (r *lfSubPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_lf_sub_policy` resource manages a LAN Firewall sub-policy " +
			"(a nested policy scoped by a SUB_POLICY_SCOPE rule).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Sub-policy ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "Sub-policy name. Changing this forces replacement.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Description:   "Sub-policy description. Changing this forces replacement.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"scope_rule_id": schema.StringAttribute{
				Description:   "ID of the underlying SUB_POLICY_SCOPE rule.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"at": schema.SingleNestedAttribute{
				Description:   "Position of the sub-policy scope within the LAN Firewall policy.",
				Optional:      true,
				PlanModifiers: []planmodifier.Object{configuredPositionModifier{}},
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule.",
						Required:    true,
					},
					"ref": schema.StringAttribute{
						Description: "Identifier of the object relative to which the position is defined.",
						Optional:    true,
					},
				},
			},
			"scope": r.policyScopeSchema(),
		},
	}
}

// policyScopeSchema derives the sub-policy scope schema from the LAN rule schema.
func (r *lfSubPolicyResource) policyScopeSchema() schema.SingleNestedAttribute {
	scopeAttr := lannetwork.LanRuleSchema()
	scopeAttr.Description = "Policy scope attributes"
	scopeAttr.Required = true
	delete(scopeAttr.Attributes, "transport")
	scopeAttr.Attributes["description"] = schema.StringAttribute{
		Description:   "API-managed scope description, synchronized with the sub-policy description.",
		Computed:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
	scopeAttr.Attributes["name"] = schema.StringAttribute{
		Description:   "API-managed scope name, synchronized with the sub-policy name.",
		Computed:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
	service := scopeAttr.Attributes["service"].(schema.SingleNestedAttribute)
	custom := service.Attributes["custom"].(schema.ListNestedAttribute)
	custom.NestedObject.Attributes["protocol"] = schema.StringAttribute{
		Description: "IP Protocol (https://api.catonetworks.com/documentation/#definition-IpProtocol)",
		Required:    true,
	}
	service.Attributes["custom"] = custom
	scopeAttr.Attributes["service"] = service

	nat := scopeAttr.Attributes["nat"].(schema.SingleNestedAttribute)
	natType := nat.Attributes["nat_type"].(schema.StringAttribute)
	natType.Computed = true
	natType.Default = stringdefault.StaticString(string(cato_models.SocketLanNatTypeDynamicPat))
	nat.Attributes["nat_type"] = natType
	scopeAttr.Attributes["nat"] = nat

	return scopeAttr
}
