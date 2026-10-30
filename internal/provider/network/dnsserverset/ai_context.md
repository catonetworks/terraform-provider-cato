# cato_dns_server_set

Status: implementation and local checks complete; live testing blocked. Draft PR.

- Package: `internal/provider/network/dnsserverset/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `NetworkConfigDnsCreateServerSetInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.networkConfig.dns.createServerSet` → SDK `NetworkConfigDNSCreateServerSet` (`sources/mutation.networkConfig.dns.createServerSet.gql`).
- `mutation.networkConfig.dns.updateServerSet` → SDK `NetworkConfigDNSUpdateServerSet` (`sources/mutation.networkConfig.dns.updateServerSet.gql`).
- `mutation.networkConfig.dns.deleteServerSet` → SDK `NetworkConfigDNSDeleteServerSet` (`sources/mutation.networkConfig.dns.deleteServerSet.gql`).
- `query.networkConfig.dns.serverSet` → SDK `NetworkConfigDNSServerSet` (`sources/query.networkConfig.dns.serverSet.gql`).

## Ownership, conversion and live-test requirements

- Flatten the one-item batch create input into one Terraform resource. Use per-item create/update/delete, never setServerSet.
- Keep server as an ordered list to preserve resolver order.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Verified implementation and live blocker

- Unit/request-shape tests cover one-item batch ownership, ordered resolver updates, explicit empty-list clearing, CRUD/import, drift hydration, missing/malformed responses, null/unknown validation, and identity retention on partial create.
- Owning backend DnsServerSetService/ProfileQueryResolver confirms maximum two IPv4 resolvers, ordered lists including empty lists, and null for missing scoped IDs. Names cannot contain surrounding whitespace because the API trims them.
- Bugbot review completed with no actionable findings. All required local checks passed; zero reachable vulnerabilities and two module-only advisories.
- Live acceptance attempted on both account 10136462 and 62920; both reject at the networkConfig parent with the rollout-not-enabled message. No DNS set was created. PR remains draft until a test account with this API is available. User clarification is pending; other resources continue.
