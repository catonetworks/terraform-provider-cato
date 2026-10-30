resource "cato_site_web_proxy_rule" "example" {
  # Omit at on update to preserve the current placement.
  # at = { position = "LAST_IN_POLICY" }
  # Relative positions require ref, for example BEFORE_RULE with an existing rule ID.
  rule = {
    name                            = "example_site_proxy"
    description                     = "Disabled site web proxy example"
    enabled                         = false
    fqdn                            = "proxy.example.com"
    port                            = "3128"
    allowed_destination_protocol    = "BOTH"
    should_associate_with_all_sites  = false
    associated_site                 = [{ name = "existing-site" }] # or { id = "existing-id" }
    authentication_config = {
      method = "NONE" # KERBEROS requires an existing encrypted keytab.
      # kerberos_config = {
      #   is_enabled       = true
      #   encrypted_keytab = var.encrypted_keytab # Sensitive encrypted content, not a raw keytab file.
      # }
    }
  }
}
