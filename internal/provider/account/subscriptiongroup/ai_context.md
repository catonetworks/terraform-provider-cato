# cato_subscription_group

Status: implemented and verified on 2026-10-29; local checks, review and live acceptance passed.

- Package: `internal/provider/account/subscriptiongroup/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `CreateSubscriptionGroupInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.notification.createSubscriptionGroup` → SDK `NotificationCreateSubscriptionGroup` (`sources/mutation.notification.createSubscriptionGroup.gql`).
- `mutation.notification.updateSubscriptionGroup` → SDK `NotificationUpdateSubscriptionGroup` (`sources/mutation.notification.updateSubscriptionGroup.gql`).
- `mutation.notification.deleteSubscriptionGroup` → SDK `NotificationDeleteSubscriptionGroup` (`sources/mutation.notification.deleteSubscriptionGroup.gql`).
- `query.notification.subscriptionGroup` → SDK `NotificationSubscriptionGroup` (`sources/query.notification.subscriptionGroup.gql`).

## Ownership, conversion and live-test requirements

- Own one subscription group and its integration/mailing-list references. Live fixtures must not connect subscriptions that emit external notifications.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Verified API behavior

- The backend treats empty replacement collections as absent. Updates read current account-scoped membership and send explicit removals for cleared collections, including out-of-band members. Nonempty collections use full replacement.
- Creation requires a member; an existing group may be emptied. Both ID and name references are supported for mailing lists and integrations.
- Focused response projections omit audit and webhook configuration/auth fields while retaining complete member ID/name references. Resource does not own or modify referenced destinations.
- Dev-US live CRUD/read/import, name-reference update, complete membership clearing, restoration, no-drift plans and deletion passed. Two temporary unattached mailing lists were removed. No event producer was connected and no email/webhook test was sent. Integration clearing is covered by unit/request-shape tests.

## Verification

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed CRUD/read/import, update and no-drift plans; owned fixtures removed.
- Bugbot review and any required re-review completed with no remaining actionable findings.
