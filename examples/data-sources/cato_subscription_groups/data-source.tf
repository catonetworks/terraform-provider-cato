## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all subscription groups
data "cato_subscription_groups" "all" {}

# Example: Read subscription groups by name
data "cato_subscription_groups" "by_name" {
  name_filter = ["NOC Alerts", "Security Notifications"]
}

# Example: Read subscription groups by ID
data "cato_subscription_groups" "by_id" {
  id_filter = ["12345", "67890"]
}
