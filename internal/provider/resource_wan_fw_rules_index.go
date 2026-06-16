package provider

import (
	"context"
	"errors"
	"sort"

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
	client              *catoClientData
	wanRulesIndexClient WanRulesIndexClient
}

func (r *wanRulesIndexResource) getWanRulesIndexClient() WanRulesIndexClient {
	if r.wanRulesIndexClient != nil {
		return r.wanRulesIndexClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.catov2
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
	"rule_name":        types.StringType,
	"description":      types.StringType,
	"enabled":          types.BoolType,
}

var WanSectionIndexResourceObjectTypes = types.ObjectType{AttrTypes: WanSectionIndexResourceAttrTypes}
var WanSectionIndexResourceAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"section_name":  types.StringType,
	"section_index": types.Int64Type,
}

func (r *wanRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
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
			"Catov2 API PolicyWanFirewallReorderPolicy error",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(diags...)
	filteredSections, sectionFilterDiags := keepPlannedMapKeys(plan.SectionData, sectionObjectsList, WanSectionIndexResourceObjectTypes)
	resp.Diagnostics.Append(sectionFilterDiags...)
	filteredRules, ruleFilterDiags := keepPlannedMapKeys(plan.RuleData, rulesObjectsList, WanRuleIndexResourceObjectTypes)
	resp.Diagnostics.Append(ruleFilterDiags...)
	plan.SectionData = filteredSections
	plan.RuleData = filteredRules

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
			"Catov2 API PolicyWanFirewallReorderPolicy error",
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(diags...)
	filteredSections, sectionFilterDiags := keepPlannedMapKeys(plan.SectionData, sectionObjectsList, WanSectionIndexResourceObjectTypes)
	resp.Diagnostics.Append(sectionFilterDiags...)
	filteredRules, ruleFilterDiags := keepPlannedMapKeys(plan.RuleData, rulesObjectsList, WanRuleIndexResourceObjectTypes)
	resp.Diagnostics.Append(ruleFilterDiags...)
	plan.SectionData = filteredSections
	plan.RuleData = filteredRules

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

func keepPlannedMapKeys(planMap types.Map, actualMap basetypes.MapValue, objectType types.ObjectType) (types.Map, diag.Diagnostics) {
	if planMap.IsNull() || planMap.IsUnknown() {
		return actualMap, nil
	}

	planElements := planMap.Elements()
	actualElements := actualMap.Elements()
	filtered := make(map[string]attr.Value, len(planElements))

	for key, planValue := range planElements {
		if actualValue, ok := actualElements[key]; ok {
			filtered[key] = actualValue
			continue
		}
		filtered[key] = planValue
	}

	return types.MapValue(objectType, filtered)
}

//nolint:gocyclo,funlen
func (r *wanRulesIndexResource) moveWanRulesAndSections(
	ctx context.Context,
	plan WanRulesIndex,
) (sectionObjects basetypes.MapValue, ruleObjects basetypes.MapValue, diagnostics diag.Diagnostics, err error) {
	diags := []diag.Diagnostic{}
	ruleObjectMap := make(map[string]attr.Value)
	client := r.getWanRulesIndexClient()
	if client == nil {
		err := errors.New("wan rules index client is not configured")
		diags = append(diags, diag.NewErrorDiagnostic("Catov2 API EntityLookup error", err.Error()))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	mutationInput, err := ensureWanDraftMutationInput(ctx, client, r.client.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic("Catov2 API PolicyWanFirewall draft revision error", err.Error()))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	if plan.SectionToStartAfterID.ValueString() != "" {
		result, err := client.PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
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

	sectionIDList := make(map[string]string)
	sectionIndexAPIData, err := client.PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
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

	for _, v := range sectionIndexAPIData.Policy.WanFirewall.Policy.Sections {
		sectionIDList[v.Section.Name] = v.Section.ID
	}

	sectionListFromPlan := make([]WanRulesSectionDataIndex, 0)
	sectionDataMapElements := plan.SectionData.Elements()
	for _, sectionValue := range sectionDataMapElements {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex WanRulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionListFromPlan = append(sectionListFromPlan, WanRulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		})
	}

	sort.Slice(sectionListFromPlan, func(i, j int) bool {
		return sectionListFromPlan[i].SectionIndex < sectionListFromPlan[j].SectionIndex
	})

	tflog.Debug(ctx, "Processing sections sorted by section_index", map[string]interface{}{
		"sections": utils.InterfaceToJSONString(sectionListFromPlan),
	})

	sectionObjectMap := make(map[string]attr.Value)
	for _, workingSection := range sectionListFromPlan {
		sectionIndexStateData, sectionDiags := types.ObjectValue(
			WanSectionIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":            types.StringValue(sectionIDList[workingSection.SectionName]),
				"section_name":  types.StringValue(workingSection.SectionName),
				"section_index": types.Int64Value(workingSection.SectionIndex),
			},
		)
		diags = append(diags, sectionDiags...)
		sectionObjectMap[workingSection.SectionName] = sectionIndexStateData
	}

	listOfSectionNames := buildWanPolicySectionOrder(
		sectionIndexAPIData.Policy.WanFirewall.Policy.Sections,
		sectionListFromPlan,
		plan.SectionToStartAfterID.ValueString(),
	)
	needsReorder := len(sectionListFromPlan) > 0 || len(plan.RuleData.Elements()) > 0
	if needsReorder {
		ruleListFromPlan := make([]WanRulesRuleDataIndex, 0)
		ruleDataMapElements := plan.RuleData.Elements()
		for _, ruleValue := range ruleDataMapElements {
			ruleObject := ruleValue.(types.Object)
			var planSourceRuleIndex WanRulesRuleItemIndex
			diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

			ruleListFromPlan = append(ruleListFromPlan, WanRulesRuleDataIndex{
				IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
				RuleName:       planSourceRuleIndex.RuleName.ValueString(),
				SectionName:    planSourceRuleIndex.SectionName.ValueString(),
				Description:    planSourceRuleIndex.Description.ValueString(),
				Enabled:        planSourceRuleIndex.Enabled.ValueBool(),
			})
		}

		sectionIndexMap := make(map[string]int64, len(sectionListFromPlan))
		for _, section := range sectionListFromPlan {
			sectionIndexMap[section.SectionName] = section.SectionIndex
		}

		sort.Slice(ruleListFromPlan, func(i, j int) bool {
			if ruleListFromPlan[i].SectionName != ruleListFromPlan[j].SectionName {
				if len(sectionIndexMap) == 0 {
					return ruleListFromPlan[i].SectionName < ruleListFromPlan[j].SectionName
				}
				return sectionIndexMap[ruleListFromPlan[i].SectionName] < sectionIndexMap[ruleListFromPlan[j].SectionName]
			}
			return ruleListFromPlan[i].IndexInSection < ruleListFromPlan[j].IndexInSection
		})

		ruleNameIDData, err := client.PolicyWanFirewallRulesIndex(ctx, r.client.AccountId)
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
		topLevelRuleIDsBySection := make(map[string][]string)
		topLevelRuleIDBySectionAndName := make(map[string]string)
		for _, ruleNameIDDataItem := range ruleNameIDData.Policy.WanFirewall.Policy.Rules {
			sectionName := ruleNameIDDataItem.Rule.Section.Name
			ruleNameIDMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.ID
			ruleNameDescriptionMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.Description
			ruleNameEnabledMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.Enabled
			topLevelRuleIDsBySection[sectionName] = append(topLevelRuleIDsBySection[sectionName], ruleNameIDDataItem.Rule.ID)
			topLevelRuleIDBySectionAndName[sectionName+"\x00"+ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.ID
		}

		plannedRuleIDsBySection := make(map[string][]string)
		for _, ruleItemFromPlan := range ruleListFromPlan {
			ruleID, ok := topLevelRuleIDBySectionAndName[ruleItemFromPlan.SectionName+"\x00"+ruleItemFromPlan.RuleName]
			if !ok {
				ruleID, ok = ruleNameIDMap[ruleItemFromPlan.RuleName]
				if !ok {
					err := errors.New("failed to resolve rule ID for reorder operation")
					diags = append(diags, diag.NewErrorDiagnostic("Catov2 API PolicyWanFirewallReorderPolicy error", err.Error()))
					return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
				}
			}
			plannedRuleIDsBySection[ruleItemFromPlan.SectionName] = append(plannedRuleIDsBySection[ruleItemFromPlan.SectionName], ruleID)
		}
		plannedRuleIDsGlobal := make(map[string]struct{})
		for _, plannedIDs := range plannedRuleIDsBySection {
			for _, id := range plannedIDs {
				plannedRuleIDsGlobal[id] = struct{}{}
			}
		}

		reorderSections := make([]*cato_models.PolicyReorderSectionInput, 0, len(listOfSectionNames))
		for _, sectionName := range listOfSectionNames {
			sectionRuleIDs := topLevelRuleIDsBySection[sectionName]
			if len(plannedRuleIDsGlobal) > 0 {
				filtered := make([]string, 0, len(sectionRuleIDs))
				for _, id := range sectionRuleIDs {
					if _, moved := plannedRuleIDsGlobal[id]; moved {
						continue
					}
					filtered = append(filtered, id)
				}
				sectionRuleIDs = filtered
			}
			if plannedIDs := plannedRuleIDsBySection[sectionName]; len(plannedIDs) > 0 {
				sectionRuleIDs = append(append(make([]string, 0, len(plannedIDs)+len(sectionRuleIDs)), plannedIDs...), sectionRuleIDs...)
			}

			reorderRules := make([]*cato_models.PolicyReorderRuleInput, 0, len(sectionRuleIDs))
			for _, ruleID := range sectionRuleIDs {
				reorderRules = append(reorderRules, &cato_models.PolicyReorderRuleInput{
					Ref: &cato_models.PolicyElementRefInput{
						By:    cato_models.ObjectRefByID,
						Input: ruleID,
					},
				})
			}

			reorderSections = append(reorderSections, &cato_models.PolicyReorderSectionInput{
				Ref: &cato_models.PolicyElementRefInput{
					By:    cato_models.ObjectRefByID,
					Input: sectionIDList[sectionName],
				},
				Rules: reorderRules,
			})
		}

		reorderInput := cato_models.PolicyReorderInput{
			Sections: reorderSections,
		}
		reorderResult, _, err := wanReorderPolicyWithRetry(ctx, client, r.client.AccountId, mutationInput, reorderInput)
		if reorderErr := wanReorderPolicyError(reorderResult, err); reorderErr != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyWanFirewallReorderPolicy error",
				reorderErr.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, reorderErr
		}

		for _, ruleFromPlan := range ruleListFromPlan {
			ruleIndexStateData, ruleDiags := buildWanRuleIndexStateData(
				ruleFromPlan,
				ruleNameIDMap,
				ruleNameDescriptionMap,
				ruleNameEnabledMap,
			)
			diags = append(diags, ruleDiags...)
			ruleObjectMap[ruleFromPlan.RuleName] = ruleIndexStateData
		}
	}

	_, err = client.PolicyWanFirewallPublishPolicyRevision(ctx, &cato_models.PolicyPublishRevisionInput{}, r.client.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
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
) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		WanRuleIndexResourceAttrTypes,
		map[string]attr.Value{
			"id":               types.StringValue(ruleNameIDMap[ruleFromPlan.RuleName]),
			"index_in_section": types.Int64Value(ruleFromPlan.IndexInSection),
			"section_name":     types.StringValue(ruleFromPlan.SectionName),
			"rule_name":        types.StringValue(ruleFromPlan.RuleName),
			"description":      types.StringValue(ruleNameDescriptionMap[ruleFromPlan.RuleName]),
			"enabled":          types.BoolValue(ruleNameEnabledMap[ruleFromPlan.RuleName]),
		},
	)
}
