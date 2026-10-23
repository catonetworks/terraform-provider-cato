package lanindex

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func lanPolicyMutationStatusError(status *cato_models.PolicyMutationStatus) error {
	if status == nil {
		return fmt.Errorf("API mutation returned no status")
	}
	if *status != cato_models.PolicyMutationStatusSuccess {
		return fmt.Errorf("API mutation returned status %q", *status)
	}
	return nil
}

func (r *lanRulesIndexResource) hydrateSections(ctx context.Context, plan *LanFwRulesIndex,
	aliases *lanFwStateAliases, policyBase *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
	diags *diag.Diagnostics,
) (policySections map[string]itemOrder, sectionData types.Map, sectionKeyToID map[string]string) {
	policySections = make(map[string]itemOrder)
	policyIDMap := r.makePolicyIDMap(policyBase) // map[policyID]policyName
	sectionDataNull := types.MapNull(types.ObjectType{AttrTypes: LanFwSectionDataTypes})

	// Prepare section map
	//   map[policyName]{  // "" means the main policy
	//   	current:[{secName,secID},...]   // from API response
	//   	target: [{secName,secID},...]   // from TF plan
	//   }
	r.parseAPISections(policyBase, policyIDMap, policySections, diags)
	if diags.HasError() { // current
		return nil, sectionDataNull, nil
	}
	sectionKeyToID = make(map[string]string)
	if plan != nil {
		r.parsePlanSections(plan, policyBase, policySections, sectionKeyToID, diags)
		if diags.HasError() { // target
			return nil, sectionDataNull, nil
		}
		r.checkSections(policySections, diags)
		if diags.HasError() { // should contain the same items
			return nil, sectionDataNull, nil
		}
	}

	sectionKeyByID := invertKeyToID(sectionKeyToID)
	if aliases != nil {
		sectionKeyByID = collisionSafeStateKeys(flattenSectionItems(policySections), aliases.sectionKeyByID)
		sectionKeyToID = invertIDToKey(sectionKeyByID)
	}
	sections := make(map[string]types.Object)
	for _, sectionLists := range policySections {
		for i, section := range sectionLists.current {
			key := stateMapKey(sectionKeyByID, section.id, section.name)
			if _, exists := sections[key]; exists {
				diags.AddError("cannot hydrate LAN firewall sections",
					fmt.Sprintf("multiple sections would use map key %q; use distinct section_data map keys and set section_name explicitly", key))
				return nil, sectionDataNull, nil
			}
			tfSection := LanFwSectionData{
				ID:            types.StringValue(section.id),
				SectionName:   types.StringValue(section.name),
				SectionIndex:  types.Int64Value(int64(i + 1)), // 1-based
				SubPolicyName: types.StringValue(sectionLists.parentName),
			}
			sectionObj, objDiags := types.ObjectValueFrom(ctx, LanFwSectionDataTypes, tfSection)
			diags.Append(objDiags...)
			if diags.HasError() {
				return nil, sectionDataNull, nil
			}
			sections[key] = sectionObj
		}
	}
	sd, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: LanFwSectionDataTypes}, sections)
	if objDiags.HasError() {
		diags.Append(objDiags...)
		return nil, sectionDataNull, nil
	}
	sectionData = sd
	return policySections, sectionData, sectionKeyToID
}

//nolint:gocyclo // Hydration keeps plan, prior-state aliases, and API state synchronized in one pass.
func (r *lanRulesIndexResource) hydrateNetRules(ctx context.Context, plan *LanFwRulesIndex,
	aliases *lanFwStateAliases, policyBase *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
	sectionKeyToID map[string]string, diags *diag.Diagnostics,
) (sectionRulesOrSubPols map[string]itemOrderType, netRuleData types.Map, netRuleKeyToID map[string]string) {
	sectionRulesOrSubPols = make(map[string]itemOrderType)
	netRuleDataNull := types.MapNull(types.ObjectType{AttrTypes: LanNetworkRuleTypes})

	// Prepare rule and subpolicy map
	//   map[sectionName]{
	//   	current:[{ruleName,id,RULE},{subPolicyName,id,SUB_POL},...]  // from API response
	//   	target: [{ruleName,id,RULE},{subPolicyName,id,SUB_POL},...]  // from TF plan
	//   }
	r.parseAPINetRules(policyBase, sectionRulesOrSubPols) // current
	if diags.HasError() {
		return nil, netRuleDataNull, nil
	}
	manageNetworkRules := plan == nil || utils.HasValue(plan.NetworkRules)
	if aliases != nil {
		manageNetworkRules = aliases.manageNetworkRules
	}
	netRuleKeyToID = make(map[string]string)
	netRuleParentKeyByID := make(map[string]string)
	if plan != nil && manageNetworkRules {
		r.parsePlanNetRules(
			plan, sectionKeyToID, sectionRulesOrSubPols, netRuleKeyToID, netRuleParentKeyByID, diags,
		) // target
		if diags.HasError() {
			return nil, netRuleDataNull, nil
		}
		r.checkNetRules(sectionRulesOrSubPols, diags) // should contain the same items
		if diags.HasError() {
			return nil, netRuleDataNull, nil
		}
	}

	netRuleKeyByID := invertKeyToID(netRuleKeyToID)
	if aliases != nil {
		netRuleKeyByID = collisionSafeStateKeys(flattenNetworkRuleItems(sectionRulesOrSubPols), aliases.netRuleKeyByID)
		netRuleKeyToID = invertIDToKey(netRuleKeyByID)
		netRuleParentKeyByID = aliases.netRuleParentKeyByID
	}
	sectionKeyByID := invertKeyToID(sectionKeyToID)
	rulesOrSubPols := make(map[string]types.Object)
	for _, rList := range sectionRulesOrSubPols {
		for i, rule := range rList.current {
			key := stateMapKey(netRuleKeyByID, rule.id, rule.name)
			if _, exists := rulesOrSubPols[key]; exists {
				diags.AddError("cannot hydrate LAN firewall network rules",
					fmt.Sprintf("multiple network rules would use map key %q; use distinct network_rules map keys and set rule_name explicitly", key))
				return nil, netRuleDataNull, nil
			}
			sectionKey := netRuleParentKeyByID[rule.id]
			if aliases != nil {
				if _, existed := aliases.netRuleKeyByID[rule.id]; !existed {
					sectionKey = sectionKeyByID[rList.parentID]
				} else if sectionKey != "" && sectionKeyToID[sectionKey] != rList.parentID {
					sectionKey = sectionKeyByID[rList.parentID]
				}
			}
			tfRuleData := LanNetworkRule{
				ID:             types.StringValue(rule.id),
				RuleType:       types.StringValue(string(rule.ruleType)),
				RuleName:       types.StringValue(rule.name),
				SectionName:    types.StringValue(rList.parentName),
				SectionKey:     optionalStringValue(sectionKey),
				IndexInSection: types.Int64Value(int64(i + 1)), // 1-based
			}
			ruleObj, objDiags := types.ObjectValueFrom(ctx, LanNetworkRuleTypes, tfRuleData)
			diags.Append(objDiags...)
			if diags.HasError() {
				return nil, netRuleDataNull, nil
			}
			rulesOrSubPols[key] = ruleObj
		}
	}

	netRuleData = netRuleDataNull
	if manageNetworkRules {
		rd, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: LanNetworkRuleTypes}, rulesOrSubPols)
		if objDiags.HasError() {
			diags.Append(objDiags...)
			return nil, netRuleDataNull, nil
		}
		netRuleData = rd
	}
	return sectionRulesOrSubPols, netRuleData, netRuleKeyToID
}

