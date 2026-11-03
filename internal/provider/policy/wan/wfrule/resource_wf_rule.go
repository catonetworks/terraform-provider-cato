package wfrule

import (
	"context"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &WanFwRuleResource{}
	_ resource.ResourceWithConfigure   = &WanFwRuleResource{}
	_ resource.ResourceWithImportState = &WanFwRuleResource{}
)

func NewWanFwRuleResource() resource.Resource {
	return &WanFwRuleResource{}
}

type WanFwRuleResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *WanFwRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wf_rule"
}

func (r *WanFwRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *WanFwRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Accept either "<rule-id>" (main-policy rule) or "<sub-policy-id>/<rule-id>"
	// (sub-policy rule).
	if subID, ruleID, ok := strings.Cut(req.ID, "/"); ok {
		if subID == "" || ruleID == "" {
			resp.Diagnostics.AddError(
				"Invalid import ID",
				"expected \"<rule-id>\" or \"<sub-policy-id>/<rule-id>\"",
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sub_policy_id"), subID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule").AtName("id"), ruleID)...)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

// adoptExistingWanRule resolves a ruleNameAlreadyExists collision on Create by
// locating the live rule with the configured name in the same policy scope
// (main policy vs. the plan's sub-policy; SUB_POLICY_SCOPE rows excluded) and
// updating it to match the configuration. It returns the adopted rule id, or a
// diagnostic when the rule cannot be found or lives in a different scope
// (updateRule cannot move rules between scopes).
//
//nolint:gocyclo,funlen
func (r *WanFwRuleResource) adoptExistingWanRule(
	ctx context.Context,
	plan Rule,
	input HydrateWanAPITypes,
) (string, diag.Diagnostic) {
	ruleName := ""
	if input.Create.Rule != nil {
		ruleName = input.Create.Rule.Name
	}
	planSubPolicyID := ""
	if !plan.SubPolicyID.IsNull() && !plan.SubPolicyID.IsUnknown() {
		planSubPolicyID = plan.SubPolicyID.ValueString()
	}

	var body *cato_go_sdk.Policy
	readErr := sharedpolicy.WithTransientReadRetry(ctx, "PolicyWanFirewall adoption lookup", func() error {
		var callErr error
		body, callErr = r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
		return callErr
	})
	if readErr != nil {
		return "", diag.NewErrorDiagnostic("Catov2 API PolicyWanFirewall error", readErr.Error())
	}

	adoptedID := ""
	for _, item := range body.GetPolicy().GetWanFirewall().GetPolicy().GetRules() {
		if rt := item.GetRuleType(); rt != nil && *rt == cato_models.PolicyRuleTypeEnumSubPolicyScope {
			continue
		}
		if item.GetRule().GetName() != ruleName {
			continue
		}
		itemSubPolicyID := ""
		if sp := item.GetSubPolicy(); sp != nil {
			itemSubPolicyID = sp.GetID()
		}
		if itemSubPolicyID != planSubPolicyID {
			return "", diag.NewErrorDiagnostic(
				"API Error Creating Resource",
				fmt.Sprintf("ruleNameAlreadyExists : rule %q already exists in a different policy scope "+
					"(existing sub-policy id %q, configured sub-policy id %q). Rule names are unique across the "+
					"whole WAN Firewall policy; rename the rule or remove the conflicting one.",
					ruleName, itemSubPolicyID, planSubPolicyID),
			)
		}
		adoptedID = item.GetRule().GetID()
		break
	}
	if adoptedID == "" {
		return "", diag.NewErrorDiagnostic(
			"API Error Creating Resource",
			fmt.Sprintf("ruleNameAlreadyExists : the API reports rule %q already exists, but it could not be "+
				"located in the policy for adoption. Retry the apply; if it persists, check for an unpublished "+
				"revision holding the rule.", ruleName),
		)
	}

	input.Update.ID = adoptedID
	updateResp, err := r.resourceAPI().PolicyWanFirewallUpdateRule(ctx, input.Update, r.client.AccountId)
	if err != nil {
		return "", diag.NewErrorDiagnostic("Catov2 API PolicyWanFirewallUpdateRule error", err.Error())
	}
	if updateResp.Policy.WanFirewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		msgs := ""
		for _, e := range updateResp.Policy.WanFirewall.UpdateRule.GetErrors() {
			if msgs != "" {
				msgs += "; "
			}
			code, message := "", ""
			if e.ErrorCode != nil {
				code = *e.ErrorCode
			}
			if e.ErrorMessage != nil {
				message = *e.ErrorMessage
			}
			msgs += fmt.Sprintf("%s : %s", code, message)
		}
		return "", diag.NewErrorDiagnostic("API Error Adopting Existing Rule", msgs)
	}
	return adoptedID, nil
}

