# API contract

Data source: `cato_roles`.
Query: `rbac.roleManagement.roleList`. Source: server-services commit
`77770519fd3c531ef830eb86b0fb16c9077bc8d8`, IAM public RBAC schema and
RoleManagementQueryService. SDK: released `v1.1.2`, commit `dc2a6a05a3b895c3de2ee65c0f70168714dda77c`.

Implemented from scratch. Role lookups support custom and predefined roles.
Role list queries follow offset pagination and preserve API name sorting.
Use shared IAM conversions and SDK interfaces; do not import resource packages.
