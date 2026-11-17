// Split Tunnel account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// Split Tunnel policy. The individual Split Tunnel rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_split_tunnel_policy" "this" {
  enabled = true // set false to disable the whole Split Tunnel policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "split_tunnel_policy_id" {
  value = cato_split_tunnel_policy.this.id
}
