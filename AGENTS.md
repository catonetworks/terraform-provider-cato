# Cato Terraform Provider — Agent Guide

## Project Overview
This repository builds the Terraform provider for Cato Networks, published as
`registry.terraform.io/catonetworks/cato`. It is a Go module using HashiCorp's
Terraform Plugin Framework and the Cato Go SDK from `github.com/catonetworks/cato-go-sdk` to be able to call Cato public APIs.

## Architecture
- Each Terraform resource has a full CRUD implementation.
- The top-level Terraform model for a resource is defined in `model_*.go` with
  strongly-typed fields. Nested objects are mapped via framework attribute types
  (`attrTypes` maps / `ObjectAsOptions`), not plain Go structs - exception here are well defined shared types, e.g. Member in `shared/groupmembers/model_group_member.go`.
- `resource_<name>.go` owns Terraform lifecycle orchestration, diagnostics, and state persistence.
- `converters_<name>.go` owns Terraform model ↔ SDK input/response conversion, including nested values and null/unknown handling.
- `client_<name>.go` owns the resource's narrow `ResourceClient` interface and SDK forwarding adapter. Resources accept an injected `apiClient`; normal execution uses the configured provider SDK. Keep SDK access, API lookup helpers, and retries behind this boundary.
- Keep publish lifecycle decisions and policy-publication diagnostics in resource files; client adapters forward the SDK requests.
- Converters must propagate diagnostics through a pointer or an explicit return value; callers must retain them in the Terraform response.
- Keep shared conversions and policy-family helpers in their existing shared packages; pass narrow client interfaces through shared helpers that perform API calls. Preserve account scoping, caching, retries, and request shapes.

## Structure
- `main.go` — provider entry point, wires `internal/provider.New(version)`.
- `internal/provider/` — provider config and registration, with one resource per
  package under the existing domain folders (`account/`, `access/`, `network/`,
  `policy/`, and `resource/`). Keep each resource implementation, model, schema,
  converters, client adapters, and tests together in its package. Tests belong beside the
  implementation they exercise; split files that test multiple resources.
  - Data sources and their models belong in `<domain>/data_sources/<name>/`,
    separate from resource implementation packages.
  - Policy families use resource-specific child packages, for example
    `policy/tls/tlsrule/`, `tlssection/`, and `tlsindex/`.
  - Share policy-family helpers in the parent package and cross-domain helpers
    in `shared/`; shared packages must not import resource implementations.
  - `resource_*.go` — resource CRUD implementations
  - `resource_*_test.go` — unit test for resource implementations
  - `datasource_*.go` — data source read implementations
  - `datasource_*_test.go` — data source for resource implementations
  - `model_*.go` — Terraform model types
  - `schema_*.go` — resource Schema methods, schema builders, and attribute type maps
  - `converters_<resource>.go` — Terraform model ↔ SDK conversion helpers
  - `client_<resource>.go` — narrow mockable API interface, SDK adapter, and API helpers
  - `validators_<resource>.go` — all resource-specific validators and plan modifier implementations in one file
  - `validators_<resource>_test.go` — tests for those validators and plan modifiers
  - `validators/` — reusable validators and validation helpers; filenames have no `validators_` prefix
  - `planmodifiers/` — reusable plan modifiers and planning helpers
  - `mocks/` — mockery-generated mocks (do not edit manually)
- `internal/acctests/` — acceptance tests that call real Cato APIs.
- `internal/accmock/` and `test_data/` — mocked API integration tests.
- `examples/` — Terraform examples per resource and data source.
- `templates/` — provider documentation templates (used by `tfplugindocs`).
- `docs/` - generated provider docs output

## Setup
- Go `1.26.10` (see `go.mod`).
- Terraform CLI for running examples and acceptance tests.
- Required env vars for API access: `CATO_BASEURL`, `CATO_TOKEN`, provider
  `account_id`. Examples also accept `TF_VAR_token` / `TF_VAR_account_id`.
- Optional retry tuning: `CATO_RETRY_MAX`, `CATO_RETRY_WAIT_MIN_SECONDS`,
  `CATO_RETRY_WAIT_MAX_SECONDS`.

## Agentic loop for implementing changes
- Start with unit tests for behavioral changes; get user sign-off if requested.
- After implementation, iterate on unit tests until green, then add/update acceptance tests when API behavior changed, run lint and vulnerability check, update docs and examples. Don't run the acceptance tests in this loop unless specifically asked.

## Definition of Done for agentic loop
- Change is implemented with clear, maintainable code that follows existing patterns in internal/provider/.
- Unit tests are added/updated for the behavior change and pass (make test).
- Acceptance tests are added/updated when real API behavior changed.
- Imports and style checks pass (make sort-imports, make lint).
- Vulnerability check passes
- Documentation is updated and generated (make docs) to reflect the final behavior.
- Example configuration in examples/ is updated to match the implementation and docs.
- New/changed resources or data sources are registered in `internal/provider/provider.go`.
- No secrets, state files, debug logs, or manual edits to generated mocks are included.
- Scope is minimal and focused; no unrelated refactors or behavior changes.
- Final verification summary is provided to the user (what changed, tests run, any residual risks).

## Commands

