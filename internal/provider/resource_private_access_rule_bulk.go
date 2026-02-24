package provider

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource               = &privAccessRuleBulkResource{}
	_ resource.ResourceWithConfigure  = &privAccessRuleBulkResource{}
	_ resource.ResourceWithModifyPlan = &privAccessRuleBulkResource{}

	ErrConvertError = errors.New("failed convert terraform types to go")
)

type privAccessRuleBulkResource struct {
	client *catoClientData
}

func NewPrivAccessRuleBulkResource() resource.Resource {
	return &privAccessRuleBulkResource{}
}

func (r *privAccessRuleBulkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_access_rule_bulk"
}

func (r *privAccessRuleBulkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *privAccessRuleBulkResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ordering and publishng private access policy rules.",
		Attributes: map[string]schema.Attribute{
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of private access rule policy Indexes keyed by rule_name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Rule id",
							Computed:    true,
						},
						"index": schema.Int64Attribute{
							Description: "Requierd position of the rule",
							Required:    true,
						},
						"cma_index": schema.Int64Attribute{
							Description: "Rule index in CMA before bulk is applied",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "IFW rule name housing rule",
							Required:    true,
						},
					},
				},
			},
			"publish": schema.Int64Attribute{
				Description: "publish policy revision if there is a change",
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

type revisionPlanModifier struct{}

func (m revisionPlanModifier) Description(_ context.Context) string {
	return "publish policy revision if there is a change"
}
func (m revisionPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m revisionPlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Always set to current time in microseconds
	nowMicro := time.Now().Unix()
	resp.PlanValue = types.Int64Value(nowMicro)
}

func (r *privAccessRuleBulkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PrivateAccessRuleBulkModel

	XXX(ctx, "Bulk Create")
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

	if checkErr(&resp.Diagnostics, r.moveRules(ctx, ruleMap)) {
		return
	}

	// publish the changes
	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Error publishing privateAcces policy", err.Error())
		return
	}

	// get final state from API
	hydratedState, ruleMap, diags, hydrateErr = r.hydratePrivAccessRuleBulkState(ctx, &plan)
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
	XXX(ctx, "Bulk Read")
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
	XXX(ctx, "Bulk Update")
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

	if checkErr(&resp.Diagnostics, r.moveRules(ctx, ruleMap)) {
		return
	}

	// publish the changes
	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Error publishing privateAcces policy", err.Error())
		return
	}

	// get final state from API
	hydratedState, ruleMap, diags, hydrateErr = r.hydratePrivAccessRuleBulkState(ctx, &plan)
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

func (r *privAccessRuleBulkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *PrivateAccessRuleBulkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan == nil {
		XXX(ctx, "Bulk Modify Plan is NULL")
		return
	}
	XXX(ctx, "Bulk Modify Plan current state: state: %v", plan.Publish)

	plan.Publish = types.Int64Value(plan.Publish.ValueInt64() + 1)
	XXX(ctx, "Bulk Modify Plan new state: %v", plan.Publish)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessRuleBulkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	XXX(ctx, "Bulk Delete")
}

// hydratePrivAccessRuleBulkState fetches the current state of a privAccessRuleBulk from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privAccessRuleBulkResource) hydratePrivAccessRuleBulkState(ctx context.Context, plan *PrivateAccessRuleBulkModel) (*PrivateAccessRuleBulkModel, map[string]*PrivateAccessBulkRule, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	// Create a Go map[rule-name]PrivateAccessBulkRule{...} from the plan/state
	stateRules := make(map[string]PrivateAccessBulkRule)
	if hasValue(plan.RuleData) {
		var tfBulk map[string]types.Object
		if checkErr(&diags, plan.RuleData.ElementsAs(ctx, &tfBulk, false)) {
			return nil, nil, diags, ErrConvertError
		}
		for name, obj := range tfBulk {
			var tfRule PrivateAccessBulkRule
			diags.Append(obj.As(ctx, &tfRule, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, nil, diags, ErrConvertError
			}
			stateRules[name] = tfRule
		}
	}

	// Call Cato API to get the policy
	result, err := r.client.catov2.PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId)
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
		return nil, nil, nil, fmt.Errorf("Rules returned by API (count=%d) do not match the PrivateAccessRuleBulkModel state (count=%d)", len(policy.Rules), len(stateRules))
	}
	for _, polRule := range policy.Rules {
		apiRule := polRule.Rule
		tfRule := &PrivateAccessBulkRule{
			ID:       types.StringValue(apiRule.ID),
			Name:     types.StringValue(apiRule.Name),
			CMAIndex: types.Int64Value(apiRule.Index),
		}
		stateRule, ok := stateRules[apiRule.Name]
		if !ok {
			return nil, nil, nil, fmt.Errorf("Rule '%s' from API is not found in the PrivateAccessRuleBulkModel state", apiRule.Name)
		}
		tfRule.Index = stateRule.Index
		ruleObj, diag := types.ObjectValueFrom(ctx, PrivateAccessBulkRuleTypes, tfRule)
		diags.Append(diag...)
		if diags.HasError() {
			return nil, nil, diags, ErrAPIResponseParse
		}
		apiRulesTf[apiRule.Name] = ruleObj
		apiRulesGo[apiRule.Name] = tfRule
	}

	// Prepate the tf state - ruleData map from enriched apiRules
	rulesMap, diag := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: PrivateAccessBulkRuleTypes}, apiRulesTf)
	diags.Append(diag...)
	if diags.HasError() {
		return nil, nil, diags, ErrAPIResponseParse
	}
	XXX(ctx, "Hydrate:  plan.Publish=%v", plan.Publish)
	state := &PrivateAccessRuleBulkModel{
		RuleData: rulesMap,
		Publish:  types.Int64Value(plan.Publish.ValueInt64()),
	}

	return state, apiRulesGo, nil, nil
}

