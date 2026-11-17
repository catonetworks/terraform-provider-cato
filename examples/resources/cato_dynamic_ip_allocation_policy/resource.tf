// Dynamic IP Allocation account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// Dynamic IP Allocation policy. The individual Dynamic IP Allocation rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_dynamic_ip_allocation_policy" "this" {
  enabled = true // set false to disable the whole Dynamic IP Allocation policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "dynamic_ip_allocation_policy_id" {
  value = cato_dynamic_ip_allocation_policy.this.id
}
