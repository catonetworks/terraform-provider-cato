resource "cato_site_web_proxy_section" "example" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Site Web Proxy Example Rules"
  }
}

# Minimal rule — only required fields populated.
resource "cato_site_web_proxy_rule" "simple" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_site_web_proxy_section.example.section.id
  }
  rule = {
    name    = "Proxy example.com"
    enabled = true
    fqdn    = "proxy.example.com"
    port    = "8080"
  }
}

# Kitchen sink — every field the schema supports. associated_site is a catalog
# reference and is shown by name with the matching id commented above; method and
# kerberos_config are literal configuration values (no catalog/id form in the schema).
# encrypted_keytab is sensitive — use a variable or secrets manager, never a literal.
resource "cato_site_web_proxy_rule" "kitchen_sink" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_site_web_proxy_section.example.section.id
  }
  rule = {
    name                            = "Site Web Proxy Kitchen Sink"
    description                     = "Kitchen sink example covering every rule-schema field"
    enabled                         = true
    fqdn                            = "proxy.kitchensink.example.com"
    port                            = "8443"
    allowed_destination_protocol    = "BOTH" # HTTPS | BOTH
    should_associate_with_all_sites = false
    associated_site = [
      {
        # id   = "144904"
        name = "site-1"
      },
      {
        # id   = "144905"
        name = "site-2"
      },
    ]
    authentication_config = {
      method = "KERBEROS" # KERBEROS | NONE
      kerberos_config = {
        is_enabled = true
        # Sensitive — obtain from mutation.webProxyConfig.encryptKeytabFile and supply
        # via a variable or secrets manager, never a literal in committed code.
        encrypted_keytab = "REPLACE_WITH_ENCRYPTED_KEYTAB"
      }
    }
  }
}
