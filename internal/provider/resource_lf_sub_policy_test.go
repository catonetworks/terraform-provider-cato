package provider

import (
	"context"
	"errors"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestLfSubPolicyAtSchemaIsImportSafe(t *testing.T) {
	t.Parallel()

	resp := &resource.SchemaResponse{}
	(&lfSubPolicyResource{}).Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), "unexpected schema diagnostics: %+v", resp.Diagnostics)

	at, ok := resp.Schema.Attributes["at"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	require.True(t, at.Optional)
	require.False(t, at.Required)
	require.Len(t, at.PlanModifiers, 1)
}

func TestLfSubPolicyPositionReplacementModifier(t *testing.T) {
	t.Parallel()

	modifier := configuredPositionModifier{}
	first := lfSubPolicyPosition("LAST_IN_POLICY", types.StringNull())
	second := lfSubPolicyPosition("AFTER_RULE", types.StringValue("rule-123"))

	t.Run("imported null position accepts initial configuration", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: types.ObjectNull(PositionAttrTypes),
			PlanValue:  first,
		}, resp)

		require.False(t, resp.RequiresReplace)
	})

	t.Run("unchanged configured position stays in place", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  first,
		}, resp)

		require.False(t, resp.RequiresReplace)
	})

	t.Run("configured position change is handled by update", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  second,
		}, resp)

		require.False(t, resp.RequiresReplace)
	})

	t.Run("unknown configured position is deferred", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  types.ObjectUnknown(PositionAttrTypes),
		}, resp)

		require.False(t, resp.RequiresReplace)
	})

	t.Run("configured position cannot be removed", func(t *testing.T) {
		t.Parallel()

		resp := &planmodifier.ObjectResponse{}
		modifier.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
			StateValue: first,
			PlanValue:  types.ObjectNull(PositionAttrTypes),
		}, resp)

		require.True(t, resp.Diagnostics.HasError())
	})
}

func TestLfSubPolicyScopeSchemaRequiresAPIEnums(t *testing.T) {
	t.Parallel()

	scope := (&lfSubPolicyResource{}).policyScopeSchema()
	service := scope.Attributes["service"].(schema.SingleNestedAttribute)
	custom := service.Attributes["custom"].(schema.ListNestedAttribute)
	protocol := custom.NestedObject.Attributes["protocol"].(schema.StringAttribute)
	require.True(t, protocol.Required)
	require.False(t, protocol.Optional)

	nat := scope.Attributes["nat"].(schema.SingleNestedAttribute)
	natType := nat.Attributes["nat_type"].(schema.StringAttribute)
	require.True(t, natType.Computed)
	require.NotNil(t, natType.Default)
}

func TestLfSubPolicyPrepareNatDefaultsType(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	nat := types.ObjectValueMust(PolicyNatSettingsTypes, map[string]attr.Value{
		"enabled":  types.BoolValue(true),
		"nat_type": types.StringNull(),
	})

	got := (&lfSubPolicyResource{}).prepareNat(context.Background(), nat, &diags)

	require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)
	require.Equal(t, cato_models.SocketLanNatTypeDynamicPat, got.NatType)
}

func TestLfSubPolicyServiceHydrationPreservesNullBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	res := &lfSubPolicyResource{}

	t.Run("empty service collections remain null", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics
		require.True(t, res.parseSimpleService(ctx, nil, &diags).IsNull())
		require.True(t, res.parseCustomService(ctx, nil, &diags).IsNull())
		require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)
	})

	t.Run("port-only custom service keeps port range null", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics
		custom := []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service_Custom{
			{
				Port:     []scalars.Port{"443"},
				Protocol: cato_models.IPProtocolTCP,
			},
		}

		value := res.parseCustomService(ctx, custom, &diags)

		require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)
		var parsed []PolicyCustomService
		require.False(t, value.ElementsAs(ctx, &parsed, false).HasError())
		require.Len(t, parsed, 1)
		require.True(t, parsed[0].PortRange.IsNull())
	})
}

func TestCheckPolicyMutationStatus(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		status := cato_models.PolicyMutationStatusSuccess
		var diags diag.Diagnostics
		require.False(t, checkPolicyMutationStatus(&status, "mutation failed", &diags))
		require.False(t, diags.HasError())
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		status := cato_models.PolicyMutationStatusFailure
		var diags diag.Diagnostics
		require.True(t, checkPolicyMutationStatus(&status, "mutation failed", &diags))
		require.True(t, diags.HasError())
	})

	t.Run("missing status", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics
		require.True(t, checkPolicyMutationStatus(nil, "mutation failed", &diags))
		require.True(t, diags.HasError())
	})
}

func TestLfSubPolicyMoveUsesScopeRuleID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := &lfSubPolicyFakeClient{}
	res := &lfSubPolicyResource{
		client:          &catoClientData{AccountId: "account-123"},
		subPolicyClient: client,
	}
	var diags diag.Diagnostics

	res.move(ctx, "scope-rule-123", lfSubPolicyPosition("AFTER_RULE", types.StringValue("rule-456")), &diags)

	require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)
	require.Equal(t, "scope-rule-123", client.moveInput.ID)
	require.Equal(t, cato_models.PolicyRulePositionEnumAfterRule, *client.moveInput.To.Position)
	require.Equal(t, "rule-456", *client.moveInput.To.Ref)
}

