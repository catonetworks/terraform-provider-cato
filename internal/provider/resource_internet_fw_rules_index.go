package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &ifwRulesIndexResource{}
	_ resource.ResourceWithConfigure = &ifwRulesIndexResource{}
	// _ resource.ResourceWithImportState = &ifwRulesIndexResource{}
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
				Computed:    true,
			},
			"rule_data": schema.ListNestedAttribute{
				Description: "List of IF Rule Policy Indexes",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "IFW rule id",
							Required:    false,
							Optional:    true,
						},
						"index_in_section": schema.Int64Attribute{
							Description: "Index value remapped per section",
							Required:    false,
							Optional:    true,
							Computed:    true,
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

var IfwRuleIndexResourceObjectTypes = types.ObjectType{AttrTypes: IfwRuleIndexResourceAttrTypes}
var IfwRuleIndexResourceAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"index_in_section": types.Int64Type,
	"section_name":     types.StringType,
	"rule_name":        types.StringType,
}

var IfwSectionIndexResourceObjectTypes = types.ObjectType{AttrTypes: IfwSectionIndexResourceAttrTypes}
var IfwSectionIndexResourceAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"section_name":  types.StringType,
	"section_index": types.Int64Type,
}

func (r *ifwRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	listOfSectionNames := make([]string, 0)

	// maps section_name -> section_id
	// initially used to find ID of Default section
	sectionIdList := make(map[string]string)
	sectionIndexApiData, err := r.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexApiData),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	for _, v := range sectionIndexApiData.Policy.InternetFirewall.Policy.Sections {
		sectionIdList[v.Section.Name] = v.Section.ID
	}

	// first section is defined by the P2P rule section
	var firstSectionId string
	if len(plan.SectionToStartAfterId.ValueString()) > 0 {
		firstSectionId = plan.SectionToStartAfterId.ValueString()
	} else if len(sectionIndexApiData.Policy.InternetFirewall.Policy.Sections) == 0 {
		input := cato_models.PolicyAddSectionInput{
			At: &cato_models.PolicySectionPositionInput{
				Position: cato_models.PolicySectionPositionEnumLastInPolicy,
			},
			Section: &cato_models.PolicyAddSectionInfoInput{
				Name: "Default Outbound Internet",
			}}
		sectionCreateApiData, err := r.client.catov2.PolicyInternetFirewallAddSection(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, input, r.client.AccountId)
		tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSectionWithinBulkMove.response", map[string]interface{}{
			"reason":   "creating new section as IFW does not have a default listed",
			"response": utils.InterfaceToJSONString(sectionCreateApiData),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API EntityLookup error",
				err.Error(),
			)
			return
		}
		firstSectionId = sectionCreateApiData.GetPolicy().GetInternetFirewall().GetAddSection().GetSection().Section.ID
	} else {
		firstSectionId = sectionIndexApiData.Policy.InternetFirewall.Policy.Sections[0].Section.ID
	}

	tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSectionWithinBulkMove.response", map[string]interface{}{
		"sectionIdList": utils.InterfaceToJSONString(sectionIdList),
	})

	plan.SectionToStartAfterId = types.StringValue(firstSectionId)

	sectionListFromPlan := make([]IfwRulesSectionDataIndex, 0)

	sectionDataFromPlanList := make([]types.Object, 0, len(plan.SectionData.Elements()))
	diags = append(diags, plan.SectionData.ElementsAs(ctx, &sectionDataFromPlanList, false)...)
	resp.Diagnostics.Append(diags...)
	var sectionSourceRuleIndex IfwRulesSectionItemIndex
	for _, item := range sectionDataFromPlanList {
		diags = append(diags, item.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)
		resp.Diagnostics.Append(diags...)

		sectionDataTmp := IfwRulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		}
		sectionListFromPlan = append(sectionListFromPlan, sectionDataTmp)

	}

	currentSectionId := firstSectionId

	var sectionObjects []attr.Value

	// create the sections from the list provided following the section ID provided in firstSectionId
	for _, workingSectionName := range sectionListFromPlan {
		listOfSectionNames = append(listOfSectionNames, workingSectionName.SectionName)
		policyMoveSectionInputInt := cato_models.PolicyMoveSectionInput{
			ID: sectionIdList[workingSectionName.SectionName],
			To: &cato_models.PolicySectionPositionInput{
				Ref:      &currentSectionId,
				Position: "AFTER_SECTION",
			},
		}
		tflog.Debug(ctx, "Write.policyMoveSectionInputInt.response", map[string]interface{}{
			"moveFrom": workingSectionName.SectionName,
			"toAfter":  currentSectionId,
			"response": utils.InterfaceToJSONString(policyMoveSectionInputInt),
		})
		sectionMoveApiData, err := r.client.catov2.PolicyInternetFirewallMoveSection(ctx, nil, policyMoveSectionInputInt, r.client.AccountId)
		if len(sectionMoveApiData.GetPolicy().InternetFirewall.GetMoveSection().Errors) != 0 {
			tflog.Debug(ctx, "Write.PolicyInternetFirewallMoveSectionError.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(sectionMoveApiData),
			})
			resp.Diagnostics.AddError(
				"Catov2 API EntityLookup error",
				err.Error(),
			)
			return

		}
		tflog.Debug(ctx, "Write.PolicyInternetFirewallMoveSection.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(sectionMoveApiData),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API EntityLookup error",
				err.Error(),
			)
			return
		}

		sectionIndexStateData, diags := types.ObjectValue(
			IfwSectionIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":            types.StringValue(sectionIdList[workingSectionName.SectionName]),
				"section_name":  types.StringValue(workingSectionName.SectionName),
				"section_index": types.Int64Value(workingSectionName.SectionIndex),
			},
		)
		resp.Diagnostics.Append(diags...)

		sectionObjects = append(sectionObjects, sectionIndexStateData)

		currentSectionId = sectionIdList[workingSectionName.SectionName]
	}

	// now that the sections are ordered properly, move the rules to the correct locations
	if len(plan.RuleData.Elements()) > 0 {
		// get all of the list elements from the plan
		ruleListFromPlan := make([]IfwRulesRuleDataIndex, 0)

		ruleDataFromPlanList := make([]types.Object, 0, len(plan.RuleData.Elements()))
		diags = plan.RuleData.ElementsAs(ctx, &ruleDataFromPlanList, false)
		resp.Diagnostics.Append(diags...)

		for _, item := range ruleDataFromPlanList {
			var planSourceRuleIndex IfwRulesRuleItemIndex
			diags = append(diags, item.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)
			resp.Diagnostics.Append(diags...)

			rulenDataTmp := IfwRulesRuleDataIndex{
				IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
				RuleName:       planSourceRuleIndex.RuleName.ValueString(),
				SectionName:    planSourceRuleIndex.SectionName.ValueString(),
			}
			ruleListFromPlan = append(ruleListFromPlan, rulenDataTmp)
		}

		tflog.Debug(ctx, "Read.ruleDataFromPlanList.response", map[string]interface{}{
			"ruleListFromPlan": utils.InterfaceToJSONString(ruleListFromPlan),
		})

		ruleNameIdData, err := r.client.catov2.PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
		tflog.Debug(ctx, "Read.PolicyInternetFirewallRulesIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(ruleNameIdData),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API PolicyInternetFirewallRulesIndex error",
				err.Error(),
			)
			return
		}
		ruleNameIdMap := make(map[string]string)

		// create map of IFW rule names from the API to their IDs for easy lookup
		for _, ruleNameIdDataItem := range ruleNameIdData.Policy.InternetFirewall.Policy.Rules {
			ruleNameIdMap[ruleNameIdDataItem.Rule.Name] = ruleNameIdDataItem.Rule.ID
		}

		tflog.Debug(ctx, "Write.PolicyInternetFirewallRulesIndexMapNamesIDs.response", map[string]interface{}{
			"ruleNameIdMap": utils.InterfaceToJSONString(ruleNameIdMap),
		})

		// loop through the ordered list of section names
		for _, sectionNameItem := range listOfSectionNames {
			mapRuleIndexToSectionName := make(map[int64]string)
			for _, ruleItemFromPlan := range ruleListFromPlan {
				if ruleItemFromPlan.SectionName == sectionNameItem {
					// section name -> rule index order -> rule name
					mapRuleIndexToSectionName[ruleItemFromPlan.IndexInSection] = ruleItemFromPlan.RuleName
				}
			}
			currentRuleId := ""

			for x := 1; x < len(mapRuleIndexToSectionName)+1; x++ {
				toPosition := &cato_models.PolicyRulePositionInput{}
				if x == 1 {
					pos := "FIRST_IN_SECTION"
					toPosition.Position = (*cato_models.PolicyRulePositionEnum)(&pos)
					secNameInt := sectionIdList[sectionNameItem]
					toPosition.Ref = &secNameInt
				} else {
					pos := "AFTER_RULE"
					toPosition.Position = (*cato_models.PolicyRulePositionEnum)(&pos)
					currentRuleId = ruleNameIdMap[mapRuleIndexToSectionName[int64(x)-1]]
					toPosition.Ref = &currentRuleId
				}

				moveRuleConfig := cato_models.PolicyMoveRuleInput{
					ID: ruleNameIdMap[mapRuleIndexToSectionName[int64(x)]],
					To: toPosition,
				}

				ruleMoveApiData, err := r.client.catov2.PolicyInternetFirewallMoveRule(ctx, nil, moveRuleConfig, r.client.AccountId)
				tflog.Debug(ctx, "Write.PolicyInternetFirewallMoveRule.response", map[string]interface{}{
					"moveRuleConfig": utils.InterfaceToJSONString(moveRuleConfig),
					"response":       utils.InterfaceToJSONString(ruleMoveApiData),
				})
				if err != nil {
					resp.Diagnostics.AddError(
						"Catov2 API PolicyInternetFirewallMoveRule error",
						err.Error(),
					)
					return
				}
			}
			tflog.Debug(ctx, "Write.PolicyInternetFirewallRulesIndexMapSectionsNamesIDs.response", map[string]interface{}{
				"sectionNameItem":           sectionNameItem,
				"numberOfMoveOperations":    len(mapRuleIndexToSectionName),
				"mapRuleIndexToSectionName": utils.InterfaceToJSONString(mapRuleIndexToSectionName),
			})
		}
	}

	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, &cato_models.PolicyPublishRevisionInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	plan.SectionToStartAfterId = types.StringValue(currentSectionId)

	sectionObjectsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: IfwSectionIndexResourceAttrTypes,
		},
		sectionObjects,
	)
	resp.Diagnostics.Append(diags...)
	plan.SectionData = sectionObjectsList

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ifwRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IfwRulesIndex
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sectionObjects []attr.Value
	var ruleObjects []attr.Value

	sectionIndexApiData, err := r.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndexInRead.response", map[string]interface{}{
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
	for k, v := range sectionIndexApiData.Policy.InternetFirewall.Policy.Sections {
		if k != 0 {

			sectionIndexListCount[v.Section.Name] = 0
			sectionIndexStateData, diags := types.ObjectValue(
				IfwSectionIndexResourceAttrTypes,
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
	}

	sectionObjectsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: IfwSectionIndexResourceAttrTypes,
		},
		sectionObjects,
	)
	resp.Diagnostics.Append(diags...)
	state.SectionData = sectionObjectsList

	ruleIndexApiData, err := r.client.catov2.PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallRulesIndex.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ruleIndexApiData),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	for _, v := range ruleIndexApiData.Policy.InternetFirewall.Policy.Rules {
		sectionIndexListCount[v.Rule.Section.Name]++
		ruleIndexStateData, diags := types.ObjectValue(
			IfwRuleIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":               types.StringValue(v.Rule.ID),
				"index_in_section": types.Int64Value(sectionIndexListCount[v.Rule.Section.Name]),
				"section_name":     types.StringValue(v.Rule.Section.Name),
				"rule_name":        types.StringValue(v.Rule.Name),
			},
		)
		resp.Diagnostics.Append(diags...)
		ruleObjects = append(ruleObjects, ruleIndexStateData)
	}

	ruleObjectsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: IfwRuleIndexResourceAttrTypes,
		},
		ruleObjects,
	)
	resp.Diagnostics.Append(diags...)
	state.RuleData = ruleObjectsList

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

	listOfSectionNames := make([]string, 0)

	// maps section_name -> section_id
	// initially used to find ID of Default section
	sectionIdList := make(map[string]string)
	sectionIndexApiData, err := r.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexApiData),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	for _, v := range sectionIndexApiData.Policy.InternetFirewall.Policy.Sections {
		sectionIdList[v.Section.Name] = v.Section.ID
	}

	// first section is defined by the P2P rule section
	var firstSectionId string
	if len(plan.SectionToStartAfterId.ValueString()) > 0 {
		firstSectionId = plan.SectionToStartAfterId.ValueString()
	} else if len(sectionIndexApiData.Policy.InternetFirewall.Policy.Sections) == 0 {
		input := cato_models.PolicyAddSectionInput{
			At: &cato_models.PolicySectionPositionInput{
				Position: cato_models.PolicySectionPositionEnumLastInPolicy,
			},
			Section: &cato_models.PolicyAddSectionInfoInput{
				Name: "Default Outbound Internet",
			}}
		sectionCreateApiData, err := r.client.catov2.PolicyInternetFirewallAddSection(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, input, r.client.AccountId)
		tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSectionWithinBulkMove.response", map[string]interface{}{
			"reason":   "creating new section as IFW does not have a default listed",
			"response": utils.InterfaceToJSONString(sectionCreateApiData),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API EntityLookup error",
				err.Error(),
			)
			return
		}
		firstSectionId = sectionCreateApiData.GetPolicy().GetInternetFirewall().GetAddSection().GetSection().Section.ID
	} else {
		firstSectionId = sectionIndexApiData.Policy.InternetFirewall.Policy.Sections[0].Section.ID
	}

	tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSectionWithinBulkMove.response", map[string]interface{}{
		"sectionIdList": utils.InterfaceToJSONString(sectionIdList),
	})

	plan.SectionToStartAfterId = types.StringValue(firstSectionId)

	sectionListFromPlan := make([]IfwRulesSectionDataIndex, 0)

	sectionDataFromPlanList := make([]types.Object, 0, len(plan.SectionData.Elements()))
	diags = append(diags, plan.SectionData.ElementsAs(ctx, &sectionDataFromPlanList, false)...)
	resp.Diagnostics.Append(diags...)
	var sectionSourceRuleIndex IfwRulesSectionItemIndex
	for _, item := range sectionDataFromPlanList {
		diags = append(diags, item.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)
		resp.Diagnostics.Append(diags...)

		sectionDataTmp := IfwRulesSectionDataIndex{
			SectionIndex: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			SectionName:  sectionSourceRuleIndex.SectionName.ValueString(),
		}
		sectionListFromPlan = append(sectionListFromPlan, sectionDataTmp)

	}

	currentSectionId := firstSectionId

	var sectionObjects []attr.Value

	// create the sections from the list provided following the section ID provided in firstSectionId
	for _, workingSectionName := range sectionListFromPlan {
		listOfSectionNames = append(listOfSectionNames, workingSectionName.SectionName)
		policyMoveSectionInputInt := cato_models.PolicyMoveSectionInput{
			ID: sectionIdList[workingSectionName.SectionName],
			To: &cato_models.PolicySectionPositionInput{
				Ref:      &currentSectionId,
				Position: "AFTER_SECTION",
			},
		}
		tflog.Debug(ctx, "Write.policyMoveSectionInputInt.response", map[string]interface{}{
			"moveFrom": workingSectionName.SectionName,
			"toAfter":  currentSectionId,
			"response": utils.InterfaceToJSONString(policyMoveSectionInputInt),
		})
		sectionMoveApiData, err := r.client.catov2.PolicyInternetFirewallMoveSection(ctx, nil, policyMoveSectionInputInt, r.client.AccountId)
		if len(sectionMoveApiData.GetPolicy().InternetFirewall.GetMoveSection().Errors) != 0 {
			tflog.Debug(ctx, "Write.PolicyInternetFirewallMoveSectionError.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(sectionMoveApiData),
			})
			resp.Diagnostics.AddError(
				"Catov2 API EntityLookup error",
				err.Error(),
			)
			return

		}
		tflog.Debug(ctx, "Write.PolicyInternetFirewallMoveSection.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(sectionMoveApiData),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API EntityLookup error",
				err.Error(),
			)
			return
		}

		sectionIndexStateData, diags := types.ObjectValue(
			IfwSectionIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":            types.StringValue(sectionIdList[workingSectionName.SectionName]),
				"section_name":  types.StringValue(workingSectionName.SectionName),
				"section_index": types.Int64Value(workingSectionName.SectionIndex),
			},
		)
		resp.Diagnostics.Append(diags...)

		sectionObjects = append(sectionObjects, sectionIndexStateData)

		currentSectionId = sectionIdList[workingSectionName.SectionName]
	}

	// now that the sections are ordered properly, move the rules to the correct locations
	if len(plan.RuleData.Elements()) > 0 {
		// get all of the list elements from the plan
		ruleListFromPlan := make([]IfwRulesRuleDataIndex, 0)

		ruleDataFromPlanList := make([]types.Object, 0, len(plan.RuleData.Elements()))
		diags = plan.RuleData.ElementsAs(ctx, &ruleDataFromPlanList, false)
		resp.Diagnostics.Append(diags...)

		for _, item := range ruleDataFromPlanList {
			var planSourceRuleIndex IfwRulesRuleItemIndex
			diags = append(diags, item.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)
			resp.Diagnostics.Append(diags...)

			rulenDataTmp := IfwRulesRuleDataIndex{
				IndexInSection: planSourceRuleIndex.IndexInSection.ValueInt64(),
				RuleName:       planSourceRuleIndex.RuleName.ValueString(),
				SectionName:    planSourceRuleIndex.SectionName.ValueString(),
			}
			ruleListFromPlan = append(ruleListFromPlan, rulenDataTmp)
		}

		tflog.Debug(ctx, "Read.ruleDataFromPlanList.response", map[string]interface{}{
			"ruleListFromPlan": utils.InterfaceToJSONString(ruleListFromPlan),
		})

		ruleNameIdData, err := r.client.catov2.PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
		tflog.Debug(ctx, "Read.PolicyInternetFirewallRulesIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(ruleNameIdData),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API PolicyInternetFirewallRulesIndex error",
				err.Error(),
			)
			return
		}
		ruleNameIdMap := make(map[string]string)

		// create map of IFW rule names from the API to their IDs for easy lookup
		for _, ruleNameIdDataItem := range ruleNameIdData.Policy.InternetFirewall.Policy.Rules {
			ruleNameIdMap[ruleNameIdDataItem.Rule.Name] = ruleNameIdDataItem.Rule.ID
		}

		tflog.Debug(ctx, "Write.PolicyInternetFirewallRulesIndexMapNamesIDs.response", map[string]interface{}{
			"ruleNameIdMap": utils.InterfaceToJSONString(ruleNameIdMap),
		})

		// loop through the ordered list of section names
		for _, sectionNameItem := range listOfSectionNames {
			mapRuleIndexToSectionName := make(map[int64]string)
			for _, ruleItemFromPlan := range ruleListFromPlan {
				if ruleItemFromPlan.SectionName == sectionNameItem {
					// section name -> rule index order -> rule name
					mapRuleIndexToSectionName[ruleItemFromPlan.IndexInSection] = ruleItemFromPlan.RuleName
				}
			}
			currentRuleId := ""

			for x := 1; x < len(mapRuleIndexToSectionName)+1; x++ {
				toPosition := &cato_models.PolicyRulePositionInput{}
				if x == 1 {
					pos := "FIRST_IN_SECTION"
					toPosition.Position = (*cato_models.PolicyRulePositionEnum)(&pos)
					secNameInt := sectionIdList[sectionNameItem]
					toPosition.Ref = &secNameInt
				} else {
					pos := "AFTER_RULE"
					toPosition.Position = (*cato_models.PolicyRulePositionEnum)(&pos)
					currentRuleId = ruleNameIdMap[mapRuleIndexToSectionName[int64(x)-1]]
					toPosition.Ref = &currentRuleId
				}

				moveRuleConfig := cato_models.PolicyMoveRuleInput{
					ID: ruleNameIdMap[mapRuleIndexToSectionName[int64(x)]],
					To: toPosition,
				}

				ruleMoveApiData, err := r.client.catov2.PolicyInternetFirewallMoveRule(ctx, nil, moveRuleConfig, r.client.AccountId)
				tflog.Debug(ctx, "Write.PolicyInternetFirewallMoveRule.response", map[string]interface{}{
					"moveRuleConfig": utils.InterfaceToJSONString(moveRuleConfig),
					"response":       utils.InterfaceToJSONString(ruleMoveApiData),
				})
				if err != nil {
					resp.Diagnostics.AddError(
						"Catov2 API PolicyInternetFirewallMoveRule error",
						err.Error(),
					)
					return
				}
			}
			tflog.Debug(ctx, "Write.PolicyInternetFirewallRulesIndexMapSectionsNamesIDs.response", map[string]interface{}{
				"sectionNameItem":           sectionNameItem,
				"numberOfMoveOperations":    len(mapRuleIndexToSectionName),
				"mapRuleIndexToSectionName": utils.InterfaceToJSONString(mapRuleIndexToSectionName),
			})
		}
	}

	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, &cato_models.PolicyPublishRevisionInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	plan.SectionToStartAfterId = types.StringValue(currentSectionId)

	sectionObjectsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: IfwSectionIndexResourceAttrTypes,
		},
		sectionObjects,
	)
	resp.Diagnostics.Append(diags...)
	plan.SectionData = sectionObjectsList

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
