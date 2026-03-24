locals {
  // list of private access policy rules, presumably read from an external JSON file
  json_file_contents_example = <<EOF
    [
      {
        "action": "BLOCK",
        "active_period": {
          "effective_from": "2026-01-20T03:04:00Z",
          "use_effective_from": true,
          "expires_at": "2027-02-20T03:04:00Z",
          "use_expires_at": true
        },
        "applications": [
          { "name": "private-app-1" }
        ],
        "connection_origins": [ "REMOTE", "SITE" ],
        "countries": [
          { "name": "United States" }
        ],
        "description": "rule1 description",
        "devices": [
          { "name": "Test device posture 1" }
        ],
        "enabled": true,
        "name": "rule1",
        "platforms": [ "LINUX", "WINDOWS" ],
        "schedule": {
          "active_on": "WORKING_HOURS",
          "custom_timeframe": {
            "from": "2026-02-20T01:02:00Z",
            "to": "2026-02-20T03:04:00Z"
          },
          "custom_recurring": {
            "days": [ "MONDAY", "FRIDAY" ],
            "from": "08:04:00",
            "to": "19:30:00"
          }
        },
        "source": {
          "users": [
            { "name": "test user_1" },
            { "name": "test user_2" }
          ],
          "user_groups": [
            { "name": "group-1" }
          ]
        },
        "tracking": {
          "event": {
            "enabled": true
          },
          "alert": {
            "enabled": true,
            "frequency": "DAILY",
            "subscription_group": [
              { "name": "subscription_group" }
            ],
            "webhook": [
              { "name": "webhook_1" }
            ],
            "mailing_list": [
              { "name": "mailing_list" }
            ]
          }
        },
        "user_attributes": {
          "risk_score": {
            "category": "LOW",
            "operator": "LTE"
          }
        }
      }
    ]
  EOF

  # Use example JSON data declared above, or read it from a file
  rule_data_list = jsondecode(local.json_file_contents_example)
  # rule_data_list = jsondecode(file("${path.module}/private-access-rules.json"))

  # Convert rules list to map keyed by rule_name for provider schema compatibility
  rule_data = {
    for i, rule in local.rule_data_list :
    rule.name => merge(rule, { index = i })
  }
}

# Enable or disable private access policy
resource "cato_private_access_policy" "this" {
  enabled = true
}

# Manage all the rules from the json
# (use cato_private_access_rule_bulk to publish the changes)
resource "cato_private_access_rule" "all" {
  for_each = local.rule_data

  action             = each.value.action
  active_period      = each.value.active_period
  applications       = each.value.applications
  connection_origins = each.value.connection_origins
  countries          = each.value.countries
  description        = each.value.description
  devices            = each.value.devices
  enabled            = each.value.enabled
  name               = each.key
  platforms          = each.value.platforms
  schedule           = each.value.schedule
  source             = each.value.source
  tracking           = each.value.tracking
  user_attributes    = each.value.user_attributes
}

# Handle ordering and publishing the rules
resource "cato_private_access_rule_bulk" "this" {
  depends_on = [cato_private_access_rule.all, cato_private_access_policy.this]
  rule_data  = local.rule_data
}

