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

//nolint:gocyclo,funlen
func (r *wanRulesIndexResource) moveWanRulesAndSections(
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
		sectionMoveAPIData, err := r.wanBulkPolicy().PolicyWanFirewallMoveSection(ctx, policyMoveSectionInputInt, r.client.AccountId)
		// Check for API errors safely with nil checks
		if sectionMoveAPIData != nil && sectionMoveAPIData.GetPolicy() != nil &&
			sectionMoveAPIData.GetPolicy().WanFirewall != nil &&
			sectionMoveAPIData.GetPolicy().WanFirewall.GetMoveSection() != nil &&
			len(sectionMoveAPIData.GetPolicy().WanFirewall.GetMoveSection().Errors) != 0 {
			tflog.Warn(ctx, "Write.PolicyWanFirewallMoveSectionMoveSection.response", map[string]interface{}{
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
		tflog.Warn(ctx, "Write.PolicyWanFirewallMoveSection.response", map[string]interface{}{
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
			WanSectionIndexResourceAttrTypes,
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

		reorderOut, reorderCallErr := r.wanBulkPolicy().PolicyWanFirewallReorderPolicy(
			ctx,
			&cato_models.WanFirewallPolicyMutationInput{},
			reorderIn,
			r.client.AccountId,
		)
		if reorderErr := wanFirewallReorderError(reorderOut, reorderCallErr); reorderErr != nil {
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
			)
			diags = append(diags, diagsSection...)
			ruleObjectMap[ruleFromPlan.RuleName] = ruleIndexStateData
		}
	}

	_, err = r.wanBulkPolicy().PolicyWanFirewallPublishPolicyRevision(ctx, &cato_models.PolicyPublishRevisionInput{}, r.client.AccountId)
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
