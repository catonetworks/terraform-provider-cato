# Minimal APPLICATION rule (GraphQL @beta). Uncomment optional blocks as needed.
resource "cato_application_control_rule" "example" {
  at = {
    position = "LAST_IN_POLICY"
    # ref      = cato_application_control_section.custom.section.id
  }
  rule = {
    name        = "Example application control rule"
    description = "Managed by Terraform"
    enabled     = true
    rule_type   = "APPLICATION"
    application_rule = {
      action   = "MONITOR"
      severity = "LOW"
      tracking = {
        event = {
          enabled = false
        }
      }
      application = {}
      source      = {}
    }
  }
}
