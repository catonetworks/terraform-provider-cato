package provider

import (
	"context"
	"fmt"
	"maps"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource                = &wfSubPolicyResource{}
	_ resource.ResourceWithConfigure   = &wfSubPolicyResource{}
	_ resource.ResourceWithImportState = &wfSubPolicyResource{}
)

func NewWfSubPolicyResource() resource.Resource {
	return &wfSubPolicyResource{}
}

type wfSubPolicyResource struct {
	client        *catoClientData
	subPolyClient WanFirewallSubPolicyClient
}

func (r *wfSubPolicyResource) getClient() WanFirewallSubPolicyClient {
	if r.subPolyClient != nil {
		return r.subPolyClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.catov2
}

func (r *wfSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wf_sub_policy"
}

func (r *wfSubPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	var ruleSchema resource.SchemaResponse
	(&wanFwRuleResource{}).Schema(ctx, resource.SchemaRequest{}, &ruleSchema)
	scopeAttr := ruleSchema.Schema.Attributes["rule"].(schema.SingleNestedAttribute)
	scopeAttr.Description = "Scope of the sub-policy. This is the SUB_POLICY_SCOPE rule that defines when the " +
		"sub-policy applies. Uses the same parameters as a cato_wf_rule rule."
	scopeAttr.Required = true
	scopeAttr.Attributes["action"] = schema.StringAttribute{
		Description: "API-managed action for the SUB_POLICY_SCOPE rule.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	scopeAttr.Attributes["name"] = schema.StringAttribute{
		Description: "API-managed scope name, synchronized with the sub-policy name.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	scopeAttr.Attributes["description"] = schema.StringAttribute{
		Description: "API-managed scope description, synchronized with the sub-policy description.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = schema.Schema{
		Description: "The `cato_wf_sub_policy` resource manages a WAN Firewall sub-policy " +
			"(a nested policy scoped by a SUB_POLICY_SCOPE rule). The Cato API has no updateSubPolicy " +
			"mutation, so changing `name` or `description` forces resource replacement. Documentation for the " +
			"underlying API can be found at mutation.policy.wanFirewall.addSubPolicy().",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Sub-policy ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Sub-policy name. Changing this forces replacement (no updateSubPolicy API).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Sub-policy description. Changing this forces replacement (no updateSubPolicy API).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope_rule_id": schema.StringAttribute{
				Description: "ID of the underlying SUB_POLICY_SCOPE rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Position of the sub-policy scope within the WAN Firewall policy.",
				Required:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule.",
						Required:    true,
					},
					"ref": schema.StringAttribute{
						Description: "Identifier of the object relative to which the position is defined.",
						Optional:    true,
					},
				},
			},
			"scope": scopeAttr,
		},
	}
}

func (r *wfSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *wfSubPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *wfSubPolicyResource) publish(ctx context.Context) error {
	_, err := r.getClient().PolicyWanFirewallPublishPolicyRevision(
		ctx,
		&cato_models.PolicyPublishRevisionInput{},
		r.client.AccountId,
	)
	return err
}

