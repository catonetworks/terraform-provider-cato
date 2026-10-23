# API contract

Data source: `cato_roles`.
Query: `rbac.roleManagement.roleList`. Source: server-services commit
`77770519fd3c531ef830eb86b0fb16c9077bc8d8`, IAM public RBAC schema and
RoleManagementQueryService. SDK commit: `5ff6a5aac4132a1e3aa4c0a420e7d5c2c6c81918`.

Implemented from scratch. Role lookups support custom and predefined roles.
Role list queries follow offset pagination and preserve API name sorting.
Use shared IAM conversions and SDK interfaces; do not import resource packages.
