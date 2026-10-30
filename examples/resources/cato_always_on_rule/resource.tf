# Terraform provider configuration
terraform {
  required_providers {
    cato = {
      source = "registry.terraform.io/catonetworks/cato"
    }
  }
}

variable "account_id" {
  description = "Cato account ID"
  type        = string
}

variable "cato_token" {
  description = "Cato API token"
  type    = string
}

variable "baseurl" {
  description = "Cato API base URL"
  type    = string
}

provider "cato" {
  baseurl    = var.baseurl
  account_id = var.account_id
  token      = var.cato_token
}


# Local data from the JSON file
locals {
  always_on_defs = jsondecode(file("${path.module}/data/rules.json"))
  section_data = {
    for section_name, data in local.always_on_defs.sections :
    section_name => {
      section_index = data.location.index_in_policy
    }
  }
  rule_data = {
    for rule_name, data in local.always_on_defs.rules :
    rule_name => {
      rule         = merge(data.input.rule, { "name" : rule_name })
      section_name = data.location.section_name
    }
  }
  reorder_data = {
    for rule_name, data in local.always_on_defs.rules :
    rule_name => {
      index_in_section = data.location.index_in_section,
      section_name     = data.location.section_name
    }
  }
}

# Enable or disable always-on policy
resource "cato_always_on_policy" "this" {
  enabled = "true"
}

# Always-on policy Sections 
resource "cato_always_on_section" "all" {
  depends_on   = [cato_always_on_policy.this]
  for_each = local.section_data
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = each.key
  }
}

# Always-on policy rules
resource "cato_always_on_rule" "all" {
  depends_on   = [cato_always_on_section.all]
  for_each = local.rule_data
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_always_on_section.all[each.value.section_name].section.id
  }
  rule = each.value.rule
}

# Ordering and publishing
resource "cato_always_on_rule_bulk" "reorder" {
  depends_on   = [cato_always_on_section.all, cato_always_on_rule.all]
  section_data = local.section_data
  rule_data    = local.reorder_data
}

# --------------------------------------------------------------------------------
# Example of rules.json file
# ~~~~~~~~~~~~~~~~~~~~~~~~~~
# 
# {
#   "sections": {
#     "Section-1": {
#       "location": {
#         "index_in_policy": 1
#       }
#     }
#   },
#   "rules": {
#     "Example Always On rule": {
#       "input": {
#         "rule": {
#           "action": "ENFORCE",
#           "allow_fail_open": false,
#           "allow_user_bypass": false,
#           "anti_tamper_mode": "PROTECT",
#           "bypass_duration": {
#             "time": 30,
#             "unit": "MINUTES"
#           },
#           "description": "Enforce Always On for Windows and macOS devices",
#           "device_posture_profile": [
#             { "name": "Example device posture 1" }
#           ],
#           "enabled": true,
#           "platform": [ "WINDOWS", "MACOS" ],
#           "source": {
#             "user_groups": [
#               { "name": "user_group_1" },
#               { "name": "user_group_2" }
#             ],
#             "users": [
#               { "name": "user_1" }
#             ]
#           }
#         }
#       },
#       "location": {
#         "index_in_section": 1,
#         "section_name": "Section-1"
#       }
#     }
#   }
# }
