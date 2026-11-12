## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all private applications
data "cato_private_apps" "all" {}

# Example: Read private applications by name
data "cato_private_apps" "by_name" {
  name_filter = ["WIKI", "Demo Portal"]
}

# Example: Read private applications by ID
data "cato_private_apps" "by_id" {
  id_filter = ["12345", "67890"]
}
