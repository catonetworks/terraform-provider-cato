# cato_allocated_ip

Status: implementation and local checks complete; live testing blocked. Draft PR.

- Package: `internal/provider/network/allocatedip/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `PopLocationAddAllocatedIpInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.popLocationMutations.addAllocatedIp` → SDK `PopLocationMutationsAddAllocatedIP` (`sources/mutation.popLocationMutations.addAllocatedIp.gql`).
- `mutation.popLocationMutations.updateAllocatedIpDescription` → SDK `PopLocationMutationsUpdateAllocatedIPDescription` (`sources/mutation.popLocationMutations.updateAllocatedIpDescription.gql`).
- `mutation.popLocationMutations.removeAllocatedIp` → SDK `PopLocationMutationsRemoveAllocatedIP` (`sources/mutation.popLocationMutations.removeAllocatedIp.gql`).
- `query.popLocations.allocatedIpList` → SDK `PopLocationsAllocatedIPList` (`sources/query.popLocations.allocatedIpList.gql`).

## Ownership, conversion and live-test requirements

- type and pop_location are replacement-only; description is mutable. Use addAllocatedIp/removeAllocatedIp and the focused allocatedIpList read.
- Expose allocated IP address, subnet and state as computed outputs. Test only an allocation created for this run and verify release.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Public API limitations and live prerequisite

- Source-verified public lifecycle supports SYSTEM with an explicit PoP. BYOIP create/release require private account/IP inputs omitted from the released public SDK; reject BYOIP before mutation. Original schema apiDefault hints are not backend guarantees.
- Description updates/release resolve the current allocation by ID within the configured account and pass the observed IP, never the Terraform ID or stale state IP. Guard account, ID, type and IPv4 validity before mutation.
- All required local checks passed. Bugbot found no actionable findings. Zero reachable vulnerabilities; two module-only advisories.
- Live attempts created no allocation: original account is at its licensed IP limit; Dev-US PoPs 1011 and 51037 have no available addresses. Await a test account/PoP with capacity. Acceptance requires explicit TFACC_ALLOCATED_IP_POP_ID, tests only a new allocation, and verifies release. Existing allocations were read only and never changed.
