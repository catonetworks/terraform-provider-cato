# NOTE on schema shape: rule.application_rule / rule.data_rule / rule.file_rule are a
# one-of selected by rule.rule_type (APPLICATION | DATA | FILE) — only one may be
# populated per rule resource. All three share an identical field shape (source,
# application, tracking, device, access_method, action_config, schedule, file_attribute,
# application_activity); dlp_profile is DATA-only, and application = {} (matching all
# cloud apps) is only valid for FILE rules.

resource "cato_application_control_section" "example" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Application Control Example Rules"
  }
}

# Minimal rule — rule_type APPLICATION, only required/common fields.
resource "cato_application_control_rule" "simple" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_application_control_section.example.section.id
  }
  rule = {
    name      = "Block Slack"
    enabled   = true
    rule_type = "APPLICATION"
    application_rule = {
      action   = "BLOCK"
      severity = "LOW"
      application = {
        application = [
          {
            # id   = "slack"
            name = "Slack"
          },
        ]
      }
      tracking = {
        event = {
          enabled = true
        }
      }
    }
  }
}

# Kitchen sink — rule_type APPLICATION, every predicate/reference field referenced by
# name with the matching id commented above each.
resource "cato_application_control_rule" "kitchen_sink" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_application_control_section.example.section.id
  }
  rule = {
    name        = "Application Control Kitchen Sink"
    description = "Kitchen sink example covering every rule-schema field"
    enabled     = true
    rule_type   = "APPLICATION"
    application_rule = {
      action   = "BLOCK"
      severity = "MEDIUM"
      schedule = {
        active_on = "CUSTOM_TIMEFRAME"
        custom_timeframe = {
          from = "2026-01-01T00:00:00Z"
          to   = "2026-12-31T23:59:59Z"
        }
        # active_on = "CUSTOM_RECURRING"
        # custom_recurring = {
        #   from = "08:00"
        #   to   = "18:00"
        #   days = ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
        # }
      }
      source = {
        country = [
          {
            # id   = "FR"
            name = "France"
          },
        ]
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
        ]
        site = [
          {
            # id   = "144905"
            name = "my-site"
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
        system_group = [
          {
            # id   = "7S"
            name = "All Floating Ranges"
          },
        ]
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
      application = {
        app_category = [
          {
            # id   = "instant_messaging"
            name = "Instant Messaging"
          },
        ]
        application = [
          {
            # id   = "slack"
            name = "Slack"
          },
        ]
        custom_app = [
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
        domain = [
          "something.com",
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
          "9.9.9.9",
        ]
        ip_range = [
          {
            from = "1.2.3.1"
            to   = "1.2.3.4"
          },
        ]
        sanctioned_apps_category = [
          {
            # id   = "cloud_storage"
            name = "Cloud Storage"
          },
        ]
        subnet = [
          "1.2.3.0/24",
        ]
      }
      application_activity = [
        {
          activity = {
            # id   = "content_transfer_action_upload"
            name = "Content Transfer - Upload"
          }
        },
      ]
      application_activity_satisfy = "ANY" # ALL | ANY
      device = [
        {
          # id   = "4202"
          name = "Test Device Posture Profile"
        },
      ]
      access_method = [
        {
          access_method = "USER_AGENT" # USER_AGENT | CLIENT | CLIENTLESS | OTHER
          operator      = "CONTAINS"   # IS | CONTAINS | NOT_CONTAINS | IS_NOT
          value         = "Mozilla"
        },
      ]
      action_config = {
        user_notification = [
          {
            # id   = "12345"
            name = "Default Notification"
          },
        ]
      }
      file_attribute = [
        {
          file_attribute = "CONTENT_SIZE" # CONTENT_SIZE | CONTENT_TYPE | FILE_EXTENSION | CONTENT_IS_ENCRYPTED (DATA only)
          operator       = "GREATER_THAN" # GREATER_THAN | LESS_THAN | IS | IS_NOT | CONTAINS | NOT_CONTAINS
          value          = "104857600"    # 100 MB in bytes
        },
      ]
      file_attribute_satisfy = "ALL" # ALL | ANY
      tracking = {
        event = {
          enabled = true
        }
        alert = {
          enabled   = false
          frequency = "HOURLY" # HOURLY | DAILY | WEEKLY | IMMEDIATE
          mailing_list = [
            {
              # id   = "-100"
              name = "All Admins"
            },
          ]
          subscription_group = [
            {
              # id   = "12345"
              name = "SOC Alerts"
            },
          ]
          webhook = [
            {
              # id   = "12345"
              name = "PagerDuty Hook"
            },
          ]
        }
      }
    }
  }
}

# Kitchen sink — rule_type FILE specific leaves. FILE rules accept application = {} to
# match all cloud apps (the only rule_type that allows an empty application matcher) and
# require at least one application_activity entry.
resource "cato_application_control_rule" "kitchen_sink_file" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_application_control_section.example.section.id
  }
  rule = {
    name        = "Application Control Kitchen Sink File"
    description = "FILE rule_type example — allow large file uploads during working hours"
    enabled     = true
    rule_type   = "FILE"
    file_rule = {
      action   = "ALLOW"
      severity = "MEDIUM"
      schedule = {
        active_on = "WORKING_HOURS"
      }
      source = {
        ip = [
          "198.51.100.30",
        ]
        subnet = [
          "10.0.0.0/8",
        ]
      }
      application = {}
      application_activity = [
        {
          activity = {
            # id   = "content_transfer_action_upload"
            name = "Content Transfer - Upload"
          }
        },
        {
          activity = {
            # id   = "content_transfer_action_download"
            name = "Content Transfer - Download"
          }
        },
      ]
      application_activity_satisfy = "ANY"
      device = [
        {
          # id   = "4202"
          name = "Test Device Posture Profile"
        },
      ]
      access_method = [
        {
          access_method = "USER_AGENT"
          operator      = "IS"
          value         = "curl/8.0"
        },
      ]
      file_attribute = [
        {
          file_attribute = "FILE_EXTENSION"
          operator       = "IS"
          value          = "zip"
        },
      ]
      file_attribute_satisfy = "ANY"
      tracking = {
        event = {
          enabled = true
        }
        alert = {
          enabled   = false
          frequency = "IMMEDIATE"
        }
      }
    }
  }
}

