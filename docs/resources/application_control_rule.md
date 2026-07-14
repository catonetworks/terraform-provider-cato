---
page_title: "cato_application_control_rule Resource - terraform-provider-cato"
subcategory: ""
description: |-
  Manages a rule in the Cato Application Control (App & Data Inline Protection) policy. Underlying GraphQL is marked @beta; behavior and fields may change.
---

# cato_application_control_rule (Resource)

Manages a rule in the Cato Application Control (App & Data Inline Protection) policy. Underlying GraphQL is marked @beta; behavior and fields may change.

## Example Usage

```terraform
# Application Control rule — rule_type APPLICATION (most common).
# Allows or blocks access to specific cloud applications.
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
      action   = "BLOCK" # ALLOW | BLOCK
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
```

## Import

Import uses the Cato **rule id** as the import `id` (written into `rule.id` in state).

```shell
terraform import cato_application_control_rule.example <rule_id>
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `at` (Attributes) Where to insert the rule (see [below for nested schema](#nestedatt--at))
- `rule` (Attributes) Rule definition (see [below for nested schema](#nestedatt--rule))

### Read-Only

- `id` (String) Rule ID

<a id="nestedatt--at"></a>
### Nested Schema for `at`

Required:

- `position` (String) Position relative to policy, section, or another rule

Optional:

- `ref` (String) Reference rule or section ID when required by position


<a id="nestedatt--rule"></a>
### Nested Schema for `rule`

Required:

- `enabled` (Boolean) Whether the rule is enabled
- `name` (String) Rule name
- `rule_type` (String) Which nested rule block is active

Optional:

- `application_rule` (Attributes) Settings when rule_type is APPLICATION (see [below for nested schema](#nestedatt--rule--application_rule))
- `data_rule` (Attributes) Settings when rule_type is DATA (see [below for nested schema](#nestedatt--rule--data_rule))
- `description` (String) Rule description
- `file_rule` (Attributes) Settings when rule_type is FILE (see [below for nested schema](#nestedatt--rule--file_rule))

Read-Only:

- `id` (String) Rule ID

<a id="nestedatt--rule--application_rule"></a>
### Nested Schema for `rule.application_rule`

Required:

- `action` (String) Rule action
- `severity` (String) Rule severity

Optional:

- `access_method` (Attributes List) Access method rows (see [below for nested schema](#nestedatt--rule--application_rule--access_method))
- `action_config` (Attributes) Action configuration (e.g. user notifications) (see [below for nested schema](#nestedatt--rule--application_rule--action_config))
- `application` (Attributes) Application matching (WAN-shaped object) (see [below for nested schema](#nestedatt--rule--application_rule--application))
- `application_activity` (Attributes List) Application activities to match (required for file/data rules when an application is specified) (see [below for nested schema](#nestedatt--rule--application_rule--application_activity))
- `application_activity_satisfy` (String) How application activities are combined (ALL | ANY)
- `device` (Attributes Set) Device profiles (see [below for nested schema](#nestedatt--rule--application_rule--device))
- `dlp_profile` (Attributes) DLP profile (data rules) (see [below for nested schema](#nestedatt--rule--application_rule--dlp_profile))
- `file_attribute` (Attributes List) File attribute rows (data/file rules) (see [below for nested schema](#nestedatt--rule--application_rule--file_attribute))
- `file_attribute_satisfy` (String) How file attributes are combined
- `schedule` (Attributes) Schedule for the typed rule (see [below for nested schema](#nestedatt--rule--application_rule--schedule))
- `source` (Attributes) Source traffic matching criteria (see [below for nested schema](#nestedatt--rule--application_rule--source))
- `tracking` (Attributes) Tracking configuration (see [below for nested schema](#nestedatt--rule--application_rule--tracking))

<a id="nestedatt--rule--application_rule--access_method"></a>
### Nested Schema for `rule.application_rule.access_method`

Required:

- `access_method` (String)
- `operator` (String)

Optional:

- `value` (String)


<a id="nestedatt--rule--application_rule--action_config"></a>
### Nested Schema for `rule.application_rule.action_config`

Optional:

- `user_notification` (Attributes Set) (see [below for nested schema](#nestedatt--rule--application_rule--action_config--user_notification))

<a id="nestedatt--rule--application_rule--action_config--user_notification"></a>
### Nested Schema for `rule.application_rule.action_config.user_notification`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--application_rule--application"></a>
### Nested Schema for `rule.application_rule.application`

Optional:

- `app_category` (Attributes Set) Application categories (see [below for nested schema](#nestedatt--rule--application_rule--application--app_category))
- `application` (Attributes Set) Predefined applications (logical OR within the set) (see [below for nested schema](#nestedatt--rule--application_rule--application--application))
- `custom_app` (Attributes Set) Custom applications (see [below for nested schema](#nestedatt--rule--application_rule--application--custom_app))
- `custom_category` (Attributes Set) Custom categories (see [below for nested schema](#nestedatt--rule--application_rule--application--custom_category))
- `domain` (List of String) Second-level domains to match
- `fqdn` (List of String) FQDNs to match
- `global_ip_range` (Attributes Set) Global IP range objects (see [below for nested schema](#nestedatt--rule--application_rule--application--global_ip_range))
- `ip` (List of String) IPv4 addresses
- `ip_range` (Attributes List) Inclusive IP ranges (see [below for nested schema](#nestedatt--rule--application_rule--application--ip_range))
- `sanctioned_apps_category` (Attributes Set) Sanctioned cloud application categories (see [below for nested schema](#nestedatt--rule--application_rule--application--sanctioned_apps_category))
- `subnet` (List of String) Subnets in CIDR notation

<a id="nestedatt--rule--application_rule--application--app_category"></a>
### Nested Schema for `rule.application_rule.application.app_category`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--application_rule--application--application"></a>
### Nested Schema for `rule.application_rule.application.application`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--application_rule--application--custom_app"></a>
### Nested Schema for `rule.application_rule.application.custom_app`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--application_rule--application--custom_category"></a>
### Nested Schema for `rule.application_rule.application.custom_category`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--application_rule--application--global_ip_range"></a>
### Nested Schema for `rule.application_rule.application.global_ip_range`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--application_rule--application--ip_range"></a>
### Nested Schema for `rule.application_rule.application.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--application_rule--application--sanctioned_apps_category"></a>
### Nested Schema for `rule.application_rule.application.sanctioned_apps_category`