//nolint:gocyclo // Hydration keeps plan, prior-state aliases, and API state synchronized in one pass.
func (r *lanRulesIndexResource) hydrateFirewallRules(ctx context.Context, plan *LanFwRulesIndex,
	aliases *lanFwStateAliases, policyBase *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
	netRuleKeyToID map[string]string, diags *diag.Diagnostics,
) (ruleFirewallRules map[string]itemOrder, firewallRuleData types.Map) {
	ruleFirewallRules = make(map[string]itemOrder)
	firewallRuleDataNull := types.MapNull(types.ObjectType{AttrTypes: LanFirewallRuleTypes})

	// Prepare firewall-rule map
	//   map[netRuleName]{
	//   	current:[{fwRuleName,id},{fwRuleName,id},...]  // from API response
	//   	target: [{fwRuleName,id},{fwRuleName,id},...]  // from TF plan
	//   }
	r.parseAPIFwRules(policyBase, ruleFirewallRules) // current
	if diags.HasError() {
		return nil, firewallRuleDataNull
	}
	manageFirewallRules := plan == nil || utils.HasValue(plan.FirewallRules)
	if aliases != nil {
		manageFirewallRules = aliases.manageFirewallRules
	}
	firewallRuleKeyToID := make(map[string]string)
	firewallParentKeyByID := make(map[string]string)
	if plan != nil && manageFirewallRules {
		r.parsePlanFwRules(
			plan, netRuleKeyToID, ruleFirewallRules, firewallRuleKeyToID, firewallParentKeyByID, diags,
		) // target
		if diags.HasError() {
			return nil, firewallRuleDataNull
		}
		r.checkFwRules(ruleFirewallRules, diags) // should contain the same items
		if diags.HasError() {
			return nil, firewallRuleDataNull
		}
	}

	firewallRuleKeyByID := invertKeyToID(firewallRuleKeyToID)
	if aliases != nil {
		firewallRuleKeyByID = collisionSafeStateKeys(flattenFirewallRuleItems(ruleFirewallRules), aliases.firewallRuleKeyByID)
		firewallParentKeyByID = aliases.firewallParentKeyByID
	}
	netRuleKeyByID := invertKeyToID(netRuleKeyToID)
	tfFirewallRules := make(map[string]types.Object)
	for _, rList := range ruleFirewallRules {
		for i, rule := range rList.current {
			key := stateMapKey(firewallRuleKeyByID, rule.id, rule.name)
			if _, exists := tfFirewallRules[key]; exists {
				diags.AddError("cannot hydrate LAN firewall rules",
					fmt.Sprintf("multiple firewall rules would use map key %q; "+
						"use distinct firewall_rules map keys and set firewall_rule_name explicitly", key))
				return nil, firewallRuleDataNull
			}
			netRuleKey := firewallParentKeyByID[rule.id]
			if aliases != nil {
				if _, existed := aliases.firewallRuleKeyByID[rule.id]; !existed {
					netRuleKey = netRuleKeyByID[rList.parentID]
				} else if netRuleKey != "" && netRuleKeyToID[netRuleKey] != rList.parentID {
					netRuleKey = netRuleKeyByID[rList.parentID]
				}
			}
			tfRuleData := LanFirewallRule{
				ID:               types.StringValue(rule.id),
				FirewallRuleName: types.StringValue(rule.name),
				NetRuleName:      types.StringValue(rList.parentName),
				NetRuleKey:       optionalStringValue(netRuleKey),
				IndexInRule:      types.Int64Value(int64(i + 1)), // 1-based
			}
			ruleObj, objDiags := types.ObjectValueFrom(ctx, LanFirewallRuleTypes, tfRuleData)
			diags.Append(objDiags...)
			if diags.HasError() {
				return nil, firewallRuleDataNull
			}
			tfFirewallRules[key] = ruleObj
		}
	}

	firewallRuleData = firewallRuleDataNull
	if manageFirewallRules {
		rd, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: LanFirewallRuleTypes}, tfFirewallRules)
		if objDiags.HasError() {
			diags.Append(objDiags...)
			return nil, firewallRuleDataNull
		}
		firewallRuleData = rd
	}

	return ruleFirewallRules, firewallRuleData
}

