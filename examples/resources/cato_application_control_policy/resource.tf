# Account-level Application Control policy toggles (App & Data Inline Protection).
# GraphQL for this policy is marked @beta; fields may change.
resource "cato_application_control_policy" "app_ctrl" {
  enabled              = true
  data_control_enabled = "ENABLED"
}
