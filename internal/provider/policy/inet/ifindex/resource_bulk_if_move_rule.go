package ifindex

import (
	"context"
	"errors"
	"fmt"
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
	_ resource.Resource              = &ifwRulesIndexResource{}
	_ resource.ResourceWithConfigure = &ifwRulesIndexResource{}
)

func NewIfwRulesIndexResource() resource.Resource {
	return &ifwRulesIndexResource{}
}

type ifwRulesIndexResource struct {
	client  *catoclient.Service
	ifwBulk InternetFirewallBulkPolicyClient // optional override for tests

	apiClient ResourceClient
}

func (r *ifwRulesIndexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_if_move_rule"
}

func (r *ifwRulesIndexResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

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
			"Catov2 API internet firewall bulk rules index error",
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

	// Create maps to store API data for ID lookups
	sectionIDMap := make(map[string]string)
	ruleIDMap := make(map[string]string)
	ruleDescriptionMap := make(map[string]string)
	ruleEnabledMap := make(map[string]bool)

	// Get current sections from API to get fresh IDs
	sectionIndexAPIData, err := r.resourceAPI().PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
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
	ruleIndexAPIData, err := r.resourceAPI().PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallRulesIndex error",
			err.Error(),
		)
		return
	}

	// Build maps of rule names to API data
	convertReadState(ctx, &state, &diags, sectionIDMap, ruleIDMap, ruleDescriptionMap, ruleEnabledMap, ruleIndexAPIData)

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
			"Catov2 API internet firewall bulk rules index error",
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
		result, err := r.resourceAPI().PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
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
	sectionIndexAPIData, err := r.resourceAPI().PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
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
		sectionCreateAPIData, err := r.resourceAPI().PolicyInternetFirewallAddSection(
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

	var ruleIndexAPIData *cato_go_sdk.IfwRulesIndexPolicy
	rules := make([]policy.BulkPolicyRuleRow, 0)
	protectedSectionIDs := make(map[string]struct{})
	if len(plan.SectionData.Elements()) > 0 || len(plan.RuleData.Elements()) > 0 {
		ruleIndexAPIData, err = r.resourceAPI().PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyInternetFirewallRulesIndex error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}

		rules = make([]policy.BulkPolicyRuleRow, 0, len(ruleIndexAPIData.Policy.InternetFirewall.Policy.Rules))
		for _, item := range ruleIndexAPIData.Policy.InternetFirewall.Policy.Rules {
			isSystem := policy.ElementHasProperty(
				item.Properties,
				cato_models.PolicyElementPropertiesEnumSystem,
			)
			rules = append(rules, policy.BulkPolicyRuleRow{
				SectionID:   item.Rule.Section.ID,
				SectionName: item.Rule.Section.Name,
				RuleID:      item.Rule.ID,
				RuleName:    item.Rule.Name,
				Index:       item.Rule.Index,
				IsSystem:    isSystem,
			})
			if isSystem {
				protectedSectionIDs[item.Rule.Section.ID] = struct{}{}
			}
		}
	}

	sectionListFromPlan := convertMoveIfwRulesAndSectionsValues(ctx, plan, &diags)

	// Sort sections by SectionIndex to ensure proper ordering
	sort.Slice(sectionListFromPlan, func(i, j int) bool {
		return sectionListFromPlan[i].SectionIndex < sectionListFromPlan[j].SectionIndex
	})

	tflog.Debug(ctx, "Processing sections in plan order (already sorted by module)", map[string]interface{}{
		"sections": utils.InterfaceToJSONString(sectionListFromPlan),
	})

	sectionObjectMap := make(map[string]attr.Value)

	// create the sections from the list provided following the section ID provided in firstSectionID
	sectionMoves := buildIfwSectionMoves(
		sectionListFromPlan,
		sectionIDList,
		protectedSectionIDs,
		plan.SectionToStartAfterID.ValueString(),
	)
	for _, move := range sectionMoves {
		workingSectionName, policyMoveSectionInputInt := convertMoveIfwRulesAndSectionsValues4(&diags, sectionIDList, sectionObjectMap, move)

		if move.protected {
			diags = append(diags, diag.NewWarningDiagnostic(
				"Internet firewall section move skipped",
				fmt.Sprintf(
					"Section %q contains a system rule and cannot be moved; its current policy position was preserved.",
					workingSectionName.SectionName,
				),
			))
			continue
		}

		tflog.Warn(ctx, "Write.policyMoveSectionInputInt.response", map[string]interface{}{
			"sectionToStartAfterId":          plan.SectionToStartAfterID.ValueString(),
			"moveFrom":                       workingSectionName.SectionName,
			"toAfter":                        policyMoveSectionInputInt.To.Ref,
			"sectionIDList":                  sectionIDList,
			"workingSectionName.SectionName": workingSectionName.SectionName,
			"sectionIDList[workingSectionName.SectionName]": sectionIDList[workingSectionName.SectionName],
			"response": utils.InterfaceToJSONString(policyMoveSectionInputInt),
		})
		var sectionMoveAPIData *cato_go_sdk.PolicyInternetFirewallMoveSection
		moveErr := policy.WithPolicyRevisionConflictRetry(ctx, "PolicyInternetFirewallMoveSection", func() error {
			var callErr error
			sectionMoveAPIData, callErr = r.resourceAPI().PolicyInternetFirewallMoveSection(
				ctx, nil, policyMoveSectionInputInt, r.client.AccountId)
			return policy.InternetFirewallMoveSectionError(sectionMoveAPIData, callErr)
		})
		if moveErr != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyInternetFirewallMoveSection error",
				moveErr.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, moveErr
		}
		tflog.Warn(ctx, "Write.PolicyInternetFirewallMoveSection.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(sectionMoveAPIData),
		})
	}

	// now that the sections are ordered properly, move the rules to the correct locations
	if len(plan.RuleData.Elements()) > 0 {
		// get all of the list elements from the plan
		ruleListFromPlan, sectionIndexMap := convertMoveIfwRulesAndSectionsValues5(ctx, plan, &diags, sectionListFromPlan)

		// Sort rules by section index first, then by IndexInSection to ensure proper ordering
		sort.Slice(ruleListFromPlan, func(i, j int) bool {
			// First sort by section index (not alphabetical section name), then by index within section
			if ruleListFromPlan[i].SectionName != ruleListFromPlan[j].SectionName {
				return sectionIndexMap[ruleListFromPlan[i].SectionName] < sectionIndexMap[ruleListFromPlan[j].SectionName]
			}
			return ruleListFromPlan[i].IndexInSection < ruleListFromPlan[j].IndexInSection
		})

		ruleNameIDMap := make(map[string]string)
		ruleNameDescriptionMap := make(map[string]string)
		ruleNameEnabledMap := make(map[string]bool)

		// create map of IFW rule names from the API to their IDs for easy lookup
		for _, ruleNameIDDataItem := range ruleIndexAPIData.Policy.InternetFirewall.Policy.Rules {
			ruleNameIDMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.ID
			ruleNameDescriptionMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.Description
			ruleNameEnabledMap[ruleNameIDDataItem.Rule.Name] = ruleNameIDDataItem.Rule.Enabled
		}

		tflog.Debug(ctx, "Processing rules in correct order", map[string]interface{}{
			"ruleListFromPlan": utils.InterfaceToJSONString(ruleListFromPlan),
			"ruleNameIDMap":    utils.InterfaceToJSONString(ruleNameIDMap),
		})

		sectionIdxAfter, err := r.resourceAPI().PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyInternetFirewallSectionsIndex error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}

		sections := make([]policy.BulkPolicySectionRef, 0, len(sectionIdxAfter.Policy.InternetFirewall.Policy.Sections))
		for _, item := range sectionIdxAfter.Policy.InternetFirewall.Policy.Sections {
			sections = append(sections, policy.BulkPolicySectionRef{
				ID:   item.Section.ID,
				Name: item.Section.Name,
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
				"Internet firewall policy reorder",
				buildErr.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, buildErr
		}

		tflog.Debug(ctx, "PolicyInternetFirewallReorderPolicy request", map[string]interface{}{
			"policyReorderInput": utils.InterfaceToJSONString(reorderIn),
		})

		var reorderOut *cato_go_sdk.PolicyInternetFirewallReorderPolicy
		reorderErr := policy.WithPolicyRevisionConflictRetry(ctx, "PolicyInternetFirewallReorderPolicy", func() error {
			var callErr error
			reorderOut, callErr = r.resourceAPI().PolicyInternetFirewallReorderPolicy(
				ctx,
				&cato_models.InternetFirewallPolicyMutationInput{},
				reorderIn,
				r.client.AccountId,
			)
			return policy.InternetFirewallReorderError(reorderOut, callErr)
		})
		if reorderErr != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyInternetFirewallReorderPolicy error",
				reorderErr.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, reorderErr
		}

		tflog.Debug(ctx, "PolicyInternetFirewallReorderPolicy response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(reorderOut),
		})

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
	pubErr := policy.WithPolicyRevisionConflictRetry(ctx, "PolicyInternetFirewallPublishPolicyRevision", func() error {
		_, errPub := r.resourceAPI().PolicyInternetFirewallPublishPolicyRevision(
			ctx,
			&cato_models.InternetFirewallPolicyMutationInput{},
			&cato_models.PolicyPublishRevisionInput{},
			r.client.AccountId,
		)
		return errPub
	})
	if pubErr != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			pubErr.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, pubErr
	}

	sectionObjectsMap, ruleObjectsMap := convertMoveIfwRulesAndSectionsValues2(&diags, ruleObjectMap, sectionObjectMap)

	return sectionObjectsMap, ruleObjectsMap, diags, nil
}
