package socketlannetworkrule

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &socketLanNetworkRuleResource{}
	_ resource.ResourceWithConfigure   = &socketLanNetworkRuleResource{}
	_ resource.ResourceWithImportState = &socketLanNetworkRuleResource{}
)

func NewSocketLanNetworkRuleResource() resource.Resource {
	return &socketLanNetworkRuleResource{}
}

type socketLanNetworkRuleResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *socketLanNetworkRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_socket_lan_network_rule"
}

func (r *socketLanNetworkRuleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *socketLanNetworkRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *socketLanNetworkRuleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SocketLanNetworkRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API input
	apiInput, inputDiags := hydrateSocketLanNetworkRuleAPI(ctx, plan)
	resp.Diagnostics.Append(inputDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySocketLanAddRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(apiInput.create),
	})

	policyChange, err := r.resourceAPI().PolicySocketLanAddRule(ctx, apiInput.create, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicySocketLanAddRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanAddRule error",
			err.Error(),
		)
		return
	}

	if len(policyChange.Policy.SocketLan.AddRule.Errors) > 0 {
		for _, e := range policyChange.Policy.SocketLan.AddRule.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
		}
		return
	}
	ruleID := policyChange.GetPolicy().GetSocketLan().GetAddRule().GetRule().GetRule().GetID()
	if ruleID == "" {
		resp.Diagnostics.AddError("Rule ID missing", "The API did not return the ID for the created LAN network rule.")
		return
	}

	// Publish the changes
	tflog.Info(ctx, "Create.publishing-rule")
	publishInput := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read back the rule to populate state
	queryResult, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPolicy error",
			err.Error(),
		)
		return
	}

	// Find the created rule
	ruleObj, diagstmp := convertCreateInput2(ctx, plan, ruleID, queryResult, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diagstmp...)

	// Set the state
	plan.Rule = ruleObj
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanNetworkRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocketLanNetworkRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from state
	ruleData := convertReadState(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := ruleData.ID.ValueString()

	// Query the API
	queryResult, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
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

	// Find the rule
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule
	ruleExists := false
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		if ruleWrapper.Rule.ID == ruleID {
			currentRule = &ruleWrapper.Rule
			ruleExists = true
			break
		}
	}

	if !ruleExists {
		tflog.Warn(ctx, "socket lan network rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	// Hydrate state from API response
	ruleObj, diagstmp := convertReadState2(ctx, state, currentRule)
	resp.Diagnostics.Append(diagstmp...)

	// Hard code position to avoid drift
	convertReadState3(&state, ruleObj, &diagstmp, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanNetworkRuleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SocketLanNetworkRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from plan
	ruleData := convertUpdateInput(ctx, plan, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := ruleData.ID.ValueString()

	// Build API input
	apiInput, inputDiags := hydrateSocketLanNetworkRuleAPI(ctx, plan)
	resp.Diagnostics.Append(inputDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the rule ID for update
	positionInput := convertUpdateInput2(ctx, plan, &diags, ruleID, &apiInput)
	resp.Diagnostics.Append(diags...)

	moveInput := convertUpdateInput3(ruleID, positionInput)

	tflog.Debug(ctx, "Update.PolicySocketLanMoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(moveInput),
	})
	moveResult, err := r.resourceAPI().PolicySocketLanMoveRule(ctx, moveInput, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicySocketLanMoveRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanMoveRule error",
			err.Error(),
		)
		return
	}

	// Update the rule
	tflog.Debug(ctx, "Update.PolicySocketLanUpdateRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(apiInput.update),
	})
	updateResult, err := r.resourceAPI().PolicySocketLanUpdateRule(ctx, nil, apiInput.update, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicySocketLanUpdateRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanUpdateRule error",
			err.Error(),
		)
		return
	}

	if updateResult.Policy.SocketLan.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateResult.Policy.SocketLan.UpdateRule.GetErrors() {
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
	_, err = r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read back and update state
	queryResult, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPolicy error",
			err.Error(),
		)
		return
	}

	ruleObj, diagstmp := convertUpdateInput5(ctx, plan, ruleID, queryResult, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diagstmp...)

	plan.Rule = ruleObj
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanNetworkRuleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SocketLanNetworkRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from state
	ruleData := convertDeleteInput(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeInput := convertDeleteInput2(ruleData)

	tflog.Debug(ctx, "Delete.PolicySocketLanRemoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeInput),
	})
	removeResult, err := r.resourceAPI().PolicySocketLanRemoveRule(ctx, nil, removeInput, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicySocketLanRemoveRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeResult),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanRemoveRule error",
			err.Error(),
		)
		return
	}

	// Publish the changes
	tflog.Info(ctx, "Delete.publishing-rule")
	publishInput := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
