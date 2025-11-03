resource "cato_wnw_section" "wnw_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Custom WAN Network Rules"
  }
}

// wan network kitchen sink rule example
resource "cato_wnw_rule" "min" {
  at = {
    position = "LAST_IN_SECTION"
    ref = cato_wnw_section.wnw_section.section.id
  }
  rule = {
    name        = "Example WAN Minimal Rule"
    description = "example description"
    enabled     = true
    route_type  = "NONE"
    rule_type   = "WAN"
    bandwidth_priority = {
      # id   = "-1"
      name = "255"
    }
    configuration = {
      active_tcp_acceleration = false
      packet_loss_mitigation  = true
      preserve_source_port    = false
      primary_transport = {
        primary_interface_role   = "WAN1"
        secondary_interface_role = "WAN1"
        transport_type           = "ALTERNATIVE_WAN"
      }
      secondary_transport = {
        primary_interface_role   = "AUTOMATIC"
        secondary_interface_role = "NONE"
        transport_type           = "NONE"
      }
    }
    application = {}
    source = {}
    destination = {}
    tracking = {
      event = {
        enabled = false
      }
    }
  }
}

resource "cato_wnw_rule" "kitchen_sink" {
    at   = {
        position = "LAST_IN_POLICY"
    }
    rule = {
        name               = "WAN Network Kitchen Sink Rule"
        description        = "example description"
        enabled            = true
        route_type         = "NONE"
        rule_type          = "WAN"
        configuration      = {
            active_tcp_acceleration = true
            packet_loss_mitigation  = true
            preserve_source_port    = false
            primary_transport       = {
                primary_interface_role   = "AUTOMATIC"
                secondary_interface_role = "NONE"
                transport_type           = "WAN"
            }
            secondary_transport     = {
                primary_interface_role   = "AUTOMATIC"
                secondary_interface_role = "NONE"
                transport_type           = "AUTOMATIC"
            }
        }
        bandwidth_priority = {
            # id   = "-1"
            name = "255"
        }
        application        = {
            app_category      = [
                {
                    # id   = "botnets"
                    name = "Botnets"
                },
            ]
            application       = [
                {
                    # id   = "wordhero"
                    name = "Wordhero"
                },
            ]
            custom_app        = [
                {
                    # id   = "CustomApp_11362_34188"
                    name = "Test Custom App"
                },
            ]
            custom_category   = [
                {
                    # id   = "27782"
                    name = "RBI-URLs"
                },
            ]
            custom_service    = [
                {
                    port     = [
                        "80",
                    ]
                    protocol = "TCP"
                },
                {
                    port_range = {
                        from = "80"
                        to   = "81"
                    }
                    protocol   = "TCP"
                },
            ]
            custom_service_ip = [
                {
                    ip   = "1.2.3.4"
                    name = "service1"
                },
                {
                    ip_range = {
                        from = "1.2.0.0"
                        to   = "1.2.255.255"
                    }
                    name     = "service2"
                },
                {
                    ip_range = {
                        from = "10.0.0.1"
                        to   = "10.0.0.24"
                    }
                    name     = "Service3"
                },
            ]
            domain            = [
                "something.com",
            ]
            fqdn              = [
                "www.something.com",
            ]
            service           = [
                {
                    # id   = "THREEPC"
                    name = "3PC"
                },
            ]
        }
        destination        = {
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
                    # Cato API does not like names with \\ in the value
                },
            ]
            site                = [
                {
                    # id   = "144905"
                    name = "1600"
                },
            ]
            site_network_subnet = [
                {
                    id   = "UzU4OTI1Mw=="
                    # name = "1600LTE \\ INT_5 \\ Direct Network Range"
                    # Cato API does not like names with \\ in the value
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
        exceptions         = [
            {
                application = {
                    app_category      = [
                        {
                            # id   = "advertisements"
                            name = "Advertisements"
                        },
                    ]
                    application       = [
                        {
                            # id   = "contentquo"
                            name = "Contentquo Oü"
                        },
                    ]
                    custom_app        = [
                        {
                            # id   = "CustomApp_11362_34188"
                            name = "Test Custom App"
                        },
                    ]
                    custom_category   = [
                        {
                            # id   = "27782"
                            name = "RBI-URLs"
                        },
                    ]
                    custom_service    = [
                        {
                            port     = [
                                "80",
                            ]
                            protocol = "TCP"
                        },
                    ]
                    custom_service_ip = [
                        {
                            ip   = "1.2.3.4"
                            name = "service1"
                        },
                        {
                            ip_range = {
                                from = "1.2.0.0"
                                to   = "1.2.255.255"
                            }
                            name     = "service2"
                        },
                        {
                            ip_range = {
                                from = "10.0.0.1"
                                to   = "10.0.0.24"
                            }
                            name     = "Service3"
                        },
                    ]
                    domain            = [
                        "something.com",
                    ]
                    fqdn              = [
                        "www.something.com",
                    ]
                    service           = [
                        {
                            # id   = "THREEPC"
                            name = "3PC"
                        },
                    ]
                }
                destination = {
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
                        "1.2.2.4",
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
                            # Cato API does not like names with \\ in the value
                        },
                    ]
                    site                = [
                        {
                            # id   = "144905"
                            name = "1600"
                        },
                    ]
                    site_network_subnet = [
                        {
                            id   = "UzU4OTI1Mw=="
                            # name = "1600LTE \\ INT_5 \\ Direct Network Range"
                            # Cato API does not like names with \\ in the value
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
                name        = "Kitchen Sink Exception"
                source      = {
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
                            # Cato API does not like names with \\ in the value
                        },
                    ]
                    site                = [
                        {
                            # id   = "144905"
                            name = "1600"
                        },
                    ]
                    site_network_subnet = [
                        {
                            id   = "UzU4OTI1Mw=="
                            # name = "1600LTE \\ INT_5 \\ Direct Network Range"
                            # Cato API does not like names with \\ in the value
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
                            # id   = "500000001"
                            name = "Operations Team"
                        },
                    ]
                }
            },
        ]
        source             = {
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
                    # Cato API does not like names with \\ in the value
                },
            ]
            site                = [
                {
                    # id   = "144905"
                    name = "1600"
                },
            ]
            site_network_subnet = [
                {
                    id   = "UzU4OTI1Mw=="
                    # name = "1600LTE \\ INT_5 \\ Direct Network Range"
                    # Cato API does not like names with \\ in the value
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
                    # id   = "500000001"
                    name = "Operations Team"
                },
            ]
        }
    }
}