package wfsubpolicy

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	wan "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan"
	wfrule "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
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

	scopeRule := wfrule.Rule{Rule: createScope, At: plan.At}
	hydrated, diags := wfrule.HydrateWanRuleAPI(ctx, scopeRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	before, err := r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}
	existing, addInput := convertCreateInput2(plan, hydrated, before)

	addResp, err := r.resourceAPI().PolicyWanFirewallAddSubPolicy(ctx, addInput, r.client.AccountId)
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

	after, err := r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}
	scope, scopeState := convertCreateInput3(ctx, &plan, &diags, scopeRule, existing, after, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
	scopeObj := convertCreateInput4(ctx, &diags, scope, &scopeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Scope = scopeObj

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

	info := wan.SubPolicyInfo(body, state.ID.ValueString())
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

	scopeState, diags := convertReadState(ctx, &state, info, scope)
	resp.Diagnostics.Append(diags...)
	scopeObj := convertReadState2(ctx, scope, &scopeState, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Scope = scopeObj

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

	updateResp, err := r.resourceAPI().PolicyWanFirewallUpdateRule(ctx, hydrated.Update, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallUpdateRule error", err.Error())
		return
	}
	if updateResp.Policy.WanFirewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, e := range updateResp.Policy.WanFirewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError("API Error Updating Sub-Policy Scope", fmt.Sprintf("%s : %s", derefStr(e.ErrorCode), derefStr(e.ErrorMessage)))
		}
		return
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallPublishPolicyRevision error", err.Error())
		return
	}

	body, err := r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}
	scope := wanScopeRule(body, plan.ID.ValueString())
	if scope == nil {
		resp.Diagnostics.AddError("Scope Rule Not Found", "Sub-policy scope rule not found after update.")
		return
	}
	scopeState, diags := wfrule.HydrateWanRuleState(ctx, scopeRule, scope)
	resp.Diagnostics.Append(diags...)
	scopeObj := convertUpdateInput(ctx, &diags, scope, &scopeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *wfSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WanFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeInput := convertDeleteInput(state)
	removeResp, err := r.resourceAPI().PolicyWanFirewallRemoveSubPolicy(ctx, removeInput, r.client.AccountId)
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
