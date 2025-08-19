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
    position = "LAST_IN_POLICY"
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

# cato_if_rule.kitchen_sink:
resource "cato_if_rule" "example_kitchen_sink" {
    at   = {
      position = "LAST_IN_SECTION"
      ref      = cato_if_section.if_section.section.id
    }
    rule = {
        name              = "Internet Firewall Test Kitchen Sink"
        action            = "ALLOW"
        active_period     = {
          effective_from = "2025-08-09T00:00:00"
          expires_at     = "2026-12-31T23:57:59"
        }
        connection_origin = "REMOTE"
        country           = [
            {
                id   = "AG"
                # name = "Antigua and Barbuda"
            },
            {
                # id   = "AW"
                name = "Aruba"
            },
        ]
        description       = "test description"
        destination       = {
            app_category             = [
                {
                    # id   = "anonymizers"
                    name = "Anonymizers"
                },
                {
                    # id   = "authentication_services"
                    name = "Authentication Services"
                },
            ]
            application              = [
                {
                    # id   = "ebix"
                    name = "Ebix Inc."
                },
                {
                    id   = "fotocasa"
                    # name = "Fotocasa"
                },
            ]
            country                  = [
                {
                    # id   = "AI"
                    name = "Anguilla"
                },
                {
                    # id   = "AQ"
                    name = "Antarctica"
                },
            ]
            custom_app               = [
                {
                    # id   = "CustomApp_11362_34188"
                    name = "Test Custom App"
                },
            ]
            custom_category          = [
                {
                    # id   = "24255"
                    name = "Test Custom Category"
                },
                {
                    # id   = "27782"
                    name = "RBI-URLs"
                },
            ]
            domain                   = [
                "test.com",
            ]
            fqdn                     = [
                "www.test.com",
            ]
            global_ip_range          = [
                {
                    # id   = "1757826"
                    name = "global_ip_range"
                },
                {
                    # id   = "1910542"
                    name = "global_ip_range2"
                },
            ]
            ip                       = [
                "1.2.3.4",
            ]
            ip_range                 = [
                {
                    from = "1.2.3.4"
                    to   = "1.2.3.5"
                },
            ]
            remote_asn               = [
                "12",
            ]
            sanctioned_apps_category = [
                {
                    # id   = "22736"
                    name = "Sanctioned Apps"
                },
            ]
            subnet                   = [
                "1.2.3.0/24",
            ]
        }
        device            = [
            {
                # id   = "4202"
                name = "Test Device Posture Profile"
            },
        ]
        device_attributes = {
            category     = [
                "IoT",
                "Mobile",
            ]
            manufacturer = [
                "ADTRAN",
                "ACTi",
            ]
            model        = [
                " 9",
                " 7+",
            ]
            os           = [
                "Aruba OS",
                "Arch Linux",
            ]
            type         = [
                "Appliance",
                "Analog Telephone Adapter",
            ]
        }
        device_os         = [
            "WINDOWS",
            "MACOS",
        ]
        enabled           = true
        schedule          = {
            active_on        = "CUSTOM_RECURRING"
            custom_recurring = {
                days = [
                    "SUNDAY",
                    "WEDNESDAY",
                    "SATURDAY",
                    "THURSDAY",
                    "MONDAY",
                    "TUESDAY",
                    "FRIDAY",
                ]
                from = "02:02:00"
                to   = "03:03:00"
            }
        }
        service           = {
            custom   = [
                {
                    port       = []
                    port_range = {
                        from = "10"
                        to   = "50"
                    }
                    protocol   = "UDP"
                },
                {
                    port     = [
                        "22",
                    ]
                    protocol = "UDP"
                },
            ]
            standard = [
                {
                    # id   = "amazon_ec2"
                    name = "Amazon EC2"
                },
            ]
        }
        source            = {
            floating_subnet     = [
                {
                    id   = "1474041"
                    # name = "floating_range"
                },
                {
                    id   = "1910541"
                    # name = "floating_range2"
                },
            ]
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
                    # id   = "1700335"
                    name = "my.hostname25"
                },
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
                },
                {
                    id   = "175651"
                    # name = "TestSite001 \\ Default"
                },
            ]
            site                = [
                {
                    # id   = "144904"
                    name = "1600LTE"
                },
                {
                    # id   = "144905"
                    name = "1600"
                },
            ]
            site_network_subnet = [
                {
                    id   = "TjE3MDE0NDI="
                    # name = "aws-site \\ LAN \\ Native Range"
                },
                {
                    id   = "TjIyMzcxODI="
                    # name = "Cato-X1600 \\ LAN88 \\ Native Range"
                },
            ]
            subnet              = [
                "1.2.3.0/24",
            ]
            system_group        = [
                {
                    # id   = "2S"
                    name = "All SDP Users"
                },
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
                    # id   = "500000001"
                    name = "Operations Team"
                },
            ]
        }
        tracking          = {
            alert = {
                enabled      = true
                frequency    = "WEEKLY"
                mailing_list = [
                    {
                        id   = "-100"
                        # name = "All Admins"
                    },
                ]
            }
            event = {
                enabled = true
            }
        }
    }
}