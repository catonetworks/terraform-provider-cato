// SDP Static IP Allocation account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// SDP Static IP Allocation policy. The individual SDP Static IP Allocation rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_sdp_static_ip_allocation_policy" "this" {
  enabled = true // set false to disable the whole SDP Static IP Allocation policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "sdp_static_ip_allocation_policy_id" {
  value = cato_sdp_static_ip_allocation_policy.this.id
}
