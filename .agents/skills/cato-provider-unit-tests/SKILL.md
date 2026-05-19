---
name: cato-provider-unit-tests
description: Write high-coverage Go unit tests for terraform-provider-cato resources, data sources, validators, plan modifiers, hydrators, and API payload translation. Use when agent is asked to add or update tests in this repository, especially for modified files under internal/provider, when SDK/API calls need mockery-generated interfaces, or when validating Terraform Plugin Framework Create/Read/Update/Delete/Import behavior without real Cato API calls.
---

# Cato Provider Unit Tests

## Goal

Add focused unit tests that maximize coverage for the modified provider behavior without hitting the real Cato API. Prefer fast tests beside `internal/provider` code, generated `testify/mock` clients from mockery, and direct Terraform Plugin Framework request/response objects.

Read `.agents/skills/cato-provider-unit-tests/references/current-testing-approach.md` when you need concrete local examples or are adding tests to a resource that makes SDK/API calls.

## Workflow

1. Identify the changed behavior first.
   - Run `git diff --name-only` and inspect modified files under `internal/provider`.
   - Map each changed branch to tests: success path, API error, validation error, null/unknown handling, missing remote object, state hydration, import, and payload conversion.
   - Prefer adding tests near the changed resource file, usually `internal/provider/resource_<name>_test.go`.

2. Check whether API calls are mockable.
   - If the resource already has an injected interface field, use the existing mock from `internal/provider/mocks`.
   - If the resource calls the SDK directly, add or extend a small interface in `internal/provider` that contains only the methods this resource needs, add the interface to `.mockery.yaml`, and run `go tool mockery`.
   - Commit generated mocks when they are needed by the tests. Do not hand-edit files under `internal/provider/mocks`.

3. Build tests around real resource methods.
   - Instantiate the concrete resource struct with `client: &catoClientData{AccountId: "account-123"}` and any injected mock client.
   - Use `resource.CreateRequest`, `resource.ReadRequest`, `resource.UpdateRequest`, `resource.DeleteRequest`, and `resource.ImportStateRequest` with synthetic `tfsdk.Plan`/`tfsdk.State`.
   - Seed plans and states through the resource schema and `plan.Set`/`state.Set`; avoid ad hoc `tftypes` unless a low-level plan modifier requires it.
   - Decode resulting state into the resource model and assert user-visible attributes, IDs, computed fields, and removed state.

4. Assert API payloads and call order where behavior matters.
   - Use `mocks.NewXxxClient(t)` and `mockClient.EXPECT().Method(...).Return(...).Once()`.
   - Use `mock.MatchedBy` for SDK input structs and assert important fields inside the matcher.
   - Return SDK response shapes that match the production code path closely enough to exercise hydration.
   - For validation-before-API behavior, create a mock client but set no expectations; the mock cleanup will fail if an API call happens.

5. Cover non-API logic directly.
   - Add table tests for validators, plan modifiers, semantic equality helpers, hydrators, conversion functions, and small pure helpers touched by the change.
   - Include null, unknown, empty, and valid values for Terraform framework types.
   - Use `t.Parallel()` for pure or independent tests; skip it when tests touch shared global state, environment variables, or httptest fixtures that are not isolated.

6. Verify narrowly, then broadly.
   - Run `go test ./internal/provider -run '<TestName>'` while iterating.
   - Run `go test ./internal/provider` before finishing provider-only changes.
   - Run `make test` when the change could affect shared provider behavior.
   - If mockery output changed, run `go tool mockery` again after final interface edits.

## Coverage Checklist

For a modified resource, aim to cover:

- `New<Resource>Resource`, `Metadata`, `Configure` with nil and valid provider data, and `ImportState` when implemented.
- `Create`: successful API path, payload shape, publish/read-after-write when present, API error, and validation error before API.
- `Read`: successful hydration, API error, and remote object missing or deleted from state.
- `Update`: changed fields, move/reorder behavior, publish/read-after-write when present, and API errors for each API step.
- `Delete`: successful delete, publish when present, API error, and idempotent/not-found behavior if production code supports it.
- Terraform framework edge cases: null, unknown, empty collections, defaulted/computed values, plan modifiers, validators, and semantic equality.

For a modified data source, aim to cover:

- Constructor/`Metadata`/`Configure` with nil and valid provider data.
- `Read`: successful API path and flattening/hydration of all user-visible attributes.
- `Read` diagnostics: API errors, validation errors, and missing/invalid remote objects.
- Terraform framework edge cases: null, unknown, empty collections, defaulted/computed values, and semantic equality where applicable.

Keep helper builders small and local to the test file unless the same model setup is already shared nearby. Use realistic account IDs, resource IDs, names, and SDK enum values, but keep fixtures minimal.
