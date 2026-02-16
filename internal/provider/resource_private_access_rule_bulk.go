package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &privAccessRuleBulkResource{}
	_ resource.ResourceWithConfigure = &privAccessRuleBulkResource{}

	ErrConvertError = errors.New("failed convert terraform types to go")
)

func NewPrivAccessRuleBulkResource() resource.Resource {
	return &privAccessRuleBulkResource{}
}

type privAccessRuleBulkResource struct {
	client *catoClientData
}

func (r *privAccessRuleBulkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_access_rule_bulk"
}

func (r *privAccessRuleBulkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ordering and publishng private access policy rules.",
		Attributes: map[string]schema.Attribute{
			"rule_data": schema.MapNestedAttribute{
				Description: "Map of private access rule policy Indexes keyed by rule_name",
				Required:    false,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Rule id",
							Computed:    true,
						},
						"index": schema.Int64Attribute{
							Description: "Requierd position of the rule",
							Required:    true,
						},
						"cma_index": schema.Int64Attribute{
							Description: "Rule index in CMA before bulk is applied",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "IFW rule name housing rule",
							Required:    true,
						},
					},
				},
			},
		},
	}

}

func (r *privAccessRuleBulkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PrivateAccessRuleBulkModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydratePrivAccessRuleBulkState(ctx, &plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privateAccessRuleBulk state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessRuleBulkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "XXX Rule Update")
	var plan PrivateAccessRuleBulkModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PrivateAccessRuleBulkModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessRuleBulkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "XXX Rule Read")
	var state PrivateAccessRuleBulkModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydratePrivAccessRuleBulkState(ctx, &state)
	if hydrateErr != nil {
		if errors.Is(hydrateErr, ErrPrivateAcccessRuleNotFound) {
			// tflog.Warn(ctx, fmt.Sprintf("Private access rule %s not found, resource removed", state.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error hydrating privateAccessRuleBulk state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessRuleBulkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "XXX Rule Delete")
}

func (r *privAccessRuleBulkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

// hydratePrivAccessRuleBulkState fetches the current state of a privAccessRuleBulk from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privAccessRuleBulkResource) hydratePrivAccessRuleBulkState(ctx context.Context, plan *PrivateAccessRuleBulkModel) (*PrivateAccessRuleBulkModel, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	// Create a Go map[rule-name]PrivateAccessBulkRule{...} from the plan/state
	stateRules := make(map[string]PrivateAccessBulkRule)
	if hasValue(plan.RuleData) {
		var tfBulk map[string]types.Object
		if checkErr(&diags, plan.RuleData.ElementsAs(ctx, &tfBulk, false)) {
			return nil, diags, ErrConvertError
		}
		for name, obj := range tfBulk {
			var tfRule PrivateAccessBulkRule
			diags.Append(obj.As(ctx, &tfRule, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags, ErrConvertError
			}
			stateRules[name] = tfRule
		}
	}

	// Call Cato API to get the policy
	result, err := r.client.catov2.PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Bulk PolicyReadPrivateAccessPolicy", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		return nil, nil, err
	}

	// Merge rules from the API and rules from the state
	apiRules := make(map[string]types.Object)
	policy := result.GetPolicy().GetPrivateAccess().GetPolicy()
	if len(policy.Rules) != len(stateRules) {
		return nil, nil, fmt.Errorf("Rules returned by API (count=%d) do not match the PrivateAccessRuleBulkModel state (count=%d)", len(policy.Rules), len(stateRules))
	}
	for _, polRule := range policy.Rules {
		apiRule := polRule.Rule
		tfRule := &PrivateAccessBulkRule{
			ID:       types.StringValue(apiRule.ID),
			Name:     types.StringValue(apiRule.Name),
			CMAIndex: types.Int64Value(apiRule.Index),
		}
		stateRule, ok := stateRules[apiRule.Name]
		if !ok {
			return nil, nil, fmt.Errorf("Rule '%s' from API is not found in the PrivateAccessRuleBulkModel state", apiRule.Name)
		}
		tfRule.Index = stateRule.Index
		ruleObj, diag := types.ObjectValueFrom(ctx, PrivateAccessBulkRuleTypes, tfRule)
		diags.Append(diag...)
		if diags.HasError() {
			return nil, diags, ErrAPIResponseParse
		}
		apiRules[apiRule.Name] = ruleObj
	}

	// Prepate the tf state - ruleData map from enriched apiRules
	rulesMap, diag := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: PrivateAccessBulkRuleTypes}, apiRules)
	diags.Append(diag...)
	if diags.HasError() {
		return nil, diags, ErrAPIResponseParse
	}
	state := &PrivateAccessRuleBulkModel{
		RuleData: rulesMap,
	}

	return state, nil, nil
}
