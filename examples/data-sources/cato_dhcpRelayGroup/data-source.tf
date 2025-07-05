## Providers ###
provider "cato" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = var.cato_token
    account_id = var.account_id
}

### Data Source Usage ###

### Retrieve all allocated IPs ###
data "cato_dhcpRelayGroup" "all" {}

### Retrieve allocated IPs by name ###
data "cato_dhcpRelayGroup" "my_dhcp_group" {
	name_filter = ["my_dhcp_group"]
}