# cato_ip_address_range_container

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/resource/ipaddressrangecontainer/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `CreateIpAddressRangeContainerFromListInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.container.ipAddressRange.createFromList` → SDK `ContainerIPAddressRangeCreateFromList` (`sources/mutation.container.ipAddressRange.createFromList.gql`).
- `mutation.container.ipAddressRange.updateFromList` → SDK `ContainerIPAddressRangeUpdateFromList` (`sources/mutation.container.ipAddressRange.updateFromList.gql`).
- `mutation.container.delete` → SDK `ContainerDelete` (`sources/mutation.container.delete.gql`).
- `query.container.ipAddressRange.searchIpAddressRange` → SDK `ContainerIPAddressRangeSearchIPAddressRange` (`sources/query.container.ipAddressRange.searchIpAddressRange.gql`).
- `mutation.container.ipAddressRange.createFromURL` → SDK `ContainerIPAddressRangeCreateFromURL` (`sources/mutation.container.ipAddressRange.createFromURL.gql`).
- `mutation.container.ipAddressRange.updateFromURL` → SDK `ContainerIPAddressRangeUpdateFromURL` (`sources/mutation.container.ipAddressRange.updateFromURL.gql`).
- `query.container.ipAddressRange.downloadFile` → SDK `ContainerIPAddressRangeDownloadFile` (`sources/query.container.ipAddressRange.downloadFile.gql`).

## Ownership, conversion and live-test requirements

- Support INLINE and URL modes with mutually exclusive values/sync_data. Mode and name are replacement-only; the available update APIs do not rename containers.
- INLINE mode owns normalized from/to ranges and reads them through downloadFile.encodedFile, not metadata-only search.
- URL mode owns URL/sync configuration, not downloaded member values. file_type and sync_data are required only in URL mode. Require HTTPS and reject URL userinfo credentials.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

- Exact account/ID and IP_RANGE-filtered metadata lookup uses the released SDK transport with focused projections; searchIPAddressRange is membership search, not exact metadata lookup. INLINE reads bounded base64 UTF-8 downloads containing singleton IPv4 or inclusive from-to lines. Overlapping configured ranges are rejected because the backend merges them; adjacent disjoint ranges are permitted.
- Only containers owned by the configured account support lifecycle/import; inherited partner containers are unsupported. Mutation scope never expands to a parent account.
- URL mode is unit-tested only: no authorized disposable notification destination was supplied, so no notifications were configured or sent.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
