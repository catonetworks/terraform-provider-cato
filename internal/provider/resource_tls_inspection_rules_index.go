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
	_ resource.Resource              = &tlsRulesIndexResource{}
	_ resource.ResourceWithConfigure = &tlsRulesIndexResource{}
)

func NewTLSRulesIndexResource() resource.Resource {
	return &tlsRulesIndexResource{}
}

type tlsRulesIndexResource struct {
	client *catoClientData
}

func (r *tlsRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_tls_move_rule"
}

func (r *tlsRulesIndexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves index values for TLS Inspection Rules.",
		Attributes: map[string]schema.Attribute{
			"section_to_start_after_id": schema.StringAttribute{
				Description: "TLS section id",
				Required:    false,
				Optional:    true,
			},
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of TLS Rule Policy Indexes keyed by rule_name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "TLS rule id",
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
							Description: "TLS section name housing rule",
							Required:    false,
							Optional:    true,
						},
						"rule_name": schema.StringAttribute{
							Description: "TLS rule name housing rule",
							Required:    false,
							Optional:    true,
						},
						"description": schema.StringAttribute{
							Description: "TLS rule description",
							Required:    false,
							Optional:    true,
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "TLS rule enabled",
							Required:    false,
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"section_data": schema.MapNestedAttribute{
				Description: "Map of TLS section Indexes keyed by section_name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "TLS section id housing rule",
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
							Description: "TLS section name housing rule",
							Required:    true,
							Optional:    false,
						},
					},
				},
			},
		},
	}
}

var TLSRuleIndexResourceObjectTypes = types.ObjectType{AttrTypes: TLSRuleIndexResourceAttrTypes}
var TLSRuleIndexResourceAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"index_in_section": types.Int64Type,
	"section_name":     types.StringType,
	"rule_name":        types.StringType,
	"description":      types.StringType,
	"enabled":          types.BoolType,
}

var TLSSectionIndexResourceObjectTypes = types.ObjectType{AttrTypes: TLSSectionIndexResourceAttrTypes}
var TLSSectionIndexResourceAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"section_name":  types.StringType,
	"section_index": types.Int64Type,
}

func (r *tlsRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *tlsRulesIndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TLSRulesIndex
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
	var state TLSRulesIndex
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
	var plan TLSRulesIndex
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
	var state TLSRulesIndex
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
	plan TLSRulesIndex,
) (sectionObjects basetypes.MapValue, ruleObjects basetypes.MapValue, diagnostics diag.Diagnostics, err error) {
	diags := []diag.Diagnostic{}
	ruleObjectMap := make(map[string]attr.Value)

	if plan.SectionToStartAfterID.ValueString() != "" {
		result, err := r.client.catov2.Tlsinspectpolicy(ctx, r.client.AccountId)
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
	sectionIndexAPIData, err := r.client.catov2.Tlsinspectpolicy(ctx, r.client.AccountId)
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
	for _, v := range sectionIndexAPIData.Policy.TLSInspect.Policy.Sections {
		sectionIDList[v.Section.Name] = v.Section.ID
	}

	sectionListFromPlan := make([]TLSRulesSectionDataIndex, 0)

	// Convert map to slice for processing
	sectionDataMapElements := plan.SectionData.Elements()
	for _, sectionValue := range sectionDataMapElements {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex TLSRulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionDataTmp := TLSRulesSectionDataIndex{
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

	// create the sections from the list provided following the section ID provided in firstSectionID
	for _, workingSectionName := range sectionListFromPlan {
		listOfSectionNames = append(listOfSectionNames, workingSectionName.SectionName)
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
		sectionMoveAPIData, err := r.client.catov2.PolicyTLSInspectMoveSection(ctx, policyMoveSectionInputInt, r.client.AccountId)
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

		sectionIndexStateData, diagsSection := types.ObjectValue(
			TLSSectionIndexResourceAttrTypes,
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
		ruleListFromPlan := make([]TLSRulesRuleDataIndex, 0)

		// Convert map to slice for processing
		ruleDataMapElements := plan.RuleData.Elements()
		for _, ruleValue := range ruleDataMapElements {
			ruleObject := ruleValue.(types.Object)
			var planSourceRuleIndex TLSRulesRuleItemIndex
			diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

			rulenDataTmp := TLSRulesRuleDataIndex{
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

		ruleNameIDData, err := r.client.catov2.Tlsinspectpolicy(ctx, r.client.AccountId)
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
					pos := ifwRulePositionFirstInSection
					toPosition.Position = (*cato_models.PolicyRulePositionEnum)(&pos)
					firstSectionID := sectionIDList[mapRuleIndexToSectionName[1]]
					toPosition.Ref = &firstSectionID
				} else {
					pos := ifwRulePositionAfterRule
					toPosition.Position = (*cato_models.PolicyRulePositionEnum)(&pos)
					currentRuleID = ruleNameIDMap[mapRuleIndexToRuleName[int64(x)-1]]
					toPosition.Ref = &currentRuleID
					tflog.Warn(ctx, "Read.sectionIDList[mapRuleIndexToSectionName[1]].response", map[string]interface{}{
						"mapRuleIndexToSectionName":         mapRuleIndexToRuleName,
						"currentRuleID":                     currentRuleID,
						"mapExternalRuleIndexToSectionName": utils.InterfaceToJSONString(mapRuleIndexToRuleName),
						"ruleNameIDMap":                     utils.InterfaceToJSONString(ruleNameIDMap),
					})
				}

				moveRuleConfig := cato_models.PolicyMoveRuleInput{
					ID: ruleNameIDMap[mapRuleIndexToRuleName[int64(x)]],
					To: toPosition,
				}
				ruleMoveAPIData, err := r.client.catov2.PolicyTLSInspectMoveRule(ctx, moveRuleConfig, r.client.AccountId)
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

	_, err = r.client.catov2.PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyTLSInspectPublishPolicyRevision error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	sectionObjectsMap, sectionMapDiags := types.MapValue(
		TLSSectionIndexResourceObjectTypes,
		sectionObjectMap,
	)
	diags = append(diags, sectionMapDiags...)

	ruleObjectsMap, ruleMapDiags := types.MapValue(
		TLSRuleIndexResourceObjectTypes,
		ruleObjectMap,
	)
	diags = append(diags, ruleMapDiags...)

	return sectionObjectsMap, ruleObjectsMap, diags, nil
}

func buildTLSRuleIndexStateData(
	ruleFromPlan TLSRulesRuleDataIndex,
	ruleNameIDMap map[string]string,
	ruleNameDescriptionMap map[string]string,
	ruleNameEnabledMap map[string]bool,
) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		TLSRuleIndexResourceAttrTypes,
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
