package ccsection

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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

type placementClient struct {
	ResourceClient
	moves, updates int
}

func (c *placementClient) PolicyClientConnectivityMoveSection(_ context.Context, _ models.PolicyMoveSectionInput, _ string, _ ...clientv2.RequestInterceptor) (*cato.PolicyClientConnectivityMoveSection, error) {
	c.moves++
	return nil, errors.New("move reached")
}
func (c *placementClient) PolicyClientConnectivityUpdateSection(_ context.Context, _ models.PolicyUpdateSectionInput, _ string, _ ...clientv2.RequestInterceptor) (*cato.PolicyClientConnectivityUpdateSection, error) {
	c.updates++
	return nil, errors.New("update reached")
}
func TestUpdateMovesOnlyWhenPlacementChanges(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "content only", true: "placement changed"}[changed], func(t *testing.T) {
			ctx := context.Background()
			client := &placementClient{}
			r := &clientConnectivitySectionResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			at := types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{"position": types.StringValue("LAST_IN_POLICY"), "ref": types.StringNull()})
			current := ClientConnectivitySection{ID: types.StringValue("object-id"), At: at, Section: types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("object-id"), "name": types.StringValue("old-name")})}
			state := tfsdk.State{Schema: schemaResponse.Schema}
			require.False(t, state.Set(ctx, current).HasError())
			planned := current
			planned.Section = types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("object-id"), "name": types.StringValue("new-name")})
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
