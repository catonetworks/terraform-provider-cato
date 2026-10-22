package globalipranges

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

// callCreateRangesAPI calls the Create API and converts the response to the TF model
func (r *globalIPRangesResource) callCreateRangesAPI(ctx context.Context, input []*cato_models.CreateGlobalIPRangeInput,
	diags *diag.Diagnostics,
) (createdRanges []tf.GlobalIPRange) {
	result, err := r.resourceAPI().ObjectCreateGlobalIPRangeBulk(ctx, r.client.AccountId, input)
	if err != nil {
		diags.AddError("Error creating global IP ranges", err.Error())
		return nil
	}

	items := result.GetObject().GetCreateGlobalIPRangeBulk().GetGlobalIPRange()
	createdRanges = make([]tf.GlobalIPRange, 0, len(items))
	for _, item := range items {
		tfRange := tf.GlobalIPRange{
			Description: types.StringPointerValue(item.GetDescription()),
			ID:          types.StringValue(item.GetID()),
			IPRange:     types.StringValue(item.GetIPRange()),
			Name:        types.StringValue(item.GetName()),
		}
		createdRanges = append(createdRanges, tfRange)
	}
	return createdRanges
}

// callUpdateRangesAPI calls the Update API and converts the response to the TF model
func (r *globalIPRangesResource) callUpdateRangesAPI(ctx context.Context, input []*cato_models.UpdateGlobalIPRangeInput,
	diags *diag.Diagnostics,
) (updatedRanges []tf.GlobalIPRange) {
	result, err := r.resourceAPI().ObjectUpdateGlobalIPRangeBulk(ctx, r.client.AccountId, input)
	if err != nil {
		diags.AddError("Error updating global IP ranges", err.Error())
		return nil
	}

	items := result.GetObject().GetUpdateGlobalIPRangeBulk().GetGlobalIPRange()
	updatedRanges = make([]tf.GlobalIPRange, 0, len(items))
	for _, item := range items {
		tfRange := tf.GlobalIPRange{
			Description: types.StringPointerValue(item.GetDescription()),
			ID:          types.StringValue(item.GetID()),
			IPRange:     types.StringValue(item.GetIPRange()),
			Name:        types.StringValue(item.GetName()),
		}
		updatedRanges = append(updatedRanges, tfRange)
	}
	return updatedRanges
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