func extractLanFwStateAliases(ctx context.Context, state *LanFwRulesIndex,
	diags *diag.Diagnostics,
) *lanFwStateAliases {
	aliases := &lanFwStateAliases{
		sectionKeyByID:        make(map[string]string),
		netRuleKeyByID:        make(map[string]string),
		firewallRuleKeyByID:   make(map[string]string),
		netRuleParentKeyByID:  make(map[string]string),
		firewallParentKeyByID: make(map[string]string),
		manageNetworkRules:    utils.HasValue(state.NetworkRules),
		manageFirewallRules:   utils.HasValue(state.FirewallRules),
	}

	var sections map[string]LanFwSectionData
	sectionDiags := state.SectionData.ElementsAs(ctx, &sections, false)
	diags.Append(sectionDiags...)
	if diags.HasError() {
		return aliases
	}
	for key, section := range sections {
		if hasConfiguredValue(section.ID) {
			aliases.sectionKeyByID[section.ID.ValueString()] = key
		}
	}

	if aliases.manageNetworkRules {
		var networkRules map[string]LanNetworkRule
		networkDiags := state.NetworkRules.ElementsAs(ctx, &networkRules, false)
		diags.Append(networkDiags...)
		if diags.HasError() {
			return aliases
		}
		for key, rule := range networkRules {
			if !hasConfiguredValue(rule.ID) {
				continue
			}
			id := rule.ID.ValueString()
			aliases.netRuleKeyByID[id] = key
			if hasConfiguredValue(rule.SectionKey) {
				aliases.netRuleParentKeyByID[id] = rule.SectionKey.ValueString()
			}
		}
	}

	if aliases.manageFirewallRules {
		var firewallRules map[string]LanFirewallRule
		firewallDiags := state.FirewallRules.ElementsAs(ctx, &firewallRules, false)
		diags.Append(firewallDiags...)
		if diags.HasError() {
			return aliases
		}
		for key, rule := range firewallRules {
			if !hasConfiguredValue(rule.ID) {
				continue
			}
			id := rule.ID.ValueString()
			aliases.firewallRuleKeyByID[id] = key
			if hasConfiguredValue(rule.NetRuleKey) {
				aliases.firewallParentKeyByID[id] = rule.NetRuleKey.ValueString()
			}
		}
	}

	return aliases
}

func flattenSectionItems(groups map[string]itemOrder) []nameID {
	var items []nameID
	for _, group := range groups {
		items = append(items, group.current...)
	}
	return items
}

func flattenNetworkRuleItems(groups map[string]itemOrderType) []nameID {
	var items []nameID
	for _, group := range groups {
		for _, item := range group.current {
			items = append(items, nameID{name: item.name, id: item.id})
		}
	}
	return items
}

func flattenFirewallRuleItems(groups map[string]itemOrder) []nameID {
	return flattenSectionItems(groups)
}

