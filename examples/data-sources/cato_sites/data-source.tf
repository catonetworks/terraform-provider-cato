## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all sites
data "cato_sites" "all" {}

# Example: Read sites by name
data "cato_sites" "by_name" {
  name_filter = ["Datacenter-NYC", "Branch-LON"]
}
