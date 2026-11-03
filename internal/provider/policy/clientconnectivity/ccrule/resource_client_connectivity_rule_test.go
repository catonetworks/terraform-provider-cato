package ccrule

import (
	"context"
	"encoding/json"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

type rejectedDeleteClient struct {
	ResourceClient
	response *cato_go_sdk.PolicyClientConnectivityRemoveRule
}

func (c *rejectedDeleteClient) PolicyClientConnectivityRemoveRule(
	_ context.Context, _ cato_models.ClientConnectivityRemoveRuleInput, _ string,
	_ ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyClientConnectivityRemoveRule, error) {
	return c.response, nil
}

func TestDeletePreservesStateOnPayloadFailure(t *testing.T) {
	for _, payload := range []string{`{}`, `{"status":"FAILURE","errors":[]}`, `{"status":"SUCCESS","errors":[{"errorMessage":"denied"}]}`} {
		t.Run(payload, func(t *testing.T) {
			ctx := context.Background()
			var response cato_go_sdk.PolicyClientConnectivityRemoveRule
			require.NoError(t, json.Unmarshal([]byte(`{"policy":{"clientConnectivity":{"removeRule":`+payload+`}}}`), &response))
			r := &clientConnectivityRuleResource{client: &catoclient.Service{AccountId: "account"}, apiClient: &rejectedDeleteClient{response: &response}}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			current, diags := hydrateClientConnectivityRuleStateFromClient(ctx, &cato_go_sdk.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Rules_Rule{ID: "rule"})
			require.False(t, diags.HasError())
			obj, diags := types.ObjectValueFrom(ctx, ClientConnectivityRuleRuleAttrTypes, current)
			require.False(t, diags.HasError())
			state := tfsdk.State{Schema: schemaResponse.Schema}
			diags = state.Set(ctx, ClientConnectivityRule{ID: types.StringValue("rule"), At: types.ObjectNull(policyrule.PositionAttrTypes), Rule: obj})
			require.False(t, diags.HasError(), "%v", diags)
			result := resource.DeleteResponse{State: state}
			r.Delete(ctx, resource.DeleteRequest{State: state}, &result)
			require.True(t, result.Diagnostics.HasError())
			require.True(t, state.Raw.Equal(result.State.Raw))
		})
	}
}
