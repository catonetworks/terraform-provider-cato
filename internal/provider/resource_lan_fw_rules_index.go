package provider

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
	client       *catoClientData
	catov2Client LanFwRuleClient
}

type LanFwRuleClient interface {
	PolicySocketLanPolicy(ctx context.Context, accountID string, socketLanPolicyInput *cato_models.SocketLanPolicyInput,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanPolicy, error)
	PolicySocketLanMoveSection(ctx context.Context, policyMoveSectionInput cato_models.PolicyMoveSectionInput,
		accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanMoveSection, error)
	PolicySocketLanMoveRule(ctx context.Context, policyMoveRuleInput cato_models.PolicyMoveRuleInput, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanMoveRule, error)
	PolicySocketLanFirewallMoveRule(ctx context.Context, accountID string,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyMoveSubRuleInput cato_models.PolicyMoveSubRuleInput, interceptors ...clientv2.RequestInterceptor) (
		*cato_go_sdk.PolicySocketLanFirewallMoveRule, error)
	PolicySocketLanPublishPolicyRevision(ctx context.Context,
		socketLanPolicyMutationInput *cato_models.SocketLanPolicyMutationInput,
		policyPublishRevisionInput *cato_models.PolicyPublishRevisionInput, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error)
}

func (r *lanRulesIndexResource) getClient() LanFwRuleClient { return r.catov2Client }

func (r *lanRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_lf_move_rule"
}

