# Enable or disable the policy.
# (see cato_private_access_rule for details about policy rules)
resource "cato_private_access_policy" "example" {
  enabled = true
}
