# Application Control rule — rule_type APPLICATION (most common).
# Monitors or blocks access to specific cloud applications.
#
# NOTE: For APPLICATION and DATA rules the `application` block must have exactly
# one non-empty matcher (e.g. application, app_category, custom_app …).
# An empty `application = {}` is only valid for FILE rules.
resource "cato_application_control_rule" "application" {
  # position values: LAST_IN_POLICY | LAST_IN_SECTION | AFTER_RULE | BEFORE_RULE
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_application_control_section.example.section.id
  }

  rule = {
    name        = "TF — Block Slack uploads"
    description = "Managed by Terraform"
    enabled     = true
    rule_type   = "APPLICATION"

    application_rule = {
      action   = "MONITOR" # ALLOW | BLOCK | MONITOR
      severity = "MEDIUM"  # LOW | MEDIUM | HIGH

      # schedule: ALWAYS | WORKING_HOURS | CUSTOM_TIMEFRAME | CUSTOM_RECURRING
      # Omit schedule to default to ALWAYS.
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
        ip       = ["198.51.100.10"]
        subnet   = ["203.0.113.0/24"]
        ip_range = [{ from = "10.0.0.1", to = "10.0.0.20" }]
        # country             = [{ name = "France" }]
        # host                = [{ name = "my-server" }]
        # site                = [{ id = "site-uuid" }]
        # global_ip_range     = [{ name = "Partner Ranges" }]
        # network_interface   = [{ id = "nic-uuid" }]
        # site_network_subnet = [{ name = "Branch LAN" }]
        # floating_subnet     = [{ name = "Guest WiFi" }]
        # user                = [{ name = "alice@example.com" }]
        # users_group         = [{ name = "All Users" }]
        # group               = [{ name = "Engineering" }]
        # system_group        = [{ name = "Remote Users" }]
      }

      application = {
        application = [{ id = "slack" }]
        # app_category             = [{ name = "Instant Messaging" }]
        # custom_app               = [{ name = "My Internal App" }]
        # custom_category          = [{ name = "My Category" }]
        # sanctioned_apps_category = [{ name = "Cloud Storage" }]
        # domain                   = ["example.com"]
        # fqdn                     = ["app.example.com"]
        # ip                       = ["203.0.113.5"]
        # subnet                   = ["203.0.113.0/24"]
        # ip_range                 = [{ from = "203.0.113.10", to = "203.0.113.20" }]
        # global_ip_range          = [{ name = "Partner IPs" }]
      }

      tracking = {
        event = { enabled = true }
        alert = {
          enabled   = false
          frequency = "HOURLY" # HOURLY | DAILY | WEEKLY | IMMEDIATE
          # When enabled = true, at least one of the following is required:
          # subscription_group = [{ name = "SOC Alerts" }]
          # webhook            = [{ name = "PagerDuty Hook" }]
          # mailing_list       = [{ name = "security@example.com" }]
        }
      }

      # device = [{ name = "Corporate Laptops" }]

      access_method = [
        {
          access_method = "USER_AGENT" # USER_AGENT | CLIENT | CLIENTLESS | OTHER
          operator      = "CONTAINS"   # IS | CONTAINS | NOT_CONTAINS | IS_NOT
          value         = "Mozilla"
        },
      ]

      # action_config = {
      #   user_notification = [{ name = "Default Notification" }]
      # }
    }
  }
}

# Application Control rule — rule_type FILE.
# Monitors file transfers based on activity and file attributes.
# FILE rules require at least one application_activity entry.
# action = MONITOR is required when using CONTENT_SIZE attribute.
resource "cato_application_control_rule" "file" {
  at = {
    position = "AFTER_RULE"
    ref      = cato_application_control_rule.application.rule.id
  }

  rule = {
    name        = "TF — Monitor large file uploads"
    description = "Flag uploads over 100 MB during working hours"
    enabled     = true
    rule_type   = "FILE"

    file_rule = {
      action   = "MONITOR" # ALLOW | BLOCK | MONITOR — BLOCK requires CONTENT_TYPE attribute
      severity = "MEDIUM"

      schedule = {
        active_on = "WORKING_HOURS"
        # active_on = "CUSTOM_RECURRING"
        # custom_recurring = {
        #   from = "08:00"
        #   to   = "18:00"
        #   days = ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
        # }
      }

      source = {
        ip = ["198.51.100.30"]
        # subnet  = ["10.0.0.0/8"]
        # country = [{ name = "France" }]
      }

      # FILE rules accept application = {} to match all cloud apps.
      # For APPLICATION and DATA rules a non-empty matcher is required.
      application = {}

      # content_transfer_action_upload / content_transfer_action_download are generic
      # upload/download activities valid for FILE rules across all cloud applications.
      application_activity = [
        { activity = { id = "content_transfer_action_upload" } }
        # { activity = { id = "content_transfer_action_download" } }
      ]
      application_activity_satisfy = "ANY" # ALL | ANY

      file_attribute = [
        {
          file_attribute = "CONTENT_SIZE" # CONTENT_SIZE | CONTENT_TYPE | FILE_EXTENSION | CONTENT_IS_ENCRYPTED (DATA only)
          operator       = "GREATER_THAN" # GREATER_THAN | LESS_THAN | IS | IS_NOT | CONTAINS | NOT_CONTAINS
          value          = "104857600"    # 100 MB in bytes
        }
      ]
      file_attribute_satisfy = "ALL" # ALL | ANY

      tracking = {
        event = { enabled = true }
        alert = {
          enabled   = false
          frequency = "DAILY"
          # subscription_group = [{ name = "SOC Alerts" }]
        }
      }
    }
  }
}

# Application Control rule — rule_type DATA (DLP).
# Requires a dlp_profile with at least one content_profile or edm_profile configured
# in your Cato account. Uncomment and configure before applying.
#
# resource "cato_application_control_rule" "data" {
#   at = {
#     position = "LAST_IN_POLICY"
#   }
#
#   rule = {
#     name      = "TF — Block sensitive data uploads"
#     enabled   = true
#     rule_type = "DATA"
#
#     data_rule = {
#       action   = "BLOCK"  # ALLOW | BLOCK | MONITOR
#       severity = "HIGH"
#
#       schedule = {
#         active_on = "ALWAYS"
#       }
#
#       source = {
#         ip = ["198.51.100.20"]
#       }
#
#       application = {
#         application = [{ id = "slack" }]
#       }
#
#       file_attribute = [
#         {
#           file_attribute = "CONTENT_TYPE"
#           operator       = "CONTAINS"
#           value          = "application/pdf"
#         },
#         {
#           file_attribute = "CONTENT_SIZE"
#           operator       = "GREATER_THAN"
#           value          = "1048576"  # 1 MB in bytes
#         },
#       ]
#       file_attribute_satisfy = "ANY"
#
#       tracking = {
#         event = { enabled = true }
#         alert = { enabled = false }
#       }
#
#       dlp_profile = {
#         content_profile = [{ name = "PCI Content Profile" }]
#         edm_profile     = [{ name = "Customer PII" }]
#       }
#     }
#   }
# }
