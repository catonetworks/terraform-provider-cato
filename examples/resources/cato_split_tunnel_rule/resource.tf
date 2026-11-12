

// route selected destinations through the tunnel
resource "cato_split_tunnel_rule" "route_selected" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name             = "Route Selected Destinations"
    enabled          = true
    action           = "ROUTE_ONLY_SELECTED"
    routing_priority = "LAN"
  }
}

resource "cato_split_tunnel_section" "st_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Split Tunnel Section"
  }
}

# cato_split_tunnel_rule.kitchen_sink:
resource "cato_split_tunnel_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_split_tunnel_section.st_section.section.id
  }
  rule = {
    name             = "Split Tunnel Kitchen Sink"
    description      = "test"
    enabled          = true
    action           = "ROUTE_ONLY_SELECTED"
    routing_priority = "LAN"
    coverage         = "ALL"
    platform = [
      "WINDOWS",
      "MACOS",
    ]
    country = [
      {
        # id   = "AG"
        name = "Antigua and Barbuda"
      },
    ]
    destination = {
      application = [
        {
          # id   = "appboy"
          name = "AppBoy By Braze, Inc."
        },
      ]
      domain = [
        "something.com",
        "something2.com",
      ]
      fqdn = [
        "www.something.com",
      ]
      global_ip_range = [
        {
          # id   = "1757826"
          name = "global_ip_range"
        },
      ]
    }
    dns_exclusion = {
      domain = [
        "excluded.something.com",
      ]
    }
    source = {
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
    source_network = {
      source_network_type = "ANY_MANAGED_NETWORK"
    }
  }
}
