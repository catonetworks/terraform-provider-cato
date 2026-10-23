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

All SDK calls use generated operations. The provider pins SDK commit
`5ff6a5aac4132a1e3aa4c0a420e7d5c2c6c81918`; replace the pseudo-version with a
released tag before merge. Role list `paging.from` is an offset; assignment
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
