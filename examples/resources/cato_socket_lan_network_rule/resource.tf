// Socket LAN Network Rules control routing of traffic between LAN segments (and
// optionally out to WAN) for a site. They define the transport (LAN/WAN), the site(s)
// and group(s) the rule applies to, and can enable NAT. Socket LAN Firewall Rules
// (see cato_socket_lan_firewall_rule) are child rules scoped underneath a network rule.

// simple network rule: route all LAN traffic for a site, no NAT
resource "cato_socket_lan_network_rule" "lan_to_wan" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "LAN to WAN"
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

resource "cato_socket_lan_section" "lan_network_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "LAN Network Section"
  }
}

# cato_socket_lan_network_rule.kitchen_sink:
resource "cato_socket_lan_network_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_socket_lan_section.lan_network_section.section.id
  }
  rule = {
    name        = "LAN Network Rule Kitchen Sink"
    description = "Comprehensive network rule exercising every available field"
    enabled     = true
    direction   = "TO"
    transport   = "LAN"
    site = {
      site = [
        {
          # id   = "117560"
          name = "Branch Office"
        },
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
    }
    source = {
      vlan = [
        100,
        200,
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
    service = {
      simple = [
        {
          name = "FTP"
        },
        {
          name = "HTTP"
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
    nat = {
      enabled  = true
      nat_type = "DYNAMIC_PAT"
    }
  }
}