func (r *lanRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ordering of LAN firewall policy items.\n\n" +
			"**LAN Firewall Policy**\n" +
			"- Section\n" +
			"  - Network rule\n" +
			"    - Firewall rule\n" +
			"  - Sub-policy\n" +
			"    - Section\n" +
			"      - Network rule\n" +
			"        - Firewall rule",

		Attributes: map[string]schema.Attribute{
			"section_data": schema.MapNestedAttribute{
				Description: "Map of section indexes keyed by section name",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:   "Section id",
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
						},
						"section_index": schema.Int64Attribute{
							Description: "Position of the section in the policy or sub-policy. Starts with 1.",
							Required:    true,
						},
						"sub_policy_name": schema.StringAttribute{
							Description: "Sub-policy name. If not set, the main policy is used.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
						},
					},
				},
			},
			"network_rules": schema.MapNestedAttribute{
				Description: "Map of network rule or sub-policy index for each section, keyed by rule or sub-policy name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:   "Network rule ID or sub-policy scope rule ID",
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
						},
						"rule_type": schema.StringAttribute{
							Description: "Rule type: POLICY_RULE, SUB_POLICY_SCOPE, SUB_RULE",
							Computed:    true,
						},
						"section_name": schema.StringAttribute{
							Description: "LAN section name housing rule",
							Required:    true,
						},
						"index_in_section": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    true,
						},
					},
				},
			},
			"firewall_rules": schema.MapNestedAttribute{
				Description: "Map of firewall rule index for each network rule, keyed by firewall rule name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:   "Firewall rule ID",
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
						},
						"net_rule_name": schema.StringAttribute{
							Description: "Parent LAN network rule name",
							Required:    true,
						},
						"index_in_rule": schema.Int64Attribute{
							Description: "Index value remapped per network rule",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *lanRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
	r.catov2Client = r.client.catov2
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

type lfIndexMap struct {
	// Section order in given policy
	// map[policyName]{
	//   current:[{sectionName:secA,id:100},{sectionName:secB,id:101},...],
	//   target: [{sectionName:secB,id:101},{sectionName:secA,id:100},...]}
	sections map[string]itemOrder

	// Rule or subPolicy order in given section
	// map[sectionName]{
	//   current:[{name:ruleA,id:100,typ:rule},{name:subPolA,id:123,typ:subPolicy},...],
	//   target: [{name:subPolA,id:123,typ:subPolicy},{name:ruleA,id:100,typ:rule},...],
	rulesOrSubPols map[string]itemOrderType

	// Firewall rule order in given network rule
	// map[networkRuleName]{
	//   current:[{name:FwRuleA,id:100},{name:FwRuleB,id:123},...],
	//   target: [{name:FwRuleB,id:123},{name:FwRuleA,id:100},...],
	firewallRules map[string]itemOrder
}
type itemOrder struct {
	parentID string
	current  []nameID
	target   []nameID
}
type nameID struct {
	name, id string
}
type itemOrderType struct {
	parentID string
	current  []nameIDType
	target   []nameIDType
}
type nameIDType struct {
	name, id string
	ruleType cato_models.PolicyRuleTypeEnum
}

func (r *lanRulesIndexResource) hydrateLanFwRulesIndex(ctx context.Context, plan *LanFwRulesIndex, diags *diag.Diagnostics,
) (newState *LanFwRulesIndex, indexMap *lfIndexMap) {
	// Call Cato API to get the policy
	result, err := r.getClient().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("failed to hydrate sub-policy", err.Error())
		return nil, nil
	}
	policyBase := result.GetPolicy().GetSocketLan().GetPolicy()

	policySections, sectionData := r.hydrateSections(ctx, plan, policyBase, diags)
	sectionRulesOrSubPols, netRuleData := r.hydrateNetRules(ctx, plan, policyBase, diags)
	ruleFirewallRules, firewallRuleData := r.hydrateFirewallRules(ctx, plan, policyBase, diags)

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

func (r *lanRulesIndexResource) hydrateSections(ctx context.Context, plan *LanFwRulesIndex,
	policyBase *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy, diags *diag.Diagnostics,
) (policySections map[string]itemOrder, sectionData types.Map) {
	policySections = make(map[string]itemOrder)
	sectionNameMap := r.makeSectionNameMap(policyBase) // map[sectionName]sectionID
	policyIDMap := r.makePolicyIDMap(policyBase)       // map[policyID]policyName
	sectionDataNull := types.MapNull(types.ObjectType{AttrTypes: LanFwSectionDataTypes})

	// Prepare section map
	//   map[policyName]{  // "" means the main policy
	//   	current:[{secName,secID},...]   // from API response
	//   	target: [{secName,secID},...]   // from TF plan
	//   }
	r.parseAPISections(policyBase, policyIDMap, policySections, diags)
	if diags.HasError() { // current
		return nil, sectionDataNull
	}
	r.parsePlanSections(plan, sectionNameMap, policySections, diags)
	if diags.HasError() { // target
		return nil, sectionDataNull
	}
	r.checkSections(policySections, diags)
	if diags.HasError() { // should contain the same items
		return nil, sectionDataNull
	}

	// create TF SectionData:  map[policyName]{secID,secIndex,subPolicyName} -> types.Map
	sections := make(map[string]types.Object)
	for polName, sectionLists := range policySections {
		for i, section := range sectionLists.current {
			tfSection := LanFwSectionData{
				ID:            types.StringValue(section.id),
				SectionIndex:  types.Int64Value(int64(i + 1)), // 1-based
				SubPolicyName: types.StringValue(polName),
			}
			sectionObj, objDiags := types.ObjectValueFrom(ctx, LanFwSectionDataTypes, tfSection)
			diags.Append(objDiags...)
			if diags.HasError() {
				return nil, sectionDataNull
			}
			sections[section.name] = sectionObj
		}
	}
	sd, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: LanFwSectionDataTypes}, sections)
	if objDiags.HasError() {
		diags.Append(objDiags...)
		return nil, sectionDataNull
	}
	sectionData = sd
	return policySections, sectionData
}

