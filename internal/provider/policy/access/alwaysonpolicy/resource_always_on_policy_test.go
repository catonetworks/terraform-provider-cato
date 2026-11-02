package alwaysonpolicy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func policySchema() schema.Schema {
	var resp resource.SchemaResponse
	NewAlwaysOnPolicyResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp.Schema
}

func policyPlan(t *testing.T, enabled types.Bool) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: policySchema()}
	require.False(t, plan.Set(context.Background(), PolicyModel{Enabled: enabled}).HasError())
	return plan
}

func policyState(t *testing.T, enabled types.Bool) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: policySchema()}
	require.False(t, state.Set(context.Background(), PolicyModel{Enabled: enabled}).HasError())
	return state
}

func decodeResponse[T any](t *testing.T, body string) *T {
	t.Helper()
	if body == "" {
		return nil
	}
	var result T
	require.NoError(t, json.Unmarshal([]byte(body), &result))
	return &result
}

func policyResponse(t *testing.T, enabled bool) *cato.PolicyZtnaAlwaysOnPolicy {
	t.Helper()
	return decodeResponse[cato.PolicyZtnaAlwaysOnPolicy](t,
		fmt.Sprintf(`{"policy":{"ztnaAlwaysOn":{"policy":{"enabled":%t,"rules":[],"sections":[]}}}}`, enabled))
}

func successfulMutation(t *testing.T, enabled bool) *cato.PolicyZtnaAlwaysOnUpdatePolicy {
	t.Helper()
	return decodeResponse[cato.PolicyZtnaAlwaysOnUpdatePolicy](t,
		fmt.Sprintf(`{"policy":{"ztnaAlwaysOn":{"updatePolicy":{"status":"SUCCESS","errors":[],"policy":{"enabled":%t,"revision":{"changes":1}}}}}}`, enabled))
}

func TestPolicyWiringAndSchema(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := NewAlwaysOnPolicyResource().(*alwaysOnPolicyResource)
	var metadata resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
	require.Equal(t, "cato_always_on_policy", metadata.TypeName)
	s := policySchema()
	require.Len(t, s.Attributes, 1)
	enabled, ok := s.Attributes["enabled"].(schema.BoolAttribute)
	require.True(t, ok)
	require.True(t, enabled.Required)
	require.False(t, enabled.Optional)
	require.False(t, enabled.Computed)
	var configured resource.ConfigureResponse
	r.Configure(ctx, resource.ConfigureRequest{}, &configured)
	require.Nil(t, r.getClient())
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "invalid"}, &configured)
	require.True(t, configured.Diagnostics.HasError())
	configured = resource.ConfigureResponse{}
	service := &catoclient.Service{AccountId: "123"}
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: service}, &configured)
	require.False(t, configured.Diagnostics.HasError())
	require.Same(t, service, r.client)
	require.Nil(t, r.getClient())
	client := mocks.NewAlwaysOnPolicyClient(t)
	r.apiClient = client
	require.Same(t, client, r.getClient())
}

func TestPolicyCreateAndUpdate(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"create", "update"} {
		for _, enabled := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/%t", operation, enabled), func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				client := mocks.NewAlwaysOnPolicyClient(t)
				toggle := cato_models.PolicyToggleStateDisabled
				if enabled {
					toggle = cato_models.PolicyToggleStateEnabled
				}
				update := client.On("PolicyZtnaAlwaysOnUpdatePolicy", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
					cato_models.ZtnaAlwaysOnPolicyUpdateInput{State: &toggle}).Return(successfulMutation(t, enabled), nil).Once()
				client.On("PolicyZtnaAlwaysOnPublishPolicyRevision", mock.Anything, "123",
					(*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil), (*cato_models.PolicyPublishRevisionInput)(nil)).
					Return(successfulPublish(t), nil).Once().NotBefore(update)
				r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
				plan := policyPlan(t, types.BoolValue(enabled))
				var state tfsdk.State
				if operation == "create" {
					resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
					r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
					require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
					state = resp.State
				} else {
					previous := policyState(t, types.BoolValue(!enabled))
					resp := resource.UpdateResponse{State: previous}
					r.Update(ctx, resource.UpdateRequest{Plan: plan, State: previous}, &resp)
					require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
					state = resp.State
				}
				var actual PolicyModel
				require.False(t, state.Get(ctx, &actual).HasError())
				require.Equal(t, types.BoolValue(enabled), actual.Enabled)
			})
		}
	}
}

