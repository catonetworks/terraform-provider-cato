# cato_cloud_interconnect_site

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/network/cloudinterconnectsite/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `AddCloudInterconnectSiteInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.site.addCloudInterconnectSite` → SDK `SiteAddCloudInterconnectSite` (`sources/mutation.site.addCloudInterconnectSite.gql`).
- `mutation.site.updateSiteGeneralDetails` → SDK `SiteUpdateSiteGeneralDetails` (`sources/mutation.site.updateSiteGeneralDetails.gql`).
- `mutation.site.removeSite` → SDK `SiteRemoveSite` (`sources/mutation.site.removeSite.gql`).
- `query.site.siteGeneralDetails` → SDK `SiteGeneralDetails` (`sources/query.site.siteGeneralDetails.gql`).

## Ownership, conversion and live-test requirements

- Own site metadata only; never adopt or remove existing physical connections. Map site_location.city to cityName for update/read.
- Use a dedicated new site for testing. Creation registers Cato site configuration; physical connections are managed by the separate resource.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Ownership and verification

- Account-scoped exact-ID configuration lookup determines existence. A separately scoped AccountSnapshot verifies CROSS_CONNECT; an unavailable snapshot for an existing site preserves state. Snapshot request shape explicitly forwards configured account ID and singleton site ID.
- Create persists a fully known identity before read convergence; update waits for requested metadata, maps create city to update/read cityName, and omits unrelated site settings.
- Delete checks both physical HA roles and fails closed on lookup errors or attached connections. Backend deletion also guards active connections. This resource never deletes physical connections.
- All required local checks passed. Bugbot final review found no actionable findings. Zero reachable vulnerabilities; two module-only advisories.
- Dev-US live acceptance passed create/read/import, metadata/location updates, description clearing and no-drift plans. Owned disconnected test sites were deleted; final cleanup verifies absence using authoritative scoped configuration lookup. No physical connection or existing site was modified.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
