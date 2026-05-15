package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource                = &internetFwSubPolicyResource{}
	_ resource.ResourceWithConfigure   = &internetFwSubPolicyResource{}
	_ resource.ResourceWithImportState = &internetFwSubPolicyResource{}
)

func NewInternetFwSubPolicyResource() resource.Resource {
	return &internetFwSubPolicyResource{}
}

type internetFwSubPolicyResource struct {
	client *catoClientData
}

func (r *internetFwSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_sub_policy"
}

func (r *internetFwSubPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_if_sub_policy` resource creates an Internet Firewall sub-policy and manages its scope rule. Rules inside the sub-policy are managed separately with `cato_if_rule`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Sub-policy ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Position of the sub-policy scope rule.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, section, or rule.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("AFTER_RULE", "BEFORE_RULE", "FIRST_IN_POLICY", "FIRST_IN_SECTION", "LAST_IN_POLICY", "LAST_IN_SECTION"),
						},
					},
					"ref": schema.StringAttribute{
						Description: "The identifier of the rule or section relative to which the sub-policy scope rule is placed.",
						Optional:    true,
					},
				},
			},
			"policy": schema.SingleNestedAttribute{
				Description: "Sub-policy metadata.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Sub-policy ID",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Sub-policy name",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"description": schema.StringAttribute{
						Description: "Sub-policy description",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"scope": schema.SingleNestedAttribute{
				Description: "Sub-policy scope rule. The API requires all-ANY scope rules to be disabled.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Scope rule ID",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Scope rule name",
						Required:    true,
					},
					"enabled": schema.BoolAttribute{
						Description: "Scope rule enabled state",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}

func (r *internetFwSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *internetFwSubPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root("policy").AtName("id"), req, resp)
}

func (r *internetFwSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InternetFirewallSubPolicy
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := buildIfSubPolicyAddInput(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Internet Firewall Sub-Policy", err.Error())
		return
	}

	createResp, err := r.addSubPolicy(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallAddSubPolicy error", err.Error())
		return
	}
	if createResp.Policy == nil || createResp.Policy.InternetFirewall == nil || createResp.Policy.InternetFirewall.AddSubPolicy == nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallAddSubPolicy error", "empty response")
		return
	}
	if createResp.Policy.InternetFirewall.AddSubPolicy.Status != "SUCCESS" {
		addPolicyMutationErrors(resp, "API Error Creating Internet Firewall Sub-Policy", createResp.Policy.InternetFirewall.AddSubPolicy.Errors)
		return
	}

	policyState, err := r.readSubPolicyState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}

	policy, err := findIfSubPolicyByName(policyState, input.Policy.Name)
	if err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy not found after create", err.Error())
		return
	}
	scopeRule, err := findIfSubPolicyScopeRule(policyState, policy.ID)
	if err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy scope rule not found after create", err.Error())
		return
	}

	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallPublishPolicyRevision error", err.Error())
		return
	}

	policyState, err = r.readSubPolicyState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	scopeRule, err = findIfSubPolicyScopeRule(policyState, policy.ID)
	if err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy scope rule not found after publish", err.Error())
		return
	}

	if err := setIfSubPolicyState(ctx, &plan, policy, scopeRule); err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy state error", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *internetFwSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InternetFirewallSubPolicy
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := state.ID.ValueString()
	if policyID == "" && !state.Policy.IsNull() && !state.Policy.IsUnknown() {
		policy := InternetFirewallSubPolicyInfo{}
		diags = state.Policy.As(ctx, &policy, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		policyID = policy.ID.ValueString()
	}

	policyState, err := r.readSubPolicyState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}

	policy, err := findIfSubPolicyByID(policyState, policyID)
	if err != nil {
		tflog.Warn(ctx, "internet firewall sub-policy not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}
	scopeRule, err := findIfSubPolicyScopeRule(policyState, policy.ID)
	if err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy scope rule not found", err.Error())
		return
	}

	if err := setIfSubPolicyState(ctx, &state, policy, scopeRule); err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy state error", err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *internetFwSubPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InternetFirewallSubPolicy
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyState, err := r.readSubPolicyState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	policy, err := findIfSubPolicyByID(policyState, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy not found", err.Error())
		return
	}
	scopeRule, err := findIfSubPolicyScopeRule(policyState, policy.ID)
	if err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy scope rule not found", err.Error())
		return
	}

	if err := r.moveSubPolicyScope(ctx, scopeRule.ID, plan.At); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallMoveRule error", err.Error())
		return
	}
	if err := r.updateSubPolicyScope(ctx, scopeRule.ID, plan.Scope); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallUpdateRule error", err.Error())
		return
	}
	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallPublishPolicyRevision error", err.Error())
		return
	}

	policyState, err = r.readSubPolicyState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	scopeRule, err = findIfSubPolicyScopeRule(policyState, policy.ID)
	if err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy scope rule not found after update", err.Error())
		return
	}

	if err := setIfSubPolicyState(ctx, &plan, policy, scopeRule); err != nil {
		resp.Diagnostics.AddError("Internet Firewall Sub-Policy state error", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *internetFwSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InternetFirewallSubPolicy
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.removeSubPolicy(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallRemoveSubPolicy error", err.Error())
		return
	}
	if err := r.publish(ctx); err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewallPublishPolicyRevision error", err.Error())
		return
	}
}

type ifSubPolicyAddInput struct {
	At     *cato_models.PolicyRulePositionInput `json:"at"`
	Policy *ifSubPolicyAddPolicyInput           `json:"policy"`
	Scope  *ifSubPolicyAddRuleInput             `json:"scope"`
}

type ifSubPolicyAddPolicyInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ifSubPolicyAddRuleInput struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Enabled     bool           `json:"enabled"`
	Action      string         `json:"action"`
	Source      map[string]any `json:"source"`
	Destination map[string]any `json:"destination"`
}

type ifSubPolicyMutationError struct {
	ErrorCode    *string `json:"errorCode"`
	ErrorMessage *string `json:"errorMessage"`
}

type ifSubPolicyAddResponse struct {
	Policy *struct {
		InternetFirewall *struct {
			AddSubPolicy *struct {
				Status string                      `json:"status"`
				Errors []*ifSubPolicyMutationError `json:"errors"`
			} `json:"addSubPolicy"`
		} `json:"internetFirewall"`
	} `json:"policy"`
}

type ifSubPolicyRemoveResponse struct {
	Policy *struct {
		InternetFirewall *struct {
			RemoveSubPolicy *struct {
				Status string                      `json:"status"`
				Errors []*ifSubPolicyMutationError `json:"errors"`
			} `json:"removeSubPolicy"`
		} `json:"internetFirewall"`
	} `json:"policy"`
}

type ifSubPolicyMoveRuleResponse struct {
	Policy *struct {
		InternetFirewall *struct {
			MoveRule *struct {
				Status string                      `json:"status"`
				Errors []*ifSubPolicyMutationError `json:"errors"`
			} `json:"moveRule"`
		} `json:"internetFirewall"`
	} `json:"policy"`
}

type ifSubPolicyUpdateRuleResponse struct {
	Policy *struct {
		InternetFirewall *struct {
			UpdateRule *struct {
				Status string                      `json:"status"`
				Errors []*ifSubPolicyMutationError `json:"errors"`
			} `json:"updateRule"`
		} `json:"internetFirewall"`
	} `json:"policy"`
}

type ifSubPolicyPublishResponse struct {
	Policy *struct {
		InternetFirewall *struct {
			PublishPolicyRevision *struct {
				Status string                      `json:"status"`
				Errors []*ifSubPolicyMutationError `json:"errors"`
			} `json:"publishPolicyRevision"`
		} `json:"internetFirewall"`
	} `json:"policy"`
}

type ifSubPolicyStateResponse struct {
	Policy *struct {
		InternetFirewall *struct {
			Policy *ifSubPolicyPolicyState `json:"policy"`
		} `json:"internetFirewall"`
	} `json:"policy"`
}

type ifSubPolicyPolicyState struct {
	Rules       []*ifSubPolicyRulePayload `json:"rules"`
	SubPolicies []*ifSubPolicyPayload     `json:"subPolicies"`
}

type ifSubPolicyPayload struct {
	Policy *ifSubPolicyInfoState `json:"policy"`
}

type ifSubPolicyInfoState struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	PolicyLevel string `json:"policyLevel"`
}

type ifSubPolicyRulePayload struct {
	Rule      *ifSubPolicyRuleState `json:"rule"`
	RuleType  string                `json:"ruleType"`
	SubPolicy *ifSubPolicyRefState  `json:"subPolicy"`
}

type ifSubPolicyRuleState struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Index   int64  `json:"index"`
	Action  string `json:"action"`
	Enabled bool   `json:"enabled"`
	Section *struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		SubPolicyID *string `json:"subPolicyId"`
	} `json:"section"`
}

type ifSubPolicyRefState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const ifSubPolicyAddDocument = `mutation policyInternetFirewallAddSubPolicy($accountId: ID!, $input: InternetFirewallAddSubPolicyInput!) {
  policy(accountId: $accountId) {
    internetFirewall {
      addSubPolicy(input: $input) {
        status
        errors {
          errorCode
          errorMessage
        }
      }
    }
  }
}`

const ifSubPolicyRemoveDocument = `mutation policyInternetFirewallRemoveSubPolicy($accountId: ID!, $id: ID!) {
  policy(accountId: $accountId) {
    internetFirewall {
      removeSubPolicy(input: { ref: { by: ID, input: $id } }) {
        status
        errors {
          errorCode
          errorMessage
        }
      }
    }
  }
}`

const ifSubPolicyMoveRuleDocument = `mutation policyInternetFirewallMoveSubPolicyScope($accountId: ID!, $input: PolicyMoveRuleInput!) {
  policy(accountId: $accountId) {
    internetFirewall {
      moveRule(input: $input) {
        status
        errors {
          errorCode
          errorMessage
        }
      }
    }
  }
}`

const ifSubPolicyUpdateRuleDocument = `mutation policyInternetFirewallUpdateSubPolicyScope($accountId: ID!, $input: InternetFirewallUpdateRuleInput!) {
  policy(accountId: $accountId) {
    internetFirewall {
      updateRule(input: $input) {
        status
        errors {
          errorCode
          errorMessage
        }
      }
    }
  }
}`

const ifSubPolicyPublishDocument = `mutation policyInternetFirewallPublishSubPolicyRevision($accountId: ID!) {
  policy(accountId: $accountId) {
    internetFirewall {
      publishPolicyRevision(input: {}) {
        status
        errors {
          errorCode
          errorMessage
        }
      }
    }
  }
}`

const ifSubPolicyStateDocument = `query policyInternetFirewallSubPolicies($accountId: ID!) {
  policy(accountId: $accountId) {
    internetFirewall {
      policy {
        rules {
          rule {
            id
			name
			index
			enabled
			section {
              id
              name
              subPolicyId
            }
          }
          ruleType
          subPolicy {
            id
            name
          }
        }
        subPolicies {
          policy {
            id
            name
            description
            enabled
            policyLevel
          }
        }
      }
    }
  }
}`

func buildIfSubPolicyAddInput(ctx context.Context, plan InternetFirewallSubPolicy) (*ifSubPolicyAddInput, error) {
	at := PolicyRulePositionInput{}
	diags := plan.At.As(ctx, &at, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("%v", diags.Errors())
	}
	policy := InternetFirewallSubPolicyInfo{}
	diags = plan.Policy.As(ctx, &policy, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("%v", diags.Errors())
	}
	scope := InternetFirewallSubPolicyScope{}
	diags = plan.Scope.As(ctx, &scope, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("%v", diags.Errors())
	}

	position := cato_models.PolicyRulePositionEnum(at.Position.ValueString())
	input := &ifSubPolicyAddInput{
		At: &cato_models.PolicyRulePositionInput{
			Position: &position,
			Ref:      at.Ref.ValueStringPointer(),
		},
		Policy: &ifSubPolicyAddPolicyInput{
			Name:        policy.Name.ValueString(),
			Description: policy.Description.ValueString(),
		},
		Scope: &ifSubPolicyAddRuleInput{
			Name:        scope.Name.ValueString(),
			Description: "",
			Enabled:     scope.Enabled.ValueBool(),
			Action:      "SUB_POLICY",
			Source:      map[string]any{},
			Destination: map[string]any{},
		},
	}
	return input, nil
}

func (r *internetFwSubPolicyResource) addSubPolicy(ctx context.Context, input *ifSubPolicyAddInput) (*ifSubPolicyAddResponse, error) {
	var res ifSubPolicyAddResponse
	vars := map[string]any{
		"accountId": r.client.AccountId,
		"input":     input,
	}
	tflog.Debug(ctx, "Create.PolicyInternetFirewallAddSubPolicy.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(vars),
	})
	err := r.client.catov2.Client.Post(ctx, "policyInternetFirewallAddSubPolicy", ifSubPolicyAddDocument, &res, vars)
	if err != nil && r.client.catov2.Client.ParseDataWhenErrors && res.Policy != nil {
		return &res, nil
	}
	return &res, err
}

func (r *internetFwSubPolicyResource) removeSubPolicy(ctx context.Context, id string) error {
	var res ifSubPolicyRemoveResponse
	vars := map[string]any{
		"accountId": r.client.AccountId,
		"id":        id,
	}
	err := r.client.catov2.Client.Post(ctx, "policyInternetFirewallRemoveSubPolicy", ifSubPolicyRemoveDocument, &res, vars)
	if err != nil {
		return err
	}
	if res.Policy == nil || res.Policy.InternetFirewall == nil || res.Policy.InternetFirewall.RemoveSubPolicy == nil {
		return fmt.Errorf("empty response")
	}
	if res.Policy.InternetFirewall.RemoveSubPolicy.Status != "SUCCESS" {
		return formatIfSubPolicyErrors(res.Policy.InternetFirewall.RemoveSubPolicy.Errors)
	}
	return nil
}

func (r *internetFwSubPolicyResource) readSubPolicyState(ctx context.Context) (*ifSubPolicyPolicyState, error) {
	var res ifSubPolicyStateResponse
	vars := map[string]any{
		"accountId": r.client.AccountId,
	}
	err := r.client.catov2.Client.Post(ctx, "policyInternetFirewallSubPolicies", ifSubPolicyStateDocument, &res, vars)
	if err != nil {
		return nil, err
	}
	if res.Policy == nil || res.Policy.InternetFirewall == nil || res.Policy.InternetFirewall.Policy == nil {
		return nil, fmt.Errorf("empty response")
	}
	return res.Policy.InternetFirewall.Policy, nil
}

func (r *internetFwSubPolicyResource) moveSubPolicyScope(ctx context.Context, scopeRuleID string, atObj types.Object) error {
	at := PolicyRulePositionInput{}
	diags := atObj.As(ctx, &at, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return fmt.Errorf("%v", diags.Errors())
	}
	position := cato_models.PolicyRulePositionEnum(at.Position.ValueString())
	input := cato_models.PolicyMoveRuleInput{
		ID: scopeRuleID,
		To: &cato_models.PolicyRulePositionInput{
			Position: &position,
			Ref:      at.Ref.ValueStringPointer(),
		},
	}
	var res ifSubPolicyMoveRuleResponse
	vars := map[string]any{
		"accountId": r.client.AccountId,
		"input":     input,
	}
	err := r.client.catov2.Client.Post(ctx, "policyInternetFirewallMoveSubPolicyScope", ifSubPolicyMoveRuleDocument, &res, vars)
	if err != nil {
		return err
	}
	if res.Policy == nil || res.Policy.InternetFirewall == nil || res.Policy.InternetFirewall.MoveRule == nil {
		return fmt.Errorf("empty response")
	}
	if res.Policy.InternetFirewall.MoveRule.Status != "SUCCESS" {
		return formatIfSubPolicyErrors(res.Policy.InternetFirewall.MoveRule.Errors)
	}
	return nil
}

func (r *internetFwSubPolicyResource) updateSubPolicyScope(ctx context.Context, scopeRuleID string, scopeObj types.Object) error {
	scope := InternetFirewallSubPolicyScope{}
	diags := scopeObj.As(ctx, &scope, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return fmt.Errorf("%v", diags.Errors())
	}
	input := cato_models.InternetFirewallUpdateRuleInput{
		ID: scopeRuleID,
		Rule: &cato_models.InternetFirewallUpdateRuleDataInput{
			Name:    scope.Name.ValueStringPointer(),
			Enabled: scope.Enabled.ValueBoolPointer(),
		},
	}
	var res ifSubPolicyUpdateRuleResponse
	vars := map[string]any{
		"accountId": r.client.AccountId,
		"input":     input,
	}
	err := r.client.catov2.Client.Post(ctx, "policyInternetFirewallUpdateSubPolicyScope", ifSubPolicyUpdateRuleDocument, &res, vars)
	if err != nil {
		return err
	}
	if res.Policy == nil || res.Policy.InternetFirewall == nil || res.Policy.InternetFirewall.UpdateRule == nil {
		return fmt.Errorf("empty response")
	}
	if res.Policy.InternetFirewall.UpdateRule.Status != "SUCCESS" {
		return formatIfSubPolicyErrors(res.Policy.InternetFirewall.UpdateRule.Errors)
	}
	return nil
}

func (r *internetFwSubPolicyResource) publish(ctx context.Context) error {
	var res ifSubPolicyPublishResponse
	vars := map[string]any{
		"accountId": r.client.AccountId,
	}
	err := r.client.catov2.Client.Post(ctx, "policyInternetFirewallPublishSubPolicyRevision", ifSubPolicyPublishDocument, &res, vars)
	if err != nil {
		return err
	}
	if res.Policy == nil || res.Policy.InternetFirewall == nil || res.Policy.InternetFirewall.PublishPolicyRevision == nil {
		return fmt.Errorf("empty response")
	}
	if res.Policy.InternetFirewall.PublishPolicyRevision.Status != "SUCCESS" {
		return formatIfSubPolicyErrors(res.Policy.InternetFirewall.PublishPolicyRevision.Errors)
	}
	return nil
}

func findIfSubPolicyByName(state *ifSubPolicyPolicyState, name string) (*ifSubPolicyInfoState, error) {
	for _, subPolicy := range state.SubPolicies {
		if subPolicy.GetPolicy().Name == name {
			return subPolicy.GetPolicy(), nil
		}
	}
	return nil, fmt.Errorf("sub-policy %q not found", name)
}

func findIfSubPolicyByID(state *ifSubPolicyPolicyState, id string) (*ifSubPolicyInfoState, error) {
	for _, subPolicy := range state.SubPolicies {
		if subPolicy.GetPolicy().ID == id {
			return subPolicy.GetPolicy(), nil
		}
	}
	return nil, fmt.Errorf("sub-policy %q not found", id)
}

func findIfSubPolicyScopeRule(state *ifSubPolicyPolicyState, policyID string) (*ifSubPolicyRuleState, error) {
	var scopeRule *ifSubPolicyRuleState

	for _, rulePayload := range state.Rules {
		if rulePayload == nil || rulePayload.Rule == nil || rulePayload.SubPolicy == nil || rulePayload.SubPolicy.ID != policyID {
			continue
		}
		if rulePayload.RuleType == "SUB_POLICY_SCOPE" {
			scopeRule = rulePayload.Rule
		}
	}

	if scopeRule == nil {
		return nil, fmt.Errorf("scope rule for sub-policy %q not found", policyID)
	}
	return scopeRule, nil
}

func setIfSubPolicyState(ctx context.Context, state *InternetFirewallSubPolicy, policy *ifSubPolicyInfoState, scopeRule *ifSubPolicyRuleState) error {
	state.ID = types.StringValue(policy.ID)

	policyObj, diags := types.ObjectValue(
		ifSubPolicyPolicyAttrTypes,
		map[string]attr.Value{
			"id":          types.StringValue(policy.ID),
			"name":        types.StringValue(policy.Name),
			"description": types.StringValue(policy.Description),
		},
	)
	if diags.HasError() {
		return fmt.Errorf("%v", diags.Errors())
	}
	state.Policy = policyObj

	scopeObj, diags := types.ObjectValue(
		ifSubPolicyScopeAttrTypes,
		map[string]attr.Value{
			"id":      types.StringValue(scopeRule.ID),
			"name":    types.StringValue(scopeRule.Name),
			"enabled": types.BoolValue(scopeRule.Enabled),
		},
	)
	if diags.HasError() {
		return fmt.Errorf("%v", diags.Errors())
	}
	state.Scope = scopeObj

	if state.At.IsNull() || state.At.IsUnknown() {
		atObj, diags := types.ObjectValue(
			PositionAttrTypes,
			map[string]attr.Value{
				"position": types.StringValue("LAST_IN_POLICY"),
				"ref":      types.StringNull(),
			},
		)
		if diags.HasError() {
			return fmt.Errorf("%v", diags.Errors())
		}
		state.At = atObj
	}

	tflog.Debug(ctx, "internet firewall sub-policy state", map[string]interface{}{
		"sub_policy_id": policy.ID,
		"scope_rule_id": scopeRule.ID,
	})
	return nil
}

func addPolicyMutationErrors(resp *resource.CreateResponse, summary string, errors []*ifSubPolicyMutationError) {
	if len(errors) == 0 {
		resp.Diagnostics.AddError(summary, "unknown API error")
		return
	}
	for _, item := range errors {
		resp.Diagnostics.AddError(summary, formatIfSubPolicyError(item))
	}
}

func formatIfSubPolicyErrors(errors []*ifSubPolicyMutationError) error {
	if len(errors) == 0 {
		return fmt.Errorf("unknown API error")
	}
	return fmt.Errorf("%s", formatIfSubPolicyError(errors[0]))
}

func formatIfSubPolicyError(item *ifSubPolicyMutationError) string {
	if item == nil {
		return "unknown API error"
	}
	code := ""
	message := ""
	if item.ErrorCode != nil {
		code = *item.ErrorCode
	}
	if item.ErrorMessage != nil {
		message = *item.ErrorMessage
	}
	if code == "" {
		return message
	}
	if message == "" {
		return code
	}
	return fmt.Sprintf("%s : %s", code, message)
}

func (p *ifSubPolicyPayload) GetPolicy() *ifSubPolicyInfoState {
	if p == nil || p.Policy == nil {
		return &ifSubPolicyInfoState{}
	}
	return p.Policy
}

var ifSubPolicyPolicyAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
}

var ifSubPolicyScopeAttrTypes = map[string]attr.Type{
	"id":      types.StringType,
	"name":    types.StringType,
	"enabled": types.BoolType,
}
