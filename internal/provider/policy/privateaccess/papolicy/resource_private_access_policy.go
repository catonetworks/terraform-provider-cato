package papolicy

import (
	"context"

	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	apperror "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *privAccessPolicyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_private_access_policy"
}

func (r *privAccessPolicyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *privAccessPolicyResource) ImportState(
	ctx context.Context,
	_ resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), false)...)
}

// Create does not really create anything as the plolicy always exists.
// It updates the 'enabled' field of the policy.
func (r *privAccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, diags := r.callUpdate(ctx, plan.Enabled.ValueBool())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read the policy status
func (r *privAccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyModel

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

// Update the policy
func (r *privAccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, diags := r.callUpdate(ctx, plan.Enabled.ValueBool())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete does not do anything, the policy always exists
func (r *privAccessPolicyResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// hydratePrivAccessPolicyState fetches the current state of a privAccessPolicy from the API
func (r *privAccessPolicyResource) hydratePrivAccessPolicyState(ctx context.Context) (*PolicyModel, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	// Call Cato API to get the policy
	result, err := r.resourceAPI().PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "PolicyReadPrivateAccessPolicy", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return nil, nil, err
	}

	// Map API response to PrivAccessPolicyModel
	state := convertHydratePrivAccessPolicyStateValues(
		ctx,
		r,
		&diags,
		result,
	)
	if diags.HasError() {
		return nil, diags, apperror.ErrAPIResponseParse
	}

	return state, nil, nil
}
