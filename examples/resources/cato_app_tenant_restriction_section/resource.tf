resource "cato_app_tenant_restriction_section" "custom" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Terraform tenant restriction section"
  }
}