func (r *lanRulesIndexResource) hydrateNetRules(ctx context.Context, plan *LanFwRulesIndex,
	policyBase *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy, diags *diag.Diagnostics,
) (sectionRulesOrSubPols map[string]itemOrderType, netRuleData types.Map) {
	sectionRulesOrSubPols = make(map[string]itemOrderType)
	ruleNameMap := r.makeRuleNameMap(policyBase) // map[netRuleName]nameIDType
	netRuleDataNull := types.MapNull(types.ObjectType{AttrTypes: LanNetworkRuleTypes})

	// Prepare rule and subpolicy map
	//   map[sectionName]{
	//   	current:[{ruleName,id,RULE},{subPolicyName,id,SUB_POL},...]  // from API response
	//   	target: [{ruleName,id,RULE},{subPolicyName,id,SUB_POL},...]  // from TF plan
	//   }
	r.parseAPINetRules(policyBase, sectionRulesOrSubPols) // current
	if diags.HasError() {
		return nil, netRuleDataNull
	}
	r.parsePlanNetRules(plan, ruleNameMap, sectionRulesOrSubPols, diags) // target
	if diags.HasError() {
		return nil, netRuleDataNull
	}
	r.checkNetRules(sectionRulesOrSubPols, diags) // should contain the same items
	if diags.HasError() {
		return nil, netRuleDataNull
	}

	// create TF NetworkRuleData:  map[rule/subPolicy name]{ruleID,sectName,sectIndex,ruleType} -> types.Map
	rulesOrSubPols := make(map[string]types.Object)
	for sectName, rList := range sectionRulesOrSubPols {
		for i, rule := range rList.current {
			tfRuleData := LanNetworkRule{
				ID:             types.StringValue(rule.id),
				RuleType:       types.StringValue(string(rule.ruleType)),
				SectionName:    types.StringValue(sectName),
				IndexInSection: types.Int64Value(int64(i + 1)), // 1-based
			}
			ruleObj, objDiags := types.ObjectValueFrom(ctx, LanNetworkRuleTypes, tfRuleData)
			diags.Append(objDiags...)
			if diags.HasError() {
				return nil, netRuleDataNull
			}
			rulesOrSubPols[rule.name] = ruleObj
		}
	}

	rd, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: LanNetworkRuleTypes}, rulesOrSubPols)
	if objDiags.HasError() {
		diags.Append(objDiags...)
		return nil, netRuleDataNull
	}
	netRuleData = rd
	return sectionRulesOrSubPols, netRuleData
}

func (r *lanRulesIndexResource) hydrateFirewallRules(ctx context.Context, plan *LanFwRulesIndex,
	policyBase *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy, diags *diag.Diagnostics,
) (ruleFirewallRules map[string]itemOrder, firewallRuleData types.Map) {
	ruleFirewallRules = make(map[string]itemOrder)
	ruleNameMap := r.makeFwRuleNameMap(policyBase) // map[fwRuleName]ruleID
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
	r.parsePlanFwRules(plan, ruleNameMap, ruleFirewallRules, diags) // target
	if diags.HasError() {
		return nil, firewallRuleDataNull
	}
	r.checkFwRules(ruleFirewallRules, diags) // should contain the same items
	if diags.HasError() {
		return nil, firewallRuleDataNull
	}

	// create TF FirewallRuleData:  map[rule name]{fwRuleID,netRuleName,Index} -> types.Map
	tfFirewallRules := make(map[string]types.Object)
	for netRuleName, rList := range ruleFirewallRules {
		for i, rule := range rList.current {
			tfRuleData := LanFirewallRule{
				ID:          types.StringValue(rule.id),
				NetRuleName: types.StringValue(netRuleName),
				IndexInRule: types.Int64Value(int64(i + 1)), // 1-based
			}
			ruleObj, objDiags := types.ObjectValueFrom(ctx, LanFirewallRuleTypes, tfRuleData)
			diags.Append(objDiags...)
			if diags.HasError() {
				return nil, firewallRuleDataNull
			}
			tfFirewallRules[rule.name] = ruleObj
		}
	}
	rd, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: LanFirewallRuleTypes}, tfFirewallRules)
	if objDiags.HasError() {
		diags.Append(objDiags...)
		return nil, firewallRuleDataNull
	}
	firewallRuleData = rd

	return ruleFirewallRules, firewallRuleData
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

// makeSectionNameMap creates a map of section names to section IDs
func (r *lanRulesIndexResource) makeSectionNameMap(apiResult *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
) map[string]string {
	out := make(map[string]string)
	for _, sp := range apiResult.Sections {
		sect := sp.GetSection()
		out[sect.GetName()] = sect.GetID()
	}
	return out
}

// makeRuleNameMap creates a map of net-rule or sub-policy names to net-rule IDs or sub-policy scope rule IDs
func (r *lanRulesIndexResource) makeRuleNameMap(apiResult *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
) map[string]nameIDType {
	out := make(map[string]nameIDType)
	for _, rul := range apiResult.Rules {
		rule := rul.GetRule()
		out[rule.GetName()] = nameIDType{
			name:     rule.GetName(),
			id:       rule.GetID(),
			ruleType: r.ruleType(rul),
		}
	}
	return out
}

