# cato_bulk_always_on_rule

Package: `internal/provider/policy/access/alwaysonindex`.
Status: approved schema implemented with unit and request-shape tests.

## Scope

Bulk ordering of existing AlwaysOn policy sections and rules, following
`internal/provider/policy/lan/lanindex/schema_bulk_lf_move_rule.go`.
The hierarchy contains only policy -> section -> rule. There are no sub-policies,
network-rule layers, or firewall sub-rules.

Preserve unrelated AlwaysOn policy and section work when extending shared
registration or mock configuration.

## API and SDK analysis

Live schema fetched and inspected on 2026-10-27:
https://system.cc.catonetworks.com/api/schema?with_undocumented=true

SDK dependency inspected: `github.com/catonetworks/cato-go-sdk v1.1.2`.

| Lifecycle | GraphQL path | SDK method |
| --- | --- | --- |
| Create/Update section order | policy.ztnaAlwaysOn.moveSection | PolicyZtnaAlwaysOnMoveSection |
| Create/Update rule order | policy.ztnaAlwaysOn.moveRule | PolicyZtnaAlwaysOnMoveRule |
| Read | policy.ztnaAlwaysOn.policy | PolicyZtnaAlwaysOnPolicy |
| Publish ordering changes | policy.ztnaAlwaysOn.publishPolicyRevision | PolicyZtnaAlwaysOnPublishPolicyRevision |
| Delete | No mutation; remove Terraform state only | None |

Both move methods accept accountID, optional ZtnaAlwaysOnPolicyMutationInput,
and a required move input. PolicyMoveSectionInput and PolicyMoveRuleInput each
contain required id and optional to. Section positions are AFTER_SECTION,
BEFORE_SECTION, LAST_IN_POLICY. Rule positions are AFTER_RULE, BEFORE_RULE,
FIRST_IN_POLICY, FIRST_IN_SECTION, LAST_IN_POLICY, LAST_IN_SECTION.
The position inputs also accept an optional reference ID.
Generated SDK input types match the inspected live schema for these operations.

The read query accepts accountID and optional ZtnaAlwaysOnPolicyInput. Its
response contains ordered sections[].section and rules[].rule. Rules expose
their ID, name, index, and parent section ID/name. Section order comes from
the section array. Move responses contain mutation errors/status and the
updated section or rule. Validate transport errors, missing payloads, payload
errors, and unsuccessful statuses before proceeding to publish.

Publish accepts accountID, optional ZtnaAlwaysOnPolicyMutationInput, and optional
PolicyPublishRevisionInput (name/description). The response includes policy,
errors, and status. Default draft revision inputs are nil, consistent with
existing AlwaysOn resources. Publishing matches the LAN bulk ordering lifecycle;
individual AlwaysOn rule and section resources leave changes in the draft.

## Terraform contract

See schema_description.yaml. section_data is a required map; rule_data is an
optional map. Each section_data map key is the API section name; its value
contains only computed id and required section_index. There is no section_name
attribute within section_data. Each rule_data map key is the API rule name;
there is no rule_name attribute within rule_data. Rules reference
their parent using required section_name, matching a section_data map key;
there is no section_key attribute. Ambiguous API section names produce a
diagnostic. Explicit section and per-section rule indexes start at 1. IDs are
computed.

Create and Update resolve and validate all requested items before moving them,
apply section order and then rule order, publish successful changes, and refresh
state. Both lifecycle methods explicitly call moveSections(), moveRules(), and
publish(). Read refreshes actual positions and preserves omitted versus empty
rule_data. It follows stored IDs across remote renames and parent changes,
refreshing map keys to the API names. Missing items are omitted from refreshed
state. Repeated applies with matching order and no pending draft changes do not
move or publish. A default private revision reporting pending changes is
published even when ordering already matches, allowing a failed publish to be
retried. Publishing includes all pending changes in that revision.
Unmanaged policy items retain their relative order. Missing or ambiguous names,
duplicate target positions, invalid indexes, and inconsistent parent references
must produce diagnostics before mutations. Partial maps are supported: configured
positions are absolute positions within all current policy sections or all rules
in a section. Unconfigured items fill the remaining positions in their original
relative order. Rules must already belong to the configured parent section.

Delete removes Terraform state and leaves the existing policy untouched. Import is unsupported,
matching the LAN bulk ordering resource; there is no unique resource ID or single
object fetch defining the requested mapping.

## Implementation and verification

Resource/model/schema/converters/client/validators files live within this package.
SDK calls use a narrow injectable ResourceClient adapter. BulkAlwaysOnRuleClient
is generated with pinned mockery through make mocks. The resource is registered
in provider.go; the example lives under examples/resources/cato_bulk_always_on_rule.
Unit tests cover lifecycle call order, payloads, Delete without API calls, omitted
and empty rule_data, drift, invalid names/positions, API failures, convergence,
and publishing retries. HTTP tests verify account/revision/interceptor forwarding.
Acceptance coverage uses disabled rule fixtures, tests Create/Update/removing the
ordering resource, and publishes fixture deletions during cleanup. Do not run
acceptance tests without an explicit request. The SDK dependency is unchanged.
