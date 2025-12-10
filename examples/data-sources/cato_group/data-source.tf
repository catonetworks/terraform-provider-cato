## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

# Example: Read all group data
data "cato_group" "all_groups" {}

# Example: Read group data using data source (by Name)
data "cato_group" "by_name" {
  name_filter = ["My Group 1", "Some Group 2"]
}

# Example: Read group data using data source (by ID)
data "cato_group" "by_id" {
  id_filter = ["abcde-12345","fghij-67890"]
}