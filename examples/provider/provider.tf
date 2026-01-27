# Configuration based authentication
terraform {
  required_providers {
    cato = {
      source  = "catonetworks/cato"
    }
  }
}

provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = "xxxxxxx"
  account_id = "xxxxxxx"
}

resource "cato_socket_site" "site1" {
  name            = "SOCKET Site - X1700"
  description     = "SOCKET Site - X1700"
  site_type       = "DATACENTER"
  connection_type = "SOCKET_X1700"

  native_range = {
    native_network_range = "192.168.37.0/24"
    local_ip             = "192.168.37.5"
    interface_dest_type = "LAN"
    dhcp_settings = {
      dhcp_type = "DHCP_RANGE"
      ip_range  = "192.168.37.100-192.168.37.150"
    }
  }

  site_location = {
    city         = "New York City"
    country_code = "US"
    state_code   = "US-NY"
    timezone     = "America/New_York"
    address      = "555 That Way"
  }
}

resource "cato_static_host" "host" {
  site_id = cato_socket_site.site1.id
  name    = "test-terraform"
  ip      = "192.168.25.24"
}