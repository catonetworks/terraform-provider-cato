# Design: a Terraform provider that company teams can extend

This document describes the implementation under `refactor/` and the intended
contribution model. The company-wide ownership rules and CI enforcement below
are proposals. The implementation is still a POC: private-access rule
import/read/delete and static-host creation are the working examples, not full
production lifecycle support. It uses the repository's existing Go module.

**TL;DR:** Separate Terraform integration, application workflows, and Cato SDK
calls so API-owning teams can deliver resources independently, with the provider
team involved through code review. Provider initialization supplies shared
dependencies; resource methods create their use cases at runtime and invoke
`Execute(ctx, command)`.

## Improvements over the previous implementation

- **Maintainability:** Resource and domain files give changes a clear home,
  reducing how much unrelated code contributors must understand or modify.
- **Separation of concerns:** `internal/provider/` handles Terraform schema,
  state, and wiring; `internal/application/` owns business workflows;
  `internal/adapters/cato_api/` wraps SDK operations and maps API data.
- **Explicit error handling at each layer:** The adapter preserves API failure
  details in `CatoAPIError`; the application reads `CatoAPIAdapterError` and
  classifies failures into `UseCaseError`; the provider translates outcomes into
  Terraform diagnostics through `tfdiagnostics.go`.
- **Independent testing:** Small application interfaces allow workflow tests
  with mocks before adapters exist. Adapter and resource tests separately verify
  API mapping and Terraform behavior.
- **Clear workflow behavior:** Sequencing, bounded retries, cancellation, and
  partial completion are application decisions, rather than being mixed with
  SDK responses and Terraform state handling.
- **Collaboration:** Domain teams can own a resource from API integration to
  Terraform support. The provider team reviews contributions and maintains the
  shared foundation instead of implementing every integration.
- **Controlled complexity:** Simple operations stay small. ID-only calls use
  parameters, richer inputs reuse application entities, and new workflows do
  not require a central service or a dependency field per operation.

## Why this matters for collaboration

**The principal advantage is removing the provider team from the implementation
path for every new API feature.** Contributors have a repeatable route from an
API capability to a reviewed Terraform resource.

| Advantage | What makes it possible | Effect on teams |
| --- | --- | --- |
| Independent delivery | A resource owns its Terraform integration; a use case owns its workflow; resource-specific adapter methods wrap the SDK. | The API-owning team can submit a complete implementation without a handoff to the provider team. |
| Parallel development | Use cases depend on small interfaces and can be tested with mocks before their adapters exist. | Domain work can merge earlier while adapter and Terraform work continue. |
| Focused reviews | Business decisions, API mapping, and Terraform state handling are separated. | Reviewers can assess each concern without untangling all three in one method. |
| Less shared-file contention | Workflows are constructed inside resource methods; bootstrap holds only shared dependencies. | Adding a workflow does not require editing a central service or adding a dependency field for each operation. |
| Faster feedback | Application tests run without Terraform or live APIs; adapter and resource tests cover their respective boundaries. | Teams can verify sequencing, errors, and mappings locally before integration testing. |
| More contained changes | SDK types stay at the API boundary and Terraform types stay at the Terraform boundary. | An SDK response change or framework change has a clearer place to be handled. |
| More reliable workflows | Sequencing, retry decisions, and completion checks are explicit application behavior. | Teams can test partial failures and concurrency decisions rather than burying them in CRUD methods. |
| Enforceable conventions | Dependency direction and extension locations can be checked in CI. | Contributors get consistent feedback before review, and shared infrastructure is harder to turn into a catch-all. |

These benefits depend on preserving the boundaries. Organizing the same mixed
methods into domain folders improves navigation, but does not provide independent
workflow tests or isolate Terraform behavior from SDK changes.

### A team's path from a new API resource to Terraform support

The API-owning team delivers the complete contribution:

1. Define commands, entities, and the adapter interfaces needed by the workflows;
   implement and unit-test the application behavior with mocks.
2. Add SDK adapter methods and the Terraform resource's schema, lifecycle
   methods, diagnostics, and tests.
3. Complete lifecycle/import support, examples, documentation, and appropriate
   acceptance coverage; register the ready resource in `Resources()`.
4. Submit for provider-team review of integration, compatibility, and conventions.

Domain code can merge earlier with passing tests while the adapter and Terraform
resource are unfinished. Keep the resource unregistered until its intended
lifecycle is ready, so existing provider releases can continue independently.

### Protect the foundation with CI

Proposed CI rules would keep contributions within the intended extension points:

- **Enforce dependencies:** Application code must not depend on Terraform, the
  SDK, or concrete adapters. Resource methods use application workflows; adapters
  handle API operations without Terraform dependencies or workflow orchestration.
- **Protect bootstrap:** Require maintainer review for changes to initialization
  and shared dependencies. Allow routine resource imports and registration;
  reject resource-specific workflows added to bootstrap or central services.
- **Verify contributions:** Run compilation, layer-specific unit tests, lint,
  and relevant compatibility checks, including `refactor/`. Route domain/resource
  changes to their owners and shared infrastructure changes to provider maintainers.

CI should reject violations with actionable feedback, rather than silently
remove code. Automated checks support code review; they cannot establish every
business boundary on their own.
