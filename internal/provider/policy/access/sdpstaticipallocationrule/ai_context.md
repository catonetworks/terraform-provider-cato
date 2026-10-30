# cato_sdp_static_ip_allocation_rule

Status: implemented and verified on 2026-10-29; local checks, Bugbot review and live acceptance passed.

- Package: `internal/provider/policy/access/sdpstaticipallocationrule/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `SdpStaticIpAllocationAddRuleInput`.
- Import target: `rule.id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.policy.sdpStaticIpAllocation.addRule` → SDK `PolicySdpStaticIPAllocationAddRule` (`sources/mutation.policy.sdpStaticIpAllocation.addRule.gql`).
- `mutation.policy.sdpStaticIpAllocation.updateRule` → SDK `PolicySdpStaticIPAllocationUpdateRule` (`sources/mutation.policy.sdpStaticIpAllocation.updateRule.gql`).
- `mutation.policy.sdpStaticIpAllocation.removeRule` → SDK `PolicySdpStaticIPAllocationRemoveRule` (`sources/mutation.policy.sdpStaticIpAllocation.removeRule.gql`).
- `query.policy.sdpStaticIpAllocation.policy` → SDK `PolicySdpStaticIPAllocationPolicy` (`sources/query.policy.sdpStaticIpAllocation.policy.gql`).
- `mutation.policy.sdpStaticIpAllocation.moveRule` → SDK `PolicySdpStaticIPAllocationMoveRule` (`sources/mutation.policy.sdpStaticIpAllocation.moveRule.gql`).
- `mutation.policy.sdpStaticIpAllocation.createPolicyRevision` → SDK `PolicySdpStaticIPAllocationCreatePolicyRevision` (`sources/mutation.policy.sdpStaticIpAllocation.createPolicyRevision.gql`).
- `mutation.policy.sdpStaticIpAllocation.publishPolicyRevision` → SDK `PolicySdpStaticIPAllocationPublishPolicyRevision` (`sources/mutation.policy.sdpStaticIpAllocation.publishPolicyRevision.gql`).
- `mutation.policy.sdpStaticIpAllocation.discardPolicyRevision` → SDK `PolicySdpStaticIPAllocationDiscardPolicyRevision` (`sources/mutation.policy.sdpStaticIpAllocation.discardPolicyRevision.gql`).

## Ownership, conversion and live-test requirements

- Use a unique named revision for each apply; all mutations/discards must use its exact explicit ID. Publish automatically after successful mutation/movement.
- Read PUBLIC policy explicitly and keep refresh observational. Retain resource identity on failed create publication and prior published state on failed updates. Reconcile lost publication responses only against confirmed published state.
- On a new apply, discard only this resource's unfinished revision; treat a confirmed missing owned revision as stale. Never discard an unrelated draft.
- For relative placement, require ref. Rules default to LAST_IN_POLICY on create; omitted rule placement preserves current position on update. Section placement is required.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

Backend validation requires source.user even for disabled rules; it is required in the schema and both conversion paths.

## Verification

- Required local checks passed: sort-imports, test, build, lint, vulnerability scan, documentation generation. Zero reachable vulnerabilities; two module-only advisories.
- Bugbot found required source.user validation; fixed schema/converters and added null/unknown regression coverage. Re-review clean.
- After configuring an isolated static client range, targeted live acceptance passed create/read/import/update/movement/no-drift/destroy. Dedicated disabled local user and rule removed.
- The temporary account pool entry was restored to its prior empty state via CMA, and the dedicated global-range fixture was deleted and verified absent via API.
- Acceptance requires TFACC_SDP_STATIC_IP from an already configured test-account static client range.
