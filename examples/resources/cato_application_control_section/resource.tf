# Creating a section can attach existing rules to the new section and open a draft
# policy revision. Review impact in the Cato UI before publish.
resource "cato_application_control_section" "custom" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Terraform application control section"
  }
}
