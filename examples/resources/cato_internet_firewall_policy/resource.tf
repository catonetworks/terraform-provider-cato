// Internet Firewall account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// Internet Firewall policy. The individual Internet Firewall rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_internet_firewall_policy" "this" {
  enabled = true // set false to disable the whole Internet Firewall policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "internet_firewall_policy_id" {
  value = cato_internet_firewall_policy.this.id
}
