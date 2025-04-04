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

resource "cato_if_rule" "kitchen_sink" {
  depends_on = [ cato_if_section.if_section ]
  at = {
    position = "LAST_IN_SECTION"
    ref=cato_if_section.if_section.section.id
  }
  rule = {
    name        = "Internet Firewall Min - test3"
    action            = "ALLOW"
    connection_origin = "REMOTE"
    country = [
      {
        name = "Afghanistan"
      },
    ]
    description = "test description"
    destination = {
      app_category = [
        {
          name = "Alcohol and Tobacco"
        },
        {
          id = "anonymizers"
        },
      ]
      application = [
        {
          name = "assembled"
        },
      ]
      country = [
        {
          name = "Afghanistan"
        },
      ]
      custom_app = [
        {
          name = "Test Custom App"
        },
      ]
      custom_category = [
        {
          name = "Test Custom Category"
        },
      ]
      domain = [
        "testdomain.com",
      ]
      fqdn = [
        "my.testedomain.com",
      ]
      # ip = []
      ip_range = [
        {
          from = "192.168.11.6"
          to   = "192.168.11.8"
        },
      ]
      remote_asn = [
        "12345",
      ]
      sanctioned_apps_category = [
        {
          name = "Sanctioned Apps"
        },
      ]
#       subnet = []
    }
    device = [
      {
        name = "Test Device Posture Profile"
      },
    ]
    device_os = ["ANDROID", "EMBEDDED", "IOS", "LINUX", "MACOS", "WINDOWS"]
    enabled = true
    exceptions = [
      {
        connection_origin = "REMOTE"
        country = [
          {
            name = "American Samoa"
          },
        ]
        destination = {
          app_category = [
            {
              name = "Anonymizers"
            },
          ]
          application = [
            {
              name = "Kyriba Corp."
            },
          ]
          country = [
            {
              name = "Anguilla"
            },
          ]
          custom_app = [
            {
              name = "Test Custom App"
            },
          ]
          custom_category = [
            {
              name = "Test Custom Category"
            },
          ]
          domain = [
            "test.com",
          ]
          fqdn = [
            "my.test.com",
          ]
          ip_range = [
            {
              from = "192.168.11.2"
              to   = "192.168.11.3"
            },
          ]
          sanctioned_apps_category = [
            {
              name = "Sanctioned Apps"
            },
          ]
        }
        name = "Kitchen Sink Exception3"
        service = {
          custom = [
            {
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
          floating_subnet = [
            {
              name = "floating_range"
            },
          ]
          group = [
            {
              name = "test group"
            },
          ]
          host = [
            {
              name = "Test Host"
            },
          ]
          ip = [
            "192.168.11.2",
          ]
          ip_range = [
            {
              from = "192.168.11.2"
              to   = "192.168.11.4"
            },
          ]
          network_interface = [
            {
              name = "Test IPSec \\ Default"
            },
          ]
          site = [
            {
              name = "test aws socket"
            },
          ]
          site_network_subnet = [
            {
              name = "Test IPSec \\ Default \\ Native Range"
            },
          ]
          system_group = [
            {
              name = "All Floating Ranges"
            },
          ]
          user = [
            {
              name = "test user"
            },
          ]
          users_group = [
            {
              name = "test group"
            },
          ]
        }
      },
    ]
    schedule = {
      active_on = "CUSTOM_RECURRING"
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
#     # section = {
#     #   id   = "db9d857c-9bfd-4395-b583-f2e70345ee8d"
#     #   name = "IFW QA Section test"
#     # }
    service = {
      custom = [
        {
          port = []
          port_range = {
            from = "10"
            to   = "50"
          }
          protocol = "UDP"
        },
        {
          port = [
            "22",
          ]
          protocol = "UDP"
        },
      ]
      standard = [
        {
          name = "Amazon EC2"
        },
      ]
    }
    source = {
      floating_subnet = [
        {
          id = "1474041"
          # name = "test subnet"
        },
      ]
      group = [
        {
          name = "test group"
        },
      ]
      host = [
        {
          name = "Test Host"
        },
      ]
      ip = [
        "192.168.11.2",
      ]
      ip_range = [
        {
          from = "192.168.11.2"
          to   = "192.168.11.5"
        },
      ]
      network_interface = [
        {
          # name = "Test IPSec \\ Default"
          id = "124986"
        },
      ]
      site = [
        {
          name = "test aws socket"
        },
      ]
      site_network_subnet = [
        {
          id = "TjE0Nzk5MTc="
          # name = "Test IPSec \\ Default \\ Native Range"
        },
      ]
      # subnet = []
      system_group = [
        {
          name = "All Floating Ranges"
        },
      ]
      user = [
        {
          name = "test user"
        },
      ]
      users_group = [
        {
          name = "test group"
        },
      ]
    }
    tracking = {
      alert = {
        enabled   = true
        frequency = "WEEKLY"
        mailing_list = [
          {
            # name = "All Admins"
            id = "-100"
          }
        ]
      }
      event = {
        enabled = true
      }
    }
  }
}

