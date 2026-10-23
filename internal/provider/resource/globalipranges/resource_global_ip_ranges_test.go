package globalipranges

import (
	"context"
	"encoding/json"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

func TestGlobalIPRangesWritesReportConversionErrors(t *testing.T) {
	ctx := context.Background()
	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			api := mocks.NewGlobalIPRangesResourceClient(t)
			r := &globalIPRangesResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
			schemaResp := &resource.SchemaResponse{}
			r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
			rangeObject, diags := types.ObjectValueFrom(ctx, tf.GlobalIPRangeTypes, tf.GlobalIPRange{
				ID: types.StringValue("range-1"), Name: types.StringValue("office"),
				IPRange: types.StringValue("192.0.2.0/24"), Description: types.StringValue(""),
			})
			require.False(t, diags.HasError())
			model := tf.GlobalIPRangesModel{Ranges: types.SetValueMust(types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes}, []attr.Value{rangeObject})}
			plan := tfsdk.Plan{Schema: schemaResp.Schema}
			require.False(t, plan.Set(ctx, model).HasError())
			state := tfsdk.State{Schema: schemaResp.Schema}
			require.False(t, state.Set(ctx, model).HasError())
			updatedObject, updateDiags := types.ObjectValueFrom(ctx, tf.GlobalIPRangeTypes, tf.GlobalIPRange{
				ID: types.StringValue("range-1"), Name: types.StringValue("office"),
				IPRange: types.StringValue("192.0.2.0/25"), Description: types.StringValue(""),
			})
			require.False(t, updateDiags.HasError())
			updated := tfsdk.State{Schema: schemaResp.Schema}
			require.False(t, updated.Set(ctx, tf.GlobalIPRangesModel{Ranges: types.SetValueMust(types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes}, []attr.Value{updatedObject})}).HasError())
			config := tfsdk.Config{Schema: schemaResp.Schema, Raw: updated.Raw}

			// The schema stays valid for request decoding; the converter's type map is
			// deliberately incompatible so the post-API state conversion fails.
			original := tf.GlobalIPRangeTypes
			t.Cleanup(func() { tf.GlobalIPRangeTypes = original })
			if operation == "create" {
				var result cato.ObjectCreateGlobalIPRangeBulk
				require.NoError(t, json.Unmarshal([]byte(`{"object":{"createGlobalIPRangeBulk":{"globalIPRange":[{"id":"range-1","name":"office","ipRange":"192.0.2.0/24"}]}}}`), &result))
				api.EXPECT().ObjectCreateGlobalIPRangeBulk(mock.Anything, "account-123", mock.Anything).
					Run(func(context.Context, string, []*cato_models.CreateGlobalIPRangeInput, ...clientv2.RequestInterceptor) {
						tf.GlobalIPRangeTypes = map[string]attr.Type{"name": types.BoolType}
					}).Return(&result, nil).Once()
				response := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
				r.Create(ctx, resource.CreateRequest{Plan: plan}, response)
				require.True(t, response.Diagnostics.HasError(), "Terraform must receive the conversion error")
				require.True(t, response.State.Raw.IsNull(), "failed conversion must not persist state")
			} else {
				var result cato.ObjectUpdateGlobalIPRangeBulk
				require.NoError(t, json.Unmarshal([]byte(`{"object":{"updateGlobalIPRangeBulk":{"globalIPRange":[{"id":"range-1","name":"office","ipRange":"192.0.2.0/25"}]}}}`), &result))
				api.EXPECT().ObjectUpdateGlobalIPRangeBulk(mock.Anything, "account-123", mock.Anything).
					Run(func(context.Context, string, []*cato_models.UpdateGlobalIPRangeInput, ...clientv2.RequestInterceptor) {
						tf.GlobalIPRangeTypes = map[string]attr.Type{"name": types.BoolType}
					}).Return(&result, nil).Once()
				response := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
				r.Update(ctx, resource.UpdateRequest{Config: config, State: state}, response)
				require.True(t, response.Diagnostics.HasError(), "Terraform must receive the conversion error")
				require.True(t, response.State.Raw.IsNull(), "failed conversion must not persist state")
			}
		})
	}
}
