package rule_test

import (
	"context"
	"testing"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resourcestest"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestRuleDeleteThroughProvider(t *testing.T) {
	t.Parallel()
	for _, rejectPublish := range []bool{false, true} {
		ctx := context.Background()
		var calls []string
		r, resourceSchema := resourcestest.ConfiguredResource(t, "cato_private_access_rule", func(request resourcestest.GraphQLRequest) string {
			calls = append(calls, request.OperationName)
			switch request.OperationName {
			case "policyPrivateAccessDeleteRule":
				require.JSONEq(t, `{"id":"rule-456"}`, string(request.Variables["input"]))
				return `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`
			case "policyPrivateAccessPublishRevision":
				if rejectPublish {
					return `{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"FAILURE","errors":[{"errorCode":"PermissionDenied","errorMessage":"sensitive backend details"}]}}}}}`
				}
				return `{"data":{"policy":{"privateAccess":{"publishPolicyRevision":{"status":"SUCCESS"}}}}}`
			case "refactorPrivateAccessPolicy":
				require.JSONEq(t, `"PUBLIC"`, string(request.Variables["revision"]))
				return `{"data":{"policy":{"privateAccess":{"policy":{"rules":[]}}}}}`
			default:
				t.Fatalf("unexpected operation: %s", request.OperationName)
				return ""
			}
		})
		state := tfsdk.State{Schema: resourceSchema}
		model := struct {
			ID types.String `tfsdk:"id"`
		}{ID: types.StringValue("rule-456")}
		require.False(t, state.Set(ctx, model).HasError())
		response := resource.DeleteResponse{State: state}
		r.Delete(ctx, resource.DeleteRequest{State: state}, &response)
		require.Equal(t, rejectPublish, response.Diagnostics.HasError())
		if rejectPublish {
			require.Equal(t, state.Raw, response.State.Raw)
			require.Len(t, calls, 2)
			require.NotContains(t, response.Diagnostics.Errors()[0].Detail(), "sensitive backend details")
		} else {
			require.Equal(t, []string{"policyPrivateAccessDeleteRule", "policyPrivateAccessPublishRevision", "refactorPrivateAccessPolicy"}, calls)
		}
	}
}
