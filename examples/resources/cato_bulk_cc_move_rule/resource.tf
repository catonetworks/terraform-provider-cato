# Client Connectivity: bulk rule/section ORDERING (move) resource.
#
# cato_bulk_cc_move_rule reorders existing Client Connectivity rules and sections into the order defined by
# two CSV files, keyed by name. Workflow:
#   1. create the sections (below) and the rules (see the cato_client_connectivity_rule example),
#   2. let this resource own their final ordering in the policy.
#
# --------------------------------------------------------------------------------
# Example CSV syntax for sections: sections.csv
# section_index,section_name
# 1,My First Section Here
# 2,My Second Section Name
# 3,My Third Section Name
#
# Example CSV syntax for rules: rules.csv
# index_in_section,section_name,rule_name
# 1,My First Section Here,My First Section 1 Rule
# 2,My First Section Here,My Second Section 1 Rule
# 1,My Second Section Name,My First Section 2 Rule
# --------------------------------------------------------------------------------

locals {
  rule_data_list    = csvdecode(file("rules.csv"))
  section_data_list = csvdecode(file("sections.csv"))

  # Provider schema wants name-keyed maps, not lists.
  section_data = { for section in local.section_data_list : section.section_name => section }
  rule_data    = { for rule in local.rule_data_list : rule.rule_name => rule }
}

# Sections are ordering anchors. (Create the rules themselves with cato_client_connectivity_rule.)
resource "cato_client_connectivity_section" "all" {
  for_each = { for section in local.section_data : section.section_index => section }
  at = {
    position = "LAST_IN_POLICY" // add last; this move resource sets final order
  }
  section = {
    name = each.value.section_name
  }
}

# Apply the CSV-defined ordering across the whole Client Connectivity policy.
resource "cato_bulk_cc_move_rule" "all" {
  depends_on   = [cato_client_connectivity_section.all]
  rule_data    = local.rule_data
  section_data = local.section_data
}

output "section_data" {
  value = local.section_data
}

output "rule_data" {
  value = local.rule_data
}
