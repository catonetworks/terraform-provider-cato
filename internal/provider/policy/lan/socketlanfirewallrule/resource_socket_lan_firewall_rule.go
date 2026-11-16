package socketlanfirewallrule

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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

	apiClient ResourceClient
}

func (r *socketLanFirewallRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_socket_lan_firewall_rule"
}

func (r *socketLanFirewallRuleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
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

func (r *socketLanFirewallRuleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
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

	var policyChange *cato_go_sdk.PolicySocketLanFirewallAddRule
	err := sharedpolicy.WithPolicyRevisionConflictRetry(ctx, "PolicySocketLanFirewallAddRule", func() error {
		var apiErr error
		policyChange, apiErr = r.resourceAPI().PolicySocketLanFirewallAddRule(ctx, r.client.AccountId, nil, apiInput.create)
		return apiErr
	})
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

	if policyChange.Policy.SocketLan.Firewall.AddRule.Status != cato_models.PolicyMutationStatusSuccess ||
		len(policyChange.Policy.SocketLan.Firewall.AddRule.Errors) > 0 {
		for _, e := range policyChange.Policy.SocketLan.Firewall.AddRule.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
		}
		if len(policyChange.Policy.SocketLan.Firewall.AddRule.Errors) == 0 {
			resp.Diagnostics.AddError(
				"API Error Creating Rule Resource",
				fmt.Sprintf("AddRule status %q", string(policyChange.Policy.SocketLan.Firewall.AddRule.Status)),
			)
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
	if !r.publishPolicyRevision(ctx, &resp.Diagnostics) {
		return
	}

	// Read back the rule to populate state
	queryResult := r.readPolicyWithRetry(ctx, &resp.Diagnostics)
	if queryResult == nil {
		return
	}

	// Find the created firewall rule
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

func (r *socketLanFirewallRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocketLanFirewallRule
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
	ruleObj, diagstmp := convertReadState2(ctx, state, currentRule)
	resp.Diagnostics.Append(diagstmp...)

	// Hard code position to avoid drift
	convertReadState3(&state, ruleObj, &diagstmp, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanFirewallRuleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SocketLanFirewallRule
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
	apiInput, inputDiags := hydrateSocketLanFirewallRuleAPI(ctx, plan)
	resp.Diagnostics.Append(inputDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the rule ID for update
	positionInput := convertUpdateInput2(ctx, plan, &diags, ruleID, &apiInput)
	resp.Diagnostics.Append(diags...)

	moveInput := convertUpdateInput3(ruleID, positionInput)

	tflog.Debug(ctx, "Update.PolicySocketLanFirewallMoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(moveInput),
	})
	var moveResult *cato_go_sdk.PolicySocketLanFirewallMoveRule
	err := sharedpolicy.WithPolicyRevisionConflictRetry(ctx, "PolicySocketLanFirewallMoveRule", func() error {
		var apiErr error
		moveResult, apiErr = r.resourceAPI().PolicySocketLanFirewallMoveRule(ctx, r.client.AccountId, nil, moveInput)
		return apiErr
	})
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
	var updateResult *cato_go_sdk.PolicySocketLanFirewallUpdateRule
	err = sharedpolicy.WithPolicyRevisionConflictRetry(ctx, "PolicySocketLanFirewallUpdateRule", func() error {
		var apiErr error
		updateResult, apiErr = r.resourceAPI().PolicySocketLanFirewallUpdateRule(ctx, r.client.AccountId, nil, apiInput.update)
		return apiErr
	})
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
	if !r.publishPolicyRevision(ctx, &resp.Diagnostics) {
		return
	}

	// Read back and update state
	queryResult := r.readPolicyWithRetry(ctx, &resp.Diagnostics)
	if queryResult == nil {
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

func (r *socketLanFirewallRuleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SocketLanFirewallRule
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

	tflog.Debug(ctx, "Delete.PolicySocketLanFirewallRemoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeInput),
	})
	var removeResult *cato_go_sdk.PolicySocketLanFirewallRemoveRule
	err := sharedpolicy.WithPolicyRevisionConflictRetry(ctx, "PolicySocketLanFirewallRemoveRule", func() error {
		var apiErr error
		removeResult, apiErr = r.resourceAPI().PolicySocketLanFirewallRemoveRule(ctx, r.client.AccountId, nil, removeInput)
		return apiErr
	})
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
	if !r.publishPolicyRevision(ctx, &resp.Diagnostics) {
		return
	}
}

// readPolicyWithRetry reads the socket LAN policy, retrying transient backend failures so a
// flaky read-back after a successful mutation doesn't orphan the just-created object.
// Failures are recorded on diags and reported as a nil result.
func (r *socketLanFirewallRuleResource) readPolicyWithRetry(
	ctx context.Context, diags *diag.Diagnostics,
) *cato_go_sdk.PolicySocketLanPolicy {
	var queryResult *cato_go_sdk.PolicySocketLanPolicy
	err := sharedpolicy.WithTransientReadRetry(ctx, "PolicySocketLanPolicy", func() error {
		var readErr error
		queryResult, readErr = r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
		return readErr
	})
	if err != nil {
		diags.AddError(
			"Catov2 API PolicySocketLanPolicy error",
			err.Error(),
		)
		return nil
	}
	return queryResult
}

// publishPolicyRevision publishes the socket LAN policy revision, retrying while the API
// reports concurrent policy revisions. Failures are recorded on diags; it reports success.
func (r *socketLanFirewallRuleResource) publishPolicyRevision(ctx context.Context, diags *diag.Diagnostics) bool {
	publishInput := &cato_models.PolicyPublishRevisionInput{}
	err := sharedpolicy.WithPolicyRevisionConflictRetry(ctx, "PolicySocketLanPublishPolicyRevision", func() error {
		_, publishErr := r.resourceAPI().PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
		return publishErr
	})
	if err != nil {
		diags.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return false
	}
	return true
}
