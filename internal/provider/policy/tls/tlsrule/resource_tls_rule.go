package tlsrule

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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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
}

func (r *tlsInspectionRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tls_rule"
}

func (r *tlsInspectionRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

//nolint:funlen
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

	createRuleResponse, err := r.client.Catov2.PolicyTLSInspectAddRule(ctx, input.create, r.client.AccountId)

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
	_, err = r.client.Catov2.PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.client.Catov2.Tlsinspectpolicy(ctx, r.client.AccountId)
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
	ruleInputRead.ID = types.StringValue(currentRule.ID)
	tflog.Info(ctx, "ruleInputRead - "+fmt.Sprintf("%v", ruleInputRead))

	ruleObject, diags := types.ObjectValueFrom(ctx, TLSInspectionRuleRuleAttrTypes, ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Rule = ruleObject
	plan.ID = types.StringValue(currentRule.ID)

	// Set the complete plan to state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
}

//nolint:gocyclo,funlen
func (r *tlsInspectionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InspectionRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Catov2.Tlsinspectpolicy(ctx, r.client.AccountId)
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
		rule := PolicyPolicyTLSInspectPolicyRulesRule{}
		diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
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
	positionValue := string(cato_models.PolicySectionPositionEnumLastInPolicy)
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
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID at root level
	diags = resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(ruleID))
	resp.Diagnostics.Append(diags...)
}

//nolint:gocyclo,funlen
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
	inputMoveRule := cato_models.PolicyMoveRuleInput{}

	// Setting at (to move rule)
	if !plan.At.IsNull() {
		inputMoveRule.To = &cato_models.PolicyRulePositionInput{}
		positionInput := policyrule.Position{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputMoveRule.To.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		inputMoveRule.To.Ref = positionInput.Ref.ValueStringPointer()
	}

	ruleInput := PolicyPolicyTLSInspectPolicyRulesRule{}
	diags = plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	// Settings other rule attributes
	inputMoveRule.ID = *ruleInput.ID.ValueStringPointer()
	input.update.ID = *ruleInput.ID.ValueStringPointer()

	// Move rule
	moveRule, err := r.client.Catov2.PolicyTLSInspectMoveRule(ctx, inputMoveRule, r.client.AccountId)
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
	updateRuleResponse, err := r.client.Catov2.PolicyTLSInspectUpdateRule(ctx, input.update, r.client.AccountId)
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
	_, err = r.client.Catov2.PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.client.Catov2.Tlsinspectpolicy(ctx, r.client.AccountId)
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
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	ruleObject, diags := types.ObjectValueFrom(ctx, TLSInspectionRuleRuleAttrTypes, ruleInputRead)
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

	ruleInput := PolicyPolicyTLSInspectPolicyRulesRule{}
	diags = state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	removeRuleInput := cato_models.TLSInspectRemoveRuleInput{
		ID: *ruleInput.ID.ValueStringPointer(),
	}

	removeRuleResponse, err := r.client.Catov2.PolicyTLSInspectRemoveRule(ctx, removeRuleInput, r.client.AccountId)
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
	_, err = r.client.Catov2.PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}

func (r *tlsInspectionRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Set rule.id to the imported ID
	resp.State.SetAttribute(ctx, path.Root("rule").AtName("id"), req.ID)
}
