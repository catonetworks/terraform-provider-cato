# API provenance

Implemented from scratch using server-services commit
`77770519fd3c531ef830eb86b0fb16c9077bc8d8`.
Source: `services/iam/src/main/resources/schema/rbac.public.graphqls` and
IAM role-management query/mutation services and RBAC assignment converters/services.

## Approved Terraform mapping

Unordered permissions, role IDs, entity grants, applicable account types, and
catalog entries use sets. Role list results retain API order. Each assignment
resource owns one user or users-group principal. Account changes replace either
resource; principal changes replace assignments. Imports use composite account IDs.

Resource: `cato_role`. Operations: createRole, role/readRole, updateRole, deleteRole under rbac.roleManagement.

All SDK calls use generated operations from released SDK `v1.1.2`
(commit `dc2a6a05a3b895c3de2ee65c0f70168714dda77c`). Role list `paging.from` is an offset; assignment
list `paging.from` is a page number. Incomplete responses produce diagnostics,
not missing-object state removal. No policy publication is required.

IAM role writes delegate to CC2, while reads query the replica. Create and Update
retain the acknowledged mutation response in state, then poll the read API for
up to 30 seconds at 250ms intervals for mutable fields to converge. Mutation
requests are never retried by this consistency loop. Poll failures report a
warning while preserving the acknowledged ID/state, avoiding Terraform taint after a successful write. Ordinary refresh still
removes a genuinely absent role.

Role deletion retries for up to 30 seconds only for HTTP-200 GraphQL
`DOWNSTREAM_SERVICE_ERROR` errors on `rbac.roleManagement.deleteRole` that
report temporary role-management unavailability or `errorType: UNAVAILABLE`.
Before retrying, it reads the role: confirmed absence completes deletion;
lookup failures, validation/authorization errors, and deadline expiration retain
diagnostics. Failed delete responses alone never remove the role from state.
If the retry request reaches the deadline, its diagnostic also retains the last
backend unavailability error and trace ID. A timeout is not evidence of deletion.
