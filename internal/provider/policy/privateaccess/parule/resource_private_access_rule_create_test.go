package parule

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// Exercise Create itself so a missing first read cannot leave a successfully
// created rule without Terraform state, even with the latest null-set schemas.
func TestPrivateAccessCreateWaitsForFirstRead(t *testing.T) {
	ctx := context.Background()
	reads := 0
	r := newPrivateAccessRuleTestResource(t, make(chan string, 3), func(w http.ResponseWriter, operation string) {
		switch operation {
		case "policyPrivateAccessAddRule":
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"addRule":{"status":"SUCCESS","rule":{"rule":{"id":"rule-123"}}}}}}}`))
		case "policyReadPrivateAccessPolicy":
			reads++
			items := ""
			if reads > 1 {
				items = `{"rule":{"id":"rule-123","name":"test rule","enabled":true,"action":{"action":"ALLOW"}}}`
			}
			_, _ = w.Write([]byte(`{"data":{"policy":{"privateAccess":{"policy":{"rules":[` + items + `]}}}}}`))
		default:
			t.Errorf("unexpected operation %s", operation)
		}
	})
	state := newPrivateAccessRuleTestState(ctx, t, r)
	plan := tfsdk.Plan(state)
	require.False(t, plan.SetAttribute(ctx, path.Root("id"), types.StringNull()).HasError())
	response := resource.CreateResponse{State: tfsdk.State{Schema: state.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	var created RuleModel
	require.False(t, response.State.Get(ctx, &created).HasError())
	require.Equal(t, "rule-123", created.ID.ValueString())
	require.Equal(t, 2, reads)
}
