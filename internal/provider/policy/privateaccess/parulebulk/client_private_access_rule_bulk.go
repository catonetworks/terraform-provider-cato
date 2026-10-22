package parulebulk

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// publish calls the API to publish the draft policy revision
func (r *privAccessRuleBulkResource) publish(ctx context.Context) error {
	result, err := r.resourceAPI().PolicyPrivateAccessPublishRevision(ctx, r.client.AccountId)
	tflog.Debug(ctx, "Bulk PolicyPrivateAccessPublishRevision", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return err
	}
	res := result.GetPolicy().GetPrivateAccess().GetPublishPolicyRevision()
	if *res.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		apiErrors := res.GetErrors()
		if len(apiErrors) > 0 {
			if *apiErrors[0].GetErrorCode() == "PolicyRevisionNotFound" {
				return nil // there was nothing to publish
			}
			return fmt.Errorf("error publishing policy - %s", *apiErrors[0].GetErrorMessage())
		}
		return fmt.Errorf("error publishing policy")
	}
	return nil
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	PolicyPrivateAccessMoveRule(
		ctx context.Context,
		accountID string,
		input cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessMoveRule, error)
	PolicyPrivateAccessPublishRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyPrivateAccessPublishRevision, error)
	PolicyReadPrivateAccessPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error)
}

func (r *privAccessRuleBulkResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) PolicyPrivateAccessMoveRule(
	ctx context.Context,
	accountID string,
	input cato_models.PolicyMoveRuleInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessMoveRule, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessMoveRule(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyPrivateAccessPublishRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyPrivateAccessPublishRevision, error) {
	return c.resource.client.Catov2.PolicyPrivateAccessPublishRevision(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyReadPrivateAccessPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyReadPrivateAccessPolicy, error) {
	return c.resource.client.Catov2.PolicyReadPrivateAccessPolicy(ctx, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *privAccessRuleBulkResource }

var _ ResourceClient = (*sdkClient)(nil)
