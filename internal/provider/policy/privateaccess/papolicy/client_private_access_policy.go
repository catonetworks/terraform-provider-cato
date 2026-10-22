package papolicy

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// callUpdate implements the actual update logic use by Update() or Create()
func (r *privAccessPolicyResource) callUpdate(
	ctx context.Context,
	isEnabled bool,
) (newState *PolicyModel, diags diag.Diagnostics) {
	// Set the enabled polEnabled
	polEnabled := cato_models.PolicyToggleStateDisabled
	if isEnabled {
		polEnabled = cato_models.PolicyToggleStateEnabled
	}
	input := cato_models.PrivateAccessPolicyUpdateInput{State: &polEnabled}

	// Call Cato API to update the policy
	tflog.Debug(ctx, "PolicyPrivateAccessUpdatePolicy", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.resourceAPI().PolicyPrivateAccessUpdatePolicy(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PolicyPrivateAccessUpdatePolicy", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	errMsg := "failed to update private access policy"
	if err != nil {
		diags.AddError(errMsg, err.Error())
		return nil, diags
	}

	pol := result.GetPolicy().GetPrivateAccess().GetUpdatePolicy()
	if pol == nil {
		diags.AddError(errMsg, "policy query returned nil")
		return nil, diags
	}
	if pol.Status != cato_models.PolicyMutationStatusSuccess {
		diags.AddError(errMsg, "returned status: "+string(pol.Status))
		for _, e := range pol.Errors {
			diags.AddError(errMsg, fmt.Sprintf("ERROR: %v [%v]", *e.GetErrorMessage(), *e.GetErrorCode()))
		}
		return nil, diags
	}

	// Hydrate state from API
	newState, hydrateDiags, hydrateErr := r.hydratePrivAccessPolicyState(ctx)
	if hydrateErr != nil {
		diags.AddError("Error hydrating privateAccessRule state", hydrateErr.Error())
		diags.Append(hydrateDiags...)
		return nil, diags
	}
	return newState, diags
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyPrivateAccessUpdatePolicy(
		ctx context.Context,
		accountID string,
		input cato_models.PrivateAccessPolicyUpdateInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessUpdatePolicy, error)
	PolicyReadPrivateAccessPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error)
}

func (r *privAccessPolicyResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyPrivateAccessUpdatePolicy(
	ctx context.Context,
	accountID string,
	input cato_models.PrivateAccessPolicyUpdateInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessUpdatePolicy, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessUpdatePolicy(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyReadPrivateAccessPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error) {
	return c.resource.client.Catov2.PolicyReadPrivateAccessPolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *privAccessPolicyResource }

var _ ResourceClient = (*sdkClient)(nil)
