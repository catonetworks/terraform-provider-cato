# cato_socket_bypass_section

Status: implemented and verified on 2026-10-29; schema approved by the user.

- Package: `internal/provider/policy/socketbypass/sbsection/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `PolicyAddSectionInput`.
- Import target: `id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.policy.socketBypass.addSection` → SDK `PolicySocketBypassAddSection` (`sources/mutation.policy.socketBypass.addSection.gql`).
- `mutation.policy.socketBypass.updateSection` → SDK `PolicySocketBypassUpdateSection` (`sources/mutation.policy.socketBypass.updateSection.gql`).
- `mutation.policy.socketBypass.removeSection` → SDK `PolicySocketBypassRemoveSection` (`sources/mutation.policy.socketBypass.removeSection.gql`).
- `query.policy.socketBypass.policy` → SDK `PolicySocketBypassPolicy` (`sources/query.policy.socketBypass.policy.gql`).
- `mutation.policy.socketBypass.moveSection` → SDK `PolicySocketBypassMoveSection` (`sources/mutation.policy.socketBypass.moveSection.gql`).
- `mutation.policy.socketBypass.createPolicyRevision` → SDK `PolicySocketBypassCreatePolicyRevision` (`sources/mutation.policy.socketBypass.createPolicyRevision.gql`).
- `mutation.policy.socketBypass.publishPolicyRevision` → SDK `PolicySocketBypassPublishPolicyRevision` (`sources/mutation.policy.socketBypass.publishPolicyRevision.gql`).
- `mutation.policy.socketBypass.discardPolicyRevision` → SDK `PolicySocketBypassDiscardPolicyRevision` (`sources/mutation.policy.socketBypass.discardPolicyRevision.gql`).

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

- make sort-imports, make test, go build ./..., make lint, make vul, and make docs passed. Zero reachable vulnerabilities; two module-only advisories.
- Targeted live acceptance passed create/read/import/update/relative movement/no-drift/destroy; owned fixtures removed.
- Bugbot review found no actionable bugs.
