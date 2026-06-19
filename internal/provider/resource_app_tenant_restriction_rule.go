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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource                = &appTenantRestrictionRuleResource{}
	_ resource.ResourceWithConfigure   = &appTenantRestrictionRuleResource{}
	_ resource.ResourceWithImportState = &appTenantRestrictionRuleResource{}
)

func NewAppTenantRestrictionRuleResource() resource.Resource {
	return &appTenantRestrictionRuleResource{}
}

type appTenantRestrictionRuleResource struct {
	client *catoClientData
}

func (r *appTenantRestrictionRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_tenant_restriction_rule"
}

//nolint:funlen
func (r *appTenantRestrictionRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule in the Cato app tenant restriction policy. " +
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
					"action": schema.StringAttribute{
						Description: "Rule action",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(cato_models.AppTenantRestrictionActionEnumBypass),
								string(cato_models.AppTenantRestrictionActionEnumInjectHeaders),
							),
						},
					},
					"severity": schema.StringAttribute{
						Description: "Rule severity",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(cato_models.AppTenantRestrictionSeverityEnumHigh),
								string(cato_models.AppTenantRestrictionSeverityEnumMedium),
								string(cato_models.AppTenantRestrictionSeverityEnumLow),
							),
						},
					},
					"application": schema.SingleNestedAttribute{
						Description: "Target application reference",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
								Computed: true,
							},
							"name": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
								Computed: true,
							},
						},
					},
					"headers": schema.ListNestedAttribute{
						Description: "HTTP headers to inject when action is INJECT_HEADERS",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required: true,
								},
								"value": schema.StringAttribute{
									Required:  true,
									Sensitive: true,
								},
							},
						},
					},
					"schedule": schema.SingleNestedAttribute{
						Description: "Schedule for the rule",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"active_on": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							"custom_timeframe": schema.SingleNestedAttribute{
								Optional: true,
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"from": schema.StringAttribute{Optional: true, Computed: true},
									"to":   schema.StringAttribute{Optional: true, Computed: true},
								},
							},
							"custom_recurring": schema.SingleNestedAttribute{
								Optional: true,
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"from": schema.StringAttribute{Optional: true, Computed: true},
									"to":   schema.StringAttribute{Optional: true, Computed: true},
									"days": schema.ListAttribute{
										ElementType: types.StringType,
										Optional:    true,
										Computed:    true,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
					},
					"source": schema.SingleNestedAttribute{
						Description: "Source traffic matching criteria",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
							planmodifiers.SourceDestObjectModifier(),
						},
						Attributes: applicationControlSourceSchemaAttributes(),
					},
				},
			},
		},
	}
}

func (r *appTenantRestrictionRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *appTenantRestrictionRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

//nolint:gocyclo,funlen
func (r *appTenantRestrictionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppTenantRestrictionRule
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateAppTenantRestrictionAddRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicyAppTenantRestrictionAddRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	res, err := r.client.catov2.PolicyAppTenantRestrictionAddRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionAddRule error", err.Error())
		return
	}
	add := res.GetPolicy().GetAppTenantRestriction().GetAddRule()
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
		resp.Diagnostics.AddError("app tenant restriction addRule failed", st)
		return
	}

	if _, err := r.client.catov2.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionPublishPolicyRevision error", err.Error())
		return
	}

	body, err := r.client.catov2.AppTenantRestrictionPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppTenantRestrictionPolicy error", err.Error())
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
	var cur *cato_go_sdk.AppTenantRestrictionPolicy_Policy_AppTenantRestriction_Policy_Rules_Rule
	for _, item := range body.GetPolicy().GetAppTenantRestriction().GetPolicy().GetRules() {
		if item != nil && item.GetRule() != nil && item.GetRule().GetID() == newID {
			cur = item.GetRule()
			break
		}
	}
	if cur == nil {
		resp.Diagnostics.AddError("Read after create failed", "rule not found in policy response")
		return
	}

	rulePlan, hdiags := hydrateAppTenantRestrictionRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)
	ruleObj, odiags := types.ObjectValueFrom(ctx, AppTenantRestrictionRuleRuleAttrTypes, rulePlan)
	resp.Diagnostics.Append(odiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appTenantRestrictionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppTenantRestrictionRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.catov2.AppTenantRestrictionPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API AppTenantRestrictionPolicy error", err.Error())
		return
	}

	rulePlan := AppTenantRestrictionRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rulePlan, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	pol := body.GetPolicy().GetAppTenantRestriction().GetPolicy()
	if pol == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var cur *cato_go_sdk.AppTenantRestrictionPolicy_Policy_AppTenantRestriction_Policy_Rules_Rule
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

	hydrated, hdiags := hydrateAppTenantRestrictionRuleStateFromClient(ctx, cur)
	resp.Diagnostics.Append(hdiags...)
	ruleObj, odiags := types.ObjectValueFrom(ctx, AppTenantRestrictionRuleRuleAttrTypes, hydrated)
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
func (r *appTenantRestrictionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AppTenantRestrictionRule
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
	rule := AppTenantRestrictionRuleRulePlan{}
	resp.Diagnostics.Append(plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	move.ID = rule.ID.ValueString()
	if resp.Diagnostics.HasError() {
		return
	}

	moveRes, err := r.client.catov2.PolicyAppTenantRestrictionMoveRule(ctx, move, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionMoveRule error", err.Error())
		return
	}
	mv := moveRes.GetPolicy().GetAppTenantRestriction().GetMoveRule()
	if mv.GetStatus() == nil || *mv.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range mv.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error moving rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	upd, diags := hydrateAppTenantRestrictionUpdateRuleInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updRes, err := r.client.catov2.PolicyAppTenantRestrictionUpdateRule(ctx, upd, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionUpdateRule error", err.Error())
		return
	}
	ur := updRes.GetPolicy().GetAppTenantRestriction().GetUpdateRule()
	if ur.GetStatus() == nil || *ur.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		for _, e := range ur.GetErrors() {
			if e != nil && e.GetErrorCode() != nil {
				resp.Diagnostics.AddError("API error updating rule", fmt.Sprintf("%s : %s", *e.GetErrorCode(), *e.GetErrorMessage()))
				return
			}
		}
	}

	if _, err := r.client.catov2.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionPublishPolicyRevision error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appTenantRestrictionRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppTenantRestrictionRule
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rule := AppTenantRestrictionRuleRulePlan{}
	resp.Diagnostics.Append(state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	rm := cato_models.AppTenantRestrictionRemoveRuleInput{ID: rule.ID.ValueString()}
	if _, err := r.client.catov2.PolicyAppTenantRestrictionRemoveRule(ctx, rm, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionRemoveRule error", err.Error())
		return
	}
	if _, err := r.client.catov2.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, r.client.AccountId); err != nil {
		resp.Diagnostics.AddError("Cato API PolicyAppTenantRestrictionPublishPolicyRevision error", err.Error())
		return
	}
}
