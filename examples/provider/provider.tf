# Configuration based authentication
terraform {
  required_providers {
    cato = {
      source = "catonetworks/cato"
    }
  }
}

provider "cato" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = "xxxxxxx"
    account_id = "xxxxxxx"
}

resource "cato_socket_site" "site1" {
    name = "site1"
    description = "site1 AWS Datacenter"
    site_type = "DATACENTER"
    connection_type = "SOCKET_AWS1500"
    native_network_range = "192.168.25.0/24"
    local_ip = "192.168.25.100"
    site_location = {
        country_code = "FR",
        timezone = "Europe/Paris"
    }
}

resource "cato_static_host" "host" {
    site_id = cato_socket_site.site1.id
    name = "test-terraform"
    ip = "192.168.25.24"
}