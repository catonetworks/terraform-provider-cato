package sbrule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestReadReflectsRemoteExceptionChanges(t *testing.T) {
	for _, removed := range []bool{false, true} {
		t.Run(map[bool]string{false: "modified", true: "removed"}[removed], func(t *testing.T) {
			ctx := context.Background()
			backend := testBackendData(t)
			if removed {
				backend.object["exception"] = []any{}
			} else {
				backend.object["exception"].([]any)[0].(map[string]any)["name"] = "remote-edit"
			}
			r := testResource(t, backend)
			state := testState(t, testModel(t))
			response := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &response)
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			var actual resourceModel
			require.False(t, response.State.Get(ctx, &actual).HasError())
			exceptions := actual.Rule.Attributes()["exception"].(types.Set)
			if removed {
				require.Empty(t, exceptions.Elements())
			} else {
				require.Len(t, exceptions.Elements(), 1)
				require.Equal(t, types.StringValue("remote-edit"), exceptions.Elements()[0].(types.Object).Attributes()["name"])
			}
		})
	}
}

func TestImportPreservesExceptionsOnContentUpdate(t *testing.T) {
	ctx := context.Background()
	backend := testBackendData(t)
	r := testResource(t, backend)
	imported := resource.ImportStateResponse{State: tfsdk.State{Schema: testSchema(t)}}
	require.False(t, imported.State.Set(ctx, resourceModel{At: types.ObjectNull(atAttrTypes), Rule: types.ObjectNull(ruleAttrTypes)}).HasError())
	r.ImportState(ctx, resource.ImportStateRequest{ID: "item-1"}, &imported)
	require.False(t, imported.Diagnostics.HasError(), "%v", imported.Diagnostics)
	read := resource.ReadResponse{State: imported.State}
	r.Read(ctx, resource.ReadRequest{State: imported.State}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	var model resourceModel
	require.False(t, read.State.Get(ctx, &model).HasError())
	require.Len(t, model.Rule.Attributes()["exception"].(types.Set).Elements(), 1)
	attrs := model.Rule.Attributes()
	attrs["description"] = types.StringValue("changed description")
	model.Rule = types.ObjectValueMust(ruleAttrTypes, attrs)
	update := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: testPlan(t, model)}, &update)
	require.False(t, update.Diagnostics.HasError(), "%v", update.Diagnostics)
	for i, operation := range backend.calls {
		if operation != "updateRule" {
			continue
		}
		input := backend.variables[i]["socketBypassUpdateRuleInput"].(map[string]any)
		rule := input["rule"].(map[string]any)
		require.Len(t, rule["exception"].([]any), 1)
		require.Equal(t, true, rule["action"].(map[string]any)["event"].(map[string]any)["enabled"])
		require.Equal(t, "DNS_TCP", rule["service"].(map[string]any)["simple"].([]any)[0].(map[string]any)["name"])
		return
	}
	t.Fatal("content update did not reach API")
}