Optional:

- `id` (String) Object ID
- `name` (String) Object name



<a id="nestedatt--rule--application_rule--application_activity"></a>
### Nested Schema for `rule.application_rule.application_activity`

Required:

- `activity` (Attributes) (see [below for nested schema](#nestedatt--rule--application_rule--application_activity--activity))

<a id="nestedatt--rule--application_rule--application_activity--activity"></a>
### Nested Schema for `rule.application_rule.application_activity.activity`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--application_rule--device"></a>
### Nested Schema for `rule.application_rule.device`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--dlp_profile"></a>
### Nested Schema for `rule.application_rule.dlp_profile`

Optional:

- `content_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--application_rule--dlp_profile--content_profile))
- `edm_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--application_rule--dlp_profile--edm_profile))

<a id="nestedatt--rule--application_rule--dlp_profile--content_profile"></a>
### Nested Schema for `rule.application_rule.dlp_profile.content_profile`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--dlp_profile--edm_profile"></a>
### Nested Schema for `rule.application_rule.dlp_profile.edm_profile`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--application_rule--file_attribute"></a>
### Nested Schema for `rule.application_rule.file_attribute`

Required:

- `file_attribute` (String)
- `operator` (String)

Optional:

- `value` (String)


<a id="nestedatt--rule--application_rule--schedule"></a>
### Nested Schema for `rule.application_rule.schedule`

Optional:

