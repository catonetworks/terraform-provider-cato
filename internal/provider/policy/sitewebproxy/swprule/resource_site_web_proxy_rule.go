package swprule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	sitewebproxy "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/sitewebproxy"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &siteWebProxyRuleResource{}
	_ resource.ResourceWithConfigure   = &siteWebProxyRuleResource{}
	_ resource.ResourceWithImportState = &siteWebProxyRuleResource{}
)

func NewSiteWebProxyRuleResource() resource.Resource {
	return &siteWebProxyRuleResource{}
}

type siteWebProxyRuleResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *siteWebProxyRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_site_web_proxy_rule"
}

func (r *siteWebProxyRuleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *siteWebProxyRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

//nolint:gocyclo,funlen
func (r *siteWebProxyRuleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SiteWebProxyRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateSiteWebProxyAddRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySiteWebProxyAddRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	res, err := r.resourceAPI().PolicySiteWebProxyAddRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyAddRule error", err.Error())
		return
	}
	add := res.GetPolicy().GetSiteWebProxy().GetAddRule()
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

	resp.Diagnostics.Append(sitewebproxy.PublishSiteWebProxyPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySiteWebProxyPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyPolicy error", err.Error())
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
	var cur *cato_go_sdk.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetSiteWebProxy().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == newID {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after create failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateSiteWebProxyRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)

	// Merge plan-only attributes the API does not echo back.
	var planRule SiteWebProxyRuleRulePlan
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

func (r *siteWebProxyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SiteWebProxyRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySiteWebProxyPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyPolicy error", err.Error())
		return
	}

	rulePlan := SiteWebProxyRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rulePlan, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetSiteWebProxy().GetPolicy()
	if pol == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var cur *cato_go_sdk.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule
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

	hydrated, hdiags := hydrateSiteWebProxyRuleStateFromClient(ctx, cur)
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
func (r *siteWebProxyRuleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SiteWebProxyRule
	var state SiteWebProxyRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rule SiteWebProxyRuleRulePlan
	move, atAsOpts := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	resp.Diagnostics.Append(plan.Rule.As(ctx, &rule, atAsOpts)...)
	move.ID = rule.ID.ValueString()
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.At.Equal(state.At) {
		moveRes, err := r.resourceAPI().PolicySiteWebProxyMoveRule(ctx, move, r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError("Cato API PolicySiteWebProxyMoveRule error", err.Error())
			return
		}
		mv := moveRes.GetPolicy().GetSiteWebProxy().GetMoveRule()
		if err := sharedpolicy.MutationResult(mv.GetStatus(), mv.GetErrors()); err != nil {
			resp.Diagnostics.AddError("Policy mutation failed", err.Error())
			return
		}
	}

	upd, diags := hydrateSiteWebProxyUpdateRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updRes, err := r.resourceAPI().PolicySiteWebProxyUpdateRule(ctx, upd, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyUpdateRule error", err.Error())
		return
	}
	ur := updRes.GetPolicy().GetSiteWebProxy().GetUpdateRule()
	if err := sharedpolicy.MutationResult(ur.GetStatus(), ur.GetErrors()); err != nil {
		resp.Diagnostics.AddError("Policy mutation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(sitewebproxy.PublishSiteWebProxyPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicySiteWebProxyPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyPolicy error", err.Error())
		return
	}

	var cur *cato_go_sdk.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetSiteWebProxy().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == rule.ID.ValueString() {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after update failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateSiteWebProxyRuleStateFromClient(ctx, cur)
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

func (r *siteWebProxyRuleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SiteWebProxyRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule := SiteWebProxyRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	rm := convertDeleteInput(rule)
	removed, err := r.resourceAPI().PolicySiteWebProxyRemoveRule(ctx, rm, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicySiteWebProxyRemoveRule error", err.Error())
		return
	}
	payload := removed.GetPolicy().GetSiteWebProxy().GetRemoveRule()
	if err := sharedpolicy.MutationResult(payload.GetStatus(), payload.GetErrors()); err != nil {
		resp.Diagnostics.AddError("Policy deletion failed", err.Error())
		return
	}
	resp.Diagnostics.Append(sitewebproxy.PublishSiteWebProxyPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