func TestLfSubPolicyCreatePreservesIDWhenPublishFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resourceSchema := lfSubPolicySchema(ctx, t)
	plan := tfsdk.Plan{Schema: resourceSchema}
	require.False(t, plan.Set(ctx, lfSubPolicyPlan()).HasError())
	client := &lfSubPolicyFakeClient{publishErr: errors.New("publish timeout")}
	res := &lfSubPolicyResource{
		client:          &catoClientData{AccountId: "account-123"},
		subPolicyClient: client,
	}
	response := &resource.CreateResponse{State: tfsdk.State{Schema: resourceSchema}}

	res.Create(ctx, resource.CreateRequest{Plan: plan}, response)

	require.True(t, response.Diagnostics.HasError())
	var state LanFirewallSubPolicy
	require.False(t, response.State.Get(ctx, &state).HasError())
	require.Equal(t, "sub-policy-123", state.ID.ValueString())
	require.Equal(t, []string{"add", "publish"}, client.calls)
}

func TestLfSubPolicyUpdateMovesImportedPositionAndPreservesPriorStateOnPublishFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resourceSchema := lfSubPolicySchema(ctx, t)
	planModel := lfSubPolicyPlan()
	planModel.ID = types.StringValue("sub-policy-123")
	planModel.ScopeRuleID = types.StringValue("scope-rule-123")
	stateModel := planModel
	stateModel.At = types.ObjectNull(PositionAttrTypes)
	plan := tfsdk.Plan{Schema: resourceSchema}
	require.False(t, plan.Set(ctx, planModel).HasError())
	state := tfsdk.State{Schema: resourceSchema}
	require.False(t, state.Set(ctx, stateModel).HasError())
	client := &lfSubPolicyFakeClient{publishErr: errors.New("publish timeout")}
	res := &lfSubPolicyResource{
		client:          &catoClientData{AccountId: "account-123"},
		subPolicyClient: client,
	}
	response := &resource.UpdateResponse{State: tfsdk.State{Schema: resourceSchema}}

	res.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, response)

	require.True(t, response.Diagnostics.HasError())
	var updatedState LanFirewallSubPolicy
	require.False(t, response.State.Get(ctx, &updatedState).HasError())
	require.True(t, updatedState.At.IsNull())
	require.Equal(t, []string{"move", "update", "publish"}, client.calls)
	require.Equal(t, "scope-rule-123", client.moveInput.ID)
}

func TestLfSubPolicyUpdateRejectsResolvedNullPosition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resourceSchema := lfSubPolicySchema(ctx, t)
	planModel := lfSubPolicyPlan()
	planModel.ID = types.StringValue("sub-policy-123")
	planModel.ScopeRuleID = types.StringValue("scope-rule-123")
	stateModel := planModel
	planModel.At = types.ObjectNull(PositionAttrTypes)
	plan := tfsdk.Plan{Schema: resourceSchema}
	require.False(t, plan.Set(ctx, planModel).HasError())
	state := tfsdk.State{Schema: resourceSchema}
	require.False(t, state.Set(ctx, stateModel).HasError())
	client := &lfSubPolicyFakeClient{}
	res := &lfSubPolicyResource{
		client:          &catoClientData{AccountId: "account-123"},
		subPolicyClient: client,
	}
	response := &resource.UpdateResponse{State: tfsdk.State{Schema: resourceSchema}}

	res.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, response)

	require.True(t, response.Diagnostics.HasError())
	require.Empty(t, client.calls)
}

func TestLfSubPolicyPublishTreatsNoDraftAsSuccess(t *testing.T) {
	t.Parallel()

	client := &lfSubPolicyFakeClient{publishNotFound: true}
	res := &lfSubPolicyResource{
		client:          &catoClientData{AccountId: "account-123"},
		subPolicyClient: client,
	}
	var diags diag.Diagnostics

	res.publish(context.Background(), &diags)

	require.False(t, diags.HasError(), "unexpected diagnostics: %+v", diags)
}

func lfSubPolicyPosition(position string, ref types.String) types.Object {
	return types.ObjectValueMust(PositionAttrTypes, map[string]attr.Value{
		"position": types.StringValue(position),
		"ref":      ref,
	})
}

func lfSubPolicySchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	(&lfSubPolicyResource{}).Schema(ctx, resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), "unexpected schema diagnostics: %+v", resp.Diagnostics)
	return resp.Schema
}

