# cato_pop_bgp_profile

Status: implementation and local checks complete; live testing blocked. Draft PR.

- Package: `internal/provider/network/popbgpprofile/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `PopLocationAddBgpProfileInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.popLocationMutations.addBgpProfile` → SDK `PopLocationMutationsAddBgpProfile` (`sources/mutation.popLocationMutations.addBgpProfile.gql`).
- `mutation.popLocationMutations.updateBgpProfile` → SDK `PopLocationMutationsUpdateBgpProfile` (`sources/mutation.popLocationMutations.updateBgpProfile.gql`).
- `mutation.popLocationMutations.removeBgpProfile` → SDK `PopLocationMutationsRemoveBgpProfile` (`sources/mutation.popLocationMutations.removeBgpProfile.gql`).
- `query.popLocations.bgpProfile` → SDK `PopLocationsBgpProfile` (`sources/query.popLocations.bgpProfile.gql`).

## Ownership, conversion and live-test requirements

- Use an unordered set of community from/to pairs; ASN scalar values remain strings. Use the focused bgpProfile read.
- The live test owns an unattached profile and must not attach it to an existing BGP session.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Verification and feature prerequisite

- Required local checks passed; zero reachable vulnerabilities and two module-only advisories. Bugbot found no actionable findings.
- Unit/request-shape tests cover CRUD/import, canonical ASN16 pairs, null/unknown values, missing/malformed responses, scoped lookups, metadata-only omission of unchanged communities, description clearing, and partial-create identity.
- Owning BgpProfileService disallows community updates on attached profiles. A fresh scoped read allows unchanged communities to be omitted during metadata-only updates.
- Both supplied development accounts reject create with “missing account tag”; frontend source requires BGP_COMMUNITIES_PROFILE. No profile was created or attached. Live CRUD/import/no-drift/deletion remains pending a feature-enabled test account; keep PR draft and state this prerequisite.
