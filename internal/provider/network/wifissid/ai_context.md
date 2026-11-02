# cato_wifi_ssid

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/network/wifissid/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `CreateWifiSsidInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.site.createWifiSsid` → SDK `SiteCreateWifiSsid` (`sources/mutation.site.createWifiSsid.gql`).
- `mutation.site.updateWifiSsid` → SDK `SiteUpdateWifiSsid` (`sources/mutation.site.updateWifiSsid.gql`).
- `mutation.site.removeWifiSsid` → SDK `SiteRemoveWifiSsid` (`sources/mutation.site.removeWifiSsid.gql`).
- `query.site.wifiSsid` → SDK `SiteWifiSsid` (`sources/query.site.wifiSsid.gql`).

## Ownership, conversion and live-test requirements

- Secret state must never appear in logs, request-shape fixtures, documentation or error request bodies.
- site and network_interface_id are replacement-only. Exclude vlan_id: the source coverage document identifies it as unsupported even though the SDL still declares it.
- PSK passkey is sensitive configuration, never computed. Retain the configured value during refresh because the API does not return it; import cannot recover it and external PSK-only drift cannot be detected.
- Validate conditional PSK settings. Redact SDK request bodies from secret-bearing errors. Live testing requires a designated Wi-Fi-capable test socket/interface.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## API compatibility and verification notes

- The local SSID list is not authoritative for legacy CC2-only configuration. Resolve and pin the account-scoped site ID, then page CC2-backed networkInterface lookups with explicit parent/account and require consistent site/port identities before creation. Occupied WBR interfaces fail closed.
- Exact ID-bound stale-CC2 errors reconcile missing configuration on Read; Delete still removes the local row and validates retained site/interface/ID.
- dhcp_enabled is deprecated and only false is supported. Omit it from requests because the API rejects even false; DHCP reset sends an empty settings object to restore account default. VLAN remains excluded.
- Remove uses the same SDK operation/variables/transport with an identity-only projection: transient subnet/localIp fields become null after deletion and would otherwise cause GraphQL errors after a successful mutation.
- PSKs remain sensitive noncomputed configuration. Read preserves known PSK state, import cannot recover it, and external PSK-only drift is undetectable. Never forward backend message text for PSK-bearing mutations, because it may echo the passkey in an arbitrary encoding.
- Dev-US Wi-Fi rollout was unavailable. Its owned temporary parent was removed and absence verified; no SSID was created there. On staging, live create/read, import (excluding unrecoverable PSK), name/band/DHCP updates, DHCP reset, PSK update, empty drift plans, deletion and owned parent cleanup passed. No existing socket or broadcast network was changed; SSID remained disabled.
- Bugbot findings on legacy-interface overwrite and stale-row recovery were fixed and re-reviewed cleanly; the live Delete projection fix was also reviewed.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
