# cato_secondary_aws_vsocket

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/network/secondaryawsvsocket/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `AddSecondaryAwsVSocketInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.site.addSecondaryAwsVSocket` → SDK `SiteAddSecondaryAWSVSocket` (`sources/mutation.site.addSecondaryAwsVSocket.gql`).
- `mutation.site.updateSecondaryAwsVSocket` → SDK `SiteUpdateSecondaryAWSVSocket` (`sources/mutation.site.updateSecondaryAwsVSocket.gql`).
- `mutation.site.removeSecondaryAwsVSocket` → SDK `SiteRemoveSecondaryAWSVSocket` (`sources/mutation.site.removeSecondaryAwsVSocket.gql`).
- `query.site.secondaryAwsVSocket` → SDK `SiteSecondaryAWSVSocket` (`sources/query.site.secondaryAwsVSocket.gql`).

## Ownership, conversion and live-test requirements

- site is replacement-only. Map eni_ip_address/eni_ip_subnet to ipAddress/subnet for update/read.
- This manages a secondary attachment in Cato and does not provision cloud VM/network infrastructure. Live tests require a dedicated compatible AWS site.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Live fixture prerequisite

The isolated acceptance fixture creates a logical, unregistered AWS primary. Enable `ADD_SECONDARY_ON_HA_SITE_CREATION` and explicitly set `TFACC_ALLOW_UNREGISTERED_SECONDARY_SITE=1` to run it. No AWS infrastructure is provisioned. Without that opt-in the test skips before creating the parent.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
