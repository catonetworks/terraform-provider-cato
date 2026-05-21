package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

type internetFwSubPolicyResource struct {
	client *catoClientData
}

type internetFwSubPolicyModel struct {
	At     types.Object `tfsdk:"at"`
	Policy types.Object `tfsdk:"policy"`
}

type internetFwSubPolicyAtModel struct {
	Position types.String `tfsdk:"position"`
	Ref      types.String `tfsdk:"ref"`
}

type internetFwSubPolicyPolicyModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewInternetFwSubPolicyResource() resource.Resource {
	return &internetFwSubPolicyResource{}
}

func (r *internetFwSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_sub_policy"
}

func (r *internetFwSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *internetFwSubPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Internet Firewall sub-policy container. This resource only manages the sub-policy object. Sub-rules are managed via separate IF rule resources that reference this sub-policy.",
		Attributes: map[string]schema.Attribute{
			"at": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Required: true,
					},
					"ref": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"policy": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"description": schema.StringAttribute{
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

func (r *internetFwSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan internetFwSubPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	at := internetFwSubPolicyAtModel{}
	resp.Diagnostics.Append(plan.At.As(ctx, &at, basetypes.ObjectAsOptions{})...)
	policy := internetFwSubPolicyPolicyModel{}
	resp.Diagnostics.Append(plan.Policy.As(ctx, &policy, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	createResult, err := r.addInternetFirewallSubPolicy(ctx, at, policy)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API addSubPolicy error", err.Error())
		return
	}
	_ = createResult

	readResult, err := r.readInternetFirewallSubPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	created, ok := findIfwSubPolicyByName(readResult.Policy.InternetFirewall.Policy.SubPolicies, policy.Name.ValueString())
	if !ok {
		resp.Diagnostics.AddError("Catov2 API addSubPolicy error", "sub-policy not returned in API response")
		return
	}
	policy.ID = types.StringValue(created.Policy.ID)
	policy.Description = types.StringValue(created.Policy.Description)
	policyObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
	}, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Policy = policyObj
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *internetFwSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state internetFwSubPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := internetFwSubPolicyPolicyModel{}
	resp.Diagnostics.Append(state.Policy.As(ctx, &policy, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	readResult, err := r.readInternetFirewallSubPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Catov2 API PolicyInternetFirewall error", err.Error())
		return
	}
	found, ok := findIfwSubPolicyByID(readResult.Policy.InternetFirewall.Policy.SubPolicies, policy.ID.ValueString())
	if !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	policy.Name = types.StringValue(found.Policy.Name)
	policy.Description = types.StringValue(found.Policy.Description)
	policyObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
	}, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Policy = policyObj
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *internetFwSubPolicyResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Internet Firewall sub-policy attributes require replacement. Terraform should replace this resource when policy attributes change.",
	)
}

func (r *internetFwSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state internetFwSubPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policy := internetFwSubPolicyPolicyModel{}
	resp.Diagnostics.Append(state.Policy.As(ctx, &policy, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	if policy.ID.IsNull() || policy.ID.IsUnknown() {
		return
	}
	if err := r.removeInternetFirewallSubPolicy(ctx, policy.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Catov2 API removeSubPolicy error", err.Error())
	}
}

func (r *internetFwSubPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("policy").AtName("id"), req, resp)
}

type ifwAddSubPolicyResponse struct {
	Policy struct {
		InternetFirewall struct {
			AddSubPolicy struct {
				Status string `json:"status"`
				Errors []struct {
					ErrorMessage *string `json:"errorMessage"`
				} `json:"errors"`
			} `json:"addSubPolicy"`
		} `json:"internetFirewall"`
	} `json:"policy"`
}

type ifwReadSubPoliciesResponse struct {
	Policy struct {
		InternetFirewall struct {
			Policy struct {
				SubPolicies []ifwSubPolicyPayload `json:"subPolicies"`
			} `json:"policy"`
		} `json:"internetFirewall"`
	} `json:"policy"`
}

type ifwRemoveSubPolicyResponse struct {
	Policy struct {
		InternetFirewall struct {
			RemoveSubPolicy struct {
				Status string `json:"status"`
				Errors []struct {
					ErrorMessage *string `json:"errorMessage"`
				} `json:"errors"`
			} `json:"removeSubPolicy"`
		} `json:"internetFirewall"`
	} `json:"policy"`
}

type ifwSubPolicyPayload struct {
	Policy struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"policy"`
}

const mutationIfwAddSubPolicy = `mutation policyInternetFirewallAddSubPolicy($internetFirewallPolicyMutationInput: InternetFirewallPolicyMutationInput, $input: InternetFirewallAddSubPolicyInput!, $accountId: ID!) {
  policy(accountId: $accountId) {
    internetFirewall(input: $internetFirewallPolicyMutationInput) {
      addSubPolicy(input: $input) {
        status
        errors {
          errorMessage
        }
      }
    }
  }
}`

const queryIfwSubPolicies = `query internetFirewallSubPolicies($accountId: ID!) {
  policy(accountId: $accountId) {
    internetFirewall {
      policy {
        subPolicies {
          policy {
            id
            name
            description
          }
        }
      }
    }
  }
}`

const mutationIfwRemoveSubPolicy = `mutation policyInternetFirewallRemoveSubPolicy($internetFirewallPolicyMutationInput: InternetFirewallPolicyMutationInput, $input: InternetFirewallRemoveSubPolicyInput!, $accountId: ID!) {
  policy(accountId: $accountId) {
    internetFirewall(input: $internetFirewallPolicyMutationInput) {
      removeSubPolicy(input: $input) {
        status
        errors {
          errorMessage
        }
      }
    }
  }
}`

func (r *internetFwSubPolicyResource) addInternetFirewallSubPolicy(ctx context.Context, at internetFwSubPolicyAtModel, policy internetFwSubPolicyPolicyModel) (*ifwAddSubPolicyResponse, error) {
	scopeName := policy.Name.ValueString() + " scope"
	vars := map[string]any{
		"accountId":                           r.client.AccountId,
		"internetFirewallPolicyMutationInput": map[string]any{},
		"input": map[string]any{
			"at": map[string]any{
				"position": at.Position.ValueString(),
				"ref":      stringOrNil(at.Ref),
			},
			"policy": map[string]any{
				"name":        policy.Name.ValueString(),
				"description": policy.Description.ValueString(),
			},
			"scope": map[string]any{
				"name":                   scopeName,
				"enabled":                true,
				"action":                 "BLOCK",
				"connectionOrigin":       "ANY",
				"connectionsOriginList":  []string{"SITE"},
				"connectionsOriginsList": []string{},
				"country":                []any{},
				"device":                 []any{},
				"deviceOS":               []string{},
				"exceptions":             []any{},
				"source": map[string]any{
					"ip":                []string{},
					"host":              []any{},
					"site":              []any{},
					"subnet":            []string{},
					"ipRange":           []any{},
					"globalIpRange":     []any{},
					"networkInterface":  []any{},
					"siteNetworkSubnet": []any{},
					"floatingSubnet":    []any{},
					"user":              []any{},
					"usersGroup":        []any{},
					"group":             []any{},
					"systemGroup":       []any{},
				},
				"destination": map[string]any{
					"application":            []any{},
					"customApp":              []any{},
					"appCategory":            []any{},
					"customCategory":         []any{},
					"sanctionedAppsCategory": []any{},
					"country":                []any{},
					"domain":                 []string{},
					"fqdn":                   []string{},
					"ip":                     []string{},
					"subnet":                 []string{},
					"ipRange":                []any{},
					"globalIpRange":          []any{},
					"remoteAsn":              []string{},
					"group":                  []any{},
					"containers": map[string]any{
						"fqdnContainer":           []any{},
						"ipAddressRangeContainer": []any{},
					},
				},
				"service": map[string]any{
					"standard": []any{},
					"custom":   []any{},
				},
				"tracking": map[string]any{
					"event": map[string]any{
						"enabled": false,
					},
					"alert": map[string]any{
						"enabled":           false,
						"frequency":         "HOURLY",
						"subscriptionGroup": []any{},
						"webhook":           []any{},
						"mailingList":       []any{},
					},
				},
				"activePeriod": map[string]any{
					"useEffectiveFrom": false,
					"useExpiresAt":     false,
				},
				"schedule": map[string]any{
					"activeOn": "ALWAYS",
				},
				"deviceAttributes": map[string]any{
					"category":     []string{},
					"type":         []string{},
					"model":        []string{},
					"manufacturer": []string{},
					"os":           []string{},
					"osVersion":    []string{},
				},
				"userAttributes": map[string]any{
					"riskScore": map[string]any{
						"category": "ANY",
						"operator": "GTE",
					},
					"userConfidenceLevel": "ANY",
				},
			},
		},
	}
	var res ifwAddSubPolicyResponse
	if err := r.client.catov2.Client.Post(ctx, "policyInternetFirewallAddSubPolicy", mutationIfwAddSubPolicy, &res, vars); err != nil {
		return nil, err
	}
	tflog.Debug(ctx, "Write.PolicyInternetFirewallAddSubPolicy.response", map[string]any{"response": utils.InterfaceToJSONString(res)})
	if res.Policy.InternetFirewall.AddSubPolicy.Status != ifwMutationStatusSuccess {
		if len(res.Policy.InternetFirewall.AddSubPolicy.Errors) > 0 && res.Policy.InternetFirewall.AddSubPolicy.Errors[0].ErrorMessage != nil {
			return nil, errors.New(*res.Policy.InternetFirewall.AddSubPolicy.Errors[0].ErrorMessage)
		}
		return nil, fmt.Errorf("add sub-policy failed")
	}
	return &res, nil
}

func (r *internetFwSubPolicyResource) readInternetFirewallSubPolicies(ctx context.Context) (*ifwReadSubPoliciesResponse, error) {
	vars := map[string]any{"accountId": r.client.AccountId}
	var res ifwReadSubPoliciesResponse
	if err := r.client.catov2.Client.Post(ctx, "internetFirewallSubPolicies", queryIfwSubPolicies, &res, vars); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *internetFwSubPolicyResource) removeInternetFirewallSubPolicy(ctx context.Context, id string) error {
	vars := map[string]any{
		"accountId":                           r.client.AccountId,
		"internetFirewallPolicyMutationInput": map[string]any{},
		"input": map[string]any{
			"ref": map[string]any{
				"by":    "ID",
				"input": id,
			},
		},
	}
	var res ifwRemoveSubPolicyResponse
	if err := r.client.catov2.Client.Post(ctx, "policyInternetFirewallRemoveSubPolicy", mutationIfwRemoveSubPolicy, &res, vars); err != nil {
		return err
	}
	if res.Policy.InternetFirewall.RemoveSubPolicy.Status != ifwMutationStatusSuccess {
		if len(res.Policy.InternetFirewall.RemoveSubPolicy.Errors) > 0 && res.Policy.InternetFirewall.RemoveSubPolicy.Errors[0].ErrorMessage != nil {
			return errors.New(*res.Policy.InternetFirewall.RemoveSubPolicy.Errors[0].ErrorMessage)
		}
		return fmt.Errorf("remove sub-policy failed")
	}
	return nil
}

func findIfwSubPolicyByID(items []ifwSubPolicyPayload, id string) (ifwSubPolicyPayload, bool) {
	for _, item := range items {
		if item.Policy.ID == id {
			return item, true
		}
	}
	return ifwSubPolicyPayload{}, false
}

func findIfwSubPolicyByName(items []ifwSubPolicyPayload, name string) (ifwSubPolicyPayload, bool) {
	for _, item := range items {
		if item.Policy.Name == name {
			return item, true
		}
	}
	return ifwSubPolicyPayload{}, false
}

func stringOrNil(v types.String) any {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueString()
}
