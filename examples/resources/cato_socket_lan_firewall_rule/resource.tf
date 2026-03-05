
# Socket LAN Firewall Rules are child rules of Socket LAN Network Rules
# They provide granular firewall control within a network rule's scope
# IMPORTANT: A parent network rule must exist before creating firewall rules

# First, create a parent network rule
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
          name = "My Site"
        }
      ]
    }
    source      = {}
    destination = {}
  }
}

# Example 1: Minimal Firewall Rule - Allow All
# Basic allow rule with event tracking
resource "cato_socket_lan_firewall_rule" "allow_all" {
  at = {
    position = "FIRST_IN_RULE"
    ref      = cato_socket_lan_network_rule.parent_rule.rule.id
  }
  rule = {
    name      = "Allow All Traffic"
    enabled   = true
    direction = "TO"
    action    = "ALLOW"
    source      = {}
    destination = {}
    tracking = {
      event = {
        enabled = true
      }
      alert = {
        enabled   = false
        frequency = "DAILY"
      }
    }
  }
}

# Example 2: Block Rule with Service Filtering
# Block specific ports with weekly alerts
resource "cato_socket_lan_firewall_rule" "block_ports" {
  at = {
    position = "LAST_IN_RULE"
    ref      = cato_socket_lan_network_rule.parent_rule.rule.id
  }
  rule = {
    name      = "Block Insecure HTTP"
    enabled   = true
    direction = "BOTH"
    action    = "BLOCK"
    source      = {}
    destination = {}
    service = {
      custom = [
        {
          port = [
            "80"
          ]
          protocol = "TCP"
        }
      ]
    }
    tracking = {
      event = {
        enabled = true
      }
      alert = {
        enabled   = true
        frequency = "WEEKLY"
        # Optional: Add subscription groups for alerts
        # subscription_group = [
        #   {
        #     id = "group-id"
        #   }
        # ]
      }
    }
  }
}

# Example 3: Firewall Rule with Source/Destination Filtering
# Control traffic between specific VLANs and IPs
resource "cato_socket_lan_firewall_rule" "vlan_control" {
  at = {
    position = "FIRST_IN_RULE"
    ref      = cato_socket_lan_network_rule.parent_rule.rule.id
  }
  rule = {
    name        = "VLAN 100 to VLAN 200"
    description = "Allow traffic from VLAN 100 to VLAN 200"
    enabled     = true
    direction   = "TO"
    action      = "ALLOW"
    source = {
      vlan = [
        100
      ]
      ip = [
        "192.168.1.0/24"
      ]
    }
    destination = {
      vlan = [
        200
      ]
    }
    tracking = {
      event = {
        enabled = true
      }
      alert = {
        enabled = false
        frequency = "DAILY"
      }
    }
  }
}

# Example 4: Comprehensive Firewall Rule (Kitchen Sink)
# Full example with all available options
resource "cato_socket_lan_firewall_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_RULE"
    ref      = cato_socket_lan_network_rule.parent_rule.rule.id
  }
  rule = {
    name        = "Comprehensive Firewall Rule"
    description = "Full example demonstrating all available options"
    enabled     = true
    direction   = "TO"
    action      = "ALLOW"
    source = {
      # VLAN IDs
      vlan = [
        100,
        200
      ]
      # MAC addresses
      mac = [
        "00:11:22:33:44:55",
        "AA:BB:CC:DD:EE:FF"
      ]
      # Specific IP addresses
      ip = [
        "192.168.1.10"
      ]
      # Subnets in CIDR notation
      subnet = [
        "10.0.0.0/24"
      ]
      # IP address ranges
      ip_range = [
        {
          from = "192.168.1.1"
          to   = "192.168.1.100"
        }
      ]
      # Reference sites by name or ID
      site = [
        {
          name = "Branch Office"
        }
      ]
      # Reference hosts by name or ID
      host = [
        {
          name = "server1"
        }
      ]
      # Global IP ranges
      global_ip_range = [
        {
          name = "Corporate IPs"
        }
      ]
      # Floating subnets
      floating_subnet = [
        {
          id = "12345"
        }
      ]
      # Network interfaces
      network_interface = [
        {
          id = "67890"
        }
      ]
      # Site network subnets
      site_network_subnet = [
        {
          id = "subnet123"
        }
      ]
      # System groups
      system_group = [
        {
          name = "All Sites"
        }
      ]
    }
    destination = {
      vlan = [
        300
      ]
      ip = [
        "10.10.10.1"
      ]
      subnet = [
        "172.16.0.0/16"
      ]
      ip_range = [
        {
          from = "10.0.0.1"
          to   = "10.0.0.50"
        }
      ]
      site = [
        {
          name = "Data Center"
        }
      ]
      host = [
        {
          name = "destination-server"
        }
      ]
      global_ip_range = [
        {
          name = "Remote Office IPs"
        }
      ]
    }
    # Application filtering (for application-aware firewall rules)
    application = {
      application = [
        {
          name = "Adobe"
        }
      ]
      custom_app = [
        {
          name = "My Custom App"
        }
      ]
      domain = [
        "example.com"
      ]
      fqdn = [
        "www.example.com"
      ]
      ip = [
        "8.8.8.8"
      ]
      subnet = [
        "8.8.0.0/16"
      ]
      ip_range = [
        {
          from = "9.9.9.1"
          to   = "9.9.9.10"
        }
      ]
      global_ip_range = [
        {
          name = "DNS Servers"
        }
      ]
    }
    # Service/port filtering
    service = {
      simple = [
        {
          name = "FTP"
        },
        {
          name = "HTTP"
        }
      ]
      standard = [
        {
          name = "3PC"
        }
      ]
      custom = [
        {
          port = [
            "443",
            "8443"
          ]
          protocol = "TCP"
        }
      ]
    }
    # Event tracking and alerting
    tracking = {
      event = {
        enabled = true
      }
      alert = {
        enabled   = true
        frequency = "WEEKLY"
        # Webhook notifications
        webhook = [
          {
            name = "My Webhook"
          }
        ]
        # Email mailing list notifications
        mailing_list = [
          {
            name = "Security Team"
          }
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

## Direction Options:
# - TO: Traffic going to the destination
# - BOTH: Traffic in both directions

## Action Options:
# - ALLOW: Allow the traffic
# - BLOCK: Block the traffic

## Alert Frequency Options:
# - HOURLY
# - DAILY
# - WEEKLY
