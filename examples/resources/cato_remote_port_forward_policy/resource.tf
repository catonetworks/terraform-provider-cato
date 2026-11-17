// Remote Port Forwarding account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// Remote Port Forwarding policy. The individual Remote Port Forwarding rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_remote_port_forward_policy" "this" {
  enabled = true // set false to disable the whole Remote Port Forwarding policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "remote_port_forward_policy_id" {
  value = cato_remote_port_forward_policy.this.id
}
