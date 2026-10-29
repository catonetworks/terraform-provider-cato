# cato_terminal_server_section

Status: implemented, reviewed and live-tested on 2026-10-29.

- Package: `internal/provider/policy/access/terminalserversection/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `PolicyAddSectionInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.policy.terminalServer.addSection` → SDK `PolicyTerminalServerAddSection` (`sources/mutation.policy.terminalServer.addSection.gql`).
- `mutation.policy.terminalServer.updateSection` → SDK `PolicyTerminalServerUpdateSection` (`sources/mutation.policy.terminalServer.updateSection.gql`).
- `mutation.policy.terminalServer.removeSection` → SDK `PolicyTerminalServerRemoveSection` (`sources/mutation.policy.terminalServer.removeSection.gql`).
- `query.policy.terminalServer.policy` → SDK `PolicyTerminalServerPolicy` (`sources/query.policy.terminalServer.policy.gql`).
- `mutation.policy.terminalServer.moveSection` → SDK `PolicyTerminalServerMoveSection` (`sources/mutation.policy.terminalServer.moveSection.gql`).
- `mutation.policy.terminalServer.createPolicyRevision` → SDK `PolicyTerminalServerCreatePolicyRevision` (`sources/mutation.policy.terminalServer.createPolicyRevision.gql`).
- `mutation.policy.terminalServer.publishPolicyRevision` → SDK `PolicyTerminalServerPublishPolicyRevision` (`sources/mutation.policy.terminalServer.publishPolicyRevision.gql`).
- `mutation.policy.terminalServer.discardPolicyRevision` → SDK `PolicyTerminalServerDiscardPolicyRevision` (`sources/mutation.policy.terminalServer.discardPolicyRevision.gql`).

## Ownership, conversion and live-test requirements

- Use a unique named revision for each apply; all mutations/discards must use its exact explicit ID. Publish automatically after successful mutation/movement.
- Read PUBLIC policy explicitly and keep refresh observational. Retain resource identity on failed create publication and prior published state on failed updates. Reconcile lost publication responses only against confirmed published state.
- On a new apply, discard only this resource's unfinished revision; treat a confirmed missing owned revision as stale. Never discard an unrelated draft.
- For relative placement, require ref. Rules default to LAST_IN_POLICY on create; omitted rule placement preserves current position on update. Section placement is required.
- Refuse deletion if the published policy or newly created revision snapshot contains a rule in this section. Never cascade into unmanaged rules.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Verification

- Required local imports, unit tests, build, lint, vulnerability scan and docs checks passed.
- Bugbot found a move-recovery edge case; fixed it and obtained a clean re-review.
- Live create/read/import/update/relative-move/no-drift/delete checks passed using isolated empty section fixtures.
- Lost publication responses must also confirm the requested published section order before updating placement state.
