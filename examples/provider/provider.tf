# Configuration based authentication
terraform {
  required_providers {
    cato-oss = {
      source = "registry.terraform.io/benekpy/cato-oss"
      version = "~> 0.3.0"
    }
  }
}

provider "cato-oss" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = "xxxxxxx"
    account_id = "xxxxxxx"
}

resource "cato-oss_socket_site" "site1" {
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

resource "cato-oss_static_host" "host" {
    site_id = cato-oss_socket_site.site1.id
    name = "test-terraform"
    ip = "192.168.25.24"
}