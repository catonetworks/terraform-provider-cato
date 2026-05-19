# Current Testing Approach

## Local Patterns

- Unit tests live beside provider code in `internal/provider/*_test.go` and use `package provider`.
- Mocked API tests use generated `testify/mock` clients from `internal/provider/mocks`.
- Mockery is configured in `.mockery.yaml`:
  - `template: testify`
  - `dir: internal/provider/mocks`
  - `filename: "{{.InterfaceName}}.go"`
  - `pkgname: mocks`
  - interfaces are listed under `github.com/catonetworks/terraform-provider-cato/internal/provider`.
- Generate mocks with `go tool mockery` or `make mocks`.
- `make test` excludes `internal/acctests`; acceptance and `accmock` tests are a different layer and should not be the default answer for unit-test requests.

## Good Examples To Inspect

- `internal/provider/resource_internet_fw_rule_test.go`
  - Direct lifecycle tests for Create, Read, Update, Delete, Import, validators, and plan modifiers.
  - Uses `mocks.NewInternetFirewallPolicyClient(t)`.
  - Builds `tfsdk.Plan` and `tfsdk.State` from the resource schema.
  - Includes SDK response helpers such as `successfulAddRuleResponse` and policy response builders.

- `internal/provider/resource_site_socket_test.go`
  - Shows interface injection for `SocketSiteClient`.
  - Uses `mock.MatchedBy` to assert exact `cato_models.AddSocketSiteInput` payload fields.
  - Tests validation-before-API behavior by expecting diagnostics and setting no API expectations.
  - Covers helper functions, validators, null/unknown/empty values, and table-driven cases.

- `internal/provider/resource_license_utils_test.go`
  - Uses `httptest.Server` for lower-level API helper coverage when interface mocking is not the right fit.
  - Tracks mutation calls and asserts operation names and inputs.

- `internal/provider/resource_private_access_rule_test.go`
  - Simple table-driven tests for pure resource helper logic.

## Mockable Client Pattern

When production code needs to call the Cato SDK:

1. Define a narrow interface near the resource or in a focused client file.
2. Add an optional interface field to the resource struct.
3. Add a getter that returns the injected test client first and falls back to `r.client.catov2`.
4. Use the getter in resource methods instead of calling `r.client.catov2` directly.
5. Add the interface to `.mockery.yaml`.
6. Run `go tool mockery`.

Example shape:

```go
type SocketSiteClient interface {
	SiteAddSocketSite(ctx context.Context, input cato_models.AddSocketSiteInput, accountID string, interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.SiteAddSocketSite, error)
}

type socketSiteResource struct {
	client           *catoClientData
	socketSiteClient SocketSiteClient
}

func (r *socketSiteResource) getSocketSiteClient() SocketSiteClient {
	if r.socketSiteClient != nil {
		return r.socketSiteClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.catov2
}
```

## Terraform Framework Test Builders

Use the resource schema to build framework objects:

```go
func getExampleSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &exampleResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	return resp.Schema
}

func newExamplePlan(ctx context.Context, t *testing.T) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{Schema: getExampleSchema(ctx, t)}
	diags := plan.Set(ctx, ExampleModel{ID: types.StringNull()})
	if diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}
	return plan
}
```

Prefer model structs and `types.ObjectValueMust`, `types.SetNull`, `types.ListNull`, `types.StringNull`, and `types.StringUnknown` to raw JSON-like construction. Decode state with `state.Get(ctx, &model)` and nested objects with `.As(ctx, &nested, basetypes.ObjectAsOptions{})`.

## Verification Commands

- Focused package: `go test ./internal/provider`
- Focused test: `go test ./internal/provider -run 'TestSocketSiteCreateTranslatedSubnetPayload'`
- All non-acceptance tests: `go test $(go list ./... | grep -v terraform-provider-cato/internal/acctests)`
- Repo target: `make test`
- Mock generation: `go tool mockery`

If a command fails because generated mocks are stale, update `.mockery.yaml` if needed, run `go tool mockery`, then rerun the tests.
