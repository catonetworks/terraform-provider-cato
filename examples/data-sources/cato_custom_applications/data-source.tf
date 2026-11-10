## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all custom applications
data "cato_custom_applications" "all" {}

# Example: Read custom applications by name
data "cato_custom_applications" "by_name" {
  name_filter = ["Internal CRM", "Legacy ERP"]
}

# Example: Read custom applications by ID
data "cato_custom_applications" "by_id" {
  id_filter = ["12345", "67890"]
}
