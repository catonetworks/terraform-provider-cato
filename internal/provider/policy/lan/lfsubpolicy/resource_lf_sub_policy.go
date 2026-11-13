package lfsubpolicy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vektah/gqlparser/v2/ast"

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
	policyID, recovered := r.addSubPolicy(ctx, input, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// publish the changes (the recovery path has already published)
	if !recovered {
		r.publish(ctx, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
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

	r.reconcileCleanup(ctx, policyID, plan.CleanupRule, hydratedState, true, &resp.Diagnostics)
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

	r.reconcileCleanup(ctx, plan.ID.ValueString(), plan.CleanupRule, hydratedState, true, &resp.Diagnostics)
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
	var removeDiags diag.Diagnostics
	failedRemove := utils.CheckAPIErrors(err, remove.GetErrors(),
		fmt.Sprintf("failed to delete lan sub-policy '%s'", state.Name.ValueString()), &removeDiags)
	if !failedRemove {
		failedRemove = checkPolicyMutationStatus(remove.GetStatus(),
			fmt.Sprintf("failed to delete lan sub-policy '%s'", state.Name.ValueString()), &removeDiags)
	}
	if failedRemove {
		// socketLan.removeSubPolicy can succeed while returning an unreadable
		// payload (the resolver nulls the non-nullable policy.access field,
		// nulling the whole socketLan payload — policy-platform
		// DOWNSTREAM_SERVICE_ERROR); a retried call then reports
		// subPolicyAlreadyRemoved. Confirm removal before failing the delete.
		alreadyRemoved := false
		for _, e := range remove.GetErrors() {
			if code := e.GetErrorCode(); code != nil && *code == "subPolicyAlreadyRemoved" {
				alreadyRemoved = true
			}
		}
		if !alreadyRemoved && (!unreadableSocketLanMutation(err, "removeSubPolicy") ||
			!r.confirmSubPolicyRemoved(ctx, state.ID.ValueString())) {
			resp.Diagnostics.Append(removeDiags...)
			return
		}
		resp.Diagnostics.AddWarning(
			"LAN sub-policy removal reported an unreadable response",
			fmt.Sprintf("policy.socketLan.removeSubPolicy returned an error payload for sub-policy '%s' "+
				"but the sub-policy was confirmed removed.", state.Name.ValueString()))
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

	// Preserve cleanup-rule identity across external reorders; use positional
	// discovery only for an initial read or import.
	r.reconcileCleanup(ctx, state.ID.ValueString(),
		types.ObjectNull(LanFirewallSubPolicyCleanupRuleTypes), hydratedState, false, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// subPolicyIDsByName snapshots the IDs of sub-policies already present, so a
// create whose response payload is unreadable can still identify the new one.
func (r *lfSubPolicyResource) subPolicyIDsByName(ctx context.Context, diags *diag.Diagnostics) map[string]bool {
	result, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("failed to list existing LAN sub-policies", err.Error())
		return nil
	}
	ids := make(map[string]bool)
	for _, polRule := range result.GetPolicy().GetSocketLan().GetPolicy().GetRules() {
		if sp := polRule.GetSubPolicy(); sp != nil {
			ids[sp.GetID()] = true
		}
	}
	return ids
}

// confirmSubPolicyRemoved publishes the pending revision and re-reads the
// policy to verify the sub-policy is gone, for removes whose response payload
// was unreadable despite the mutation succeeding.
func (r *lfSubPolicyResource) confirmSubPolicyRemoved(ctx context.Context, subPolicyID string) bool {
	var diags diag.Diagnostics
	r.publish(ctx, &diags)
	if diags.HasError() {
		return false
	}
	result, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		return false
	}
	for _, polRule := range result.GetPolicy().GetSocketLan().GetPolicy().GetRules() {
		if sp := polRule.GetSubPolicy(); sp != nil && sp.GetID() == subPolicyID {
			return false
		}
	}
	return true
}

// findCreatedSubPolicyID re-reads the policy and returns the ID of the
// sub-policy with the given name that was not present in preExisting.
func (r *lfSubPolicyResource) findCreatedSubPolicyID(
	ctx context.Context,
	name string,
	preExisting map[string]bool,
	diags *diag.Diagnostics,
) string {
	result, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("failed to re-read LAN policy while recovering sub-policy create", err.Error())
		return ""
	}
	for _, polRule := range result.GetPolicy().GetSocketLan().GetPolicy().GetRules() {
		sp := polRule.GetSubPolicy()
		if sp == nil || preExisting[sp.GetID()] || sp.GetName() != name {
			continue
		}
		return sp.GetID()
	}
	return ""
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

// reconcileCleanup resolves the sub-policy's API-created cleanup rule (its
// LAST rule), optionally applies the configured cleanup_rule changes, and
// hydrates cleanup_rule/cleanup_rule_id on the state. Mirrors the
// cato_if_sub_policy / cato_wf_sub_policy cleanup embed.
func (r *lfSubPolicyResource) reconcileCleanup(
	ctx context.Context,
	subPolicyID string,
	desired types.Object,
	st *LanFirewallSubPolicy,
	apply bool,
	diags *diag.Diagnostics,
) {
	result, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("failed to resolve sub-policy cleanup rule", err.Error())
		return
	}
	policy := result.GetPolicy().GetSocketLan().GetPolicy()
	cleanup := lanCleanupRule(policy, subPolicyID, st.CleanupRuleID.ValueString())
	if cleanup == nil {
		// Only an error when the config actually manages the cleanup rule;
		// otherwise surface it as informational and leave the fields null.
		if st.CleanupRuleID.ValueString() != "" || (apply && !desired.IsNull() && !desired.IsUnknown()) {
			diags.AddError("Cleanup Rule Not Found",
				"cleanup_rule is configured but no API-created cleanup rule was found for sub-policy "+subPolicyID+".")
			return
		}
		diags.AddWarning("Cleanup Rule Not Found",
			"No API-created cleanup rule found for sub-policy "+subPolicyID+"; cleanup_rule left unset.")
		st.CleanupRuleID = types.StringNull()
		st.CleanupRule = types.ObjectNull(LanFirewallSubPolicyCleanupRuleTypes)
		return
	}

	if apply {
		if input := lanCleanupUpdateInput(ctx, desired, cleanup.GetID(), diags); input != nil && !diags.HasError() {
			updResult, err := r.resourceAPI().PolicySocketLanUpdateRule(ctx, nil, *input, r.client.AccountId)
			upd := updResult.GetPolicy().GetSocketLan().GetUpdateRule()
			if utils.CheckAPIErrors(err, upd.GetErrors(), "failed to update LAN sub-policy cleanup rule", diags) {
				return
			}
			if checkPolicyMutationStatus(upd.GetStatus(), "failed to update LAN sub-policy cleanup rule", diags) {
				return
			}
			r.publish(ctx, diags)
			if diags.HasError() {
				return
			}
			result, err = r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
			if err != nil {
				diags.AddError("failed to re-read sub-policy cleanup rule", err.Error())
				return
			}
			cleanup = lanCleanupRule(result.GetPolicy().GetSocketLan().GetPolicy(), subPolicyID, cleanup.GetID())
			if cleanup == nil {
				diags.AddError("Cleanup Rule Not Found", "Sub-policy cleanup rule not found after update.")
				return
			}
		}
	}

	st.CleanupRuleID = types.StringValue(cleanup.GetID())
	st.CleanupRule = lanCleanupObject(cleanup, diags)
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

// publish calls the API to publish the draft policy revision
func (r *lfSubPolicyResource) publish(ctx context.Context, diags *diag.Diagnostics) {
	const summary = "failed to publish LAN firewall policy"
	const notFound = "PolicyRevisionNotFound"
	result, err := r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, nil, r.client.AccountId)
	if err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	publish := result.GetPolicy().GetSocketLan().GetPublishPolicyRevision()
	apiErrors := publish.GetErrors()
	if len(apiErrors) > 0 {
		allNotFound := true
		for _, apiError := range apiErrors {
			if code := apiError.GetErrorCode(); code == nil || *code != notFound {
				allNotFound = false
				break
			}
		}
		if allNotFound {
			return
		}
	}
	if utils.CheckAPIErrors(nil, apiErrors, summary, diags) {
		return
	}
	checkPolicyMutationStatus(publish.GetStatus(), summary, diags)
}

func (r *lfSubPolicyResource) addSubPolicy(
	ctx context.Context,
	input cato_models.SocketLanAddSubPolicyInput,
	plan *LanFirewallSubPolicy,
	diags *diag.Diagnostics,
) (string, bool) {
	preExisting := r.subPolicyIDsByName(ctx, diags)
	if diags.HasError() {
		return "", false
	}
	result, err := r.resourceAPI().PolicySocketLanAddSubPolicy(ctx, input, r.client.AccountId)
	subpol := result.GetPolicy().GetSocketLan().GetAddSubPolicy()
	var addDiags diag.Diagnostics
	var policyID string
	if !utils.CheckAPIErrors(err, subpol.GetErrors(),
		fmt.Sprintf("failed to add LAN sub-policy '%s'", plan.Name), &addDiags) {
		policyID = convertCreateInput2(plan, subpol, &addDiags)
	}
	recovered := false
	if policyID == "" {
		if !unreadableSocketLanMutation(err, "addSubPolicy") {
			diags.Append(addDiags...)
			return "", false
		}
		// socketLan.addSubPolicy can violate its own response contract: the
		// resolver returns null for the non-nullable policy.rules field, which
		// nulls the whole socketLan payload (policy-platform
		// DOWNSTREAM_SERVICE_ERROR) even though the sub-policy was created.
		// Recover by publishing, re-reading the policy, and locating the new
		// sub-policy by name among those that did not exist before the add.
		var pubDiags diag.Diagnostics
		r.publish(ctx, &pubDiags)
		if !pubDiags.HasError() {
			policyID = r.findCreatedSubPolicyID(ctx, plan.Name.ValueString(), preExisting, &pubDiags)
		}
		if policyID == "" || pubDiags.HasError() {
			diags.Append(addDiags...)
			diags.Append(pubDiags...)
			return "", false
		}
		recovered = true
		plan.ID = types.StringValue(policyID)
		diags.AddWarning(
			"Recovered LAN sub-policy from unreadable addSubPolicy response",
			fmt.Sprintf("policy.socketLan.addSubPolicy returned an unreadable payload for sub-policy '%s'; "+
				"the created sub-policy (ID %s) was recovered by re-reading the policy.", plan.Name, policyID))
	}
	return policyID, recovered
}

// Recovery may publish the pending revision. Only the known structured null
// payload failure proves recovery is appropriate; ordinary failures must stop.
func unreadableSocketLanMutation(err error, operation string) bool {
	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response.NetworkError != nil || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
		return false
	}
	for _, item := range *response.GqlErrors {
		if item == nil || len(item.Path) < 2 || item.Path[0] != ast.PathName("policy") || item.Path[1] != ast.PathName("socketLan") {
			return false
		}
		message := strings.ToLower(item.Message)
		if !strings.Contains(message, "non null") && !strings.Contains(message, "non-null") {
			return false
		}
		if (len(item.Path) < 3 || item.Path[2] != ast.PathName(operation)) &&
			!strings.Contains(item.Message, "/policy/socketLan/"+operation+"/") {
			return false
		}
	}
	return true
}
