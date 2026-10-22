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
| Delete | policy.ztnaAlwaysOn.removeRule | PolicyZtnaAlwaysOnRemoveRule |

All client methods accept accountID; mutation methods also accept optional
ZtnaAlwaysOnPolicyMutationInput for revision selection. Read accepts optional
ZtnaAlwaysOnPolicyInput. Create uses ZtnaAlwaysOnAddRuleInput; update uses
ZtnaAlwaysOnUpdateRuleInput; delete uses ZtnaAlwaysOnRemoveRuleInput.
The rule ID is nested in the create payload at rule.rule.id; read locates the
matching ID in policy.rules[].rule.

## Proposed schema

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

## Implementation considerations

- anti_tamper_mode is required in Terraform despite being optional in the API
  with default OFF. Empty enum values are rejected.
- `at` is a create positioning input. Updates need explicit positioning semantics;
  the updateRule mutation does not accept `at` (moveRule exists separately).
- SDK update collection fields have omitempty tags. Verify explicit empty-array
  forwarding when clearing matching criteria; omission would leave old values.
- Preserve existing untracked files and root schema_description.yaml.
- The user approved the schema and requested schema implementation, then directed
  alignment with the updated skill and coding rules. Schema implementation exists;
  model, CRUD, registration, and their tests remain to be implemented.
