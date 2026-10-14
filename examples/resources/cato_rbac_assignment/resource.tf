# Use a role with only the permissions required by this assignment.
resource "cato_role" "site_viewer" {
  name = "Terraform site viewer"
  permissions = [{ resource = "Sites", action = "VIEW" }]
}

resource "cato_rbac_assignment" "example" {
  user_id = "12345" # Existing user ID. Replace with your account's user ID.
  # group_id = "67890" # Use an existing users group instead of user_id.
  # account_id = "123456" # Defaults to the provider account_id.

  role_ids = [cato_role.site_viewer.role_id]
  allowed_entities = [
    { id = "54321", type = "SITE", action = "VIEW" }
    # Use SYSTEMGROUP entity IDs for account-wide scope.
    # Other supported types: USERSGROUP, GROUP, SYSTEMGROUP,
    # WAN_FIREWALL_POLICY, INTERNET_FIREWALL_POLICY.
  ]
  # Use allowed_entities = [] for no individual entities.
  # allow_viewing_unobfuscated_data = false
  # description = "View the selected site"
}
