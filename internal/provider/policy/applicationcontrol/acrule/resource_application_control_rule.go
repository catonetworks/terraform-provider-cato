package acrule

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	applicationcontrol "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/applicationcontrol"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &applicationControlRuleResource{}
	_ resource.ResourceWithConfigure   = &applicationControlRuleResource{}
	_ resource.ResourceWithImportState = &applicationControlRuleResource{}
)

func NewApplicationControlRuleResource() resource.Resource {
	return &applicationControlRuleResource{}
}

type applicationControlRuleResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *applicationControlRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_application_control_rule"
}

func (r *applicationControlRuleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *applicationControlRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

//nolint:gocyclo,funlen
func (r *applicationControlRuleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ApplicationControlRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateApplicationControlAddRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyApplicationControlAddRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	res, err := r.resourceAPI().PolicyApplicationControlAddRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlAddRule error", err.Error())
		return
	}
	add := res.GetPolicy().GetApplicationControl().GetAddRule()
	for _, e := range add.GetErrors() {
		if e != nil && e.GetErrorCode() != nil {
			resp.Diagnostics.AddError("API error: "+*e.GetErrorCode(), *e.GetErrorMessage())
			return
		}
	}
	if add.GetStatus() == nil || *add.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		st := ""
		if add.GetStatus() != nil {
			st = string(*add.GetStatus())
		}
		resp.Diagnostics.AddError("Application Control addRule failed", st)
		return
	}

	resp.Diagnostics.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().ApplicationControlPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API ApplicationControlPolicy error", err.Error())
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

	var cur *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetApplicationControl().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == newID {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after create failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateApplicationControlRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)
	ruleObj, odiags := applicationControlRuleRuleObjectFromPlan(rulePlan)
	resp.Diagnostics.Append(odiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pos policyrule.Position
	resp.Diagnostics.Append(plan.At.As(ctx, &pos, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}
	atObj, adiags := convertCreateInput(pos)
	resp.Diagnostics.Append(adiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	st := convertCreateInput2(newID, ruleObj, atObj)
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *applicationControlRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationControlRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().ApplicationControlPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API ApplicationControlPolicy error", err.Error())
		return
	}

	rulePlan := ApplicationControlRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rulePlan, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetApplicationControl().GetPolicy()
	if pol == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var cur *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule
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

	hydrated, hdiags := hydrateApplicationControlRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)
	ruleObj, odiags := applicationControlRuleRuleObjectFromPlan(hydrated)
	resp.Diagnostics.Append(odiags...)

	// Preserve 'at' from state — the API does not return placement info.
	// Default to LAST_IN_POLICY only when state has no 'at' (e.g. after import).
	convertReadState(&state, cur, ruleObj, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

//nolint:gocyclo,funlen
func (r *applicationControlRuleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ApplicationControlRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rule ApplicationControlRuleRulePlan
	move, atAsOpts := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	resp.Diagnostics.Append(plan.Rule.As(ctx, &rule, atAsOpts)...)
	move.ID = rule.ID.ValueString()
	if resp.Diagnostics.HasError() {
		return
	}

	moveRes, err := r.resourceAPI().PolicyApplicationControlMoveRule(ctx, move, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlMoveRule error", err.Error())
		return
	}
	mv := moveRes.GetPolicy().GetApplicationControl().GetMoveRule()
	if mv.GetStatus() == nil || *mv.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range mv.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error moving rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	upd, diags := hydrateApplicationControlUpdateRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updRes, err := r.resourceAPI().PolicyApplicationControlUpdateRule(ctx, upd, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlUpdateRule error", err.Error())
		return
	}
	ur := updRes.GetPolicy().GetApplicationControl().GetUpdateRule()
	if ur.GetStatus() == nil || *ur.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range ur.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error updating rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	resp.Diagnostics.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().ApplicationControlPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API ApplicationControlPolicy error", err.Error())
		return
	}

	var cur *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetApplicationControl().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == rule.ID.ValueString() {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after update failed", "rule not found in policy response")
		return
	}

	hydrated, hdiags := hydrateApplicationControlRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)
	ruleObj, odiags := applicationControlRuleRuleObjectFromPlan(hydrated)
	resp.Diagnostics.Append(odiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pos policyrule.Position
	resp.Diagnostics.Append(plan.At.As(ctx, &pos, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}
	atObj, adiags := convertUpdateInput2(pos)
	resp.Diagnostics.Append(adiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	st := convertUpdateInput3(rule, ruleObj, atObj)
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *applicationControlRuleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ApplicationControlRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule := ApplicationControlRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	rm := convertDeleteInput(rule)
	if _, err := r.resourceAPI().PolicyApplicationControlRemoveRule(ctx, rm, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlRemoveRule error", err.Error())
		return
	}
	resp.Diagnostics.Append(applicationcontrol.PublishApplicationControlPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
