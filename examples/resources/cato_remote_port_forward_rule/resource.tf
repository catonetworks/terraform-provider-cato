# NOTE: cato_remote_port_forward_rule requires description, remote_ips, and tracking
# to be fully populated (all Required in this resource's schema) — unlike some sibling
# policy families where those blocks are optional. See schema_remote_port_forward_rule.go.

resource "cato_remote_port_forward_section" "example" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Remote Port Forward Example Rules"
  }
}

# Minimal rule — only required fields populated.
resource "cato_remote_port_forward_rule" "simple" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_remote_port_forward_section.example.section.id
  }
  rule = {
    name        = "Forward RDP"
    description = "Managed by Terraform"
    enabled     = true
    external_ip = {
      # id   = "1234567"
      name = "my-allocated-ip"
    }
    external_port_range = {
      from = "10000"
      to   = "10001"
    }
    internal_ip = "192.0.2.10"
    internal_port_range = {
      from = "443"
      to   = "443"
    }
    forward_icmp     = false
    restriction_type = "ALLOW_LIST" # ALLOW_LIST | BLOCK_LIST
    remote_ips = {
      global_ip_range = []
      ip = [
        "192.0.2.100",
      ]
      ip_range = []
      subnet   = []
    }
    tracking = {
      enabled            = false
      frequency          = "DAILY" # HOURLY | DAILY | WEEKLY | IMMEDIATE
      mailing_list       = []
      subscription_group = []
      webhook            = []
    }
  }
}

# Kitchen sink — every predicate the schema supports, referenced by name with the
# matching id commented above each.
resource "cato_remote_port_forward_rule" "kitchen_sink" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_remote_port_forward_section.example.section.id
  }
  rule = {
    name        = "Remote Port Forwarding Kitchen Sink"
    description = "Kitchen sink example covering every rule-schema field"
    enabled     = true
    external_ip = {
      # id   = "7654321"
      name = "my-allocated-ip-2"
    }
    external_port_range = {
      from = "20000"
      to   = "20010"
    }
    internal_ip = "192.0.2.21"
    internal_port_range = {
      from = "8080"
      to   = "8090"
    }
    forward_icmp     = true
    restriction_type = "BLOCK_LIST"
    remote_ips = {
      global_ip_range = [
        {
          # id   = "1757826"
          name = "global_ip_range"
        },
        {
          # id   = "1910542"
          name = "global_ip_range2"
        },
      ]
      ip = [
        "1.2.3.4",
        "1.2.3.5",
      ]
      ip_range = [
        {
          from = "1.2.3.10"
          to   = "1.2.3.20"
        },
      ]
      subnet = [
        "1.2.3.0/24",
      ]
    }
    tracking = {
      enabled   = true
      frequency = "DAILY"
      mailing_list = [
        {
          # id   = "-100"
          name = "All Admins"
        },
      ]
      subscription_group = [
        {
          # id   = "12345"
          name = "SOC Alerts"
        },
      ]
      webhook = [
        {
          # id   = "12345"
          name = "PagerDuty Hook"
        },
      ]
    }
  }
}
