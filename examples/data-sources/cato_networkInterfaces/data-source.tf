## Providers ###
provider "cato" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = var.cato_token
    account_id = var.account_id
}

### Data Source ###
data "cato_networkInterfaces" "test-site" {
	site_id = "12345"
    network_interface_name = "INT_7"
}

### Example Response Output ###
data "cato_networkInterfaces" "test-site" {
    items   = [
        {
            dest_type              = "LAN"
            id                     = "12345"
            name                   = "MY-LAN"
            site_id                = "89328"
            site_name              = "SD-Office"
            socket_interface_id    = "INT_7"
            socket_interface_index = "6"
            subnet                 = "192.168.2.0/24"
        },
        {
            dest_type              = "LAN"
            id                     = "12346"
            name                   = "MY-WAN"
            site_id                = "09876"
            site_name              = "MY-Office"
            socket_interface_id    = "INT_1"
            socket_interface_index = "5"
            subnet                 = "192.168.1.0/24"
        },
    ]
    site_id = "89328"
}
