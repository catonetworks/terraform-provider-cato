## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all ZTNA app connectors
data "cato_app_connectors" "all" {}

# Example: Read app connectors by name
data "cato_app_connectors" "by_name" {
  name_filter = ["kvm-connector-01"]
}

# Example: Read app connectors by connector group
data "cato_app_connectors" "by_group" {
  group_filter = ["getsase-apps"]
}
