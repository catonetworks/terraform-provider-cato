resource "cato_socket_bypass_section" "example" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Socket Bypass Example Rules"
  }
}

# Minimal rule — bypass all traffic matched (empty source/destination = match all).
resource "cato_socket_bypass_rule" "simple" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_socket_bypass_section.example.section.id
  }
  rule = {
    name        = "Bypass all traffic"
    enabled     = true
    source      = {}
    destination = {}
  }
}

# Minimal rule — bypass HTTP on WAN1 with event tracking. Note: service.simple is a
# plain list of standard service names (strings), not objects.
resource "cato_socket_bypass_rule" "simple_http" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_socket_bypass_section.example.section.id
  }
  rule = {
    name    = "Bypass HTTP on WAN1"
    enabled = true
    action = {
      event_enabled = true
      port          = "WAN1" # AUTOMATIC | WAN1 | WAN2 | ...
    }
    service = {
      simple = [
        "HTTP",
      ]
    }
    site        = {}
    source      = {}
    destination = {}
  }
}

# Kitchen sink — every predicate the schema supports, referenced by name with the
# matching id commented above each.
resource "cato_socket_bypass_rule" "kitchen_sink" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_socket_bypass_section.example.section.id
  }
  rule = {
    name        = "Socket Bypass Kitchen Sink"
    description = "Kitchen sink example covering every rule-schema field"
    enabled     = true
    action = {
      event_enabled = true
      port          = "WAN1"
    }
    site = {
      group = [
        {
          # id   = "623603"
          name = "test group"
        },
      ]
      site = [
        {
          # id   = "144904"
          name = "site-1"
        },
        {
          # id   = "144905"
          name = "site-2"
        },
      ]
    }
    source = {
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
        {
          # id   = "1910542"
          name = "global_ip_range2"
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
        {
          # id   = "175651"
          name = "second-site-default"
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
      vlan = [
        100,
        200,
      ]
    }
    destination = {
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
      ip = [
        "5.6.7.8",
      ]
      ip_range = [
        {
          from = "5.6.7.8"
          to   = "5.6.7.9"
        },
      ]
      subnet = [
        "5.6.7.0/24",
      ]
    }
    service = {
      custom = [
        {
          port_range = {
            from = "80"
            to   = "81"
          }
          protocol = "TCP"
        },
        {
          port = [
            "443",
          ]
          protocol = "TCP"
        },
      ]
      simple = [
        "HTTP",
        "HTTPS",
      ]
    }
    exception = [
      {
        name = "Socket Bypass Exception"
        site = {
          group = [
            {
              # id   = "623604"
              name = "test group 2"
            },
          ]
          site = [
            {
              # id   = "144904"
              name = "site-1"
            },
          ]
        }
        source = {
          floating_subnet = [
            {
              # id   = "1910541"
              name = "floating_range2"
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
              # id   = "623604"
              name = "test group 2"
            },
          ]
          host = [
            {
              # id   = "1778359"
              name = "host31"
            },
            {
              # id   = "1700335"
              name = "host25"
            },
          ]
          ip = [
            "1.2.3.4",
            "1.2.3.5",
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
          site_network_subnet = [
            {
              # id   = "UzU4OTI1Mw=="
              name = "branch-lan-direct-range"
            },
          ]
          subnet = [
            "1.2.3.0/24",
          ]
          vlan = [
            300,
          ]
        }
        destination = {
          application = [
            {
              # id   = "parsec"
              name = "Parsec"
            },
          ]
          custom_app = [
            {
              # id   = "CustomApp_11362_34188"
              name = "Test Custom App"
            },
          ]
          domain = [
            "www.example.com",
          ]
          fqdn = [
            "my.domain.com",
          ]
          global_ip_range = [
            {
              # id   = "1910542"
              name = "global_ip_range2"
            },
          ]
          ip = [
            "5.6.7.8",
          ]
          ip_range = [
            {
              from = "5.6.7.8"
              to   = "5.6.7.9"
            },
          ]
          subnet = [
            "5.6.7.0/24",
          ]
        }
        service = {
          custom = [
            {
              port_range = {
                from = "80"
                to   = "81"
              }
              protocol = "TCP"
            },
            {
              port = [
                "443",
              ]
              protocol = "TCP"
            },
          ]
          simple = [
            "HTTP",
          ]
        }
      },
    ]
  }
}