func TestPolicyRead(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, message string
		enabled             bool
		err                 error
	}{
		{name: "enabled", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"enabled":true,"rules":[],"sections":[]}}}}`, enabled: true},
		{name: "disabled", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"enabled":false,"rules":[],"sections":[]}}}}`},
		{name: "transport", err: errors.New("read failed"), message: "read failed"},
		{name: "nil response", message: "policy"},
		{name: "missing Always On response", body: `{"policy":{}}`, message: "policy"},
		{name: "empty policy defaults to disabled", body: `{"policy":{"ztnaAlwaysOn":{}}}`},
		{name: "without unused lists", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"enabled":false}}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := mocks.NewAlwaysOnPolicyClient(t)
			client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyInput)(nil)).
				Return(decodeResponse[cato.PolicyZtnaAlwaysOnPolicy](t, tc.body), tc.err).Once()
			r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := policyState(t, types.BoolValue(!tc.enabled))
			resp := resource.ReadResponse{State: state}
			r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
			if tc.message != "" {
				require.True(t, resp.Diagnostics.HasError())
				require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
				require.True(t, state.Raw.Equal(resp.State.Raw))
				return
			}
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var actual PolicyModel
			require.False(t, resp.State.Get(context.Background(), &actual).HasError())
			require.Equal(t, types.BoolValue(tc.enabled), actual.Enabled)
		})
	}
}

func TestPolicyMutationFailures(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, message string
		err                 error
	}{
		{name: "transport", err: errors.New("update failed"), message: "update failed"},
		{name: "nil response", message: "mutation status"},
		{name: "failure status", body: `{"policy":{"ztnaAlwaysOn":{"updatePolicy":{"status":"FAILURE"}}}}`, message: "FAILURE"},
		{name: "payload error", body: `{"policy":{"ztnaAlwaysOn":{"updatePolicy":{"status":"SUCCESS","errors":[{"errorMessage":"Policy locked"}]}}}}`, message: "Policy locked"},
	} {
		for _, operation := range []string{"create", "update"} {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				client := mocks.NewAlwaysOnPolicyClient(t)
				client.On("PolicyZtnaAlwaysOnUpdatePolicy", mock.Anything, "123", mock.Anything, mock.Anything).
					Return(decodeResponse[cato.PolicyZtnaAlwaysOnUpdatePolicy](t, tc.body), tc.err).Once()
				r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
				plan := policyPlan(t, types.BoolValue(true))
				if operation == "create" {
					resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
					r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
					require.True(t, resp.Diagnostics.HasError())
					require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
				} else {
					state := policyState(t, types.BoolValue(false))
					resp := resource.UpdateResponse{State: state}
					r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, &resp)
					require.True(t, resp.Diagnostics.HasError())
					require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
					require.True(t, state.Raw.Equal(resp.State.Raw))
				}
			})
		}
	}
}

func TestPolicyUpdateRejectsUnchangedToggle(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{true, false} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			t.Parallel()
			client := mocks.NewAlwaysOnPolicyClient(t)
			client.On("PolicyZtnaAlwaysOnUpdatePolicy", mock.Anything, "123", mock.Anything, mock.Anything).
				Return(successfulMutation(t, !enabled), nil).Once()
			r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := policyState(t, types.BoolValue(!enabled))
			resp := resource.UpdateResponse{State: state}
			r.Update(context.Background(), resource.UpdateRequest{Plan: policyPlan(t, types.BoolValue(enabled)), State: state}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "did not match the requested value")
			require.True(t, state.Raw.Equal(resp.State.Raw))
		})
	}
}

func TestPolicyDelete(t *testing.T) {
	t.Parallel()
	client := mocks.NewAlwaysOnPolicyClient(t)
	client.On("PolicyZtnaAlwaysOnPublishPolicyRevision", mock.Anything, "123",
		(*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil), (*cato_models.PolicyPublishRevisionInput)(nil)).
		Return(successfulPublish(t), nil).Once()
	r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	state := policyState(t, types.BoolValue(true))
	resp := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.True(t, resp.State.Raw.IsNull())
}

func TestPolicyInvalidInput(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"null", "unknown"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: mocks.NewAlwaysOnPolicyClient(t)}
			enabled := types.BoolValue(true)
			switch mode {
			case "null":
				enabled = types.BoolNull()
			case "unknown":
				enabled = types.BoolUnknown()
			}
			plan := policyPlan(t, enabled)
			create := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &create)
			require.True(t, create.Diagnostics.HasError())
			state := policyState(t, types.BoolValue(false))
			update := resource.UpdateResponse{State: state}
			r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, &update)
			require.True(t, update.Diagnostics.HasError())
		})
	}
}

