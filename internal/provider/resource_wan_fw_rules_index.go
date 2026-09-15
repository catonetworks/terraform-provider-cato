package provider

import (
	"context"
	"errors"
	"fmt"
	"sort"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource              = &wanRulesIndexResource{}
	_ resource.ResourceWithConfigure = &wanRulesIndexResource{}
	// _ resource.ResourceWithImportState = &wanRulesIndexResource{}
)

func NewWanRulesIndexResource() resource.Resource {
	return &wanRulesIndexResource{}
}

type wanRulesIndexResource struct {
	client  *catoClientData
	wanBulk WanFirewallBulkPolicyClient // optional override for tests
}

func (r *wanRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_wf_move_rule"
}

func (r *wanRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves index values for WAN Firewall Rules.",
		Attributes: map[string]schema.Attribute{
			"section_to_start_after_id": schema.StringAttribute{
				Description: "WAN rule id",
				Required:    false,
				Optional:    true,
				// Computed:    true,
			},
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of WAN Rule Policy Indexes keyed by rule_name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "IFW rule id",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"index_in_section": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"section_name": schema.StringAttribute{
							Description: "WAN section name housing rule",
							Required:    false,
							Optional:    true,
						},
						"sub_policy_name": schema.StringAttribute{
							Description: "WAN sub-policy name housing rule",
							Required:    false,
							Optional:    true,
						},
						"rule_name": schema.StringAttribute{
							Description: "WAN rule name housing rule",
							Required:    false,
							Optional:    true,
						},
						"description": schema.StringAttribute{
							Description: "WAN rule description",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "IFW rule enabled",
							Required:    false,
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(), // Avoid drift
							},
						},
					},
				},
			},
			"section_data": schema.MapNestedAttribute{
				Description: "Map of IFW section Indexes keyed by section_name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "IFW section id housing rule",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"section_index": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    true,
							Optional:    false,
						},
						"section_name": schema.StringAttribute{
							Description: "IFW section name housing rule",
							Required:    true,
							Optional:    false,
						},
						"sub_policy_name": schema.StringAttribute{
							Description: "WAN sub-policy name owning section",
							Required:    false,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

var WanRuleIndexResourceObjectTypes = types.ObjectType{AttrTypes: WanRuleIndexResourceAttrTypes}
var WanRuleIndexResourceAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"index_in_section": types.Int64Type,
	"section_name":     types.StringType,
	"sub_policy_name":  types.StringType,
	"rule_name":        types.StringType,
	"description":      types.StringType,
	"enabled":          types.BoolType,
}

var WanSectionIndexResourceObjectTypes = types.ObjectType{AttrTypes: WanSectionIndexResourceAttrTypes}
var WanSectionIndexResourceAttrTypes = map[string]attr.Type{
	"id":              types.StringType,
	"section_name":    types.StringType,
	"section_index":   types.Int64Type,
	"sub_policy_name": types.StringType,
}

func (r *wanRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *wanRulesIndexResource) wanBulkPolicy() WanFirewallBulkPolicyClient {
	if r.wanBulk != nil {
		return r.wanBulk
	}
	if r.client == nil {
		return nil
	}
	return r.client.catov2
}

// func (r *wanRulesIndexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	// Retrieve import ID and save to id attribute
// 	// resource.ImportStatePassthroughID(ctx, path.Root("Id"), req, resp)
// }

