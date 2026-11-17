// Terminal Server account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// Terminal Server policy. The individual Terminal Server rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_terminal_server_policy" "this" {
  enabled = true // set false to disable the whole Terminal Server policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "terminal_server_policy_id" {
  value = cato_terminal_server_policy.this.id
}
