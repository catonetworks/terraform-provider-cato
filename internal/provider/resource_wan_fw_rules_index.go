package provider

import (
	"context"
	"errors"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
	client *catoClientData
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
			"rule_data": schema.ListNestedAttribute{
				Description: "List of WAN Rule Policy Indexes",
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
							Description: "WAN rule enabled",
							Required:    false,
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"section_data": schema.ListNestedAttribute{
				Description: "List of IFW section Indexes",
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

func (r *wanRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
			"Catov2 API PolicyWanFirewallMoveSection error",
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

	var sectionObjects []attr.Value
	var ruleObjects []attr.Value

	sectionIndexApiData, err := r.client.catov2.PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Warn(ctx, "Read.PolicyWanFirewallSectionsIndexInRead.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexApiData),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	sectionCount := int64(1)

	sectionIndexListCount := make(map[string]int64)

	// pass in the api sections list during read to the state
	for _, v := range sectionIndexApiData.Policy.WanFirewall.Policy.Sections {
		sectionIndexListCount[v.Section.Name] = 0
		sectionIndexStateData, diags := types.ObjectValue(
			WanSectionIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":            types.StringValue(v.Section.ID),
				"section_name":  types.StringValue(v.Section.Name),
				"section_index": types.Int64Value(sectionCount),
			},
		)
		resp.Diagnostics.Append(diags...)
		sectionObjects = append(sectionObjects, sectionIndexStateData)
		sectionCount++
	}

	sectionObjectsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: WanSectionIndexResourceAttrTypes,
		},
		sectionObjects,
	)
	resp.Diagnostics.Append(diags...)
	state.SectionData = sectionObjectsList

	ruleIndexApiData, err := r.client.catov2.PolicyWanFirewallRulesIndex(ctx, r.client.AccountId)
	tflog.Warn(ctx, "Read.PolicyWanFirewallRulesIndex.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ruleIndexApiData),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	for _, v := range ruleIndexApiData.Policy.WanFirewall.Policy.Rules {
		sectionIndexListCount[v.Rule.Section.Name]++
		ruleIndexStateData, diags := types.ObjectValue(
			WanRuleIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":               types.StringValue(v.Rule.ID),
				"index_in_section": types.Int64Value(sectionIndexListCount[v.Rule.Section.Name]),
				"section_name":     types.StringValue(v.Rule.Section.Name),
				"rule_name":        types.StringValue(v.Rule.Name),
				"description":      types.StringValue(v.Rule.Description),
				"enabled":          types.BoolValue(v.Rule.Enabled),
			},
		)
		resp.Diagnostics.Append(diags...)
		ruleObjects = append(ruleObjects, ruleIndexStateData)
	}

	ruleObjectsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: WanRuleIndexResourceAttrTypes,
		},
		ruleObjects,
	)
	resp.Diagnostics.Append(diags...)
	state.RuleData = ruleObjectsList

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
			"Catov2 API PolicyWanFirewallMoveSection error",
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