```sh
make build          # compile terraform-provider-cato
make install        # build + install dev provider override (~/.terraformrc-dev)
make install-mirror # build + install filesystem mirror (needed for terraform test)
make test           # run unit tests (excludes acceptance tests)
make lint           # run golangci-lint (includes acctest build tag)
make sort-imports   # run goimports + gci on internal/
make docs           # regenerate provider docs with tfplugindocs
make mocks          # regenerate mockery mocks
make vul            # vulnerability check with govulncheck
make acctest        # run acceptance tests (real API — see Testing section)
```

Target a single package or test during development:

```sh
go test ./internal/provider/... -run TestName
```

## Testing

**Unit tests** live alongside the code in `internal/provider/`. They mock Cato
backend API calls and cover business logic and Terraform state mapping. Run with
`make test`.

**Acceptance tests** in `internal/acctests/` call real Cato APIs and create real
resources. Run with:

```sh
make acctest
```

To load credentials and account settings from a local shell file (not committed
to the repo), run `source /path/to/your-env.sh` in the same shell before
`make acctest` or a targeted `go test -tags=acctest ./internal/acctests/...`.

`make acctest-flaky` runs `test_data/flaky_acctest.sh`, which retries a package on
transient failures (including `reorderPolicyBlockedByActiveSessions` when the API
reports concurrent policy revisions). Pass `t=<suite>` to run one directory under
`internal/acctests/` (basename only, e.g. `t=wf_rules_index` or `t=wf_rules_index_with_rule_data`). The script exports
`DISABLE_POLICY_RULE_CLEANUP=true` (same as `make acctest`).

Set `DISABLE_POLICY_RULE_CLEANUP=true` when running acceptance tests to prevent
the provider from discarding draft policy revisions during `Configure()` (already
included in `make acctest`).

## Import Ordering
Imports must follow the order enforced by `make sort-imports`:
1. Standard library
2. Third-party packages
3. `github.com/softopus-io/*` packages
4. Local module packages

Violating this order will cause `make lint` to fail.

## Adding a New Resource
1. `internal/provider/<domain>/<name>/model_<name>.go` — define the Terraform model.
   Put the resource `Schema` method and schema builders in `schema_<name>.go` in the same package.
2. `internal/provider/<domain>/<name>/resource_<name>.go` — implement `Create`, `Read`, `Update`, `Delete`, `ImportState`.
3. `internal/provider/<domain>/<name>/converters_<name>.go` — Terraform state → API input conversion (if non-trivial).
4. `internal/provider/<domain>/<name>/client_<name>.go` — define the narrow API interface and SDK adapter; keep response-to-state conversion in `converters_<name>.go`.
5. Register in `internal/provider/provider.go` under `Resources()`.
6. Add `examples/resources/cato_<name>/resource.tf` — complete example using all available parameters; comment out optional ones.
7. Add `templates/resources/<name>.md.tmpl` if custom docs are needed; otherwise `tfplugindocs` generates them automatically.
8. Run `make docs` and `make sort-imports` before committing.

## Development Notes
- When bumping the provider version, update it in both `GNUmakefile` and `main.go` (the `version` variable).
- Use the same example from `examples/` in the corresponding `*.md`
- Make an example complete, self-explanatory and use maximum number of available parameters, comment out the optional ones (see example in `.agents/examples/example.tf`)

## Never
- Commit Terraform state files, API tokens, debug logs, or compiled provider binaries.
- Commit secrets or sensitive internal information.
- Edit files under `internal/provider/mocks/` by hand — regenerate with `make mocks`.

## API-First Provider Contract

Treat SDK regeneration fallout as a compatibility task, not permission for broad provider rewrites.

- Start from clean provider `main`; never reuse a dirty migration experiment.
- Pin the candidate SDK commit pseudo-version only for branch/PR testing. Replace it with the released SDK tag before provider merge.
- Localize added optional generated arguments in one narrow adapter when practical.
- Change resource logic directly only for genuine API behavior or generated response-shape changes.
- Preserve AccountSnapshot account scoping and verify request/header shape when its generated arguments change.
- Use an explicit, tested `GroupMembersListInput`; do not assume zero-value semantics without request-shape coverage.
- Preserve Socket LAN section `subPolicyID` in the SDK projection. Never infer empty-section ownership from rule rows.
- For policy revision discard, send the same explicit revision ID in discard and policy-mutation inputs; keep request-shape tests.
- Regenerate affected mocks only with pinned `make mocks`. Never hand-edit files under `internal/provider/mocks/`.

### Required Checks

- Add focused unit/request-shape tests for changed behavior or forwarding.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, and `make vul`.
- Do not run acceptance tests unless explicitly requested.
- Keep the provider PR dependent on the exact candidate SDK commit and unmerged until the SDK dependency is released.

## Skills
Find all skills in `.agents/skills/`
- Use `prepare-release/SKILL.md` to prepare the release of a new version of the terraform provider
- Use `publish-release/SKILL.md` to publish the release of a new version of the terraform provider
- Use `send-release-info/SKILL.md` to announce a release on Slack (#rn-kb-opensource-api-announcements + #management-integrations)
- Use `cato-provider-unit-tests/SKILL.md` to prepare unit tests
- Use `integration-test-mock/SKILL.md` to prepare a mock of an integration test
