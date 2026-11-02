# cato_custom_application

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/resource/customapplication/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `AddCustomApplicationInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.customAppData.addCustomApplication` → SDK `CustomAppDataAddCustomApplication` (`sources/mutation.customAppData.addCustomApplication.gql`).
- `mutation.customAppData.updateCustomApplication` → SDK `CustomAppDataUpdateCustomApplication` (`sources/mutation.customAppData.updateCustomApplication.gql`).
- `mutation.customAppData.deleteCustomApplication` → SDK `CustomAppDataDeleteCustomApplication` (`sources/mutation.customAppData.deleteCustomApplication.gql`).
- `query.customAppData.customApplication` → SDK `CustomAppDataCustomApplication` (`sources/query.customAppData.customApplication.gql`).

## Ownership, conversion and live-test requirements

- Use unordered sets for OR criteria and their unordered members; preserve all typed protocol/port/destination fields.
- Use the focused customApplication read. Live test fixtures remain unattached to existing policies.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
