package provider

import (
	"context"

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

func (r *wanRulesIndexResource) moveWanRulesAndSections(ctx context.Context, plan WanRulesIndex) (basetypes.MapValue, basetypes.MapValue, diag.Diagnostics, error) {
	diags := []diag.Diagnostic{}
	ruleObjectMap := make(map[string]attr.Value)
	sectionObjectMap := make(map[string]attr.Value)

	sectionIndexApiData, err := r.client.catov2.PolicyWanFirewallSectionsIndex(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Read.PolicyWanFirewallSectionsIndexInCreate.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(sectionIndexApiData),
	})
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyWanFirewallSectionsIndex error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	ruleIndexApiData, err := r.client.catov2.PolicyWanFirewallRulesIndex(ctx, r.client.AccountId)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyWanFirewallRulesIndex error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	sectionListFromPlan := make([]bulkPlanSection, 0, len(plan.SectionData.Elements()))
	for _, sectionValue := range plan.SectionData.Elements() {
		sectionObject := sectionValue.(types.Object)
		var sectionSourceRuleIndex WanRulesSectionItemIndex
		diags = append(diags, sectionObject.As(ctx, &sectionSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		sectionListFromPlan = append(sectionListFromPlan, bulkPlanSection{
			Index: sectionSourceRuleIndex.SectionIndex.ValueInt64(),
			Name:  sectionSourceRuleIndex.SectionName.ValueString(),
		})
	}

	ruleListFromPlan := make([]bulkPlanRule, 0, len(plan.RuleData.Elements()))
	for _, ruleValue := range plan.RuleData.Elements() {
		ruleObject := ruleValue.(types.Object)
		var planSourceRuleIndex WanRulesRuleItemIndex
		diags = append(diags, ruleObject.As(ctx, &planSourceRuleIndex, basetypes.ObjectAsOptions{})...)

		ruleListFromPlan = append(ruleListFromPlan, bulkPlanRule{
			Index:       planSourceRuleIndex.IndexInSection.ValueInt64(),
			Name:        planSourceRuleIndex.RuleName.ValueString(),
			SectionName: planSourceRuleIndex.SectionName.ValueString(),
			Description: planSourceRuleIndex.Description.ValueString(),
			Enabled:     planSourceRuleIndex.Enabled.ValueBool(),
		})
	}

	currentSections := make([]bulkPolicySection, 0, len(sectionIndexApiData.Policy.WanFirewall.Policy.Sections))
	for _, item := range sectionIndexApiData.Policy.WanFirewall.Policy.Sections {
		currentSections = append(currentSections, bulkPolicySection{
			ID:         item.Section.ID,
			Name:       item.Section.Name,
			Properties: item.Properties,
		})
	}

	currentRules := make([]bulkPolicyRule, 0, len(ruleIndexApiData.Policy.WanFirewall.Policy.Rules))
	for _, item := range ruleIndexApiData.Policy.WanFirewall.Policy.Rules {
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
			"Catov2 API PolicyWanFirewallReorderPolicy input error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	tflog.Debug(ctx, "Write.PolicyWanFirewallReorderPolicy.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(reorderInput),
	})
	reorderResponse, err := r.client.catov2.PolicyWanFirewallReorderPolicy(ctx, &cato_models.WanFirewallPolicyMutationInput{}, reorderInput, r.client.AccountId)
	tflog.Debug(ctx, "Write.PolicyWanFirewallReorderPolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(reorderResponse),
	})
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyWanFirewallReorderPolicy error",
			err.Error(),
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	reorderResult := reorderResponse.GetPolicy().GetWanFirewall().GetReorderPolicy()
	if status := reorderResult.GetStatus(); status == nil || *status != cato_models.PolicyMutationStatusSuccess {
		errorMessage := "reorderPolicy returned a non-success status"
		for _, item := range reorderResult.GetErrors() {
			if item.GetErrorCode() != nil || item.GetErrorMessage() != nil {
				errorMessage = errorMessage + ": " + utils.InterfaceToJSONString(reorderResult.GetErrors())
				break
			}
		}
		diags = append(diags, diag.NewErrorDiagnostic(
			"Catov2 API PolicyWanFirewallReorderPolicy error",
			errorMessage,
		))
		return basetypes.MapValue{}, basetypes.MapValue{}, diags, err
	}

	for _, sectionFromPlan := range sectionListFromPlan {
		sectionIndexStateData, sectionDiags := types.ObjectValue(
			WanSectionIndexResourceAttrTypes,
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
			WanRuleIndexResourceAttrTypes,
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

	_, err = r.client.catov2.PolicyWanFirewallPublishPolicyRevision(ctx, &cato_models.PolicyPublishRevisionInput{}, r.client.AccountId)
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