func (r *wanRulesIndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WanRulesIndex
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionObjectsList, rulesObjectsList, diags, err := r.moveWanRulesAndSections(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(diags...)
	plan.SectionData = sectionObjectsList
	plan.RuleData = rulesObjectsList

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WanRulesIndex
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

func (r *wanRulesIndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WanRulesIndex
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionObjectsList, rulesObjectsList, diags, err := r.moveWanRulesAndSections(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(diags...)
	plan.SectionData = sectionObjectsList
	plan.RuleData = rulesObjectsList

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanRulesIndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WanRulesIndex
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *wanRulesIndexResource) moveWanRulesAndSections(
	ctx context.Context,
	plan WanRulesIndex,
) (sectionObjects, ruleObjects basetypes.MapValue, diags diag.Diagnostics, err error) {
	if !wanPlanHasSubPolicyData(ctx, plan) {
		return r.moveWanRulesAndSectionsMain(ctx, plan)
	}

	mainPlan, subPolicyPlan, diags, err := splitWanPlanBySubPolicy(ctx, plan)
	if err != nil {
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	if len(mainPlan.SectionData.Elements()) > 0 || len(mainPlan.RuleData.Elements()) > 0 {
		var mainDiags diag.Diagnostics
		var mainErr error
		sectionObjects, ruleObjects, mainDiags, mainErr = r.moveWanRulesAndSectionsMain(ctx, mainPlan)
		diags = append(diags, mainDiags...)
		if mainErr != nil {
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, mainErr
		}
	}

	subSectionObjects, subRuleObjects, subDiags, subErr := r.moveWanSubPolicyRulesAndSections(ctx, subPolicyPlan)
	diags = append(diags, subDiags...)
	if subErr != nil {
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, subErr
	}

	sectionObjects, mergeDiags := mergeWanMapValues(
		WanSectionIndexResourceObjectTypes,
		sectionObjects,
		subSectionObjects,
	)
	diags = append(diags, mergeDiags...)
	ruleObjects, mergeDiags = mergeWanMapValues(
		WanRuleIndexResourceObjectTypes,
		ruleObjects,
		subRuleObjects,
	)
	diags = append(diags, mergeDiags...)

	return sectionObjects, ruleObjects, diags, nil
}

func wanPlanHasSubPolicyData(ctx context.Context, plan WanRulesIndex) bool {
	for _, value := range plan.SectionData.Elements() {
		var item WanRulesSectionItemIndex
		if diags := value.(types.Object).As(ctx, &item, basetypes.ObjectAsOptions{}); !diags.HasError() &&
			item.SubPolicyName.ValueString() != "" {
			return true
		}
	}
	for _, value := range plan.RuleData.Elements() {
		var item WanRulesRuleItemIndex
		if diags := value.(types.Object).As(ctx, &item, basetypes.ObjectAsOptions{}); !diags.HasError() &&
			item.SubPolicyName.ValueString() != "" {
			return true
		}
	}
	return false
}

func splitWanPlanBySubPolicy(
	ctx context.Context,
	plan WanRulesIndex,
) (mainPlan, subPolicyPlan WanRulesIndex, diags diag.Diagnostics, err error) {
	diags = diag.Diagnostics{}
	mainPlan = WanRulesIndex{
		SectionToStartAfterID: plan.SectionToStartAfterID,
	}
	subPolicyPlan = WanRulesIndex{}

	mainSections, subSections, sectionDiags, err := splitWanMapBySubPolicy(
		plan.SectionData,
		WanSectionIndexResourceObjectTypes,
		func(value types.Object) (string, diag.Diagnostics) {
			var item WanRulesSectionItemIndex
			diags := value.As(ctx, &item, basetypes.ObjectAsOptions{})
			return item.SubPolicyName.ValueString(), diags
		},
	)
	diags = append(diags, sectionDiags...)
	if err != nil {
		return mainPlan, subPolicyPlan, diags, err
	}
	mainRules, subRules, ruleDiags, err := splitWanMapBySubPolicy(
		plan.RuleData,
		WanRuleIndexResourceObjectTypes,
		func(value types.Object) (string, diag.Diagnostics) {
			var item WanRulesRuleItemIndex
			diags := value.As(ctx, &item, basetypes.ObjectAsOptions{})
			return item.SubPolicyName.ValueString(), diags
		},
	)
	diags = append(diags, ruleDiags...)
	if err != nil {
		return mainPlan, subPolicyPlan, diags, err
	}

	mainPlan.SectionData = mainSections
	mainPlan.RuleData = mainRules
	subPolicyPlan.SectionData = subSections
	subPolicyPlan.RuleData = subRules
	return mainPlan, subPolicyPlan, diags, nil
}

func splitWanMapBySubPolicy(
	input basetypes.MapValue,
	objectType types.ObjectType,
	subPolicyName func(types.Object) (string, diag.Diagnostics),
) (mainValue, subPolicyValue basetypes.MapValue, diags diag.Diagnostics, err error) {
	diags = diag.Diagnostics{}
	mainElements := make(map[string]attr.Value)
	subPolicyElements := make(map[string]attr.Value)

	for key, value := range input.Elements() {
		object, ok := value.(types.Object)
		if !ok {
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, fmt.Errorf("map element %q is not an object", key)
		}
		name, elementDiags := subPolicyName(object)
		diags = append(diags, elementDiags...)
		if diags.HasError() {
			continue
		}
		if name == "" {
			mainElements[key] = value
		} else {
			subPolicyElements[key] = value
		}
	}

	mainValue, mainDiags := types.MapValue(objectType, mainElements)
	diags = append(diags, mainDiags...)
	subPolicyValue, subPolicyDiags := types.MapValue(objectType, subPolicyElements)
	diags = append(diags, subPolicyDiags...)
	return mainValue, subPolicyValue, diags, nil
}

func mergeWanMapValues(objectType types.ObjectType, values ...basetypes.MapValue) (basetypes.MapValue, diag.Diagnostics) {
	elements := make(map[string]attr.Value)
	diags := diag.Diagnostics{}
	for _, value := range values {
		for key, element := range value.Elements() {
			if _, exists := elements[key]; exists {
				diags.AddError("WAN firewall ordering state", fmt.Sprintf("duplicate map key %q", key))
				continue
			}
			elements[key] = element
		}
	}
	merged, mapDiags := types.MapValue(objectType, elements)
	diags = append(diags, mapDiags...)
	return merged, diags
}

type wanSubPolicyOrderingData struct {
	ID       string
	Name     string
	Sections []BulkPolicySectionRef
	Rules    []BulkPolicyRuleRow
}

//nolint:gocyclo
func wanSubPolicyOrderingDataFromPolicy(policy *cato_go_sdk.Policy) map[string]wanSubPolicyOrderingData {
	result := make(map[string]wanSubPolicyOrderingData)
	if policy == nil || policy.GetPolicy().GetWanFirewall() == nil {
		return result
	}

	wanPolicy := policy.GetPolicy().GetWanFirewall().GetPolicy()
	for _, subPolicyPayload := range wanPolicy.GetSubPolicies() {
		if subPolicyPayload == nil || subPolicyPayload.GetPolicy() == nil {
			continue
		}
		subPolicy := subPolicyPayload.GetPolicy()
		result[subPolicy.GetName()] = wanSubPolicyOrderingData{
			ID:   subPolicy.GetID(),
			Name: subPolicy.GetName(),
		}
	}

	sectionOrder := make(map[string][]string)
	sectionSeen := make(map[string]map[string]struct{})
	for _, rulePayload := range wanPolicy.GetRules() {
		if rulePayload == nil || rulePayload.GetSubPolicy() == nil {
			continue
		}
		subPolicyName := rulePayload.GetSubPolicy().GetName()
		data, exists := result[subPolicyName]
		if !exists {
			continue
		}
		rule := rulePayload.GetRule()
		section := rule.GetSection()
		if section == nil || section.GetID() == "" {
			// Sub-policy child and cleanup rules do not belong to a reorder section.
			continue
		}
		if sectionSeen[subPolicyName] == nil {
			sectionSeen[subPolicyName] = make(map[string]struct{})
		}
		if _, seen := sectionSeen[subPolicyName][section.GetID()]; !seen {
			sectionSeen[subPolicyName][section.GetID()] = struct{}{}
			sectionOrder[subPolicyName] = append(sectionOrder[subPolicyName], section.GetID())
			data.Sections = append(data.Sections, BulkPolicySectionRef{
				ID:   section.GetID(),
				Name: section.GetName(),
			})
		}
		data.Rules = append(data.Rules, BulkPolicyRuleRow{
			SectionID:   section.GetID(),
			SectionName: section.GetName(),
			RuleID:      rule.GetID(),
			RuleName:    rule.GetName(),
			Index:       rule.GetIndex(),
			IsSystem: policyElementHasProperty(
				rulePayload.GetProperties(),
				cato_models.PolicyElementPropertiesEnumSystem,
			),
		})
		result[subPolicyName] = data
	}

	for name, data := range result {
		orderedSections := make([]BulkPolicySectionRef, 0, len(sectionOrder[name]))
		sectionsByID := make(map[string]BulkPolicySectionRef, len(data.Sections))
		for _, section := range data.Sections {
			sectionsByID[section.ID] = section
		}
		for _, sectionID := range sectionOrder[name] {
			orderedSections = append(orderedSections, sectionsByID[sectionID])
		}
		data.Sections = orderedSections
		result[name] = data
	}
	return result
}

//nolint:gocyclo,funlen
func (r *wanRulesIndexResource) moveWanSubPolicyRulesAndSections(
	ctx context.Context,
	plan WanRulesIndex,
) (sectionObjects, ruleObjects basetypes.MapValue, diags diag.Diagnostics, err error) {
	diags = diag.Diagnostics{}
	policy, err := r.wanBulkPolicy().PolicyWanFirewall(
		ctx,
		&cato_models.WanFirewallPolicyInput{},
		r.client.AccountId,
	)
	if err != nil {
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}
	topologies := wanSubPolicyOrderingDataFromPolicy(policy)

	sectionPlans := make(map[string][]WanRulesSectionDataIndex)
	for _, value := range plan.SectionData.Elements() {
		var item WanRulesSectionItemIndex
		diags = append(diags, value.(types.Object).As(ctx, &item, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			continue
		}
		subPolicyName := item.SubPolicyName.ValueString()
		sectionPlans[subPolicyName] = append(sectionPlans[subPolicyName], WanRulesSectionDataIndex{
			ID:            item.ID.ValueString(),
			SectionIndex:  item.SectionIndex.ValueInt64(),
			SectionName:   item.SectionName.ValueString(),
			SubPolicyName: subPolicyName,
		})
	}

	rulePlans := make(map[string][]WanRulesRuleDataIndex)
	for _, value := range plan.RuleData.Elements() {
		var item WanRulesRuleItemIndex
		diags = append(diags, value.(types.Object).As(ctx, &item, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			continue
		}
		subPolicyName := item.SubPolicyName.ValueString()
		rulePlans[subPolicyName] = append(rulePlans[subPolicyName], WanRulesRuleDataIndex{
			ID:             item.ID.ValueString(),
			IndexInSection: item.IndexInSection.ValueInt64(),
			SectionName:    item.SectionName.ValueString(),
			RuleName:       item.RuleName.ValueString(),
			Description:    item.Description.ValueString(),
			Enabled:        item.Enabled.ValueBool(),
			SubPolicyName:  subPolicyName,
		})
	}
	if diags.HasError() {
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, errors.New("failed to parse WAN sub-policy ordering plan")
	}

	subPolicyNames := make([]string, 0, len(sectionPlans)+len(rulePlans))
	seenSubPolicies := make(map[string]struct{})
	for name := range sectionPlans {
		if name == "" {
			continue
		}
		seenSubPolicies[name] = struct{}{}
		subPolicyNames = append(subPolicyNames, name)
	}
	for name := range rulePlans {
		if name == "" {
			continue
		}
		if _, seen := seenSubPolicies[name]; !seen {
			seenSubPolicies[name] = struct{}{}
			subPolicyNames = append(subPolicyNames, name)
		}
	}
	sort.Strings(subPolicyNames)

	sectionObjectMap := make(map[string]attr.Value)
	ruleObjectMap := make(map[string]attr.Value)
	for _, subPolicyName := range subPolicyNames {
		topology, exists := topologies[subPolicyName]
		if !exists {
			return basetypes.MapValue{}, basetypes.MapValue{}, diags,
				fmt.Errorf("WAN sub-policy %q was not found in the API response", subPolicyName)
		}
		orderedSections, err := orderBulkPolicySections(topology.Sections, sectionPlans[subPolicyName])
		if err != nil {
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, fmt.Errorf(
				"WAN sub-policy %q section ordering: %w", subPolicyName, err,
			)
		}

		planned := make([]BulkPlannedRuleIndex, 0, len(rulePlans[subPolicyName]))
		for _, rule := range rulePlans[subPolicyName] {
			planned = append(planned, BulkPlannedRuleIndex{
				SectionName:    rule.SectionName,
				RuleName:       rule.RuleName,
				IndexInSection: rule.IndexInSection,
			})
		}
		reorderIn, buildErr := buildPolicyReorderInput(orderedSections, topology.Rules, planned)
		if buildErr != nil {
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, fmt.Errorf(
				"WAN sub-policy %q rule ordering: %w", subPolicyName, buildErr,
			)
		}
		subPolicyID := topology.ID
		reorderIn.SubPolicyID = &subPolicyID

		var reorderOut *cato_go_sdk.PolicyWanFirewallReorderPolicy
		reorderErr := withAcctestPolicyRevisionCleanupRetryOnce(ctx, "PolicyWanFirewallReorderPolicy", func() error {
			return discardFirewallAndWANPolicyRevisions(ctx, r.client.catov2, r.client.AccountId)
		}, func() error {
			var callErr error
			reorderOut, callErr = r.wanBulkPolicy().PolicyWanFirewallReorderPolicy(
				ctx,
				&cato_models.WanFirewallPolicyMutationInput{},
				reorderIn,
				r.client.AccountId,
			)
			return wanFirewallReorderError(reorderOut, callErr)
		})
		if reorderErr != nil {
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, fmt.Errorf(
				"WAN sub-policy %q reorder: %w", subPolicyName, reorderErr,
			)
		}

		sectionIDs := make(map[string]string, len(topology.Sections))
		for _, section := range topology.Sections {
			sectionIDs[section.Name] = section.ID
		}
		for _, section := range sectionPlans[subPolicyName] {
			sectionObject, sectionDiags := types.ObjectValue(
				WanSectionIndexResourceAttrTypes,
				map[string]attr.Value{
					"id":              types.StringValue(sectionIDs[section.SectionName]),
					"section_name":    types.StringValue(section.SectionName),
					"section_index":   types.Int64Value(section.SectionIndex),
					"sub_policy_name": types.StringValue(subPolicyName),
				},
			)
			diags = append(diags, sectionDiags...)
			sectionObjectMap[section.SectionName] = sectionObject
		}

		ruleIDs := make(map[string]BulkPolicyRuleRow, len(topology.Rules))
		for _, rule := range topology.Rules {
			ruleIDs[rule.RuleName] = rule
		}
		for _, rule := range rulePlans[subPolicyName] {
			apiRule, exists := ruleIDs[rule.RuleName]
			if !exists {
				return basetypes.MapValue{}, basetypes.MapValue{}, diags,
					fmt.Errorf("WAN sub-policy %q rule %q was not found in the API response", subPolicyName, rule.RuleName)
			}
			ruleObject, ruleDiags := buildWanRuleIndexStateData(
				rule,
				map[string]string{rule.RuleName: apiRule.RuleID},
				map[string]string{rule.RuleName: rule.Description},
				map[string]bool{rule.RuleName: rule.Enabled},
				subPolicyName,
			)
			diags = append(diags, ruleDiags...)
			ruleObjectMap[rule.RuleName] = ruleObject
		}
	}

	pubErr := withPolicyRevisionConflictRetry(ctx, "PolicyWanFirewallPublishPolicyRevision", func() error {
		_, errPub := r.wanBulkPolicy().PolicyWanFirewallPublishPolicyRevision(
			ctx,
			&cato_models.PolicyPublishRevisionInput{},
			r.client.AccountId,
		)
		return errPub
	})
	if pubErr != nil {
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, pubErr
	}

	sectionObjects, sectionDiags := types.MapValue(WanSectionIndexResourceObjectTypes, sectionObjectMap)
	diags = append(diags, sectionDiags...)
	ruleObjects, ruleDiags := types.MapValue(WanRuleIndexResourceObjectTypes, ruleObjectMap)
	diags = append(diags, ruleDiags...)
	return sectionObjects, ruleObjects, diags, nil
}

//nolint:gocyclo,funlen
func (r *wanRulesIndexResource) moveWanRulesAndSectionsMain(
	ctx context.Context,
	plan WanRulesIndex,
) (sectionObjects basetypes.MapValue, ruleObjects basetypes.MapValue, diagnostics diag.Diagnostics, err error) {
	diags := []diag.Diagnostic{}
	ruleObjectMap := make(map[string]attr.Value)

	if plan.SectionToStartAfterID.ValueString() != "" {
		result, err := r.wanBulkPolicy().PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
		tflog.Debug(ctx, "Read.PolicyWanFirewallSectionsIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic("Catov2 API PolicyWanFirewallSectionsIndex error", err.Error()))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}
		isPresent := false
		for _, item := range result.Policy.WanFirewall.Policy.Sections {
			sectionID := cast.ToString(item.Section.ID)
			if sectionID == plan.SectionToStartAfterID.ValueString() {
				isPresent = true
				break
			}
		}
		if !isPresent {
			diags = append(diags, diag.NewErrorDiagnostic(
				"SectionToStartAfterID '"+plan.SectionToStartAfterID.ValueString()+"' not found",
				"Please check the section ID and try again.",
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, errors.New("sectionToStartAfterId not found")
		}
	}

	// maps section_name -> section_id
	sectionIDList := make(map[string]string)
	sectionIndexAPIData, err := r.wanBulkPolicy().PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Warn(ctx, "Read.PolicyWanFirewallSectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexAPIData),
	})
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API EntityLookup error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	// for easier processing, a map of section name to ID is created
	for _, v := range sectionIndexAPIData.Policy.WanFirewall.Policy.Sections {
		sectionIDList[v.Section.Name] = v.Section.ID
	}

	sectionListFromPlan := make([]WanRulesSectionDataIndex, 0)

	// Convert map to slice for processing
	sectionDataMapElements := plan.SectionData.Elements()
	for _, sectionValue := range sectionDataMapElements {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex WanRulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionDataTmp := WanRulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		}
		sectionListFromPlan = append(sectionListFromPlan, sectionDataTmp)
	}

	// Sort sections by SectionIndex to ensure proper ordering
	sort.Slice(sectionListFromPlan, func(i, j int) bool {
		return sectionListFromPlan[i].SectionIndex < sectionListFromPlan[j].SectionIndex
	})

	tflog.Debug(ctx, "Processing sections sorted by section_index", map[string]interface{}{
		"sections": utils.InterfaceToJSONString(sectionListFromPlan),
	})

	currentSectionID := ""

	sectionObjectMap := make(map[string]attr.Value)

	// create the sections from the list provided following the section ID provided in firstSectionId
	for _, workingSectionName := range sectionListFromPlan {
		policyMoveSectionInputInt := cato_models.PolicyMoveSectionInput{
			ID: sectionIDList[workingSectionName.SectionName],
		}

		// For the first element, check for sectionToStartAfterId, if not, start at last LAST_IN_POLICY
		// initializing currentSectionID to the SectionToStartAfterID otherwise set to id of first section for next in list
		if currentSectionID == "" {
			if plan.SectionToStartAfterID.ValueString() != "" {
				policyMoveSectionInputInt.To = &cato_models.PolicySectionPositionInput{
					Ref:      plan.SectionToStartAfterID.ValueStringPointer(),
					Position: "AFTER_SECTION",
				}
			} else {
				policyMoveSectionInputInt.To = &cato_models.PolicySectionPositionInput{
					Position: "LAST_IN_POLICY",
				}
			}
		} else {
			policyMoveSectionInputInt.To = &cato_models.PolicySectionPositionInput{
				Ref:      &currentSectionID,
				Position: "AFTER_SECTION",
			}
		}
		tflog.Warn(ctx, "Write.policyMoveSectionInputInt.response", map[string]interface{}{
			"sectionToStartAfterId":          plan.SectionToStartAfterID.ValueString(),
			"moveFrom":                       workingSectionName.SectionName,
			"toAfter":                        currentSectionID,
			"sectionIDList":                  sectionIDList,
			"workingSectionName.SectionName": workingSectionName.SectionName,
			"sectionIDList[workingSectionName.SectionName]": sectionIDList[workingSectionName.SectionName],
			"response": utils.InterfaceToJSONString(policyMoveSectionInputInt),
		})
		var sectionMoveAPIData *cato_go_sdk.PolicyWanFirewallMoveSection
		moveErr := withPolicyRevisionConflictRetry(ctx, "PolicyWanFirewallMoveSection", func() error {
			var callErr error
			sectionMoveAPIData, callErr = r.wanBulkPolicy().PolicyWanFirewallMoveSection(
				ctx, policyMoveSectionInputInt, r.client.AccountId)
			return wanFirewallMoveSectionError(sectionMoveAPIData, callErr)
		})
		if moveErr != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyWanFirewallMoveSection error",
				moveErr.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, moveErr
		}
		tflog.Warn(ctx, "Write.PolicyWanFirewallMoveSection.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(sectionMoveAPIData),
		})

		sectionIndexStateData, diagsSection := types.ObjectValue(
			WanSectionIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":              types.StringValue(sectionIDList[workingSectionName.SectionName]),
				"section_name":    types.StringValue(workingSectionName.SectionName),
				"section_index":   types.Int64Value(workingSectionName.SectionIndex),
				"sub_policy_name": types.StringNull(),
			},
		)
		diags = append(diags, diagsSection...)

		sectionObjectMap[workingSectionName.SectionName] = sectionIndexStateData

		currentSectionID = sectionIDList[workingSectionName.SectionName]
	}

	// now that the sections are ordered properly, move the rules to the correct locations
	if len(plan.RuleData.Elements()) > 0 {
		// get all of the list elements from the plan
		ruleListFromPlan := make([]WanRulesRuleDataIndex, 0)

		// Convert map to slice for processing
		ruleDataMapElements := plan.RuleData.Elements()
		for _, ruleValue := range ruleDataMapElements {
			ruleObject := ruleValue.(types.Object)
			var planSourceRuleIndex WanRulesRuleItemIndex
			diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

			rulenDataTmp := WanRulesRuleDataIndex{
				IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
				RuleName:       planSourceRuleIndex.RuleName.ValueString(),
				SectionName:    planSourceRuleIndex.SectionName.ValueString(),
				Description:    planSourceRuleIndex.Description.ValueString(),
				Enabled:        planSourceRuleIndex.Enabled.ValueBool(),
				SubPolicyName:  planSourceRuleIndex.SubPolicyName.ValueString(),
			}
			ruleListFromPlan = append(ruleListFromPlan, rulenDataTmp)
			tflog.Warn(ctx, "Read.rulenDataTmp.response", map[string]interface{}{
				"rulenDataTmp": utils.InterfaceToJSONString(rulenDataTmp),
			})
		}

		// Sort rules by IndexInSection to ensure proper ordering within sections
		sort.Slice(ruleListFromPlan, func(i, j int) bool {
			// First sort by section name, then by index within section
			if ruleListFromPlan[i].SectionName != ruleListFromPlan[j].SectionName {
				return ruleListFromPlan[i].SectionName < ruleListFromPlan[j].SectionName
			}
			return ruleListFromPlan[i].IndexInSection < ruleListFromPlan[j].IndexInSection
		})

		tflog.Warn(ctx, "Read.ruleListFromPlan.response (sorted by section_name and index_in_section)", map[string]interface{}{
			"ruleListFromPlan": utils.InterfaceToJSONString(ruleListFromPlan),
		})

		ruleNameIDData, err := r.wanBulkPolicy().PolicyWanFirewallRulesIndex(ctx, r.client.AccountId)
		tflog.Warn(ctx, "Read.PolicyWanFirewallRulesIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(ruleNameIDData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API EntityLookup error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}
		ruleNameIDMap := make(map[string]string)
		ruleNameDescriptionMap := make(map[string]string)
		ruleNameEnabledMap := make(map[string]bool)

		// create map of IFW rule names from the API to their IDs for easy lookup
		for _, ruleNameIDDataItem := range ruleNameIDData.Policy.WanFirewall.Policy.Rules {
			ruleNameIDMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.ID
			ruleNameDescriptionMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.Description
			ruleNameEnabledMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.Enabled
		}

		tflog.Warn(ctx, "Read.ruleNameIDMap.response", map[string]interface{}{
			"ruleNameIDMap": utils.InterfaceToJSONString(ruleNameIDMap),
		})

		sectionIdxAfter, err := r.wanBulkPolicy().PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyWanFirewallSectionsIndex error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}

		sections := make([]BulkPolicySectionRef, 0, len(sectionIdxAfter.Policy.WanFirewall.Policy.Sections))
		for _, item := range sectionIdxAfter.Policy.WanFirewall.Policy.Sections {
			sections = append(sections, BulkPolicySectionRef{
				ID:   item.Section.ID,
				Name: item.Section.Name,
			})
		}

		rules := make([]BulkPolicyRuleRow, 0, len(ruleNameIDData.Policy.WanFirewall.Policy.Rules))
		for _, item := range ruleNameIDData.Policy.WanFirewall.Policy.Rules {
			rules = append(rules, BulkPolicyRuleRow{
				SectionID:   item.Rule.Section.ID,
				SectionName: item.Rule.Section.Name,
				RuleID:      item.Rule.ID,
				RuleName:    item.Rule.Name,
				Index:       item.Rule.Index,
				IsSystem: policyElementHasProperty(
					item.Properties,
					cato_models.PolicyElementPropertiesEnumSystem,
				),
			})
		}

		planned := make([]BulkPlannedRuleIndex, 0, len(ruleListFromPlan))
		for _, r := range ruleListFromPlan {
			planned = append(planned, BulkPlannedRuleIndex{
				SectionName:    r.SectionName,
				RuleName:       r.RuleName,
				IndexInSection: r.IndexInSection,
			})
		}

		reorderIn, buildErr := buildPolicyReorderInput(sections, rules, planned)
		if buildErr != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"WAN firewall policy reorder",
				buildErr.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, buildErr
		}

		tflog.Debug(ctx, "PolicyWanFirewallReorderPolicy request", map[string]interface{}{
			"policyReorderInput": utils.InterfaceToJSONString(reorderIn),
		})

		var reorderOut *cato_go_sdk.PolicyWanFirewallReorderPolicy
		reorderErr := withAcctestPolicyRevisionCleanupRetryOnce(ctx, "PolicyWanFirewallReorderPolicy", func() error {
			return discardFirewallAndWANPolicyRevisions(ctx, r.client.catov2, r.client.AccountId)
		}, func() error {
			var callErr error
			reorderOut, callErr = r.wanBulkPolicy().PolicyWanFirewallReorderPolicy(
				ctx,
				&cato_models.WanFirewallPolicyMutationInput{},
				reorderIn,
				r.client.AccountId,
			)
			return wanFirewallReorderError(reorderOut, callErr)
		})
		if reorderErr != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyWanFirewallReorderPolicy error",
				reorderErr.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, reorderErr
		}

		tflog.Debug(ctx, "PolicyWanFirewallReorderPolicy response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(reorderOut),
		})

		// Build final state using API values for computed fields so they are always known post-apply.
		for _, ruleFromPlan := range ruleListFromPlan {
			ruleIndexStateData, diagsSection := buildWanRuleIndexStateData(
				ruleFromPlan,
				ruleNameIDMap,
				ruleNameDescriptionMap,
				ruleNameEnabledMap,
				"",
			)
			diags = append(diags, diagsSection...)
			ruleObjectMap[ruleFromPlan.RuleName] = ruleIndexStateData
		}
	}

	pubErr := withPolicyRevisionConflictRetry(ctx, "PolicyWanFirewallPublishPolicyRevision", func() error {
		_, errPub := r.wanBulkPolicy().PolicyWanFirewallPublishPolicyRevision(
			ctx,
			&cato_models.PolicyPublishRevisionInput{},
			r.client.AccountId,
		)
		return errPub
	})
	if pubErr != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			pubErr.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, pubErr
	}

	sectionObjectsMap, sectionMapDiags := types.MapValue(
		WanSectionIndexResourceObjectTypes,
		sectionObjectMap,
	)
	diags = append(diags, sectionMapDiags...)

	ruleObjectsMap, ruleMapDiags := types.MapValue(
		WanRuleIndexResourceObjectTypes,
		ruleObjectMap,
	)
	diags = append(diags, ruleMapDiags...)

	return sectionObjectsMap, ruleObjectsMap, diags, nil
}

func buildWanRuleIndexStateData(
	ruleFromPlan WanRulesRuleDataIndex,
	ruleNameIDMap map[string]string,
	ruleNameDescriptionMap map[string]string,
	ruleNameEnabledMap map[string]bool,
	subPolicyName string,
) (basetypes.ObjectValue, diag.Diagnostics) {
	subPolicyValue := types.StringNull()
	if subPolicyName != "" {
		subPolicyValue = types.StringValue(subPolicyName)
	}
	return types.ObjectValue(
		WanRuleIndexResourceAttrTypes,
		map[string]attr.Value{
			"id":               types.StringValue(ruleNameIDMap[ruleFromPlan.RuleName]),
			"index_in_section": types.Int64Value(ruleFromPlan.IndexInSection),
			"section_name":     types.StringValue(ruleFromPlan.SectionName),
			"sub_policy_name":  subPolicyValue,
			"rule_name":        types.StringValue(ruleFromPlan.RuleName),
			"description":      types.StringValue(ruleNameDescriptionMap[ruleFromPlan.RuleName]),
			"enabled":          types.BoolValue(ruleNameEnabledMap[ruleFromPlan.RuleName]),
		},
	)
}
