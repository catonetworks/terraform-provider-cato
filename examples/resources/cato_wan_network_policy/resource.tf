// WAN Network account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// WAN Network policy. The individual WAN Network rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_wan_network_policy" "this" {
  enabled = true // set false to disable the whole WAN Network policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "wan_network_policy_id" {
  value = cato_wan_network_policy.this.id
}
