package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/clientinterfaces"
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
)

var (
	_ resource.Resource              = &ifwRulesIndexResource{}
	_ resource.ResourceWithConfigure = &ifwRulesIndexResource{}
)

func NewIfwRulesIndexResource() resource.Resource {
	return &ifwRulesIndexResource{}
}

type ifwRulesIndexResource struct {
	client    *catoClientData
	ifwClient clientinterfaces.BulkInternetFirewallPolicyClient
}

func (r *ifwRulesIndexResource) getIfwClient() clientinterfaces.BulkInternetFirewallPolicyClient {
	if r.ifwClient != nil {
		return r.ifwClient
	}
	return r.client.catov2
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

	// Create maps to store API data for ID lookups
	sectionIdMap := make(map[string]string)
	ruleIdMap := make(map[string]string)
	ruleDescriptionMap := make(map[string]string)
	ruleEnabledMap := make(map[string]bool)

	// Get current sections from API to get fresh IDs
	ifwClient := r.getIfwClient()
	sectionIndexApiData, err := ifwClient.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallSectionsIndex error",
			err.Error(),
		)
		return
	}

	// Build map of section names to IDs from API
	for _, v := range sectionIndexApiData.Policy.InternetFirewall.Policy.Sections {
		sectionIdMap[v.Section.Name] = v.Section.ID
	}

	// Get current rules from API to get fresh IDs and computed values
	ruleIndexApiData, err := ifwClient.PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallRulesIndex error",
			err.Error(),
		)
		return
	}

	// Build maps of rule names to API data
	for _, v := range ruleIndexApiData.Policy.InternetFirewall.Policy.Rules {
		ruleIdMap[v.Rule.Name] = v.Rule.ID
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
					"id":            types.StringValue(sectionIdMap[sectionName]),
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
					"id":               types.StringValue(ruleIdMap[ruleName]),
					"index_in_section": existingRule.IndexInSection,                     // Preserve planned value
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

func (r *ifwRulesIndexResource) moveIfwRulesAndSections(ctx context.Context, plan IfwRulesIndex) (basetypes.MapValue, basetypes.MapValue, diag.Diagnostics, error) {
	diags := []diag.Diagnostic{}
	ruleObjectMap := make(map[string]attr.Value)
	sectionObjectMap := make(map[string]attr.Value)

	ifwClient := r.getIfwClient()
	sectionIndexApiData, err := ifwClient.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyInternetFirewallSectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexApiData),
	})
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyInternetFirewallSectionsIndex error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	if len(sectionIndexApiData.Policy.InternetFirewall.Policy.Sections) == 0 {
		input := cato_models.PolicyAddSectionInput{
			At: &cato_models.PolicySectionPositionInput{
				Position: cato_models.PolicySectionPositionEnumLastInPolicy,
			},
			Section: &cato_models.PolicyAddSectionInfoInput{
				Name: "Default Outbound Internet",
			},
		}
		sectionCreateApiData, err := ifwClient.PolicyInternetFirewallAddSection(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, input, r.client.AccountId)
		tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSectionWithinBulkMove.response", map[string]interface{}{
			"reason":   "creating new section as IFW does not have a default listed",
			"response": utils.InterfaceToJSONString(sectionCreateApiData),
		})
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyInternetFirewallAddSection error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}

		sectionIndexApiData, err = ifwClient.PolicyInternetFirewallSectionsIndex(ctx, r.client.AccountId)
		if err != nil {
			diags = append(diags, diag.NewErrorDiagnostic(
				"Catov2 API PolicyInternetFirewallSectionsIndex error",
				err.Error(),
			))
			return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
		}
	}

	ruleIndexApiData, err := ifwClient.PolicyInternetFirewallRulesIndex(ctx, r.client.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyInternetFirewallRulesIndex error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	sectionListFromPlan := make([]bulkPlanSection, 0, len(plan.SectionData.Elements()))
	for _, sectionValue := range plan.SectionData.Elements() {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex IfwRulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionListFromPlan = append(sectionListFromPlan, bulkPlanSection{
			Index: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			Name:  sectionSourceRuleIndex.SectionName.ValueString(),
		})
	}

	ruleListFromPlan := make([]bulkPlanRule, 0, len(plan.RuleData.Elements()))
	for _, ruleValue := range plan.RuleData.Elements() {
		ruleObject := ruleValue.(types.Object)
		var planSourceRuleIndex IfwRulesRuleItemIndex
		diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		ruleListFromPlan = append(ruleListFromPlan, bulkPlanRule{
			Index:       planSourceRuleIndex.IndexInSection.ValueInt64(),
			Name:        planSourceRuleIndex.RuleName.ValueString(),
			SectionName: planSourceRuleIndex.SectionName.ValueString(),
			Description: planSourceRuleIndex.Description.ValueString(),
			Enabled:     planSourceRuleIndex.Enabled.ValueBool(),
		})
	}

	currentSections := make([]bulkPolicySection, 0, len(sectionIndexApiData.Policy.InternetFirewall.Policy.Sections))
	for _, item := range sectionIndexApiData.Policy.InternetFirewall.Policy.Sections {
		currentSections = append(currentSections, bulkPolicySection{
			ID:         item.Section.ID,
			Name:       item.Section.Name,
			Properties: item.Properties,
		})
	}

	currentRules := make([]bulkPolicyRule, 0, len(ruleIndexApiData.Policy.InternetFirewall.Policy.Rules))
	for _, item := range ruleIndexApiData.Policy.InternetFirewall.Policy.Rules {
		currentRules = append(currentRules, bulkPolicyRule{
			ID:          item.Rule.ID,
			Name:        item.Rule.Name,
			SectionID:   item.Rule.Section.ID,
			SectionName: item.Rule.Section.Name,
			Description: item.Rule.Description,
			Enabled:     item.Rule.Enabled,
			Index:       item.Rule.Index,
			Properties:  item.Properties,
		})
	}

	reorderInput, sectionIDByName, ruleIDByName, err := buildBulkFirewallReorderInput(currentSections, currentRules, sectionListFromPlan, ruleListFromPlan, plan.SectionToStartAfterId.ValueString())
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyInternetFirewallReorderPolicy input error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	tflog.Debug(ctx, "Write.PolicyInternetFirewallReorderPolicy.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(reorderInput),
	})
	reorderResponse, err := ifwClient.PolicyInternetFirewallReorderPolicy(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, reorderInput, r.client.AccountId)
	tflog.Debug(ctx, "Write.PolicyInternetFirewallReorderPolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(reorderResponse),
	})
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyInternetFirewallReorderPolicy error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	reorderResult := reorderResponse.GetPolicy().GetInternetFirewall().GetReorderPolicy()
	if status := reorderResult.GetStatus(); status == nil || *status != cato_models.PolicyMutationStatusSuccess {
		errorMessage := "reorderPolicy returned a non-success status"
		for _, item := range reorderResult.GetErrors() {
			if item.GetErrorCode() != nil || item.GetErrorMessage() != nil {
				errorMessage = errorMessage + ": " + utils.InterfaceToJSONString(reorderResult.GetErrors())
				break
			}
		}
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyInternetFirewallReorderPolicy error",
			errorMessage,
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	for _, sectionFromPlan := range sectionListFromPlan {
		sectionIndexStateData, sectionDiags := types.ObjectValue(
			IfwSectionIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":            types.StringValue(sectionIDByName[sectionFromPlan.Name]),
				"section_name":  types.StringValue(sectionFromPlan.Name),
				"section_index": types.Int64Value(sectionFromPlan.Index),
			},
		)
		diags = append(diags, sectionDiags...)
		sectionObjectMap[sectionFromPlan.Name] = sectionIndexStateData
	}

	for _, ruleFromPlan := range ruleListFromPlan {
		ruleIndexStateData, ruleDiags := types.ObjectValue(
			IfwRuleIndexResourceAttrTypes,
			map[string]attr.Value{
				"id":               types.StringValue(ruleIDByName[ruleFromPlan.Name]),
				"index_in_section": types.Int64Value(ruleFromPlan.Index),
				"section_name":     types.StringValue(ruleFromPlan.SectionName),
				"rule_name":        types.StringValue(ruleFromPlan.Name),
				"description":      types.StringValue(ruleFromPlan.Description),
				"enabled":          types.BoolValue(ruleFromPlan.Enabled),
			},
		)
		diags = append(diags, ruleDiags...)
		ruleObjectMap[ruleFromPlan.Name] = ruleIndexStateData
	}

	_, err = ifwClient.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, &cato_models.PolicyPublishRevisionInput{}, r.client.AccountId)
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
