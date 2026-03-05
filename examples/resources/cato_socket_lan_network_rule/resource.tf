
# Socket LAN Network Rules control network traffic routing between LAN segments
# Network rules define the transport (LAN/WAN) and can enable NAT

# Example 1: Minimal Network Rule with NAT
# Basic rule for traffic going TO a destination with NAT enabled
resource "cato_socket_lan_network_rule" "minimal_nat" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "LAN to WAN with NAT"
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
    nat = {
      enabled  = true
      nat_type = "DYNAMIC_PAT"
    }
  }
}

# Example 2: Network Rule within a Section
# Place rules within sections for better organization
resource "cato_socket_lan_section" "network_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Network Rules Section"
  }
}

resource "cato_socket_lan_network_rule" "in_section" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_socket_lan_section.network_section.section.id
  }
  rule = {
    name      = "Network Rule in Section"
    enabled   = true
    direction = "BOTH"
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

# Example 3: Network Rule with Service Filtering
# Filter traffic by specific ports and protocols
resource "cato_socket_lan_network_rule" "with_service" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "Web Traffic Only"
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
    service = {
      simple = [
        {
          name = "HTTP"
        },
        {
          name = "HTTPS"
        }
      ]
      custom = [
        {
          port = [
            "8080",
            "8443"
          ]
          protocol = "TCP"
        }
      ]
    }
  }
}

# Example 4: Comprehensive Network Rule (Kitchen Sink)
# Full example with all source, destination, and service options
resource "cato_socket_lan_network_rule" "kitchen_sink" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name        = "Comprehensive Network Rule"
    description = "Full example with all available options"
    enabled     = true
    direction   = "TO"
    transport   = "LAN"
    site = {
      site = [
        {
          name = "My Site"
        }
      ]
    }
    source = {
      # VLAN IDs
      vlan = [
        100,
        200
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
      # Reference hosts by ID or name
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
    service = {
      # Simple service by name
      simple = [
        {
          name = "FTP"
        },
        {
          name = "HTTP"
        }
      ]
      # Custom port/protocol combinations
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
    nat = {
      enabled  = true
      nat_type = "DYNAMIC_PAT"
    }
  }
}
