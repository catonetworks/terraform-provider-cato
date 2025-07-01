provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token_sase
  account_id = var.account_id_sase
}

# --------------------------------------------------------------------------------
# README: Example CSV syntax for sections: sections.csv
# section_index,section_name
# 1,My First Section Here
# 2,My Second Section Name
# 3,My Third Section Name

# README: Example CSV syntax for sections: rules.csv
# index_in_section,section_name,rule_name
# 1,My First Section Here,My First Section 1 Rule
# 2,My First Section Here,My Second Section 1 Rule
# 1,My Second Section Name,My First Section 2 Rule
# 2,My Second Section Name,My Second Section 2 Rule
# 1,My Third Section Name,My First Section 3 Rule
# 2,My Third Section Name,My Second Section 3 Rule
# 3,My Third Section Name,My Third Section 3 Rule
# --------------------------------------------------------------------------------

locals {
  rule_data = csvdecode(file("${path.module}/rules.csv"))
  section_data = csvdecode(file("${path.module}/sections.csv"))
}

resource "cato_wf_section" "all_wf_sections" {
  for_each = { for section in local.section_data : section.section_index => section}
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = each.value.section_name
  }
}
resource "cato_wf_rule" "all_fw_rules" {
  at = {
    position = "LAST_IN_POLICY" // adding last, to reorder in cato_bulk_if_move_rule
  }
  for_each = { for rule in local.rule_data : rule.rule_name => rule }
  rule = {
    name        = each.value.rule_name
    description = each.value.rule_name
    enabled     = true
    direction   = "TO"
    source = {}
    destination = {}
    application = {}
    action = "ALLOW"
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

resource "cato_bulk_wf_move_rule" "all_wf_rules" {
 depends_on = [ cato_wf_section.all_wf_sections, cato_wf_rule.all_fw_rules ]
 rule_data = local.rule_data
 section_data = local.section_data
}

output "section_data" {
  value = local.section_data
}

output "rule_data" {
  value = local.rule_data
}