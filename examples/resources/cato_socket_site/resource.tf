// socket site for AWS
data "cato_siteLocation" "ny" {
  filters = [{
    field = "city"
    search = "New York"
    operation = "startsWith"
  },
  {
    field = "state_name"
    search = "New York"
    operation = "exact"
  },
 {
    field = "country_name"
    search = "United"
    operation = "contains"
  }]
}

resource "cato_socket_site" "aws_site" {
  name            = "aws_site"
  description     = "site description"
  site_type       = "DATACENTER"
  connection_type = "SOCKET_AWS1500"

  native_range = {
    native_network_range = "192.168.25.0/24"
    local_ip             = "192.168.25.5"
  }

  site_location = {
    country_code = data.cato_siteLocation.ny.locations[1].country_code
    state_code = data.cato_siteLocation.ny.locations[1].state_code
    timezone = data.cato_siteLocation.ny.locations[1].timezone
 }
}

// socket site x1500 with DHCP settings
resource "cato_socket_site" "branch_site" {
  name            = "branch_site"
  description     = "site description"
  site_type       = "BRANCH"
  connection_type = "SOCKET_X1500"

  native_range = {
    native_network_range = "192.168.20.0/24"
    local_ip             = "192.168.20.1"
    dhcp_settings = {
      dhcp_type = "DHCP_RANGE"
      ip_range  = "192.168.20.10-192.168.20.22"
    }
  }

  site_location = {
    country_code = "FR"
    timezone     = "Europe/Paris"
  }
}