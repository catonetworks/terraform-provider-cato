# App Tenant Restriction section — groups ATR rules in the policy UI.
# GraphQL for this policy is marked @beta; fields may change.
resource "cato_app_tenant_restriction_section" "example" {
  # position values: LAST_IN_POLICY | AFTER_SECTION | BEFORE_SECTION
  # Use ref to anchor relative positions; omit ref for LAST_IN_POLICY.
  at = {
    position = "LAST_IN_POLICY"
    # ref = cato_app_tenant_restriction_section.other.section.id
  }

  section = {
    name = "Terraform — Tenant Restriction"
  }
}
