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
	_ resource.Resource              = &ifwRulesIndexResource{}
	_ resource.ResourceWithConfigure = &ifwRulesIndexResource{}
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
							Computed:    true,
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

var IfwRuleIndexResourceObjectTypes = types.ObjectType{AttrTypes: IfwRuleIndexResourceAttrTypes}
var IfwRuleIndexResourceAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"index_in_section": types.Int64Type,
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

func (r *ifwRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IfwRulesIndex
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sectionObjects []attr.Value
	var ruleObjects []attr.Value

	// Get current sections from API
	sectionIndexApiData, err := r.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallSectionsIndex error",
			err.Error(),
		)
		return
	}

	sectionCount := int64(1)
	sectionIndexListCount := make(map[string]int64)

	// Build section data from API response
	for i, v := range sectionIndexApiData.Policy.InternetFirewall.Policy.Sections {
		// Ignore the first section which includes P2P and system rules
		if i != 0 {
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

	// Get current rules from API
	ruleIndexApiData, err := r.client.catov2.PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallRulesIndex error",
			err.Error(),
		)
		return
	}

	// Build rule data from API response with current indexes
	for i, v := range ruleIndexApiData.Policy.InternetFirewall.Policy.Rules {
		// Ignore the first section which includes P2P and system rules
		if i != 0 {

			sectionIndexListCount[v.Rule.Section.Name]++
			ruleIndexStateData, diags := types.ObjectValue(
				IfwRuleIndexResourceAttrTypes,
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

func (r *ifwRulesIndexResource) moveIfwRulesAndSections(ctx context.Context, plan IfwRulesIndex) (basetypes.ListValue, basetypes.ListValue, diag.Diagnostics, error) {
	diags := []diag.Diagnostic{}

	if len(plan.SectionToStartAfterId.ValueString()) > 0 {
		result, err := r.client.catov2.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
		tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic("Catov2 API PolicyInternetFirewallSectionsIndex error", err.Error()))
			return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
		}
		isPresent := false
		for _, item := range result.Policy.InternetFirewall.Policy.Sections {
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
			return basetypes.ListValue{}, basetypes.ListValue{}, diags, errors.New("SectionToStartAfterId not found")
		}
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
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API EntityLookup error",
			err.Error(),
		))
		return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
	}

	for _, v := range sectionIndexApiData.Policy.InternetFirewall.Policy.Sections {
		sectionIdList[v.Section.Name] = v.Section.ID
	}

	// Check if no sections exist and create a default section for P2P and system rules
	if len(sectionIndexApiData.Policy.InternetFirewall.Policy.Sections) == 0 {
		input := cato_models.PolicyAddSectionInput{
			At: &cato_models.PolicySectionPositionInput{
				Position: cato_models.PolicySectionPositionEnumLastInPolicy,
			},
			Section: &cato_models.PolicyAddSectionInfoInput{
				Name: "Default Outbound Internet",
			},
		}
		sectionCreateApiData, err := r.client.catov2.PolicyInternetFirewallAddSection(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, input, r.client.AccountId)
		tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSectionWithinBulkMove.response", map[string]interface{}{
			"reason":   "creating new section as IFW does not have a default listed",
			"response": utils.InterfaceToJSONString(sectionCreateApiData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API EntityLookup error",
				err.Error(),
			))
			return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
		}
	}

	sectionListFromPlan := make([]IfwRulesSectionDataIndex, 0)

	sectionDataFromPlanList := make([]types.Object, 0, len(plan.SectionData.Elements()))
	diags = append(diags, plan.SectionData.ElementsAs(ctx, &sectionDataFromPlanList, false)...)
	var sectionSourceRuleIndex IfwRulesSectionItemIndex
	for _, item := range sectionDataFromPlanList {
		diags = append(diags, item.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionDataTmp := IfwRulesSectionDataIndex{
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
		sectionMoveApiData, err := r.client.catov2.PolicyInternetFirewallMoveSection(ctx, nil, policyMoveSectionInputInt, r.client.AccountId)
		// Check for API errors safely with nil checks
		if sectionMoveApiData != nil && sectionMoveApiData.GetPolicy() != nil &&
			sectionMoveApiData.GetPolicy().InternetFirewall != nil &&
			sectionMoveApiData.GetPolicy().InternetFirewall.GetMoveSection() != nil &&
			len(sectionMoveApiData.GetPolicy().InternetFirewall.GetMoveSection().Errors) != 0 {
			tflog.Warn(ctx, "Write.PolicyInternetFirewallMoveSectionMoveSection.response", map[string]interface{}{
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
		tflog.Warn(ctx, "Write.PolicyInternetFirewallMoveSection.response", map[string]interface{}{
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
			IfwSectionIndexResourceAttrTypes,
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

	var ruleObjects []attr.Value

	// now that the sections are ordered properly, move the rules to the correct locations
	if len(plan.RuleData.Elements()) > 0 {
		// get all of the list elements from the plan
		ruleListFromPlan := make([]IfwRulesRuleDataIndex, 0)

		ruleDataFromPlanList := make([]types.Object, 0, len(plan.RuleData.Elements()))
		diags = append(diags, plan.RuleData.ElementsAs(ctx, &ruleDataFromPlanList, false)...)

		for _, item := range ruleDataFromPlanList {
			var planSourceRuleIndex IfwRulesRuleItemIndex
			diags = append(diags, item.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

			rulenDataTmp := IfwRulesRuleDataIndex{
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

		tflog.Debug(ctx, "Read.ruleDataFromPlanList.response", map[string]interface{}{
			"ruleListFromPlan": utils.InterfaceToJSONString(ruleListFromPlan),
		})

		ruleNameIdData, err := r.client.catov2.PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
		tflog.Debug(ctx, "Read.PolicyInternetFirewallRulesIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(ruleNameIdData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyInternetFirewallRulesIndex error",
				err.Error(),
			))
			return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
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
				ruleMoveApiData, err := r.client.catov2.PolicyInternetFirewallMoveRule(ctx, nil, moveRuleConfig, r.client.AccountId)
				tflog.Warn(ctx, "Write.PolicyInternetFirewallMoveRule.response", map[string]interface{}{
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

		// Now create the rule objects with proper IDs from the API
		for _, ruleFromPlan := range ruleListFromPlan {
			ruleId := ruleNameIdMap[ruleFromPlan.RuleName]
			ruleIndexStateData, diagsSection := types.ObjectValue(
				IfwRuleIndexResourceAttrTypes,
				map[string]attr.Value{
					"id":               types.StringValue(ruleId),
					"index_in_section": types.Int64Value(ruleFromPlan.IndexInSection),
					"section_name":     types.StringValue(ruleFromPlan.SectionName),
					"rule_name":        types.StringValue(ruleFromPlan.RuleName),
					"description":      types.StringValue(ruleFromPlan.Description),
					"enabled":          types.BoolValue(ruleFromPlan.Enabled),
				},
			)
			diags = append(diags, diagsSection...)
			ruleObjects = append(ruleObjects, ruleIndexStateData)
		}
	}

	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, &cato_models.PolicyPublishRevisionInput{}, r.client.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		))
		return basetypes.ListValue{}, basetypes.ListValue{}, diags, err
	}

	sectionObjectsList, listDiags := types.ListValue(
		types.ObjectType{
			AttrTypes: IfwSectionIndexResourceAttrTypes,
		},
		sectionObjects,
	)
	diags = append(diags, listDiags...)

	ruleObjectsList, listDiags := types.ListValue(
		types.ObjectType{
			AttrTypes: IfwRuleIndexResourceAttrTypes,
		},
		ruleObjects,
	)
	diags = append(diags, listDiags...)

	return sectionObjectsList, ruleObjectsList, diags, nil

}
