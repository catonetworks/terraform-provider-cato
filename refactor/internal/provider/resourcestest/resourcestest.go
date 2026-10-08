package resourcestest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/adapters/catoapi"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/adapters/memory"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/provider"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resources"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// GraphQLRequest contains the SDK request fields inspected by resource tests.
type GraphQLRequest struct {
	OperationName string                     `json:"operationName"`
	Variables     map[string]json.RawMessage `json:"variables"`
}

// ConfiguredResource configures a registered resource with the real SDK and an in-memory HTTP transport.
func ConfiguredResource(t *testing.T, name string, handle func(GraphQLRequest) string) (resource.Resource, schema.Schema) {
	t.Helper()
	ctx := context.Background()
	p := provider.NewProvider("test")()
	for _, factory := range p.Resources(ctx) {
		r := factory()
		var metadata resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
		if metadata.TypeName != name {
			continue
		}
		client, err := cato.New("https://cato.invalid/graphql", "", "account-123", &http.Client{
			Timeout: time.Second,
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				var request GraphQLRequest
				require.NoError(t, json.NewDecoder(req.Body).Decode(&request))
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}},
					Body: io.NopCloser(strings.NewReader(handle(request))), Request: req}, nil
			}),
		}, nil)
		require.NoError(t, err)
		var configured resource.ConfigureResponse
		r.(resource.ResourceWithConfigure).Configure(ctx, resource.ConfigureRequest{ProviderData: &resources.Dependencies{
			CatoAPI: catoapi.NewAdapter(client), AccountID: "account-123", ResourceLock: memory.NewResourceLock(),
		}}, &configured)
		require.False(t, configured.Diagnostics.HasError(), configured.Diagnostics)
		var schemaResponse resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
		return r, schemaResponse.Schema
	}
	t.Fatalf("resource %s is not registered", name)
	return nil, schema.Schema{}
}