func TestPolicyImport(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{true, false} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			t.Parallel()
			client := mocks.NewAlwaysOnPolicyClient(t)
			client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyInput)(nil)).
				Return(policyResponse(t, enabled), nil).Once()
			r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			resp := resource.ImportStateResponse{State: tfsdk.State{Schema: policySchema()}}
			r.ImportState(context.Background(), resource.ImportStateRequest{ID: "123"}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var actual PolicyModel
			require.False(t, resp.State.Get(context.Background(), &actual).HasError())
			require.Equal(t, types.BoolValue(enabled), actual.Enabled)
		})
	}
}

func TestPolicyImportErrors(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"unconfigured", "missing account", "missing client", "read error", "missing policy"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			client := mocks.NewAlwaysOnPolicyClient(t)
			r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			id := "123"
			switch mode {
			case "unconfigured":
				r.client = nil
			case "missing account":
				r.client.AccountId = ""
			case "missing client":
				r.apiClient = nil
			case "read error", "missing policy":
				var result *cato.PolicyZtnaAlwaysOnPolicy
				var err error
				if mode == "read error" {
					err = errors.New("read failed")
				}
				client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", mock.Anything).Return(result, err).Once()
			}
			resp := resource.ImportStateResponse{State: tfsdk.State{Schema: policySchema()}}
			r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &resp)
			require.True(t, resp.Diagnostics.HasError())
		})
	}
}

func successfulPublish(t *testing.T) *cato.PolicyZtnaAlwaysOnPublishPolicyRevision {
	t.Helper()
	return decodeResponse[cato.PolicyZtnaAlwaysOnPublishPolicyRevision](t,
		`{"policy":{"ztnaAlwaysOn":{"publishPolicyRevision":{"status":"SUCCESS","errors":[]}}}}`)
}

func TestPolicyPublishFailures(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"create", "update"} {
		for _, tc := range []struct {
			name, body, message string
			err                 error
		}{
			{name: "transport", err: errors.New("publish failed"), message: "publish failed"},
			{name: "nil response", message: "mutation status"},
			{name: "missing status", body: `{"policy":{"ztnaAlwaysOn":{"publishPolicyRevision":{}}}}`, message: "mutation status"},
			{name: "failure status", body: `{"policy":{"ztnaAlwaysOn":{"publishPolicyRevision":{"status":"FAILURE"}}}}`, message: "FAILURE"},
			{name: "empty status", body: `{"policy":{"ztnaAlwaysOn":{"publishPolicyRevision":{"status":""}}}}`, message: "mutation status"},
			{name: "unexpected status", body: `{"policy":{"ztnaAlwaysOn":{"publishPolicyRevision":{"status":"UNKNOWN"}}}}`, message: "UNKNOWN"},
			{name: "payload error", body: `{"policy":{"ztnaAlwaysOn":{"publishPolicyRevision":{"status":"SUCCESS","errors":[{"errorMessage":"Policy locked"}]}}}}`, message: "Policy locked"},
		} {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				client := mocks.NewAlwaysOnPolicyClient(t)
				update := client.On("PolicyZtnaAlwaysOnUpdatePolicy", mock.Anything, "123", mock.Anything, mock.Anything).
					Return(successfulMutation(t, true), nil).Once()
				client.On("PolicyZtnaAlwaysOnPublishPolicyRevision", mock.Anything, "123",
					(*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil), (*cato_models.PolicyPublishRevisionInput)(nil)).
					Return(decodeResponse[cato.PolicyZtnaAlwaysOnPublishPolicyRevision](t, tc.body), tc.err).Once().NotBefore(update)
				r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
				plan := policyPlan(t, types.BoolValue(true))
				if operation == "create" {
					resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
					r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
					require.True(t, resp.Diagnostics.HasError())
					require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "publish")
					require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
					require.True(t, resp.State.Raw.IsNull())
				} else {
					state := policyState(t, types.BoolValue(false))
					resp := resource.UpdateResponse{State: state}
					r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, &resp)
					require.True(t, resp.Diagnostics.HasError())
					require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "publish")
					require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
					require.True(t, state.Raw.Equal(resp.State.Raw))
				}
			})
		}
	}
}
