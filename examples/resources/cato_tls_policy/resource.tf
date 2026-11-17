// TLS Inspection account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// TLS Inspection policy. The individual TLS Inspection rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_tls_policy" "this" {
  enabled = true // set false to disable the whole TLS Inspection policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "tls_policy_id" {
  value = cato_tls_policy.this.id
}
