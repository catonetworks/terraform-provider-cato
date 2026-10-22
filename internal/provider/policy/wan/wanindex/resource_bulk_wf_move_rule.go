package wanindex

import (
	"context"
	"errors"
	"sort"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"
	cast "github.com/spf13/cast"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	policy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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
	client  *catoclient.Service
	wanBulk WanFirewallBulkPolicyClient // optional override for tests

	apiClient ResourceClient
}

func (r *wanRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_wf_move_rule"
}

func (r *wanRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

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
		result, err := r.resourceAPI().PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
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
	sectionIndexAPIData, err := r.resourceAPI().PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
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
	sectionListFromPlan := convertMoveWanRulesAndSectionsValues(ctx, plan, &diags, sectionIDList, sectionIndexAPIData)

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
		policyMoveSectionInputInt := convertMoveWanRulesAndSectionsValues3(plan, sectionIDList, &currentSectionID, workingSectionName)
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
		moveErr := policy.WithPolicyRevisionConflictRetry(ctx, "PolicyWanFirewallMoveSection", func() error {
			var callErr error
			sectionMoveAPIData, callErr = r.resourceAPI().PolicyWanFirewallMoveSection(
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

		convertMoveWanRulesAndSectionsValues4(&diags, sectionIDList, &currentSectionID, sectionObjectMap, workingSectionName)
	}

	// now that the sections are ordered properly, move the rules to the correct locations
	if len(plan.RuleData.Elements()) > 0 {
		// get all of the list elements from the plan
		ruleListFromPlan := make([]WanRulesRuleDataIndex, 0)

		// Convert map to slice for processing
		ruleDataMapElements := plan.RuleData.Elements()
		for _, ruleValue := range ruleDataMapElements {
			rulenDataTmp := convertMoveWanRulesAndSectionsValues5(ctx, &diags, &ruleListFromPlan, ruleValue)
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

		ruleNameIDData, err := r.resourceAPI().PolicyWanFirewallRulesIndex(ctx, r.client.AccountId)
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

		sectionIdxAfter, err := r.resourceAPI().PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyWanFirewallSectionsIndex error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}

		sections := make([]policy.BulkPolicySectionRef, 0, len(sectionIdxAfter.Policy.WanFirewall.Policy.Sections))
		for _, item := range sectionIdxAfter.Policy.WanFirewall.Policy.Sections {
			sections = append(sections, policy.BulkPolicySectionRef{
				ID:   item.Section.ID,
				Name: item.Section.Name,
			})
		}

		rules := make([]policy.BulkPolicyRuleRow, 0, len(ruleNameIDData.Policy.WanFirewall.Policy.Rules))
		for _, item := range ruleNameIDData.Policy.WanFirewall.Policy.Rules {
			rules = append(rules, policy.BulkPolicyRuleRow{
				SectionID:   item.Rule.Section.ID,
				SectionName: item.Rule.Section.Name,
				RuleID:      item.Rule.ID,
				RuleName:    item.Rule.Name,
				Index:       item.Rule.Index,
				IsSystem: policy.ElementHasProperty(
					item.Properties,
					cato_models.PolicyElementPropertiesEnumSystem,
				),
			})
		}

		planned := make([]policy.BulkPlannedRuleIndex, 0, len(ruleListFromPlan))
		for _, r := range ruleListFromPlan {
			planned = append(planned, policy.BulkPlannedRuleIndex{
				SectionName:    r.SectionName,
				RuleName:       r.RuleName,
				IndexInSection: r.IndexInSection,
			})
		}

		reorderIn, buildErr := policy.BuildPolicyReorderInput(sections, rules, planned)
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
		reorderErr := policy.WithAcctestPolicyRevisionCleanupRetryOnce(ctx, "PolicyWanFirewallReorderPolicy", func() error {
			return r.resourceAPI().DiscardPolicyRevisions(ctx)
		}, func() error {
			var callErr error
			reorderOut, callErr = r.resourceAPI().PolicyWanFirewallReorderPolicy(
				ctx,
				&cato_models.WanFirewallPolicyMutationInput{},
				reorderIn,
				r.client.AccountId,
			)
			return policy.WanFirewallReorderError(reorderOut, callErr)
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
			)
			diags = append(diags, diagsSection...)
			ruleObjectMap[ruleFromPlan.RuleName] = ruleIndexStateData
		}
	}

	pubErr := policy.WithPolicyRevisionConflictRetry(ctx, "PolicyWanFirewallPublishPolicyRevision", func() error {
		_, errPub := r.resourceAPI().PolicyWanFirewallPublishPolicyRevision(
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

	sectionObjectsMap, ruleObjectsMap := convertMoveWanRulesAndSectionsValues2(&diags, ruleObjectMap, sectionObjectMap)

	return sectionObjectsMap, ruleObjectsMap, diags, nil
}
