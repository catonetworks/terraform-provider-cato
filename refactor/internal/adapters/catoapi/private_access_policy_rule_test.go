package catoapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/stretchr/testify/require"

	privateaccess "github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network"
)

func TestRemovePrivateAccessPolicyRule(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, response string
		want           error
	}{
		{"success", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS"}}}}}`, nil},
		{"absent", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"ruleNotExist"}]}}}}}`, privateaccess.ErrRejected},
		{"other absent code", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"RuleNotFound"}]}}}}}`, privateaccess.ErrRejected},
		{"mixed failure", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"ruleNotExist"},{"errorCode":"PermissionDenied"}]}}}}}`, privateaccess.ErrRejected},
		{"missing status", `{"data":{"policy":{"privateAccess":{"removeRule":{}}}}}`, privateaccess.ErrInvalidResponse},
		{"null result", `{"data":{"policy":{"privateAccess":{"removeRule":null}}}}`, privateaccess.ErrInvalidResponse},
		{"null policy", `{"data":{"policy":null}}`, privateaccess.ErrInvalidResponse},
		{"empty failure", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[]}}}}}`, privateaccess.ErrInvalidResponse},
		{"nil error", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[null]}}}}}`, privateaccess.ErrInvalidResponse},
		{"missing code", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorMessage":"denied"}]}}}}}`, privateaccess.ErrInvalidResponse},
		{"contradictory success", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"SUCCESS","errors":[{"errorCode":"PermissionDenied"}]}}}}}`, privateaccess.ErrInvalidResponse},
		{"unknown status", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"PENDING"}}}}}`, privateaccess.ErrInvalidResponse},
		{"conflict", `{"data":{"policy":{"privateAccess":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"reorderPolicyBlockedByActiveSessions"}]}}}}}`, privateaccess.ErrRejected},
		{"graphql", `{"errors":[{"message":"sensitive backend details"}]}`, privateaccess.ErrUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, requests := testCatoAPI(t, tc.response)
			err := NewAdapter(client).RemovePrivateAccessPolicyRule(context.Background(), "account-123", "rule-456")
			if tc.want == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.want)
				require.NotContains(t, err.Error(), "sensitive backend details")
				if tc.want != privateaccess.ErrRuleNotFound {
					require.NotErrorIs(t, err, privateaccess.ErrRuleNotFound)
				}
			}
			request := <-requests
			require.Equal(t, "policyPrivateAccessDeleteRule", request.OperationName)
			require.JSONEq(t, `"account-123"`, string(request.Variables["accountID"]))
			require.JSONEq(t, `{"id":"rule-456"}`, string(request.Variables["input"]))
		})
	}
}

type graphQLRequest struct {
	OperationName string                     `json:"operationName"`
	Query         string                     `json:"query"`
	Variables     map[string]json.RawMessage `json:"variables"`
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// Exercise the real SDK without opening sockets or calling Cato.
func testCatoAPI(t *testing.T, response string) (client *cato.Client, calls <-chan graphQLRequest) {
	t.Helper()
	requests := make(chan graphQLRequest, 1)
	client, err := cato.New("https://cato.invalid/graphql", "", "account-123",
		&http.Client{Timeout: time.Second, Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			var request graphQLRequest
			require.NoError(t, json.NewDecoder(req.Body).Decode(&request))
			requests <- request
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}},
				Body: io.NopCloser(strings.NewReader(response)), Request: req}, nil
		})}, nil)
	require.NoError(t, err)
	return client, requests
}
