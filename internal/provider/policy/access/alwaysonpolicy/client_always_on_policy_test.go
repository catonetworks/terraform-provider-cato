package alwaysonpolicy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestPolicySDKForwarding(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"read", "enable", "disable", "publish"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			type recordedRequest struct {
				OperationName string         `json:"operationName"`
				Variables     map[string]any `json:"variables"`
				Forwarded     string
			}
			requests := make(chan recordedRequest, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				defer func() {
					if err := req.Body.Close(); err != nil {
						t.Error(err)
					}
				}()
				var request recordedRequest
				if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				request.Forwarded = req.Header.Get("X-Policy-Test")
				requests <- request
				w.Header().Set("Content-Type", "application/json")
				if _, err := fmt.Fprint(w, `{"data":{"policy":{"ztnaAlwaysOn":{}}}}`); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			client, err := cato.New(server.URL, "test-token", "sdk-default-account", nil, nil)
			require.NoError(t, err)
			r := &alwaysOnPolicyResource{client: &catoclient.Service{Catov2: catoclient.NewProviderSDKClient(client)}}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			interceptor := func(ctx context.Context, req *http.Request, info *clientv2.GQLRequestInfo, result any,
				next clientv2.RequestInterceptorFunc,
			) error {
				req.Header.Set("X-Policy-Test", "forwarded")
				return next(ctx, req, info, result)
			}
			revision := "revision-1"
			switch operation {
			case "read":
				_, err = r.getClient().PolicyZtnaAlwaysOnPolicy(ctx, "123",
					&cato_models.ZtnaAlwaysOnPolicyInput{Revision: &cato_models.PolicyRevisionInput{ID: &revision}}, interceptor)
			case "publish":
				_, err = r.getClient().PolicyZtnaAlwaysOnPublishPolicyRevision(ctx, "123",
					&cato_models.ZtnaAlwaysOnPolicyMutationInput{Revision: &cato_models.PolicyMutationRevisionInput{ID: &revision}},
					&cato_models.PolicyPublishRevisionInput{}, interceptor)
			default:
				var diags diag.Diagnostics
				input := buildPolicyInput(PolicyModel{Enabled: types.BoolValue(operation == "enable")}, &diags)
				require.False(t, diags.HasError())
				_, err = r.getClient().PolicyZtnaAlwaysOnUpdatePolicy(ctx, "123",
					&cato_models.ZtnaAlwaysOnPolicyMutationInput{Revision: &cato_models.PolicyMutationRevisionInput{ID: &revision}}, input, interceptor)
			}
			require.NoError(t, err)
			request := <-requests
			require.Equal(t, "123", request.Variables["accountId"])
			require.Equal(t, "forwarded", request.Forwarded)
			switch operation {
			case "read":
				require.Equal(t, "policyZtnaAlwaysOnPolicy", request.OperationName)
				require.Equal(t, map[string]any{"revision": map[string]any{"id": revision}}, request.Variables["ztnaAlwaysOnPolicyInput"])
			case "publish":
				require.Equal(t, "policyZtnaAlwaysOnPublishPolicyRevision", request.OperationName)
				require.Equal(t, map[string]any{"revision": map[string]any{"id": revision}}, request.Variables["ztnaAlwaysOnPolicyMutationInput"])
				require.Equal(t, map[string]any{}, request.Variables["policyPublishRevisionInput"])
			default:
				require.Equal(t, "policyZtnaAlwaysOnUpdatePolicy", request.OperationName)
				require.Equal(t, map[string]any{"revision": map[string]any{"id": revision}}, request.Variables["ztnaAlwaysOnPolicyMutationInput"])
				expected := "DISABLED"
				if operation == "enable" {
					expected = "ENABLED"
				}
				require.Equal(t, map[string]any{"state": expected}, request.Variables["ztnaAlwaysOnPolicyUpdateInput"])
			}
		})
	}
}
