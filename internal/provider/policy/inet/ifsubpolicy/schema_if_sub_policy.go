package ifsubpolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifrule"
)

//nolint:funlen // Scope and cleanup schemas intentionally mirror the rule schema.
func (r *ifSubPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Reuse the exact cato_if_rule "rule" schema for the sub-policy scope so the
	// scope stays in lockstep (drift-free) with the rule resource.
	var ruleSchema resource.SchemaResponse
	ifrule.IfwSchema(ctx, resource.SchemaRequest{}, &ruleSchema)
	scopeAttr := ruleSchema.Schema.Attributes["rule"].(schema.SingleNestedAttribute)
	scopeAttr.Description = "Scope of the sub-policy. This is the SUB_POLICY_SCOPE rule that defines when the " +
		"sub-policy applies. Uses the same parameters as a cato_if_rule rule."
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
	// The rule schema marks tracking as Required, but the API rejects tracking
	// settings on SUB_POLICY_SCOPE rules ("You can't include tracking settings in
	// a sub-policy scoping rule"). Relax it to Optional+Computed so configs omit
	// it; the rule hydrator skips null tracking, keeping it out of the mutation.
	scopeTracking := scopeAttr.Attributes["tracking"].(schema.SingleNestedAttribute)
	scopeTracking.Description = "Tracking is API-managed on SUB_POLICY_SCOPE rules and must be omitted here — " +
		"the API rejects user-supplied tracking settings on scoping rules."
	scopeTracking.Required = false
	scopeTracking.Optional = true
	scopeTracking.Computed = true
	scopeTracking.PlanModifiers = []planmodifier.Object{
		objectplanmodifier.UseStateForUnknown(),
	}
	scopeAttr.Attributes["tracking"] = scopeTracking

	// Second, independent clone of the rule schema for the API-created cleanup
	// rule (a fresh IfwSchema call — the nested Attributes map is shared by
	// reference, so reusing ruleSchema would leak the scope's overrides).
	// addSubPolicy always materializes a "<name> - Cleanup Rule" inside the new
	// sub-policy; it is managed here, embedded in the sub-policy resource (the
	// same relationship as a socket site's default interface / native range).
	var cleanupSchema resource.SchemaResponse
	ifrule.IfwSchema(ctx, resource.SchemaRequest{}, &cleanupSchema)
	cleanupAttr := cleanupSchema.Schema.Attributes["rule"].(schema.SingleNestedAttribute)
	cleanupAttr.Description = "The sub-policy's cleanup rule — auto-created by the API with the sub-policy " +
		"and managed as part of this resource. Omit to keep the API defaults (ALLOW, tracking disabled). " +
		"Uses the same parameters as a cato_if_rule rule. The discovered cleanup rule ID is preserved across external reorders."
	cleanupAttr.Required = false
	cleanupAttr.Optional = true
	cleanupAttr.Computed = true
	cleanupAttr.PlanModifiers = []planmodifier.Object{
		objectplanmodifier.UseStateForUnknown(),
	}
	cleanupAttr.Attributes["name"] = schema.StringAttribute{
		Description: "Cleanup rule name. Defaults to the API-generated \"<sub-policy name> - Cleanup Rule\"; " +
			"set to rename the cleanup rule.",
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	// Within an Optional+Computed object the framework cannot mark individual
	// children Required, so relax the rule schema's Required children to
	// Optional+Computed.
	for _, key := range []string{"description", "enabled", "action"} {
		if attr, ok := cleanupAttr.Attributes[key].(schema.StringAttribute); ok {
			attr.Required = false
			attr.Optional = true
			attr.Computed = true
			attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.UseStateForUnknown())
			cleanupAttr.Attributes[key] = attr
		}
		if attr, ok := cleanupAttr.Attributes[key].(schema.BoolAttribute); ok {
			attr.Required = false
			attr.Optional = true
			attr.Computed = true
			cleanupAttr.Attributes[key] = attr
		}
	}
	cleanupTracking := cleanupAttr.Attributes["tracking"].(schema.SingleNestedAttribute)
	cleanupTracking.Description = "Tracking for the cleanup rule. Optional — the API default is disabled."
	cleanupTracking.Required = false
	cleanupTracking.Optional = true
	cleanupTracking.Computed = true
	cleanupTracking.PlanModifiers = []planmodifier.Object{
		objectplanmodifier.UseStateForUnknown(),
	}
	cleanupAttr.Attributes["tracking"] = cleanupTracking

	resp.Schema = schema.Schema{
		Description: "The `cato_if_sub_policy` resource manages an Internet Firewall sub-policy " +
			"(a nested policy scoped by a SUB_POLICY_SCOPE rule) together with the cleanup rule the API " +
			"auto-creates inside it. The Cato API has no updateSubPolicy " +
			"mutation, so changing `name` or `description` forces resource replacement. Documentation for the " +
			"underlying API can be found at mutation.policy.internetFirewall.addSubPolicy().",
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
			"cleanup_rule_id": schema.StringAttribute{
				Description: "ID of the cleanup rule the API auto-created inside the sub-policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at":           ifSubPolicyAtAttribute(),
			"scope":        scopeAttr,
			"cleanup_rule": cleanupAttr,
		},
	}
}

// ifSubPolicyAtAttribute defines the position of the sub-policy scope within the
// main policy. Changes move the scope rule in place. The API does not return
// relative placement, so Read preserves the last applied value.
func ifSubPolicyAtAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Position of the sub-policy scope within the Internet Firewall policy. " +
			"Changing position or ref moves the scope rule in place without replacing the sub-policy or its child rules. " +
			"Refresh preserves the last applied placement and does not detect external reordering. " +
			"When cato_bulk_if_move_rule owns ordering, keep at unchanged to avoid conflicting moves.",
		Required: true,
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
	}
}