// makeFwRuleNameMap creates a map of firewall rule names to IDs
func (r *lanRulesIndexResource) makeFwRuleNameMap(apiResult *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
) map[string]string {
	out := make(map[string]string)
	for _, rul := range apiResult.Rules {
		rule := rul.GetRule()
		for _, fr := range rule.GetFirewall() {
			if fr == nil {
				continue
			}
			fRule := fr.GetRule()
			if fRule.GetName() == "" {
				continue
			}
			out[fRule.GetName()] = fRule.GetID()
		}
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
		iOrder := policySections[policyName]
		iOrder.parentID = *subPolicyID
		iOrder.current = append(iOrder.current, nameID{name: section.GetName(), id: section.GetID()})
		policySections[policyName] = iOrder
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
		sectionName := rule.GetSection().GetName()
		iOrder := sectionRulesOrSubPols[sectionName]
		iOrder.parentID = rule.GetSection().GetID()
		iOrder.current = append(iOrder.current, nameIDType{
			name:     rule.GetName(),
			id:       rule.GetID(),
			ruleType: r.ruleType(rul),
		})
		sectionRulesOrSubPols[sectionName] = iOrder
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
		netRuleName := netRule.GetName()
		iOrder := ruleFwRules[netRuleName]
		iOrder.parentID = netRule.GetID()
		for _, firewall := range netRule.GetFirewall() {
			iOrder.current = append(iOrder.current, nameID{
				name: firewall.GetRule().GetName(),
				id:   firewall.GetRule().GetID(),
			})
		}
		ruleFwRules[netRuleName] = iOrder
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
func (r *lanRulesIndexResource) parsePlanSections(plan *LanFwRulesIndex, sectionNameMap map[string]string,
	policySections map[string]itemOrder, diags *diag.Diagnostics,
) {
	type sectItem struct {
		name, id string
		index    int64
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

	if plan == nil || !utils.HasValue(plan.SectionData) {
		return
	}

	tfSections := make(map[string]LanFwSectionData)
	if utils.CheckErr(diags, plan.SectionData.ElementsAs(context.Background(), &tfSections, false)) {
		return
	}

	// make a helper map:    map[policyName][]{section_name,id,index}
	sectionIndexes := make(map[string][]sectItem)
	for sectionName, section := range tfSections {
		policyName := section.SubPolicyName.ValueString()
		sectionIndex := section.SectionIndex.ValueInt64()
		sectionID := sectionNameMap[sectionName]
		if sectionID == "" {
			diags.AddError("error parsing plan", "failed to find ID for section "+sectionName)
			return
		}
		sectionIndexes[policyName] = append(sectionIndexes[policyName],
			sectItem{name: sectionName, id: sectionID, index: sectionIndex})
	}

	// sort the sections in each policy by sectionIndex
	// and update policySections[].target
	for policyName, sectSlice := range sectionIndexes {
		slices.SortFunc(sectSlice, func(a, b sectItem) int { return cmp.Compare(a.index, b.index) })
		if checkIndexes(policyName, sectSlice) {
			return
		}
		item := policySections[policyName]
		item.target = make([]nameID, len(sectSlice))
		for i, sect := range sectSlice {
			item.target[i] = nameID{name: sect.name, id: sect.id}
		}
		policySections[policyName] = item
	}
}

// checkSections ensures the policySections contains the same sections in current and target lists
func (r *lanRulesIndexResource) checkSections(policySections map[string]itemOrder, diags *diag.Diagnostics) {
	const summary = "LAN firewall section validation failed"
	mkMap := func(items []nameID, policy string, diags *diag.Diagnostics) map[string]struct{} {
		m := make(map[string]struct{}, len(items))
		for _, i := range items {
			if _, ok := m[i.name]; ok {
				diags.AddError(summary, fmt.Sprintf("duplicate section name %q in policy %s", i.name, policy))
			}
			m[i.name] = struct{}{}
		}
		return m
	}

	for policyName, sections := range policySections {
		if policyName == "" {
			policyName = "<main>"
		}
		if len(sections.current) != len(sections.target) {
			diags.AddError(summary, fmt.Sprintf("policy %q has %d current sections but %d target sections",
				policyName, len(sections.current), len(sections.target)),
			)
		}

		currentNames := mkMap(sections.current, policyName, diags)
		mkMap(sections.target, policyName, diags)
		if diags.HasError() {
			return
		}

		for _, t := range sections.target {
			if _, ok := currentNames[t.name]; !ok {
				diags.AddError(summary, fmt.Sprintf("planned section '%s' not found in policy '%s'", t.name, policyName))
			}
		}
	}
}

// checkNetRules ensures the sectionRules contains the same rules or sub-policies in current and target lists
func (r *lanRulesIndexResource) checkNetRules(sectionRules map[string]itemOrderType, diags *diag.Diagnostics) {
	for sectionName, rules := range sectionRules {
		if len(rules.current) != len(rules.target) {
			diags.AddError(
				"LAN firewall rule validation failed",
				fmt.Sprintf("section %q has %d current rules or sub-policies but %d target rules or sub-policies",
					sectionName, len(rules.current), len(rules.target)),
			)
		}

		currentNames := make(map[string]int, len(rules.current))
		for _, rule := range rules.current {
			currentNames[rule.name]++
		}
		targetNames := make(map[string]int, len(rules.target))
		for _, rule := range rules.target {
			targetNames[rule.name]++
		}

		for name, count := range currentNames {
			if targetNames[name] != count {
				diags.AddError("LAN firewall rule validation failed",
					fmt.Sprintf("section %q has current rule or sub-policy %q but target contains it %d times",
						sectionName, name, targetNames[name]),
				)
			}
		}
		for name, count := range targetNames {
			if currentNames[name] != count {
				diags.AddError("LAN firewall rule validation failed",
					fmt.Sprintf("section %q has target rule or sub-policy %q but current contains it %d times",
						sectionName, name, currentNames[name]),
				)
			}
		}
	}
}

// checkFwRules ensures the ruleFwRules contains the same fwRules in current and target lists
func (r *lanRulesIndexResource) checkFwRules(ruleFwRules map[string]itemOrder, diags *diag.Diagnostics) {
	for netRuleName, fwRules := range ruleFwRules {
		if len(fwRules.current) != len(fwRules.target) {
			diags.AddError(
				"LAN firewall rules validation failed",
				fmt.Sprintf("network rule %q has %d current firewall rules but %d target firewall rules",
					netRuleName, len(fwRules.current), len(fwRules.target)),
			)
		}

		currentNames := make(map[string]int, len(fwRules.current))
		for _, fwRule := range fwRules.current {
			currentNames[fwRule.name]++
		}
		targetNames := make(map[string]int, len(fwRules.target))
		for _, fwRule := range fwRules.target {
			targetNames[fwRule.name]++
		}

		for name, count := range currentNames {
			if targetNames[name] != count {
				diags.AddError("LAN firewall rule validation failed",
					fmt.Sprintf("network rule %q has current firewal rules %q but target contains it %d times",
						netRuleName, name, targetNames[name]),
				)
			}
		}
		for name, count := range targetNames {
			if currentNames[name] != count {
				diags.AddError("LAN firewall rule validation failed",
					fmt.Sprintf("network rule %q has target section %q but current contains it %d times",
						netRuleName, name, currentNames[name]),
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
func (r *lanRulesIndexResource) parsePlanNetRules(plan *LanFwRulesIndex, ruleNameMap map[string]nameIDType,
	sectionRulesOrSubPols map[string]itemOrderType, diags *diag.Diagnostics,
) {
	type ruleItem struct {
		name, id string
		ruleType cato_models.PolicyRuleTypeEnum
		index    int64
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
	if plan == nil || !utils.HasValue(plan.NetworkRules) {
		return
	}

	tfRuleData := make(map[string]LanNetworkRule)
	if utils.CheckErr(diags, plan.NetworkRules.ElementsAs(context.Background(), &tfRuleData, false)) {
		return
	}

	// make a helper map:    map[section][]{rule_or_subpol_name,id,type,index}
	ruleIndexes := make(map[string][]ruleItem)
	for ruleOrSPolName, rule := range tfRuleData {
		sectionName := rule.SectionName.ValueString()
		ruleInfo, ok := ruleNameMap[ruleOrSPolName]
		if !ok {
			diags.AddError("error parsing plan", fmt.Sprintf("cannot find details for rule '%s'", ruleOrSPolName))
			return
		}
		ruleIndexes[sectionName] = append(ruleIndexes[sectionName],
			ruleItem{
				name:     ruleOrSPolName,
				id:       ruleInfo.id,
				ruleType: ruleInfo.ruleType,
				index:    rule.IndexInSection.ValueInt64(),
			})
	}

	// sort the rules/sub-policies in each section by index
	// and update policySections[].target
	for sectionName, ruleSlice := range ruleIndexes {
		slices.SortFunc(ruleSlice, func(a, b ruleItem) int { return cmp.Compare(a.index, b.index) })
		if checkIndexes(sectionName, ruleSlice) {
			return
		}
		item := sectionRulesOrSubPols[sectionName]
		item.target = make([]nameIDType, len(ruleSlice))
		for i, rl := range ruleSlice {
			item.target[i] = nameIDType{name: rl.name, id: rl.id, ruleType: rl.ruleType}
		}
		sectionRulesOrSubPols[sectionName] = item
	}
}

// parsePlanFwRules takes TF state and creates a list of firewall rules in each network rule.
// Sets the ruleFwRules[netRuleName].target[] list, i.e. what is currently in state.
// On error, adds the error to *diags
//
// netRuleName -> {target: [{fwRuleName,id},{fwRuleName,id},...]}
func (r *lanRulesIndexResource) parsePlanFwRules(plan *LanFwRulesIndex, ruleNameMap map[string]string,
	ruleFwRules map[string]itemOrder, diags *diag.Diagnostics,
) {
	type ruleItem struct {
		name, id string
		index    int64
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
	if plan == nil || !utils.HasValue(plan.FirewallRules) {
		return
	}

	tfRuleData := make(map[string]LanFirewallRule)
	if utils.CheckErr(diags, plan.FirewallRules.ElementsAs(context.Background(), &tfRuleData, false)) {
		return
	}

	// make a helper map:    map[netRule][]{rule_name,id,index}
	ruleIndexes := make(map[string][]ruleItem)
	for fwRuleName, fwRule := range tfRuleData {
		netRuleName := fwRule.NetRuleName.ValueString()
		fwRuleID, ok := ruleNameMap[fwRuleName]
		if !ok {
			diags.AddError("error parsing plan", fmt.Sprintf("cannot find details for firewall rule '%s'", fwRuleName))
			return
		}
		ruleIndexes[netRuleName] = append(ruleIndexes[netRuleName],
			ruleItem{
				name:  fwRuleName,
				id:    fwRuleID,
				index: fwRule.IndexInRule.ValueInt64(),
			})
	}

	// sort the rules in each netRule by index
	// and update ruleFwRules[].target
	for netRuleName, fwRuleSlice := range ruleIndexes {
		slices.SortFunc(fwRuleSlice, func(a, b ruleItem) int { return cmp.Compare(a.index, b.index) })
		if checkIndexes(netRuleName, fwRuleSlice) {
			return
		}
		item := ruleFwRules[netRuleName]
		item.target = make([]nameID, len(fwRuleSlice))
		for i, rl := range fwRuleSlice {
			item.target[i] = nameID{name: rl.name, id: rl.id}
		}
		ruleFwRules[netRuleName] = item
	}
}

func (r *lanRulesIndexResource) moveSections(ctx context.Context, sections map[string]itemOrder) (diags diag.Diagnostics) {
	movePolicySections := func(sect itemOrder) error {
		// we need to loop backwards because the API does not support movint to the beginning
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

	for policyName, sectionLists := range sections {
		if err := movePolicySections(sectionLists); err != nil {
			if policyName == "" {
				policyName = "<main>"
			}
			diags.AddError(fmt.Sprintf("failed to move section in LAN policy '%s'", policyName), err.Error())
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

	curPossition := -1
	for i := range currentSections {
		if currentSections[i].id == sectionID {
			curPossition = i
			mySection = currentSections[i]
			break
		}
	}
	if curPossition == -1 {
		return fmt.Errorf("internal error: failed to find sectioID %s", sectionID)
	}
	if curPossition == newPosition {
		return nil // nothing to do
	}

	for i := curPossition; i < newPosition; i++ {
		currentSections[i] = currentSections[i+1]
	}
	currentSections[newPosition] = mySection

	// Prepare input for API to move the section
	input := cato_models.PolicyMoveSectionInput{
		ID: sectionID,
		To: &cato_models.PolicySectionPositionInput{},
	}
	if newPosition == len(currentSections)-1 { // move to LAST_IN_POLICY
		input.To.Position = cato_models.PolicySectionPositionEnumLastInPolicy
		if parentPolicyID != "" { // it is a sub-policy
			input.To.Ref = new(parentPolicyID)
		}
	} else {
		input.To.Position = cato_models.PolicySectionPositionEnumBeforeSection
		input.To.Ref = new(currentSections[newPosition+1].id)
	}

	// Call the API to move the section
	result, err := r.getClient().PolicySocketLanMoveSection(ctx, input, r.client.AccountId)
	if err != nil {
		return err
	}
	if errors := result.GetPolicy().GetSocketLan().GetMoveSection().GetErrors(); len(errors) > 0 {
		msg := "unknown error"
		if m := errors[0].GetErrorMessage(); m != nil {
			msg = *m
		}
		return fmt.Errorf("failed to move LAN policy section '%s': %v", sectionName, msg)
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

	for ruleOrSubPolName, ruleLists := range rulesOrSubPols {
		if err := moveSectionRules(ruleLists); err != nil {
			if ruleOrSubPolName == "" {
				ruleOrSubPolName = "<main>"
			}
			diags.AddError(fmt.Sprintf("failed to move section in LAN policy '%s'", ruleOrSubPolName), err.Error())
			return
		}
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

	for ruleName, ruleLists := range fwRules {
		if err := moveFwRules(ruleLists); err != nil {
			diags.AddError(fmt.Sprintf("failed to move firewall rule in LAN policy '%s'", ruleName), err.Error())
			return
		}
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

	curPossition := -1
	for i := range currentRulesOrSubPols {
		if currentRulesOrSubPols[i].id == ruleID {
			curPossition = i
			myRule = currentRulesOrSubPols[i]
			break
		}
	}
	if curPossition == -1 {
		return fmt.Errorf("internal error: failed to find rule ID %s", ruleID)
	}
	if curPossition == newPosition {
		return nil // nothing to do
	}

	for i := curPossition; i < newPosition; i++ {
		currentRulesOrSubPols[i] = currentRulesOrSubPols[i+1]
	}
	currentRulesOrSubPols[newPosition] = myRule

	// Prepare input for API to move the network rule
	input := cato_models.PolicyMoveRuleInput{
		ID: ruleID,
		To: &cato_models.PolicyRulePositionInput{},
	}
	if newPosition == len(currentRulesOrSubPols)-1 { // move to LAST_IN_SECTION
		input.To.Position = new(cato_models.PolicyRulePositionEnumLastInSection)
		input.To.Ref = new(parentSectionID)
	} else {
		input.To.Position = new(cato_models.PolicyRulePositionEnumBeforeRule)
		input.To.Ref = new(currentRulesOrSubPols[newPosition+1].id)
	}

	// Call the API to move the rule
	result, err := r.getClient().PolicySocketLanMoveRule(ctx, input, r.client.AccountId)
	if err != nil {
		return err
	}
	if errors := result.GetPolicy().GetSocketLan().GetMoveRule().GetErrors(); len(errors) > 0 {
		msg := "unknown error"
		if m := errors[0].GetErrorMessage(); m != nil {
			msg = *m
		}
		return fmt.Errorf("failed to move LAN policy rule '%s': %v", ruleName, msg)
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

	curPossition := -1
	for i := range currentRulesOrSubPols {
		if currentRulesOrSubPols[i].id == policyID {
			curPossition = i
			myPolicy = currentRulesOrSubPols[i]
			break
		}
	}
	if curPossition == -1 {
		return fmt.Errorf("internal error: failed to find sub-policy ID %s", policyID)
	}
	if curPossition == newPosition {
		return nil // nothing to do
	}

	for i := curPossition; i < newPosition; i++ {
		currentRulesOrSubPols[i] = currentRulesOrSubPols[i+1]
	}
	currentRulesOrSubPols[newPosition] = myPolicy

	// Prepare input for API to move the sub-policy
	input := cato_models.PolicyMoveRuleInput{
		ID: policyID,
		To: &cato_models.PolicyRulePositionInput{},
	}
	if newPosition == len(currentRulesOrSubPols)-1 { // move to LAST_IN_SECTION
		input.To.Position = new(cato_models.PolicyRulePositionEnumLastInSection)
		input.To.Ref = new(parentSectionID)
	} else {
		input.To.Position = new(cato_models.PolicyRulePositionEnumBeforeRule)
		input.To.Ref = new(currentRulesOrSubPols[newPosition+1].id)
	}

	// Call the API to move the sub-policy
	result, err := r.getClient().PolicySocketLanMoveRule(ctx, input, r.client.AccountId)
	if err != nil {
		return err
	}
	if errors := result.GetPolicy().GetSocketLan().GetMoveRule().GetErrors(); len(errors) > 0 {
		msg := "unknown error"
		if m := errors[0].GetErrorMessage(); m != nil {
			msg = *m
		}
		return fmt.Errorf("failed to move LAN sub-policy '%s': %v", policyName, msg)
	}
	return nil
}

// moveFwRuleToPosition moves the firewall rule with given ID to the given position in []currentRules
// (shifting the rest up) and calls the API to move the rule in the CMA
// Warning: it only moves the rules down, moving up is not supported!
func (r *lanRulesIndexResource) moveFwRuleToPosition(ctx context.Context, parentRuleID string,
	currentRules []nameID, fwRuleID, fwRuleName string, newPosition int,
) error {
	var myRule nameID

	tflog.Debug(ctx, "moving firewall rule '"+fwRuleName+"' to position "+strconv.Itoa(newPosition))

	curPossition := -1
	for i := range currentRules {
		if currentRules[i].id == fwRuleID {
			curPossition = i
			myRule = currentRules[i]
			break
		}
	}
	if curPossition == -1 {
		return fmt.Errorf("internal error: failed to find firewall rule ID %s", fwRuleID)
	}
	if curPossition == newPosition {
		return nil // nothing to do
	}

	for i := curPossition; i < newPosition; i++ {
		currentRules[i] = currentRules[i+1]
	}
	currentRules[newPosition] = myRule

	// Prepare input for API to move the sub-policy
	input := cato_models.PolicyMoveSubRuleInput{
		ID: fwRuleID,
		To: &cato_models.PolicySubRulePositionInput{},
	}
	if newPosition == len(currentRules)-1 { // move to LAST_IN_RULE
		input.To.Position = cato_models.PolicySubRulePositionEnumLastInRule
		input.To.Ref = parentRuleID
	} else {
		input.To.Position = cato_models.PolicySubRulePositionEnumBeforeSubRule
		input.To.Ref = currentRules[newPosition+1].id
	}

	// Call the API to move the sub-policy
	result, err := r.getClient().PolicySocketLanFirewallMoveRule(ctx, r.client.AccountId, nil, input)
	if err != nil {
		return err
	}
	if errors := result.GetPolicy().GetSocketLan().GetFirewall().GetMoveRule().GetErrors(); len(errors) > 0 {
		msg := "unknown error"
		if m := errors[0].GetErrorMessage(); m != nil {
			msg = *m
		}
		return fmt.Errorf("failed to move LAN firewall rule '%s': %v", fwRuleName, msg)
	}
	return nil
}

func (r *lanRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LanFwRulesIndex
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For this resource, we should preserve the state as-is since it represents
	// the intended configuration/ordering rather than reading all data from API.
	// The state is already properly set during Create/Update operations.
	// Only refresh IDs if needed, but preserve planned values.

	// No changes needed - preserve existing state
	if diags := resp.State.Set(ctx, &state); diags.HasError() {
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

// publish calls the API to publish the draft policy revision
func (r *lanRulesIndexResource) publish(ctx context.Context, diags *diag.Diagnostics) {
	const summary = "failed to publish LAN firewall policy"
	const notFound = "PolicyRevisionNotFound"
	result, err := r.getClient().PolicySocketLanPublishPolicyRevision(ctx, nil, nil, r.client.AccountId)
	if err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	errors := result.GetPolicy().GetSocketLan().GetPublishPolicyRevision().GetErrors()
	if len(errors) > 0 {
		for _, e := range errors {
			if code := e.GetErrorCode(); code != nil && *code == notFound {
				continue
			}
			if msg := e.GetErrorMessage(); msg != nil {
				diags.AddError(summary, *msg)
			}
		}
		return
	}
}
