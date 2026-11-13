package strule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	splittunnel "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/splittunnel"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &splitTunnelRuleResource{}
	_ resource.ResourceWithConfigure   = &splitTunnelRuleResource{}
	_ resource.ResourceWithImportState = &splitTunnelRuleResource{}
)

func NewSplitTunnelRuleResource() resource.Resource {
	return &splitTunnelRuleResource{}
}

type splitTunnelRuleResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *splitTunnelRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_split_tunnel_rule"
}

func (r *splitTunnelRuleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *splitTunnelRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

//nolint:gocyclo,funlen
func (r *splitTunnelRuleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SplitTunnelRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateSplitTunnelAddRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySplitTunnelAddRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	res, err := r.resourceAPI().PolicySplitTunnelAddRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelAddRule error", err.Error())
		return
	}
	add := res.GetPolicy().GetSplitTunnel().GetAddRule()
	for _, e := range add.GetErrors() {
		if e != nil && e.GetErrorCode() != nil {
			resp.Diagnostics.AddError("API error: "+*e.GetErrorCode(), *e.GetErrorMessage())
			return
		}
	}
	if err := sharedpolicy.MutationResult(add.GetStatus(), add.GetErrors()); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(splittunnel.PublishSplitTunnelPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySplitTunnelPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelPolicy error", err.Error())
		return
	}

	newID := ""
	if add.GetRule() != nil && add.GetRule().GetRule() != nil {
		newID = add.GetRule().GetRule().GetID()
	}
	if newID == "" {
		resp.Diagnostics.AddError("Read after create failed", "addRule response did not return rule id")
		return
	}
	var cur *cato_go_sdk.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetSplitTunnel().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == newID {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after create failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateSplitTunnelRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)

	// Merge plan-only attributes the API does not echo back.
	var planRule SplitTunnelRuleRulePlan
	resp.Diagnostics.Append(plan.Rule.As(
		ctx,
		&planRule,
		basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true},
	)...)
	ruleObj, odiags := convertCreateInput(ctx, &rulePlan, planRule)
	resp.Diagnostics.Append(odiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	convertCreateInput2(&plan, newID, ruleObj)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *splitTunnelRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SplitTunnelRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySplitTunnelPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelPolicy error", err.Error())
		return
	}

	rulePlan := SplitTunnelRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rulePlan, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetSplitTunnel().GetPolicy()
	if pol == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var cur *cato_go_sdk.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule
	for _, item := range pol.GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == rulePlan.ID.ValueString() {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	hydrated, hdiags := hydrateSplitTunnelRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)

	// Merge state-only attributes the API does not echo back.
	ruleObj, odiags := convertReadState(ctx, rulePlan, &hydrated)
	resp.Diagnostics.Append(odiags...)

	// Preserve the 'at' block from state — the API does not return placement
	// information, so we keep whatever was previously configured. On import the
	// state has no 'at', so we fall back to the safe default LAST_IN_POLICY.
	convertReadState2(&state, cur, ruleObj, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

//nolint:gocyclo,funlen
func (r *splitTunnelRuleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SplitTunnelRule
	var state SplitTunnelRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rule SplitTunnelRuleRulePlan
	move, atAsOpts := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	resp.Diagnostics.Append(plan.Rule.As(ctx, &rule, atAsOpts)...)
	move.ID = rule.ID.ValueString()
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.At.Equal(state.At) {
		moveRes, err := r.resourceAPI().PolicySplitTunnelMoveRule(ctx, move, r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError("Cato API PolicySplitTunnelMoveRule error", err.Error())
			return
		}
		mv := moveRes.GetPolicy().GetSplitTunnel().GetMoveRule()
		if err := sharedpolicy.MutationResult(mv.GetStatus(), mv.GetErrors()); err != nil {
			resp.Diagnostics.AddError("Policy mutation failed", err.Error())
			return
		}
	}

	upd, diags := hydrateSplitTunnelUpdateRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updRes, err := r.resourceAPI().PolicySplitTunnelUpdateRule(ctx, upd, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelUpdateRule error", err.Error())
		return
	}
	ur := updRes.GetPolicy().GetSplitTunnel().GetUpdateRule()
	if err := sharedpolicy.MutationResult(ur.GetStatus(), ur.GetErrors()); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(splittunnel.PublishSplitTunnelPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySplitTunnelPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelPolicy error", err.Error())
		return
	}

	var cur *cato_go_sdk.PolicySplitTunnelPolicy_Policy_SplitTunnel_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetSplitTunnel().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == rule.ID.ValueString() {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after update failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateSplitTunnelRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)

	// Merge plan-only attributes the API does not echo back.
	ruleObj, odiags := convertUpdateInput2(ctx, rule, &rulePlan)
	resp.Diagnostics.Append(odiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pos policyrule.Position
	resp.Diagnostics.Append(plan.At.As(ctx, &pos, atAsOpts)...)
	if resp.Diagnostics.HasError() {
		return
	}
	atObj, adiags := convertUpdateInput3(pos)
	resp.Diagnostics.Append(adiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	st := convertUpdateInput4(rule, ruleObj, atObj)
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *splitTunnelRuleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SplitTunnelRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule := SplitTunnelRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	rm := convertDeleteInput(rule)
	removed, err := r.resourceAPI().PolicySplitTunnelRemoveRule(ctx, rm, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySplitTunnelRemoveRule error", err.Error())
		return
	}
	payload := removed.GetPolicy().GetSplitTunnel().GetRemoveRule()
	if err := sharedpolicy.MutationResult(payload.GetStatus(), payload.GetErrors()); err != nil {
		resp.Diagnostics.AddError("Policy deletion failed", err.Error())
		return
	}
	resp.Diagnostics.Append(splittunnel.PublishSplitTunnelPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
