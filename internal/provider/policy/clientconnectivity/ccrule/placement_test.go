package ccrule

import (
	"context"
	"errors"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

type placementClient struct {
	ResourceClient
	moves, updates int
}

func (c *placementClient) PolicyClientConnectivityMoveRule(_ context.Context, _ models.PolicyMoveRuleInput, _ string, _ ...clientv2.RequestInterceptor) (*cato.PolicyClientConnectivityMoveRule, error) {
	c.moves++
	return nil, errors.New("move reached")
}
func (c *placementClient) PolicyClientConnectivityUpdateRule(_ context.Context, _ models.ClientConnectivityUpdateRuleInput, _ string, _ ...clientv2.RequestInterceptor) (*cato.PolicyClientConnectivityUpdateRule, error) {
	c.updates++
	return nil, errors.New("update reached")
}
func TestUpdateMovesOnlyWhenPlacementChanges(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "content only", true: "placement changed"}[changed], func(t *testing.T) {
			ctx := context.Background()
			client := &placementClient{}
			r := &clientConnectivityRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			at := types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{"position": types.StringValue("LAST_IN_POLICY"), "ref": types.StringNull()})
			hydrated, diags := hydrateClientConnectivityRuleStateFromClient(ctx, &cato.PolicyClientConnectivityPolicy_Policy_ClientConnectivity_Policy_Rules_Rule{ID: "object-id", Name: "old-name"})
			require.False(t, diags.HasError(), "%v", diags)
			object, diags := types.ObjectValueFrom(ctx, ClientConnectivityRuleRuleAttrTypes, hydrated)
			require.False(t, diags.HasError(), "%v", diags)
			current := ClientConnectivityRule{ID: types.StringValue("object-id"), At: at, Rule: object}
			state := tfsdk.State{Schema: schemaResponse.Schema}
			require.False(t, state.Set(ctx, current).HasError())
			planned := current
			attrs := planned.Rule.Attributes()
			attrs["description"] = types.StringValue("new-description")
			planned.Rule = types.ObjectValueMust(ClientConnectivityRuleRuleAttrTypes, attrs)
			if changed {
				planned.At = types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{"position": types.StringValue("FIRST_IN_POLICY"), "ref": types.StringNull()})
			}
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			require.False(t, plan.Set(ctx, planned).HasError())
			response := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, &response)
			require.True(t, response.Diagnostics.HasError(), "stop after the first mutation")
			if changed {
				require.Equal(t, 1, client.moves)
				require.Zero(t, client.updates)
			} else {
				require.Zero(t, client.moves, "content edits must preserve bulk ordering")
				require.Equal(t, 1, client.updates)
			}
		})
	}
}
