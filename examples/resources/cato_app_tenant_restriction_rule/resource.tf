# App Tenant Restriction rule — action BYPASS.
# Allows traffic to bypass tenant restriction enforcement for the specified application.
# GraphQL for this policy is marked @beta; fields may change.
resource "cato_app_tenant_restriction_rule" "bypass" {
  # position values: LAST_IN_POLICY | LAST_IN_SECTION | AFTER_RULE | BEFORE_RULE
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_app_tenant_restriction_section.example.section.id
  }

  rule = {
    name        = "TF ATR — Allow corporate Office 365 access"
    description = "Bypass tenant restriction for corporate users on weekdays"
    enabled     = true
    action      = "BYPASS"
    severity    = "MEDIUM" # LOW | MEDIUM | HIGH

    application = {
      id = "microsoft_office_login"
      # name = "Office365"  # use id OR name, not both
    }

    # schedule: ALWAYS | WORKING_HOURS | CUSTOM_TIMEFRAME | CUSTOM_RECURRING
    # Omit schedule to default to ALWAYS.
    schedule = {
      active_on = "CUSTOM_RECURRING"
      custom_recurring = {
        from = "08:00"
        to   = "18:00"
        days = ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
      }
      # active_on = "CUSTOM_TIMEFRAME"
      # custom_timeframe = {
      #   from = "2026-01-01T00:00:00Z"
      #   to   = "2026-12-31T23:59:59Z"
      # }
    }

    source = {
      ip = ["198.51.100.40"]
      # subnet              = ["10.0.0.0/8"]
      # ip_range            = [{ from = "10.1.0.10", to = "10.1.0.20" }]
      # country             = [{ name = "United States" }]
      # host                = [{ name = "my-server" }]
      # site                = [{ id = "site-uuid" }]
      # global_ip_range     = [{ name = "Corporate Ranges" }]
      # network_interface   = [{ id = "nic-uuid" }]
      # site_network_subnet = [{ name = "Branch LAN" }]
      # floating_subnet     = [{ name = "Guest WiFi" }]
      # user                = [{ name = "alice@example.com" }]
      # users_group         = [{ name = "All Users" }]
      # group               = [{ name = "Engineering" }]
      # system_group        = [{ name = "Remote Users" }]
    }
  }
}

# App Tenant Restriction rule — action INJECT_HEADERS.
# Injects HTTP headers into matching traffic to enforce tenant restrictions
# (e.g. Microsoft 365 Restrict-Access-To-Tenants header pattern).
# header values are marked sensitive in the provider schema.
resource "cato_app_tenant_restriction_rule" "inject_headers" {
  at = {
    position = "AFTER_RULE"
    ref      = cato_app_tenant_restriction_rule.bypass.rule.id
  }

  rule = {
    name        = "TF ATR — Inject M365 tenant restriction headers"
    description = "Restrict Microsoft 365 access to approved tenants only"
    enabled     = true
    action      = "INJECT_HEADERS"
    severity    = "HIGH"

    application = {
      id = "microsoft_office_login"
    }

    # Multiple headers can be injected; values are sensitive in the provider schema.
    headers = [
      {
        name  = "Restrict-Access-To-Tenants"
        value = "allowed-tenant-uuid" # sensitive — use a variable or secrets manager
      },
      {
        name  = "Restrict-Access-Context"
        value = "your-directory-id" # sensitive — use a variable or secrets manager
      },
    ]

    schedule = {
      active_on = "ALWAYS"
      # active_on = "WORKING_HOURS"
    }

    source = {
      ip = ["198.51.100.50"]
      # country = [{ name = "Germany" }]
      # site    = [{ id = "site-uuid" }]
    }
  }
}
