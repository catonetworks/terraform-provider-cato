# Application Control section — groups rules in the policy UI.
# Creating a section opens a draft policy revision; review in the Cato UI before publishing.
resource "cato_application_control_section" "example" {
  # position values: LAST_IN_POLICY | AFTER_SECTION | BEFORE_SECTION
  # Use ref to anchor relative positions; omit ref for LAST_IN_POLICY.
  at = {
    position = "LAST_IN_POLICY"
    # ref = cato_application_control_section.other.section.id
  }

  section = {
    name = "Terraform — Application Control"
  }
}
