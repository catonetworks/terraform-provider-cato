package wnwrule

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

	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &wanNetworkRuleResource{}
	_ resource.ResourceWithConfigure   = &wanNetworkRuleResource{}
	_ resource.ResourceWithImportState = &wanNetworkRuleResource{}
)

func NewWanNetworkRuleResource() resource.Resource {
	return &wanNetworkRuleResource{}
}

type wanNetworkRuleResource struct {
	client *catoclient.Service
}

func (r *wanNetworkRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wnw_rule"
}

func (r *wanNetworkRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *wanNetworkRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

//nolint:gocyclo,funlen
func (r *wanNetworkRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Rule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateWanNetworkRuleAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_input.create", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.create),
	})

	createRuleResponse, err := r.client.Catov2.PolicyWanNetworkAddRule(ctx, input.create, r.client.AccountId)

	tflog.Warn(ctx, "TFLOG_WARN_WAN_createRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(createRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkAddRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if createRuleResponse.Policy.WanNetwork.AddRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range createRuleResponse.Policy.WanNetwork.AddRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publish policy revision (align with WAN FW behavior)
	tflog.Info(ctx, "publishing new rule")
	_, err = r.client.Catov2.PolicyWanNetworkPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.client.Catov2.WanNetworkPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API WanNetworkPolicy error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().WanNetwork.Policy.Rules
	currentRule := &cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule{}
	// Get current rule from response by ID
	if createRuleResponse.Policy.WanNetwork.AddRule.Rule != nil {
		createdRuleID := createRuleResponse.Policy.WanNetwork.AddRule.Rule.GetRule().ID
		for _, ruleListItem := range ruleList {
			if ruleListItem.GetRule().ID == createdRuleID {
				currentRule = ruleListItem.GetRule()
				resp.State.SetAttribute(
					ctx,
					path.Root("rule").AtName("id"),
					ruleListItem.GetRule().ID)
				break
			}
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_WAN_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})
	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := hydrateWanNetworkRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	ruleInputRead.ID = types.StringValue(currentRule.ID)
	tflog.Info(ctx, "ruleInputRead - "+fmt.Sprintf("%v", ruleInputRead))

	// Handle exceptions correlation manually to preserve plan structure
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := wan.CorrelateWanExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	ruleObject, diags := types.ObjectValueFrom(ctx, WanNetworkRuleRuleAttrTypes, ruleInputRead)
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
}

//nolint:gocyclo,funlen
func (r *wanNetworkRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Rule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query WAN Network policy
	body, err := r.client.Catov2.WanNetworkPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API WanNetworkPolicy error",
			err.Error(),
		)
		return
	}

	// Extract rule ID from state
	var ruleID string
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		rule := PolicyPolicyWanNetworkPolicyRulesRule{}
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

	// Find the rule in the policy
	ruleList := body.GetPolicy().WanNetwork.Policy.Rules
	ruleExist := false
	var currentRule *cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule
	for _, ruleListItem := range ruleList {
		rule := ruleListItem.GetRule()
		if rule.ID == ruleID {
			ruleExist = true
			currentRule = rule
			break
		}
	}

	// Remove resource if it doesn't exist anymore
	if !ruleExist {
		tflog.Warn(ctx, "WAN Network rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "WAN Network rule found, hydrating state", map[string]interface{}{
		"rule_id": ruleID,
	})

	// Hydrate the state from the API response
	ruleInput, hydrateDiags := hydrateWanNetworkRuleState(ctx, state, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle exceptions correlation manually to preserve state structure
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		stateRule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = state.Rule.As(ctx, &stateRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !stateRule.Exceptions.IsNull() && !stateRule.Exceptions.IsUnknown() {
			// Correlate exceptions between state and hydrated response
			correlatedExceptions := wan.CorrelateWanExceptions(ctx, stateRule.Exceptions, ruleInput.Exceptions)
			if correlatedExceptions != nil {
				ruleInput.Exceptions = *correlatedExceptions
			}
		}
	}

	// Set the rule attribute in state
	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hard code LAST_IN_POLICY position as the API does not return any value
	// This supports the use case of bulk rule import/export
	positionValue := "LAST_IN_POLICY"
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
	diags = resp.State.SetAttribute(ctx, path.Root("at"), curAtObj)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(diagstmp...)
}

//nolint:gocyclo,funlen
func (r *wanNetworkRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Rule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateWanNetworkRuleAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract rule ID from plan
	var ruleID string
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		rule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ruleID = rule.ID.ValueString()
	}

	// Prepare update input with rule ID
	updateInput := cato_models.WanNetworkUpdateRuleInput{
		ID:   ruleID,
		Rule: input.update.Rule,
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_NW_input.update", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(updateInput),
	})

	updateRuleResponse, err := r.client.Catov2.PolicyWanNetworkUpdateRule(ctx, updateInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkUpdateRule error",
			err.Error(),
		)
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_NW_input.update.response", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(updateRuleResponse),
	})

	if updateRuleResponse.Policy.WanNetwork.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateRuleResponse.Policy.WanNetwork.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publish policy revision after update
	tflog.Info(ctx, "publishing updated rule")
	_, err = r.client.Catov2.PolicyWanNetworkPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.client.Catov2.WanNetworkPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API WanNetworkPolicy error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().WanNetwork.Policy.Rules
	currentRule := &cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == updateRuleResponse.GetPolicy().GetWanNetwork().GetUpdateRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_WAN_updateRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := hydrateWanNetworkRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	// Handle exceptions correlation manually to preserve plan structure
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := wan.CorrelateWanExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	ruleObject, diags := types.ObjectValueFrom(ctx, WanNetworkRuleRuleAttrTypes, ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Rule = ruleObject

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wanNetworkRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Rule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract rule ID from state
	var ruleID string
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		rule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ruleID = rule.ID.ValueString()
	}

	// Prepare remove input
	removeInput := cato_models.WanNetworkRemoveRuleInput{
		ID: ruleID,
	}

	deleteRuleResponse, err := r.client.Catov2.PolicyWanNetworkRemoveRule(ctx, removeInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkRemoveRule error",
			err.Error(),
		)
		return
	}

	if deleteRuleResponse.Policy.WanNetwork.RemoveRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range deleteRuleResponse.Policy.WanNetwork.RemoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Deleting Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publish policy revision after update
	tflog.Info(ctx, "publishing updated rule")
	_, err = r.client.Catov2.PolicyWanNetworkPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
