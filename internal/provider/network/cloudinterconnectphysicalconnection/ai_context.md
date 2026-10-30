# cato_cloud_interconnect_physical_connection

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/network/cloudinterconnectphysicalconnection/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `AddCloudInterconnectPhysicalConnectionInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.site.addCloudInterconnectPhysicalConnection` → SDK `SiteAddCloudInterconnectPhysicalConnection` (`sources/mutation.site.addCloudInterconnectPhysicalConnection.gql`).
- `mutation.site.updateCloudInterconnectPhysicalConnection` → SDK `SiteUpdateCloudInterconnectPhysicalConnection` (`sources/mutation.site.updateCloudInterconnectPhysicalConnection.gql`).
- `mutation.site.removeCloudInterconnectPhysicalConnection` → SDK `SiteRemoveCloudInterconnectPhysicalConnection` (`sources/mutation.site.removeCloudInterconnectPhysicalConnection.gql`).
- `query.site.cloudInterconnectPhysicalConnection` → SDK `SiteCloudInterconnectPhysicalConnection` (`sources/query.site.cloudInterconnectPhysicalConnection.gql`).

## Ownership, conversion and live-test requirements

- site and ha_role are replacement-only because the update contract does not accept them. Other creation attributes are mutable when present in the update API.
- Hydrate QinQ c_vlan/s_vlan from the flat read projection. Use numeric VLANs with validation and encapsulation-dependent configuration.
- Use a dedicated test site, permitted PoP and non-overlapping test addressing; do not attach a production circuit.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Proposed schema amendment, awaiting approval

- Change only vlan and qinq_vlan_configuration from optional to optional_computed so API-assigned tags can hydrate/import without inconsistent state. Neither field is required for automatic allocation. Explicit configuration still requires mode-correct fields and VLAN range 1..4094.
- The existing SDK read maps numeric NetworkBandwidth to string; a resource-local adapter preserves the SDK document and transport and normalizes only the two bandwidth fields before typed decoding. Request-shape/normalization tests pass.
- Backend automatically allocates VLANs when tags are omitted. Custom VLAN inputs require SUPPORT_INTERCONNECT_VLAN_CONFIGURATION; global VLAN-inventory access additionally requires a reseller. The ordinary test account cannot read global reservations. No physical connection/site fixture has been created for this resource yet.

## Approved schema amendment

The user approved optional+computed `vlan` and `qinq_vlan_configuration` on 2026-10-29. Configuration presence determines ownership: omit auto-assigned tags from mutation inputs, retain allocated active tags in Read/import state, and preserve fully known partial identity after read/publication failures. Custom tags retain range/mode validation and require SUPPORT_INTERCONNECT_VLAN_CONFIGURATION; automatic allocation does not require that custom-tag feature.

## Review fixes and live boundary

- Encapsulation changes fail before mutation and require explicit replacement: frontend fills omitted prior tags, and the GraphQL validated scalar prevents zero-clearing. No zero-tag requests are sent.
- Equivalent provider-name casing is preserved with EqualFold convergence; genuinely different remote providers remain visible as drift.
- Isolated mock PoP 10013 passed live create/read/import/BW update/no-drift/delete and parent cleanup. Mock 10012 lacked capacity; no connection was created there and its parent was removed. Movement, QINQ and custom-tag live checks remain unverified.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
