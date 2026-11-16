resource "cato_app_tenant_restriction_section" "example" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "App Tenant Restriction Example Rules"
  }
}

# Minimal rule — action BYPASS. Allows traffic to bypass tenant restriction enforcement
# for the specified application.
resource "cato_app_tenant_restriction_rule" "simple" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_app_tenant_restriction_section.example.section.id
  }
  rule = {
    name     = "Allow corporate Office 365 access"
    enabled  = true
    action   = "BYPASS"
    severity = "LOW" # LOW | MEDIUM | HIGH
    application = {
      # id   = "microsoft_office_login"
      name = "Office365"
    }
    source = {}
  }
}

# action INJECT_HEADERS — injects HTTP headers into matching traffic to enforce tenant
# restrictions (e.g. the Microsoft 365 Restrict-Access-To-Tenants header pattern).
# Header values are marked sensitive in the provider schema.
resource "cato_app_tenant_restriction_rule" "inject_headers" {
  at = {
    position = "AFTER_RULE"
    ref      = cato_app_tenant_restriction_rule.simple.id
  }
  rule = {
    name     = "Restrict M365 access to approved tenants"
    enabled  = true
    action   = "INJECT_HEADERS"
    severity = "MEDIUM"
    application = {
      # id   = "microsoft_office_login"
      name = "Office365"
    }
    headers = [
      {
        name  = "Restrict-Access-To-Tenants"
        value = "your-tenant-id" # sensitive — use a variable or secrets manager
      },
    ]
    # schedule: ALWAYS | WORKING_HOURS | CUSTOM_TIMEFRAME | CUSTOM_RECURRING
    schedule = {
      active_on = "CUSTOM_TIMEFRAME"
      custom_timeframe = {
        from = "2026-01-01T00:00:00Z"
        to   = "2026-12-31T23:59:59Z"
      }
    }
    source = {
      ip = [
        "198.51.100.50",
      ]
    }
  }
}

# Kitchen sink — every source predicate the schema supports, referenced by name with
# the matching id commented above each.
resource "cato_app_tenant_restriction_rule" "kitchen_sink" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_app_tenant_restriction_section.example.section.id
  }
  rule = {
    name        = "App Tenant Restriction Kitchen Sink"
    description = "Kitchen sink example covering every rule-schema field"
    enabled     = true
    action      = "INJECT_HEADERS"
    severity    = "HIGH"
    application = {
      # id   = "microsoft_office_login"
      name = "Office365"
    }
    headers = [
      {
        name  = "Restrict-Access-To-Tenants"
        value = "allowed-tenant-id" # sensitive — use a variable or secrets manager
      },
      {
        name  = "Restrict-Access-Context"
        value = "directory-id" # sensitive — use a variable or secrets manager
      },
    ]
    schedule = {
      active_on = "CUSTOM_RECURRING"
      custom_recurring = {
        from = "08:00"
        to   = "18:00"
        days = ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
      }
    }
    source = {
      country = [
        {
          # id   = "AG"
          name = "Antigua and Barbuda"
        },
      ]
      floating_subnet = [
        {
          # id   = "1474041"
          name = "floating_range"
        },
      ]
      global_ip_range = [
        {
          # id   = "1757826"
          name = "global_ip_range"
        },
      ]
      group = [
        {
          # id   = "623603"
          name = "test group"
        },
      ]
      host = [
        {
          # id   = "1778359"
          name = "host31"
        },
      ]
      ip = [
        "1.2.3.4",
      ]
      ip_range = [
        {
          from = "1.2.3.4"
          to   = "1.2.3.5"
        },
      ]
      network_interface = [
        {
          # id   = "124986"
          name = "main-site-default"
        },
      ]
      site = [
        {
          # id   = "144905"
          name = "my-site"
        },
      ]
      site_network_subnet = [
        {
          # id   = "UzU4OTI1Mw=="
          name = "branch-lan-direct-range"
        },
      ]
      subnet = [
        "1.2.3.0/24",
      ]
      system_group = [
        {
          # id   = "7S"
          name = "All Floating Ranges"
        },
      ]
      user = [
        {
          # id   = "0"
          name = "test user"
        },
      ]
      users_group = [
        {
          # id   = "500000000"
          name = "Test User Group"
        },
      ]
    }
  }
}
