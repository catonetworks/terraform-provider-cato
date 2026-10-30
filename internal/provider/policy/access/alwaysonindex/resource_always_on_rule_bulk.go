package alwaysonindex

import (
	"context"
	"fmt"
	"strconv"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource              = &alwaysOnRuleBulkResource{}
	_ resource.ResourceWithConfigure = &alwaysOnRuleBulkResource{}
)

type alwaysOnRuleBulkResource struct {
	client     *catoclient.Service
	mockClient ResourceClient
}

// NewAlwaysOnRuleBulkResource returns the cato_always_on_rule_bulk ordering resource.
func NewAlwaysOnRuleBulkResource() resource.Resource { return &alwaysOnRuleBulkResource{} }

func (r *alwaysOnRuleBulkResource) getClient() ResourceClient {
	if r.mockClient != nil {
		return r.mockClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.Catov2
}

func (r *alwaysOnRuleBulkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_always_on_rule_bulk"
}

func (r *alwaysOnRuleBulkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *alwaysOnRuleBulkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config RulesIndex
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := r.readPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newState := r.reorderAll(ctx, *state, config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *alwaysOnRuleBulkResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state := r.readPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *alwaysOnRuleBulkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config RulesIndex
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// it is necessary to read the state again, TF may call Read() before creating sections/rules
	state := r.readPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newState := r.reorderAll(ctx, *state, config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *alwaysOnRuleBulkResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Publish changes
	r.publish(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.RemoveResource(ctx)
}

// - config as target
// - validate it matches
// - add IDs
// - move
func (r *alwaysOnRuleBulkResource) reorderAll(ctx context.Context, state, config RulesIndex, diags *diag.Diagnostics,
) (newState *RulesIndex) {
	sectionOrder, sectionIDs := buildSectionOrder(ctx, state, config, diags)
	ruleOrder := buildRuleOrder(ctx, state, config, sectionIDs, diags)
	if diags.HasError() {
		return nil
	}

	// Reorder sections
	r.moveSections(ctx, sectionOrder, diags)
	if diags.HasError() {
		return nil
	}

	// Reorder rules
	r.moveRules(ctx, ruleOrder, diags)
	if diags.HasError() {
		return nil
	}

	// Publish changes
	r.publish(ctx, diags)
	if diags.HasError() {
		return nil
	}

	// Re-read the policy
	newState = r.readPolicy(ctx, diags)
	if diags.HasError() {
		return nil
	}
	return newState
}

func (r *alwaysOnRuleBulkResource) moveRules(ctx context.Context, orders map[string]ruleOrder, diags *diag.Diagnostics) {
	// we need to loop backwards because the API does not support moving to the beginning
	for secName, rules := range orders {
		for i := len(rules.target) - 1; i >= 0; i-- {
			if rules.target[i].id != rules.current[i].id {
				r.moveRuleToPosition(ctx, rules.current, rules.target[i].id, rules.target[i].name, i,
					rules.sectionID, secName, diags)
				if diags.HasError() {
					return
				}
			}
		}
	}
}

func (r *alwaysOnRuleBulkResource) publish(ctx context.Context, diags *diag.Diagnostics) {
	const summary = "Failed to publish Always On policy ordering"
	result, err := r.getClient().PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, r.client.AccountId, nil, nil)
	if err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetPublishPolicyRevision()
	if errors := payload.GetErrors(); len(errors) == 1 {
		if errors[0].ErrorCode != nil && *errors[0].ErrorCode == apperror.PolicyRevisionNotFound {
			return // no policy draft to discard; OK
		}
	}
	if utils.CheckAPIErrors(nil, payload.GetErrors(), summary, diags) {
		return
	}
	checkMutationStatus(payload.GetStatus(), summary, diags)
}

func checkMutationStatus(status *cato_models.PolicyMutationStatus, summary string, diags *diag.Diagnostics) bool {
	if status == nil {
		diags.AddError(summary, "API did not return a mutation status")
		return false
	}
	if *status != cato_models.PolicyMutationStatusSuccess {
		diags.AddError(summary, fmt.Sprintf("API returned mutation status %q", *status))
		return false
	}
	return true
}

func (r *alwaysOnRuleBulkResource) moveSections(ctx context.Context, sect sectionOrder, diags *diag.Diagnostics) {
	// we need to loop backwards because the API does not support moving to the beginning
	for i := len(sect.target) - 1; i >= 0; i-- {
		if sect.target[i].id != sect.current[i].id {
			r.moveSectionToPosition(ctx, sect.current, sect.target[i].id, sect.target[i].name, i, diags)
			if diags.HasError() {
				return
			}
		}
	}
}

// moveSectionToPosition moves the section with given ID to the given position in []currentSections (shifting the rest up)
// and calls the API to move the section in the CMA
// Warning: it only moves the sections down, moving up is not supported!
func (r *alwaysOnRuleBulkResource) moveSectionToPosition(ctx context.Context,
	currentSections []nameID, sectionID, sectionName string, newPosition int, diags *diag.Diagnostics,
) {
	var mySection nameID

	tflog.Debug(ctx, "moving section '"+sectionName+"' to position "+strconv.Itoa(newPosition))
	summary := fmt.Sprintf("failed to move AlwaysOn policy section '%s'", sectionName)

	// Find the section that is to be moved in []currentSections --> curPosition
	curPosition := -1
	for i := range currentSections {
		if currentSections[i].id == sectionID {
			curPosition = i
			break
		}
	}
	if curPosition == -1 {
		diags.AddError(summary, "internal error: failed to find sectionID "+sectionID)
	}
	if curPosition == newPosition {
		return // nothing to do
	}

	// Move the section item to the bottom of []currentSections
	mySection = currentSections[curPosition]
	for i := curPosition; i < newPosition; i++ {
		currentSections[i] = currentSections[i+1]
	}
	currentSections[newPosition] = mySection

	// Call the API to move the section
	input := buildMoveSectionInput(currentSections, newPosition)
	result, err := r.getClient().PolicyZtnaAlwaysOnMoveSection(ctx, r.client.AccountId, nil, input)
	if utils.CheckAPIErrors(err, result.GetPolicy().GetZtnaAlwaysOn().GetMoveSection().GetErrors(), summary, diags) {
		return
	}
}

// moveRuleToPosition moves the section with given ID to the given position in []currentSections (shifting the rest up)
// and calls the API to move the section in the CMA
// Warning: it only moves the sections down, moving up is not supported!
func (r *alwaysOnRuleBulkResource) moveRuleToPosition(ctx context.Context,
	currentRules []nameID, ruleID, ruleName string, newPosition int, secID, secName string, diags *diag.Diagnostics,
) {
	var myRule nameID

	tflog.Debug(ctx, "moving rule '"+ruleName+"' in section '"+secName+"' to position "+strconv.Itoa(newPosition))
	summary := fmt.Sprintf("failed to move AlwaysOn policy rule '%s'", ruleName)

	// Find the rule that is to be moved in []currentRules --> curPosition
	curPosition := -1
	for i := range currentRules {
		if currentRules[i].id == ruleID {
			curPosition = i
			break
		}
	}
	if curPosition == -1 {
		diags.AddError(summary, "internal error: failed to find ruleID "+ruleID)
	}
	if curPosition == newPosition {
		return // nothing to do
	}

	// Move the section item to the bottom of []currentRules
	myRule = currentRules[curPosition]
	for i := curPosition; i < newPosition; i++ {
		currentRules[i] = currentRules[i+1]
	}
	currentRules[newPosition] = myRule

	// Call the API to move the section
	input := buildMoveRuleInput(currentRules, newPosition, secID)
	result, err := r.getClient().PolicyZtnaAlwaysOnMoveRule(ctx, r.client.AccountId, nil, input)
	if utils.CheckAPIErrors(err, result.GetPolicy().GetZtnaAlwaysOn().GetMoveRule().GetErrors(), summary, diags) {
		return
	}
}
