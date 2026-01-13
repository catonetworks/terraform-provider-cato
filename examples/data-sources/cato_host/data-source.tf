## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

### Retrieve all static hosts ###
data "cato_host" "all" {}

### Retrieve static hosts by IP ###
data "cato_host" "by_ips" {
  ip_filter = ["10.20.6.100"]
}

### Retrieve static hosts by name ###
data "cato_host" "by_names" {
  name_filter = ["test-host-100", "test-host-101"]
}

### Retrieve static hosts by name or IP ###
data "cato_host" "by_names_ips" {
  ip_filter = ["10.20.6.100"]
  name_filter = ["test-host-100", "test-host-101"]
}