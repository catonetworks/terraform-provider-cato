## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all IP address range containers
data "cato_ip_address_range_container" "all" {}

# Example: Read IP address range containers by name
data "cato_ip_address_range_container" "by_name" {
  name_filter = ["Corp-Subnets", "Partner-Subnets"]
}

# Example: Read IP address range containers by ID
data "cato_ip_address_range_container" "by_id" {
  id_filter = ["12345", "67890"]
}
