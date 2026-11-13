package wfrule

import (
	"context"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestWanFwRuleCreateNameCollisionRequiresImport(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := mocks.NewWanFirewallRuleClient(t)
	code, message := "ruleNameAlreadyExists", "Rule with the same name already exists"
	result := &cato.PolicyWanFirewallAddRule{
		Policy: cato.PolicyWanFirewallAddRule_Policy{
			WanFirewall: &cato.PolicyWanFirewallAddRule_Policy_WanFirewall{
				AddRule: cato.PolicyWanFirewallAddRule_Policy_WanFirewall_AddRule{
					Status: "FAILURE",
					Errors: []*cato.PolicyWanFirewallAddRule_Policy_WanFirewall_AddRule_Errors{
						{ErrorCode: &code, ErrorMessage: &message},
					},
				},
			},
		},
	}
	// No lookup, update, or publish is permitted after a name collision.
	client.EXPECT().PolicyWanFirewallAddRule(mock.Anything, mock.Anything, "account-123").Return(result, nil).Once()
	r := &WanFwRuleResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: client}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	require.False(t, plan.Set(ctx, NewMinimalWanFwRuleModel("")).HasError())
	response := resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
	require.True(t, response.Diagnostics.HasError())
	require.True(t, strings.Contains(response.Diagnostics.Errors()[0].Detail(), "import the existing rule explicitly"))
}
