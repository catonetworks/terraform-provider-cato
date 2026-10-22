package lfsubpolicy

import (
	"context"
	"errors"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &lfSubPolicyResource{}
	_ resource.ResourceWithConfigure   = &lfSubPolicyResource{}
	_ resource.ResourceWithImportState = &lfSubPolicyResource{}

	ErrLanRuleNotFound = errors.New("lan rule not found")
)

// NewLfSubPolicyResource creates a LAN Firewall sub-policy resource instance.
func NewLfSubPolicyResource() resource.Resource {
	return &lfSubPolicyResource{}
}

type lfSubPolicyResource struct {
	client          *catoclient.Service
	subPolicyClient SocketLanSubPolicyClient

	apiClient ResourceClient
}

// Metadata sets the Terraform resource type name for LAN Firewall sub-policies.
func (r *lfSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lf_sub_policy"
}

// Configure stores the configured Cato API client for resource operations.
func (r *lfSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
	r.subPolicyClient = r.client.Catov2
}

// ImportState imports a LAN Firewall sub-policy by its ID.
func (r *lfSubPolicyResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create adds a LAN Firewall sub-policy and hydrates its Terraform state from the API.
func (r *lfSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LanFirewallSubPolicy
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := convertCreateInput(
		ctx,
		r,
		plan,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.resourceAPI().PolicySocketLanAddSubPolicy(ctx, input, r.client.AccountId)
	subpol := result.GetPolicy().GetSocketLan().GetAddSubPolicy()
	if utils.CheckAPIErrors(err, subpol.GetErrors(),
		fmt.Sprintf("failed to add LAN sub-policy '%s'", plan.Name), &resp.Diagnostics) {
		return
	}
	policyID := convertCreateInput2(&plan, subpol, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// publish the changes
	r.publish(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Hydrate state from API
	hydratedState, notFound := r.hydrateLfSubPolicy(ctx, policyID, &plan, &resp.Diagnostics)
	if notFound {
		resp.Diagnostics.AddError("failed to create sub policy", "sub-policy not found in API response")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
}

// Update modifies the scope rule of a LAN Firewall sub-policy and refreshes its state.
func (r *lfSubPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LanFirewallSubPolicy
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LanFirewallSubPolicy
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !preserveLfSubPolicyUpdateState(ctx, plan, state, resp) {
		return
	}
	tfScope := convertUpdateInput(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	input := socketLanUpdateRuleInputFromTerraform(ctx, r, &resp.Diagnostics, plan, tfScope)

	if utils.HasValue(plan.At) && !plan.At.Equal(state.At) {
		r.move(ctx, plan.ScopeRuleID.ValueString(), plan.At, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	result, err := r.resourceAPI().PolicySocketLanUpdateRule(ctx, nil, input, r.client.AccountId)
	subpol := result.GetPolicy().GetSocketLan().GetUpdateRule()
	if utils.CheckAPIErrors(err, subpol.GetErrors(),
		fmt.Sprintf("failed to update LAN sub-policy '%s'", plan.Name), &resp.Diagnostics) {
		return
	}
	if checkPolicyMutationStatus(subpol.GetStatus(),
		fmt.Sprintf("failed to update LAN sub-policy '%s'", plan.Name), &resp.Diagnostics) {
		return
	}

	// publish the changes
	r.publish(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, notFound := r.hydrateLfSubPolicy(ctx, plan.ID.ValueString(), &plan, &resp.Diagnostics)
	if notFound {
		resp.Diagnostics.AddError("failed to update sub policy", "sub-policy not found in API response")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
}

// Delete removes a LAN Firewall sub-policy through the Cato API.
func (r *lfSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LanFirewallSubPolicy
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := convertDeleteInput(state)

	// Call Cato API to delete the subpolicy
	res, err := r.resourceAPI().PolicySocketLanRemoveSubPolicy(ctx, nil, input, r.client.AccountId)
	remove := res.GetPolicy().GetSocketLan().GetRemoveSubPolicy()
	if utils.CheckAPIErrors(err, remove.GetErrors(),
		fmt.Sprintf("failed to delete lan sub-policy '%s'", state.Name.ValueString()), &resp.Diagnostics) {
		return
	}
	if checkPolicyMutationStatus(remove.GetStatus(),
		fmt.Sprintf("failed to delete lan sub-policy '%s'", state.Name.ValueString()), &resp.Diagnostics) {
		return
	}

	// publish the changes
	r.publish(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes a LAN Firewall sub-policy from the Cato API and removes missing resources from state.
func (r *lfSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LanFirewallSubPolicy
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, notFound := r.hydrateLfSubPolicy(ctx, state.ID.ValueString(), &state, &resp.Diagnostics)
	if notFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// hydrateLfSubPolicy fetches the current state of a lan firewall sub-policy from the API
func (r *lfSubPolicyResource) hydrateLfSubPolicy(ctx context.Context, subPolicyID string,
	planOrState *LanFirewallSubPolicy, diags *diag.Diagnostics,
) (newState *LanFirewallSubPolicy, notFound bool) {
	// Call Cato API to get the policy
	result, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("failed to hydrate sub-policy", err.Error())
		return nil, false
	}

	var state *LanFirewallSubPolicy

	// Map API response to LanFirewallSubPolicy
	policy := result.GetPolicy().GetSocketLan().GetPolicy()
	for _, polRule := range policy.Rules {
		if polRule.GetSubPolicy().GetID() != subPolicyID {
			continue
		}
		if rType := polRule.GetRuleType(); rType == nil || *rType != cato_models.PolicyRuleTypeEnumSubPolicyScope {
			continue
		}
		convertHydrateLfSubPolicyValues(
			ctx,
			r,
			subPolicyID,
			planOrState,
			diags,
			&state,
			polRule,
		)
		break
	}

	if state == nil {
		return nil, true
	}
	if diags.HasError() {
		return nil, false
	}

	return state, false
}

func (r *lfSubPolicyResource) move(
	ctx context.Context,
	scopeRuleID string,
	at types.Object,
	diags *diag.Diagnostics,
) {
	const summary = "failed to move LAN sub-policy"
	input := convertMoveValues(
		ctx,
		r,
		scopeRuleID,
		at,
		diags,
	)
	if diags.HasError() {
		return
	}

	result, err := r.resourceAPI().PolicySocketLanMoveRule(ctx, input, r.client.AccountId)
	move := result.GetPolicy().GetSocketLan().GetMoveRule()
	if utils.CheckAPIErrors(err, move.GetErrors(), summary, diags) {
		return
	}
	checkPolicyMutationStatus(move.GetStatus(), summary, diags)
}

func preserveLfSubPolicyUpdateState(
	ctx context.Context,
	plan LanFirewallSubPolicy,
	state LanFirewallSubPolicy,
	resp *resource.UpdateResponse,
) bool {
	if utils.HasValue(state.At) && !utils.HasValue(plan.At) {
		resp.Diagnostics.AddAttributeError(
			path.Root("at"),
			"LAN sub-policy position cannot be removed",
			"The at attribute resolved to null or unknown after previously being configured.",
		)
		return false
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	return !resp.Diagnostics.HasError()
}
