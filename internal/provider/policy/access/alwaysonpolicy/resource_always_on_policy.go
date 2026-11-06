package alwaysonpolicy

import (
	"context"
	"errors"
	"fmt"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &alwaysOnPolicyResource{}
	_ resource.ResourceWithConfigure   = &alwaysOnPolicyResource{}
	_ resource.ResourceWithImportState = &alwaysOnPolicyResource{}
)

type alwaysOnPolicyResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewAlwaysOnPolicyResource returns the cato_always_on_policy resource.
func NewAlwaysOnPolicyResource() resource.Resource { return &alwaysOnPolicyResource{} }

func (r *alwaysOnPolicyResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return r.client.Catov2
}

func (r *alwaysOnPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_always_on_policy"
}

func (r *alwaysOnPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *alwaysOnPolicyResource) ImportState(ctx context.Context, _ resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	state := r.readPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alwaysOnPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.updatePolicy(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	r.publishPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	state := r.readPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alwaysOnPolicyResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state := r.readPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alwaysOnPolicyResource) readPolicy(ctx context.Context, diags *diag.Diagnostics) *PolicyModel {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("failed to read Always On policy", "An API client and account ID are required")
		return nil
	}
	var result *cato.PolicyZtnaAlwaysOnPolicy
	err := policy.WithTransientReadRetry(ctx, "Read Always On policy", func() error {
		var callErr error
		result, callErr = r.getClient().PolicyZtnaAlwaysOnPolicy(ctx, r.client.AccountId, nil)
		return callErr
	})
	if err != nil {
		diags.AddError("failed to read Always On policy", err.Error())
		return nil
	}
	if result.GetPolicy().GetZtnaAlwaysOn() == nil {
		diags.AddError("failed to read Always On policy", "API did not return a policy")
		return nil
	}
	return parsePolicy(result)
}

func (r *alwaysOnPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.updatePolicy(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	r.publishPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	state := r.readPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alwaysOnPolicyResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.publishPolicy(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *alwaysOnPolicyResource) updatePolicy(ctx context.Context, plan PolicyModel, diags *diag.Diagnostics) {
	input := buildPolicyInput(plan, diags)
	if diags.HasError() {
		return
	}
	errMsg := "failed to update Always On policy"
	result, err := r.getClient().PolicyZtnaAlwaysOnUpdatePolicy(ctx, r.client.AccountId, nil, input)
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetUpdatePolicy()
	if utils.CheckAPIErrors(err, payload.GetErrors(), errMsg, diags) {
		return
	}
	status := payload.GetStatus()
	if status == nil || *status != cato_models.PolicyMutationStatusSuccess {
		value := ""
		if status != nil {
			value = string(*status)
		}
		diags.AddError(errMsg, fmt.Sprintf("API returned mutation status %q", value))
		return
	}

	state := parsePolicyUpdate(payload)
	if !state.Enabled.Equal(plan.Enabled) {
		diags.AddError(errMsg, "API reported success but the policy enabled value did not match the requested value on read.")
		return
	}
}

func (r *alwaysOnPolicyResource) publishPolicy(ctx context.Context, diags *diag.Diagnostics) {
	err := policy.WithPolicyRevisionConflictRetry(ctx, "Publish Always On policy", func() error {
		var attempt diag.Diagnostics
		r.publishPolicyOnce(ctx, &attempt)
		var failures []error
		for _, diagnostic := range attempt.Errors() {
			failures = append(failures, errors.New(diagnostic.Detail()))
		}
		return errors.Join(failures...)
	})
	if err != nil {
		diags.AddError("failed to publish Always On policy", err.Error())
	}
}

func (r *alwaysOnPolicyResource) publishPolicyOnce(ctx context.Context, diags *diag.Diagnostics) {
	const summary = "failed to publish Always On policy"
	result, err := r.getClient().PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, r.client.AccountId, nil, nil)
	if err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	payload := result.GetPolicy().GetZtnaAlwaysOn().GetPublishPolicyRevision()
	if apiErrors := payload.GetErrors(); len(apiErrors) == 1 {
		if apiErrors[0].ErrorCode != nil && *apiErrors[0].ErrorCode == apperror.PolicyRevisionNotFound {
			return // no policy draft to discard; OK
		}
	}
	if utils.CheckAPIErrors(err, payload.GetErrors(), summary, diags) {
		return
	}
	status := payload.GetStatus()
	if status == nil || *status != cato_models.PolicyMutationStatusSuccess {
		value := ""
		if status != nil {
			value = string(*status)
		}
		diags.AddError(summary, fmt.Sprintf("API returned mutation status %q", value))
	}
}
