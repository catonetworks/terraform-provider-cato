## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source - All ranges ###
data "cato_networkRanges" "range" {}

### Data Source - Filter by site or name ###
data "cato_networkRanges" "range" {
  site_id_filter = ["12345"]
  name_filter    = ["VLAN", "Native Range"]
}