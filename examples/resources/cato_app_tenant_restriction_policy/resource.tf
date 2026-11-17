// App Tenant Restriction account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// App Tenant Restriction policy. The individual App Tenant Restriction rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_app_tenant_restriction_policy" "this" {
  enabled = true // set false to disable the whole App Tenant Restriction policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "app_tenant_restriction_policy_id" {
  value = cato_app_tenant_restriction_policy.this.id
}
