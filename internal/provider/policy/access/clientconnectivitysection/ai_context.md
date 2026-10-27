# cato_client_connectivity_section

Status: schema approved by the user; implementation uses resource-local helpers until the later shared-helper review.

- Package: `internal/provider/policy/access/clientconnectivitysection`.
- SDK: `github.com/catonetworks/cato-go-sdk v1.1.3`, commit `506577112e8425ee31d8aae08d22d8b3d20ff675`.
- GraphQL schema: `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.
- Schema SHA-256: `93d6289ff914e4b5b8a41c10156b5521d5e23e90ce73e0df9aa8f1fb4e8e70d9`.
- Read: `query.policy.clientConnectivity.policy`, SDK `PolicyClientConnectivityPolicy`.
- Mutation: `mutation.policy.clientConnectivity.addSection`, SDK `PolicyClientConnectivityAddSection`.
- Mutation: `mutation.policy.clientConnectivity.updateSection`, SDK `PolicyClientConnectivityUpdateSection`.
- Mutation: `mutation.policy.clientConnectivity.removeSection`, SDK `PolicyClientConnectivityRemoveSection`.
- Mutation: `mutation.policy.clientConnectivity.moveSection`, SDK `PolicyClientConnectivityMoveSection`.
- Mutation: `mutation.policy.clientConnectivity.createPolicyRevision`, SDK `PolicyClientConnectivityCreatePolicyRevision`.
- Mutation: `mutation.policy.clientConnectivity.publishPolicyRevision`, SDK `PolicyClientConnectivityPublishPolicyRevision`.
- Mutation: `mutation.policy.clientConnectivity.discardPolicyRevision`, SDK `PolicyClientConnectivityDiscardPolicyRevision`.

## Schema conventions

- Follow the create-input tree; retain required fields even where the API has a default.
- Use sets for unordered match criteria; OR within criteria and AND between criteria.
- Object references expose optional/computed `id` and `name`; require at least one configured reference selector and reuse shared id/name helpers.
- Enum validators use SDK enum constants.
- Optional `at` on rules defaults to last in policy on create; omission during update preserves placement. Section `at` is required. Relative placements require `ref`.
- Computed identity: `id` and `section.id`, with import using `id`.
- Automatically publish create, update, move and delete; scope all requests to the provider account and preserve the operation's revision identity.
- Resource owns its section only, without managing account policy settings. Section deletion must not cascade into unmanaged rules.
- On create publication failure, persist the returned resource ID before returning diagnostics so the draft can be recovered.
- Diagnose SDK/GraphQL payload errors and absent expected IDs. Do not discard another session's revisions.
- Add unit/request-shape and acceptance-test definitions; do not execute live acceptance tests or mutations.