# Kitchen sink — rule_type DATA specific leaves (dlp_profile, MONITOR action,
# CONTENT_IS_ENCRYPTED file_attribute).
resource "cato_application_control_rule" "kitchen_sink_data" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_application_control_section.example.section.id
  }
  rule = {
    name        = "Application Control Kitchen Sink Data"
    description = "DATA (DLP) rule_type example — monitor sensitive data uploads"
    enabled     = true
    rule_type   = "DATA"
    data_rule = {
      action   = "MONITOR" # ALLOW | BLOCK | MONITOR
      severity = "HIGH"
      schedule = {
        active_on = "ALWAYS"
      }
      source = {
        ip = [
          "198.51.100.20",
        ]
      }
      application = {
        application = [
          {
            # id   = "slack"
            name = "Slack"
          },
        ]
      }
      file_attribute = [
        {
          file_attribute = "CONTENT_TYPE"
          operator       = "CONTAINS"
          value          = "application/pdf"
        },
        {
          file_attribute = "CONTENT_IS_ENCRYPTED" # DATA rules only
          operator       = "IS"
          value          = "true"
        },
      ]
      file_attribute_satisfy = "ANY"
      tracking = {
        event = {
          enabled = true
        }
        alert = {
          enabled = false
        }
      }
      dlp_profile = {
        content_profile = [
          {
            # id   = "12345"
            name = "PCI Content Profile"
          },
        ]
        edm_profile = [
          {
            # id   = "12345"
            name = "Customer PII"
          },
        ]
      }
    }
  }
}
