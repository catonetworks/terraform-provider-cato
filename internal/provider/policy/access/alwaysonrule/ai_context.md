# cato_always_on_rule

Package: `internal/provider/policy/access/alwaysonrule`.

## API and SDK

Schema inspected: https://system.cc.catonetworks.com/api/schema?with_undocumented=true

Dependency: `github.com/catonetworks/cato-go-sdk v1.1.0` (unchanged).

| Operation | GraphQL path | SDK client method |
| --- | --- | --- |
| Create | policy.ztnaAlwaysOn.addRule | PolicyZtnaAlwaysOnAddRule |
| Read | policy.ztnaAlwaysOn.policy | PolicyZtnaAlwaysOnPolicy |
| Update | policy.ztnaAlwaysOn.updateRule | PolicyZtnaAlwaysOnUpdateRule |
| Move | policy.ztnaAlwaysOn.moveRule | PolicyZtnaAlwaysOnMoveRule |
| Delete | policy.ztnaAlwaysOn.removeRule | PolicyZtnaAlwaysOnRemoveRule |

All client methods accept accountID; mutation methods also accept optional
ZtnaAlwaysOnPolicyMutationInput for revision selection. Read accepts optional
ZtnaAlwaysOnPolicyInput. Create uses ZtnaAlwaysOnAddRuleInput; update uses
ZtnaAlwaysOnUpdateRuleInput; delete uses ZtnaAlwaysOnRemoveRuleInput.
The rule ID is nested in the create payload at rule.rule.id; read locates the
matching ID in policy.rules[].rule.

## Schema

See schema_description.yaml. Preserve the create input tree: optional `at` and
required `rule`. Compute rule.id, rule.index, and rule.section from responses.
Required modes follow GraphQL non-null markers even when GraphQL defines defaults.
All enum attributes are required per tf-resource section 1.2.5, including
at.position (when at is configured) and rule.anti_tamper_mode. Enum validation
uses stringvalidator.OneOf(utils.ToStrings(cato_models.AllEnumName)...), per
coding rule TFSCHEMA-010.
Use unordered sets for platforms, device profiles, users, and user groups.
Use id/name references instead of by/input objects, following the skill.
Enum hints identify actual SDK model types.

The live schema returns section.subPolicyId, but the SDK generated add, update,
and policy query selections expose only section.id and section.name. The proposal
excludes sub_policy_id until SDK selections support it. Create input types match
the live API fields and nullability; no input discrepancy found.

## Implementation

CRUD is implemented in resource_always_on_rule.go, using a narrow ResourceClient
interface, an injected apiClient, and an SDK forwarding adapter in the client file.
Builders/parsers live in converters_always_on_rule.go. The existing model matches
the approved schema. The schema description marks
at.ref optional_computed with modifyPlan, matching shared SchemaPosition.
The live schema was rechecked on 2026-10-23; CRUD inputs match SDK v1.1.0.

Create calls addRule once in the default draft revision (nil policy mutation
input), checks transport errors, payload errors, mutation status and the rule ID,
then hydrates state from the returned rule. It preserves at from the plan and
does not publish. Empty matching criteria are sent as JSON arrays. References
use idname.PrepareIDNameSet per CONV-007. The shared helper resolves by known
non-empty ID first, then name; empty references produce diagnostics. The helper
always returns a non-nil slice, including for empty/null/unknown sets and errors.
Nullable antiTamperMode output uses the API default OFF.

Read queries the default draft policy using PolicyZtnaAlwaysOnPolicy with the
configured account ID and nil policy input. It locates the rule by rule.id and
refreshes every rule attribute while preserving at. Import passes the ID to
rule.id; Read hydrates all remaining fields after import. A valid policy with no
matching rule removes the resource from state. Transport errors, missing policy
or rule list, and entries without an ID produce diagnostics and preserve state.
All rule response parsers share scalar, duration, source, and reference conversion logic,
including empty-set normalization and the OFF anti-tamper default.

Update builds a full updateRule input from the plan using the state ID. It reuses
the Create builder for required fields and reference criteria, preserving false
booleans, empty strings, and non-nil empty slices. The pinned SDK's clientv2
serializer includes non-nil empty slices despite omitempty tags; HTTP request
tests verify all four criteria arrays are sent as [] when clearing them.
When a non-null at differs from state, Update calls moveRule after updateRule.
Omitting at preserves the current position. The move response supplies the new
index and section. If moving fails after the update succeeds, state records the
updated rule with the previous at, allowing the move to be retried.

Delete calls removeRule with the state ID and validates transport errors, payload
errors, and mutation status. A FAILURE containing only ruleNotExist or RuleNotFound
errors is treated as already removed. Other errors preserve state. Successful
deletion removes the resource from state.

All mutations use nil policy input for the default draft revision and do not
publish. Update and Move require the response rule ID to match the state ID.
The resource is registered in provider.go.

Unit tests cover CRUD, API errors, missing rules, malformed responses, state
refresh, import, defaults, positioning, partial update/move failure, and reference
inputs. HTTP tests verify account/revision/interceptor forwarding and explicit
empty-array updates. The acceptance test exercises CRUD, import, moving, and
refresh after deletion with a disabled draft rule. SDK cleanup is a fallback if
the lifecycle test fails before Delete. Do not run it unless explicitly requested.

AlwaysOnRuleClient is generated from ResourceClient by pinned mockery v3.6.0
using make mocks.

## Remaining considerations

- anti_tamper_mode is required in Terraform despite being optional in the API
  with default OFF. Empty enum values are rejected.
- `at` controls initial placement and explicit moves on update; it is not derived
  from the read response. API index and section are refreshed separately.
