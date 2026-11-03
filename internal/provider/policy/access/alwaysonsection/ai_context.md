# cato_always_on_section

Package: `internal/provider/policy/access/alwaysonsection`.

Status: schema approved by the user on 2026-10-27; resource implemented and
registered in provider.go.

## API and SDK

Live schema inspected on 2026-10-27:
https://system.cc.catonetworks.com/api/schema?with_undocumented=true

Pinned dependency: `github.com/catonetworks/cato-go-sdk v1.1.2` (unchanged).

| Operation | GraphQL path | SDK client method | SDK input |
| --- | --- | --- | --- |
| Create | policy.ztnaAlwaysOn.addSection | PolicyZtnaAlwaysOnAddSection | PolicyAddSectionInput |
| Read | policy.ztnaAlwaysOn.policy.sections | PolicyZtnaAlwaysOnPolicy | ZtnaAlwaysOnPolicyInput |
| Update | policy.ztnaAlwaysOn.updateSection | PolicyZtnaAlwaysOnUpdateSection | PolicyUpdateSectionInput |
| Move | policy.ztnaAlwaysOn.moveSection | PolicyZtnaAlwaysOnMoveSection | PolicyMoveSectionInput |
| Delete | policy.ztnaAlwaysOn.removeSection | PolicyZtnaAlwaysOnRemoveSection | PolicyRemoveSectionInput |

All SDK methods accept the configured account ID and request interceptors.
Mutations also accept an optional ZtnaAlwaysOnPolicyMutationInput for revision
selection. Read accepts an optional ZtnaAlwaysOnPolicyInput. Follow the existing
alwaysonrule resource: use nil revision inputs for the default draft policy and
do not publish policy changes.

Create input recursively:

- PolicyAddSectionInput.at: PolicySectionPositionInput!.
  - position: PolicySectionPositionEnum!.
  - ref: ID.
- PolicyAddSectionInput.section: PolicyAddSectionInfoInput!.
  - name: String!.

The SDK create input matches the live GraphQL input. Generated Go pointers for
required nested objects do not make those objects optional in Terraform.
PolicySectionPositionEnum contains AFTER_SECTION, BEFORE_SECTION, LAST_IN_POLICY.
Validate against cato_models.AllPolicySectionPositionEnum.

All section mutations return PolicySectionMutationPayload: errors, revision,
section, and status. The nested section is PolicySectionPayload, containing
audit, access, metadata, properties, and section: PolicySectionInfo!. The latter
contains id: ID!, name: String!, subPolicyId: ID. The matching configuration
subtree is PolicySectionInfo; expose its id as computed and name as required.
The pinned SDK selects only id and name for this subtree in create, read,
update, move, and delete operations. Exclude sub_policy_id until the SDK selects
it; this is a response-projection discrepancy with the live schema. Audit,
access, metadata, properties, revision, status, and errors are operation or
policy metadata outside the resource configuration subtree.

Create response ID path:
policy.ztnaAlwaysOn.addSection.section.section.id.
Read locates the section by ID in policy.ztnaAlwaysOn.policy.sections[].section.
Read the section list directly, including empty sections; do not infer sections
from policy rules.

Update input requires id and section, with an optional section.name. Terraform
keeps name required to match the create input. Delete requires only id.
updateSection cannot change placement; the proposal includes moveSection when
at changes, using PolicyMoveSectionInput with the same section ID and to = at.

## Terraform schema

See schema_description.yaml, sorted alphabetically at every level. Preserve the
create input structure: required at and section. section.id is the sole resource
identifier; do not add a duplicate top-level id. Import passes the ID through
to section.id, and Read hydrates section.name by scanning the section list.

at.position is required and uses the SDK enum validator. at.ref is
optional_computed with the modifyPlan hint, per the skill's optional-string
mapping; normalize an omitted value to an empty string and omit an empty ref
from API requests. Require a nonempty ref for BEFORE_SECTION and AFTER_SECTION.
Preserve at from plan/state because the read response does not return the
original relative-placement input. For import, derive a valid position from
the ordered section list rather than claiming an arbitrary placement.

section.id is computed with UseNonNullStateForUnknown. section.name is required.
PolicySectionInfo is not an ObjectRefBy/String input pair, so do not apply the
idName reference special case to the section configuration.

## Implementation and verification

Lifecycle orchestration, diagnostics, and state live in the resource file;
conversion builders/parsers in converters; the narrow injected ResourceClient,
SDK forwarding adapter, and section lookup helpers in the client file. The
resource reuses policyrule.Position attribute types with section-specific enum
validation. All calls preserve configured account scoping and request interceptors.
Mutations validate status, payload errors, and returned IDs before updating state.
A valid section list without the resource ID removes it from state; malformed
responses produce diagnostics. Empty sections are read from sections directly.

Update renames first, then moves only when at changes. If the move fails, state
retains the completed rename and previous placement so Terraform can retry.
Live testing on 2026-10-27 confirmed updateSection can report SUCCESS with the
previous section name in its payload even though the rename persisted. When the
mutation name differs from the request, Update reads the section and validates
the requested name before saving state. A failed read or an unchanged remote
name produces diagnostics rather than hiding the discrepancy.

Unit tests cover CRUD, import, reference validation, null/unknown values, empty
sections, missing and duplicate IDs, malformed responses, API errors, account,
revision and interceptor forwarding, placement changes, partial update/move
failure state handling, and stale rename responses. AlwaysOnSectionClient was
generated with pinned make mocks; do not edit it manually.

Acceptance coverage is in internal/acctests/always_on_section and follows the
explicitly invoked terraform-acceptance-tests skill for account 62015. It covers
create, rename, import using section.id, before/after moves against a section
owned by the same test, and automatic deletion. Import verification selects
section.id as its identifier and ignores at because the original insertion
input is not returned by the API. Move checks also query the API section list
to verify actual ordering. Run only when explicitly authorized, with
DISABLE_POLICY_RULE_CLEANUP=true to preserve unrelated draft policy changes.

The example and generated resource documentation include all configurable
attributes and an import command. Required repository checks: make sort-imports,
make test, go build ./..., make lint, make vul, and make docs.
