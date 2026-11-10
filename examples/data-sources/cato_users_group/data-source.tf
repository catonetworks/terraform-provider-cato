## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all users groups
data "cato_users_group" "all" {}

# Example: Read users groups by name
data "cato_users_group" "by_name" {
  name_filter = ["Engineering", "Contractors"]
}

# Example: Read only SCIM-provisioned groups
# import_type_filter accepts one of: DYNAMIC, INTEGRATION, LDAP, MANUAL, SCIM
data "cato_users_group" "scim_only" {
  import_type_filter = "SCIM"
}
