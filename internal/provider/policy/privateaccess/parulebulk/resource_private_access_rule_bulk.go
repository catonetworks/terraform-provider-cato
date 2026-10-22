package parulebulk

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	apperror "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource               = &privAccessRuleBulkResource{}
	_ resource.ResourceWithConfigure  = &privAccessRuleBulkResource{}
	_ resource.ResourceWithModifyPlan = &privAccessRuleBulkResource{}

	ErrConvertError = errors.New("failed convert terraform types to go")
)

type privAccessRuleBulkResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func NewPrivAccessRuleBulkResource() resource.Resource {
	return &privAccessRuleBulkResource{}
}

func (r *privAccessRuleBulkResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_private_access_rule_bulk"
}

func (r *privAccessRuleBulkResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *privAccessRuleBulkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PrivateAccessRuleBulkModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, ruleMap, diags, hydrateErr := r.hydratePrivAccessRuleBulkState(ctx, &plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privateAccessRuleBulk state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	if utils.CheckErr(&resp.Diagnostics, r.moveRules(ctx, ruleMap)) {
		return
	}

	// publish the changes
	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Error publishing privateAcces policy", err.Error())
		return
	}

	// get final state from API
	hydratedState, _, diags, hydrateErr = r.hydratePrivAccessRuleBulkState(ctx, &plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privateAccessRuleBulk state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessRuleBulkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PrivateAccessRuleBulkModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, _, diags, hydrateErr := r.hydratePrivAccessRuleBulkState(ctx, &state)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privateAccessRuleBulk state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}
	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessRuleBulkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PrivateAccessRuleBulkModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PrivateAccessRuleBulkModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get state from API
	hydratedState, ruleMap, diags, hydrateErr := r.hydratePrivAccessRuleBulkState(ctx, &plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privateAccessRuleBulk state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	if utils.CheckErr(&resp.Diagnostics, r.moveRules(ctx, ruleMap)) {
		return
	}

	// publish the changes
	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Error publishing privateAcces policy", err.Error())
		return
	}

	// get final state from API
	hydratedState, _, diags, hydrateErr = r.hydratePrivAccessRuleBulkState(ctx, &plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privateAccessRuleBulk state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessRuleBulkResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	fmt.Println("bulk Delete")
}

// moveToPosition moves the rule with given ID to the given position in	[]currentRules (shifting the rest down)
// and calls the API to move the rule in the CMA
// Warning: it only moves the rule up, moving down is not supported!
func (r *privAccessRuleBulkResource) moveToPosition(
	ctx context.Context,
	currentRules []*PrivateAccessBulkRule,
	ruleID, ruleName string,
	newPosition int,
) error {
	var myRule *PrivateAccessBulkRule
	tflog.Debug(ctx, "moving private-access rule '"+ruleName+"' to position "+strconv.Itoa(newPosition))

	curPossition := -1
	for i := len(currentRules) - 1; i >= 0; i-- {
		if currentRules[i].ID.ValueString() == ruleID {
			curPossition = i
			myRule = currentRules[i]
			break
		}
	}
	if curPossition == -1 {
		return fmt.Errorf("internal error: failed to find ruleID")
	}
	if curPossition == newPosition {
		return nil // nothing to do
	}

	input := convertMoveToPositionValues(currentRules, ruleID, newPosition, myRule, curPossition)
	if r.client == nil {
		return nil
	}

	// Call the API to move the rule
	tflog.Debug(ctx, "Bulk PolicyPrivateAccessMoveRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().PolicyPrivateAccessMoveRule(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "Bulk PolicyPrivateAccessMoveRule", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return err
	}
	res := result.GetPolicy().GetPrivateAccess().GetMoveRule()
	if *res.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		apiErrors := res.GetErrors()
		if len(apiErrors) > 0 {
			return fmt.Errorf("error moving rule '%s' [%s] - %s", ruleName, *apiErrors[0].GetErrorCode(), *apiErrors[0].GetErrorMessage())
		}
		return fmt.Errorf("error moving rule '%s'", ruleName)
	}
	return nil
}

