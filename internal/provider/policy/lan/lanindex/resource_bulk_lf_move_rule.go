package lanindex

import (
	"context"
	"fmt"
	"strconv"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource              = &lanRulesIndexResource{}
	_ resource.ResourceWithConfigure = &lanRulesIndexResource{}
	// _ resource.ResourceWithImportState = &lanRulesIndexResource{}
)

func NewLanRulesIndexResource() resource.Resource {
	return &lanRulesIndexResource{}
}

type lanRulesIndexResource struct {
	client       *catoclient.Service
	catov2Client LanFwRuleClient

	apiClient ResourceClient
}

func (r *lanRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_lf_move_rule"
}

func (r *lanRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
	r.catov2Client = r.client.Catov2
}

func (r *lanRulesIndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LanFwRulesIndex
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, indexMap := r.hydrateLanFwRulesIndex(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if utils.CheckErr(&resp.Diagnostics, r.moveSections(ctx, indexMap.sections)) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, r.moveRulesOrSubPolicies(ctx, indexMap.rulesOrSubPols)) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, r.moveFirewallRules(ctx, indexMap.firewallRules)) {
		return
	}

	// publish the changes
	r.publish(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// get final state from API
	hydratedState, _ = r.hydrateLanFwRulesIndex(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lanRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LanFwRulesIndex
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, _ := r.hydrateLanFwRulesIndexForRead(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.NetworkRules.IsNull() {
		hydratedState.NetworkRules = state.NetworkRules
	}
	if state.FirewallRules.IsNull() {
		hydratedState.FirewallRules = state.FirewallRules
	}
	if diags := resp.State.Set(ctx, &hydratedState); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *lanRulesIndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LanFwRulesIndex
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, indexMap := r.hydrateLanFwRulesIndex(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if utils.CheckErr(&resp.Diagnostics, r.moveSections(ctx, indexMap.sections)) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, r.moveRulesOrSubPolicies(ctx, indexMap.rulesOrSubPols)) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, r.moveFirewallRules(ctx, indexMap.firewallRules)) {
		return
	}

	// publish the changes
	r.publish(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// get final state from API
	hydratedState, _ = r.hydrateLanFwRulesIndex(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lanRulesIndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LanFwRulesIndex
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

// moveFwRuleToPosition moves the firewall rule with given ID to the given position in []currentRules
// (shifting the rest up) and calls the API to move the rule in the CMA
// Warning: it only moves the rules down, moving up is not supported!
func (r *lanRulesIndexResource) moveFwRuleToPosition(ctx context.Context, parentRuleID string,
	currentRules []nameID, fwRuleID, fwRuleName string, newPosition int,
) error {
	var myRule nameID

	tflog.Debug(ctx, "moving firewall rule '"+fwRuleName+"' to position "+strconv.Itoa(newPosition))

	curPosition := -1
	for i := range currentRules {
		if currentRules[i].id == fwRuleID {
			curPosition = i
			myRule = currentRules[i]
			break
		}
	}
	if curPosition == -1 {
		return fmt.Errorf("internal error: failed to find firewall rule ID %s", fwRuleID)
	}
	if curPosition == newPosition {
		return nil // nothing to do
	}

	input := convertMoveFwRuleToPositionValues(parentRuleID, currentRules, fwRuleID, newPosition, myRule, curPosition)

	// Call the API to move the sub-policy
	result, err := r.resourceAPI().PolicySocketLanFirewallMoveRule(ctx, r.client.AccountId, nil, input)
	if err != nil {
		return err
	}
	move := result.GetPolicy().GetSocketLan().GetFirewall().GetMoveRule()
	if errors := move.GetErrors(); len(errors) > 0 {
		msg := "unknown error"
		if m := errors[0].GetErrorMessage(); m != nil {
			msg = *m
		}
		return fmt.Errorf("failed to move LAN firewall rule '%s': %v", fwRuleName, msg)
	}
	if err := lanPolicyMutationStatusError(move.GetStatus()); err != nil {
		return fmt.Errorf("failed to move LAN firewall rule '%s': %w", fwRuleName, err)
	}
	return nil
}

// moveSubPolicyToPosition moves the sub-policy with given ID to the given position in []currentRulesOrSubPols
// (shifting the rest up) and calls the API to move the section in the CMA
// Warning: it only moves the sections down, moving up is not supported!
func (r *lanRulesIndexResource) moveSubPolicyToPosition(ctx context.Context, parentSectionID string,
	currentRulesOrSubPols []nameIDType, policyID, policyName string, newPosition int,
) error {
	var myPolicy nameIDType

	tflog.Debug(ctx, "moving sub-policy '"+policyName+"' to position "+strconv.Itoa(newPosition))

	curPosition := -1
	for i := range currentRulesOrSubPols {
		if currentRulesOrSubPols[i].id == policyID {
			curPosition = i
			myPolicy = currentRulesOrSubPols[i]
			break
		}
	}
	if curPosition == -1 {
		return fmt.Errorf("internal error: failed to find sub-policy ID %s", policyID)
	}
	if curPosition == newPosition {
		return nil // nothing to do
	}

	input := convertMoveSubPolicyToPositionValues(parentSectionID, currentRulesOrSubPols, policyID, newPosition, myPolicy, curPosition)

	// Call the API to move the sub-policy
	result, err := r.resourceAPI().PolicySocketLanMoveRule(ctx, input, r.client.AccountId)
	if err != nil {
		return err
	}
	move := result.GetPolicy().GetSocketLan().GetMoveRule()
	if errors := move.GetErrors(); len(errors) > 0 {
		msg := "unknown error"
		if m := errors[0].GetErrorMessage(); m != nil {
			msg = *m
		}
		return fmt.Errorf("failed to move LAN sub-policy '%s': %v", policyName, msg)
	}
	if err := lanPolicyMutationStatusError(move.GetStatus()); err != nil {
		return fmt.Errorf("failed to move LAN sub-policy '%s': %w", policyName, err)
	}
	return nil
}

// moveNetRuleToPosition moves the rule with given ID to the given position in []currentRulesOrSubPols
// (shifting the rest up) and calls the API to move the section in the CMA
// Warning: it only moves the sections down, moving up is not supported!
func (r *lanRulesIndexResource) moveNetRuleToPosition(ctx context.Context, parentSectionID string,
	currentRulesOrSubPols []nameIDType, ruleID, ruleName string, newPosition int,
) error {
	var myRule nameIDType

	tflog.Debug(ctx, "moving network rule '"+ruleName+"' to position "+strconv.Itoa(newPosition))

	curPosition := -1
	for i := range currentRulesOrSubPols {
		if currentRulesOrSubPols[i].id == ruleID {
			curPosition = i
			myRule = currentRulesOrSubPols[i]
			break
		}
	}
	if curPosition == -1 {
		return fmt.Errorf("internal error: failed to find rule ID %s", ruleID)
	}
	if curPosition == newPosition {
		return nil // nothing to do
	}

	input := convertMoveNetRuleToPositionValues(parentSectionID, currentRulesOrSubPols, ruleID, newPosition, myRule, curPosition)

	// Call the API to move the rule
	result, err := r.resourceAPI().PolicySocketLanMoveRule(ctx, input, r.client.AccountId)
	if err != nil {
		return err
	}
	move := result.GetPolicy().GetSocketLan().GetMoveRule()
	if errors := move.GetErrors(); len(errors) > 0 {
		msg := "unknown error"
		if m := errors[0].GetErrorMessage(); m != nil {
			msg = *m
		}
		return fmt.Errorf("failed to move LAN policy rule '%s': %v", ruleName, msg)
	}
	if err := lanPolicyMutationStatusError(move.GetStatus()); err != nil {
		return fmt.Errorf("failed to move LAN policy rule '%s': %w", ruleName, err)
	}
	return nil
}

func (r *lanRulesIndexResource) moveFirewallRules(ctx context.Context, fwRules map[string]itemOrder,
) (diags diag.Diagnostics) {
	moveFwRules := func(rules itemOrder) (err error) {
		for i := len(rules.target) - 1; i >= 0; i-- {
			if rules.target[i].id != rules.current[i].id {
				err = r.moveFwRuleToPosition(ctx, rules.parentID, rules.current, rules.target[i].id, rules.target[i].name, i)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	for _, ruleLists := range fwRules {
		if err := moveFwRules(ruleLists); err != nil {
			diags.AddError(fmt.Sprintf("failed to move firewall rule under network rule '%s'",
				ruleLists.parentName), err.Error())
			return
		}
	}

	return nil
}

func (r *lanRulesIndexResource) moveRulesOrSubPolicies(ctx context.Context, rulesOrSubPols map[string]itemOrderType,
) (diags diag.Diagnostics) {
	moveSectionRules := func(sect itemOrderType) (err error) {
		for i := len(sect.target) - 1; i >= 0; i-- {
			if sect.target[i].id != sect.current[i].id {
				if sect.target[i].ruleType == cato_models.PolicyRuleTypeEnumPolicyRule {
					err = r.moveNetRuleToPosition(ctx, sect.parentID, sect.current, sect.target[i].id, sect.target[i].name, i)
				} else {
					err = r.moveSubPolicyToPosition(ctx, sect.parentID, sect.current, sect.target[i].id, sect.target[i].name, i)
				}
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	for _, ruleLists := range rulesOrSubPols {
		if err := moveSectionRules(ruleLists); err != nil {
			diags.AddError(fmt.Sprintf("failed to move rule in LAN section '%s'", ruleLists.parentName), err.Error())
			return
		}
	}

	return nil
}

// moveSectionToPosition moves the section with given ID to the given position in []currentSections (shifting the rest up)
// and calls the API to move the section in the CMA
// Warning: it only moves the sections down, moving up is not supported!
func (r *lanRulesIndexResource) moveSectionToPosition(ctx context.Context, parentPolicyID string,
	currentSections []nameID, sectionID, sectionName string, newPosition int,
) error {
	var mySection nameID

	tflog.Debug(ctx, "moving section '"+sectionName+"' to position "+strconv.Itoa(newPosition))

	curPosition := -1
	for i := range currentSections {
		if currentSections[i].id == sectionID {
			curPosition = i
			mySection = currentSections[i]
			break
		}
	}
	if curPosition == -1 {
		return fmt.Errorf("internal error: failed to find sectionID %s", sectionID)
	}
	if curPosition == newPosition {
		return nil // nothing to do
	}

	input := convertMoveSectionToPositionValues(parentPolicyID, currentSections, sectionID, newPosition, mySection, curPosition)

	// Call the API to move the section
	result, err := r.resourceAPI().PolicySocketLanMoveSection(ctx, input, r.client.AccountId)
	if err != nil {
		return err
	}
	move := result.GetPolicy().GetSocketLan().GetMoveSection()
	if errors := move.GetErrors(); len(errors) > 0 {
		msg := "unknown error"
		if m := errors[0].GetErrorMessage(); m != nil {
			msg = *m
		}
		return fmt.Errorf("failed to move LAN policy section '%s': %v", sectionName, msg)
	}
	if err := lanPolicyMutationStatusError(move.GetStatus()); err != nil {
		return fmt.Errorf("failed to move LAN policy section '%s': %w", sectionName, err)
	}
	return nil
}

func (r *lanRulesIndexResource) moveSections(ctx context.Context, sections map[string]itemOrder) (diags diag.Diagnostics) {
	movePolicySections := func(sect itemOrder) error {
		// we need to loop backwards because the API does not support moving to the beginning
		for i := len(sect.target) - 1; i >= 0; i-- {
			if sect.target[i].id != sect.current[i].id {
				err := r.moveSectionToPosition(ctx, sect.parentID, sect.current, sect.target[i].id, sect.target[i].name, i)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	for _, sectionLists := range sections {
		if err := movePolicySections(sectionLists); err != nil {
			policyName := sectionLists.parentName
			if policyName == "" {
				policyName = "<main>"
			}
			diags.AddError(fmt.Sprintf("failed to move section in LAN policy '%s'", policyName), err.Error())
			return
		}
	}

	return nil
}

func (r *lanRulesIndexResource) hydrateLanFwRulesIndexWithAliases(ctx context.Context, plan *LanFwRulesIndex,
	aliases *lanFwStateAliases, diags *diag.Diagnostics,
) (newState *LanFwRulesIndex, indexMap *lfIndexMap) {
	// Call Cato API to get the policy
	result, err := r.resourceAPI().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("failed to hydrate sub-policy", err.Error())
		return nil, nil
	}
	policyBase := result.GetPolicy().GetSocketLan().GetPolicy()

	policySections, sectionData, sectionKeyToID := r.hydrateSections(ctx, plan, aliases, policyBase, diags)
	sectionRulesOrSubPols, netRuleData, netRuleKeyToID := r.hydrateNetRules(
		ctx, plan, aliases, policyBase, sectionKeyToID, diags,
	)
	ruleFirewallRules, firewallRuleData := r.hydrateFirewallRules(
		ctx, plan, aliases, policyBase, netRuleKeyToID, diags,
	)

	newState = &LanFwRulesIndex{
		SectionData:   sectionData,
		NetworkRules:  netRuleData,
		FirewallRules: firewallRuleData,
	}
	indexMap = &lfIndexMap{
		sections:       policySections,
		rulesOrSubPols: sectionRulesOrSubPols,
		firewallRules:  ruleFirewallRules,
	}
	return newState, indexMap
}

func (r *lanRulesIndexResource) hydrateLanFwRulesIndexForRead(ctx context.Context, state *LanFwRulesIndex,
	diags *diag.Diagnostics,
) (newState *LanFwRulesIndex, indexMap *lfIndexMap) {
	aliases := extractLanFwStateAliases(ctx, state, diags)
	if diags.HasError() {
		return nil, nil
	}
	return r.hydrateLanFwRulesIndexWithAliases(ctx, nil, aliases, diags)
}

func (r *lanRulesIndexResource) hydrateLanFwRulesIndex(ctx context.Context, plan *LanFwRulesIndex, diags *diag.Diagnostics,
) (newState *LanFwRulesIndex, indexMap *lfIndexMap) {
	return r.hydrateLanFwRulesIndexWithAliases(ctx, plan, nil, diags)
}
