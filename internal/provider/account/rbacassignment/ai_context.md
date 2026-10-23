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

Resource: `cato_rbac_assignment`. Operations: addRbacAssignments, rbacAssignmentList, updateRbacAssignment, removeRbacAssignment under rbac.

All SDK calls use generated operations from released SDK `v1.1.2`
(commit `dc2a6a05a3b895c3de2ee65c0f70168714dda77c`). Role list `paging.from` is an offset; assignment
list `paging.from` is a page number. Incomplete responses produce diagnostics,
not missing-object state removal. No policy publication is required.

The public schema intentionally manages one ID-based principal per resource.
Unconfigured principal references stay null to preserve mutual exclusion.
The optional obfuscation preference also stays null when omitted: defaulting it
to false would overwrite a stored permission. Effective all-sites/all-users grants
are computed from SYSTEMGROUP scope. Send NONE in their API shortcut inputs to
avoid maintaining a second representation. Preserve hidden preferences when the
API reports DISABLED. Reject legacy ADMIN mutations as the backend does.

Assignment reads send `filter: {}` even without criteria because the IAM service
dereferences the filter. Omitting it produces an HTTP-200 GraphQL error.

The assignment acceptance suite creates a unique user through the generated
user API and removes it with `t.Cleanup` after Terraform destroy. It uses an
`example.invalid` address and requires no pre-existing user or CI user-ID setting.

The public `RBACAssignmentFilterInput` has no assignment-ID field. Its public
filters are search, creationDate, lastModifiedDate, and type; admin is private.
Lookup therefore scans account-scoped pages with an explicit empty filter. If
the total changes between pages, restart from page zero up to three attempts.
Persistent churn produces an error and retains state rather than reporting an
assignment absent after rows may have shifted across page boundaries. Duplicate
IDs, malformed metadata, cross-account results, and API failures still fail closed.
