## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all users
data "cato_users" "all" {}

# Example: Read users by display name
data "cato_users" "by_name" {
  name_filter = ["Jane Doe", "John Smith"]
}

# Example: Read users by email
data "cato_users" "by_email" {
  email_filter = ["jane.doe@example.com", "john.smith@example.com"]
}

# Example: Read only locally-provisioned (MANUAL) users
# import_type_filter accepts one of: INTEGRATION, LDAP, MANUAL, SCIM
data "cato_users" "manual_only" {
  import_type_filter = "MANUAL"
}
