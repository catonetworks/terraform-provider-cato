package ifrule

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
	_ resource.Resource                = &InternetFwRuleResource{}
	_ resource.ResourceWithConfigure   = &InternetFwRuleResource{}
	_ resource.ResourceWithImportState = &InternetFwRuleResource{}
)

type InternetFwRuleResource struct {
	Client    *catoclient.Service
	IfwClient InternetFirewallPolicyClient

	apiClient ResourceClient
}

func NewInternetFwRuleResource() resource.Resource {
	return &InternetFwRuleResource{}
}

func (r *InternetFwRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_rule"
}

func (r *InternetFwRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.Client = req.ProviderData.(*catoclient.Service)
}

func (r *InternetFwRuleResource) ImportState(
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

// adoptExistingRule resolves a ruleNameAlreadyExists collision on Create by
// locating the live rule with the configured name in the same policy scope
// (main policy vs. the plan's sub-policy; SUB_POLICY_SCOPE rows excluded) and
// updating it to match the configuration. It returns the adopted rule id, or a
// diagnostic when the rule cannot be found or lives in a different scope
// (updateRule cannot move rules between scopes).
func (r *InternetFwRuleResource) adoptExistingRule(
	ctx context.Context,
	plan InternetFirewallRule,
	input HydrateIfwAPITypes,
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
	readErr := sharedpolicy.WithTransientReadRetry(ctx, "PolicyInternetFirewall adoption lookup", func() error {
		var callErr error
		body, callErr = r.resourceAPI().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.Client.AccountId)
		return callErr
	})
	if readErr != nil {
		return "", diag.NewErrorDiagnostic("Catov2 API PolicyInternetFirewall error", readErr.Error())
	}

	adoptedID := ""
	for _, item := range body.GetPolicy().GetInternetFirewall().GetPolicy().GetRules() {
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
					"whole Internet Firewall policy; rename the rule or remove the conflicting one.",
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
	updateResp, err := r.resourceAPI().PolicyInternetFirewallUpdateRule(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		input.Update,
		r.Client.AccountId,
	)
	if err != nil {
		return "", diag.NewErrorDiagnostic("Catov2 API PolicyInternetFirewallUpdateRule error", err.Error())
	}
	if updateResp.Policy.InternetFirewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		msgs := ""
		for _, e := range updateResp.Policy.InternetFirewall.UpdateRule.GetErrors() {
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

//nolint:gocyclo,funlen
func (r *InternetFwRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InternetFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := HydrateIfwRuleAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// When the rule is owned by a sub-policy, anchor its position before the
	// sub-policy cleanup rule so the API places it inside the sub-policy. The
	// sub-policy has no policy-level position of its own, so any policy-level
	// position from config would otherwise land the rule in the main policy.
	if !plan.SubPolicyID.IsNull() && !plan.SubPolicyID.IsUnknown() {
		subID := plan.SubPolicyID.ValueString()
		anchorBody, err := r.resourceAPI().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.Client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
			return
		}
		convertCreateValues(&input, subID, anchorBody, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW_input.create", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.Create),
	})

	// Creating new rule
	createRuleResponse, err := r.resourceAPI().PolicyInternetFirewallAddRule(ctx, input.Create, r.Client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallAddRule error",
			err.Error(),
		)
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW_createRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(createRuleResponse),
	})

	// check for errors
	effectiveRuleID := ""
	if createRuleResponse.Policy.InternetFirewall.AddRule.Status != cato_models.PolicyMutationStatusSuccess {
		nameCollision := false
		for _, item := range createRuleResponse.Policy.InternetFirewall.AddRule.GetErrors() {
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
		adoptedID, adoptDiag := r.adoptExistingRule(ctx, plan, input)
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
				"Its current policy position is unchanged (ordering is managed by cato_bulk_if_move_rule).",
				ruleName, adoptedID),
		)
	} else {
		effectiveRuleID = createRuleResponse.GetPolicy().GetInternetFirewall().GetAddRule().Rule.GetRule().ID
	}

	// Publishing new rule
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		publishDataIfEnabled,
		r.Client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state. The mutation already succeeded
	// and published at this point — retry transient backend failures instead of
	// failing the operation, which would desync state from the live rule (a
	// failed Create leaves the rule orphaned and the next apply collides with
	// ruleNameAlreadyExists).
	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	var body *cato_go_sdk.Policy
	readErr := sharedpolicy.WithTransientReadRetry(ctx, "PolicyInternetFirewall post-mutation read", func() error {
		var callErr error
		body, callErr = r.resourceAPI().PolicyInternetFirewall(ctx, queryIfwPolicy, r.Client.AccountId)
		return callErr
	})
	if readErr != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			readErr.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID (created or adopted)
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == effectiveRuleID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleObject := internetFirewallRuleObjectFromResponse(ctx, plan, &diags, currentRule)
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
	resp.Diagnostics.Append(diags...)
}

