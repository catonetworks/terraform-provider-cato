// Client Connectivity account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// Client Connectivity policy. The individual Client Connectivity rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_client_connectivity_policy" "this" {
  enabled = true // set false to disable the whole Client Connectivity policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "client_connectivity_policy_id" {
  value = cato_client_connectivity_policy.this.id
}
