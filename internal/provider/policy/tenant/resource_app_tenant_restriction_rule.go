//nolint:lll
package tenant

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy_rule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &appTenantRestrictionRuleResource{}
	_ resource.ResourceWithConfigure   = &appTenantRestrictionRuleResource{}
	_ resource.ResourceWithImportState = &appTenantRestrictionRuleResource{}
)

func NewAppTenantRestrictionRuleResource() resource.Resource {
	return &appTenantRestrictionRuleResource{}
}

type appTenantRestrictionRuleResource struct {
	client *catoclient.Service
}

func (r *appTenantRestrictionRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_tenant_restriction_rule"
}

func (r *appTenantRestrictionRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *appTenantRestrictionRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

//nolint:gocyclo,funlen
func (r *appTenantRestrictionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppTenantRestrictionRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateAppTenantRestrictionAddRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyAppTenantRestrictionAddRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	res, err := r.client.Catov2.PolicyAppTenantRestrictionAddRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionAddRule error", err.Error())
		return
	}
	add := res.GetPolicy().GetAppTenantRestriction().GetAddRule()
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
		resp.Diagnostics.AddError("app tenant restriction addRule failed", st)
		return
	}

	resp.Diagnostics.Append(publishAppTenantRestrictionPolicyRevision(ctx, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Catov2.AppTenantRestrictionPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppTenantRestrictionPolicy error", err.Error())
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
	var cur *cato_go_sdk.AppTenantRestrictionPolicy_Policy_AppTenantRestriction_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetAppTenantRestriction().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == newID {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after create failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateAppTenantRestrictionRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)

	// Preserve plan headers — the API never returns sensitive header values.
	var planRule AppTenantRestrictionRuleRulePlan
	resp.Diagnostics.Append(plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if !planRule.Headers.IsNull() && !planRule.Headers.IsUnknown() {
		rulePlan.Headers = planRule.Headers
	}

	ruleObj, odiags := types.ObjectValueFrom(ctx, AppTenantRestrictionRuleRuleAttrTypes, rulePlan)
	resp.Diagnostics.Append(odiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appTenantRestrictionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppTenantRestrictionRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Catov2.AppTenantRestrictionPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppTenantRestrictionPolicy error", err.Error())
		return
	}

	rulePlan := AppTenantRestrictionRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rulePlan, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetAppTenantRestriction().GetPolicy()
	if pol == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var cur *cato_go_sdk.AppTenantRestrictionPolicy_Policy_AppTenantRestriction_Policy_Rules_Rule
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

	hydrated, hdiags := hydrateAppTenantRestrictionRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)

	// Preserve headers from state — the API never returns sensitive header values.
	if !rulePlan.Headers.IsNull() && !rulePlan.Headers.IsUnknown() {
		hydrated.Headers = rulePlan.Headers
	}

	ruleObj, odiags := types.ObjectValueFrom(ctx, AppTenantRestrictionRuleRuleAttrTypes, hydrated)
	resp.Diagnostics.Append(odiags...)

	// Preserve the 'at' block from state — the API does not return placement
	// information, so we keep whatever was previously configured. On import the
	// state has no 'at', so we fall back to the safe default LAST_IN_POLICY.
	if state.At.IsNull() || state.At.IsUnknown() {
		atObj, d := types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		})
		resp.Diagnostics.Append(d...)
		state.At = atObj
	}

	state.ID = types.StringValue(cur.GetID())
	state.Rule = ruleObj
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

//nolint:gocyclo,funlen
func (r *appTenantRestrictionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AppTenantRestrictionRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	move := cato_models.PolicyMoveRuleInput{}
	atAsOpts := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	if !plan.At.IsNull() {
		pos := policyrule.Position{}
		resp.Diagnostics.Append(plan.At.As(ctx, &pos, atAsOpts)...)
		move.To = &cato_models.PolicyRulePositionInput{
			Position: (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer()),
			Ref:      pos.Ref.ValueStringPointer(),
		}
	}
	rule := AppTenantRestrictionRuleRulePlan{}
	resp.Diagnostics.Append(plan.Rule.As(ctx, &rule, atAsOpts)...)
	move.ID = rule.ID.ValueString()
	if resp.Diagnostics.HasError() {
		return
	}

	moveRes, err := r.client.Catov2.PolicyAppTenantRestrictionMoveRule(ctx, move, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionMoveRule error", err.Error())
		return
	}
	mv := moveRes.GetPolicy().GetAppTenantRestriction().GetMoveRule()
	if mv.GetStatus() == nil || *mv.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range mv.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error moving rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	upd, diags := hydrateAppTenantRestrictionUpdateRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updRes, err := r.client.Catov2.PolicyAppTenantRestrictionUpdateRule(ctx, upd, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionUpdateRule error", err.Error())
		return
	}
	ur := updRes.GetPolicy().GetAppTenantRestriction().GetUpdateRule()
	if ur.GetStatus() == nil || *ur.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range ur.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error updating rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	resp.Diagnostics.Append(publishAppTenantRestrictionPolicyRevision(ctx, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Catov2.AppTenantRestrictionPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppTenantRestrictionPolicy error", err.Error())
		return
	}

	var cur *cato_go_sdk.AppTenantRestrictionPolicy_Policy_AppTenantRestriction_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetAppTenantRestriction().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == rule.ID.ValueString() {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after update failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateAppTenantRestrictionRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)

	// Preserve plan headers — the API never returns sensitive header values.
	if !rule.Headers.IsNull() && !rule.Headers.IsUnknown() {
		rulePlan.Headers = rule.Headers
	}

	ruleObj, odiags := types.ObjectValueFrom(ctx, AppTenantRestrictionRuleRuleAttrTypes, rulePlan)
	resp.Diagnostics.Append(odiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pos policyrule.Position
	resp.Diagnostics.Append(plan.At.As(ctx, &pos, atAsOpts)...)
	if resp.Diagnostics.HasError() {
		return
	}
	atObj, adiags := types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
		"position": pos.Position,
		"ref":      pos.Ref,
	})
	resp.Diagnostics.Append(adiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	st := AppTenantRestrictionRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, st)...)
}

func (r *appTenantRestrictionRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppTenantRestrictionRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule := AppTenantRestrictionRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	rm := cato_models.AppTenantRestrictionRemoveRuleInput{ID: rule.ID.ValueString()}
	if _, err := r.client.Catov2.PolicyAppTenantRestrictionRemoveRule(ctx, rm, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionRemoveRule error", err.Error())
		return
	}
	resp.Diagnostics.Append(publishAppTenantRestrictionPolicyRevision(ctx, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