//nolint:funlen
func (r *wfSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WanFirewallSubPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API requires a normal action while creating a scope rule, then returns
	// the API-owned SUB_POLICY action. Keep that bootstrap value out of config.
	scopeAttrs := maps.Clone(plan.Scope.Attributes())
	scopeAttrs["action"] = types.StringValue(string(cato_models.WanFirewallActionEnumAllow))
	scopeAttrs["name"] = plan.Name
	scopeAttrs["description"] = plan.Description
	createScope, diags := types.ObjectValue(WanFirewallRuleRuleAttrTypes, scopeAttrs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeRule := WanFirewallRule{Rule: createScope, At: plan.At}
	hydrated, diags := hydrateWanRuleAPI(ctx, scopeRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	before, err := r.getClient().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}
	existing := wanSubPolicyIDs(before)

	addInput := cato_models.WanFirewallAddSubPolicyInput{
		At: hydrated.create.At,
		Policy: &cato_models.WanFirewallAddSubPolicyDataInput{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
		},
		Scope: hydrated.create.Rule,
	}

	addResp, err := r.getClient().PolicyWanFirewallAddSubPolicy(ctx, addInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallAddSubPolicy error", err.Error())
		return
	}
	addPayload := addResp.GetPolicy().GetWanFirewall().GetAddSubPolicy()
	if addPayload.GetStatus() == nil || *addPayload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range addPayload.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Sub-Policy",
				fmt.Sprintf("%s : %s", derefStr(e.ErrorCode), derefStr(e.ErrorMessage)),
			)
		}
		return
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallPublishPolicyRevision error", err.Error())
		return
	}

	after, err := r.getClient().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}
	subID := wanSubPolicyIDByName(after, plan.Name.ValueString(), existing)
	if subID == "" {
		resp.Diagnostics.AddError("Sub-Policy Not Found", "Created WAN Firewall sub-policy could not be located after publish.")
		return
	}
	scope := wanScopeRule(after, subID)
	if scope == nil {
		resp.Diagnostics.AddError("Scope Rule Not Found", fmt.Sprintf("No SUB_POLICY_SCOPE rule found for sub-policy %s.", subID))
		return
	}

	plan.ID = types.StringValue(subID)
	plan.ScopeRuleID = types.StringValue(scope.GetID())
	scopeState, diags := hydrateWanRuleState(ctx, scopeRule, scope)
	resp.Diagnostics.Append(diags...)
	scopeState.ID = types.StringValue(scope.GetID())
	scopeObj, diags := types.ObjectValueFrom(ctx, WanFirewallRuleRuleAttrTypes, scopeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *wfSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WanFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.getClient().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}

	info := wanSubPolicyInfo(body, state.ID.ValueString())
	if info == nil {
		tflog.Warn(ctx, "wan firewall sub-policy not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}
	scope := wanScopeRule(body, state.ID.ValueString())
	if scope == nil {
		tflog.Warn(ctx, "wan firewall sub-policy scope rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(info.GetName())
	if info.GetDescription() == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(info.GetDescription())
	}
	state.ScopeRuleID = types.StringValue(scope.GetID())

	scopeRule := WanFirewallRule{Rule: state.Scope, At: state.At}
	scopeState, diags := hydrateWanRuleState(ctx, scopeRule, scope)
	resp.Diagnostics.Append(diags...)
	scopeState.ID = types.StringValue(scope.GetID())
	scopeObj, diags := types.ObjectValueFrom(ctx, WanFirewallRuleRuleAttrTypes, scopeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wfSubPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WanFirewallSubPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state WanFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeRuleID := state.ScopeRuleID.ValueString()
	plan.ID = state.ID
	plan.ScopeRuleID = state.ScopeRuleID

	scopeRule := WanFirewallRule{Rule: plan.Scope, At: plan.At}
	hydrated, diags := hydrateWanRuleAPI(ctx, scopeRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	hydrated.update.ID = scopeRuleID
	// The API synchronizes scope name/description back to sub-policy metadata.
	hydrated.update.Rule.Name = state.Name.ValueStringPointer()
	hydrated.update.Rule.Description = state.Description.ValueStringPointer()

	updateResp, err := r.getClient().PolicyWanFirewallUpdateRule(ctx, hydrated.update, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallUpdateRule error", err.Error())
		return
	}
	if updateResp.Policy.WanFirewall.UpdateRule.Status != ifwMutationStatusSuccess {
		for _, e := range updateResp.Policy.WanFirewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError("API Error Updating Sub-Policy Scope", fmt.Sprintf("%s : %s", derefStr(e.ErrorCode), derefStr(e.ErrorMessage)))
		}
		return
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallPublishPolicyRevision error", err.Error())
		return
	}

	body, err := r.getClient().PolicyWanFirewall(ctx, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewall error", err.Error())
		return
	}
	scope := wanScopeRule(body, plan.ID.ValueString())
	if scope == nil {
		resp.Diagnostics.AddError("Scope Rule Not Found", "Sub-policy scope rule not found after update.")
		return
	}
	scopeState, diags := hydrateWanRuleState(ctx, scopeRule, scope)
	resp.Diagnostics.Append(diags...)
	scopeState.ID = types.StringValue(scope.GetID())
	scopeObj, diags := types.ObjectValueFrom(ctx, WanFirewallRuleRuleAttrTypes, scopeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *wfSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WanFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeInput := cato_models.WanFirewallRemoveSubPolicyInput{Ref: wanObjectRefByID(state.ID.ValueString())}
	removeResp, err := r.getClient().PolicyWanFirewallRemoveSubPolicy(ctx, removeInput, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicyWanFirewallRemoveSubPolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeResp),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyWanFirewallRemoveSubPolicy error", err.Error())
		return
	}
	removePayload := removeResp.GetPolicy().GetWanFirewall().GetRemoveSubPolicy()
	if removePayload.GetStatus() != nil && *removePayload.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range removePayload.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Removing Sub-Policy",
				fmt.Sprintf("%s : %s", derefStr(e.ErrorCode), derefStr(e.ErrorMessage)),
			)
		}
		return
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API Delete/PolicyWanFirewallPublishPolicyRevision error", err.Error())
		return
	}
}
