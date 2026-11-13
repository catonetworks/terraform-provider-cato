## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all mailing lists
data "cato_mailing_lists" "all" {}

# Example: Read mailing lists by name
data "cato_mailing_lists" "by_name" {
  name_filter = ["All Admins", "Security Team"]
}

# Example: Read mailing lists by ID
data "cato_mailing_lists" "by_id" {
  id_filter = ["12345", "67890"]
}