func (r *privAccessRuleBulkResource) moveRules(ctx context.Context, ruleMap map[string]*PrivateAccessBulkRule) diag.Diagnostics {
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
		return cmp.Compare(a.CMAIndex.ValueInt64(), b.CMAIndex.ValueInt64())
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

// moveToPosition moves the rule with given ID to the given position in	[]currentRules (shifting the rest down)
// and calls the API to move the rule in the CMA
func (r *privAccessRuleBulkResource) moveToPosition(ctx context.Context, currentRules []*PrivateAccessBulkRule, ruleID, ruleName string, newPosition int) error {
	var myRule *PrivateAccessBulkRule
	tflog.Debug(ctx, "moving private-access rule '"+ruleName+"' to position "+strconv.Itoa(newPosition))
	XXX(ctx, "moving private-access rule '"+ruleName+"' to position "+strconv.Itoa(newPosition))

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

	for i := curPossition; i > newPosition; i-- {
		currentRules[i] = currentRules[i-1]
	}
	currentRules[newPosition] = myRule

	// Prepare input for API to move the rule
	var input cato_models.PolicyMoveRuleInput
	if newPosition == 0 {
		input = cato_models.PolicyMoveRuleInput{
			ID: ruleID,
			To: &cato_models.PolicyRulePositionInput{Position: ptr(cato_models.PolicyRulePositionEnumFirstInPolicy)},
		}
	} else {
		input = cato_models.PolicyMoveRuleInput{
			ID: ruleID,
			To: &cato_models.PolicyRulePositionInput{
				Position: ptr(cato_models.PolicyRulePositionEnumAfterRule),
				Ref:      currentRules[newPosition-1].ID.ValueStringPointer(),
			},
		}
	}
	// Call the API to move the rule
	tflog.Debug(ctx, "Bulk PolicyPrivateAccessMoveRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.catov2.PolicyPrivateAccessMoveRule(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "Bulk PolicyPrivateAccessMoveRule", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return err
	}
	res := result.GetPolicy().GetPrivateAccess().GetMoveRule()
	if *res.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		errors := res.GetErrors()
		if len(errors) > 0 {
			return fmt.Errorf("error moving rule '%s' [%s] - %s", ruleName, *errors[0].GetErrorCode(), *errors[0].GetErrorMessage())
		}
		return fmt.Errorf("error moving rule '%s'", ruleName)
	}
	return nil
}

// publish calls the API to publish the draft policy revision
func (r *privAccessRuleBulkResource) publish(ctx context.Context) error {
	result, err := r.client.catov2.PolicyPrivateAccessPublishRevision(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Bulk PolicyPrivateAccessPublishRevision", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return err
	}
	res := result.GetPolicy().GetPrivateAccess().GetPublishPolicyRevision()
	if *res.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		errors := res.GetErrors()
		if len(errors) > 0 {
			if *errors[0].GetErrorCode() == "PolicyRevisionNotFound" {
				return nil // there was nothing to publish
			}
			return fmt.Errorf("error publishing policy - %s", *errors[0].GetErrorMessage())
		}
		return fmt.Errorf("error publishing policy")
	}
	return nil
}

func XXX(ctx context.Context, msg string, args ...any) {
	tflog.Warn(ctx, fmt.Sprintf("\033[0;33mXXX "+msg+"\033[0m", args...))
}

// TODO: check status": "SUCCESS" on API calls
