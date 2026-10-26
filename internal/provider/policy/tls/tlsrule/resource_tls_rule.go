package tlsrule

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &tlsInspectionRuleResource{}
	_ resource.ResourceWithConfigure   = &tlsInspectionRuleResource{}
	_ resource.ResourceWithImportState = &tlsInspectionRuleResource{}
)

func NewTLSInspectionRuleResource() resource.Resource {
	return &tlsInspectionRuleResource{}
}

type tlsInspectionRuleResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *tlsInspectionRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_tls_rule"
}

func (r *tlsInspectionRuleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *tlsInspectionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InspectionRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateTLSRuleAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_TLS_input.create", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.create),
	})

	createRuleResponse, err := r.resourceAPI().PolicyTLSInspectAddRule(ctx, input.create, r.client.AccountId)

	tflog.Warn(ctx, "TFLOG_WARN_TLS_createRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(createRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectAddRule error",
			err.Error(),
		)
		return
	}

	// Check for errors
	if createRuleResponse.Policy.TLSInspect.AddRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range createRuleResponse.Policy.TLSInspect.AddRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publishing new rule
	tflog.Info(ctx, "publishing new TLS rule")
	_, err = r.resourceAPI().PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.resourceAPI().Tlsinspectpolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspect error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().TLSInspect.Policy.GetRules()
	currentRule := &cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == createRuleResponse.GetPolicy().GetTLSInspect().GetAddRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			resp.State.SetAttribute(
				ctx,
				path.Root("rule").AtName("id"),
				ruleListItem.GetRule().ID)
			break
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_TLS_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := hydrateTLSRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	convertCreateInput(currentRule, &ruleInputRead)
	tflog.Info(ctx, "ruleInputRead - "+fmt.Sprintf("%v", ruleInputRead))

	ruleObject := convertCreateInput2(ctx, &diags, ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	populateTLSRuleState(&plan, currentRule, ruleObject)

	// Set the complete plan to state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
}

func (r *tlsInspectionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InspectionRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.resourceAPI().Tlsinspectpolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspect error",
			err.Error(),
		)
		return
	}

	// Retrieve rule ID
	var ruleID string
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		rule := convertReadValues(ctx, state, &diags)
		if diags.HasError() {
			tflog.Debug(ctx, "Failed to convert full state to struct, attempting to extract ID only")
			resp.Diagnostics = resp.Diagnostics[:0]

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

	ruleList := body.GetPolicy().TLSInspect.Policy.GetRules()
	ruleExist := false
	currentRule := &cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule{}
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == ruleID {
			ruleExist = true
			currentRule = ruleListItem.GetRule()

			resp.State.SetAttribute(
				ctx,
				path.Root("rule").AtName("id"),
				ruleListItem.GetRule().ID)
		}
	}

	// Remove resource if it doesn't exist anymore
	if !ruleExist {
		tflog.Warn(ctx, "TLS inspection rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_TLS_readRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	ruleInput, hydrateDiags := hydrateTLSRuleState(ctx, state, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)

	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)

	// Check if position is set in state, if not default to LAST_IN_POLICY
	curAtObj, diagstmp := convertReadState(ctx, state, &diags, &resp.Diagnostics)
	resp.Diagnostics.Append(diagstmp...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.SetAttribute(ctx, path.Root("at"), curAtObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID at root level
	diags = resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(ruleID))
	resp.Diagnostics.Append(diags...)
}

//nolint:funlen
func (r *tlsInspectionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InspectionRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateTLSRuleAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Setting input for moving rule
	inputMoveRule, ruleInput := convertUpdateInput(ctx, plan, &diags, &resp.Diagnostics)
	resp.Diagnostics.Append(diags...)

	// Settings other rule attributes
	inputMoveRule.ID = *ruleInput.ID.ValueStringPointer()
	input.update.ID = *ruleInput.ID.ValueStringPointer()

	// Move rule
	moveRule, err := r.resourceAPI().PolicyTLSInspectMoveRule(ctx, inputMoveRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectMoveRule error",
			err.Error(),
		)
		return
	}

	// Check for errors
	if moveRule.Policy.TLSInspect.MoveRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range moveRule.Policy.TLSInspect.MoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Rule Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_TLS_input.update", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.update),
	})

	// Updating rule
	updateRuleResponse, err := r.resourceAPI().PolicyTLSInspectUpdateRule(ctx, input.update, r.client.AccountId)
	tflog.Warn(ctx, "TFLOG_WARN_TLS_updateRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(updateRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectUpdateRule error",
			err.Error(),
		)
		return
	}

	// Check for errors
	if updateRuleResponse.Policy.TLSInspect.UpdateRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range updateRuleResponse.Policy.TLSInspect.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publishing updated rule
	tflog.Info(ctx, "publishing updated TLS rule")
	_, err = r.resourceAPI().PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.resourceAPI().Tlsinspectpolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspect error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().TLSInspect.Policy.GetRules()
	currentRule := &cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule{}
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == updateRuleResponse.GetPolicy().GetTLSInspect().GetUpdateRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}

	tflog.Warn(ctx, "TFLOG_WARN_TLS_updateRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := hydrateTLSRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	ruleObject := convertUpdateInput2(ctx, &diags, currentRule, &ruleInputRead)
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

func (r *tlsInspectionRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InspectionRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleInput := convertDeleteInput(ctx, state, &diags)
	resp.Diagnostics.Append(diags...)

	removeRuleInput := convertDeleteInput2(ruleInput)

	removeRuleResponse, err := r.resourceAPI().PolicyTLSInspectRemoveRule(ctx, removeRuleInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectRemoveRule error",
			err.Error(),
		)
		return
	}

	// Check for errors
	if removeRuleResponse.Policy.TLSInspect.RemoveRule.Status != cato_models.PolicyMutationStatusSuccess {
		for _, item := range removeRuleResponse.Policy.TLSInspect.RemoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Deleting Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publishing rule deletion
	tflog.Info(ctx, "publishing TLS rule deletion")
	_, err = r.resourceAPI().PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}

func (r *tlsInspectionRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Set rule.id to the imported ID
	resp.State.SetAttribute(ctx, path.Root("rule").AtName("id"), req.ID)
}
