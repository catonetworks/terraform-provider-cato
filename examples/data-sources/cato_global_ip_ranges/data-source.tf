## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all global IP ranges
data "cato_global_ip_ranges" "all" {}

# Example: Read global IP ranges by name
data "cato_global_ip_ranges" "by_name" {
  name_filter = ["Corp-DMZ", "Guest-WiFi"]
}

# Example: Read global IP ranges by ID
data "cato_global_ip_ranges" "by_id" {
  id_filter = ["12345", "67890"]
}