func lfSubPolicyPlan() LanFirewallSubPolicy {
	scope := types.ObjectValueMust(LanFirewallSubPolicyScopeTypes, map[string]attr.Value{
		"description": types.StringUnknown(),
		"destination": types.ObjectNull(SocketLanDestinationAttrTypes),
		"direction":   types.StringValue("BOTH"),
		"enabled":     types.BoolValue(true),
		"id":          types.StringUnknown(),
		"name":        types.StringUnknown(),
		"nat":         types.ObjectNull(PolicyNatSettingsTypes),
		"service":     types.ObjectNull(PolicyServiceTypes),
		"site":        types.ObjectNull(PolicySiteTypes),
		"source":      types.ObjectNull(SocketLanSourceAttrTypes),
	})
	return LanFirewallSubPolicy{
		ID:          types.StringUnknown(),
		Name:        types.StringValue("sub-policy"),
		Description: types.StringValue("description"),
		At:          lfSubPolicyPosition("LAST_IN_POLICY", types.StringNull()),
		ScopeRuleID: types.StringUnknown(),
		Scope:       scope,
	}
}

type lfSubPolicyFakeClient struct {
	moveInput       cato_models.PolicyMoveRuleInput
	publishErr      error
	publishNotFound bool
	calls           []string
}

func (c *lfSubPolicyFakeClient) PolicySocketLanMoveRule(
	_ context.Context,
	input cato_models.PolicyMoveRuleInput,
	_ string,
	_ ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanMoveRule, error) {
	c.moveInput = input
	c.calls = append(c.calls, "move")
	return &cato_go_sdk.PolicySocketLanMoveRule{
		Policy: &cato_go_sdk.PolicySocketLanMoveRule_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanMoveRule_Policy_SocketLan{
				MoveRule: cato_go_sdk.PolicySocketLanMoveRule_Policy_SocketLan_MoveRule{
					Status: cato_models.PolicyMutationStatusSuccess,
				},
			},
		},
	}, nil
}

func (*lfSubPolicyFakeClient) PolicySocketLanPolicy(
	context.Context,
	string,
	*cato_models.SocketLanPolicyInput,
	...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPolicy, error) {
	return nil, nil
}

func (c *lfSubPolicyFakeClient) PolicySocketLanAddSubPolicy(
	context.Context,
	cato_models.SocketLanAddSubPolicyInput,
	string,
	...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanAddSubPolicy, error) {
	c.calls = append(c.calls, "add")
	return &cato_go_sdk.PolicySocketLanAddSubPolicy{
		Policy: &cato_go_sdk.PolicySocketLanAddSubPolicy_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanAddSubPolicy_Policy_SocketLan{
				AddSubPolicy: cato_go_sdk.PolicySocketLanAddSubPolicy_Policy_SocketLan_AddSubPolicy{
					Status: cato_models.PolicyMutationStatusSuccess,
					Policy: &cato_go_sdk.PolicySocketLanAddSubPolicy_Policy_SocketLan_AddSubPolicy_Policy{
						ID: "sub-policy-123",
					},
				},
			},
		},
	}, nil
}

func (c *lfSubPolicyFakeClient) PolicySocketLanUpdateRule(
	context.Context,
	*cato_models.SocketLanPolicyMutationInput,
	cato_models.SocketLanUpdateRuleInput,
	string,
	...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanUpdateRule, error) {
	c.calls = append(c.calls, "update")
	return &cato_go_sdk.PolicySocketLanUpdateRule{
		Policy: &cato_go_sdk.PolicySocketLanUpdateRule_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanUpdateRule_Policy_SocketLan{
				UpdateRule: cato_go_sdk.PolicySocketLanUpdateRule_Policy_SocketLan_UpdateRule{
					Status: cato_models.PolicyMutationStatusSuccess,
				},
			},
		},
	}, nil
}

func (*lfSubPolicyFakeClient) PolicySocketLanRemoveSubPolicy(
	context.Context,
	*cato_models.SocketLanPolicyMutationInput,
	cato_models.SocketLanRemoveSubPolicyInput,
	string,
	...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanRemoveSubPolicy, error) {
	return nil, nil
}

func (c *lfSubPolicyFakeClient) PolicySocketLanPublishPolicyRevision(
	context.Context,
	*cato_models.SocketLanPolicyMutationInput,
	*cato_models.PolicyPublishRevisionInput,
	string,
	...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicySocketLanPublishPolicyRevision, error) {
	c.calls = append(c.calls, "publish")
	if c.publishErr != nil {
		return nil, c.publishErr
	}
	if c.publishNotFound {
		code := "PolicyRevisionNotFound"
		return &cato_go_sdk.PolicySocketLanPublishPolicyRevision{
			Policy: &cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy{
				SocketLan: &cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy_SocketLan{
					PublishPolicyRevision: cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy_SocketLan_PublishPolicyRevision{
						Status: cato_models.PolicyMutationStatusFailure,
						Errors: []*cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy_SocketLan_PublishPolicyRevision_Errors{
							{ErrorCode: &code},
						},
					},
				},
			},
		}, nil
	}
	return &cato_go_sdk.PolicySocketLanPublishPolicyRevision{
		Policy: &cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy{
			SocketLan: &cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy_SocketLan{
				PublishPolicyRevision: cato_go_sdk.PolicySocketLanPublishPolicyRevision_Policy_SocketLan_PublishPolicyRevision{
					Status: cato_models.PolicyMutationStatusSuccess,
				},
			},
		},
	}, nil
}
