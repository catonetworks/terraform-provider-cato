# As a best practice, add custom categories to map to the following phases for TLS Inspection used in rules
# https://support.catonetworks.com/hc/en-us/articles/13314286857501-Working-with-Categories
# TLSi_Phase_1: Alcohol and Tobacco, Anonymizers, Botnets, Cheating, Compromised, Criminal Activity, Cults, Hacking, Illegal Drugs, Keyloggers, Malware, Nudity, P2P, Parked domains, Phishing, Porn, Questionable, SPAM, Spyware, Tasteless, Violence and Hate, Weapons, 
# TLSi_Phase_2: Entertainment, Gambling, Games, Greeting Cards, Leisure and Recreation, News, Politics, Real Estate, Religion, Sex education, Shopping, Sports, Vehicles, Web Hosting, 

# Create section for custom TLS rules 
resource "cato_tls_section" "section1" {
    at = {
        position = "LAST_IN_POLICY"
    }
    section = {
        name = "Custom TLS Rules"
    }    
}

# Create first section for TLS phases
resource "cato_tls_section" "section2" {
    at      = {
        position = "AFTER_SECTION"
        ref = cato_tls_section.section1.section.id
    }
    section = {
        name = "TLS Phases"
    }    
}

# Create rules in sections
resource "cato_tls_rule" "TLSi_Phase_1" {
    at = {
        position = "FIRST_IN_SECTION"
        ref = cato_tls_section.section2.section.id
    }
    rule = {
        action                       = "INSPECT"
        application                  = {
            custom_category = [
                {
                    # id   = "12345"
                    name = "TLSi_Phase_1"
                },
            ]
        }
        connection_origin            = "ANY"
        enabled                      = true
        name                         = "TLSi_Phase_1_inspection"
        source                       = {
            # Filter phase 1 to applyto specific set of sites or users
            users_group = [
                {
                    # id   = "12345"
                    name = "Your User Group Here"
                },
            ]
            site = [
                {
                    # id   = "12345"
                    name = "Your Site Name"
                },
            ]
        }
        untrusted_certificate_action = "ALLOW"
    }
}

resource "cato_tls_rule" "TLSi_Phase_2" {
    at   = {
        position = "AFTER_RULE"
        ref = cato_tls_rule.TLSi_Phase_1.id
    }
    rule = {
        action                       = "INSPECT"
        application                  = {
            custom_category = [
                {
                    id   = "37294"
                    name = "TLSi_Phase_2"
                },
            ]
        }
        connection_origin            = "ANY"
        enabled                      = true
        name                         = "TLSi_Phase_2_inspection"
        source                       = {
            users_group = [
                {
                    # id   = "12345"
                    name = "Your User Group Here"
                },
            ]
            site = [
                {
                    # id   = "12345"
                    name = "Your Site Name"
                },
            ]
        }
        untrusted_certificate_action = "ALLOW"
    }
}

resource "cato_tls_rule" "TLSi_Phase_2" {
    at   = {
        position = "AFTER_RULE"
        ref = cato_tls_rule.TLSi_Phase_2.id
    }
    rule = {
        action                       = "INSPECT"
        application                  = {}
        connection_origin            = "ANY"
        enabled                      = true
        name                         = "TLSi_Phase_3_inspection"
        source                       = {}
        untrusted_certificate_action = "ALLOW"
    }
}

resource "cato_if_rule" "example_kitchen_sink" {
    at   = {
      position = "LAST_IN_SECTION"
      ref      = cato_tls_section.section1.section.id
    }
    rule = {
        name              = "Internet Firewall Kitchen Sink"
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