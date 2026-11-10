## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all FQDN containers
data "cato_fqdn_container" "all" {}

# Example: Read FQDN containers by name
data "cato_fqdn_container" "by_name" {
  name_filter = ["Allowed-SaaS", "Blocked-Domains"]
}

# Example: Read FQDN containers by ID
data "cato_fqdn_container" "by_id" {
  id_filter = ["12345", "67890"]
}