func collisionSafeStateKeys(items []nameID, priorKeys map[string]string) map[string]string {
	keys := make(map[string]string, len(items))
	present := make(map[string]nameID, len(items))
	for _, item := range items {
		present[item.id] = item
	}

	ids := make([]string, 0, len(present))
	for id := range present {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	used := make(map[string]struct{}, len(items))
	for _, id := range ids {
		if key := priorKeys[id]; key != "" {
			keys[id] = key
			used[key] = struct{}{}
		}
	}

	slices.SortFunc(items, func(a, b nameID) int {
		if byName := cmp.Compare(a.name, b.name); byName != 0 {
			return byName
		}
		return cmp.Compare(a.id, b.id)
	})
	for _, item := range items {
		if keys[item.id] != "" {
			continue
		}
		base := item.name
		if base == "" {
			base = item.id
		}
		candidate := base
		if _, exists := used[candidate]; exists {
			candidate = base + "__" + item.id
			for suffix := 2; ; suffix++ {
				if _, exists := used[candidate]; !exists {
					break
				}
				candidate = fmt.Sprintf("%s__%s__%d", base, item.id, suffix)
			}
		}
		keys[item.id] = candidate
		used[candidate] = struct{}{}
	}
	return keys
}

func invertKeyToID(keyToID map[string]string) map[string]string {
	idToKey := make(map[string]string, len(keyToID))
	for key, id := range keyToID {
		idToKey[id] = key
	}
	return idToKey
}

func invertIDToKey(idToKey map[string]string) map[string]string {
	keyToID := make(map[string]string, len(idToKey))
	for id, key := range idToKey {
		keyToID[key] = id
	}
	return keyToID
}

func stateMapKey(keyByID map[string]string, id, fallbackName string) string {
	if key := keyByID[id]; key != "" {
		return key
	}
	return fallbackName
}

func optionalStringValue(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func configuredName(value types.String, fallback string) string {
	if hasConfiguredValue(value) {
		return value.ValueString()
	}
	return fallback
}

func hasConfiguredValue(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown() && value.ValueString() != ""
}

func resolveNameID(name, preferredID string, items []nameID) (string, error) {
	if preferredID != "" {
		for _, item := range items {
			if item.id == preferredID {
				if item.name != name {
					return "", fmt.Errorf("stored ID %q now has name %q", preferredID, item.name)
				}
				return item.id, nil
			}
		}
	}
	var matches []string
	for _, item := range items {
		if item.name == name {
			matches = append(matches, item.id)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("name not found")
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("name is ambiguous within its parent")
	}
}

func resolveNameIDType(name, preferredID string, items []nameIDType) (nameIDType, error) {
	if preferredID != "" {
		for _, item := range items {
			if item.id == preferredID {
				if item.name != name {
					return nameIDType{}, fmt.Errorf("stored ID %q now has name %q", preferredID, item.name)
				}
				return item, nil
			}
		}
	}
	var matches []nameIDType
	for _, item := range items {
		if item.name == name {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return nameIDType{}, fmt.Errorf("name not found")
	case 1:
		return matches[0], nil
	default:
		return nameIDType{}, fmt.Errorf("name is ambiguous within its parent")
	}
}

func resolveParentID[T itemOrder | itemOrderType](kind, parentName string, parentKey types.String,
	keyToID map[string]string, parents map[string]T,
) (string, error) {
	keyField := "section_key"
	if kind == "network rule" {
		keyField = "net_rule_key"
	}
	if hasConfiguredValue(parentKey) {
		key := parentKey.ValueString()
		id, ok := keyToID[key]
		if !ok {
			return "", fmt.Errorf("%s %q does not match a configured parent map key", keyField, key)
		}
		parent, ok := parents[id]
		if !ok {
			return "", fmt.Errorf("%s %q resolves to missing %s ID %q", keyField, key, kind, id)
		}
		if name := parentDisplayName(parent); name != parentName {
			return "", fmt.Errorf("%s %q identifies %s %q, not %q", keyField, key, kind, name, parentName)
		}
		return id, nil
	}

	var matches []string
	for id, parent := range parents {
		if parentDisplayName(parent) == parentName {
			matches = append(matches, id)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("%s name %q not found", kind, parentName)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("%s name %q is ambiguous; set %s to the parent map key",
			kind, parentName, keyField)
	}
}

func parentDisplayName[T itemOrder | itemOrderType](parent T) string {
	switch value := any(parent).(type) {
	case itemOrder:
		return value.parentName
	case itemOrderType:
		return value.parentName
	default:
		return ""
	}
}

// makePolicyIDMap creates a map of policy IDs to policy names.
func (r *lanRulesIndexResource) makePolicyIDMap(apiResult *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
) map[string]string {
	out := make(map[string]string)
	for _, sp := range apiResult.SubPolicies {
		pol := sp.GetPolicy()
		out[pol.GetID()] = pol.GetName()
	}
	return out
}

// parseAPISections takes API result and creates a list of sections in each policy.
// Sets the policySections[policyName].current[] list, i.e. what is currently in CMA.
// On error, adds the error to *diags
//
// policyName -> {current: [{sectName,id},{sectName,id},...]}
func (r *lanRulesIndexResource) parseAPISections(apiResult *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
	policyIDMap map[string]string, policySections map[string]itemOrder, diags *diag.Diagnostics,
) {
	for _, sec := range apiResult.Sections {
		section := sec.GetSection()
		policyName := "" // default main policy
		subPolicyID := section.GetSubPolicyID()
		if subPolicyID != nil { // get sub-policy name if defined
			polName, ok := policyIDMap[*subPolicyID]
			if !ok {
				diags.AddError("processing policy API response",
					fmt.Sprintf("subpolicy id '%s' not found in API response", *subPolicyID))
				return
			}
			policyName = polName
		} else {
			subPolicyID = new("")
		}
		iOrder := policySections[*subPolicyID]
		iOrder.parentID = *subPolicyID
		iOrder.parentName = policyName
		iOrder.current = append(iOrder.current, nameID{name: section.GetName(), id: section.GetID()})
		policySections[*subPolicyID] = iOrder
	}
}

// parseAPINetRules takes API result and creates a list of rules or sub-policies in each section.
// Sets the sectionRulesOrSubPols[sectionName].current[] list, i.e. what is currently in CMA.
// On error, adds the error to *diags
//
// sectionName -> {current: [{ruleName,id,RULE}, {subPolicyName,id,SUB_POLICY}, ...]}
func (r *lanRulesIndexResource) parseAPINetRules(apiResult *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
	sectionRulesOrSubPols map[string]itemOrderType,
) {
	for _, rul := range apiResult.Rules {
		rule := rul.GetRule()
		sectionID := rule.GetSection().GetID()
		iOrder := sectionRulesOrSubPols[sectionID]
		iOrder.parentID = sectionID
		iOrder.parentName = rule.GetSection().GetName()
		iOrder.current = append(iOrder.current, nameIDType{
			name:     rule.GetName(),
			id:       rule.GetID(),
			ruleType: r.ruleType(rul),
		})
		sectionRulesOrSubPols[sectionID] = iOrder
	}
}

// parseAPIFwRules takes API result and creates a list of firewall rules in each network rule.
// Sets the policyRules[sectionName].current[] list, i.e. what is currently in CMA.
// On error, adds the error to *diags
//
// netRule -> {current: [{fwRuleName,id}, {fwRuleName,id}, ...]}
func (r *lanRulesIndexResource) parseAPIFwRules(apiResult *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
	ruleFwRules map[string]itemOrder,
) {
	for _, rul := range apiResult.Rules {
		if r.ruleType(rul) != cato_models.PolicyRuleTypeEnumPolicyRule {
			continue
		}
		netRule := rul.GetRule()
		netRuleID := netRule.GetID()
		iOrder := ruleFwRules[netRuleID]
		iOrder.parentID = netRuleID
		iOrder.parentName = netRule.GetName()
		for _, firewall := range netRule.GetFirewall() {
			if firewall == nil {
				continue
			}
			iOrder.current = append(iOrder.current, nameID{
				name: firewall.GetRule().GetName(),
				id:   firewall.GetRule().GetID(),
			})
		}
		ruleFwRules[netRuleID] = iOrder
	}
}

func (r *lanRulesIndexResource) ruleType(ru *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules,
) cato_models.PolicyRuleTypeEnum {
	if t := ru.GetRuleType(); t != nil {
		return *t
	}
	return cato_models.PolicyRuleTypeEnum("")
}

// parsePlanSections takes TF state and creates a list of sections in each policy.
// Sets the policySections[policyName].target[] list, i.e. what is currently in state.
// On error, adds the error to *diags
//
// policyName -> {target: [{sectName,id},{sectName,id},...]}
func (r *lanRulesIndexResource) parsePlanSections(plan *LanFwRulesIndex,
	apiResult *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy, policySections map[string]itemOrder,
	sectionKeyToID map[string]string, diags *diag.Diagnostics,
) {
	type sectItem struct {
		name, id string
		index    int64
	}
	if plan == nil || !utils.HasValue(plan.SectionData) {
		return
	}

	checkIndexes := func(policyName string, items []sectItem) bool {
		for i, item := range items {
			if item.index != int64(i+1) {
				if policyName == "" {
					policyName = "<main>"
				}
				diags.AddError("error parsing plan, mismatched index_in_policy numbers",
					fmt.Sprintf("index for section '%s' in policy '%s' is %d but should be %d",
						item.name, policyName, item.index, i+1))
				return true
			}
		}
		return false
	}

	tfSections := make(map[string]LanFwSectionData)
	if utils.CheckErr(diags, plan.SectionData.ElementsAs(context.Background(), &tfSections, false)) {
		return
	}

	policyIDsByName := make(map[string][]string)
	policyIDsByName[""] = []string{""}
	for _, subPolicy := range apiResult.SubPolicies {
		policy := subPolicy.GetPolicy()
		policyIDsByName[policy.GetName()] = append(policyIDsByName[policy.GetName()], policy.GetID())
	}

	// make a helper map: map[policyID][]{section_name,id,index}
	sectionIndexes := make(map[string][]sectItem)
	for sectionKey, section := range tfSections {
		sectionName := configuredName(section.SectionName, sectionKey)
		policyName := section.SubPolicyName.ValueString()
		policyIDs := policyIDsByName[policyName]
		if len(policyIDs) == 0 {
			diags.AddError("error parsing plan",
				fmt.Sprintf("cannot find sub-policy %q for section map key %q", policyName, sectionKey))
			return
		}
		if len(policyIDs) > 1 {
			diags.AddError("error parsing plan",
				fmt.Sprintf("sub-policy name %q is ambiguous for section map key %q", policyName, sectionKey))
			return
		}
		policyID := policyIDs[0]
		sectionIndex := section.SectionIndex.ValueInt64()
		sectionID, err := resolveNameID(sectionName, section.ID.ValueString(), policySections[policyID].current)
		if err != nil {
			diags.AddError("error parsing plan",
				fmt.Sprintf("cannot resolve section map key %q (section_name %q) in sub-policy %q: %v",
					sectionKey, sectionName, policyName, err))
			return
		}
		sectionKeyToID[sectionKey] = sectionID
		sectionIndexes[policyID] = append(sectionIndexes[policyID],
			sectItem{name: sectionName, id: sectionID, index: sectionIndex})
	}

	// sort the sections in each policy by sectionIndex
	// and update policySections[].target
	for policyID, sectSlice := range sectionIndexes {
		slices.SortFunc(sectSlice, func(a, b sectItem) int { return cmp.Compare(a.index, b.index) })
		policyName := policySections[policyID].parentName
		if checkIndexes(policyName, sectSlice) {
			return
		}
		item := policySections[policyID]
		item.target = make([]nameID, len(sectSlice))
		for i, sect := range sectSlice {
			item.target[i] = nameID{name: sect.name, id: sect.id}
		}
		policySections[policyID] = item
	}
}

// checkSections ensures the policySections contains the same sections in current and target lists
func (r *lanRulesIndexResource) checkSections(policySections map[string]itemOrder, diags *diag.Diagnostics) {
	const summary = "LAN firewall section validation failed"
	for _, sections := range policySections {
		policyName := sections.parentName
		if policyName == "" {
			policyName = "<main>"
		}
		if len(sections.current) != len(sections.target) {
			diags.AddError(summary, fmt.Sprintf("policy %q has %d current sections but %d target sections",
				policyName, len(sections.current), len(sections.target)),
			)
		}

		currentIDs := make(map[string]struct{}, len(sections.current))
		targetIDs := make(map[string]struct{}, len(sections.target))
		for _, current := range sections.current {
			currentIDs[current.id] = struct{}{}
		}
		for _, t := range sections.target {
			targetIDs[t.id] = struct{}{}
			if _, ok := currentIDs[t.id]; !ok {
				diags.AddError(summary, fmt.Sprintf("planned section '%s' not found in policy '%s'", t.name, policyName))
			}
		}
		for _, current := range sections.current {
			if _, ok := targetIDs[current.id]; !ok {
				diags.AddError(summary, fmt.Sprintf("current section '%s' missing from plan for policy '%s'",
					current.name, policyName))
			}
		}
	}
}

// checkNetRules ensures the sectionRules contains the same rules or sub-policies in current and target lists
func (r *lanRulesIndexResource) checkNetRules(sectionRules map[string]itemOrderType, diags *diag.Diagnostics) {
	for _, rules := range sectionRules {
		sectionName := rules.parentName
		if len(rules.current) != len(rules.target) {
			diags.AddError(
				"LAN firewall rule validation failed",
				fmt.Sprintf("section %q has %d current rules or sub-policies but %d target rules or sub-policies",
					sectionName, len(rules.current), len(rules.target)),
			)
		}

		currentIDs := make(map[string]nameIDType, len(rules.current))
		for _, rule := range rules.current {
			currentIDs[rule.id] = rule
		}
		targetIDs := make(map[string]nameIDType, len(rules.target))
		for _, rule := range rules.target {
			targetIDs[rule.id] = rule
		}

		for id, rule := range currentIDs {
			if _, ok := targetIDs[id]; !ok {
				diags.AddError("LAN firewall rule validation failed",
					fmt.Sprintf("section %q has current rule or sub-policy %q missing from target",
						sectionName, rule.name),
				)
			}
		}
		for id, rule := range targetIDs {
			if _, ok := currentIDs[id]; !ok {
				diags.AddError("LAN firewall rule validation failed",
					fmt.Sprintf("section %q has target rule or sub-policy %q not found in current policy",
						sectionName, rule.name),
				)
			}
		}
	}
}

// checkFwRules ensures the ruleFwRules contains the same fwRules in current and target lists
func (r *lanRulesIndexResource) checkFwRules(ruleFwRules map[string]itemOrder, diags *diag.Diagnostics) {
	for _, fwRules := range ruleFwRules {
		netRuleName := fwRules.parentName
		if len(fwRules.current) != len(fwRules.target) {
			diags.AddError(
				"LAN firewall rules validation failed",
				fmt.Sprintf("network rule %q has %d current firewall rules but %d target firewall rules",
					netRuleName, len(fwRules.current), len(fwRules.target)),
			)
		}

		currentIDs := make(map[string]nameID, len(fwRules.current))
		for _, fwRule := range fwRules.current {
			currentIDs[fwRule.id] = fwRule
		}
		targetIDs := make(map[string]nameID, len(fwRules.target))
		for _, fwRule := range fwRules.target {
			targetIDs[fwRule.id] = fwRule
		}

		for id, rule := range currentIDs {
			if _, ok := targetIDs[id]; !ok {
				diags.AddError("LAN firewall rule validation failed",
					fmt.Sprintf("network rule %q has current firewall rule %q missing from target",
						netRuleName, rule.name),
				)
			}
		}
		for id, rule := range targetIDs {
			if _, ok := currentIDs[id]; !ok {
				diags.AddError("LAN firewall rule validation failed",
					fmt.Sprintf("network rule %q has target firewall rule %q not found in current policy",
						netRuleName, rule.name),
				)
			}
		}
	}
}

// parsePlanNetRules takes TF state and creates a list of rules/sub-policies in each section.
// Sets the sectionRulesOrSubPols[sectionName].target[] list, i.e. what is currently in state.
// On error, adds the error to *diags
//
// sectionName -> {target: [{ruleName,id,RULE},{subPolName,id,SUB_POL},...]}
func (r *lanRulesIndexResource) parsePlanNetRules(plan *LanFwRulesIndex, sectionKeyToID map[string]string,
	sectionRulesOrSubPols map[string]itemOrderType, netRuleKeyToID, netRuleParentKeyByID map[string]string,
	diags *diag.Diagnostics,
) {
	type ruleItem struct {
		name, id string
		ruleType cato_models.PolicyRuleTypeEnum
		index    int64
	}
	if plan == nil || !utils.HasValue(plan.NetworkRules) {
		return
	}
	checkIndexes := func(sectionName string, items []ruleItem) bool {
		for i, item := range items {
			if item.index != int64(i+1) {
				diags.AddError("error parsing plan, mismatched index_in_section numbers",
					fmt.Sprintf("index for rule '%s' in section '%s' is %d but should be %d",
						item.name, sectionName, item.index, i+1))
				return true
			}
		}
		return false
	}

	tfRuleData := make(map[string]LanNetworkRule)
	if utils.CheckErr(diags, plan.NetworkRules.ElementsAs(context.Background(), &tfRuleData, false)) {
		return
	}

	// make a helper map: map[sectionID][]{rule_or_subpol_name,id,type,index}
	ruleIndexes := make(map[string][]ruleItem)
	for ruleKey, rule := range tfRuleData {
		ruleOrSPolName := configuredName(rule.RuleName, ruleKey)
		sectionName := rule.SectionName.ValueString()
		sectionID, err := resolveParentID(
			"section", sectionName, rule.SectionKey, sectionKeyToID, sectionRulesOrSubPols,
		)
		if err != nil {
			diags.AddError("error parsing plan",
				fmt.Sprintf("cannot resolve parent for network_rules map key %q: %v", ruleKey, err))
			return
		}
		ruleInfo, err := resolveNameIDType(
			ruleOrSPolName, rule.ID.ValueString(), sectionRulesOrSubPols[sectionID].current,
		)
		if err != nil {
			diags.AddError("error parsing plan",
				fmt.Sprintf("cannot resolve network_rules map key %q (rule_name %q) in section %q: %v",
					ruleKey, ruleOrSPolName, sectionName, err))
			return
		}
		if existingKey := invertKeyToID(netRuleKeyToID)[ruleInfo.id]; existingKey != "" {
			diags.AddError("error parsing plan",
				fmt.Sprintf("network_rules map keys %q and %q resolve to the same rule", existingKey, ruleKey))
			return
		}
		netRuleKeyToID[ruleKey] = ruleInfo.id
		if hasConfiguredValue(rule.SectionKey) {
			netRuleParentKeyByID[ruleInfo.id] = rule.SectionKey.ValueString()
		}
		ruleIndexes[sectionID] = append(ruleIndexes[sectionID],
			ruleItem{
				name:     ruleOrSPolName,
				id:       ruleInfo.id,
				ruleType: ruleInfo.ruleType,
				index:    rule.IndexInSection.ValueInt64(),
			})
	}

	// sort the rules/sub-policies in each section by index
	// and update policySections[].target
	for sectionID, ruleSlice := range ruleIndexes {
		slices.SortFunc(ruleSlice, func(a, b ruleItem) int { return cmp.Compare(a.index, b.index) })
		sectionName := sectionRulesOrSubPols[sectionID].parentName
		if checkIndexes(sectionName, ruleSlice) {
			return
		}
		item := sectionRulesOrSubPols[sectionID]
		item.target = make([]nameIDType, len(ruleSlice))
		for i, rl := range ruleSlice {
			item.target[i] = nameIDType{name: rl.name, id: rl.id, ruleType: rl.ruleType}
		}
		sectionRulesOrSubPols[sectionID] = item
	}
}

// parsePlanFwRules takes TF state and creates a list of firewall rules in each network rule.
// Sets the ruleFwRules[netRuleName].target[] list, i.e. what is currently in state.
// On error, adds the error to *diags
//
// netRuleName -> {target: [{fwRuleName,id},{fwRuleName,id},...]}
func (r *lanRulesIndexResource) parsePlanFwRules(plan *LanFwRulesIndex, netRuleKeyToID map[string]string,
	ruleFwRules map[string]itemOrder, firewallRuleKeyToID, firewallParentKeyByID map[string]string,
	diags *diag.Diagnostics,
) {
	type ruleItem struct {
		name, id string
		index    int64
	}
	if plan == nil || !utils.HasValue(plan.FirewallRules) {
		return
	}
	checkIndexes := func(netRuleName string, items []ruleItem) bool {
		for i, item := range items {
			if item.index != int64(i+1) {
				diags.AddError("error parsing plan, mismatched index_in_rule numbers",
					fmt.Sprintf("index for firewall rule '%s' in network rule '%s' is %d but should be %d",
						item.name, netRuleName, item.index, i+1))
				return true
			}
		}
		return false
	}

	tfRuleData := make(map[string]LanFirewallRule)
	if utils.CheckErr(diags, plan.FirewallRules.ElementsAs(context.Background(), &tfRuleData, false)) {
		return
	}

	// make a helper map: map[netRuleID][]{rule_name,id,index}
	ruleIndexes := make(map[string][]ruleItem)
	for fwRuleKey, fwRule := range tfRuleData {
		fwRuleName := configuredName(fwRule.FirewallRuleName, fwRuleKey)
		netRuleName := fwRule.NetRuleName.ValueString()
		netRuleID, err := resolveParentID(
			"network rule", netRuleName, fwRule.NetRuleKey, netRuleKeyToID, ruleFwRules,
		)
		if err != nil {
			diags.AddError("error parsing plan",
				fmt.Sprintf("cannot resolve parent for firewall_rules map key %q: %v", fwRuleKey, err))
			return
		}
		fwRuleID, err := resolveNameID(fwRuleName, fwRule.ID.ValueString(), ruleFwRules[netRuleID].current)
		if err != nil {
			diags.AddError("error parsing plan",
				fmt.Sprintf("cannot resolve firewall_rules map key %q (firewall_rule_name %q) in network rule %q: %v",
					fwRuleKey, fwRuleName, netRuleName, err))
			return
		}
		if existingKey := invertKeyToID(firewallRuleKeyToID)[fwRuleID]; existingKey != "" {
			diags.AddError("error parsing plan",
				fmt.Sprintf("firewall_rules map keys %q and %q resolve to the same firewall rule",
					existingKey, fwRuleKey))
			return
		}
		firewallRuleKeyToID[fwRuleKey] = fwRuleID
		if hasConfiguredValue(fwRule.NetRuleKey) {
			firewallParentKeyByID[fwRuleID] = fwRule.NetRuleKey.ValueString()
		}
		ruleIndexes[netRuleID] = append(ruleIndexes[netRuleID],
			ruleItem{
				name:  fwRuleName,
				id:    fwRuleID,
				index: fwRule.IndexInRule.ValueInt64(),
			})
	}

	// sort the rules in each netRule by index
	// and update ruleFwRules[].target
	for netRuleID, fwRuleSlice := range ruleIndexes {
		slices.SortFunc(fwRuleSlice, func(a, b ruleItem) int { return cmp.Compare(a.index, b.index) })
		netRuleName := ruleFwRules[netRuleID].parentName
		if checkIndexes(netRuleName, fwRuleSlice) {
			return
		}
		item := ruleFwRules[netRuleID]
		item.target = make([]nameID, len(fwRuleSlice))
		for i, rl := range fwRuleSlice {
			item.target[i] = nameID{name: rl.name, id: rl.id}
		}
		ruleFwRules[netRuleID] = item
	}
}

// convertMoveFwRuleToPositionValues converts between the Terraform model and SDK values.
func convertMoveFwRuleToPositionValues(
	parentRuleID string,
	currentRules []nameID,
	fwRuleID string,
	newPosition int,
	myRule nameID,
	curPosition int,
) (resultInput cato_models.PolicyMoveSubRuleInput) {
	for i := curPosition; i < newPosition; i++ {
		currentRules[i] = currentRules[i+1]
	}
	currentRules[newPosition] = myRule

	// Prepare input for API to move the sub-policy
	resultInput = cato_models.PolicyMoveSubRuleInput{
		ID: fwRuleID,
		To: &cato_models.PolicySubRulePositionInput{},
	}
	if newPosition == len(currentRules)-1 { // move to LAST_IN_RULE
		resultInput.To.Position = cato_models.PolicySubRulePositionEnumLastInRule
		resultInput.To.Ref = parentRuleID
	} else {
		resultInput.To.Position = cato_models.PolicySubRulePositionEnumBeforeSubRule
		resultInput.To.Ref = currentRules[newPosition+1].id
	}
	return resultInput
}

// convertMoveSubPolicyToPositionValues converts between the Terraform model and SDK values.
func convertMoveSubPolicyToPositionValues(
	parentSectionID string,
	currentRulesOrSubPols []nameIDType,
	policyID string,
	newPosition int,
	myPolicy nameIDType,
	curPosition int,
) (resultInput cato_models.PolicyMoveRuleInput) {
	for i := curPosition; i < newPosition; i++ {
		currentRulesOrSubPols[i] = currentRulesOrSubPols[i+1]
	}
	currentRulesOrSubPols[newPosition] = myPolicy

	// Prepare input for API to move the sub-policy
	resultInput = cato_models.PolicyMoveRuleInput{
		ID: policyID,
		To: &cato_models.PolicyRulePositionInput{},
	}
	if newPosition == len(currentRulesOrSubPols)-1 { // move to LAST_IN_SECTION
		resultInput.To.Position = new(cato_models.PolicyRulePositionEnumLastInSection)
		resultInput.To.Ref = new(parentSectionID)
	} else {
		resultInput.To.Position = new(cato_models.PolicyRulePositionEnumBeforeRule)
		resultInput.To.Ref = new(currentRulesOrSubPols[newPosition+1].id)
	}
	return resultInput
}

// convertMoveNetRuleToPositionValues converts between the Terraform model and SDK values.
func convertMoveNetRuleToPositionValues(
	parentSectionID string,
	currentRulesOrSubPols []nameIDType,
	ruleID string,
	newPosition int,
	myRule nameIDType,
	curPosition int,
) (resultInput cato_models.PolicyMoveRuleInput) {
	for i := curPosition; i < newPosition; i++ {
		currentRulesOrSubPols[i] = currentRulesOrSubPols[i+1]
	}
	currentRulesOrSubPols[newPosition] = myRule

	// Prepare input for API to move the network rule
	resultInput = cato_models.PolicyMoveRuleInput{
		ID: ruleID,
		To: &cato_models.PolicyRulePositionInput{},
	}
	if newPosition == len(currentRulesOrSubPols)-1 { // move to LAST_IN_SECTION
		resultInput.To.Position = new(cato_models.PolicyRulePositionEnumLastInSection)
		resultInput.To.Ref = new(parentSectionID)
	} else {
		resultInput.To.Position = new(cato_models.PolicyRulePositionEnumBeforeRule)
		resultInput.To.Ref = new(currentRulesOrSubPols[newPosition+1].id)
	}
	return resultInput
}

// convertMoveSectionToPositionValues converts between the Terraform model and SDK values.
func convertMoveSectionToPositionValues(
	parentPolicyID string,
	currentSections []nameID,
	sectionID string,
	newPosition int,
	mySection nameID,
	curPosition int,
) (resultInput cato_models.PolicyMoveSectionInput) {
	for i := curPosition; i < newPosition; i++ {
		currentSections[i] = currentSections[i+1]
	}
	currentSections[newPosition] = mySection

	// Prepare input for API to move the section
	resultInput = cato_models.PolicyMoveSectionInput{
		ID: sectionID,
		To: &cato_models.PolicySectionPositionInput{},
	}
	if newPosition == len(currentSections)-1 { // move to LAST_IN_POLICY
		resultInput.To.Position = cato_models.PolicySectionPositionEnumLastInPolicy
		if parentPolicyID != "" { // it is a sub-policy
			resultInput.To.Ref = new(parentPolicyID)
		}
	} else {
		resultInput.To.Position = cato_models.PolicySectionPositionEnumBeforeSection
		resultInput.To.Ref = new(currentSections[newPosition+1].id)
	}
	return resultInput
}