func (r *WanFwRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Rule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := HydrateWanRuleAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// When the rule is owned by a sub-policy, anchor its position before the
	// sub-policy cleanup rule so the API places it inside the sub-policy.
	if !plan.SubPolicyID.IsNull() && !plan.SubPolicyID.IsUnknown() {
		subID := plan.SubPolicyID.ValueString()
		anchorBody, err := r.resourceAPI().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
			return
		}
		convertCreateValues(&input, subID, anchorBody, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_input.create", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.Create),
	})

	createRuleResponse, err := r.resourceAPI().PolicyWanFirewallAddRule(ctx, input.Create, r.client.AccountId)

	tflog.Warn(ctx, "TFLOG_WARN_WAN_createRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(createRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallAddRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	effectiveRuleID := ""
	if createRuleResponse.Policy.WanFirewall.AddRule.Status != cato_models.PolicyMutationStatusSuccess {
		nameCollision := false
		for _, item := range createRuleResponse.Policy.WanFirewall.AddRule.GetErrors() {
			if item.ErrorCode != nil && *item.ErrorCode == "ruleNameAlreadyExists" {
				nameCollision = true
				continue
			}
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		if !nameCollision || resp.Diagnostics.HasError() {
			return
		}
		// The rule already exists live — typically orphaned by an interrupted
		// apply (created + published, then the read-back failed before state was
		// written). Adopt it: resolve by name in the same policy scope, update
		// it to match the configuration, and continue with its id so Update and
		// Delete work natively from state.
		adoptedID, adoptDiag := r.adoptExistingWanRule(ctx, plan, input)
		if adoptDiag != nil {
			resp.Diagnostics.Append(adoptDiag)
			return
		}
		effectiveRuleID = adoptedID
		ruleName := ""
		if input.Create.Rule != nil {
			ruleName = input.Create.Rule.Name
		}
		resp.Diagnostics.AddWarning(
			"Adopted Existing Rule",
			fmt.Sprintf("Rule %q (id %s) already existed in the tenant (likely orphaned by an interrupted apply); "+
				"it was updated to match the configuration and is now managed by this resource. "+
				"Its current policy position is unchanged (ordering is managed by cato_bulk_wf_move_rule).",
				ruleName, adoptedID),
		)
	} else {
		effectiveRuleID = createRuleResponse.GetPolicy().GetWanFirewall().GetAddRule().Rule.GetRule().ID
	}

	// publishing new rule
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Sleep to allow the rule to propagate with active_period, bug in create/read
	// time.Sleep(30 * time.Second)

	// Read rule and hydrate response to state. The mutation already succeeded
	// and published at this point — retry transient backend failures instead of
	// failing the operation, which would desync state from the live rule (a
	// failed Create leaves the rule orphaned and the next apply collides with
	// ruleNameAlreadyExists).
	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	var body *cato_go_sdk.Policy
	readErr := sharedpolicy.WithTransientReadRetry(ctx, "PolicyWanFirewall post-mutation read", func() error {
		var callErr error
		body, callErr = r.resourceAPI().PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
		return callErr
	})
	if readErr != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			readErr.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().WanFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == effectiveRuleID {
			currentRule = ruleListItem.GetRule()
			resp.State.SetAttribute(
				ctx,
				path.Root("rule").AtName("id"),
				ruleListItem.GetRule().ID)
			break
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_WAN_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})
	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := HydrateWanRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	populateWanRuleID(currentRule, &ruleInputRead)
	tflog.Info(ctx, "ruleInputRead - "+fmt.Sprintf("%v", ruleInputRead))

	// Handle exceptions correlation manually to preserve plan structure
	ruleObject := convertCreateInput4(ctx, plan, &diags, &ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Rule = ruleObject

	// Set the complete plan to state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
}

func (r *WanFwRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Rule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	body, err := r.resourceAPI().PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			err.Error(),
		)
		return
	}

	// retrieve rule ID
	// Handle case where state might be incomplete (e.g., during import)
	var ruleID string
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		rule := convertReadValues(ctx, state, &diags)
		if diags.HasError() {
			// If conversion fails, try to extract just the ID from the state
			tflog.Debug(ctx, "Failed to convert full state to struct, attempting to extract ID only")
			// Clear the diagnostics as we'll handle this differently
			resp.Diagnostics = resp.Diagnostics[:0]

			// Try to extract ID directly from the state
			ruleAttrs := state.Rule.Attributes()
			if idValue, ok := ruleAttrs["id"]; ok && !idValue.IsNull() {
				ruleID = idValue.(types.String).ValueString()
			} else {
				resp.Diagnostics.AddError(
					"Unable to extract rule ID from state",
					"Could not find rule ID in state during import",
				)
				return
			}
		} else {
			ruleID = rule.ID.ValueString()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid state",
			"Rule state is null or unknown",
		)
		return
	}

	ruleList := body.GetPolicy().WanFirewall.Policy.GetRules()
	ruleExist := false
	currentRule := &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	currentSubPolicyID := types.StringNull()
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == ruleID {
			convertReadValues2(&ruleExist, &currentRule, &currentSubPolicyID, ruleListItem)

			// Need to refresh STATE
			resp.State.SetAttribute(
				ctx,
				path.Root("rule").AtName("id"),
				ruleListItem.GetRule().ID)
		}
	}

	// remove resource if it doesn't exist anymore
	if !ruleExist {
		tflog.Warn(ctx, "wan firewall rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_readRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	ruleInput, hydrateDiags := HydrateWanRuleState(ctx, state, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)

	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)

	// Reflect the sub-policy that currently owns the rule so config drift forces
	// replacement (sub_policy_id is immutable).
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sub_policy_id"), currentSubPolicyID)...)

	// Check if position is set in state, if not default to LAST_IN_POLICY
	// Hard coding LAST_IN_POLICY position as the API does not return any value and
	// hardcoding position supports the use case of bulk rule import/export
	// getting around state changes for the position field
	curAtObj, diagstmp := convertReadState2(ctx, state, &diags, &resp.Diagnostics)
	resp.Diagnostics.Append(diagstmp...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.SetAttribute(ctx, path.Root("at"), curAtObj)
	resp.Diagnostics.Append(diags...)
}

//nolint:funlen
func (r *WanFwRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Rule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := HydrateWanRuleAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input for moving rule
	inputMoveRule, ruleInput := convertUpdateInput(ctx, plan, &diags, &resp.Diagnostics)
	resp.Diagnostics.Append(diags...)

	// settings other rule attributes
	inputMoveRule.ID = *ruleInput.ID.ValueStringPointer()
	input.Update.ID = *ruleInput.ID.ValueStringPointer()

	// move rule
	moveRule, err := r.resourceAPI().PolicyWanFirewallMoveRule(ctx, inputMoveRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallMoveRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveRule.Policy.WanFirewall.MoveRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveRule.Policy.WanFirewall.MoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Rule Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_input.update", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.Update),
	})

	// updating rule
	updateRuleResponse, err := r.resourceAPI().PolicyWanFirewallUpdateRule(ctx, input.Update, r.client.AccountId)
	tflog.Warn(ctx, "TFLOG_WARN_WAN_updateRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(updateRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallUpdateRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateRuleResponse.Policy.WanFirewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateRuleResponse.Policy.WanFirewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// publishing new rule
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	wanFWQueryResponse, err := r.resourceAPI().PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
	tflog.Debug(ctx, "wanFWQueryResponse", map[string]interface{}{
		"wanFWQueryResponse": utils.InterfaceToJSONString(wanFWQueryResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			err.Error(),
		)
		return
	}

	ruleList := wanFWQueryResponse.GetPolicy().WanFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == updateRuleResponse.GetPolicy().GetWanFirewall().GetUpdateRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_WAN_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := HydrateWanRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)

	ruleObject := convertUpdateInput4(ctx, plan, &diags, currentRule, &ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Assign ruleObject to state
	plan.Rule = ruleObject

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *WanFwRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Rule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve rule ID
	rule := convertDeleteInput(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeRule := convertDeleteInput2(rule)
	tflog.Debug(ctx, "wan_fw_policy delete", map[string]interface{}{
		"input": utils.InterfaceToJSONString(removeRule),
	})

	_, err := r.resourceAPI().PolicyWanFirewallRemoveRule(ctx, removeRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
