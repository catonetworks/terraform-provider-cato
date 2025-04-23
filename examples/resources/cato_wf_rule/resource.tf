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
resource "cato_wf_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_POLICY"
    # ref = cato_wf_section.wf_section.section.id
  }
  rule = {
    direction = "TO"
    enabled   = true
    name      = "WAN Firewall Kitchen Sink Example"
    action    = "ALLOW"
    application = {
      app_category = [
        {
          name = "Advertisements"
        },
      ]
      application = [
        {
          name = "AppBoy By Braze, Inc."
        },
      ]
      custom_app = [
        {
          name = "My Custom App"
        },
      ]
      custom_category = [
        {
          name = "Some Custom Category"
        },
        {
          name = "RBI-URLs"
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
          name = "global_ip_range"
        },
      ]
      ip_range = [
        {
          from = "1.2.3.1"
          to   = "1.2.3.4"
        },
      ]
      sanctioned_apps_category = [
        {
          name = "Sanctioned Apps"
        },
      ]
    }
    connection_origin = "SITE"
    country = [
      {
        name = "Antigua and Barbuda"
      },
    ]
    description = "WAN Firewall Test Kitchen Sink"
    destination = {
      host = [
        {
          # id = "000000"
          name = "my.hostname25"
        },
      ]
      site = [
        {
          // id = "000000",
          name = "aws-site"
        }
      ]
      ip = [
        "1.2.3.4"
      ]
      ip_range = [
        {
          from: "1.2.3.4",
          to: "1.2.3.5"
        }
      ]
      global_ip_range = [
        {
          // id = "000000",
          name = "global_ip_range"
        }
      ]
      network_interface = [
        {
          # id = "000000"
          name = "Test IPSec \\ Default"
        },
      ]
      site_network_subnet = [
        {
          name = "Test IPSec \\ Default \\ Native Range"
          # id = "ABCDEFG="
        },
      ]
      floating_subnet = [
        {
          id = "123456",
          # name = "floating_range"          
        },
      ]
      user = [
        {
          # id = "0",
          name = "test user"
        },
      ]
      users_group = [
        {
          # id: "0000000",
          name = "test group"
        },
      ]
      group = [
        {
          # id: "00",
          name = "test group"
        },
      ]
      system_group = [
        {
          # id: "00",
          name = "All Floating Ranges"
        },
      ]
    }
    device = [
      {
        name = "Device Posture Profile"
      },
    ]
    device_os = [
      "WINDOWS",
    ]
    schedule = {
      active_on = "ALWAYS"
    }
    service = {
      custom = [
        {
          port     = ["80"]
          protocol = "TCP"
        },
        {
          port     = ["81"]
          protocol = "TCP"
        },
      ]
      standard = [
        {
          name = "Agora"
        },
      ]
    }
    source = {
      host = [
        {
          # id = "123456"
          name = "my.hostname25"
        },
      ]
      site = [
        {
          // id = "123456",
          name = "aws-site"
        },
        {
          # "id": "123456",
          name = "test aws socket"
        }
      ]
      ip = [
        "1.2.3.4"
      ]
      ip_range = [
        {
          from: "1.2.3.4",
          to: "1.2.3.5"
        }
      ]
      global_ip_range = [
        {
          // id = "123456",
          name = "global_ip_range2"
        },
        {
          // id = "123456",
          name = "global_ip_range"
        }
      ]
      network_interface = [
        {
            "id": "123456",
            # "name": "test aws socket \\ LAN"
        },
        {
          id = "123456"
          // name = "Test IPSec \\ Default"
        },
      ]
      site_network_subnet = [
        {
          # name = "Test IPSec \\ Default \\ Native Range"
          id = "ABCDEFG="
        },
      ]
      floating_subnet = [
        {
          id = "123456",  // TODO see why this does not take name
          # name = "floating_range"          
        },
      ]
      user = [
        {
          # id = "0",
          name = "test user"
        },
      ]
      users_group = [
        {
          # id: "0000000",
          name = "test group"
        },
      ]
      group = [
        {
          # id: "000",
          name = "test group"
        },
      ]
      system_group = [
        {
          # id: "000",
          name = "All Floating Ranges"
        },
      ]
    }
    exceptions = [
      {
        application = {
          app_category = [
            {
              #id   = "advertisements"
              name = "Advertisements"
            },
          ]
          application = [
            {
              #id   = "parsec"
              name = "Parsec"
            },
          ]
          custom_app = [
            {
              #id   = "CustomApp_11362_34188"
              name = "Test Custom App"
            },
          ]
          custom_category = [
            {
              #id   = "123456"
              name = "Test Custom Category"
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
              #id   = "123456"
              name = "global_ip_range"
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
          sanctioned_apps_category = [
            {
              #id   = "123456"
              name = "Sanctioned Apps"
            },
          ]
        }
        connection_origin = "ANY"
        country = [
          {
            # id   = "AF"
            name = "Afghanistan"
          },
        ]
        destination = {
          floating_subnet = [
            {
              id   = "123456"
              #name = "floating_range"
            },
          ]
          group = [
            {
              id   = "123456"
              #name = "test group"
            },
          ]
          host = [
            {
              #id   = "123456"
              name = "my.hostname25"
            },
            {
              #id   = "123456"
              name = "host31"
            },
          ]
          ip = [
            "1.2.3.4",
            "1.2.3.5",
            "1.2.3.6",
          ]
          ip_range = [
            {
              from = "1.2.3.4"
              to   = "1.2.3.5"
            },
          ]
          network_interface = [
            {
              id   = "123456"
              #name = "IPSec Site \\ Default"
            },
          ]
          site = [
            {
              #id   = "123456"
              name = "aws-site"
            },
          ]
          site_network_subnet = [
            {
              id   = "TjE0Nzk5MTc="
              #name = "IPSec Site \\ Default \\ Native Range"
            },
          ]
          system_group = [
            {
              #id   = "000"
              name = "All Floating Ranges"
            },
          ]
          user = [
            {
              #id   = "0"
              name = "test user"
            },
          ]
        }
        device = [
          {
            # id   = "1234"
            name = "Test Device Posture Profile"
          },
        ]
        device_os = [
          "WINDOWS",
          "MACOS",
        ]
        direction = "TO"
        name      = "WAN Exception"
        service = {
          custom = [
            {
              port = [
                "80",
              ]
              protocol = "TCP"
            },
          ]
          standard = [
            {
              #id   = "amazon_ec2"
              name = "Amazon EC2"
            },
          ]
        }
        source = {
          floating_subnet = [
            {
              id   = "123456"
            },
          ]
          global_ip_range = [
            {
              #id   = "123456"
              name = "global_ip_range"
            },
          ]
          group = [
            {
              #id   = "123456"
              name = "test group"
            },
          ]
          host = [
            {
              #id   = "123456"
              name = "my.hostname25"
            },
            {
              #id   = "123456"
              name = "host31"
            },
          ]
          ip = [
            "1.2.3.4",
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
              id   = "124986"
              #name = "IPSec Site \\ Default"
            },
          ]
          site = [
            {
              #id   = "123456"
              name = "aws-site"
            },
            {
              #id   = "123456"
              name = "ipsec-test-site"
            },
          ]
          site_network_subnet = [
            {
              id   = "ABCDEFG="
              #name = "IPSec Site \\ Default \\ Native Range"
            },
          ]
          system_group = [
            {
              #id   = "000"
              name = "All Floating Ranges"
            },
          ]
          user = [
            {
              #id   = "0"
              name = "test user"
            },
          ]
          users_group = [
            {
              #id   = "00000000"
              name = "Operations Team"
            },
          ]
        }
      },
    ]
    tracking = {
      alert = {
        enabled   = true
        frequency = "DAILY"
      }
      event = {
        enabled = true
      }
    }
  }
}


