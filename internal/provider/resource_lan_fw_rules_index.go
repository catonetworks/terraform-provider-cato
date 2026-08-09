package provider

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	PolicySocketLanPolicy(ctx context.Context, accountID string, socketLanPolicyInput *cato_models.SocketLanPolicyInput, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicySocketLanPolicy, error)
}

func (r *lanRulesIndexResource) getClient() LanFwRuleClient { return r.catov2Client }

func (r *lanRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_lf_move_rule"
}

func (r *lanRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves index values for LAN Firewall Rules.",
		Attributes: map[string]schema.Attribute{
			"section_data": schema.MapNestedAttribute{
				Description: "Map of section indexes keyed by section name",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Section id",
							Computed:    true,
						},
						"section_index": schema.Int64Attribute{
							Description: "Position of the section in the policy or sub-policy. Starts with 1.",
							Required:    true,
						},
						"sub_policy_name": schema.Int64Attribute{
							Description: "Sub-policy name. If not set, the main policy is used.",
							Optional:    true,
						},
					},
				},
			},
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of network rule or sub-policy index for each section, keyed by rule or sub-policy name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Rule id",
							Computed:    true,
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

	// if utils.CheckErr(&resp.Diagnostics, r.moveRules(ctx, indexMap)) {
	// 	return
	// }
	_ = indexMap
	_ = hydratedState
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
}
type itemOrder struct {
	current []nameID
	target  []nameID
}
type nameID struct {
	name, id string
}
type itemOrderType struct {
	current []nameIDType
	target  []nameIDType
}
type nameIDType struct {
	name, id string
	ruleType cato_models.PolicyRuleTypeEnum
}

func (r *lanRulesIndexResource) hydrateLanFwRulesIndex(ctx context.Context, plan *LanFwRulesIndex, diags *diag.Diagnostics,
) (newState *LanFwRulesIndex, indexMap *lfIndexMap) {
	policySections := make(map[string]itemOrder)
	sectionRulesOrSubPols := make(map[string]itemOrderType)

	// Call Cato API to get the policy
	result, err := r.getClient().PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("failed to hydrate sub-policy", err.Error())
		return nil, nil
	}
	policyBase := result.GetPolicy().GetSocketLan().GetPolicy()
	policyIDMap := r.makePolicyIDMap(policyBase) // map[policyID]policyName

	// Prepare section map
	//   map[policyName]{  // "" means the main policy
	//   	current:[{secName,secID},...]   // from API response
	//   	target: [{secName,secID},...]   // from TF plan
	//   }
	r.parseAPISections(policyBase, policyIDMap, policySections, diags) // current
	r.parsePlanSections(plan, policySections, diags)                   // target
	if diags.HasError() {
		return
	}

	// Prepare rule and subpolicy map
	//   map[sectionName]{
	//   	current:[{ruleName,id,RULE},{subPolicyName,id,SUB_POL},...]  // from API response
	//   	target: [{ruleName,id,RULE},{subPolicyName,id,SUB_POL},...]  // from TF plan
	//   }
	r.parseAPIRules(policyBase, sectionRulesOrSubPols, diags) // current
	r.parsePlanRules(plan, sectionRulesOrSubPols, diags)      // target
	if diags.HasError() {
		return
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
				return nil, nil
			}
			sections[section.name] = sectionObj
		}
	}
	sectionData := types.MapNull(types.ObjectType{AttrTypes: LanFwSectionDataTypes})
	if len(sections) > 0 {
		sd, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: LanFwSectionDataTypes}, sections)
		if objDiags.HasError() {
			diags.Append(objDiags...)
			return
		}
		sectionData = sd
	}

	// create TF RuleData:  map[rule/subPolicy name]{ruleID,sectName,sectIndex,ruleTypw} -> types.Map
	rulesOrSubPols := make(map[string]types.Object)
	for sectName, rList := range sectionRulesOrSubPols {
		for i, rule := range rList.current {
			tfRuleData := LanFwRuleData{
				ID:             types.StringValue(rule.id),
				RuleType:       types.StringValue(string(rule.ruleType)),
				SectionName:    types.StringValue(sectName),
				IndexInSection: types.Int64Value(int64(i + 1)), // 1-based
			}
			ruleObj, objDiags := types.ObjectValueFrom(ctx, LanFwRuleDataTypes, tfRuleData)
			diags.Append(objDiags...)
			if diags.HasError() {
				return nil, nil
			}
			rulesOrSubPols[rule.name] = ruleObj
		}
	}
	ruleData := types.MapNull(types.ObjectType{AttrTypes: LanFwRuleDataTypes})
	if len(rulesOrSubPols) > 0 {
		rd, objDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: LanFwRuleDataTypes}, rulesOrSubPols)
		if objDiags.HasError() {
			diags.Append(objDiags...)
			return
		}
		ruleData = rd
	}

	newState = &LanFwRulesIndex{
		SectionData: sectionData,
		RuleData:    ruleData,
	}
	indexMap = &lfIndexMap{
		sections:       policySections,
		rulesOrSubPols: sectionRulesOrSubPols,
	}
	return newState, indexMap
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
		policyName := ""                                   // default main policy
		if spID := section.GetSubPolicyID(); spID != nil { // get sub-policy name if defined
			polName, ok := policyIDMap[*spID]
			if !ok {
				diags.AddError("processing policy API response",
					fmt.Sprintf("subpolicy id '%s' not found in API response", *spID))
				return
			}
			policyName = polName
		}
		iOrder := policySections[policyName]
		iOrder.current = append(iOrder.current, nameID{name: section.GetName(), id: section.GetID()})
		policySections[policyName] = iOrder
	}
}

