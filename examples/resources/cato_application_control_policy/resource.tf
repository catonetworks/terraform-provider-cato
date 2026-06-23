# Account-level Application Control policy toggles (App & Data Inline Protection).
# GraphQL for this policy is marked @beta; fields may change.
#
# NOTE: Only one Application Control policy exists per account. Managing this
# resource from multiple Terraform configurations (or multiple workspaces) will
# cause conflicts — all configurations must agree on the same attribute values.
resource "cato_application_control_policy" "example" {
  enabled              = true
  data_control_enabled = "ENABLED" # ENABLED | DISABLED
}
