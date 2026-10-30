package alwaysonrule

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &alwaysOnRuleResource{}
	_ resource.ResourceWithConfigure   = &alwaysOnRuleResource{}
	_ resource.ResourceWithImportState = &alwaysOnRuleResource{}
)

type alwaysOnRuleResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewAlwaysOnRuleResource returns the cato_always_on_rule resource.
func NewAlwaysOnRuleResource() resource.Resource {
	return &alwaysOnRuleResource{}
}

func (r *alwaysOnRuleResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return r.client.Catov2
}

func (r *alwaysOnRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_always_on_rule"
}

func (r *alwaysOnRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *alwaysOnRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *alwaysOnRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlwaysOnRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := buildAlwaysOnRuleInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil || r.getClient() == nil {
		resp.Diagnostics.AddError("Resource not configured", "The Cato client must be configured before creating an Always On rule.")
		return
	}

	errMsg := fmt.Sprintf("failed to create Always On rule %s", input.Rule.Name)
	result, err := r.getClient().PolicyZtnaAlwaysOnAddRule(ctx, r.client.AccountId, nil, input)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetAddRule()
	if utils.CheckAPIErrors(nil, payload.GetErrors(), errMsg, &resp.Diagnostics) {
		return
	}
	rule := payload.GetRule().GetRule()
	if rule.GetID() == "" {
		resp.Diagnostics.AddError(errMsg, "API did not return resource ID")
		return
	}
	if *payload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		resp.Diagnostics.AddError(errMsg, fmt.Sprintf("API returned mutation status %q", *payload.GetStatus()))
		return
	}

	plan.Rule = parseAlwaysOnRule(ctx, rule, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alwaysOnRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlwaysOnRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("rule").AtName("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !utils.HasValue(id) || id.ValueString() == "" {
		resp.Diagnostics.AddError("Invalid Always On rule ID", "Reading an Always On rule requires a known, non-empty rule ID in state.")
		return
	}
	if r.client == nil || r.getClient() == nil {
		resp.Diagnostics.AddError("Resource not configured", "The Cato client must be configured before reading an Always On rule.")
		return
	}

	errMsg := fmt.Sprintf("failed to read Always On rule %s", id.ValueString())
	result, err := r.getClient().PolicyZtnaAlwaysOnPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	rule, err := findAlwaysOnRule(result, id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Rule = parseAlwaysOnPolicyRule(ctx, rule, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *alwaysOnRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AlwaysOnRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildAlwaysOnUpdateRuleInput(ctx, plan, state, &resp.Diagnostics)
	move := buildAlwaysOnMoveRuleInput(ctx, plan, state, input.ID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil || r.getClient() == nil {
		resp.Diagnostics.AddError("Resource not configured", "The Cato client must be configured before updating an Always On rule.")
		return
	}
	plan.Rule = r.updateRule(ctx, input, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if move != nil {
		// Persist the completed update so a failed move can be retried with the previous placement input.
		resp.Diagnostics.Append(resp.State.Set(ctx, AlwaysOnRule{At: state.At, Rule: plan.Rule})...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Rule = r.moveRule(ctx, *move, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alwaysOnRuleResource) updateRule(ctx context.Context, input cato_models.ZtnaAlwaysOnUpdateRuleInput,
	diags *diag.Diagnostics,
) types.Object {
	errMsg := fmt.Sprintf("failed to update Always On rule %s", input.ID)
	result, err := r.getClient().PolicyZtnaAlwaysOnUpdateRule(ctx, r.client.AccountId, nil, input)
	if err != nil {
		diags.AddError(errMsg, err.Error())
		return types.ObjectNull(AlwaysOnRuleDataAttrTypes)
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetUpdateRule()
	if utils.CheckAPIErrors(nil, payload.GetErrors(), errMsg, diags) {
		return types.ObjectNull(AlwaysOnRuleDataAttrTypes)
	}
	if *payload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		diags.AddError(errMsg, fmt.Sprintf("API returned mutation status %q", *payload.GetStatus()))
		return types.ObjectNull(AlwaysOnRuleDataAttrTypes)
	}
	rule := payload.GetRule().GetRule()
	if rule.GetID() != input.ID {
		diags.AddError(errMsg, fmt.Sprintf("API returned resource ID %q; expected %q", rule.GetID(), input.ID))
		return types.ObjectNull(AlwaysOnRuleDataAttrTypes)
	}
	return parseUpdatedAlwaysOnRule(ctx, rule, diags)
}

func (r *alwaysOnRuleResource) moveRule(ctx context.Context, input cato_models.PolicyMoveRuleInput,
	diags *diag.Diagnostics,
) types.Object {
	errMsg := fmt.Sprintf("failed to move Always On rule %s", input.ID)
	result, err := r.getClient().PolicyZtnaAlwaysOnMoveRule(ctx, r.client.AccountId, nil, input)
	if err != nil {
		diags.AddError(errMsg, err.Error())
		return types.ObjectNull(AlwaysOnRuleDataAttrTypes)
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetMoveRule()
	if utils.CheckAPIErrors(nil, payload.GetErrors(), errMsg, diags) {
		return types.ObjectNull(AlwaysOnRuleDataAttrTypes)
	}
	if *payload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		diags.AddError(errMsg, fmt.Sprintf("API returned mutation status %q", *payload.GetStatus()))
		return types.ObjectNull(AlwaysOnRuleDataAttrTypes)
	}
	rule := payload.GetRule().GetRule()
	if rule.GetID() != input.ID {
		diags.AddError(errMsg, fmt.Sprintf("API returned resource ID %q; expected %q", rule.GetID(), input.ID))
		return types.ObjectNull(AlwaysOnRuleDataAttrTypes)
	}
	return parseMovedAlwaysOnRule(ctx, rule, diags)
}

func (r *alwaysOnRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlwaysOnRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildAlwaysOnRemoveRuleInput(ctx, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil || r.getClient() == nil {
		resp.Diagnostics.AddError("Resource not configured", "The Cato client must be configured before deleting an Always On rule.")
		return
	}
	errMsg := fmt.Sprintf("failed to delete Always On rule %s", input.ID)
	result, err := r.getClient().PolicyZtnaAlwaysOnRemoveRule(ctx, r.client.AccountId, nil, input)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetRemoveRule()
	status := *payload.GetStatus()
	alreadyRemoved := false
	apiErrors := payload.GetErrors()
	for i, apiError := range apiErrors {
		code := utils.GqlOptionalStr(apiError.GetErrorCode())
		if status == cato_models.PolicyMutationStatusFailure && (code == "ruleNotExist" || code == "RuleNotFound") {
			alreadyRemoved = true
			continue
		}
		utils.CheckAPIErrors(nil, apiErrors[i:i+1], errMsg, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	if status != cato_models.PolicyMutationStatusSuccess && !alreadyRemoved {
		resp.Diagnostics.AddError(errMsg, fmt.Sprintf("API returned mutation status %q", status))
		return
	}
	resp.State.RemoveResource(ctx)
}
