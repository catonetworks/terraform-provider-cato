resource "cato_role" "viewer" {
  name = "Terraform site viewer"
  # description = "Read access to sites"
  permissions = [{ resource = "Sites", action = "VIEW" }]
  # account_id = "123456" # Defaults to the provider account.
}
