# Use the names of an existing policy section and rule.
resource "cato_bulk_sdp_static_ip_allocation_move_rule" "policy_order" {
  section_data = {
    "Office" = { section_index = 1 }
  }
  rule_data = {
    "Allow HTTPS" = { index_in_section = 1, section_name = "Office" }
  }
}
