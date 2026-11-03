package wfsubpolicy

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	wfrule "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &wfSubPolicyResource{}
	_ resource.ResourceWithConfigure   = &wfSubPolicyResource{}
	_ resource.ResourceWithImportState = &wfSubPolicyResource{}
)

func NewWfSubPolicyResource() resource.Resource {
	return &wfSubPolicyResource{}
}

type wfSubPolicyResource struct {
	client        *catoclient.Service
	subPolyClient WanFirewallSubPolicyClient

	apiClient ResourceClient
}

func (r *wfSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wf_sub_policy"
}

func (r *wfSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *wfSubPolicyResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *wfSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WanFirewallSubPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API requires a normal action while creating a scope rule, then returns
	// the API-owned SUB_POLICY action. Keep that bootstrap value out of config.
	createScope, diags := convertCreateInput(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reuse the rule hydrator to build the scope rule data + position.
	scopeRule := wfrule.Rule{Rule: createScope, At: plan.At}
	hydrated, diags := wfrule.HydrateWanRuleAPI(ctx, scopeRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Snapshot existing sub-policy ids to identify the new one after creation.
	before, err := r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}
	existing, addInput := convertCreateInput2(plan, hydrated, before)

	addResp, err := r.resourceAPI().PolicyWanFirewallAddSubPolicy(
		ctx,
		addInput,
		r.client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallAddSubPolicy error", err.Error())
		return
	}
	addPayload := addResp.GetPolicy().GetWanFirewall().GetAddSubPolicy()
	if addPayload.GetStatus() == nil || *addPayload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range addPayload.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Sub-Policy",
				fmt.Sprintf("%s : %s", derefStr(e.ErrorCode), derefStr(e.ErrorMessage)),
			)
		}
		return
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallPublishPolicyRevision error", err.Error())
		return
	}

	// Locate the new sub-policy + its scope rule.
	after, err := r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}
	scopeObj := wanFirewallSubPolicyStateFromPolicy(ctx, &plan, &diags, scopeRule, existing, after, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Scope = scopeObj

	// addSubPolicy auto-created the sub-policy's cleanup rule (its only rule at
	// this point). Capture it, apply any configured cleanup settings, and manage
	// it as part of this resource — like a socket site's default interface.
	cleanup := wanCleanupRule(after, plan.ID.ValueString())
	if cleanup == nil {
		resp.Diagnostics.AddError(
			"Cleanup Rule Not Found",
			fmt.Sprintf("No API-created cleanup rule found for sub-policy %s.", plan.ID.ValueString()),
		)
		return
	}
	if !plan.CleanupRule.IsNull() && !plan.CleanupRule.IsUnknown() {
		cleanupRule := wfrule.Rule{Rule: plan.CleanupRule, At: plan.At}
		hydratedCleanup, cuDiags := wfrule.HydrateWanRuleAPI(ctx, cleanupRule)
		resp.Diagnostics.Append(cuDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		hydratedCleanup.Update.ID = cleanup.GetID()
		// Keep the API-generated name unless the config customizes it.
		if !cleanupNameConfigured(plan.CleanupRule) {
			cleanupName := cleanup.GetName()
			hydratedCleanup.Update.Rule.Name = &cleanupName
		}
		updateResp, err := r.resourceAPI().PolicyWanFirewallUpdateRule(
			ctx,
			hydratedCleanup.Update,
			r.client.AccountId,
		)
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallUpdateRule error", err.Error())
			return
		}
		if updateResp.Policy.WanFirewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
			for _, e := range updateResp.Policy.WanFirewall.UpdateRule.GetErrors() {
				resp.Diagnostics.AddError(
					"API Error Updating Sub-Policy Cleanup Rule",
					cleanupUpdateErrorDetail(derefStr(e.ErrorCode), derefStr(e.ErrorMessage),
						cleanupConfiguredName(plan.CleanupRule), cleanup.GetID()),
				)
			}
			return
		}
		if err := r.publish(ctx); err != nil {
			resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallPublishPolicyRevision error", err.Error())
			return
		}
		after, err = r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
			return
		}
		cleanup = wanCleanupRule(after, plan.ID.ValueString())
		if cleanup == nil {
			resp.Diagnostics.AddError("Cleanup Rule Not Found", "Sub-policy cleanup rule not found after update.")
			return
		}
	}
	cleanupObj, cuDiags := hydrateCleanupState(ctx, &plan, cleanup)
	resp.Diagnostics.Append(cuDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.CleanupRule = cleanupObj

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *wfSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WanFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}

	info := wanSubPolicyInfo(body, state.ID.ValueString())
	if info == nil {
		tflog.Warn(ctx, "wan firewall sub-policy not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}
	scope := wanScopeRule(body, state.ID.ValueString())
	if scope == nil {
		tflog.Warn(ctx, "wan firewall sub-policy scope rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	scopeObj, diags := convertReadState(ctx, &state, info, scope)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Scope = scopeObj

	// Cleanup rule: matched by stored id, falling back to the sub-policy's last
	// rule — the fallback is what populates it on `terraform import`.
	cleanup := wanCleanupRule(body, state.ID.ValueString())
	if cleanup == nil {
		tflog.Warn(ctx, "wan firewall sub-policy cleanup rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}
	cleanupObj, cuDiags := hydrateCleanupState(ctx, &state, cleanup)
	resp.Diagnostics.Append(cuDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.CleanupRule = cleanupObj

	// `at` is input-only (the API never returns a position): preserve the stored
	// value, defaulting a null one (fresh import) to LAST_IN_POLICY — the same
	// normalization the cato_wf_rule Read applies.
	if state.At.IsNull() || state.At.IsUnknown() {
		atObj, atDiags := types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue(string(cato_models.PolicySectionPositionEnumLastInPolicy)),
			"ref":      types.StringNull(),
		})
		resp.Diagnostics.Append(atDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.At = atObj
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wfSubPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WanFirewallSubPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state WanFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeRuleID := state.ScopeRuleID.ValueString()
	plan.ID = state.ID
	plan.ScopeRuleID = state.ScopeRuleID
	plan.CleanupRuleID = state.CleanupRuleID

	// name/description are RequiresReplace, so only the scope can change here.
	scopeRule := wfrule.Rule{Rule: plan.Scope, At: plan.At}
	hydrated, diags := wfrule.HydrateWanRuleAPI(ctx, scopeRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	hydrated.Update.ID = scopeRuleID
	// The API synchronizes scope name/description back to sub-policy metadata.
	hydrated.Update.Rule.Name = state.Name.ValueStringPointer()
	hydrated.Update.Rule.Description = state.Description.ValueStringPointer()

	updateResp, err := r.resourceAPI().PolicyWanFirewallUpdateRule(
		ctx,
		hydrated.Update,
		r.client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallUpdateRule error", err.Error())
		return
	}
	if updateResp.Policy.WanFirewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, e := range updateResp.Policy.WanFirewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Sub-Policy Scope",
				fmt.Sprintf("%s : %s", derefStr(e.ErrorCode), derefStr(e.ErrorMessage)),
			)
		}
		return
	}

	// Apply cleanup rule changes in the same revision, before the single publish.
	if !plan.CleanupRule.IsNull() && !plan.CleanupRule.IsUnknown() && !plan.CleanupRule.Equal(state.CleanupRule) {
		cleanupID := state.CleanupRuleID.ValueString()
		if cleanupID == "" {
			// Legacy state (pre-cleanup_rule): resolve via the last-rule lookup.
			body0, err := r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
			if err != nil {
				resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
				return
			}
			if cr := wanCleanupRule(body0, plan.ID.ValueString()); cr != nil {
				cleanupID = cr.GetID()
			}
		}
		if cleanupID == "" {
			resp.Diagnostics.AddError("Cleanup Rule Not Found", "Sub-policy cleanup rule could not be resolved for update.")
			return
		}
		cleanupRule := wfrule.Rule{Rule: plan.CleanupRule, At: plan.At}
		hydratedCleanup, cuDiags := wfrule.HydrateWanRuleAPI(ctx, cleanupRule)
		resp.Diagnostics.Append(cuDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		hydratedCleanup.Update.ID = cleanupID
		cleanupUpdateResp, err := r.resourceAPI().PolicyWanFirewallUpdateRule(
			ctx,
			hydratedCleanup.Update,
			r.client.AccountId,
		)
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallUpdateRule error", err.Error())
			return
		}
		if cleanupUpdateResp.Policy.WanFirewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
			for _, e := range cleanupUpdateResp.Policy.WanFirewall.UpdateRule.GetErrors() {
				resp.Diagnostics.AddError(
					"API Error Updating Sub-Policy Cleanup Rule",
					cleanupUpdateErrorDetail(derefStr(e.ErrorCode), derefStr(e.ErrorMessage),
						cleanupConfiguredName(plan.CleanupRule), cleanupID),
				)
			}
			return
		}
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallPublishPolicyRevision error", err.Error())
		return
	}

	// Re-read scope for consistent state.
	body, err := r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}
	scopeObj := convertUpdateInput(ctx, plan, scopeRule, &diags, body, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Scope = scopeObj

	cleanup := wanCleanupRule(body, plan.ID.ValueString())
	if cleanup == nil {
		resp.Diagnostics.AddError("Cleanup Rule Not Found", "Sub-policy cleanup rule not found after update.")
		return
	}
	cleanupObj, cuDiags := hydrateCleanupState(ctx, &plan, cleanup)
	resp.Diagnostics.Append(cuDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.CleanupRule = cleanupObj

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *wfSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WanFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeInput := convertDeleteInput(state)
	removeResp, err := r.resourceAPI().PolicyWanFirewallRemoveSubPolicy(
		ctx,
		removeInput,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Delete.PolicyWanFirewallRemoveSubPolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeResp),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallRemoveSubPolicy error", err.Error())
		return
	}
	removePayload := removeResp.GetPolicy().GetWanFirewall().GetRemoveSubPolicy()
	if removePayload.GetStatus() != nil && *removePayload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range removePayload.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Removing Sub-Policy",
				fmt.Sprintf("%s : %s", derefStr(e.ErrorCode), derefStr(e.ErrorMessage)),
			)
		}
		return
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API Delete/PolicyWanFirewallPublishPolicyRevision error", err.Error())
		return
	}
}

// subPolicyScopeRuleType is the ruleType value the API assigns to the scope rule
// that a sub-policy is created with.
const subPolicyScopeRuleType = cato_models.PolicyRuleTypeEnumSubPolicyScope

func (r *wfSubPolicyResource) publish(ctx context.Context) error {
	_, err := r.resourceAPI().PolicyWanFirewallPublishPolicyRevision(
		ctx,
		&cato_models.PolicyPublishRevisionInput{},
		r.client.AccountId,
	)
	return err
}
