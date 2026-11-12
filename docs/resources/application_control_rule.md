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
- `application_criteria` (Attributes) Attribute-based application matching (compliance/security attributes, risk score, origin country, AI risk). Typically combined with application.application_type. Omitted attributes default to ANY. (see [below for nested schema](#nestedatt--rule--application_rule--application_criteria))
- `application_criteria_satisfy` (String) How application criteria are combined (ALL | ANY)
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

- `id` (String) User notification ID
- `name` (String) User notification name



<a id="nestedatt--rule--application_rule--application"></a>
### Nested Schema for `rule.application_rule.application`

Optional:

- `app_category` (Attributes Set) Application categories (see [below for nested schema](#nestedatt--rule--application_rule--application--app_category))
- `application` (Attributes Set) Predefined applications (logical OR within the set) (see [below for nested schema](#nestedatt--rule--application_rule--application--application))
- `application_type` (List of String) Match applications by type instead of by reference (APPLICATION, CLOUD_APPLICATION, SERVICE). Combine with application_criteria on the rule to match by attributes (e.g. all CLOUD_APPLICATIONs supporting PCI DSS).
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

- `id` (String) App category ID
- `name` (String) App category name


<a id="nestedatt--rule--application_rule--application--application"></a>
### Nested Schema for `rule.application_rule.application.application`

Optional:

- `id` (String) Application ID
- `name` (String) Application name


<a id="nestedatt--rule--application_rule--application--custom_app"></a>
### Nested Schema for `rule.application_rule.application.custom_app`

Optional:

- `id` (String) Custom application ID
- `name` (String) Custom application name


<a id="nestedatt--rule--application_rule--application--custom_category"></a>
### Nested Schema for `rule.application_rule.application.custom_category`

Optional:

- `id` (String) Custom category ID
- `name` (String) Custom category name


<a id="nestedatt--rule--application_rule--application--global_ip_range"></a>
### Nested Schema for `rule.application_rule.application.global_ip_range`

Optional:

- `id` (String) Global IP range ID
- `name` (String) Global IP range name


<a id="nestedatt--rule--application_rule--application--ip_range"></a>
### Nested Schema for `rule.application_rule.application.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--application_rule--application--sanctioned_apps_category"></a>
### Nested Schema for `rule.application_rule.application.sanctioned_apps_category`

Optional:

- `id` (String) Sanctioned apps category ID
- `name` (String) Sanctioned apps category name



<a id="nestedatt--rule--application_rule--application_activity"></a>
### Nested Schema for `rule.application_rule.application_activity`

Required:

- `activity` (Attributes) (see [below for nested schema](#nestedatt--rule--application_rule--application_activity--activity))

Optional:

- `field` (Attributes) Activity condition field (e.g. URL) — some activities require a condition (see [below for nested schema](#nestedatt--rule--application_rule--application_activity--field))
- `operator` (String) Activity condition operator (e.g. IS, CONTAINS)
- `value` (String) Activity condition value (required by some activities, e.g. Full Path URL)
- `value_set` (Attributes) Activity condition value set reference (see [below for nested schema](#nestedatt--rule--application_rule--application_activity--value_set))

<a id="nestedatt--rule--application_rule--application_activity--activity"></a>
### Nested Schema for `rule.application_rule.application_activity.activity`

Optional:

- `id` (String) Activity ID
- `name` (String) Activity name


<a id="nestedatt--rule--application_rule--application_activity--field"></a>
### Nested Schema for `rule.application_rule.application_activity.field`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--application_activity--value_set"></a>
### Nested Schema for `rule.application_rule.application_activity.value_set`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--application_rule--application_criteria"></a>
### Nested Schema for `rule.application_rule.application_criteria`

Optional:

- `ai_risk` (List of String) AI risk levels (LOW, MEDIUM, HIGH, CRITICAL)
- `compliance_attributes` (Map of String) Compliance attribute constraints keyed by API field name (hippa, isae3402, iso27001, pciDss, soc1, soc2, soc3, sox). Values: SUPPORTED or NOT_SUPPORTED — omit a key for ANY.
- `origin_country` (Attributes Set) Application origin countries (see [below for nested schema](#nestedatt--rule--application_rule--application_criteria--origin_country))
- `risk` (Attributes List) Risk score constraints (see [below for nested schema](#nestedatt--rule--application_rule--application_criteria--risk))
- `security_attributes` (Map of String) Security attribute constraints keyed by API field name (auditTrail, encryptionAtRest, httpSecurityHeaders, mfa, rbac, rememberPassword, sso, tlsEnforcement, trustedCertificate). Values: SUPPORTED or NOT_SUPPORTED — omit a key for ANY.

<a id="nestedatt--rule--application_rule--application_criteria--origin_country"></a>
### Nested Schema for `rule.application_rule.application_criteria.origin_country`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--application_criteria--risk"></a>
### Nested Schema for `rule.application_rule.application_criteria.risk`

Required:

- `risk` (String) Risk score (e.g. "7")
- `risk_operator` (String) Comparison operator (e.g. GREATER_THAN, IS)



<a id="nestedatt--rule--application_rule--device"></a>
### Nested Schema for `rule.application_rule.device`

Optional:

- `id` (String) Device ID
- `name` (String) Device name


<a id="nestedatt--rule--application_rule--dlp_profile"></a>
### Nested Schema for `rule.application_rule.dlp_profile`

Optional:

- `content_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--application_rule--dlp_profile--content_profile))
- `edm_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--application_rule--dlp_profile--edm_profile))

<a id="nestedatt--rule--application_rule--dlp_profile--content_profile"></a>
### Nested Schema for `rule.application_rule.dlp_profile.content_profile`

Optional:

- `id` (String) Content profile ID
- `name` (String) Content profile name


<a id="nestedatt--rule--application_rule--dlp_profile--edm_profile"></a>
### Nested Schema for `rule.application_rule.dlp_profile.edm_profile`

Optional:

- `id` (String) EDM profile ID
- `name` (String) EDM profile name



<a id="nestedatt--rule--application_rule--file_attribute"></a>
### Nested Schema for `rule.application_rule.file_attribute`

Required:

- `file_attribute` (String)
- `operator` (String)

Optional:

- `content_type_group_values` (Attributes Set) Content type GROUP references (CONTENT_TYPE rows, e.g. Binaries, Scripts) (see [below for nested schema](#nestedatt--rule--application_rule--file_attribute--content_type_group_values))
- `content_type_values` (Attributes Set) Content type references (CONTENT_TYPE rows) (see [below for nested schema](#nestedatt--rule--application_rule--file_attribute--content_type_values))
- `value` (String) Attribute value (e.g. a size threshold). MUST be omitted for CONTENT_TYPE rows — use the content-type ref sets instead.

<a id="nestedatt--rule--application_rule--file_attribute--content_type_group_values"></a>
### Nested Schema for `rule.application_rule.file_attribute.content_type_group_values`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--application_rule--file_attribute--content_type_values"></a>
### Nested Schema for `rule.application_rule.file_attribute.content_type_values`

Optional:

- `id` (String)
- `name` (String)



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

- `id` (String) Country ID
- `name` (String) Country name


<a id="nestedatt--rule--application_rule--source--floating_subnet"></a>
### Nested Schema for `rule.application_rule.source.floating_subnet`

Optional:

- `id` (String) Floating subnet ID
- `name` (String) Floating subnet name


<a id="nestedatt--rule--application_rule--source--global_ip_range"></a>
### Nested Schema for `rule.application_rule.source.global_ip_range`

Optional:

- `id` (String) Global IP range ID
- `name` (String) Global IP range name


<a id="nestedatt--rule--application_rule--source--group"></a>
### Nested Schema for `rule.application_rule.source.group`

Optional:

- `id` (String) Group ID
- `name` (String) Group name


<a id="nestedatt--rule--application_rule--source--host"></a>
### Nested Schema for `rule.application_rule.source.host`

Optional:

- `id` (String) Host ID
- `name` (String) Host name


<a id="nestedatt--rule--application_rule--source--ip_range"></a>
### Nested Schema for `rule.application_rule.source.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--application_rule--source--network_interface"></a>
### Nested Schema for `rule.application_rule.source.network_interface`

Optional:

- `id` (String) Network interface ID
- `name` (String) Network interface name


<a id="nestedatt--rule--application_rule--source--site"></a>
### Nested Schema for `rule.application_rule.source.site`

Optional:

- `id` (String) Site ID
- `name` (String) Site name


<a id="nestedatt--rule--application_rule--source--site_network_subnet"></a>
### Nested Schema for `rule.application_rule.source.site_network_subnet`

Optional:

- `id` (String) Site network subnet ID
- `name` (String) Site network subnet name


<a id="nestedatt--rule--application_rule--source--system_group"></a>
### Nested Schema for `rule.application_rule.source.system_group`

Optional:

- `id` (String) System group ID
- `name` (String) System group name


<a id="nestedatt--rule--application_rule--source--user"></a>
### Nested Schema for `rule.application_rule.source.user`

Optional:

- `id` (String) User ID
- `name` (String) User name


<a id="nestedatt--rule--application_rule--source--users_group"></a>
### Nested Schema for `rule.application_rule.source.users_group`

Optional:

- `id` (String) User group ID
- `name` (String) User group name



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

- `id` (String) Mailing list ID
- `name` (String) Mailing list name


<a id="nestedatt--rule--application_rule--tracking--alert--subscription_group"></a>
### Nested Schema for `rule.application_rule.tracking.alert.subscription_group`

Optional:

- `id` (String) Subscription group ID
- `name` (String) Subscription group name


<a id="nestedatt--rule--application_rule--tracking--alert--webhook"></a>
### Nested Schema for `rule.application_rule.tracking.alert.webhook`

Optional:

- `id` (String) Webhook ID
- `name` (String) Webhook name



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
- `application_criteria` (Attributes) Attribute-based application matching (compliance/security attributes, risk score, origin country, AI risk). Typically combined with application.application_type. Omitted attributes default to ANY. (see [below for nested schema](#nestedatt--rule--data_rule--application_criteria))
- `application_criteria_satisfy` (String) How application criteria are combined (ALL | ANY)
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

- `id` (String) User notification ID
- `name` (String) User notification name



<a id="nestedatt--rule--data_rule--application"></a>
### Nested Schema for `rule.data_rule.application`

Optional:

- `app_category` (Attributes Set) Application categories (see [below for nested schema](#nestedatt--rule--data_rule--application--app_category))
- `application` (Attributes Set) Predefined applications (logical OR within the set) (see [below for nested schema](#nestedatt--rule--data_rule--application--application))
- `application_type` (List of String) Match applications by type instead of by reference (APPLICATION, CLOUD_APPLICATION, SERVICE). Combine with application_criteria on the rule to match by attributes (e.g. all CLOUD_APPLICATIONs supporting PCI DSS).
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

- `id` (String) App category ID
- `name` (String) App category name


<a id="nestedatt--rule--data_rule--application--application"></a>
### Nested Schema for `rule.data_rule.application.application`

Optional:

- `id` (String) Application ID
- `name` (String) Application name


<a id="nestedatt--rule--data_rule--application--custom_app"></a>
### Nested Schema for `rule.data_rule.application.custom_app`

Optional:

- `id` (String) Custom application ID
- `name` (String) Custom application name


<a id="nestedatt--rule--data_rule--application--custom_category"></a>
### Nested Schema for `rule.data_rule.application.custom_category`

Optional:

- `id` (String) Custom category ID
- `name` (String) Custom category name


<a id="nestedatt--rule--data_rule--application--global_ip_range"></a>
### Nested Schema for `rule.data_rule.application.global_ip_range`

Optional:

- `id` (String) Global IP range ID
- `name` (String) Global IP range name


<a id="nestedatt--rule--data_rule--application--ip_range"></a>
### Nested Schema for `rule.data_rule.application.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--data_rule--application--sanctioned_apps_category"></a>
### Nested Schema for `rule.data_rule.application.sanctioned_apps_category`

Optional:

- `id` (String) Sanctioned apps category ID
- `name` (String) Sanctioned apps category name



<a id="nestedatt--rule--data_rule--application_activity"></a>
### Nested Schema for `rule.data_rule.application_activity`

Required:

- `activity` (Attributes) (see [below for nested schema](#nestedatt--rule--data_rule--application_activity--activity))

Optional:

- `field` (Attributes) Activity condition field (e.g. URL) — some activities require a condition (see [below for nested schema](#nestedatt--rule--data_rule--application_activity--field))
- `operator` (String) Activity condition operator (e.g. IS, CONTAINS)
- `value` (String) Activity condition value (required by some activities, e.g. Full Path URL)
- `value_set` (Attributes) Activity condition value set reference (see [below for nested schema](#nestedatt--rule--data_rule--application_activity--value_set))

<a id="nestedatt--rule--data_rule--application_activity--activity"></a>
### Nested Schema for `rule.data_rule.application_activity.activity`

Optional:

- `id` (String) Activity ID
- `name` (String) Activity name


<a id="nestedatt--rule--data_rule--application_activity--field"></a>
### Nested Schema for `rule.data_rule.application_activity.field`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--application_activity--value_set"></a>
### Nested Schema for `rule.data_rule.application_activity.value_set`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--data_rule--application_criteria"></a>
### Nested Schema for `rule.data_rule.application_criteria`

Optional:

- `ai_risk` (List of String) AI risk levels (LOW, MEDIUM, HIGH, CRITICAL)
- `compliance_attributes` (Map of String) Compliance attribute constraints keyed by API field name (hippa, isae3402, iso27001, pciDss, soc1, soc2, soc3, sox). Values: SUPPORTED or NOT_SUPPORTED — omit a key for ANY.
- `origin_country` (Attributes Set) Application origin countries (see [below for nested schema](#nestedatt--rule--data_rule--application_criteria--origin_country))
- `risk` (Attributes List) Risk score constraints (see [below for nested schema](#nestedatt--rule--data_rule--application_criteria--risk))
- `security_attributes` (Map of String) Security attribute constraints keyed by API field name (auditTrail, encryptionAtRest, httpSecurityHeaders, mfa, rbac, rememberPassword, sso, tlsEnforcement, trustedCertificate). Values: SUPPORTED or NOT_SUPPORTED — omit a key for ANY.

<a id="nestedatt--rule--data_rule--application_criteria--origin_country"></a>
### Nested Schema for `rule.data_rule.application_criteria.origin_country`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--application_criteria--risk"></a>
### Nested Schema for `rule.data_rule.application_criteria.risk`

Required:

- `risk` (String) Risk score (e.g. "7")
- `risk_operator` (String) Comparison operator (e.g. GREATER_THAN, IS)



<a id="nestedatt--rule--data_rule--device"></a>
### Nested Schema for `rule.data_rule.device`

Optional:

- `id` (String) Device ID
- `name` (String) Device name


<a id="nestedatt--rule--data_rule--dlp_profile"></a>
### Nested Schema for `rule.data_rule.dlp_profile`

Optional:

- `content_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--data_rule--dlp_profile--content_profile))
- `edm_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--data_rule--dlp_profile--edm_profile))

<a id="nestedatt--rule--data_rule--dlp_profile--content_profile"></a>
### Nested Schema for `rule.data_rule.dlp_profile.content_profile`

Optional:

- `id` (String) Content profile ID
- `name` (String) Content profile name


<a id="nestedatt--rule--data_rule--dlp_profile--edm_profile"></a>
### Nested Schema for `rule.data_rule.dlp_profile.edm_profile`

Optional:

- `id` (String) EDM profile ID
- `name` (String) EDM profile name



<a id="nestedatt--rule--data_rule--file_attribute"></a>
### Nested Schema for `rule.data_rule.file_attribute`

Required:

- `file_attribute` (String)
- `operator` (String)

Optional:

- `content_type_group_values` (Attributes Set) Content type GROUP references (CONTENT_TYPE rows, e.g. Binaries, Scripts) (see [below for nested schema](#nestedatt--rule--data_rule--file_attribute--content_type_group_values))
- `content_type_values` (Attributes Set) Content type references (CONTENT_TYPE rows) (see [below for nested schema](#nestedatt--rule--data_rule--file_attribute--content_type_values))
- `value` (String) Attribute value (e.g. a size threshold). MUST be omitted for CONTENT_TYPE rows — use the content-type ref sets instead.

<a id="nestedatt--rule--data_rule--file_attribute--content_type_group_values"></a>
### Nested Schema for `rule.data_rule.file_attribute.content_type_group_values`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--data_rule--file_attribute--content_type_values"></a>
### Nested Schema for `rule.data_rule.file_attribute.content_type_values`

Optional:

- `id` (String)
- `name` (String)



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

- `id` (String) Country ID
- `name` (String) Country name


<a id="nestedatt--rule--data_rule--source--floating_subnet"></a>
### Nested Schema for `rule.data_rule.source.floating_subnet`

Optional:

- `id` (String) Floating subnet ID
- `name` (String) Floating subnet name


<a id="nestedatt--rule--data_rule--source--global_ip_range"></a>
### Nested Schema for `rule.data_rule.source.global_ip_range`

Optional:

- `id` (String) Global IP range ID
- `name` (String) Global IP range name


<a id="nestedatt--rule--data_rule--source--group"></a>
### Nested Schema for `rule.data_rule.source.group`

Optional:

- `id` (String) Group ID
- `name` (String) Group name


<a id="nestedatt--rule--data_rule--source--host"></a>
### Nested Schema for `rule.data_rule.source.host`

Optional:

- `id` (String) Host ID
- `name` (String) Host name


<a id="nestedatt--rule--data_rule--source--ip_range"></a>
### Nested Schema for `rule.data_rule.source.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--data_rule--source--network_interface"></a>
### Nested Schema for `rule.data_rule.source.network_interface`

Optional:

- `id` (String) Network interface ID
- `name` (String) Network interface name


<a id="nestedatt--rule--data_rule--source--site"></a>
### Nested Schema for `rule.data_rule.source.site`

Optional:

- `id` (String) Site ID
- `name` (String) Site name


<a id="nestedatt--rule--data_rule--source--site_network_subnet"></a>
### Nested Schema for `rule.data_rule.source.site_network_subnet`

Optional:

- `id` (String) Site network subnet ID
- `name` (String) Site network subnet name


<a id="nestedatt--rule--data_rule--source--system_group"></a>
### Nested Schema for `rule.data_rule.source.system_group`

Optional:

- `id` (String) System group ID
- `name` (String) System group name


<a id="nestedatt--rule--data_rule--source--user"></a>
### Nested Schema for `rule.data_rule.source.user`

Optional:

- `id` (String) User ID
- `name` (String) User name


<a id="nestedatt--rule--data_rule--source--users_group"></a>
### Nested Schema for `rule.data_rule.source.users_group`

Optional:

- `id` (String) User group ID
- `name` (String) User group name



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

- `id` (String) Mailing list ID
- `name` (String) Mailing list name


<a id="nestedatt--rule--data_rule--tracking--alert--subscription_group"></a>
### Nested Schema for `rule.data_rule.tracking.alert.subscription_group`

Optional:

- `id` (String) Subscription group ID
- `name` (String) Subscription group name


<a id="nestedatt--rule--data_rule--tracking--alert--webhook"></a>
### Nested Schema for `rule.data_rule.tracking.alert.webhook`

Optional:

- `id` (String) Webhook ID
- `name` (String) Webhook name



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
- `application_criteria` (Attributes) Attribute-based application matching (compliance/security attributes, risk score, origin country, AI risk). Typically combined with application.application_type. Omitted attributes default to ANY. (see [below for nested schema](#nestedatt--rule--file_rule--application_criteria))
- `application_criteria_satisfy` (String) How application criteria are combined (ALL | ANY)
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

- `id` (String) User notification ID
- `name` (String) User notification name



<a id="nestedatt--rule--file_rule--application"></a>
### Nested Schema for `rule.file_rule.application`

Optional:

- `app_category` (Attributes Set) Application categories (see [below for nested schema](#nestedatt--rule--file_rule--application--app_category))
- `application` (Attributes Set) Predefined applications (logical OR within the set) (see [below for nested schema](#nestedatt--rule--file_rule--application--application))
- `application_type` (List of String) Match applications by type instead of by reference (APPLICATION, CLOUD_APPLICATION, SERVICE). Combine with application_criteria on the rule to match by attributes (e.g. all CLOUD_APPLICATIONs supporting PCI DSS).
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

- `id` (String) App category ID
- `name` (String) App category name


<a id="nestedatt--rule--file_rule--application--application"></a>
### Nested Schema for `rule.file_rule.application.application`

Optional:

- `id` (String) Application ID
- `name` (String) Application name


<a id="nestedatt--rule--file_rule--application--custom_app"></a>
### Nested Schema for `rule.file_rule.application.custom_app`

Optional:

- `id` (String) Custom application ID
- `name` (String) Custom application name


<a id="nestedatt--rule--file_rule--application--custom_category"></a>
### Nested Schema for `rule.file_rule.application.custom_category`

Optional:

- `id` (String) Custom category ID
- `name` (String) Custom category name


<a id="nestedatt--rule--file_rule--application--global_ip_range"></a>
### Nested Schema for `rule.file_rule.application.global_ip_range`

Optional:

- `id` (String) Global IP range ID
- `name` (String) Global IP range name


<a id="nestedatt--rule--file_rule--application--ip_range"></a>
### Nested Schema for `rule.file_rule.application.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--file_rule--application--sanctioned_apps_category"></a>
### Nested Schema for `rule.file_rule.application.sanctioned_apps_category`

Optional:

- `id` (String) Sanctioned apps category ID
- `name` (String) Sanctioned apps category name



<a id="nestedatt--rule--file_rule--application_activity"></a>
### Nested Schema for `rule.file_rule.application_activity`

Required:

- `activity` (Attributes) (see [below for nested schema](#nestedatt--rule--file_rule--application_activity--activity))

Optional:

- `field` (Attributes) Activity condition field (e.g. URL) — some activities require a condition (see [below for nested schema](#nestedatt--rule--file_rule--application_activity--field))
- `operator` (String) Activity condition operator (e.g. IS, CONTAINS)
- `value` (String) Activity condition value (required by some activities, e.g. Full Path URL)
- `value_set` (Attributes) Activity condition value set reference (see [below for nested schema](#nestedatt--rule--file_rule--application_activity--value_set))

<a id="nestedatt--rule--file_rule--application_activity--activity"></a>
### Nested Schema for `rule.file_rule.application_activity.activity`

Optional:

- `id` (String) Activity ID
- `name` (String) Activity name


<a id="nestedatt--rule--file_rule--application_activity--field"></a>
### Nested Schema for `rule.file_rule.application_activity.field`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--application_activity--value_set"></a>
### Nested Schema for `rule.file_rule.application_activity.value_set`

Optional:

- `id` (String)
- `name` (String)



<a id="nestedatt--rule--file_rule--application_criteria"></a>
### Nested Schema for `rule.file_rule.application_criteria`

Optional:

- `ai_risk` (List of String) AI risk levels (LOW, MEDIUM, HIGH, CRITICAL)
- `compliance_attributes` (Map of String) Compliance attribute constraints keyed by API field name (hippa, isae3402, iso27001, pciDss, soc1, soc2, soc3, sox). Values: SUPPORTED or NOT_SUPPORTED — omit a key for ANY.
- `origin_country` (Attributes Set) Application origin countries (see [below for nested schema](#nestedatt--rule--file_rule--application_criteria--origin_country))
- `risk` (Attributes List) Risk score constraints (see [below for nested schema](#nestedatt--rule--file_rule--application_criteria--risk))
- `security_attributes` (Map of String) Security attribute constraints keyed by API field name (auditTrail, encryptionAtRest, httpSecurityHeaders, mfa, rbac, rememberPassword, sso, tlsEnforcement, trustedCertificate). Values: SUPPORTED or NOT_SUPPORTED — omit a key for ANY.

<a id="nestedatt--rule--file_rule--application_criteria--origin_country"></a>
### Nested Schema for `rule.file_rule.application_criteria.origin_country`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--application_criteria--risk"></a>
### Nested Schema for `rule.file_rule.application_criteria.risk`

Required:

- `risk` (String) Risk score (e.g. "7")
- `risk_operator` (String) Comparison operator (e.g. GREATER_THAN, IS)



<a id="nestedatt--rule--file_rule--device"></a>
### Nested Schema for `rule.file_rule.device`

Optional:

- `id` (String) Device ID
- `name` (String) Device name


<a id="nestedatt--rule--file_rule--dlp_profile"></a>
### Nested Schema for `rule.file_rule.dlp_profile`

Optional:

- `content_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--file_rule--dlp_profile--content_profile))
- `edm_profile` (Attributes Set) (see [below for nested schema](#nestedatt--rule--file_rule--dlp_profile--edm_profile))

<a id="nestedatt--rule--file_rule--dlp_profile--content_profile"></a>
### Nested Schema for `rule.file_rule.dlp_profile.content_profile`

Optional:

- `id` (String) Content profile ID
- `name` (String) Content profile name


<a id="nestedatt--rule--file_rule--dlp_profile--edm_profile"></a>
### Nested Schema for `rule.file_rule.dlp_profile.edm_profile`

Optional:

- `id` (String) EDM profile ID
- `name` (String) EDM profile name



<a id="nestedatt--rule--file_rule--file_attribute"></a>
### Nested Schema for `rule.file_rule.file_attribute`

Required:

- `file_attribute` (String)
- `operator` (String)

Optional:

- `content_type_group_values` (Attributes Set) Content type GROUP references (CONTENT_TYPE rows, e.g. Binaries, Scripts) (see [below for nested schema](#nestedatt--rule--file_rule--file_attribute--content_type_group_values))
- `content_type_values` (Attributes Set) Content type references (CONTENT_TYPE rows) (see [below for nested schema](#nestedatt--rule--file_rule--file_attribute--content_type_values))
- `value` (String) Attribute value (e.g. a size threshold). MUST be omitted for CONTENT_TYPE rows — use the content-type ref sets instead.

<a id="nestedatt--rule--file_rule--file_attribute--content_type_group_values"></a>
### Nested Schema for `rule.file_rule.file_attribute.content_type_group_values`

Optional:

- `id` (String)
- `name` (String)


<a id="nestedatt--rule--file_rule--file_attribute--content_type_values"></a>
### Nested Schema for `rule.file_rule.file_attribute.content_type_values`

Optional:

- `id` (String)
- `name` (String)



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

- `id` (String) Country ID
- `name` (String) Country name


<a id="nestedatt--rule--file_rule--source--floating_subnet"></a>
### Nested Schema for `rule.file_rule.source.floating_subnet`

Optional:

- `id` (String) Floating subnet ID
- `name` (String) Floating subnet name


<a id="nestedatt--rule--file_rule--source--global_ip_range"></a>
### Nested Schema for `rule.file_rule.source.global_ip_range`

Optional:

- `id` (String) Global IP range ID
- `name` (String) Global IP range name


<a id="nestedatt--rule--file_rule--source--group"></a>
### Nested Schema for `rule.file_rule.source.group`

Optional:

- `id` (String) Group ID
- `name` (String) Group name


<a id="nestedatt--rule--file_rule--source--host"></a>
### Nested Schema for `rule.file_rule.source.host`

Optional:

- `id` (String) Host ID
- `name` (String) Host name


<a id="nestedatt--rule--file_rule--source--ip_range"></a>
### Nested Schema for `rule.file_rule.source.ip_range`

Optional:

- `from` (String)
- `to` (String)


<a id="nestedatt--rule--file_rule--source--network_interface"></a>
### Nested Schema for `rule.file_rule.source.network_interface`

Optional:

- `id` (String) Network interface ID
- `name` (String) Network interface name


<a id="nestedatt--rule--file_rule--source--site"></a>
### Nested Schema for `rule.file_rule.source.site`

Optional:

- `id` (String) Site ID
- `name` (String) Site name


<a id="nestedatt--rule--file_rule--source--site_network_subnet"></a>
### Nested Schema for `rule.file_rule.source.site_network_subnet`

Optional:

- `id` (String) Site network subnet ID
- `name` (String) Site network subnet name


<a id="nestedatt--rule--file_rule--source--system_group"></a>
### Nested Schema for `rule.file_rule.source.system_group`

Optional:

- `id` (String) System group ID
- `name` (String) System group name


<a id="nestedatt--rule--file_rule--source--user"></a>
### Nested Schema for `rule.file_rule.source.user`

Optional:

- `id` (String) User ID
- `name` (String) User name


<a id="nestedatt--rule--file_rule--source--users_group"></a>
### Nested Schema for `rule.file_rule.source.users_group`

Optional:

- `id` (String) User group ID
- `name` (String) User group name



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

- `id` (String) Mailing list ID
- `name` (String) Mailing list name


<a id="nestedatt--rule--file_rule--tracking--alert--subscription_group"></a>
### Nested Schema for `rule.file_rule.tracking.alert.subscription_group`

Optional:

- `id` (String) Subscription group ID
- `name` (String) Subscription group name


<a id="nestedatt--rule--file_rule--tracking--alert--webhook"></a>
### Nested Schema for `rule.file_rule.tracking.alert.webhook`

Optional:

- `id` (String) Webhook ID
- `name` (String) Webhook name



<a id="nestedatt--rule--file_rule--tracking--event"></a>
### Nested Schema for `rule.file_rule.tracking.event`

Optional:

- `enabled` (Boolean)
