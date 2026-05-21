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
	_ resource.Resource              = &ifwRulesIndexResource{}
	_ resource.ResourceWithConfigure = &ifwRulesIndexResource{}
)

const (
	ifwRulePositionFirstInSection = "FIRST_IN_SECTION"
	ifwRulePositionAfterRule      = "AFTER_RULE"
	reorderPolicyAPIErrorMsg      = "reorder policy failed with api error"
)

func NewIfwRulesIndexResource() resource.Resource {
	return &ifwRulesIndexResource{}
}

type ifwRulesIndexResource struct {
	client *catoClientData
}

func (r *ifwRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_if_move_rule"
}

func (r *ifwRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves index values for Internet Firewall Rules.",
		Attributes: map[string]schema.Attribute{
			"section_to_start_after_id": schema.StringAttribute{
				Description: "IFW rule id",
				Required:    false,
				Optional:    true,
				// Computed:    true,
			},
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of IF Rule Policy Indexes keyed by rule_name",
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
						"index_in_parent": schema.Int64Attribute{
							Description: "Index value remapped within parent sub-policy rule",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"parent_rule_name": schema.StringAttribute{
							Description: "Parent sub-policy rule name for sub-rules",
							Required:    false,
							Optional:    true,
						},
						"section_name": schema.StringAttribute{
							Description: "IFW section name housing rule",
							Required:    false,
							Optional:    true,
						},
						"rule_name": schema.StringAttribute{
							Description: "IFW rule name housing rule",
							Required:    false,
							Optional:    true,
						},
						"description": schema.StringAttribute{
							Description: "IFW rule description",
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

var IfwRuleIndexResourceObjectTypes = types.ObjectType{AttrTypes: IfwRuleIndexResourceAttrTypes}
var IfwRuleIndexResourceAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"index_in_section": types.Int64Type,
	"index_in_parent":  types.Int64Type,
	"parent_rule_name": types.StringType,
	"section_name":     types.StringType,
	"rule_name":        types.StringType,
	"description":      types.StringType,
	"enabled":          types.BoolType,
}

var IfwSectionIndexResourceObjectTypes = types.ObjectType{AttrTypes: IfwSectionIndexResourceAttrTypes}
var IfwSectionIndexResourceAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"section_name":  types.StringType,
	"section_index": types.Int64Type,
}

func (r *ifwRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

// func (r *ifwRulesIndexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	// Retrieve import ID and save to id attribute
// 	// resource.ImportStatePassthroughID(ctx, path.Root("Id"), req, resp)
// }

func (r *ifwRulesIndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IfwRulesIndex
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionObjectsList, rulesObjectsList, diags, err := r.moveIfwRulesAndSections(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallMoveSection error",
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

//nolint:funlen
func (r *ifwRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IfwRulesIndex
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create maps to store API data for ID lookups
	sectionIDMap := make(map[string]string)
	ruleIDMap := make(map[string]string)
	ruleDescriptionMap := make(map[string]string)
	ruleEnabledMap := make(map[string]bool)

	// Get current sections from API to get fresh IDs
	sectionIndexAPIData, err := r.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallSectionsIndex error",
			err.Error(),
		)
		return
	}

	// Build map of section names to IDs from API
	for _, v := range sectionIndexAPIData.Policy.InternetFirewall.Policy.Sections {
		sectionIDMap[v.Section.Name] = v.Section.ID
	}

	// Get current rules from API to get fresh IDs and computed values
	ruleIndexAPIData, err := r.client.catov2.PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallRulesIndex error",
			err.Error(),
		)
		return
	}

	// Build maps of rule names to API data
	for _, v := range ruleIndexAPIData.Policy.InternetFirewall.Policy.Rules {
		ruleIDMap[v.Rule.Name] = v.Rule.ID
		ruleDescriptionMap[v.Rule.Name] = v.Rule.Description
		ruleEnabledMap[v.Rule.Name] = v.Rule.Enabled
	}

	// Update only computed fields in existing state, preserving planned configuration
	sectionObjectMap := make(map[string]attr.Value)
	if !state.SectionData.IsNull() && !state.SectionData.IsUnknown() {
		sectionDataMapElements := state.SectionData.Elements()
		for sectionName, sectionValue := range sectionDataMapElements {
			sectionObject := sectionValue.(types.Object)
			var existingSection IfwRulesSectionItemIndex
			diags = append(diags, sectionObject.As(ctx, &existingSection, basetypes.ObjectAsOptions{})...)

			// Preserve planned values, only update computed ID
			sectionIndexStateData, sectionDiags := types.ObjectValue(
				IfwSectionIndexResourceAttrTypes,
				map[string]attr.Value{
					"id":            types.StringValue(sectionIDMap[sectionName]),
					"section_name":  existingSection.SectionName,  // Preserve planned value
					"section_index": existingSection.SectionIndex, // Preserve planned value
				},
			)
			diags = append(diags, sectionDiags...)
			sectionObjectMap[sectionName] = sectionIndexStateData
		}
	}

	sectionObjectsMap, sectionMapDiags := types.MapValue(
		IfwSectionIndexResourceObjectTypes,
		sectionObjectMap,
	)
	diags = append(diags, sectionMapDiags...)
	state.SectionData = sectionObjectsMap

	// Update only computed fields in existing rule state, preserving planned configuration
	ruleObjectMap := make(map[string]attr.Value)
	if !state.RuleData.IsNull() && !state.RuleData.IsUnknown() {
		ruleDataMapElements := state.RuleData.Elements()
		for ruleName, ruleValue := range ruleDataMapElements {
			ruleObject := ruleValue.(types.Object)
			var existingRule IfwRulesRuleItemIndex
			diags = append(diags, ruleObject.As(ctx, &existingRule, basetypes.ObjectAsOptions{})...)

			// Preserve planned values, only update computed fields
			ruleIndexStateData, ruleDiags := types.ObjectValue(
				IfwRuleIndexResourceAttrTypes,
				map[string]attr.Value{
					"id":               types.StringValue(ruleIDMap[ruleName]),
					"index_in_section": existingRule.IndexInSection,                     // Preserve planned value
					"index_in_parent":  existingRule.IndexInParent,                      // Preserve planned value
					"parent_rule_name": existingRule.ParentRuleName,                     // Preserve planned value
					"section_name":     existingRule.SectionName,                        // Preserve planned value
					"rule_name":        existingRule.RuleName,                           // Preserve planned value
					"description":      types.StringValue(ruleDescriptionMap[ruleName]), // Update computed value
					"enabled":          types.BoolValue(ruleEnabledMap[ruleName]),       // Update computed value
				},
			)
			diags = append(diags, ruleDiags...)
			ruleObjectMap[ruleName] = ruleIndexStateData
		}
	}

	ruleObjectsMap, ruleMapDiags := types.MapValue(
		IfwRuleIndexResourceObjectTypes,
		ruleObjectMap,
	)
	diags = append(diags, ruleMapDiags...)
	state.RuleData = ruleObjectsMap

	resp.Diagnostics.Append(diags...)
	if diags := resp.State.Set(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *ifwRulesIndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IfwRulesIndex
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionObjectsList, rulesObjectsList, diags, err := r.moveIfwRulesAndSections(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallMoveSection error",
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

func (r *ifwRulesIndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IfwRulesIndex
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

//nolint:gocyclo,funlen,gocritic
func (r *ifwRulesIndexResource) moveIfwRulesAndSections(
	ctx context.Context,
	plan IfwRulesIndex,
) (basetypes.MapValue, basetypes.MapValue, diag.Diagnostics, error) {
	diags := []diag.Diagnostic{}

	ruleObjectMap := make(map[string]attr.Value)

	if plan.SectionToStartAfterID.ValueString() != "" {
		result, err := r.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
		tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic("Catov2 API PolicyInternetFirewallSectionsIndex error", err.Error()))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}
		isPresent := false
		for _, item := range result.Policy.InternetFirewall.Policy.Sections {
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
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, errors.New("SectionToStartAfterID not found")
		}
	}

	// maps section_name -> sectionID
	// initially used to find ID of Default section
	sectionIDList := make(map[string]string)
	sectionIndexAPIData, err := r.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexAPIData),
	})
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API EntityLookup error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	for _, v := range sectionIndexAPIData.Policy.InternetFirewall.Policy.Sections {
		sectionIDList[v.Section.Name] = v.Section.ID
	}

	// Check if no sections exist and create a default section for P2P and system rules
	if len(sectionIndexAPIData.Policy.InternetFirewall.Policy.Sections) == 0 {
		input := cato_models.PolicyAddSectionInput{
			At: &cato_models.PolicySectionPositionInput{
				Position: cato_models.PolicySectionPositionEnumLastInPolicy,
			},
			Section: &cato_models.PolicyAddSectionInfoInput{
				Name: "Default Outbound Internet",
			},
		}
		sectionCreateAPIData, err := r.client.catov2.PolicyInternetFirewallAddSection(
			ctx,
			&cato_models.InternetFirewallPolicyMutationInput{},
			input,
			r.client.AccountId,
		)
		tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSectionWithinBulkMove.response", map[string]interface{}{
			"reason":   "creating new section as IFW does not have a default listed",
			"response": utils.InterfaceToJSONString(sectionCreateAPIData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API EntityLookup error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}
	}

	sectionListFromPlan := make([]IfwRulesSectionDataIndex, 0)

	// Convert map to slice for processing
	sectionDataMapElements := plan.SectionData.Elements()
	for _, sectionValue := range sectionDataMapElements {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex IfwRulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionDataTmp := IfwRulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		}
		sectionListFromPlan = append(sectionListFromPlan, sectionDataTmp)
	}

	// Sort sections by SectionIndex to ensure proper ordering
	sort.Slice(sectionListFromPlan, func(i, j int) bool {
		return sectionListFromPlan[i].SectionIndex < sectionListFromPlan[j].SectionIndex
	})

	tflog.Debug(ctx, "Processing sections in plan order (already sorted by module)", map[string]interface{}{
		"sections": utils.InterfaceToJSONString(sectionListFromPlan),
	})

	currentSectionID := ""

	sectionObjectMap := make(map[string]attr.Value)

	// create the sections from the list provided following the section ID provided in firstSectionID
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
		sectionMoveAPIData, err := r.client.catov2.PolicyInternetFirewallMoveSection(ctx, nil, policyMoveSectionInputInt, r.client.AccountId)
		// Check for API errors safely with nil checks
		if sectionMoveAPIData != nil && sectionMoveAPIData.GetPolicy() != nil &&
			sectionMoveAPIData.GetPolicy().InternetFirewall != nil &&
			sectionMoveAPIData.GetPolicy().InternetFirewall.GetMoveSection() != nil &&
			len(sectionMoveAPIData.GetPolicy().InternetFirewall.GetMoveSection().Errors) != 0 {
			tflog.Warn(ctx, "Write.PolicyInternetFirewallMoveSectionMoveSection.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(sectionMoveAPIData),
			})
			if err != nil {
				diags = append(diags, diag.NewErrorDiagnostic(
					"Catov2 API EntityLookup error",
					err.Error(),
				))
				return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
			}
		}
		tflog.Warn(ctx, "Write.PolicyInternetFirewallMoveSection.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(sectionMoveAPIData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API EntityLookup error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}

		sectionIndexStateData, diagsSection := types.ObjectValue(
			IfwSectionIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":            types.StringValue(sectionIDList[workingSectionName.SectionName]),
				"section_name":  types.StringValue(workingSectionName.SectionName),
				"section_index": types.Int64Value(workingSectionName.SectionIndex),
			},
		)
		diags = append(diags, diagsSection...)

		sectionObjectMap[workingSectionName.SectionName] = sectionIndexStateData

		currentSectionID = sectionIDList[workingSectionName.SectionName]
	}

	// now that the sections are ordered properly, move the rules to the correct locations
	if len(plan.RuleData.Elements()) > 0 {
		// get all of the list elements from the plan
		ruleListFromPlan := make([]IfwRulesRuleDataIndex, 0)

		// Convert map to slice for processing
		ruleDataMapElements := plan.RuleData.Elements()
		for _, ruleValue := range ruleDataMapElements {
			ruleObject := ruleValue.(types.Object)
			var planSourceRuleIndex IfwRulesRuleItemIndex
			diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

			rulenDataTmp := IfwRulesRuleDataIndex{
				IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
				IndexInParent:  planSourceRuleIndex.IndexInParent.ValueInt64(),
				ParentRuleName: planSourceRuleIndex.ParentRuleName.ValueString(),
				RuleName:       planSourceRuleIndex.RuleName.ValueString(),
				SectionName:    planSourceRuleIndex.SectionName.ValueString(),
				Description:    planSourceRuleIndex.Description.ValueString(),
				Enabled:        planSourceRuleIndex.Enabled,
			}
			ruleListFromPlan = append(ruleListFromPlan, rulenDataTmp)
		}

		// Create a map of section name to section index for proper ordering
		sectionIndexMap := make(map[string]int64)
		for _, section := range sectionListFromPlan {
			sectionIndexMap[section.SectionName] = section.SectionIndex
		}

		// Sort entries so deterministic order is preserved for:
		// - top-level rules: section index + index_in_section
		// - sub-rules: parent_rule_name + index_in_parent
		sort.Slice(ruleListFromPlan, func(i, j int) bool {
			isSubI := ruleListFromPlan[i].ParentRuleName != ""
			isSubJ := ruleListFromPlan[j].ParentRuleName != ""
			if isSubI != isSubJ {
				return !isSubI // keep top-level rules first
			}
			if isSubI {
				if ruleListFromPlan[i].ParentRuleName != ruleListFromPlan[j].ParentRuleName {
					return ruleListFromPlan[i].ParentRuleName < ruleListFromPlan[j].ParentRuleName
				}
				return ruleListFromPlan[i].IndexInParent < ruleListFromPlan[j].IndexInParent
			}
			if ruleListFromPlan[i].SectionName != ruleListFromPlan[j].SectionName {
				return sectionIndexMap[ruleListFromPlan[i].SectionName] < sectionIndexMap[ruleListFromPlan[j].SectionName]
			}
			return ruleListFromPlan[i].IndexInSection < ruleListFromPlan[j].IndexInSection
		})

		fullPolicy, err := r.client.catov2.PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.client.AccountId)
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyInternetFirewall error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}
		ruleNameIDMap := make(map[string]string)
		ruleNameDescriptionMap := make(map[string]string)
		ruleNameEnabledMap := make(map[string]bool)
		topLevelRuleIDsBySection := make(map[string][]string)
		topLevelRuleIDBySectionAndName := make(map[string]string)
		topLevelRuleIDByName := make(map[string]string)
		subRuleIDsByParent := make(map[string][]string)
		subRuleIDByParentAndName := make(map[string]string)
		currentSubPolicyParentID := ""
		for _, apiRulePayload := range fullPolicy.Policy.InternetFirewall.Policy.Rules {
			apiRule := apiRulePayload.Rule
			ruleNameIDMap[apiRule.Name] = apiRule.ID
			ruleNameDescriptionMap[apiRule.Name] = apiRule.Description
			ruleNameEnabledMap[apiRule.Name] = apiRule.Enabled
			if apiRule.Section.ID != "" {
				topLevelRuleIDsBySection[apiRule.Section.Name] = append(topLevelRuleIDsBySection[apiRule.Section.Name], apiRule.ID)
				topLevelRuleIDBySectionAndName[apiRule.Section.Name+"\x00"+apiRule.Name] = apiRule.ID
				topLevelRuleIDByName[apiRule.Name] = apiRule.ID
				if apiRule.Action == cato_models.InternetFirewallActionEnumSubPolicy {
					currentSubPolicyParentID = apiRule.ID
				} else {
					currentSubPolicyParentID = ""
				}
				continue
			}
			if currentSubPolicyParentID == "" {
				continue
			}
			subRuleIDsByParent[currentSubPolicyParentID] = append(subRuleIDsByParent[currentSubPolicyParentID], apiRule.ID)
			subRuleIDByParentAndName[currentSubPolicyParentID+"\x00"+apiRule.Name] = apiRule.ID
		}

		tflog.Debug(ctx, "Processing rules in correct order", map[string]interface{}{
			"ruleListFromPlan": utils.InterfaceToJSONString(ruleListFromPlan),
			"ruleNameIDMap":    utils.InterfaceToJSONString(ruleNameIDMap),
		})

		plannedRuleIDsBySection := make(map[string][]string)
		plannedSubRuleIDsByParent := make(map[string][]string)
		reorderSections := make([]*cato_models.PolicyReorderSectionInput, 0, len(sectionIndexAPIData.Policy.InternetFirewall.Policy.Sections))
		for _, ruleItemFromPlan := range ruleListFromPlan {
			if ruleItemFromPlan.ParentRuleName != "" {
				parentRuleID := topLevelRuleIDByName[ruleItemFromPlan.ParentRuleName]
				if parentRuleID == "" {
					diags = append(diags, diag.NewWarningDiagnostic(
						"Skipped sub-rule in IF reorder",
						"Parent sub-policy rule '"+ruleItemFromPlan.ParentRuleName+"' was not found for sub-rule '"+ruleItemFromPlan.RuleName+"'.",
					))
					continue
				}
				subRuleID := subRuleIDByParentAndName[parentRuleID+"\x00"+ruleItemFromPlan.RuleName]
				if subRuleID == "" {
					diags = append(diags, diag.NewWarningDiagnostic(
						"Skipped sub-rule in IF reorder",
						"Sub-rule '"+ruleItemFromPlan.RuleName+"' was not found under parent '"+ruleItemFromPlan.ParentRuleName+"'.",
					))
					continue
				}
				plannedSubRuleIDsByParent[parentRuleID] = append(plannedSubRuleIDsByParent[parentRuleID], subRuleID)
				continue
			}
			ruleID, ok := topLevelRuleIDBySectionAndName[ruleItemFromPlan.SectionName+"\x00"+ruleItemFromPlan.RuleName]
			if !ok {
				diags = append(diags, diag.NewWarningDiagnostic(
					"Skipped rule in IF reorder",
					"Rule '"+ruleItemFromPlan.RuleName+"' was not found as a top-level rule in section '"+ruleItemFromPlan.SectionName+"'.",
				))
				continue
			}
			plannedRuleIDsBySection[ruleItemFromPlan.SectionName] = append(plannedRuleIDsBySection[ruleItemFromPlan.SectionName], ruleID)
		}

		for _, section := range sectionIndexAPIData.Policy.InternetFirewall.Policy.Sections {
			sectionName := section.Section.Name
			sectionRuleIDs := topLevelRuleIDsBySection[sectionName]
			if plannedIDs := plannedRuleIDsBySection[sectionName]; len(plannedIDs) > 0 {
				plannedSet := make(map[string]struct{}, len(plannedIDs))
				for _, id := range plannedIDs {
					plannedSet[id] = struct{}{}
				}
				remaining := make([]string, 0, len(sectionRuleIDs))
				for _, id := range sectionRuleIDs {
					if _, managed := plannedSet[id]; managed {
						continue
					}
					remaining = append(remaining, id)
				}
				sectionRuleIDs = append(plannedIDs, remaining...)
			}

			reorderRules := make([]*cato_models.PolicyReorderRuleInput, 0, len(sectionRuleIDs))
			for _, ruleID := range sectionRuleIDs {
				reorderRule := &cato_models.PolicyReorderRuleInput{
					Ref: &cato_models.PolicyElementRefInput{
						By:    cato_models.ObjectRefByID,
						Input: ruleID,
					},
				}
				if plannedSubRules := plannedSubRuleIDsByParent[ruleID]; len(plannedSubRules) > 0 {
					plannedSet := make(map[string]struct{}, len(plannedSubRules))
					for _, id := range plannedSubRules {
						plannedSet[id] = struct{}{}
					}
					remainingSubRules := make([]string, 0, len(subRuleIDsByParent[ruleID]))
					for _, id := range subRuleIDsByParent[ruleID] {
						if _, managed := plannedSet[id]; managed {
							continue
						}
						remainingSubRules = append(remainingSubRules, id)
					}
					finalSubRuleIDs := append(plannedSubRules, remainingSubRules...)
					reorderRule.SubRules = make([]*cato_models.PolicyReorderSubRuleInput, 0, len(finalSubRuleIDs))
					for _, subRuleID := range finalSubRuleIDs {
						reorderRule.SubRules = append(reorderRule.SubRules, &cato_models.PolicyReorderSubRuleInput{
							Ref: &cato_models.PolicyElementRefInput{
								By:    cato_models.ObjectRefByID,
								Input: subRuleID,
							},
						})
					}
				}
				reorderRules = append(reorderRules, reorderRule)
			}

			reorderSections = append(reorderSections, &cato_models.PolicyReorderSectionInput{
				Ref: &cato_models.PolicyElementRefInput{
					By:    cato_models.ObjectRefByID,
					Input: section.Section.ID,
				},
				Rules: reorderRules,
			})
		}
		reorderInput := cato_models.PolicyReorderInput{
			Sections: reorderSections,
		}

		reorderResult, err := r.client.catov2.PolicyInternetFirewallReorderPolicy(
			ctx,
			&cato_models.InternetFirewallPolicyMutationInput{},
			reorderInput,
			r.client.AccountId,
		)
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyInternetFirewallReorderPolicy error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}
		reorderPayload := reorderResult.GetPolicy().GetInternetFirewall().GetReorderPolicy()
		if reorderPayload != nil && reorderPayload.GetStatus() != nil && *reorderPayload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
			apiErrors := reorderPayload.GetErrors()
			if len(apiErrors) > 0 {
				errMsg := reorderPolicyAPIErrorMsg
				if apiErrors[0].GetErrorMessage() != nil {
					errMsg = *apiErrors[0].GetErrorMessage()
				}
				err := errors.New(errMsg)
				diags = append(diags, diag.NewErrorDiagnostic("Catov2 API PolicyInternetFirewallReorderPolicy error", err.Error()))
				return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
			}
			err := errors.New("reorder policy failed")
			diags = append(diags, diag.NewErrorDiagnostic("Catov2 API PolicyInternetFirewallReorderPolicy error", err.Error()))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}

		// Build final state using API values for computed fields so they are always known post-apply.
		for _, ruleFromPlan := range ruleListFromPlan {
			ruleIndexStateData, ruleDiags := buildIfwRuleIndexStateData(
				ruleFromPlan,
				ruleNameIDMap,
				ruleNameDescriptionMap,
				ruleNameEnabledMap,
			)
			diags = append(diags, ruleDiags...)
			ruleObjectMap[ruleFromPlan.RuleName] = ruleIndexStateData
		}
	}

	// Publish changes
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		&cato_models.PolicyPublishRevisionInput{},
		r.client.AccountId,
	)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	sectionObjectsMap, sectionMapDiags := types.MapValue(
		IfwSectionIndexResourceObjectTypes,
		sectionObjectMap,
	)
	diags = append(diags, sectionMapDiags...)

	ruleObjectsMap, ruleMapDiags := types.MapValue(
		IfwRuleIndexResourceObjectTypes,
		ruleObjectMap,
	)
	diags = append(diags, ruleMapDiags...)

	return sectionObjectsMap, ruleObjectsMap, diags, nil
}

func buildIfwRuleIndexStateData(
	ruleFromPlan IfwRulesRuleDataIndex,
	ruleNameIDMap map[string]string,
	ruleNameDescriptionMap map[string]string,
	ruleNameEnabledMap map[string]bool,
) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		IfwRuleIndexResourceAttrTypes,
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
