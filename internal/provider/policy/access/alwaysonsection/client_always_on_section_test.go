package alwaysonsection

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
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestSectionSDKForwarding(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"AddSection", "Policy", "UpdateSection", "MoveSection", "RemoveSection"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			type recordedRequest struct {
				OperationName string         `json:"operationName"`
				Variables     map[string]any `json:"variables"`
				Query         string         `json:"query"`
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
				request.Forwarded = req.Header.Get("X-Section-Test")
				requests <- request
				w.Header().Set("Content-Type", "application/json")
				if _, err := fmt.Fprint(w, `{"data":{"policy":{"ztnaAlwaysOn":{}}}}`); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			client, err := cato.New(server.URL, "test-token", "sdk-default-account", nil, nil)
			require.NoError(t, err)
			r := &alwaysOnSectionResource{client: &catoclient.Service{Catov2: catoclient.NewProviderSDKClient(client)}}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			interceptor := func(ctx context.Context, req *http.Request, info *clientv2.GQLRequestInfo, result any,
				next clientv2.RequestInterceptorFunc,
			) error {
				req.Header.Set("X-Section-Test", "forwarded")
				return next(ctx, req, info, result)
			}
			revision, ref, name := "revision-1", "other-section", "Example"
			policy := &cato_models.ZtnaAlwaysOnPolicyMutationInput{Revision: &cato_models.PolicyMutationRevisionInput{ID: &revision}}
			position := &cato_models.PolicySectionPositionInput{Position: "BEFORE_SECTION", Ref: &ref}
			switch operation {
			case "AddSection":
				_, err = r.getClient().PolicyZtnaAlwaysOnAddSection(ctx, "123", policy,
					cato_models.PolicyAddSectionInput{At: position, Section: &cato_models.PolicyAddSectionInfoInput{Name: name}}, interceptor)
			case "Policy":
				_, err = r.getClient().PolicyZtnaAlwaysOnPolicy(ctx, "123",
					&cato_models.ZtnaAlwaysOnPolicyInput{Revision: &cato_models.PolicyRevisionInput{ID: &revision}}, interceptor)
			case "UpdateSection":
				_, err = r.getClient().PolicyZtnaAlwaysOnUpdateSection(ctx, "123", policy,
					cato_models.PolicyUpdateSectionInput{ID: "section-123", Section: &cato_models.PolicyUpdateSectionInfoInput{Name: &name}}, interceptor)
			case "MoveSection":
				_, err = r.getClient().PolicyZtnaAlwaysOnMoveSection(ctx, "123", policy,
					cato_models.PolicyMoveSectionInput{ID: "section-123", To: position}, interceptor)
			case "RemoveSection":
				_, err = r.getClient().PolicyZtnaAlwaysOnRemoveSection(ctx, "123", policy,
					cato_models.PolicyRemoveSectionInput{ID: "section-123"}, interceptor)
			}
			require.NoError(t, err)
			request := <-requests
			require.Equal(t, "policyZtnaAlwaysOn"+operation, request.OperationName)
			require.Equal(t, "123", request.Variables["accountId"])
			require.Equal(t, "forwarded", request.Forwarded)
			policyVariable := "ztnaAlwaysOnPolicyMutationInput"
			if operation == "Policy" {
				policyVariable = "ztnaAlwaysOnPolicyInput"
				require.Contains(t, request.Query, "sections")
			}
			require.Equal(t, map[string]any{"revision": map[string]any{"id": revision}}, request.Variables[policyVariable])
			expectedPosition := map[string]any{"position": "BEFORE_SECTION", "ref": ref}
			switch operation {
			case "AddSection":
				require.Equal(t, map[string]any{"at": expectedPosition, "section": map[string]any{"name": name}}, request.Variables["policyAddSectionInput"])
			case "UpdateSection":
				require.Equal(t, map[string]any{"id": "section-123", "section": map[string]any{"name": name}}, request.Variables["policyUpdateSectionInput"])
			case "MoveSection":
				require.Equal(t, map[string]any{"id": "section-123", "to": expectedPosition}, request.Variables["policyMoveSectionInput"])
			case "RemoveSection":
				require.Equal(t, map[string]any{"id": "section-123"}, request.Variables["policyRemoveSectionInput"])
			}
		})
	}
}
