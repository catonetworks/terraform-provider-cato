package tlsindex

import (
	"context"
	"errors"
	"sort"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"
	cast "github.com/spf13/cast"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource              = &tlsRulesIndexResource{}
	_ resource.ResourceWithConfigure = &tlsRulesIndexResource{}
)

func NewTLSRulesIndexResource() resource.Resource {
	return &tlsRulesIndexResource{}
}

type tlsRulesIndexResource struct {
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *tlsRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_tls_move_rule"
}

func (r *tlsRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *tlsRulesIndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RulesIndex
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionObjectsList, rulesObjectsList, diags, err := r.moveTLSRulesAndSections(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTLSInspectMoveSection error",
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

func (r *tlsRulesIndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RulesIndex
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

func (r *tlsRulesIndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RulesIndex
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sectionObjectsList, rulesObjectsList, diags, err := r.moveTLSRulesAndSections(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTLSInspectMoveSection error",
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

func (r *tlsRulesIndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RulesIndex
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

//nolint:gocyclo,funlen
func (r *tlsRulesIndexResource) moveTLSRulesAndSections(
	ctx context.Context,
	plan RulesIndex,
) (sectionObjects basetypes.MapValue, ruleObjects basetypes.MapValue, diagnostics diag.Diagnostics, err error) {
	diags := []diag.Diagnostic{}
	ruleObjectMap := make(map[string]attr.Value)

	if plan.SectionToStartAfterID.ValueString() != "" {
		result, err := r.resourceAPI().Tlsinspectpolicy(ctx, r.client.AccountId)
		tflog.Debug(ctx, "Read.TlsinspectpolicySectionsIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic("Catov2 API Tlsinspectpolicy error", err.Error()))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}
		isPresent := false
		for _, item := range result.Policy.TLSInspect.Policy.Sections {
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

	// as the name indicates, a slice of string containing TLS sections names
	listOfSectionNames := make([]string, 0)

	// maps section_name -> section_id
	sectionIDList := make(map[string]string)
	sectionIndexAPIData, err := r.resourceAPI().Tlsinspectpolicy(ctx, r.client.AccountId)
	tflog.Warn(ctx, "Read.TlsinspectpolicySectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexAPIData),
	})
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API Tlsinspectpolicy error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	// for easier processing, a map of section name to ID is created
	sectionListFromPlan := convertMoveTLSRulesAndSectionsValues(ctx, plan, &diags, sectionIDList, sectionIndexAPIData)

	// Sort sections by SectionIndex to ensure proper ordering
	sort.Slice(sectionListFromPlan, func(i, j int) bool {
		return sectionListFromPlan[i].SectionIndex < sectionListFromPlan[j].SectionIndex
	})

	tflog.Debug(ctx, "Processing sections sorted by section_index", map[string]interface{}{
		"sections": utils.InterfaceToJSONString(sectionListFromPlan),
	})

	currentSectionID := ""

	sectionObjectMap := make(map[string]attr.Value)

	// create the sections from the list provided following the section ID provided in firstSectionID
	for _, workingSectionName := range sectionListFromPlan {
		policyMoveSectionInputInt := convertMoveTLSRulesAndSectionsValues3(
			plan,
			&listOfSectionNames,
			sectionIDList,
			&currentSectionID,
			workingSectionName,
		)
		tflog.Warn(ctx, "Write.policyMoveSectionInputInt.response", map[string]interface{}{
			"sectionToStartAfterId":          plan.SectionToStartAfterID.ValueString(),
			"moveFrom":                       workingSectionName.SectionName,
			"toAfter":                        currentSectionID,
			"sectionIDList":                  sectionIDList,
			"workingSectionName.SectionName": workingSectionName.SectionName,
			"sectionIDList[workingSectionName.SectionName]": sectionIDList[workingSectionName.SectionName],
			"response": utils.InterfaceToJSONString(policyMoveSectionInputInt),
		})
		sectionMoveAPIData, err := r.resourceAPI().PolicyTLSInspectMoveSection(ctx, policyMoveSectionInputInt, r.client.AccountId)
		// Check for API errors safely with nil checks
		if sectionMoveAPIData != nil && sectionMoveAPIData.GetPolicy() != nil &&
			sectionMoveAPIData.GetPolicy().TLSInspect != nil &&
			sectionMoveAPIData.GetPolicy().TLSInspect.GetMoveSection() != nil &&
			len(sectionMoveAPIData.GetPolicy().TLSInspect.GetMoveSection().Errors) != 0 {
			tflog.Warn(ctx, "Write.PolicyTLSInspectMoveSectionMoveSection.response", map[string]interface{}{
				"response": utils.InterfaceToJSONString(sectionMoveAPIData),
			})
			if err != nil {
				diags = append(diags, diag.NewErrorDiagnostic(
					"Catov2 API PolicyTLSInspectMoveSection error",
					err.Error(),
				))
				return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
			}
		}
		tflog.Warn(ctx, "Write.PolicyTLSInspectMoveSection.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(sectionMoveAPIData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyTLSInspectMoveSection error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}

		convertMoveTLSRulesAndSectionsValues4(&diags, sectionIDList, &currentSectionID, sectionObjectMap, workingSectionName)
	}

	// now that the sections are ordered properly, move the rules to the correct locations
	if len(plan.RuleData.Elements()) > 0 {
		// get all of the list elements from the plan
		ruleListFromPlan := make([]RulesRuleDataIndex, 0)

		// Convert map to slice for processing
		ruleDataMapElements := plan.RuleData.Elements()
		for _, ruleValue := range ruleDataMapElements {
			rulenDataTmp := convertMoveTLSRulesAndSectionsValues5(ctx, &diags, &ruleListFromPlan, ruleValue)
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

		ruleNameIDData, err := r.resourceAPI().Tlsinspectpolicy(ctx, r.client.AccountId)
		tflog.Warn(ctx, "Read.TlsinspectpolicyRulesIndex.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(ruleNameIDData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API Tlsinspectpolicy error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}
		ruleNameIDMap := make(map[string]string)
		ruleNameDescriptionMap := make(map[string]string)
		ruleNameEnabledMap := make(map[string]bool)

		// create map of TLS rule names from the API to their IDs for easy lookup
		for _, ruleNameIDDataItem := range ruleNameIDData.Policy.TLSInspect.Policy.Rules {
			ruleNameIDMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.ID
			ruleNameDescriptionMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.Description
			ruleNameEnabledMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.Enabled
		}

		tflog.Warn(ctx, "Read.ruleNameIDMap.response", map[string]interface{}{
			"ruleNameIDMap": utils.InterfaceToJSONString(ruleNameIDMap),
		})

		// loop through the ordered list of section names
		for _, sectionNameItem := range listOfSectionNames {
			tflog.Warn(ctx, "Read.ProcessingSectionFromList.response", map[string]interface{}{
				"sectionNameItem": sectionNameItem,
				"ruleNameIDMap":   utils.InterfaceToJSONString(listOfSectionNames),
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

			currentRuleID := ""
			for x := 1; x < len(mapRuleIndexToRuleName)+1; x++ {
				toPosition := &cato_models.PolicyRulePositionInput{}
				if x == 1 {
					pos := cato_models.PolicyRulePositionEnumFirstInSection
					toPosition.Position = &pos
					firstSectionID := sectionIDList[mapRuleIndexToSectionName[1]]
					toPosition.Ref = &firstSectionID
				} else {
					pos := cato_models.PolicyRulePositionEnumAfterRule
					toPosition.Position = &pos
					currentRuleID = ruleNameIDMap[mapRuleIndexToRuleName[int64(x)-1]]
					toPosition.Ref = &currentRuleID
					tflog.Warn(ctx, "Read.sectionIDList[mapRuleIndexToSectionName[1]].response", map[string]interface{}{
						"mapRuleIndexToSectionName":         mapRuleIndexToRuleName,
						"currentRuleID":                     currentRuleID,
						"mapExternalRuleIndexToSectionName": utils.InterfaceToJSONString(mapRuleIndexToRuleName),
						"ruleNameIDMap":                     utils.InterfaceToJSONString(ruleNameIDMap),
					})
				}

				moveRuleConfig := convertMoveTLSRulesAndSectionsValues7(ruleNameIDMap, mapRuleIndexToRuleName, x, toPosition)
				ruleMoveAPIData, err := r.resourceAPI().PolicyTLSInspectMoveRule(ctx, moveRuleConfig, r.client.AccountId)
				tflog.Warn(ctx, "Write.PolicyTLSInspectMoveRule.response", map[string]interface{}{
					"ruleNameIDMap":             utils.InterfaceToJSONString(ruleNameIDMap),
					"mapRuleIndexToSectionName": utils.InterfaceToJSONString(mapRuleIndexToRuleName),
					"moveRuleConfig":            utils.InterfaceToJSONString(moveRuleConfig),
					"response":                  utils.InterfaceToJSONString(ruleMoveAPIData),
				})
				if err != nil {
					diags = append(diags, diag.NewErrorDiagnostic(
						"Catov2 API PolicyTLSInspectMoveRule error",
						err.Error(),
					))
					return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
				}
			}
		}

		// Build final state using API values for computed fields so they are always known post-apply.
		for _, ruleFromPlan := range ruleListFromPlan {
			ruleIndexStateData, diagsSection := buildTLSRuleIndexStateData(
				ruleFromPlan,
				ruleNameIDMap,
				ruleNameDescriptionMap,
				ruleNameEnabledMap,
			)
			diags = append(diags, diagsSection...)
			ruleObjectMap[ruleFromPlan.RuleName] = ruleIndexStateData
		}
	}

	_, err = r.resourceAPI().PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyTLSInspectPublishPolicyRevision error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	sectionObjectsMap, ruleObjectsMap := convertMoveTLSRulesAndSectionsValues2(&diags, ruleObjectMap, sectionObjectMap)

	return sectionObjectsMap, ruleObjectsMap, diags, nil
}
