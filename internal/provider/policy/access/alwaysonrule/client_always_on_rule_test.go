package alwaysonrule

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestAlwaysOnRuleSDKForwarding(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"policyZtnaAlwaysOnPolicy", "policyZtnaAlwaysOnAddRule"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			body := `{"data":{"policy":{"ztnaAlwaysOn":{"policy":{"rules":[]}}}}}`
			if operation == "policyZtnaAlwaysOnAddRule" {
				body = `{"data":{"policy":{"ztnaAlwaysOn":{"addRule":{"status":"SUCCESS","errors":[]}}}}}`
			}
			var request struct {
				OperationName string         `json:"operationName"`
				Variables     map[string]any `json:"variables"`
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() { require.NoError(t, r.Body.Close()) }()
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(body))
				require.NoError(t, err)
			}))
			defer server.Close()
			client, err := cato.New(server.URL, "test-token", "sdk-default-account", nil, nil)
			require.NoError(t, err)
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123", Catov2: catoclient.NewProviderSDKClient(client)}}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			revision := "revision-1"
			if operation == "policyZtnaAlwaysOnPolicy" {
				result, err := r.getClient().PolicyZtnaAlwaysOnPolicy(ctx, r.client.AccountId,
					&cato_models.ZtnaAlwaysOnPolicyInput{Revision: &cato_models.PolicyRevisionInput{ID: &revision}})
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, result.GetPolicy().GetZtnaAlwaysOn().GetPolicy().GetRules())
				require.Equal(t, map[string]any{"revision": map[string]any{"id": revision}}, request.Variables["ztnaAlwaysOnPolicyInput"])
			} else {
				result, err := r.getClient().PolicyZtnaAlwaysOnAddRule(ctx, r.client.AccountId,
					&cato_models.ZtnaAlwaysOnPolicyMutationInput{Revision: &cato_models.PolicyMutationRevisionInput{ID: &revision}},
					cato_models.ZtnaAlwaysOnAddRuleInput{Rule: &cato_models.ZtnaAlwaysOnAddRuleDataInput{Name: "Example rule"}})
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, map[string]any{"revision": map[string]any{"id": revision}}, request.Variables["ztnaAlwaysOnPolicyMutationInput"])
				ruleInput := request.Variables["ztnaAlwaysOnAddRuleInput"].(map[string]any)["rule"].(map[string]any)
				require.Equal(t, "Example rule", ruleInput["name"])
			}
			require.Equal(t, operation, request.OperationName)
			require.Equal(t, "123", request.Variables["accountId"])
		})
	}
}

