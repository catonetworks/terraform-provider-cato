## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Usage ###

### Retrieve all allocated IPs ###
data "cato_allocatedIp" "ips" {}

### Retrieve allocated IPs by name ###
data "cato_allocatedIp" "ips" {
  name_filter = ["11.22.33.44"]
}