# cato_user

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/access/user/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `CreateUserInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.user.createUser` → SDK `UserCreateUser` (`sources/mutation.user.createUser.gql`).
- `mutation.user.updateUser` → SDK `UserUpdateUser` (`sources/mutation.user.updateUser.gql`).
- `mutation.user.deleteUser` → SDK `UserDeleteUser` (`sources/mutation.user.deleteUser.gql`).
- `query.user.userList` → SDK `UserUserList` (`sources/query.user.userList.gql`).
- `mutation.user.enableUser` → SDK `UserEnableUser` (`sources/mutation.user.enableUser.gql`).
- `mutation.user.disableUser` → SDK `UserDisableUser` (`sources/mutation.user.disableUser.gql`).

## Ownership, conversion and live-test requirements

- Own local users only; reject imports/updates of directory-managed users. Paginate userList and compare exact IDs.
- Map phone_number to the phone response. Manage enabled through the separate enableUser/disableUser APIs; default true. Disable live fixtures and do not grant remote-access eligibility.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

Ownership is enforced before update/delete and during import refresh. LOCKED users remain enabled; no API operation grants remote-access eligibility.

API clearing requires explicit null rather than empty strings. The narrow update adapter uses the SDK operation constant and original transport to emit nulls for the three optional profile fields; create omits empty optional fields. Live tests cover clearing and keep their owned user disabled. Enabling is unit-tested only.

Create preserves known identity before enable/disable follow-up failures and also on partial SDK responses with errors. Unknown computed fields become null in partial error state.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
