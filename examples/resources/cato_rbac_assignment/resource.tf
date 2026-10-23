resource "cato_role" "viewer" {
  name        = "Terraform site viewer"
  permissions = [{ resource = "Sites", action = "VIEW" }]
}

resource "cato_rbac_assignment" "viewer" {
  user_id  = "12345" # Existing admin-user principal.
  role_ids = [cato_role.viewer.role_id]
  allowed_entities = [{
    id     = "67890"
    type   = "SITE"
    action = "VIEW"
  }]
  # group_id = "12345" # Use instead of user_id for a users-group principal.
  # account_id = "123456" # Defaults to the provider account.
  # description = "Site viewer assignment"
  # allow_viewing_unobfuscated_data = false
}
