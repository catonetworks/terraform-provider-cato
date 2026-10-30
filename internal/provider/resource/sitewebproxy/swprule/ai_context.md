# cato_site_web_proxy_rule

Status: implementation and local checks complete; live testing blocked. Draft PR.

- Package: `internal/provider/resource/sitewebproxy/swprule/`.
- Base: `beta`; released Cato Go SDK v1.1.3 (`506577112e8425ee31d8aae08d22d8b3d20ff675`).
- Source: lifecycle section of Confluence page 7832371287, version 3.
- SDL: fetched 2026-10-29 from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- SDL SHA-256: `528547eb33c6c57b7c8978bc4a528ebde997dc1359cb2afc6efe250b380546de`.
- Creation input: `SiteWebProxyAddRuleInput`.
- Import target: `rule.id`; exact remote ID, scoped to the configured account.

## API operations

- `mutation.policy.siteWebProxy.addRule` → SDK `PolicySiteWebProxyAddRule` (`sources/mutation.policy.siteWebProxy.addRule.gql`).
- `mutation.policy.siteWebProxy.updateRule` → SDK `PolicySiteWebProxyUpdateRule` (`sources/mutation.policy.siteWebProxy.updateRule.gql`).
- `mutation.policy.siteWebProxy.removeRule` → SDK `PolicySiteWebProxyRemoveRule` (`sources/mutation.policy.siteWebProxy.removeRule.gql`).
- `query.policy.siteWebProxy.policy` → SDK `PolicySiteWebProxyPolicy` (`sources/query.policy.siteWebProxy.policy.gql`).
- `mutation.policy.siteWebProxy.moveRule` → SDK `PolicySiteWebProxyMoveRule` (`sources/mutation.policy.siteWebProxy.moveRule.gql`).
- `mutation.policy.siteWebProxy.createPolicyRevision` → SDK `PolicySiteWebProxyCreatePolicyRevision` (`sources/mutation.policy.siteWebProxy.createPolicyRevision.gql`).
- `mutation.policy.siteWebProxy.publishPolicyRevision` → SDK `PolicySiteWebProxyPublishPolicyRevision` (`sources/mutation.policy.siteWebProxy.publishPolicyRevision.gql`).
- `mutation.policy.siteWebProxy.discardPolicyRevision` → SDK `PolicySiteWebProxyDiscardPolicyRevision` (`sources/mutation.policy.siteWebProxy.discardPolicyRevision.gql`).

## Ownership, conversion and live-test requirements

- Use a unique named revision for each apply; all mutations/discards must use its exact explicit ID. Publish automatically after successful mutation/movement.
- Read PUBLIC policy explicitly and keep refresh observational. Retain resource identity on failed create publication and prior published state on failed updates. Reconcile lost publication responses only against confirmed published state.
- On a new apply, discard only this resource's unfinished revision; treat a confirmed missing owned revision as stale. Never discard an unrelated draft.
- For relative placement, require ref. Rules default to LAST_IN_POLICY on create; omitted rule placement preserves current position on update. Section placement is required.
- Own outer proxy rules only; nested webProxyTraffic operations require separate SDK coverage and remain outside this resource.
- Accept the API encrypted_keytab field as sensitive input/readback. Raw keytab file upload is a separate operational API and is not part of resource CRUD.
- Live fixtures remain disabled and use only test-owned or explicitly designated test sites.
- Write unit tests first; generate mocks through pinned `make mocks`. Cover null/unknown values, payload shapes, CRUD failures, missing objects, import and state hydration.
- Review with Bugbot, fix findings, and run actual CRUD/import/movement where supported plus no-drift plans in the authorized development accounts. Remove only fixtures created for this test.
- Run `make sort-imports`, `make test`, `go build ./...`, `make lint`, `make vul`, and `make docs`. Keep account credentials and live data out of source control.
- Create one independent resource branch/PR from `beta`; leave PRs unmerged. Shared-helper extraction remains deferred.

## Verification boundary

- Live fallback account permits reads and revisions but rejects addRule with Unsupported field [addRule]. Further live mutations skipped per user instruction. No rule was created; temporary parents were removed and exact absence verified. Draft until backend support enables full live validation.
- Omitted Kerberos blocks actively reset false/empty update fields. Only the disabled/empty API default is canonicalized to null. Explicit empty disabled configuration must be omitted; meaningful disabled keytabs remain supported and external nondefault values remain visible.
- SDK projections keep operation variable contracts while excluding nested traffic and audit data. Diagnostics do not print encrypted keytabs or request bodies.
