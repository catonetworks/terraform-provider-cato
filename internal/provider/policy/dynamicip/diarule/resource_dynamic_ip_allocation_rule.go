package diarule

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	dynamicip "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/dynamicip"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &dynamicIpAllocationRuleResource{}
	_ resource.ResourceWithConfigure   = &dynamicIpAllocationRuleResource{}
	_ resource.ResourceWithImportState = &dynamicIpAllocationRuleResource{}
)

func NewDynamicIpAllocationRuleResource() resource.Resource {
	return &dynamicIpAllocationRuleResource{}
}

type dynamicIpAllocationRuleResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *dynamicIpAllocationRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_dynamic_ip_allocation_rule"
}

func (r *dynamicIpAllocationRuleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *dynamicIpAllocationRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

//nolint:gocyclo,funlen
func (r *dynamicIpAllocationRuleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan DynamicIpAllocationRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateDynamicIpAllocationAddRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyDynamicIPAllocationAddRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	res, err := r.resourceAPI().PolicyDynamicIPAllocationAddRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyDynamicIPAllocationAddRule error", err.Error())
		return
	}
	add := res.GetPolicy().GetDynamicIPAllocation().GetAddRule()
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
		resp.Diagnostics.AddError("dynamic IP allocation addRule failed", st)
		return
	}

	resp.Diagnostics.Append(dynamicip.PublishDynamicIpAllocationPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicyDynamicIPAllocationPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyDynamicIPAllocationPolicy error", err.Error())
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
	var cur *cato_go_sdk.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetDynamicIPAllocation().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == newID {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after create failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateDynamicIpAllocationRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)

	// Merge plan-only attributes the API does not echo back.
	var planRule DynamicIpAllocationRuleRulePlan
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

func (r *dynamicIpAllocationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DynamicIpAllocationRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicyDynamicIPAllocationPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyDynamicIPAllocationPolicy error", err.Error())
		return
	}

	rulePlan := DynamicIpAllocationRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rulePlan, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetDynamicIPAllocation().GetPolicy()
	if pol == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var cur *cato_go_sdk.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule
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

	hydrated, hdiags := hydrateDynamicIpAllocationRuleStateFromClient(ctx, cur)
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
func (r *dynamicIpAllocationRuleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan DynamicIpAllocationRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rule DynamicIpAllocationRuleRulePlan
	move, atAsOpts := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	resp.Diagnostics.Append(plan.Rule.As(ctx, &rule, atAsOpts)...)
	move.ID = rule.ID.ValueString()
	if resp.Diagnostics.HasError() {
		return
	}

	moveRes, err := r.resourceAPI().PolicyDynamicIPAllocationMoveRule(ctx, move, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyDynamicIPAllocationMoveRule error", err.Error())
		return
	}
	mv := moveRes.GetPolicy().GetDynamicIPAllocation().GetMoveRule()
	if mv.GetStatus() == nil || *mv.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range mv.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error moving rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	upd, diags := hydrateDynamicIpAllocationUpdateRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updRes, err := r.resourceAPI().PolicyDynamicIPAllocationUpdateRule(ctx, upd, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyDynamicIPAllocationUpdateRule error", err.Error())
		return
	}
	ur := updRes.GetPolicy().GetDynamicIPAllocation().GetUpdateRule()
	if ur.GetStatus() == nil || *ur.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range ur.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error updating rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	resp.Diagnostics.Append(dynamicip.PublishDynamicIpAllocationPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().PolicyDynamicIPAllocationPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyDynamicIPAllocationPolicy error", err.Error())
		return
	}

	var cur *cato_go_sdk.PolicyDynamicIpAllocationPolicy_Policy_DynamicIPAllocation_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetDynamicIPAllocation().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == rule.ID.ValueString() {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after update failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateDynamicIpAllocationRuleStateFromClient(ctx, cur)
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

func (r *dynamicIpAllocationRuleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state DynamicIpAllocationRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule := DynamicIpAllocationRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	rm := convertDeleteInput(rule)
	if _, err := r.resourceAPI().PolicyDynamicIPAllocationRemoveRule(ctx, rm, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyDynamicIPAllocationRemoveRule error", err.Error())
		return
	}
	resp.Diagnostics.Append(dynamicip.PublishDynamicIpAllocationPolicyRevision(ctx, r.resourceAPI(), r.client.AccountId)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
