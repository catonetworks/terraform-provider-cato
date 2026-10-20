package lanfirewall

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
	_ resource.Resource                = &socketLanFirewallRuleResource{}
	_ resource.ResourceWithConfigure   = &socketLanFirewallRuleResource{}
	_ resource.ResourceWithImportState = &socketLanFirewallRuleResource{}
)

func NewSocketLanFirewallRuleResource() resource.Resource {
	return &socketLanFirewallRuleResource{}
}

type socketLanFirewallRuleResource struct {
	client *catoclient.Service
}

func (r *socketLanFirewallRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_lan_firewall_rule"
}

func (r *socketLanFirewallRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *socketLanFirewallRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *socketLanFirewallRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SocketLanFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API input
	apiInput, inputDiags := hydrateSocketLanFirewallRuleAPI(ctx, plan)
	resp.Diagnostics.Append(inputDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySocketLanFirewallAddRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(apiInput.create),
	})

	policyChange, err := r.client.Catov2.PolicySocketLanFirewallAddRule(ctx, r.client.AccountId, nil, apiInput.create)
	tflog.Debug(ctx, "Create.PolicySocketLanFirewallAddRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanFirewallAddRule error",
			err.Error(),
		)
		return
	}

	if len(policyChange.Policy.SocketLan.Firewall.AddRule.Errors) > 0 {
		for _, e := range policyChange.Policy.SocketLan.Firewall.AddRule.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
		}
		return
	}
	ruleID := policyChange.GetPolicy().GetSocketLan().GetFirewall().GetAddRule().GetRule().GetRule().GetID()
	if ruleID == "" {
		resp.Diagnostics.AddError("Rule ID missing", "The API did not return the ID for the created LAN firewall rule.")
		return
	}

	// Publish the changes
	tflog.Info(ctx, "Create.publishing-rule")
	publishInput := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.Catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read back the rule to populate state
	queryResult, err := r.client.Catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPolicy error",
			err.Error(),
		)
		return
	}

	// Find the created firewall rule
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		for _, fwWrapper := range ruleWrapper.Rule.Firewall {
			if fwWrapper.Rule.ID == ruleID {
				currentRule = &fwWrapper.Rule
				break
			}
		}
		if currentRule != nil {
			break
		}
	}

	if currentRule == nil {
		resp.Diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find created firewall rule with ID %s", ruleID),
		)
		return
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanFirewallRuleState(ctx, plan, currentRule)

	// Build the rule object
	ruleObj, diagstmp := types.ObjectValueFrom(ctx, SocketLanFirewallRuleRuleAttrTypes, ruleState)
	resp.Diagnostics.Append(diagstmp...)

	// Set the state
	plan.Rule = ruleObj
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanFirewallRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocketLanFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from state
	var ruleData SocketLanFirewallRuleData
	diags = state.Rule.As(ctx, &ruleData, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := ruleData.ID.ValueString()

	// Query the API
	queryResult, err := r.client.Catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	tflog.Debug(ctx, "Read.PolicySocketLanPolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// Find the firewall rule
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule
	ruleExists := false
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		for _, fwWrapper := range ruleWrapper.Rule.Firewall {
			if fwWrapper.Rule.ID == ruleID {
				currentRule = &fwWrapper.Rule
				ruleExists = true
				break
			}
		}
		if currentRule != nil {
			break
		}
	}

	if !ruleExists {
		tflog.Warn(ctx, "socket lan firewall rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanFirewallRuleState(ctx, state, currentRule)

	// Build the rule object
	ruleObj, diagstmp := types.ObjectValueFrom(ctx, SocketLanFirewallRuleRuleAttrTypes, ruleState)
	resp.Diagnostics.Append(diagstmp...)

	// Hard code position to avoid drift
	curAtObj := state.At
	if !utils.HasValue(state.At) {
		curAtObj, diagstmp = types.ObjectValue(
			policyrule.PositionAttrTypes,
			map[string]attr.Value{
				"position": types.StringValue("LAST_IN_RULE"),
				"ref":      types.StringNull(),
			},
		)
		resp.Diagnostics.Append(diagstmp...)
	}

	state.Rule = ruleObj
	state.At = curAtObj

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

//nolint:funlen
func (r *socketLanFirewallRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SocketLanFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from plan
	var ruleData SocketLanFirewallRuleData
	diags = plan.Rule.As(ctx, &ruleData, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := ruleData.ID.ValueString()

	// Build API input
	apiInput, inputDiags := hydrateSocketLanFirewallRuleAPI(ctx, plan)
	resp.Diagnostics.Append(inputDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the rule ID for update
	apiInput.update.ID = ruleID

	// Move rule if position changed
	var positionInput policyrule.Position
	diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	moveInput := cato_models.PolicyMoveSubRuleInput{
		ID: ruleID,
		To: &cato_models.PolicySubRulePositionInput{
			Position: cato_models.PolicySubRulePositionEnum(positionInput.Position.ValueString()),
			Ref:      positionInput.Ref.ValueString(),
		},
	}

	tflog.Debug(ctx, "Update.PolicySocketLanFirewallMoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(moveInput),
	})
	moveResult, err := r.client.Catov2.PolicySocketLanFirewallMoveRule(ctx, r.client.AccountId, nil, moveInput)
	tflog.Debug(ctx, "Update.PolicySocketLanFirewallMoveRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanFirewallMoveRule error",
			err.Error(),
		)
		return
	}

	// Update the rule
	tflog.Debug(ctx, "Update.PolicySocketLanFirewallUpdateRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(apiInput.update),
	})
	updateResult, err := r.client.Catov2.PolicySocketLanFirewallUpdateRule(ctx, r.client.AccountId, nil, apiInput.update)
	tflog.Debug(ctx, "Update.PolicySocketLanFirewallUpdateRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanFirewallUpdateRule error",
			err.Error(),
		)
		return
	}

	if updateResult.Policy.SocketLan.Firewall.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateResult.Policy.SocketLan.Firewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Rule Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publish the changes
	tflog.Info(ctx, "Update.publishing-rule")
	publishInput := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.Catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read back and update state
	queryResult, err := r.client.Catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPolicy error",
			err.Error(),
		)
		return
	}

	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		for _, fwWrapper := range ruleWrapper.Rule.Firewall {
			if fwWrapper.Rule.ID == ruleID {
				currentRule = &fwWrapper.Rule
				break
			}
		}
		if currentRule != nil {
			break
		}
	}

	if currentRule == nil {
		resp.Diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find firewall rule with ID %s after update", ruleID),
		)
		return
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanFirewallRuleState(ctx, plan, currentRule)

	// Build the rule object
	ruleObj, diagstmp := types.ObjectValueFrom(ctx, SocketLanFirewallRuleRuleAttrTypes, ruleState)
	resp.Diagnostics.Append(diagstmp...)

	plan.Rule = ruleObj
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanFirewallRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SocketLanFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from state
	var ruleData SocketLanFirewallRuleData
	diags = state.Rule.As(ctx, &ruleData, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := ruleData.ID.ValueString()

	removeInput := cato_models.SocketLanFirewallRemoveRuleInput{
		ID: ruleID,
	}

	tflog.Debug(ctx, "Delete.PolicySocketLanFirewallRemoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeInput),
	})
	removeResult, err := r.client.Catov2.PolicySocketLanFirewallRemoveRule(ctx, r.client.AccountId, nil, removeInput)
	tflog.Debug(ctx, "Delete.PolicySocketLanFirewallRemoveRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeResult),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanFirewallRemoveRule error",
			err.Error(),
		)
		return
	}

	// Publish the changes
	tflog.Info(ctx, "Delete.publishing-rule")
	publishInput := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.Catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