- `active_on` (String)
- `custom_recurring` (Attributes) (see [below for nested schema](#nestedatt--rule--application_rule--schedule--custom_recurring))
- `custom_timeframe` (Attributes) (see [below for nested schema](#nestedatt--rule--application_rule--schedule--custom_timeframe))

<a id="nestedatt--rule--application_rule--schedule--custom_recurring"></a>
### Nested Schema for `rule.application_rule.schedule.custom_recurring`

Optional:

- `days` (List of String)
- `from` (String)
- `to` (String)


<a id="nestedatt--rule--application_rule--schedule--custom_timeframe"></a>
### Nested Schema for `rule.application_rule.schedule.custom_timeframe`

Optional:

- `from` (String)
- `to` (String)



<a id="nestedatt--rule--application_rule--source"></a>
### Nested Schema for `rule.application_rule.source`

Optional:

- `country` (Attributes Set) Source country matching criteria (see [below for nested schema](#nestedatt--rule--application_rule--source--country))
- `floating_subnet` (Attributes Set) Floating subnets (see [below for nested schema](#nestedatt--rule--application_rule--source--floating_subnet))
- `global_ip_range` (Attributes Set) Globally defined IP range, IP and subnet objects (see [below for nested schema](#nestedatt--rule--application_rule--source--global_ip_range))
- `group` (Attributes Set) Groups defined for your account (see [below for nested schema](#nestedatt--rule--application_rule--source--group))
- `host` (Attributes Set) Hosts and servers defined for your account (see [below for nested schema](#nestedatt--rule--application_rule--source--host))
- `ip` (List of String) IPv4 address list
- `ip_range` (Attributes List) Multiple separate IP addresses or an IP range (see [below for nested schema](#nestedatt--rule--application_rule--source--ip_range))
- `network_interface` (Attributes Set) Network interface defined for a site (see [below for nested schema](#nestedatt--rule--application_rule--source--network_interface))
- `site` (Attributes Set) Site defined for the account (see [below for nested schema](#nestedatt--rule--application_rule--source--site))
- `site_network_subnet` (Attributes Set) Site network subnet (see [below for nested schema](#nestedatt--rule--application_rule--source--site_network_subnet))
- `subnet` (List of String) Subnets and network ranges defined for the LAN interfaces of a site
- `system_group` (Attributes Set) Predefined Cato groups (see [below for nested schema](#nestedatt--rule--application_rule--source--system_group))
- `user` (Attributes Set) Individual users defined for the account (see [below for nested schema](#nestedatt--rule--application_rule--source--user))
- `users_group` (Attributes Set) Group of users (see [below for nested schema](#nestedatt--rule--application_rule--source--users_group))

<a id="nestedatt--rule--application_rule--source--country"></a>
### Nested Schema for `rule.application_rule.source.country`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--floating_subnet"></a>
### Nested Schema for `rule.application_rule.source.floating_subnet`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--global_ip_range"></a>
### Nested Schema for `rule.application_rule.source.global_ip_range`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--group"></a>
### Nested Schema for `rule.application_rule.source.group`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--host"></a>
### Nested Schema for `rule.application_rule.source.host`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--ip_range"></a>
### Nested Schema for `rule.application_rule.source.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--application_rule--source--network_interface"></a>
### Nested Schema for `rule.application_rule.source.network_interface`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--site"></a>
### Nested Schema for `rule.application_rule.source.site`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--site_network_subnet"></a>
### Nested Schema for `rule.application_rule.source.site_network_subnet`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--system_group"></a>
### Nested Schema for `rule.application_rule.source.system_group`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--user"></a>
### Nested Schema for `rule.application_rule.source.user`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--source--users_group"></a>
### Nested Schema for `rule.application_rule.source.users_group`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--application_rule--tracking"></a>
### Nested Schema for `rule.application_rule.tracking`

Optional:

- `alert` (Attributes) (see [below for nested schema](#nestedatt--rule--application_rule--tracking--alert))
- `event` (Attributes) (see [below for nested schema](#nestedatt--rule--application_rule--tracking--event))

<a id="nestedatt--rule--application_rule--tracking--alert"></a>
### Nested Schema for `rule.application_rule.tracking.alert`

Optional:

- `enabled` (Boolean)
- `frequency` (String)
- `mailing_list` (Attributes Set) (see [below for nested schema](#nestedatt--rule--application_rule--tracking--alert--mailing_list))
- `subscription_group` (Attributes Set) (see [below for nested schema](#nestedatt--rule--application_rule--tracking--alert--subscription_group))
- `webhook` (Attributes Set) (see [below for nested schema](#nestedatt--rule--application_rule--tracking--alert--webhook))

<a id="nestedatt--rule--application_rule--tracking--alert--mailing_list"></a>
### Nested Schema for `rule.application_rule.tracking.alert.mailing_list`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--tracking--alert--subscription_group"></a>
### Nested Schema for `rule.application_rule.tracking.alert.subscription_group`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--tracking--alert--webhook"></a>
### Nested Schema for `rule.application_rule.tracking.alert.webhook`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--application_rule--tracking--event"></a>
### Nested Schema for `rule.application_rule.tracking.event`

Optional:

- `enabled` (Boolean)




<a id="nestedatt--rule--data_rule"></a>
### Nested Schema for `rule.data_rule`

Required:

- `action` (String) Rule action
- `severity` (String) Rule severity

Optional:

- `access_method` (Attributes List) Access method rows (see [below for nested schema](#nestedatt--rule--data_rule--access_method))
- `action_config` (Attributes) Action configuration (e.g. user notifications) (see [below for nested schema](#nestedatt--rule--data_rule--action_config))
- `application` (Attributes) Application matching (WAN-shaped object) (see [below for nested schema](#nestedatt--rule--data_rule--application))
- `application_activity` (Attributes List) Application activities to match (required for file/data rules when an application is specified) (see [below for nested schema](#nestedatt--rule--data_rule--application_activity))
- `application_activity_satisfy` (String) How application activities are combined (ALL | ANY)
- `device` (Attributes Set) Device profiles (see [below for nested schema](#nestedatt--rule--data_rule--device))
- `dlp_profile` (Attributes) DLP profile (data rules) (see [below for nested schema](#nestedatt--rule--data_rule--dlp_profile))
- `file_attribute` (Attributes List) File attribute rows (data/file rules) (see [below for nested schema](#nestedatt--rule--data_rule--file_attribute))
- `file_attribute_satisfy` (String) How file attributes are combined
- `schedule` (Attributes) Schedule for the typed rule (see [below for nested schema](#nestedatt--rule--data_rule--schedule))
- `source` (Attributes) Source traffic matching criteria (see [below for nested schema](#nestedatt--rule--data_rule--source))
- `tracking` (Attributes) Tracking configuration (see [below for nested schema](#nestedatt--rule--data_rule--tracking))

<a id="nestedatt--rule--data_rule--access_method"></a>
### Nested Schema for `rule.data_rule.access_method`

Required:

- `access_method` (String)
- `operator` (String)

Optional:

- `value` (String)


<a id="nestedatt--rule--data_rule--action_config"></a>
### Nested Schema for `rule.data_rule.action_config`

Optional:

- `user_notification` (Attributes Set) (see [below for nested schema](#nestedatt--rule--data_rule--action_config--user_notification))

<a id="nestedatt--rule--data_rule--action_config--user_notification"></a>
### Nested Schema for `rule.data_rule.action_config.user_notification`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--data_rule--application"></a>
### Nested Schema for `rule.data_rule.application`

Optional:

- `app_category` (Attributes Set) Application categories (see [below for nested schema](#nestedatt--rule--data_rule--application--app_category))
- `application` (Attributes Set) Predefined applications (logical OR within the set) (see [below for nested schema](#nestedatt--rule--data_rule--application--application))
- `custom_app` (Attributes Set) Custom applications (see [below for nested schema](#nestedatt--rule--data_rule--application--custom_app))
- `custom_category` (Attributes Set) Custom categories (see [below for nested schema](#nestedatt--rule--data_rule--application--custom_category))
- `domain` (List of String) Second-level domains to match
- `fqdn` (List of String) FQDNs to match
- `global_ip_range` (Attributes Set) Global IP range objects (see [below for nested schema](#nestedatt--rule--data_rule--application--global_ip_range))
- `ip` (List of String) IPv4 addresses
- `ip_range` (Attributes List) Inclusive IP ranges (see [below for nested schema](#nestedatt--rule--data_rule--application--ip_range))
- `sanctioned_apps_category` (Attributes Set) Sanctioned cloud application categories (see [below for nested schema](#nestedatt--rule--data_rule--application--sanctioned_apps_category))
- `subnet` (List of String) Subnets in CIDR notation

<a id="nestedatt--rule--data_rule--application--app_category"></a>
### Nested Schema for `rule.data_rule.application.app_category`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--data_rule--application--application"></a>
### Nested Schema for `rule.data_rule.application.application`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--data_rule--application--custom_app"></a>
### Nested Schema for `rule.data_rule.application.custom_app`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--data_rule--application--custom_category"></a>
### Nested Schema for `rule.data_rule.application.custom_category`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--data_rule--application--global_ip_range"></a>
### Nested Schema for `rule.data_rule.application.global_ip_range`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--data_rule--application--ip_range"></a>
### Nested Schema for `rule.data_rule.application.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--data_rule--application--sanctioned_apps_category"></a>
### Nested Schema for `rule.data_rule.application.sanctioned_apps_category`

Optional:

- `id` (String) Object ID
- `name` (String) Object name



<a id="nestedatt--rule--data_rule--application_activity"></a>
### Nested Schema for `rule.data_rule.application_activity`

Required:

- `activity` (Attributes) (see [below for nested schema](#nestedatt--rule--data_rule--application_activity--activity))

<a id="nestedatt--rule--data_rule--application_activity--activity"></a>
### Nested Schema for `rule.data_rule.application_activity.activity`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--data_rule--device"></a>
### Nested Schema for `rule.data_rule.device`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--dlp_profile"></a>
### Nested Schema for `rule.data_rule.dlp_profile`

Optional:

- `content_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--data_rule--dlp_profile--content_profile))
- `edm_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--data_rule--dlp_profile--edm_profile))

<a id="nestedatt--rule--data_rule--dlp_profile--content_profile"></a>
### Nested Schema for `rule.data_rule.dlp_profile.content_profile`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--dlp_profile--edm_profile"></a>
### Nested Schema for `rule.data_rule.dlp_profile.edm_profile`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--data_rule--file_attribute"></a>
### Nested Schema for `rule.data_rule.file_attribute`

Required:

- `file_attribute` (String)
- `operator` (String)

Optional:

- `value` (String)


<a id="nestedatt--rule--data_rule--schedule"></a>
### Nested Schema for `rule.data_rule.schedule`

Optional:

- `active_on` (String)
- `custom_recurring` (Attributes) (see [below for nested schema](#nestedatt--rule--data_rule--schedule--custom_recurring))
- `custom_timeframe` (Attributes) (see [below for nested schema](#nestedatt--rule--data_rule--schedule--custom_timeframe))

<a id="nestedatt--rule--data_rule--schedule--custom_recurring"></a>
### Nested Schema for `rule.data_rule.schedule.custom_recurring`

Optional:

- `days` (List of String)
- `from` (String)
- `to` (String)


<a id="nestedatt--rule--data_rule--schedule--custom_timeframe"></a>
### Nested Schema for `rule.data_rule.schedule.custom_timeframe`

Optional:

- `from` (String)
- `to` (String)



<a id="nestedatt--rule--data_rule--source"></a>
### Nested Schema for `rule.data_rule.source`

Optional:

- `country` (Attributes Set) Source country matching criteria (see [below for nested schema](#nestedatt--rule--data_rule--source--country))
- `floating_subnet` (Attributes Set) Floating subnets (see [below for nested schema](#nestedatt--rule--data_rule--source--floating_subnet))
- `global_ip_range` (Attributes Set) Globally defined IP range, IP and subnet objects (see [below for nested schema](#nestedatt--rule--data_rule--source--global_ip_range))
- `group` (Attributes Set) Groups defined for your account (see [below for nested schema](#nestedatt--rule--data_rule--source--group))
- `host` (Attributes Set) Hosts and servers defined for your account (see [below for nested schema](#nestedatt--rule--data_rule--source--host))
- `ip` (List of String) IPv4 address list
- `ip_range` (Attributes List) Multiple separate IP addresses or an IP range (see [below for nested schema](#nestedatt--rule--data_rule--source--ip_range))
- `network_interface` (Attributes Set) Network interface defined for a site (see [below for nested schema](#nestedatt--rule--data_rule--source--network_interface))
- `site` (Attributes Set) Site defined for the account (see [below for nested schema](#nestedatt--rule--data_rule--source--site))
- `site_network_subnet` (Attributes Set) Site network subnet (see [below for nested schema](#nestedatt--rule--data_rule--source--site_network_subnet))
- `subnet` (List of String) Subnets and network ranges defined for the LAN interfaces of a site
- `system_group` (Attributes Set) Predefined Cato groups (see [below for nested schema](#nestedatt--rule--data_rule--source--system_group))
- `user` (Attributes Set) Individual users defined for the account (see [below for nested schema](#nestedatt--rule--data_rule--source--user))
- `users_group` (Attributes Set) Group of users (see [below for nested schema](#nestedatt--rule--data_rule--source--users_group))

<a id="nestedatt--rule--data_rule--source--country"></a>
### Nested Schema for `rule.data_rule.source.country`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--floating_subnet"></a>
### Nested Schema for `rule.data_rule.source.floating_subnet`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--global_ip_range"></a>
### Nested Schema for `rule.data_rule.source.global_ip_range`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--group"></a>
### Nested Schema for `rule.data_rule.source.group`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--host"></a>
### Nested Schema for `rule.data_rule.source.host`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--ip_range"></a>
### Nested Schema for `rule.data_rule.source.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--data_rule--source--network_interface"></a>
### Nested Schema for `rule.data_rule.source.network_interface`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--site"></a>
### Nested Schema for `rule.data_rule.source.site`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--site_network_subnet"></a>
### Nested Schema for `rule.data_rule.source.site_network_subnet`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--system_group"></a>
### Nested Schema for `rule.data_rule.source.system_group`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--user"></a>
### Nested Schema for `rule.data_rule.source.user`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--source--users_group"></a>
### Nested Schema for `rule.data_rule.source.users_group`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--data_rule--tracking"></a>
### Nested Schema for `rule.data_rule.tracking`

Optional:

- `alert` (Attributes) (see [below for nested schema](#nestedatt--rule--data_rule--tracking--alert))
- `event` (Attributes) (see [below for nested schema](#nestedatt--rule--data_rule--tracking--event))

<a id="nestedatt--rule--data_rule--tracking--alert"></a>
### Nested Schema for `rule.data_rule.tracking.alert`

Optional:

- `enabled` (Boolean)
- `frequency` (String)
- `mailing_list` (Attributes Set) (see [below for nested schema](#nestedatt--rule--data_rule--tracking--alert--mailing_list))
- `subscription_group` (Attributes Set) (see [below for nested schema](#nestedatt--rule--data_rule--tracking--alert--subscription_group))
- `webhook` (Attributes Set) (see [below for nested schema](#nestedatt--rule--data_rule--tracking--alert--webhook))

<a id="nestedatt--rule--data_rule--tracking--alert--mailing_list"></a>
### Nested Schema for `rule.data_rule.tracking.alert.mailing_list`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--tracking--alert--subscription_group"></a>
### Nested Schema for `rule.data_rule.tracking.alert.subscription_group`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--tracking--alert--webhook"></a>
### Nested Schema for `rule.data_rule.tracking.alert.webhook`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--data_rule--tracking--event"></a>
### Nested Schema for `rule.data_rule.tracking.event`

Optional:

- `enabled` (Boolean)




<a id="nestedatt--rule--file_rule"></a>
### Nested Schema for `rule.file_rule`

Required:

- `action` (String) Rule action
- `severity` (String) Rule severity

Optional:

- `access_method` (Attributes List) Access method rows (see [below for nested schema](#nestedatt--rule--file_rule--access_method))
- `action_config` (Attributes) Action configuration (e.g. user notifications) (see [below for nested schema](#nestedatt--rule--file_rule--action_config))
- `application` (Attributes) Application matching (WAN-shaped object) (see [below for nested schema](#nestedatt--rule--file_rule--application))
- `application_activity` (Attributes List) Application activities to match (required for file/data rules when an application is specified) (see [below for nested schema](#nestedatt--rule--file_rule--application_activity))
- `application_activity_satisfy` (String) How application activities are combined (ALL | ANY)
- `device` (Attributes Set) Device profiles (see [below for nested schema](#nestedatt--rule--file_rule--device))
- `dlp_profile` (Attributes) DLP profile (data rules) (see [below for nested schema](#nestedatt--rule--file_rule--dlp_profile))
- `file_attribute` (Attributes List) File attribute rows (data/file rules) (see [below for nested schema](#nestedatt--rule--file_rule--file_attribute))
- `file_attribute_satisfy` (String) How file attributes are combined
- `schedule` (Attributes) Schedule for the typed rule (see [below for nested schema](#nestedatt--rule--file_rule--schedule))
- `source` (Attributes) Source traffic matching criteria (see [below for nested schema](#nestedatt--rule--file_rule--source))
- `tracking` (Attributes) Tracking configuration (see [below for nested schema](#nestedatt--rule--file_rule--tracking))

<a id="nestedatt--rule--file_rule--access_method"></a>
### Nested Schema for `rule.file_rule.access_method`

Required:

- `access_method` (String)
- `operator` (String)

Optional:

- `value` (String)


<a id="nestedatt--rule--file_rule--action_config"></a>
### Nested Schema for `rule.file_rule.action_config`

Optional:

- `user_notification` (Attributes Set) (see [below for nested schema](#nestedatt--rule--file_rule--action_config--user_notification))

<a id="nestedatt--rule--file_rule--action_config--user_notification"></a>
### Nested Schema for `rule.file_rule.action_config.user_notification`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--file_rule--application"></a>
### Nested Schema for `rule.file_rule.application`

Optional:

- `app_category` (Attributes Set) Application categories (see [below for nested schema](#nestedatt--rule--file_rule--application--app_category))
- `application` (Attributes Set) Predefined applications (logical OR within the set) (see [below for nested schema](#nestedatt--rule--file_rule--application--application))
- `custom_app` (Attributes Set) Custom applications (see [below for nested schema](#nestedatt--rule--file_rule--application--custom_app))
- `custom_category` (Attributes Set) Custom categories (see [below for nested schema](#nestedatt--rule--file_rule--application--custom_category))
- `domain` (List of String) Second-level domains to match
- `fqdn` (List of String) FQDNs to match
- `global_ip_range` (Attributes Set) Global IP range objects (see [below for nested schema](#nestedatt--rule--file_rule--application--global_ip_range))
- `ip` (List of String) IPv4 addresses
- `ip_range` (Attributes List) Inclusive IP ranges (see [below for nested schema](#nestedatt--rule--file_rule--application--ip_range))
- `sanctioned_apps_category` (Attributes Set) Sanctioned cloud application categories (see [below for nested schema](#nestedatt--rule--file_rule--application--sanctioned_apps_category))
- `subnet` (List of String) Subnets in CIDR notation

<a id="nestedatt--rule--file_rule--application--app_category"></a>
### Nested Schema for `rule.file_rule.application.app_category`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--file_rule--application--application"></a>
### Nested Schema for `rule.file_rule.application.application`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--file_rule--application--custom_app"></a>
### Nested Schema for `rule.file_rule.application.custom_app`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--file_rule--application--custom_category"></a>
### Nested Schema for `rule.file_rule.application.custom_category`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--file_rule--application--global_ip_range"></a>
### Nested Schema for `rule.file_rule.application.global_ip_range`

Optional:

- `id` (String) Object ID
- `name` (String) Object name


<a id="nestedatt--rule--file_rule--application--ip_range"></a>
### Nested Schema for `rule.file_rule.application.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--file_rule--application--sanctioned_apps_category"></a>
### Nested Schema for `rule.file_rule.application.sanctioned_apps_category`

Optional:

- `id` (String) Object ID
- `name` (String) Object name



<a id="nestedatt--rule--file_rule--application_activity"></a>
### Nested Schema for `rule.file_rule.application_activity`

Required:

- `activity` (Attributes) (see [below for nested schema](#nestedatt--rule--file_rule--application_activity--activity))

<a id="nestedatt--rule--file_rule--application_activity--activity"></a>
### Nested Schema for `rule.file_rule.application_activity.activity`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--file_rule--device"></a>
### Nested Schema for `rule.file_rule.device`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--dlp_profile"></a>
### Nested Schema for `rule.file_rule.dlp_profile`

Optional:

- `content_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--file_rule--dlp_profile--content_profile))
- `edm_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--file_rule--dlp_profile--edm_profile))

<a id="nestedatt--rule--file_rule--dlp_profile--content_profile"></a>
### Nested Schema for `rule.file_rule.dlp_profile.content_profile`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--dlp_profile--edm_profile"></a>
### Nested Schema for `rule.file_rule.dlp_profile.edm_profile`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--file_rule--file_attribute"></a>
### Nested Schema for `rule.file_rule.file_attribute`

Required:

- `file_attribute` (String)
- `operator` (String)

Optional:

- `value` (String)


<a id="nestedatt--rule--file_rule--schedule"></a>
### Nested Schema for `rule.file_rule.schedule`

Optional:

- `active_on` (String)
- `custom_recurring` (Attributes) (see [below for nested schema](#nestedatt--rule--file_rule--schedule--custom_recurring))
- `custom_timeframe` (Attributes) (see [below for nested schema](#nestedatt--rule--file_rule--schedule--custom_timeframe))

<a id="nestedatt--rule--file_rule--schedule--custom_recurring"></a>
### Nested Schema for `rule.file_rule.schedule.custom_recurring`

Optional:

- `days` (List of String)
- `from` (String)
- `to` (String)


<a id="nestedatt--rule--file_rule--schedule--custom_timeframe"></a>
### Nested Schema for `rule.file_rule.schedule.custom_timeframe`

Optional:

- `from` (String)
- `to` (String)



<a id="nestedatt--rule--file_rule--source"></a>
### Nested Schema for `rule.file_rule.source`

Optional:

- `country` (Attributes Set) Source country matching criteria (see [below for nested schema](#nestedatt--rule--file_rule--source--country))
- `floating_subnet` (Attributes Set) Floating subnets (see [below for nested schema](#nestedatt--rule--file_rule--source--floating_subnet))
- `global_ip_range` (Attributes Set) Globally defined IP range, IP and subnet objects (see [below for nested schema](#nestedatt--rule--file_rule--source--global_ip_range))
- `group` (Attributes Set) Groups defined for your account (see [below for nested schema](#nestedatt--rule--file_rule--source--group))
- `host` (Attributes Set) Hosts and servers defined for your account (see [below for nested schema](#nestedatt--rule--file_rule--source--host))
- `ip` (List of String) IPv4 address list
- `ip_range` (Attributes List) Multiple separate IP addresses or an IP range (see [below for nested schema](#nestedatt--rule--file_rule--source--ip_range))
- `network_interface` (Attributes Set) Network interface defined for a site (see [below for nested schema](#nestedatt--rule--file_rule--source--network_interface))
- `site` (Attributes Set) Site defined for the account (see [below for nested schema](#nestedatt--rule--file_rule--source--site))
- `site_network_subnet` (Attributes Set) Site network subnet (see [below for nested schema](#nestedatt--rule--file_rule--source--site_network_subnet))
- `subnet` (List of String) Subnets and network ranges defined for the LAN interfaces of a site
- `system_group` (Attributes Set) Predefined Cato groups (see [below for nested schema](#nestedatt--rule--file_rule--source--system_group))
- `user` (Attributes Set) Individual users defined for the account (see [below for nested schema](#nestedatt--rule--file_rule--source--user))
- `users_group` (Attributes Set) Group of users (see [below for nested schema](#nestedatt--rule--file_rule--source--users_group))

<a id="nestedatt--rule--file_rule--source--country"></a>
### Nested Schema for `rule.file_rule.source.country`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--floating_subnet"></a>
### Nested Schema for `rule.file_rule.source.floating_subnet`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--global_ip_range"></a>
### Nested Schema for `rule.file_rule.source.global_ip_range`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--group"></a>
### Nested Schema for `rule.file_rule.source.group`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--host"></a>
### Nested Schema for `rule.file_rule.source.host`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--ip_range"></a>
### Nested Schema for `rule.file_rule.source.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--file_rule--source--network_interface"></a>
### Nested Schema for `rule.file_rule.source.network_interface`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--site"></a>
### Nested Schema for `rule.file_rule.source.site`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--site_network_subnet"></a>
### Nested Schema for `rule.file_rule.source.site_network_subnet`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--system_group"></a>
### Nested Schema for `rule.file_rule.source.system_group`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--user"></a>
### Nested Schema for `rule.file_rule.source.user`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--source--users_group"></a>
### Nested Schema for `rule.file_rule.source.users_group`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--file_rule--tracking"></a>
### Nested Schema for `rule.file_rule.tracking`

Optional:

- `alert` (Attributes) (see [below for nested schema](#nestedatt--rule--file_rule--tracking--alert))
- `event` (Attributes) (see [below for nested schema](#nestedatt--rule--file_rule--tracking--event))

<a id="nestedatt--rule--file_rule--tracking--alert"></a>
### Nested Schema for `rule.file_rule.tracking.alert`

Optional:

- `enabled` (Boolean)
- `frequency` (String)
- `mailing_list` (Attributes Set) (see [below for nested schema](#nestedatt--rule--file_rule--tracking--alert--mailing_list))
- `subscription_group` (Attributes Set) (see [below for nested schema](#nestedatt--rule--file_rule--tracking--alert--subscription_group))
- `webhook` (Attributes Set) (see [below for nested schema](#nestedatt--rule--file_rule--tracking--alert--webhook))

<a id="nestedatt--rule--file_rule--tracking--alert--mailing_list"></a>
### Nested Schema for `rule.file_rule.tracking.alert.mailing_list`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--tracking--alert--subscription_group"></a>
### Nested Schema for `rule.file_rule.tracking.alert.subscription_group`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--tracking--alert--webhook"></a>
### Nested Schema for `rule.file_rule.tracking.alert.webhook`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--file_rule--tracking--event"></a>
### Nested Schema for `rule.file_rule.tracking.event`

Optional:

- `enabled` (Boolean)