func (r *privAccessRuleBulkResource) moveRules(
	ctx context.Context,
	ruleMap map[string]*PrivateAccessBulkRule,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(ruleMap) == 0 {
		return nil
	}
	// Get the list of rules
	currentRules := make([]*PrivateAccessBulkRule, 0, len(ruleMap))
	targetRules := make([]*PrivateAccessBulkRule, 0, len(ruleMap))
	for _, rule := range ruleMap {
		targetRules = append(targetRules, rule)
		currentRules = append(currentRules, rule)
	}
	// rules as currently ordered in the draft state in CMA
	slices.SortFunc(currentRules, func(a, b *PrivateAccessBulkRule) int {
		return cmp.Compare(a.CMAIndex, b.CMAIndex)
	})
	// rules in the order we want
	slices.SortFunc(targetRules, func(a, b *PrivateAccessBulkRule) int {
		return cmp.Compare(a.Index.ValueInt64(), b.Index.ValueInt64())
	})

	for i := range targetRules {
		if targetRules[i].ID.ValueString() != currentRules[i].ID.ValueString() {
			err := r.moveToPosition(ctx, currentRules, targetRules[i].ID.ValueString(), targetRules[i].Name.ValueString(), i)
			if err != nil {
				diags.AddError("failed to reorganize private-access rules", err.Error())
				return diags
			}
			continue
		}
	}
	return nil
}

// hydratePrivAccessRuleBulkState fetches the current state of a privAccessRuleBulk from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privAccessRuleBulkResource) hydratePrivAccessRuleBulkState(
	ctx context.Context,
	plan *PrivateAccessRuleBulkModel,
) (*PrivateAccessRuleBulkModel, map[string]*PrivateAccessBulkRule, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	// Create a Go map[rule-name]PrivateAccessBulkRule{...} from the plan/state
	stateRules := make(map[string]PrivateAccessBulkRule)
	if utils.HasValue(plan.RuleData) {
		var tfBulk map[string]types.Object
		if utils.CheckErr(&diags, plan.RuleData.ElementsAs(ctx, &tfBulk, false)) {
			return nil, nil, diags, ErrConvertError
		}
		for name, obj := range tfBulk {
			tfRule := convertHydratePrivAccessRuleBulkStateValues3(ctx, diags, obj)
			if diags.HasError() {
				return nil, nil, diags, ErrConvertError
			}
			stateRules[name] = tfRule
		}
	}

	// Call Cato API to get the policy
	result, err := r.resourceAPI().PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Bulk PolicyReadPrivateAccessPolicy", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Merge rules from the API and rules from the state
	apiRulesTf := make(map[string]types.Object)
	apiRulesGo := make(map[string]*PrivateAccessBulkRule)
	policy := result.GetPolicy().GetPrivateAccess().GetPolicy()
	if len(policy.Rules) != len(stateRules) {
		return nil, nil, nil, fmt.Errorf(
			"rules returned by API (count=%d) do not match the PrivateAccessRuleBulkModel state (count=%d)",
			len(policy.Rules),
			len(stateRules),
		)
	}
	for _, polRule := range policy.Rules {
		apiRule, tfRule, stateRule, ok := convertHydratePrivAccessRuleBulkStateValues4(stateRules, polRule)
		if !ok {
			return nil, nil, nil, fmt.Errorf("rule '%s' from API is not found in the PrivateAccessRuleBulkModel state", apiRule.Name)
		}
		ruleObj := convertHydratePrivAccessRuleBulkStateValues5(ctx, diags, tfRule, stateRule)
		if diags.HasError() {
			return nil, nil, diags, apperror.ErrAPIResponseParse
		}
		apiRulesTf[apiRule.Name] = ruleObj
		apiRulesGo[apiRule.Name] = tfRule
	}

	// Prepate the tf state - ruleData map from enriched apiRules
	rulesMap := convertHydratePrivAccessRuleBulkStateValues(ctx, diags, apiRulesTf)
	if diags.HasError() {
		return nil, nil, diags, apperror.ErrAPIResponseParse
	}
	state := convertHydratePrivAccessRuleBulkStateValues2(plan, rulesMap)

	return state, apiRulesGo, nil, nil
}
