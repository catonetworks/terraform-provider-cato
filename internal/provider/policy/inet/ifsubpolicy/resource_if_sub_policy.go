package ifsubpolicy

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	ifrule "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifrule"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &ifSubPolicyResource{}
	_ resource.ResourceWithConfigure   = &ifSubPolicyResource{}
	_ resource.ResourceWithImportState = &ifSubPolicyResource{}
)

func NewIfSubPolicyResource() resource.Resource {
	return &ifSubPolicyResource{}
}

type ifSubPolicyResource struct {
	client        *catoclient.Service
	subPolyClient InternetFirewallSubPolicyClient

	apiClient ResourceClient
}

func (r *ifSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_sub_policy"
}

func (r *ifSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *ifSubPolicyResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ifSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InternetFirewallSubPolicy
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
	scopeRule := ifrule.InternetFirewallRule{Rule: createScope, At: plan.At}
	hydrated, diags := ifrule.HydrateIfwRuleAPI(ctx, scopeRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Snapshot existing sub-policy ids to identify the new one after creation.
	before, err := r.resourceAPI().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	existing, addInput := convertCreateInput2(plan, hydrated, before)

	addResp, err := r.resourceAPI().PolicyInternetFirewallAddSubPolicy(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		addInput,
		r.client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallAddSubPolicy error", err.Error())
		return
	}
	addPayload := addResp.GetPolicy().GetInternetFirewall().GetAddSubPolicy()
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
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallPublishPolicyRevision error", err.Error())
		return
	}

	// Locate the new sub-policy + its scope rule.
	after, err := r.resourceAPI().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	scopeObj := convertCreateInput3(ctx, &plan, &diags, scopeRule, existing, after, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ifSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InternetFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}

	info := ifwSubPolicyInfo(body, state.ID.ValueString())
	if info == nil {
		tflog.Warn(ctx, "internet firewall sub-policy not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}
	scope := ifwScopeRule(body, state.ID.ValueString())
	if scope == nil {
		tflog.Warn(ctx, "internet firewall sub-policy scope rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	scopeObj, diags := convertReadState(ctx, &state, info, scope)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ifSubPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InternetFirewallSubPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state InternetFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeRuleID := state.ScopeRuleID.ValueString()
	plan.ID = state.ID
	plan.ScopeRuleID = state.ScopeRuleID

	// name/description are RequiresReplace, so only the scope can change here.
	scopeRule := ifrule.InternetFirewallRule{Rule: plan.Scope, At: plan.At}
	hydrated, diags := ifrule.HydrateIfwRuleAPI(ctx, scopeRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	hydrated.Update.ID = scopeRuleID
	// The API synchronizes scope name/description back to sub-policy metadata.
	hydrated.Update.Rule.Name = state.Name.ValueStringPointer()
	hydrated.Update.Rule.Description = state.Description.ValueStringPointer()

	updateResp, err := r.resourceAPI().PolicyInternetFirewallUpdateRule(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		hydrated.Update,
		r.client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallUpdateRule error", err.Error())
		return
	}
	if updateResp.Policy.InternetFirewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, e := range updateResp.Policy.InternetFirewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Sub-Policy Scope",
				fmt.Sprintf("%s : %s", derefStr(e.ErrorCode), derefStr(e.ErrorMessage)),
			)
		}
		return
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallPublishPolicyRevision error", err.Error())
		return
	}

	// Re-read scope for consistent state.
	body, err := r.resourceAPI().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
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

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ifSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InternetFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeInput := convertDeleteInput(state)
	removeResp, err := r.resourceAPI().PolicyInternetFirewallRemoveSubPolicy(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		removeInput,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Delete.PolicyInternetFirewallRemoveSubPolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeResp),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallRemoveSubPolicy error", err.Error())
		return
	}
	removePayload := removeResp.GetPolicy().GetInternetFirewall().GetRemoveSubPolicy()
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
		resp.Diagnostics.AddError("Catov2 API Delete/PolicyInternetFirewallPublishPolicyRevision error", err.Error())
		return
	}
}

// subPolicyScopeRuleType is the ruleType value the API assigns to the scope rule
// that a sub-policy is created with.
const subPolicyScopeRuleType = cato_models.PolicyRuleTypeEnumSubPolicyScope
