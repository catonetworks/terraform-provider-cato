package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &privAccessPolicyResource{}
	_ resource.ResourceWithConfigure   = &privAccessPolicyResource{}
	_ resource.ResourceWithImportState = &privAccessPolicyResource{}
)

func NewPrivAccessPolicyResource() resource.Resource {
	return &privAccessPolicyResource{}
}

type privAccessPolicyResource struct {
	client *catoClientData
}

func (r *privAccessPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_access_policy"
}

func (r *privAccessPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_private_access_policy` resource contains the configuration parameters for private access policies in the Cato platform.",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Is the private access policy enabled?",
				Required:    true,
			},
			"audit": schema.SingleNestedAttribute{
				Description:   "Audit log record",
				Computed:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"published_time": schema.StringAttribute{
						Description: "Audit log published time",
						Computed:    true,
					},
					"published_by": schema.StringAttribute{
						Description: "Author of the published log record",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (r *privAccessPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *privAccessPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	// TODO: implement
}

func (r *privAccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PrivAccessPolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, diags := r.callUpdate(ctx, plan.Enabled.ValueBool())
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PrivAccessPolicyModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, diags := r.callUpdate(ctx, plan.Enabled.ValueBool())
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessPolicyResource) callUpdate(ctx context.Context, isEnabled bool) (newState *PrivAccessPolicyModel, diags diag.Diagnostics) {
	// Set the enabled polEnabled
	polEnabled := cato_models.PolicyToggleStateDisabled
	if isEnabled {
		polEnabled = cato_models.PolicyToggleStateEnabled
	}
	input := cato_models.PrivateAccessPolicyUpdateInput{State: &polEnabled}

	// Call Cato API to update the policy
	tflog.Debug(ctx, "PolicyPrivateAccessUpdatePolicy", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.catov2.PolicyPrivateAccessUpdatePolicy(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PolicyPrivateAccessUpdatePolicy", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	errMsg := "failed to update private access policy"
	if err != nil {
		diags.AddError(errMsg, err.Error())
		return nil, diags
	}
	pol := result.GetPolicy().GetPrivateAccess().GetUpdatePolicy()
	if pol.Status != cato_models.PolicyMutationStatusSuccess {
		diags.AddError(errMsg, "returned status: "+string(pol.Status))
		for _, e := range pol.Errors {
			diags.AddError(errMsg, fmt.Sprintf("ERROR: %v [%v]", *e.GetErrorMessage(), *e.GetErrorCode()))
		}
		return nil, diags
	}

	// Hydrate state from API
	newState, diag, hydrateErr := r.hydratePrivAccessPolicyState(ctx)
	if hydrateErr != nil {
		diags.AddError("Error hydrating privateAccessRule state", hydrateErr.Error())
		diags.Append(diag...)
		return nil, diags
	}
	return newState, diags
}

func (r *privAccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PrivAccessPolicyModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, diags, hydrateErr := r.hydratePrivAccessPolicyState(ctx)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privatea access policy state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

// hydratePrivAccessPolicyState fetches the current state of a privAccessPolicy from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privAccessPolicyResource) hydratePrivAccessPolicyState(ctx context.Context) (*PrivAccessPolicyModel, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	// Call Cato API to get the policy
	result, err := r.client.catov2.PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "PolicyReadPrivateAccessPolicy", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return nil, nil, err
	}

	// Map API response to PrivAccessPolicyModel
	policy := result.GetPolicy().GetPrivateAccess().GetPolicy()
	state := &PrivAccessPolicyModel{
		Enabled: types.BoolValue(policy.Enabled),
		Audit:   r.parseAudit(ctx, policy.Audit, &diags),
	}
	if diags.HasError() {
		return nil, diags, ErrAPIResponseParse
	}

	return state, nil, nil
}

func (r *privAccessPolicyResource) parseAudit(ctx context.Context, aud *cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Audit,
	diags *diag.Diagnostics,
) types.Object {
	var diag diag.Diagnostics

	// Prepare PolicyAudit object
	var auditObj types.Object = types.ObjectNull(PolicyAuditTypes)
	if aud != nil {
		tfAudit := PolicyAudit{
			PublishedBy:   types.StringValue(aud.PublishedBy),
			PublishedTime: types.StringValue(aud.PublishedTime),
		}
		auditObj, diag = types.ObjectValueFrom(ctx, PolicyAuditTypes, tfAudit)
		diags.Append(diag...)
		if diags.HasError() {
			return auditObj
		}
	}

	return auditObj
}
