// Site Web Proxy account-level policy toggle (SINGLETON).
//
// This resource owns the account-wide enable/disable state for the entire
// Site Web Proxy policy. The individual Site Web Proxy rules/sections only take effect while
// this policy is enabled. There is exactly ONE of these per account.
resource "cato_site_web_proxy_policy" "this" {
  enabled = true // set false to disable the whole Site Web Proxy policy
}

// All other attributes are read-only, populated after apply:
//   id, published_by, published_time, revision_id, revision_name
output "site_web_proxy_policy_id" {
  value = cato_site_web_proxy_policy.this.id
}