func (r *wanRulesIndexResource) moveWanRulesAndSections(ctx context.Context, plan WanRulesIndex) (basetypes.ListValue, basetypes.ListValue, diag.Diagnostics, error) {
	diags := []diag.Diagnostic{}

	if len(plan.SectionToStartAfterId.ValueString()) > 0 {
		result, err := r.client.catov2.PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
		tflog.Debug(ctx, "Read.PolicyWanFirewallSectionsIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic("Catov2 API PolicyWanFirewallSectionsIndex error", err.Error()))
			return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
		}
		isPresent := false
		for _, item := range result.Policy.WanFirewall.Policy.Sections {
			section_id := cast.ToString(item.Section.ID)
			if section_id == plan.SectionToStartAfterId.ValueString() {
				isPresent = true
				break
			}
		}
		if !isPresent {
			diags = append(diags, diag.NewErrorDiagnostic(
				"SectionToStartAfterId '"+plan.SectionToStartAfterId.ValueString()+"' not found",
				"Please check the section ID and try again.",
			))
			return basetypes.ListValue{}, basetypes.ListValue{}, diags, errors.New("sectionToStartAfterId not found")
		}
	}

	// as the name indicates, a slice of string containing WF sections names
	listOfSectionNames := make([]string, 0)

	// maps section_name -> section_id
	// initially used to find ID of Default section
	sectionIdList := make(map[string]string)
	sectionIndexApiData, err := r.client.catov2.PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Warn(ctx, "Read.PolicyWanFirewallSectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexApiData),
	})
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API EntityLookup error",
			err.Error(),
		))
		return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
	}

	// for easier processing, a map of section name to ID is created
	for _, v := range sectionIndexApiData.Policy.WanFirewall.Policy.Sections {
		sectionIdList[v.Section.Name] = v.Section.ID
	}

	sectionListFromPlan := make([]WanRulesSectionDataIndex, 0)

	sectionDataFromPlanList := make([]types.Object, 0, len(plan.SectionData.Elements()))
	diags = append(diags, plan.SectionData.ElementsAs(ctx, &sectionDataFromPlanList, false)...)
	var sectionSourceRuleIndex WanRulesSectionItemIndex
	for _, item := range sectionDataFromPlanList {
		diags = append(diags, item.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionDataTmp := WanRulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		}
		sectionListFromPlan = append(sectionListFromPlan, sectionDataTmp)

	}

	tflog.Debug(ctx, "Processing sections in plan order (already sorted by module)", map[string]interface{}{
		"sections": utils.InterfaceToJSONString(sectionListFromPlan),
	})

	currentSectionId := ""

	var sectionObjects []attr.Value

	// create the sections from the list provided following the section ID provided in firstSectionId
	for _, workingSectionName := range sectionListFromPlan {

		listOfSectionNames = append(listOfSectionNames, workingSectionName.SectionName)
		policyMoveSectionInputInt := cato_models.PolicyMoveSectionInput{
			ID: sectionIdList[workingSectionName.SectionName],
		}

		// For the first element, check for sectionToStartAfterId, if not, start at last LAST_IN_POLICY
		// initializing currentSectionId to the SectionToStartAfterId otherwise set to id of first section for next in list
		if currentSectionId == "" {
			if plan.SectionToStartAfterId.ValueString() != "" {
				policyMoveSectionInputInt.To = &cato_models.PolicySectionPositionInput{
					Ref:      plan.SectionToStartAfterId.ValueStringPointer(),
					Position: "AFTER_SECTION",
				}
			} else {
				policyMoveSectionInputInt.To = &cato_models.PolicySectionPositionInput{
					Position: "LAST_IN_POLICY",
				}
			}
		} else {
			policyMoveSectionInputInt.To = &cato_models.PolicySectionPositionInput{
				Ref:      &currentSectionId,
				Position: "AFTER_SECTION",
			}
		}
		tflog.Warn(ctx, "Write.policyMoveSectionInputInt.response", map[string]interface{}{
			"sectionToStartAfterId":          plan.SectionToStartAfterId.ValueString(),
			"moveFrom":                       workingSectionName.SectionName,
			"toAfter":                        currentSectionId,
			"sectionIdList":                  sectionIdList,
			"workingSectionName.SectionName": workingSectionName.SectionName,
			"sectionIdList[workingSectionName.SectionName]": sectionIdList[workingSectionName.SectionName],
			"response": utils.InterfaceToJSONString(policyMoveSectionInputInt),
		})
		sectionMoveApiData, err := r.client.catov2.PolicyWanFirewallMoveSection(ctx, policyMoveSectionInputInt, r.client.AccountId)
		// Check for API errors safely with nil checks
		if sectionMoveApiData != nil && sectionMoveApiData.GetPolicy() != nil &&
			sectionMoveApiData.GetPolicy().WanFirewall != nil &&
			sectionMoveApiData.GetPolicy().WanFirewall.GetMoveSection() != nil &&
			len(sectionMoveApiData.GetPolicy().WanFirewall.GetMoveSection().Errors) != 0 {
			tflog.Warn(ctx, "Write.PolicyWanFirewallMoveSectionMoveSection.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(sectionMoveApiData),
			})
			if err != nil {
				diags = append(diags, diag.NewErrorDiagnostic(
					"Catov2 API EntityLookup error",
					err.Error(),
				))
				return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
			}
		}
		tflog.Warn(ctx, "Write.PolicyWanFirewallMoveSection.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(sectionMoveApiData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API EntityLookup error",
				err.Error(),
			))
			return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
		}

		sectionIndexStateData, diagsSection := types.ObjectValue(
			WanSectionIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":            types.StringValue(sectionIdList[workingSectionName.SectionName]),
				"section_name":  types.StringValue(workingSectionName.SectionName),
				"section_index": types.Int64Value(workingSectionName.SectionIndex),
			},
		)
		diags = append(diags, diagsSection...)

		sectionObjects = append(sectionObjects, sectionIndexStateData)

		currentSectionId = sectionIdList[workingSectionName.SectionName]
	}

	// now that the sections are ordered properly, move the rules to the correct locations
	if len(plan.RuleData.Elements()) > 0 {
		// get all of the list elements from the plan
		ruleListFromPlan := make([]WanRulesRuleDataIndex, 0)

		ruleDataFromPlanList := make([]types.Object, 0, len(plan.SectionData.Elements()))
		diags = append(diags, plan.RuleData.ElementsAs(ctx, &ruleDataFromPlanList, false)...)

		var planSourceRuleIndex WanRulesRuleItemIndex
		for _, item := range ruleDataFromPlanList {
			diags = append(diags, item.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

			rulenDataTmp := WanRulesRuleDataIndex{
				IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
				RuleName:       planSourceRuleIndex.RuleName.ValueString(),
				SectionName:    planSourceRuleIndex.SectionName.ValueString(),
				Description:    planSourceRuleIndex.Description.ValueString(),
				Enabled:        planSourceRuleIndex.Enabled.ValueBool(),
			}
			ruleListFromPlan = append(ruleListFromPlan, rulenDataTmp)
			tflog.Warn(ctx, "Read.rulenDataTmp.response", map[string]interface{}{
				"rulenDataTmp": utils.InterfaceToJSONString(rulenDataTmp),
			})
		}

		tflog.Warn(ctx, "Read.ruleListFromPlan.response", map[string]interface{}{
			"ruleListFromPlan": utils.InterfaceToJSONString(ruleListFromPlan),
		})

		ruleNameIdData, err := r.client.catov2.PolicyWanFirewallRulesIndex(ctx, r.client.AccountId)
		tflog.Warn(ctx, "Read.PolicyWanFirewallRulesIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(ruleNameIdData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API EntityLookup error",
				err.Error(),
			))
			return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
		}
		ruleNameIdMap := make(map[string]string)

		// create map of IFW rule names from the API to their IDs for easy lookup
		for _, ruleNameIdDataItem := range ruleNameIdData.Policy.WanFirewall.Policy.Rules {
			ruleNameIdMap[ruleNameIdDataItem.Rule.Name] = ruleNameIdDataItem.Rule.ID
		}

		tflog.Warn(ctx, "Read.ruleNameIdMap.response", map[string]interface{}{
			"ruleNameIdMap": utils.InterfaceToJSONString(ruleNameIdMap),
		})

		// loop through the ordered list of section names
		for _, sectionNameItem := range listOfSectionNames {
			tflog.Warn(ctx, "Read.ProcessingSectionFromList.response", map[string]interface{}{
				"sectionNameItem": sectionNameItem,
				"ruleNameIdMap":   utils.InterfaceToJSONString(listOfSectionNames),
			})

			// for easier processing and visualization, we are creating two maps
			// 1 - mapRuleIndexToRuleName
			//   this maps the rule index in section to the rule name
			// 2 - mapRuleIndexToSectionName
			//  this maps the rule index in section to the section name
			mapRuleIndexToRuleName := make(map[int64]string)
			mapRuleIndexToSectionName := make(map[int64]string)

			for _, ruleItemFromPlan := range ruleListFromPlan {
				tflog.Warn(ctx, "Read.CompareruleItemFromPlanAndruleListFromPlan", map[string]interface{}{
					"ruleItemFromPlan.SectionName": ruleItemFromPlan.SectionName,
					"sectionNameItem":              sectionNameItem,
				})
				if ruleItemFromPlan.SectionName == sectionNameItem {
					// section name -> rule index order -> rule name
					mapRuleIndexToRuleName[ruleItemFromPlan.IndexInSection] = ruleItemFromPlan.RuleName
					mapRuleIndexToSectionName[ruleItemFromPlan.IndexInSection] = ruleItemFromPlan.SectionName
					tflog.Warn(ctx, "Read.mapRuleIndexToRuleName.response", map[string]interface{}{
						"ruleItemFromPlan.IndexInSection":   ruleItemFromPlan.IndexInSection,
						"ruleItemFromPlan.RuleName":         ruleItemFromPlan.RuleName,
						"mapInternalRuleIndexToSectionName": utils.InterfaceToJSONString(mapRuleIndexToRuleName),
					})
				}
			}

			tflog.Warn(ctx, "Read.mapRuleIndexToSectionName.response", map[string]interface{}{
				"mapExternalRuleIndexToSectionName": utils.InterfaceToJSONString(mapRuleIndexToRuleName),
			})

			currentRuleId := ""
			for x := 1; x < len(mapRuleIndexToRuleName)+1; x++ {
				toPosition := &cato_models.PolicyRulePositionInput{}
				if x == 1 {
					pos := "FIRST_IN_SECTION"
					toPosition.Position = (*cato_models.PolicyRulePositionEnum)(&pos)
					firstSectionId := sectionIdList[mapRuleIndexToSectionName[1]]
					toPosition.Ref = &firstSectionId
				} else {
					pos := "AFTER_RULE"
					toPosition.Position = (*cato_models.PolicyRulePositionEnum)(&pos)
					currentRuleId = ruleNameIdMap[mapRuleIndexToRuleName[int64(x)-1]]
					toPosition.Ref = &currentRuleId
					tflog.Warn(ctx, "Read.sectionIdList[mapRuleIndexToSectionName[1]].response", map[string]interface{}{
						"mapRuleIndexToSectionName":         mapRuleIndexToRuleName,
						"currentRuleId":                     currentRuleId,
						"mapExternalRuleIndexToSectionName": utils.InterfaceToJSONString(mapRuleIndexToRuleName),
						"ruleNameIdMap":                     utils.InterfaceToJSONString(ruleNameIdMap),
					})
				}

				moveRuleConfig := cato_models.PolicyMoveRuleInput{
					ID: ruleNameIdMap[mapRuleIndexToRuleName[int64(x)]],
					To: toPosition,
				}
				ruleMoveApiData, err := r.client.catov2.PolicyWanFirewallMoveRule(ctx, moveRuleConfig, r.client.AccountId)
				tflog.Warn(ctx, "Write.PolicyWanFirewallMoveRule.response", map[string]interface{}{
					"ruleNameIdMap":             utils.InterfaceToJSONString(ruleNameIdMap),
					"mapRuleIndexToSectionName": utils.InterfaceToJSONString(mapRuleIndexToRuleName),
					"moveRuleConfig":            utils.InterfaceToJSONString(moveRuleConfig),
					"response":                  utils.InterfaceToJSONString(ruleMoveApiData),
				})
				if err != nil {
					diags = append(diags, diag.NewErrorDiagnostic(
						"Catov2 API EntityLookup error",
						err.Error(),
					))
					return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
				}
			}
		}

	}

	_, err = r.client.catov2.PolicyWanFirewallPublishPolicyRevision(ctx, &cato_models.PolicyPublishRevisionInput{}, r.client.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		))
		return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
	}

	// Always set SectionToStartAfterId in the state (it's a computed field)
	// plan.SectionToStartAfterId = types.StringValue(firstSectionId)

	sectionObjectsList, diagsList := types.ListValue(
		types.ObjectType{
			AttrTypes: WanSectionIndexResourceAttrTypes,
		},
		sectionObjects,
	)
	diags = append(diags, diagsList...)

	// After moving sections and rules, get the current state from API to build rule objects
	var ruleObjects []attr.Value
	currentRuleIndexApiData, err := r.client.catov2.PolicyWanFirewallRulesIndex(ctx, r.client.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyWanFirewallRulesIndex error",
			err.Error(),
		))
		return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
	}

	// Create index count per section
	sectionIndexListCount := make(map[string]int64)
	for _, v := range currentRuleIndexApiData.Policy.WanFirewall.Policy.Rules {
		sectionIndexListCount[v.Rule.Section.Name] = 0
	}

	// Build rule objects with current data from API
	for _, v := range currentRuleIndexApiData.Policy.WanFirewall.Policy.Rules {
		sectionIndexListCount[v.Rule.Section.Name]++
		ruleIndexStateData, ruleObjDiags := types.ObjectValue(
			WanRuleIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":               types.StringValue(v.Rule.ID),
				"index_in_section": types.Int64Value(sectionIndexListCount[v.Rule.Section.Name]),
				"section_name":     types.StringValue(v.Rule.Section.Name),
				"rule_name":        types.StringValue(v.Rule.Name),
				"description":      types.StringValue(v.Rule.Description),
				"enabled":          types.BoolValue(v.Rule.Enabled),
			},
		)
		diags = append(diags, ruleObjDiags...)
		ruleObjects = append(ruleObjects, ruleIndexStateData)
	}

	ruleObjectsList, ruleListDiags := types.ListValue(
		types.ObjectType{
			AttrTypes: WanRuleIndexResourceAttrTypes,
		},
		ruleObjects,
	)
	diags = append(diags, ruleListDiags...)

	return sectionObjectsList, ruleObjectsList, diags, nil
}
