# cato_mailing_list

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/account/mailinglist/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `CreateMailingListInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.notification.createMailingList` → SDK `NotificationCreateMailingList` (`sources/mutation.notification.createMailingList.gql`).
- `mutation.notification.updateMailingList` → SDK `NotificationUpdateMailingList` (`sources/mutation.notification.updateMailingList.gql`).
- `mutation.notification.deleteMailingList` → SDK `NotificationDeleteMailingList` (`sources/mutation.notification.deleteMailingList.gql`).
- `query.notification.mailingList` → SDK `NotificationMailingList` (`sources/query.notification.mailingList.gql`).

## Ownership, conversion and live-test requirements

- Own one mailing list and replace only its address/admin members; do not use incremental membership operations as the authoritative state. Creating this object must not send test email.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Verified API constraints

- At least one address or administrator must remain. Administrator inputs require IDs; the returned name is display metadata. Backend source: notification-service GraphqlValidatorUtil.validateCreateMailingListInput / validateUpdateMailingListInput.
- The SDK's unused audit.updatedBy projection fails for API-key actors. The adapter retains SDK transport/interceptors/response types but selects only managed fields; request-shape tests cover the projection and explicit empty membership arrays.
- Dev-US live acceptance passed create/read/import/update, clearing addresses while retaining an existing administrator, clearing administrators while retaining an address, no-drift plans, and deletion verification. No notification subscription was attached and no test email was sent.
- The first failed create's unique-name fixture was recovered and removed; subsequent test fixtures were destroyed and checked absent.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
