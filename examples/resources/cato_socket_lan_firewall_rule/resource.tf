// Socket LAN Firewall Rules are child rules of Socket LAN Network Rules. They provide
// granular firewall control within a network rule's scope. A parent network rule must
// exist before creating firewall rules under it.

resource "cato_socket_lan_network_rule" "parent_rule" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "Parent Network Rule"
    enabled   = true
    direction = "TO"
    transport = "LAN"
    site = {
      site = [
        {
          # id   = "117560"
          name = "Branch Office"
        },
      ]
    }
    source      = {}
    destination = {}
  }
}

// simple firewall rule: allow all traffic under the parent network rule, with event tracking
resource "cato_socket_lan_firewall_rule" "allow_all" {
  at = {
    position = "FIRST_IN_RULE"
    ref      = cato_socket_lan_network_rule.parent_rule.rule.id
  }
  rule = {
    name        = "Allow All Traffic"
    enabled     = true
    direction   = "TO"
    action      = "ALLOW"
    source      = {}
    destination = {}
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

# cato_socket_lan_firewall_rule.kitchen_sink:
resource "cato_socket_lan_firewall_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_RULE"
    ref      = cato_socket_lan_network_rule.parent_rule.rule.id
  }
  rule = {
    name        = "LAN Firewall Rule Kitchen Sink"
    description = "Comprehensive firewall rule exercising every available field"
    enabled     = true
    direction   = "TO"
    action      = "ALLOW"
    source = {
      vlan = [
        100,
        200,
      ]
      mac = [
        "00:11:22:33:44:55",
        "AA:BB:CC:DD:EE:FF",
      ]
      ip = [
        "192.168.1.10",
      ]
      subnet = [
        "10.0.0.0/24",
      ]
      ip_range = [
        {
          from = "192.168.1.1"
          to   = "192.168.1.100"
        },
      ]
      host = [
        {
          # id   = "1700335"
          name = "server1"
        },
      ]
      site = [
        {
          # id   = "117560"
          name = "Branch Office"
        },
      ]
      group = [
        {
          # id   = "623603"
          name = "Site Group"
        },
      ]
      system_group = [
        {
          # id   = "7S"
          name = "All Floating Ranges"
        },
      ]
      network_interface = [
        {
          # id   = "124986"
          name = "Default"
        },
      ]
      global_ip_range = [
        {
          # id   = "1757826"
          name = "Corporate IPs"
        },
      ]
      floating_subnet = [
        {
          # id   = "1474041"
          name = "Floating Range"
        },
      ]
      site_network_subnet = [
        {
          # id   = "UzU4OTI1Mw=="
          name = "Direct Network Range"
        },
      ]
    }
    destination = {
      vlan = [
        300,
      ]
      ip = [
        "10.10.10.1",
      ]
      subnet = [
        "172.16.0.0/16",
      ]
      ip_range = [
        {
          from = "10.0.0.1"
          to   = "10.0.0.50"
        },
      ]
      host = [
        {
          # id   = "1778359"
          name = "destination-server"
        },
      ]
      site = [
        {
          # id   = "144905"
          name = "Data Center"
        },
      ]
      group = [
        {
          # id   = "623603"
          name = "Site Group"
        },
      ]
      system_group = [
        {
          # id   = "2S"
          name = "All SDP Users"
        },
      ]
      network_interface = [
        {
          # id   = "124988"
          name = "LAN"
        },
      ]
      global_ip_range = [
        {
          # id   = "1910542"
          name = "Remote Office IPs"
        },
      ]
      floating_subnet = [
        {
          # id   = "1474042"
          name = "Standby Floating Range"
        },
      ]
      site_network_subnet = [
        {
          # id   = "TjE0Nzk5MTc="
          name = "Native Range"
        },
      ]
    }
    // Application-aware firewall matching (optional)
    application = {
      application = [
        {
          # id   = "appboy"
          name = "AppBoy By Braze, Inc."
        },
      ]
      custom_app = [
        {
          # id   = "CustomApp_11362_34188"
          name = "Test Custom App"
        },
      ]
      domain = [
        "example.com",
      ]
      fqdn = [
        "www.example.com",
      ]
      ip = [
        "8.8.8.8",
      ]
      subnet = [
        "8.8.0.0/16",
      ]
      ip_range = [
        {
          from = "9.9.9.1"
          to   = "9.9.9.10"
        },
      ]
      global_ip_range = [
        {
          # id   = "1757826"
          name = "DNS Servers"
        },
      ]
    }
    service = {
      simple = [
        {
          name = "FTP"
        },
        {
          name = "HTTP"
        },
      ]
      standard = [
        {
          # id   = "THREEPC"
          name = "3PC"
        },
      ]
      custom = [
        {
          port = [
            "443",
            "8443",
          ]
          protocol = "TCP"
        },
        {
          port_range = {
            from = "8000"
            to   = "8080"
          }
          protocol = "TCP"
        },
      ]
    }
    tracking = {
      event = {
        enabled = true
      }
      alert = {
        enabled   = true
        frequency = "WEEKLY"
        subscription_group = [
          {
            # id   = "500000001"
            name = "Operations Team"
          },
        ]
        webhook = [
          {
            # id   = "7001"
            name = "My Webhook"
          },
        ]
        mailing_list = [
          {
            # id   = "-100"
            name = "Security Team"
          },
        ]
      }
    }
  }
}

## Position Options for 'at' block:
# - FIRST_IN_RULE: Insert at the beginning of the parent rule's firewall rules
# - LAST_IN_RULE: Insert at the end of the parent rule's firewall rules
# - BEFORE_RULE: Insert before a specific firewall rule (requires ref)
# - AFTER_RULE: Insert after a specific firewall rule (requires ref)
