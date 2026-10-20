package wfrule

import (
	"context"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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

func (r *WanFwRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

//nolint:gocyclo,funlen
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
		anchorBody, err := r.client.Catov2.PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
			return
		}
		cleanupID := wan.SubPolicyCleanupRuleID(anchorBody, subID)
		if cleanupID == "" {
			resp.Diagnostics.AddError(
				"Sub-policy not found",
				fmt.Sprintf("could not locate cleanup rule for sub_policy_id %q; verify the sub-policy exists", subID),
			)
			return
		}
		beforeRule := cato_models.PolicyRulePositionEnumBeforeRule
		input.Create.At = &cato_models.PolicyRulePositionInput{
			Position: &beforeRule,
			Ref:      &cleanupID,
		}
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_input.create", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.Create),
	})

	createRuleResponse, err := r.client.Catov2.PolicyWanFirewallAddRule(ctx, input.Create, r.client.AccountId)

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
	if createRuleResponse.Policy.WanFirewall.AddRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range createRuleResponse.Policy.WanFirewall.AddRule.GetErrors() {
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
	_, err = r.client.Catov2.PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Sleep to allow the rule to propagate with active_period, bug in create/read
	// time.Sleep(30 * time.Second)

	// Read rule and hydrate response to state
	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	body, err := r.client.Catov2.PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().WanFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == createRuleResponse.GetPolicy().GetWanFirewall().GetAddRule().Rule.GetRule().ID {
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
	ruleInputRead.ID = types.StringValue(currentRule.ID)
	tflog.Info(ctx, "ruleInputRead - "+fmt.Sprintf("%v", ruleInputRead))

	// Handle exceptions correlation manually to preserve plan structure
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := PolicyPolicyWanFirewallPolicyRulesRule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := wan.CorrelateWanExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	ruleObject, diags := types.ObjectValueFrom(ctx, WanFirewallRuleRuleAttrTypes, ruleInputRead)
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

//nolint:gocyclo,funlen
func (r *WanFwRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Rule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	body, err := r.client.Catov2.PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
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
		rule := PolicyPolicyWanFirewallPolicyRulesRule{}
		diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
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
			ruleExist = true
			currentRule = ruleListItem.GetRule()
			if sp := ruleListItem.GetSubPolicy(); sp != nil && sp.GetID() != "" {
				currentSubPolicyID = types.StringValue(sp.GetID())
			}

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
	positionValue := string(cato_models.PolicyRulePositionEnumLastInPolicy)
	refValue := types.StringNull()

	if !state.At.IsNull() && !state.At.IsUnknown() {
		statePosInput := policyrule.Position{}
		diags = state.At.As(ctx, &statePosInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if !diags.HasError() && !statePosInput.Position.IsNull() && !statePosInput.Position.IsUnknown() {
			positionValue = statePosInput.Position.ValueString()
			if !statePosInput.Ref.IsNull() && !statePosInput.Ref.IsUnknown() {
				refValue = statePosInput.Ref
			}
		}
	}

	curAtObj, diagstmp := types.ObjectValue(
		policyrule.PositionAttrTypes,
		map[string]attr.Value{
			"position": types.StringValue(positionValue),
			"ref":      refValue,
		},
	)
	resp.Diagnostics.Append(diagstmp...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.SetAttribute(ctx, path.Root("at"), curAtObj)
	resp.Diagnostics.Append(diags...)
}

//nolint:gocyclo,funlen
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
	inputMoveRule := cato_models.PolicyMoveRuleInput{}

	// setting at (to move rule)
	if !plan.At.IsNull() {
		inputMoveRule.To = &cato_models.PolicyRulePositionInput{}
		positionInput := policyrule.Position{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputMoveRule.To.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		inputMoveRule.To.Ref = positionInput.Ref.ValueStringPointer()
	}

	ruleInput := PolicyPolicyWanFirewallPolicyRulesRule{}
	diags = plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	// settings other rule attributes
	inputMoveRule.ID = *ruleInput.ID.ValueStringPointer()
	input.Update.ID = *ruleInput.ID.ValueStringPointer()

	// move rule
	moveRule, err := r.client.Catov2.PolicyWanFirewallMoveRule(ctx, inputMoveRule, r.client.AccountId)
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
	updateRuleResponse, err := r.client.Catov2.PolicyWanFirewallUpdateRule(ctx, input.Update, r.client.AccountId)
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
	_, err = r.client.Catov2.PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	wanFWQueryResponse, err := r.client.Catov2.PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
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

	ruleInputRead.ID = types.StringValue(currentRule.ID)

	// Handle exceptions correlation manually to preserve plan structure
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := PolicyPolicyWanFirewallPolicyRulesRule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := wan.CorrelateWanExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	ruleObject, diags := types.ObjectValueFrom(ctx, WanFirewallRuleRuleAttrTypes, ruleInputRead)
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
	rule := PolicyPolicyWanFirewallPolicyRulesRule{}
	diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeRule := cato_models.WanFirewallRemoveRuleInput{
		ID: rule.ID.ValueString(),
	}
	tflog.Debug(ctx, "wan_fw_policy delete", map[string]interface{}{
		"input": utils.InterfaceToJSONString(removeRule),
	})

	_, err := r.client.Catov2.PolicyWanFirewallRemoveRule(ctx, removeRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.Catov2.PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