func TestAlwaysOnRuleSDKMutationForwarding(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"update empty", "update populated", "update omitted", "move", "delete"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			operation, mutation := "policyZtnaAlwaysOnUpdateRule", "updateRule"
			if mode == "move" {
				operation, mutation = "policyZtnaAlwaysOnMoveRule", "moveRule"
			}
			if mode == "delete" {
				operation, mutation = "policyZtnaAlwaysOnRemoveRule", "removeRule"
			}
			var request struct {
				OperationName string         `json:"operationName"`
				Variables     map[string]any `json:"variables"`
			}
			body, err := json.Marshal(map[string]any{"data": map[string]any{"policy": map[string]any{"ztnaAlwaysOn": map[string]any{
				mutation: map[string]any{"status": "SUCCESS", "errors": []any{}},
			}}}})
			require.NoError(t, err)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() { require.NoError(t, r.Body.Close()) }()
				require.Equal(t, "forwarded", r.Header.Get("X-AlwaysOn-Test"))
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(body)
				require.NoError(t, err)
			}))
			defer server.Close()
			client, err := cato.New(server.URL, "test-token", "sdk-default-account", nil, nil)
			require.NoError(t, err)
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123", Catov2: catoclient.NewProviderSDKClient(client)}}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			revision := "revision-1"
			policy := &cato_models.ZtnaAlwaysOnPolicyMutationInput{Revision: &cato_models.PolicyMutationRevisionInput{ID: &revision}}
			interceptor := func(ctx context.Context, req *http.Request, info *clientv2.GQLRequestInfo, result any, next clientv2.RequestInterceptorFunc) error {
				req.Header.Set("X-AlwaysOn-Test", "forwarded")
				return next(ctx, req, info, result)
			}
			switch mode {
			case "move":
				position := cato_models.PolicyRulePositionEnum("LAST_IN_POLICY")
				_, err = r.getClient().PolicyZtnaAlwaysOnMoveRule(ctx, "123", policy,
					cato_models.PolicyMoveRuleInput{ID: "rule-123", To: &cato_models.PolicyRulePositionInput{Position: &position}}, interceptor)
				require.NoError(t, err)
				require.Equal(t, map[string]any{"id": "rule-123", "to": map[string]any{"position": "LAST_IN_POLICY"}}, request.Variables["policyMoveRuleInput"])
			case "delete":
				_, err = r.getClient().PolicyZtnaAlwaysOnRemoveRule(ctx, "123", policy, cato_models.ZtnaAlwaysOnRemoveRuleInput{ID: "rule-123"}, interceptor)
				require.NoError(t, err)
				require.Equal(t, map[string]any{"id": "rule-123"}, request.Variables["ztnaAlwaysOnRemoveRuleInput"])
			default:
				var model AlwaysOnRule
				require.False(t, alwaysOnState(t).Get(ctx, &model).HasError())
				var diags diag.Diagnostics
				input := buildAlwaysOnUpdateRuleInput(ctx, alwaysOnUpdateModel(t), model, &diags)
				require.False(t, diags.HasError(), "%v", diags)
				switch mode {
				case "update populated":
					input.Rule.Platform = []cato_models.OperatingSystem{"WINDOWS"}
					input.Rule.DevicePostureProfile = []*cato_models.DeviceProfileRefInput{{By: cato_models.ObjectRefByID, Input: "profile-1"}}
					input.Rule.Source.User = []*cato_models.UserRefInput{{By: cato_models.ObjectRefByName, Input: "User"}}
					input.Rule.Source.UsersGroup = []*cato_models.UsersGroupRefInput{{By: cato_models.ObjectRefByID, Input: "group-1"}}
				case "update omitted":
					input.Rule.Platform, input.Rule.DevicePostureProfile, input.Rule.Source = nil, nil, nil
				}
				_, err = r.getClient().PolicyZtnaAlwaysOnUpdateRule(ctx, "123", policy, input, interceptor)
				require.NoError(t, err)
				payload := request.Variables["ztnaAlwaysOnUpdateRuleInput"].(map[string]any)
				require.Equal(t, "rule-123", payload["id"])
				rule := payload["rule"].(map[string]any)
				require.Equal(t, false, rule["enabled"])
				require.Equal(t, false, rule["allowUserBypass"])
				require.Equal(t, "", rule["description"])
				switch mode {
				case "update empty":
					require.Equal(t, []any{}, rule["platform"])
					require.Equal(t, []any{}, rule["devicePostureProfile"])
					require.Equal(t, map[string]any{"user": []any{}, "usersGroup": []any{}}, rule["source"])
				case "update populated":
					require.Equal(t, []any{"WINDOWS"}, rule["platform"])
					require.Equal(t, []any{map[string]any{"by": "ID", "input": "profile-1"}}, rule["devicePostureProfile"])
					require.Equal(t, map[string]any{"user": []any{map[string]any{"by": "NAME", "input": "User"}}, "usersGroup": []any{map[string]any{"by": "ID", "input": "group-1"}}}, rule["source"])
				default:
					require.NotContains(t, rule, "platform")
					require.NotContains(t, rule, "devicePostureProfile")
					require.NotContains(t, rule, "source")
				}
			}
			require.Equal(t, operation, request.OperationName)
			require.Equal(t, "123", request.Variables["accountId"])
			require.Equal(t, map[string]any{"revision": map[string]any{"id": revision}}, request.Variables["ztnaAlwaysOnPolicyMutationInput"])
		})
	}
}