func (r *InternetFwRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InternetFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	body, err := r.resourceAPI().PolicyInternetFirewall(ctx, queryIfwPolicy, r.Client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			err.Error(),
		)
		return
	}

	// Retrieve rule ID
	rule := convertReadState2(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	ruleExist := false
	currentRule := &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	currentSubPolicyID := types.StringNull()
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == rule.ID.ValueString() {
			convertReadValues(&ruleExist, &currentRule, &currentSubPolicyID, ruleListItem)

			// Need to refresh STATE
			resp.State.SetAttribute(
				ctx,
				path.Root("rule").AtName("id"),
				ruleListItem.GetRule().ID)
		}
	}

	// remove resource if it doesn't exist anymore
	if !ruleExist {
		tflog.Warn(ctx, "internet firewall rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	ruleInput := HydrateIfwRuleState(ctx, state, currentRule)
	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	resp.Diagnostics.Append(diags...)

	// Reflect the sub-policy that currently owns the rule so config drift forces
	// replacement (sub_policy_id is immutable).
	diags = resp.State.SetAttribute(ctx, path.Root("sub_policy_id"), currentSubPolicyID)
	resp.Diagnostics.Append(diags...)

	// Check if position is set in state, if not default to LAST_IN_POLICY
	// Hard coding LAST_IN_POLICY position as the API does not return any value and
	// hardcoding position supports the use case of bulk rule import/export
	// getting around state changes for the position field
	curAtObj, diagstmp := convertReadState3(ctx, state, &diags, &resp.Diagnostics)
	diags = resp.State.SetAttribute(ctx, path.Root("at"), curAtObj)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(diagstmp...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:funlen
func (r *InternetFwRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InternetFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := HydrateIfwRuleAPI(ctx, plan)
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

	// Move rule
	moveRule, err := r.resourceAPI().PolicyInternetFirewallMoveRule(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		inputMoveRule,
		r.Client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallMoveRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveRule.Policy.InternetFirewall.MoveRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveRule.Policy.InternetFirewall.MoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Rule Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW_input.update", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.Update),
	})

	// Update new rule
	updateRuleResponse, err := r.resourceAPI().PolicyInternetFirewallUpdateRule(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		input.Update,
		r.Client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallUpdateRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateRuleResponse.Policy.InternetFirewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateRuleResponse.Policy.InternetFirewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Update Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publishing new rule
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		publishDataIfEnabled,
		r.Client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state. The mutation already succeeded
	// and published at this point — retry transient backend failures instead of
	// failing the operation, which would desync state from the live rule (a
	// failed Create leaves the rule orphaned and the next apply collides with
	// ruleNameAlreadyExists).
	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	var body *cato_go_sdk.Policy
	readErr := sharedpolicy.WithTransientReadRetry(ctx, "PolicyInternetFirewall post-mutation read", func() error {
		var callErr error
		body, callErr = r.resourceAPI().PolicyInternetFirewall(ctx, queryIfwPolicy, r.Client.AccountId)
		return callErr
	})
	if readErr != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			readErr.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == updateRuleResponse.GetPolicy().GetInternetFirewall().GetUpdateRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_IFW_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleObject := convertUpdateInput4(ctx, plan, &diags, currentRule)
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

func (r *InternetFwRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InternetFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve rule ID
	rule := convertDeleteInput(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeMutations, removeRule := convertDeleteInput2(rule)
	tflog.Debug(ctx, "internet_fw_policy delete", map[string]interface{}{
		"query": utils.InterfaceToJSONString(removeMutations),
		"input": utils.InterfaceToJSONString(removeRule),
	})

	_, err := r.resourceAPI().PolicyInternetFirewallRemoveRule(ctx, removeMutations, removeRule, r.Client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		publishDataIfEnabled,
		r.Client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
