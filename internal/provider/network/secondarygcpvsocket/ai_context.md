# cato_secondary_gcp_vsocket

Status: implementation and local checks complete; live testing blocked. Draft PR.

- Package: `internal/provider/network/secondarygcpvsocket/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `AddSecondaryGcpVSocketInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.site.addSecondaryGcpVSocket` → SDK `SiteAddSecondaryGCPVSocket` (`sources/mutation.site.addSecondaryGcpVSocket.gql`).
- `mutation.site.updateSecondaryGcpVSocket` → SDK `SiteUpdateSecondaryGCPVSocket` (`sources/mutation.site.updateSecondaryGcpVSocket.gql`).
- `mutation.site.removeSecondaryGcpVSocket` → SDK `SiteRemoveSecondaryGCPVSocket` (`sources/mutation.site.removeSecondaryGcpVSocket.gql`).
- `query.site.secondaryGcpVSocket` → SDK `SiteSecondaryGCPVSocket` (`sources/query.site.secondaryGcpVSocket.gql`).

## Ownership, conversion and live-test requirements

- site is replacement-only. Hydrate gcp_config from the flat interfaceIp/loadBalancerIp response.
- This manages a secondary attachment in Cato and does not provision cloud VM/network infrastructure. Live tests require a dedicated compatible GCP site and account rollout.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Live validation boundary

A temporary logical GCP primary and secondary were created on the dev-US account. Create returned the new socket ID and the provider retained it, but the typed read rejected the unregistered secondary as not a GCP socket. Source also requires REGISTERED for updates. The exact owned secondary was removed via the SDK after account/site/role/configuration ownership checks; the parent was removed and absence verified. No VM was provisioned.

Full read/import/update/no-drift/delete acceptance now requires an explicitly disposable registered secondary, `TFACC_GCP_SECONDARY_ALLOW_MUTATIONS=1`, `TFACC_GCP_SECONDARY_SOCKET_ID`, and two unused native-range addresses in `TFACC_GCP_INTERFACE_IP`/`TFACC_GCP_LOAD_BALANCER_IP`. This test imports, updates and deletes the designated secondary, retaining its parent. It cannot validate successful creation/registration without externally provisioned GCP infrastructure. Leave the PR in draft pending this validation.

Read-only missing-ID probes on both dev accounts returned the exact account-scoped `404 Not Found: "SOCKET with id [0] not found"` GraphQL message. Missing classification recognizes that exact ID-bound message and the service's explicit not-found message, never generic HTTP 404s.
