

// internet firewall allowing all & logs
resource "cato_if_rule" "allow_all_and_log" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name    = "Allow all & logs"
    enabled = true
    action  = "ALLOW"
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

// block all remote users except "Marketing" using category domain "test.com"
resource "cato_if_rule" "block_test_com_for_remote_users" {
  at = {
    position = "FIRST_IN_POLICY"
  }
  rule = {
    name              = "Block Test.com for Remote Users"
    enabled           = true
    action            = "BLOCK"
    connection_origin = "REMOTE"
    destination = {
      domain = [
        "test.com"
      ]
    }
    source = {}
    tracking = {
      event = {
        enabled = true
      }
    }
    exceptions = [
      {
        name = "Exclude Marketing Teams"
        source = {
          users_group = [
            {
              name = "Marketing-Teams"
            }
          ]
        }
      }
    ]
  }
}

resource "cato_if_section" "if_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "IFW Section"
  }
}

resource "cato_if_rule" "example_min" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_if_section.if_section.section.id
  }
  rule = {
    name        = "Internet Firewall Test Kitchen Sink"
    description = "test description"
    enabled     = true
    source      = {}
    destination = {}
    action      = "ALLOW"
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

# cato_tls_rule.kitchen_sink:
resource "cato_tls_rule" "kitchen_sink" {
    at   = {
        position = "LAST_IN_POLICY"
    }
    rule = {
        enabled                      = true
        name                         = "Kitchen Sink New"
        platform                     = "EMBEDDED"
        action                       = "INSPECT"
        application                  = {
            app_category    = [
                {
                    # id   = "advertisements"
                    name = "Advertisements"
                },
            ]
            application     = [
                {
                    # id   = "buildmyteam"
                    name = "buildmyteam"
                },
            ]
            country         = [
                {
                    # id   = "AF"
                    name = "Afghanistan"
                },
            ]
            custom_app      = [
                {
                    # id   = "CustomApp_11362_34188"
                    name = "Test Custom App"
                },
            ]
            custom_category = [
                {
                    # id   = "24255"
                    name = "Test Custom Category"
                },
            ]
            domain          = [
                "something.com",
                "www.something.com",
            ]
            fqdn            = [
                "www.something.com",
            ]
            global_ip_range = [
                {
                    # id   = "1757826"
                    name = "global_ip_range"
                },
            ]
            ip              = [
                "1.2.3.4",
            ]
            ip_range        = [
                {
                    from = "1.2.3.4"
                    to   = "1.2.3.5"
                },
            ]
            remote_asn      = [
                "1234",
            ]
            service         = [
                {
                    # id   = "THREEPC"
                    name = "3PC"
                },
            ]
            subnet          = [
                "1.2.3.0/24",
            ]
        }
        connection_origin            = "REMOTE"
        description                  = "test"
        device_posture_profile       = [
            {
                id   = "4202"
                name = "Test Device Posture Profile"
            },
        ]
        source                       = {
            global_ip_range     = [
                {
                    # id   = "1757826"
                    name = "global_ip_range"
                },
                {
                    # id   = "1910542"
                    name = "global_ip_range2"
                },
            ]
            group               = [
                {
                    # id   = "623603"
                    name = "test group"
                },
            ]
            host                = [
                {
                    # id   = "1778359"
                    name = "host31"
                },
            ]
            ip                  = [
                "1.2.3.4",
            ]
            ip_range            = [
                {
                    from = "1.2.3.4"
                    to   = "1.2.3.5"
                },
            ]
            network_interface   = [
                {
                    id   = "124986"
                    # name = "ipsec-dev-site \\ Default" 
                    # API does not like \\ charaacters in name values
                },
            ]
            site                = [
                {
                    # id   = "144905"
                    # name = "1600"
                },
            ]
            site_network_subnet = [
                {
                    id   = "UzU4OTI1Mw=="
                    # name = "1600LTE \\ INT_5 \\ Direct Network Range" 
                    # API does not like \\ charaacters in name values
                },
            ]
            subnet              = [
                "1.2.3.0/24",
            ]
            system_group        = [
                {
                    # id   = "7S"
                    name = "All Floating Ranges"
                },
            ]
            user                = [
                {
                    # id   = "0"
                    name = "test user"
                },
            ]
            users_group         = [
                {
                    # id   = "500000000"
                    name = "Test User Group"
                },
            ]
        }
        untrusted_certificate_action = "ALLOW"
    }
}