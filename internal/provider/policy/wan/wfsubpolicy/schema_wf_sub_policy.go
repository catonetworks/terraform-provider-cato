package wfsubpolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
)

func (r *wfSubPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	var ruleSchema resource.SchemaResponse
	(&wfrule.WanFwRuleResource{}).Schema(ctx, resource.SchemaRequest{}, &ruleSchema)
	scopeAttr := ruleSchema.Schema.Attributes["rule"].(schema.SingleNestedAttribute)
	scopeAttr.Description = "Scope of the sub-policy. This is the SUB_POLICY_SCOPE rule that defines when the " +
		"sub-policy applies. Uses the same parameters as a cato_wf_rule rule."
	scopeAttr.Required = true
	scopeAttr.Attributes["action"] = schema.StringAttribute{
		Description: "API-managed action for the SUB_POLICY_SCOPE rule.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	scopeAttr.Attributes["name"] = schema.StringAttribute{
		Description: "API-managed scope name, synchronized with the sub-policy name.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	scopeAttr.Attributes["description"] = schema.StringAttribute{
		Description: "API-managed scope description, synchronized with the sub-policy description.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = schema.Schema{
		Description: "The `cato_wf_sub_policy` resource manages a WAN Firewall sub-policy " +
			"(a nested policy scoped by a SUB_POLICY_SCOPE rule). The Cato API has no updateSubPolicy " +
			"mutation, so changing `name` or `description` forces resource replacement. Documentation for the " +
			"underlying API can be found at mutation.policy.wanFirewall.addSubPolicy().",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Sub-policy ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Sub-policy name. Changing this forces replacement (no updateSubPolicy API).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Sub-policy description. Changing this forces replacement (no updateSubPolicy API).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope_rule_id": schema.StringAttribute{
				Description: "ID of the underlying SUB_POLICY_SCOPE rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Position of the sub-policy scope within the WAN Firewall policy.",
				Required:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
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
			"scope": scopeAttr,
		},
	}
}
