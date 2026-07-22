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
	_ resource.Resource                = &ifSubPolicyResource{}
	_ resource.ResourceWithConfigure   = &ifSubPolicyResource{}
	_ resource.ResourceWithImportState = &ifSubPolicyResource{}
)

func NewIfSubPolicyResource() resource.Resource {
	return &ifSubPolicyResource{}
}

type ifSubPolicyResource struct {
	client        *catoClientData
	subPolyClient InternetFirewallSubPolicyClient
}

func (r *ifSubPolicyResource) getClient() InternetFirewallSubPolicyClient {
	if r.subPolyClient != nil {
		return r.subPolyClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.catov2
}

func (r *ifSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_sub_policy"
}

func (r *ifSubPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Reuse the exact cato_if_rule "rule" schema for the sub-policy scope so the
	// scope stays in lockstep (drift-free) with the rule resource.
	var ruleSchema resource.SchemaResponse
	(&internetFwRuleResource{}).Schema(ctx, resource.SchemaRequest{}, &ruleSchema)
	scopeAttr := ruleSchema.Schema.Attributes["rule"].(schema.SingleNestedAttribute)
	scopeAttr.Description = "Scope of the sub-policy. This is the SUB_POLICY_SCOPE rule that defines when the " +
		"sub-policy applies. Uses the same parameters as a cato_if_rule rule."
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
		Description: "The `cato_if_sub_policy` resource manages an Internet Firewall sub-policy " +
			"(a nested policy scoped by a SUB_POLICY_SCOPE rule). The Cato API has no updateSubPolicy " +
			"mutation, so changing `name` or `description` forces resource replacement. Documentation for the " +
			"underlying API can be found at mutation.policy.internetFirewall.addSubPolicy().",
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
			"at":    ifSubPolicyAtAttribute(),
			"scope": scopeAttr,
		},
	}
}

// ifSubPolicyAtAttribute defines the position of the sub-policy scope within the
// main policy. It forces replacement because there is no stable, drift-free way
// to reconcile the position from a read.
func ifSubPolicyAtAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Position of the sub-policy scope within the Internet Firewall policy.",
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
	}
}

func (r *ifSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *ifSubPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ifSubPolicyResource) publish(ctx context.Context) error {
	_, err := r.getClient().PolicyInternetFirewallPublishPolicyRevision(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		&cato_models.PolicyPublishRevisionInput{},
		r.client.AccountId,
	)
	return err
}

