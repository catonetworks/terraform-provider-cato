package ifrule

import (
	"context"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	activeperiod "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/activeperiod"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	policyschedule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyschedule"
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
	if createRuleResponse.Policy.InternetFirewall.AddRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range createRuleResponse.Policy.InternetFirewall.AddRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
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

	// Read rule and hydrate response to state
	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	body, err := r.resourceAPI().PolicyInternetFirewall(ctx, queryIfwPolicy, r.Client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == createRuleResponse.GetPolicy().GetInternetFirewall().GetAddRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleObject := convertCreateInput3(ctx, plan, &diags, currentRule)
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

	// Read rule and hydrate response to state
	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	body, err := r.resourceAPI().PolicyInternetFirewall(ctx, queryIfwPolicy, r.Client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			err.Error(),
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

func NewMinimalInternetFwRuleModel(ruleID string) InternetFirewallRule {
	ruleAttrs := map[string]attr.Value{
		"id":                convert.NullableString(ruleID),
		"name":              types.StringValue("test-ifw-rule"),
		"description":       types.StringNull(),
		"enabled":           types.BoolValue(true),
		"source":            emptyIfwSourceObject(),
		"connection_origin": types.StringNull(),
		"active_period":     types.ObjectNull(activeperiod.ActivePeriodAttrTypes),
		"country":           types.SetNull(idname.RefObjectType),
		"device":            types.SetNull(idname.RefObjectType),
		"device_attributes": types.ObjectNull(IfwDeviceAttrAttrTypes),
		"device_os":         types.ListNull(types.StringType),
		"destination":       emptyIfwDestinationObject(),
		"service":           emptyIfwServiceObject(),
		"action":            types.StringValue("ALLOW"),
		"tracking":          minimalTrackingObject(),
		"schedule":          types.ObjectNull(policyschedule.ScheduleAttrTypes),
		"exceptions":        types.SetNull(IfwExceptionObjectType),
	}

	return InternetFirewallRule{
		Rule: types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs),
		At: types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		}),
	}
}