// parseAPIRules takes API result and creates a list of rules or sub-policies in each section.
// Sets the policyRules[sectionName].current[] list, i.e. what is currently in CMA.
// On error, adds the error to *diags
//
// sectionName -> {current: [{ruleName,id,RULE}, {subPolicyName,id,SUB_POLICY}, ...]}
func (r *lanRulesIndexResource) parseAPIRules(apiResult *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy,
	sectionRulesOrSubPols map[string]itemOrderType, diags *diag.Diagnostics,
) {
	for _, rul := range apiResult.Rules {
		rule := rul.GetRule()
		sectionName := rule.GetSection().GetName()
		iOrder := sectionRulesOrSubPols[sectionName]
		iOrder.current = append(iOrder.current, nameIDType{name: rule.GetName(), id: rule.GetID(), ruleType: r.ruleType(rul)})
		sectionRulesOrSubPols[sectionName] = iOrder
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
	policySections map[string]itemOrder, diags *diag.Diagnostics,
) {
	if plan == nil || !utils.HasValue(plan.SectionData) {
		return
	}

	tfSections := make(map[string]LanFwSectionData)
	if utils.CheckErr(diags, plan.SectionData.ElementsAs(context.Background(), &tfSections, false)) {
		return
	}

	type sectItem struct {
		name, id string
		index    int64
	}
	// make a helper map:    map[policyName][]{section_name,id,index}
	sectionIndexes := make(map[string][]sectItem)
	for sectionName, section := range tfSections {
		policyName := section.SubPolicyName.ValueString()
		sectionIndex := section.SectionIndex.ValueInt64()
		sectionID := section.ID.ValueString()
		sectionIndexes[policyName] = append(sectionIndexes[policyName],
			sectItem{name: sectionName, id: sectionID, index: sectionIndex})
	}

	// sort the sections in each policy by sectionIndex
	// and update policySections[].target
	for policyName, sectSlice := range sectionIndexes {
		slices.SortFunc(sectSlice, func(a, b sectItem) int { return cmp.Compare(a.index, b.index) })
		item := policySections[policyName]
		item.target = make([]nameID, len(sectSlice))
		for i, sect := range sectSlice {
			item.target[i] = nameID{name: sect.name, id: sect.id}
		}
		policySections[policyName] = item
	}
}

// parsePlanRules takes TF state and creates a list of rules/sub-policies in each section.
// Sets the sectionRulesOrSubPols[sectionName].target[] list, i.e. what is currently in state.
// On error, adds the error to *diags
//
// sectionName -> {target: [{ruleName,id,RULE},{subPolName,id,SUB_POL},...]}
func (r *lanRulesIndexResource) parsePlanRules(plan *LanFwRulesIndex,
	sectionRulesOrSubPols map[string]itemOrderType, diags *diag.Diagnostics,
) {
	if plan == nil || !utils.HasValue(plan.RuleData) {
		return
	}

	tfRuleData := make(map[string]LanFwRuleData)
	if utils.CheckErr(diags, plan.RuleData.ElementsAs(context.Background(), &tfRuleData, false)) {
		return
	}

	type ruleItem struct {
		name, id string
		ruleType cato_models.PolicyRuleTypeEnum
		index    int64
	}
	// make a helper map:    map[section][]{rule_or_subpol_name,id,type,index}
	ruleIndexes := make(map[string][]ruleItem)
	for ruleOrSPolName, rule := range tfRuleData {
		sectionName := rule.SectionName.ValueString()
		ruleIndexes[sectionName] = append(ruleIndexes[sectionName],
			ruleItem{
				name:     ruleOrSPolName,
				id:       rule.ID.ValueString(),
				ruleType: cato_models.PolicyRuleTypeEnum(rule.RuleType.ValueString()),
				index:    rule.IndexInSection.ValueInt64(),
			})
	}

	// sort the rules/sub-policies in each section by index
	// and update policySections[].target
	for sectionName, ruleSlice := range ruleIndexes {
		slices.SortFunc(ruleSlice, func(a, b ruleItem) int { return cmp.Compare(a.index, b.index) })
		item := sectionRulesOrSubPols[sectionName]
		item.target = make([]nameIDType, len(ruleSlice))
		for i, rl := range ruleSlice {
			item.target[i] = nameIDType{name: rl.name, id: rl.id, ruleType: rl.ruleType}
		}
		sectionRulesOrSubPols[sectionName] = item
	}
}

func (r *lanRulesIndexResource) moveRules(ctx context.Context, indexMap lfIndexMap) diag.Diagnostics {
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
