package parule

import (
	"context"
	"errors"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	apperror "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &privAccessRuleResource{}
	_ resource.ResourceWithConfigure   = &privAccessRuleResource{}
	_ resource.ResourceWithImportState = &privAccessRuleResource{}

	ErrPrivateAcccessRuleNotFound = errors.New("private access rule not found")
)

func NewPrivAccessRuleResource() resource.Resource {
	return &privAccessRuleResource{}
}

type privAccessRuleResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *privAccessRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_access_rule"
}

func (r *privAccessRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *privAccessRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create private access policy rule
func (r *privAccessRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleName, input := convertCreateInput(
		ctx,
		r,
		plan,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call Cato API to create a new rule
	tflog.Debug(ctx, "PolicyPrivateAccessAddRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().PolicyPrivateAccessAddRule(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PolicyPrivateAccessAddRule", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	errMsg := fmt.Sprintf("failed to add private access rule '%s'", ruleName)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	hydratedState := convertCreateInput2(
		ctx,
		r,
		&plan,
		&diags,
		result,
		errMsg,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read private access policy rule
func (r *privAccessRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydratePrivAccessRuleState(ctx, state.ID.ValueString(), false)
	if hydrateErr != nil {
		if errors.Is(hydrateErr, ErrPrivateAcccessRuleNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Private access rule %s not found, resource removed", state.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error hydrating privateAccessRule state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update private access policy rule
func (r *privAccessRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RuleModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("PolicyPrivateAccessUpdateRule: ID in unknown", "Rule ID is not set in TF state")
		return
	}

	input := cato_models.PrivateAccessUpdateRuleInput{
		ID: id,
		Rule: &cato_models.PrivateAccessUpdateRuleDataInput{
			Action:                r.prepareActionUpdate(plan.Action),
			ActivePeriod:          r.prepareActivePeriodUpdate(ctx, plan.ActivePeriod, &resp.Diagnostics),
			Applications:          r.prepareApplicationsUpdate(ctx, plan.Applications, &resp.Diagnostics),
			ConnectionsOriginList: r.prepareConnectionOrigins(ctx, plan.ConnectionOrigins, &resp.Diagnostics),
			Country:               r.prepareCountries(ctx, plan.Countries, &resp.Diagnostics),
			Description:           convert.KnownStringPointer(plan.Description),
			Device:                r.prepareDevice(ctx, plan.Devices, &resp.Diagnostics),
			Enabled:               convert.KnownBoolPointer(plan.Enabled),
			Name:                  convert.KnownStringPointer(plan.Name),
			Platform:              r.preparePlatforms(ctx, plan.Platforms, &resp.Diagnostics),
			Schedule:              r.prepareScheduleUpdate(ctx, plan.Schedule, &resp.Diagnostics),
			Source:                r.prepareSourceUpdate(ctx, plan.Source, &resp.Diagnostics),
			Tracking:              r.prepareTrackingUpdate(ctx, plan.Tracking, &resp.Diagnostics),
			UserAttributes:        r.prepareUserAttributesUpdate(ctx, plan.UserAttributes, &resp.Diagnostics),
		},
	}

	tflog.Debug(ctx, "PolicyPrivateAccessUpdateRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().PolicyPrivateAccessUpdateRule(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PolicyPrivateAccessUpdateRule", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	errMsg := fmt.Sprintf("failed to update private access rule '%s'", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	res := result.GetPolicy().GetPrivateAccess().GetUpdateRule()
	if *res.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		apiErrors := res.GetErrors()
		if len(apiErrors) == 0 {
			resp.Diagnostics.AddError(errMsg, "returned status: "+string(*res.GetStatus()))
			return
		}
		for _, e := range apiErrors {
			resp.Diagnostics.AddError(errMsg, fmt.Sprintf("ERROR: %v [%v]", *e.GetErrorMessage(), *e.GetErrorCode()))
		}
		return
	}

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydratePrivAccessRuleState(ctx, id, false)
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.AddError("Error hydrating private access policy rule state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete private access policy rule
//
//nolint:gocyclo,funlen
func (r *privAccessRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := convertDeleteInput(state)

	// Call Cato API to delete the private access rule.
	tflog.Debug(ctx, "PolicyPrivateAccessDeleteRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().PolicyPrivateAccessDeleteRule(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PolicyPrivateAccessDeleteRule", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	errMsg := fmt.Sprintf("Catov2 API PolicyPrivateAccessDeleteRule failed for '%s'", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	res := result.GetPolicy().GetPrivateAccess().GetRemoveRule()
	status := res.GetStatus()
	if status == nil || *status == "" {
		resp.Diagnostics.AddError(errMsg, "status not returned")
		return
	}
	if *status != cato_models.PolicyMutationStatusSuccess {
		apiErrors := res.GetErrors()
		if len(apiErrors) == 0 {
			resp.Diagnostics.AddError(errMsg, fmt.Sprintf("returned status: %s", string(*status)))
			return
		}
		for _, e := range apiErrors {
			errorCode := utils.GqlOptionalStr(e.GetErrorCode())
			errorMessage := utils.GqlOptionalStr(e.GetErrorMessage())
			// A previous Delete may have removed the rule but failed to publish.
			if errorCode == "ruleNotExist" || errorCode == "RuleNotFound" {
				continue
			}
			resp.Diagnostics.AddError(errMsg, fmt.Sprintf("ERROR: %v [%v]", errorMessage, errorCode))
		}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Publish the draft revision so the deletion takes effect and the rule no longer
	// appears in the active policy.
	// Mirrors resource_wf_rule.go Delete() and the bulk publish() helper.
	pubResult, err := r.resourceAPI().PolicyPrivateAccessPublishRevision(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Delete/PolicyPrivateAccessPublishRevision",
		map[string]interface{}{"response": utils.InterfaceToJSONString(pubResult)})
	errMsg = fmt.Sprintf("Catov2 API Delete/PolicyPrivateAccessPublishRevision failed for '%s'", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	pubRes := pubResult.GetPolicy().GetPrivateAccess().GetPublishPolicyRevision()
	status = pubRes.GetStatus()
	if status == nil || *status == "" {
		resp.Diagnostics.AddError(errMsg, "status not returned")
		return
	}
	if *status != cato_models.PolicyMutationStatusSuccess {
		apiErrors := pubRes.GetErrors()
		if len(apiErrors) == 0 {
			resp.Diagnostics.AddError(errMsg, "returned status: "+string(*status))
			return
		}
		for _, e := range apiErrors {
			errorCode := utils.GqlOptionalStr(e.GetErrorCode())
			errorMessage := utils.GqlOptionalStr(e.GetErrorMessage())
			// The deletion may already be published, leaving no draft to publish.
			// Publishing may have succeeded in a previous try, even if Terraform did not
			// receive the response. On retry, no remaining draft means there is nothing
			// left to publish.
			if errorCode == "PolicyRevisionNotFound" {
				continue
			}
			resp.Diagnostics.AddError(errMsg, fmt.Sprintf("ERROR: %v [%v]", errorMessage, errorCode))
		}
	}
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	ctx context.Context,
	r *privAccessRuleResource,
	plan *RuleModel,
	priorDiags *diag.Diagnostics,
	result *cato_go_sdk.PolicyPrivateAccessAddRule,
	errMsg string,
	diagnostics *diag.Diagnostics,
) (resultHydratedState *RuleModel) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	respRule := result.GetPolicy().GetPrivateAccess().GetAddRule()
	if respRule.Status != cato_models.PolicyMutationStatusSuccess {
		diagnostics.AddError(errMsg, "returned status: "+string(respRule.Status))
		for _, e := range respRule.Errors {
			diagnostics.AddError(errMsg, fmt.Sprintf("ERROR: %v [%v]", *e.GetErrorMessage(), *e.GetErrorCode()))
		}
		return resultHydratedState
	}
	ruleID := result.GetPolicy().GetPrivateAccess().GetAddRule().GetRule().GetRule().GetID()
	if ruleID == "" {
		diagnostics.AddError("Cato API PolicyPrivateAccessAddRule error", "rule ID is empty")
		return resultHydratedState
	}

	// Set the ID from the response
	plan.ID = types.StringValue(ruleID)

	// Hydrate state from API
	resultHydratedState, diags, hydrateErr := r.hydratePrivAccessRuleState(ctx, ruleID, true)
	if hydrateErr != nil {
		diagnostics.AddError("Error hydrating privateAccessRule state", hydrateErr.Error())
		diagnostics.Append(diags...)
		return resultHydratedState
	}
	return resultHydratedState
}

// hydratePrivAccessRuleState fetches the current state of a privAccessRule from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privAccessRuleResource) hydratePrivAccessRuleState(
	ctx context.Context,
	ruleID string,
	waitForCreation bool,
) (*RuleModel, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	polRule, err := lookupPrivateAccessRule(ctx, r.resourceAPI(), r.client.AccountId, ruleID, waitForCreation)
	if err != nil {
		return nil, diags, err
	}
	var state *RuleModel
	convertHydratePrivAccessRuleStateValues(ctx, r, &diags, &state, polRule)
	if diags.HasError() {
		return nil, diags, apperror.ErrAPIResponseParse
	}

	return state, nil, nil
}
