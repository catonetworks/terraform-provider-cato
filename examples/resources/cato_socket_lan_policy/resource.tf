// Socket LAN account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// Socket LAN policy (both Socket LAN firewall and network rules). The individual
// Socket LAN rules/sections only take effect while this policy is enabled.
// There is exactly ONE of these per account.
resource "cato_socket_lan_policy" "this" {
  enabled = true // set false to disable the whole Socket LAN policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "socket_lan_policy_id" {
  value = cato_socket_lan_policy.this.id
}
