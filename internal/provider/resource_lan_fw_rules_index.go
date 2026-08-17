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

//nolint:funlen // Declarative nested schema remains clearer in one function.
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
				Description: "Map of section indexes keyed by a caller-chosen stable key. " +
					"For backward compatibility, the key is used as section_name when section_name is omitted.",
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:   "Section id",
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
						},
						"section_name": schema.StringAttribute{
							Description: "LAN section name. Defaults to the map key for backward compatibility.",
							Optional:    true,
							Computed:    true,
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
				Description: "Map of network rule or sub-policy indexes keyed by a caller-chosen stable key. " +
					"For backward compatibility, the key is used as rule_name when rule_name is omitted.",
				Required: false,
				Optional: true,
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
						"rule_name": schema.StringAttribute{
							Description: "Network rule or sub-policy name. Defaults to the map key for backward compatibility.",
							Optional:    true,
							Computed:    true,
						},
						"section_name": schema.StringAttribute{
							Description: "LAN section name housing rule",
							Required:    true,
						},
						"section_key": schema.StringAttribute{
							Description: "Map key of the parent section. Required when section_name is ambiguous.",
							Optional:    true,
						},
						"index_in_section": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    true,
						},
					},
				},
			},
			"firewall_rules": schema.MapNestedAttribute{
				Description: "Map of firewall rule indexes keyed by a caller-chosen stable key. " +
					"For backward compatibility, the key is used as firewall_rule_name when firewall_rule_name is omitted.",
				Required: false,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:   "Firewall rule ID",
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
						},
						"firewall_rule_name": schema.StringAttribute{
							Description: "Firewall rule name. Defaults to the map key for backward compatibility.",
							Optional:    true,
							Computed:    true,
						},
						"net_rule_name": schema.StringAttribute{
							Description: "Parent LAN network rule name",
							Required:    true,
						},
						"net_rule_key": schema.StringAttribute{
							Description: "Map key of the parent network rule. Required when net_rule_name is ambiguous.",
							Optional:    true,
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
	// map[policyID]{
	//   current:[{sectionName:secA,id:100},{sectionName:secB,id:101},...],
	//   target: [{sectionName:secB,id:101},{sectionName:secA,id:100},...]}
	sections map[string]itemOrder

	// Rule or subPolicy order in given section
	// map[sectionID]{
	//   current:[{name:ruleA,id:100,typ:rule},{name:subPolA,id:123,typ:subPolicy},...],
	//   target: [{name:subPolA,id:123,typ:subPolicy},{name:ruleA,id:100,typ:rule},...],
	rulesOrSubPols map[string]itemOrderType

	// Firewall rule order in given network rule
	// map[networkRuleID]{
	//   current:[{name:FwRuleA,id:100},{name:FwRuleB,id:123},...],
	//   target: [{name:FwRuleB,id:123},{name:FwRuleA,id:100},...],
	firewallRules map[string]itemOrder
}
type itemOrder struct {
	parentID   string
	parentName string
	current    []nameID
	target     []nameID
}
type nameID struct {
	name, id string
}
type itemOrderType struct {
	parentID   string
	parentName string
	current    []nameIDType
	target     []nameIDType
}
type nameIDType struct {
	name, id string
	ruleType cato_models.PolicyRuleTypeEnum
}

type lanFwStateAliases struct {
	sectionKeyByID        map[string]string
	netRuleKeyByID        map[string]string
	firewallRuleKeyByID   map[string]string
	netRuleParentKeyByID  map[string]string
	firewallParentKeyByID map[string]string
	manageNetworkRules    bool
	manageFirewallRules   bool
}

func (r *lanRulesIndexResource) hydrateLanFwRulesIndex(ctx context.Context, plan *LanFwRulesIndex, diags *diag.Diagnostics,
) (newState *LanFwRulesIndex, indexMap *lfIndexMap) {
	return r.hydrateLanFwRulesIndexWithAliases(ctx, plan, nil, diags)
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

func (r *lanRulesIndexResource) hydrateLanFwRulesIndexWithAliases(ctx context.Context, plan *LanFwRulesIndex,
	aliases *lanFwStateAliases, diags *diag.Diagnostics,
) (newState *LanFwRulesIndex, indexMap *lfIndexMap) {
	// Call Cato API to get the policy
	result, err := r.getClient().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
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

	for _, ruleLists := range rulesOrSubPols {
		if err := moveSectionRules(ruleLists); err != nil {
			diags.AddError(fmt.Sprintf("failed to move rule in LAN section '%s'", ruleLists.parentName), err.Error())
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

	for _, ruleLists := range fwRules {
		if err := moveFwRules(ruleLists); err != nil {
			diags.AddError(fmt.Sprintf("failed to move firewall rule under network rule '%s'",
				ruleLists.parentName), err.Error())
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
			} else {
				diags.AddError(summary, "unknown error")
			}
		}
		return
	}
}
