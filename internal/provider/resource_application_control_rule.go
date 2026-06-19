//nolint:lll
package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource                = &applicationControlRuleResource{}
	_ resource.ResourceWithConfigure   = &applicationControlRuleResource{}
	_ resource.ResourceWithImportState = &applicationControlRuleResource{}
)

func NewApplicationControlRuleResource() resource.Resource {
	return &applicationControlRuleResource{}
}

type applicationControlRuleResource struct {
	client *catoClientData
}

func (r *applicationControlRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_control_rule"
}

func (r *applicationControlRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	typed := applicationControlTypedRuleSchemaAttributes()
	resp.Schema = schema.Schema{
		Description: "Manages a rule in the Cato Application Control (App & Data Inline Protection) policy. " +
			"Underlying GraphQL is marked @beta; behavior and fields may change.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Rule ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Where to insert the rule",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to policy, section, or another rule",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(
								"AFTER_RULE", "BEFORE_RULE", "FIRST_IN_POLICY", "FIRST_IN_SECTION",
								"LAST_IN_POLICY", "LAST_IN_SECTION",
							),
						},
					},
					"ref": schema.StringAttribute{
						Description: "Reference rule or section ID when required by position",
						Optional:    true,
					},
				},
			},
			"rule": schema.SingleNestedAttribute{
				Description: "Rule definition",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Rule ID",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Rule name",
						Required:    true,
					},
					"description": schema.StringAttribute{
						Description: "Rule description",
						Optional:    true,
					},
					"enabled": schema.BoolAttribute{
						Description: "Whether the rule is enabled",
						Required:    true,
					},
					"rule_type": schema.StringAttribute{
						Description: "Which nested rule block is active",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(cato_models.ApplicationControlRuleTypeApplication),
								string(cato_models.ApplicationControlRuleTypeData),
								string(cato_models.ApplicationControlRuleTypeFile),
							),
						},
					},
					"application_rule": schema.SingleNestedAttribute{
						Description: "Settings when rule_type is APPLICATION",
						Optional:    true,
						Attributes:  typed,
					},
					"data_rule": schema.SingleNestedAttribute{
						Description: "Settings when rule_type is DATA",
						Optional:    true,
						Attributes:  typed,
					},
					"file_rule": schema.SingleNestedAttribute{
						Description: "Settings when rule_type is FILE",
						Optional:    true,
						Attributes:  typed,
					},
				},
			},
		},
	}
}

func (r *applicationControlRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *applicationControlRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

//nolint:gocyclo,funlen
func (r *applicationControlRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationControlRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateApplicationControlAddRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyApplicationControlAddRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	res, err := r.client.catov2.PolicyApplicationControlAddRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlAddRule error", err.Error())
		return
	}
	add := res.GetPolicy().GetApplicationControl().GetAddRule()
	for _, e := range add.GetErrors() {
		if e != nil && e.GetErrorCode() != nil {
			resp.Diagnostics.AddError("API error: "+*e.GetErrorCode(), *e.GetErrorMessage())
			return
		}
	}
	if add.GetStatus() == nil || *add.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		st := ""
		if add.GetStatus() != nil {
			st = string(*add.GetStatus())
		}
		resp.Diagnostics.AddError("Application Control addRule failed", st)
		return
	}

	if _, err := r.client.catov2.PolicyApplicationControlPublishPolicyRevision(ctx, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlPublishPolicyRevision error", err.Error())
		return
	}

	body, err := r.client.catov2.ApplicationControlPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API ApplicationControlPolicy error", err.Error())
		return
	}

	newID := ""
	if add.GetRule() != nil && add.GetRule().GetRule() != nil {
		newID = add.GetRule().GetRule().GetID()
	}
	if newID == "" {
		resp.Diagnostics.AddError("Read after create failed", "addRule response did not return rule id")
		return
	}

	var cur *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetApplicationControl().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == newID {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after create failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateApplicationControlRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)
	ruleObj, odiags := types.ObjectValueFrom(ctx, ApplicationControlRuleRuleAttrTypes, rulePlan)
	resp.Diagnostics.Append(odiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationControlRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationControlRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.catov2.ApplicationControlPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API ApplicationControlPolicy error", err.Error())
		return
	}

	rulePlan := ApplicationControlRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rulePlan, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetApplicationControl().GetPolicy()
	if pol == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var cur *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule
	for _, item := range pol.GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == rulePlan.ID.ValueString() {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	hydrated, hdiags := hydrateApplicationControlRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)
	ruleObj, odiags := types.ObjectValueFrom(ctx, ApplicationControlRuleRuleAttrTypes, hydrated)
	resp.Diagnostics.Append(odiags...)

	atObj, d := types.ObjectValue(PositionAttrTypes, map[string]attr.Value{
		"position": types.StringValue("LAST_IN_POLICY"),
		"ref":      types.StringNull(),
	})
	resp.Diagnostics.Append(d...)

	state.ID = types.StringValue(cur.GetID())
	state.Rule = ruleObj
	state.At = atObj
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

//nolint:gocyclo
func (r *applicationControlRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationControlRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	move := cato_models.PolicyMoveRuleInput{}
	if !plan.At.IsNull() {
		pos := PolicyRulePositionInput{}
		resp.Diagnostics.Append(plan.At.As(ctx, &pos, basetypes.ObjectAsOptions{})...)
		move.To = &cato_models.PolicyRulePositionInput{
			Position: (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer()),
			Ref:      pos.Ref.ValueStringPointer(),
		}
	}
	rule := ApplicationControlRuleRulePlan{}
	resp.Diagnostics.Append(plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	move.ID = rule.ID.ValueString()
	if resp.Diagnostics.HasError() {
		return
	}

	moveRes, err := r.client.catov2.PolicyApplicationControlMoveRule(ctx, move, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlMoveRule error", err.Error())
		return
	}
	mv := moveRes.GetPolicy().GetApplicationControl().GetMoveRule()
	if mv.GetStatus() == nil || *mv.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range mv.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error moving rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	upd, diags := hydrateApplicationControlUpdateRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updRes, err := r.client.catov2.PolicyApplicationControlUpdateRule(ctx, upd, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlUpdateRule error", err.Error())
		return
	}
	ur := updRes.GetPolicy().GetApplicationControl().GetUpdateRule()
	if ur.GetStatus() == nil || *ur.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range ur.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error updating rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	if _, err := r.client.catov2.PolicyApplicationControlPublishPolicyRevision(ctx, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlPublishPolicyRevision error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationControlRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationControlRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule := ApplicationControlRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	rm := cato_models.ApplicationControlRemoveRuleInput{ID: rule.ID.ValueString()}
	if _, err := r.client.catov2.PolicyApplicationControlRemoveRule(ctx, rm, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlRemoveRule error", err.Error())
		return
	}
	if _, err := r.client.catov2.PolicyApplicationControlPublishPolicyRevision(ctx, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyApplicationControlPublishPolicyRevision error", err.Error())
		return
	}
}