//nolint:funlen
func (r *ifSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InternetFirewallSubPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API requires a normal action while creating a scope rule, then returns
	// the API-owned SUB_POLICY action. Keep that bootstrap value out of config.
	scopeAttrs := maps.Clone(plan.Scope.Attributes())
	scopeAttrs["action"] = types.StringValue(string(cato_models.InternetFirewallActionEnumAllow))
	scopeAttrs["name"] = plan.Name
	scopeAttrs["description"] = plan.Description
	createScope, diags := types.ObjectValue(InternetFirewallRuleRuleAttrTypes, scopeAttrs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reuse the rule hydrator to build the scope rule data + position.
	scopeRule := InternetFirewallRule{Rule: createScope, At: plan.At}
	hydrated, diags := hydrateIfwRuleAPI(ctx, scopeRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Snapshot existing sub-policy ids to identify the new one after creation.
	before, err := r.getClient().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	existing := ifwSubPolicyIDs(before)

	addInput := cato_models.InternetFirewallAddSubPolicyInput{
		At: hydrated.create.At,
		Policy: &cato_models.InternetFirewallAddSubPolicyDataInput{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
		},
		Scope: hydrated.create.Rule,
	}

	addResp, err := r.getClient().PolicyInternetFirewallAddSubPolicy(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		addInput,
		r.client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallAddSubPolicy error", err.Error())
		return
	}
	addPayload := addResp.GetPolicy().GetInternetFirewall().GetAddSubPolicy()
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
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallPublishPolicyRevision error", err.Error())
		return
	}

	// Locate the new sub-policy + its scope rule.
	after, err := r.getClient().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	subID := ifwSubPolicyIDByName(after, plan.Name.ValueString(), existing)
	if subID == "" {
		resp.Diagnostics.AddError("Sub-Policy Not Found", "Created Internet Firewall sub-policy could not be located after publish.")
		return
	}
	scope := ifwScopeRule(after, subID)
	if scope == nil {
		resp.Diagnostics.AddError("Scope Rule Not Found", fmt.Sprintf("No SUB_POLICY_SCOPE rule found for sub-policy %s.", subID))
		return
	}

	plan.ID = types.StringValue(subID)
	plan.ScopeRuleID = types.StringValue(scope.GetID())
	scopeState := hydrateIfwRuleState(ctx, scopeRule, scope)
	scopeState.ID = types.StringValue(scope.GetID())
	scopeObj, diags := types.ObjectValueFrom(ctx, InternetFirewallRuleRuleAttrTypes, scopeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ifSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InternetFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.getClient().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}

	info := ifwSubPolicyInfo(body, state.ID.ValueString())
	if info == nil {
		tflog.Warn(ctx, "internet firewall sub-policy not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}
	scope := ifwScopeRule(body, state.ID.ValueString())
	if scope == nil {
		tflog.Warn(ctx, "internet firewall sub-policy scope rule not found, resource removed")
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

	scopeRule := InternetFirewallRule{Rule: state.Scope, At: state.At}
	scopeState := hydrateIfwRuleState(ctx, scopeRule, scope)
	scopeState.ID = types.StringValue(scope.GetID())
	scopeObj, diags := types.ObjectValueFrom(ctx, InternetFirewallRuleRuleAttrTypes, scopeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ifSubPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InternetFirewallSubPolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state InternetFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeRuleID := state.ScopeRuleID.ValueString()
	plan.ID = state.ID
	plan.ScopeRuleID = state.ScopeRuleID

	// name/description are RequiresReplace, so only the scope can change here.
	scopeRule := InternetFirewallRule{Rule: plan.Scope, At: plan.At}
	hydrated, diags := hydrateIfwRuleAPI(ctx, scopeRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	hydrated.update.ID = scopeRuleID
	// The API synchronizes scope name/description back to sub-policy metadata.
	hydrated.update.Rule.Name = state.Name.ValueStringPointer()
	hydrated.update.Rule.Description = state.Description.ValueStringPointer()

	updateResp, err := r.getClient().PolicyInternetFirewallUpdateRule(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		hydrated.update,
		r.client.AccountId,
	)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallUpdateRule error", err.Error())
		return
	}
	if updateResp.Policy.InternetFirewall.UpdateRule.Status != ifwMutationStatusSuccess {
		for _, e := range updateResp.Policy.InternetFirewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Sub-Policy Scope",
				fmt.Sprintf("%s : %s", derefStr(e.ErrorCode), derefStr(e.ErrorMessage)),
			)
		}
		return
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallPublishPolicyRevision error", err.Error())
		return
	}

	// Re-read scope for consistent state.
	body, err := r.getClient().PolicyInternetFirewall(ctx, &cato_models.InternetFirewallPolicyInput{}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	scope := ifwScopeRule(body, plan.ID.ValueString())
	if scope == nil {
		resp.Diagnostics.AddError("Scope Rule Not Found", "Sub-policy scope rule not found after update.")
		return
	}
	scopeState := hydrateIfwRuleState(ctx, scopeRule, scope)
	scopeState.ID = types.StringValue(scope.GetID())
	scopeObj, diags := types.ObjectValueFrom(ctx, InternetFirewallRuleRuleAttrTypes, scopeState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ifSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InternetFirewallSubPolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeInput := cato_models.InternetFirewallRemoveSubPolicyInput{Ref: objectRefByID(state.ID.ValueString())}
	removeResp, err := r.getClient().PolicyInternetFirewallRemoveSubPolicy(
		ctx,
		&cato_models.InternetFirewallPolicyMutationInput{},
		removeInput,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Delete.PolicyInternetFirewallRemoveSubPolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeResp),
	})
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallRemoveSubPolicy error", err.Error())
		return
	}
	removePayload := removeResp.GetPolicy().GetInternetFirewall().GetRemoveSubPolicy()
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
		resp.Diagnostics.AddError("Catov2 API Delete/PolicyInternetFirewallPublishPolicyRevision error", err.Error())
		return
	}
}

// derefStr safely dereferences an optional string pointer.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
