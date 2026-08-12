provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token_sase
  account_id = var.account_id_sase
}

locals {
  firewall_defs = jsondecode(file("data/lanFirewall.json"))
  section_data = {
    for section_name, data in local.firewall_defs.sections :
    section_name => {
      section_index   = data.location.index_in_policy,
      sub_policy_name = try(data.location.policy_name, null)
    }
  }
  network_rules = {
    for rule_name, data in local.firewall_defs.network_rules :
    rule_name => {
      index_in_section = data.location.index_in_section,
      section_name     = data.location.section_name
    }
  }
  sub_policies = {
    for pol_name, data in local.firewall_defs.policies :
    pol_name => {
      index_in_section = data.location.index_in_section,
      section_name     = data.location.section_name
    }
  }
  firewall_rules = {
    for rule_name, data in local.firewall_defs.firewall_rules :
    rule_name => {
      index_in_rule = data.location.index_in_rule,
      net_rule_name = data.location.network_rule_name
    }
  }
}


resource "cato_socket_lan_section" "primary_sections" {
  for_each = { for section_name, data in local.firewall_defs.sections : section_name => data if try(data.location.policy_name, "") == "" }
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = each.key
  }
}


resource "cato_lf_sub_policy" "all_policies" {
  for_each = { for policy_name, data in local.firewall_defs.policies : policy_name => data }

  name        = each.key
  description = each.value.input.description

  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_socket_lan_section.primary_sections[each.value.location.section_name].id
  }
  scope = each.value.input.scope
}


resource "cato_socket_lan_section" "sub_policy_sections" {
  for_each = { for section_name, data in local.firewall_defs.sections : section_name => data if try(data.location.policy_name, "") != "" }
  at = {
    position = "LAST_IN_POLICY"
    ref      = cato_lf_sub_policy.all_policies[each.value.location.policy_name].id
  }
  section = {
    name = each.key
  }
}


resource "cato_socket_lan_network_rule" "network_rules" {
  for_each = { for rule_name, data in local.firewall_defs.network_rules : rule_name => data }
  at = {
    position = "LAST_IN_SECTION"
    ref      = try(cato_socket_lan_section.primary_sections[each.value.location.section_name].id, cato_socket_lan_section.sub_policy_sections[each.value.location.section_name].id)
  }
  rule = each.value.input.rule
}


resource "cato_socket_lan_firewall_rule" "firewall_rules" {
  for_each = { for rule_name, data in local.firewall_defs.firewall_rules : rule_name => data }
  at = {
    position = "LAST_IN_RULE"
    ref      = cato_socket_lan_network_rule.network_rules[each.value.location.network_rule_name].rule.id
  }
  rule = each.value.input.rule
}


resource "cato_bulk_lf_move_rule" "reorder" {
  section_data   = local.section_data
  network_rules  = merge(local.network_rules, local.sub_policies)
  firewall_rules = local.firewall_rules

  depends_on = [
    cato_socket_lan_section.primary_sections,
    cato_socket_lan_network_rule.network_rules,
    cato_lf_sub_policy.all_policies,
    cato_socket_lan_section.sub_policy_sections,
    cato_socket_lan_firewall_rule.firewall_rules
  ]
}


# --------------------------------------------------------------------------------
# README: Example JSON for sections, network rules, firewall rules, sub-policies
# {
#   "sections": {
#     "Section-1": {
#       "location": { "index_in_policy": 1 }
#     },
# 
#     "Section-1 [Sub-policy-1]": {
#       "location": { "index_in_policy": 1, "policy_name": "Sub-policy-1" }
#     }
#   },
# 
# 
#   "network_rules": {
#     "Network-Rule-1": {
#       "input": {
#         "rule": {
#           "name": "Network-Rule-1",
#           "enabled": true,
#           "direction": "BOTH",
#           "transport": "LAN",
#           "site": {},
#           "source": {},
#           "destination": {}
#         }
#       },
#       "location": { "index_in_section": 1, "section_name": "Section-1" }
#     },
# 
#     "Network-Rule-1 [Sub-policy-1]": {
#       "input": {
#         "rule": {
#           "name": "Network-Rule-1 [Sub-policy-1]",
#           "enabled": true,
#           "direction": "BOTH",
#           "transport": "LAN",
#           "site": {},
#           "source": {},
#           "destination": {}
#         }
#       },
#       "location": { "index_in_section": 1, "section_name": "Section-1 [Sub-policy-1]" }
#     }
#   },
# 
# 
#   "firewall_rules": {
#     "Firewall-Rule-1 [Network-Rule-1]": {
#       "input": {
#         "rule": {
#           "name": "Firewall-Rule-1 [Network-Rule-1]",
#           "enabled": true,
#           "direction": "TO",
#           "action": "ALLOW",
#           "site": {},
#           "source": {},
#           "destination": {},
#           "tracking": { "event": { "enabled": false }, "alert": { "enabled": false } }
#         }
#       },
#       "location": { "index_in_rule": 1, "network_rule_name": "Network-Rule-1" }
#     }
#   },
# 
# 
#   "policies": {
#     "Sub-policy-1": {
#       "input": {
#         "description": "Example sub-policy",
#         "scope": {
#           "destination": {},
#           "direction": "TO",
#           "enabled": true,
#           "site": {},
#           "source": {}
#         }
#       },
#       "location": { "index_in_section": 2, "section_name": "Section-1" }
#     }
#   }
# }
# --------------------------------------------------------------------------------
