# cato_dns_suffix_set

Status: implementation and local checks complete; live testing blocked. Draft PR.

- Package: `internal/provider/network/dnssuffixset/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `NetworkConfigDnsCreateSuffixSetInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.networkConfig.dns.createSuffixSet` → SDK `NetworkConfigDNSCreateSuffixSet` (`sources/mutation.networkConfig.dns.createSuffixSet.gql`).
- `mutation.networkConfig.dns.updateSuffixSet` → SDK `NetworkConfigDNSUpdateSuffixSet` (`sources/mutation.networkConfig.dns.updateSuffixSet.gql`).
- `mutation.networkConfig.dns.deleteSuffixSet` → SDK `NetworkConfigDNSDeleteSuffixSet` (`sources/mutation.networkConfig.dns.deleteSuffixSet.gql`).
- `query.networkConfig.dns.suffixSet` → SDK `NetworkConfigDNSSuffixSet` (`sources/query.networkConfig.dns.suffixSet.gql`).

## Ownership, conversion and live-test requirements

- Flatten the one-item batch create input into one Terraform resource. Use per-item create/update/delete, never setSuffixSet.
- Keep suffix as an ordered list to preserve search order.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Verification and rollout prerequisite

- Local unit/request-shape tests cover order changes, case preservation, duplicates, null/unknowns, whitespace and Java UTF-16 length limits, explicit empty-list clearing, single-item batch ownership, CRUD/import/drift/missing/malformed responses and partial-create identity.
- All required local checks passed. Bugbot found no actionable findings. Zero reachable vulnerabilities; two module-only advisories.
- Live CRUD/import/no-drift testing requires networkConfig rollout on the target account. Both supplied development accounts reject the parent API, as confirmed during DNS Server Set validation; no suffix-set live mutation was attempted. User explicitly requested PRs with the rollout prerequisite in their descriptions. Leave draft until live validation can pass.
