// wan firewall allowing all & logs
resource "cato_wf_rule" "allow_all_and_log" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "Allow all & logs"
    enabled   = true
    action    = "ALLOW"
    direction = "BOTH"
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

// all SMBV3 for all domain users to the site named Datacenter
resource "cato_wf_rule" "allow_smbv3_to_dc" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "Allow SMBv3 to DC"
    enabled   = true
    action    = "ALLOW"
    direction = "TO"
    source = {
      users_group = [
        {
          name = "Domain Users"
        }
      ]
    }
    destination = {
      site = [
        {
          name = "Datacenter"
        }
      ]
    }
    service = {
      standard = [
        {
          name = "SMBV3"
        }
      ]
    }
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

// Comprehensive kitchen sink rule example
resource "cato_wf_rule" "example_kitchen_sink" {
    at   = {
        position = "LAST_IN_POLICY"
    }
    rule = {
        name              = "WAN Firewall Test Kitchen Sink"
        action            = "ALLOW"
        active_period     = {
            effective_from     = "2025-08-31T10:10:10"
            expires_at         = "2026-12-31T11:11:12"
        }
        application       = {
            app_category             = [
                {
                    # id   = "advertisements"
                    name = "Advertisements"
                },
            ]
            application              = [
                {
                    # id   = "appboy"
                    name = "AppBoy By Braze, Inc."
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
                "something.com",
                "something2.com",
            ]
            fqdn                     = [
                "www.something.com",
            ]
            global_ip_range          = [
                {
                    # id   = "1757826"
                    name = "global_ip_range"
                },
            ]
            ip_range                 = [
                {
                    from = "1.2.3.1"
                    to   = "1.2.3.4"
                },
            ]
            sanctioned_apps_category = [
                {
                    # id   = "22736"
                    name = "Sanctioned Apps"
                },
            ]
        }
        connection_origin = "SITE"
        country           = [
            {
                # id   = "AG"
                name = "Antigua and Barbuda"
            },
        ]
        description       = "WAN Firewall Test Kitchen Sink"
        destination       = {
            floating_subnet     = [
                {
                    id   = "1474041"
                    # name = "floating_range"
                },
            ]
            global_ip_range     = [
                {
                    # id   = "1757826"
                    name = "global_ip_range"
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
            ]
            site                = [
                {
                    # id   = "117560"
                    name = "aws-site"
                },
            ]
            site_network_subnet = [
                {
                    id   = "TjE0Nzk5MTc="
                    # name = "ipsec-dev-site \\ Default \\ Native Range"
                },
            ]
            system_group        = [
                {
                    # id   = "7S"
                    name = "All Floating Ranges"
                },
            ]
            user                = [
                {
                    id   = "0"
                    # name = "test user"
                },
            ]
            users_group         = [
                {
                    id   = "500000001"
                    # name = "Operations Team"
                },
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
                "2N",
                "2N, part of Axis group",
            ]
            model        = [
                " 3",
                " 4",
                "002RMZ",
            ]
            os           = [
                "Amazon Linux 2",
                "Android",
            ]
            type         = [
                "105",
                "3D Printer",
            ]
        }
        device_os         = [
            "WINDOWS",
        ]
        direction         = "TO"
        enabled           = true
        exceptions        = [
            {
                application       = {
                    app_category             = [
                        {
                            # id   = "advertisements"
                            name = "Advertisements"
                        },
                    ]
                    application              = [
                        {
                            # id   = "parsec"
                            name = "Parsec"
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
                    ]
                    domain                   = [
                        "www.example.com",
                    ]
                    fqdn                     = [
                        "my.domain.com",
                    ]
                    global_ip_range          = [
                        {
                            # id   = "1757826"
                            name = "global_ip_range"
                        },
                    ]
                    ip_range                 = [
                        {
                            from = "1.2.3.4"
                            to   = "1.2.3.5"
                        },
                    ]
                    sanctioned_apps_category = [
                        {
                            # id   = "22736"
                            name = "Sanctioned Apps"
                        },
                    ]
                }
                connection_origin = "ANY"
                country           = [
                    {
                        # id   = "AF"
                        name = "Afghanistan"
                    },
                ]
                destination       = {
                    floating_subnet     = [
                        {
                            # id   = "1474041"
                            name = "floating_range"
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
                        "1.2.3.5",
                        "1.2.3.6",
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
                    ]
                    site                = [
                        {
                            # id   = "117560"
                            name = "aws-site"
                        },
                    ]
                    site_network_subnet = [
                        {
                            id   = "TjE0Nzk5MTc="
                            # name = "ipsec-dev-site \\ Default \\ Native Range"
                        },
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
                }
                device            = [
                    {
                        # id   = "4202"
                        name = "Test Device Posture Profile"
                    },
                ]
                device_os         = [
                    "WINDOWS",
                    "MACOS",
                ]
                direction         = "TO"
                name              = "WAN Exception"
                service           = {
                    custom   = [
                        {
                            port     = [
                                "80",
                            ]
                            protocol = "TCP"
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
                    ]
                    global_ip_range     = [
                        {
                            # id   = "1757826"
                            name = "global_ip_range"
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
                    ]
                    site                = [
                        {
                            # id   = "101055"
                            name = "ipsec-dev-site"
                        },
                        {
                            # id   = "117560"
                            name = "aws-site"
                        },
                    ]
                    site_network_subnet = [
                        {
                            id   = "TjE0Nzk5MTc="
                            # name = "ipsec-dev-site \\ Default \\ Native Range"
                        },
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
                            # id   = "500000001"
                            name = "Operations Team"
                        },
                    ]
                }
            },
        ]
        schedule          = {
            active_on = "ALWAYS"
        }
        service           = {
            custom   = [
                {
                    port     = [
                        "80",
                    ]
                    protocol = "TCP"
                },
                {
                    port     = [
                        "81",
                    ]
                    protocol = "TCP"
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
                    # id   = "1474041"
                    name = "floating_range"
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
                    id   = "124988"
                    # name = "test aws socket \\ LAN"
                },
            ]
            site                = [
                {
                    # id   = "101057"
                    name = "test aws socket"
                },
                {
                    # id   = "117560"
                    name = "aws-site"
                },
            ]
            site_network_subnet = [
                {
                    id   = "TjE0Nzk5MTc="
                    # name = "ipsec-dev-site \\ Default \\ Native Range"
                },
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
                    # id   = "500000001"
                    name = "Operations Team"
                },
            ]
        }
        tracking          = {
            alert = {
                enabled      = false
                frequency    = "DAILY"
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