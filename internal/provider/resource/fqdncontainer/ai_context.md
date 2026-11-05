# cato_fqdn_container

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/resource/fqdncontainer/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `CreateFqdnContainerFromListInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.container.fqdn.createFromList` → SDK `ContainerFqdnCreateFromList` (`sources/mutation.container.fqdn.createFromList.gql`).
- `mutation.container.fqdn.updateFromList` → SDK `ContainerFqdnUpdateFromList` (`sources/mutation.container.fqdn.updateFromList.gql`).
- `mutation.container.delete` → SDK `ContainerDelete` (`sources/mutation.container.delete.gql`).
- `query.container.list` → exact ID + FQDN type filter, with focused SDK transport projection. `searchFqdn` is membership search and cannot read exact metadata.
- `mutation.container.fqdn.createFromURL` → SDK `ContainerFqdnCreateFromURL` (`sources/mutation.container.fqdn.createFromURL.gql`).
- `mutation.container.fqdn.updateFromURL` → SDK `ContainerFqdnUpdateFromURL` (`sources/mutation.container.fqdn.updateFromURL.gql`).
- `query.container.fqdn.downloadFile` → SDK `ContainerFqdnDownloadFile` (`sources/query.container.fqdn.downloadFile.gql`).

## Ownership, conversion and live-test requirements

- Support INLINE and URL modes with mutually exclusive values/sync_data. Mode and name are replacement-only; the available update APIs do not rename containers.
- INLINE mode owns normalized FQDN values and reads them through downloadFile.encodedFile, not metadata-only search.
- URL mode owns URL/sync configuration, not downloaded member values. file_type and sync_data are required only in URL mode. Require HTTPS and reject URL userinfo credentials.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

- Only containers owned by the configured account support lifecycle/import. Partner-inherited objects may be visible, but mutations remain scoped to the configured owner account and fail without ownership.
- URL mode is unit-tested only: live sync requires an authorized disposable notification destination, which was not supplied. No notifications were configured or sent.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
