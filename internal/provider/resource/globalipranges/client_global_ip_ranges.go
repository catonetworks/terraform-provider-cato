package globalipranges

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// callCreateRangesAPI forwards the SDK request without converting Terraform values.
func (r *globalIPRangesResource) callCreateRangesAPI(
	ctx context.Context, input []*cato_models.CreateGlobalIPRangeInput,
) (*cato_go_sdk.ObjectCreateGlobalIPRangeBulk, error) {
	return r.resourceAPI().ObjectCreateGlobalIPRangeBulk(ctx, r.client.AccountId, input)
}

// callUpdateRangesAPI forwards the SDK request without converting Terraform values.
func (r *globalIPRangesResource) callUpdateRangesAPI(
	ctx context.Context, input []*cato_models.UpdateGlobalIPRangeInput,
) (*cato_go_sdk.ObjectUpdateGlobalIPRangeBulk, error) {
	return r.resourceAPI().ObjectUpdateGlobalIPRangeBulk(ctx, r.client.AccountId, input)
}

// callDeleteRangesAPI calls the Delete API
func (r *globalIPRangesResource) callDeleteRangesAPI(ctx context.Context, input []*cato_models.GlobalIPRangeRefInput,
	diags *diag.Diagnostics,
) {
	_, err := r.resourceAPI().ObjectDeleteGlobalIPRangeBulk(ctx, r.client.AccountId, input)
	if err != nil {
		diags.AddError("Error deleting global IP ranges", err.Error())
		return
	}
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	ObjectCreateGlobalIPRangeBulk(
		ctx context.Context,
		accountID string,
		input []*cato_models.CreateGlobalIPRangeInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ObjectCreateGlobalIPRangeBulk, error)
	ObjectDeleteGlobalIPRangeBulk(
		ctx context.Context,
		accountID string,
		input []*cato_models.GlobalIPRangeRefInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ObjectDeleteGlobalIPRangeBulk, error)
	ObjectGlobalIPRangeList(
		ctx context.Context,
		accountID string,
		input *cato_models.GlobalIPRangeListInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ObjectGlobalIPRangeList, error)
	ObjectUpdateGlobalIPRangeBulk(
		ctx context.Context,
		accountID string,
		input []*cato_models.UpdateGlobalIPRangeInput,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ObjectUpdateGlobalIPRangeBulk, error)
}

func (r *globalIPRangesResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) ObjectCreateGlobalIPRangeBulk(
	ctx context.Context,
	accountID string,
	input []*cato_models.CreateGlobalIPRangeInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ObjectCreateGlobalIPRangeBulk, error) {
	return c.resource.client.Catov2.ObjectCreateGlobalIPRangeBulk(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) ObjectDeleteGlobalIPRangeBulk(
	ctx context.Context,
	accountID string,
	input []*cato_models.GlobalIPRangeRefInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ObjectDeleteGlobalIPRangeBulk, error) {
	return c.resource.client.Catov2.ObjectDeleteGlobalIPRangeBulk(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) ObjectGlobalIPRangeList(
	ctx context.Context,
	accountID string,
	input *cato_models.GlobalIPRangeListInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ObjectGlobalIPRangeList, error) {
	return c.resource.client.Catov2.ObjectGlobalIPRangeList(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) ObjectUpdateGlobalIPRangeBulk(
	ctx context.Context,
	accountID string,
	input []*cato_models.UpdateGlobalIPRangeInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ObjectUpdateGlobalIPRangeBulk, error) {
	return c.resource.client.Catov2.ObjectUpdateGlobalIPRangeBulk(ctx, accountID, input, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct{ resource *globalIPRangesResource }

var _ ResourceClient = (*sdkClient)(nil)
