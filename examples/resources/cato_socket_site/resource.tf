// Data Source for site location
data "cato_siteLocation" "ny" {
  filters = [{
    field     = "city"
    search    = "New York City"
    operation = "startsWith"
    },
    {
      field     = "state_name"
      search    = "New York"
      operation = "exact"
    },
    {
      field     = "country_name"
      search    = "United States"
      operation = "contains"
  }]
}

// socket site for AWS
resource "cato_socket_site" "aws_site" {
  name            = "aws_site"
  description     = "site description"
  site_type       = "DATACENTER"
  connection_type = "SOCKET_AWS1500"

  # Optional. Assign both AWS vSocket serial numbers during site creation.
  # secondary_vsocket = {
  #   eni_ip_address = "192.168.26.10"
  #   eni_ip_subnet  = "192.168.26.0/24"
  #   route_table_id = aws_route_table.lan.id
  # }

  native_range = {
    native_network_range = "192.168.25.0/24"
    local_ip             = "192.168.25.5"
  }

  site_location = {
    city         = data.cato_siteLocation.ny.locations[0].city
    country_code = data.cato_siteLocation.ny.locations[0].country_code
    state_code   = data.cato_siteLocation.ny.locations[0].state_code
    timezone     = data.cato_siteLocation.ny.locations[0].timezone[0]
    address      = "555 That Way"
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
// GCP HA site. Use the two serials when creating the cloud VMs.
resource "cato_socket_site" "gcp_ha" {
  name            = "gcp-ha"
  site_type       = "DATACENTER"
  connection_type = "SOCKET_GCP1500"

  native_range = {
    native_network_range = "192.168.30.0/24"
    local_ip             = "192.168.30.5"
  }

  secondary_vsocket = {
    interface_ip     = "192.168.30.6"
    load_balancer_ip = "192.168.30.10"
  }

  # Optional. Enable Cloud Router BGP peering for NCC Router Appliance mode.
  # is_cloud_router = true

  site_location = {
    country_code = "US"
    state_code   = "US-NY"
    timezone     = "America/New_York"
  }
}

output "gcp_ha_serial_numbers" {
  value = {
    primary   = one([for socket in cato_socket_site.gcp_ha.sockets : socket.serial_number if socket.is_primary])
    secondary = one([for socket in cato_socket_site.gcp_ha.sockets : socket.serial_number if !socket.is_primary])
  }
}

# Azure HA uses the same resource with different secondary settings.
# resource "cato_socket_site" "azure_ha" {
#   name            = "azure-ha"
#   site_type       = "DATACENTER"
#   connection_type = "SOCKET_AZ1500"
#
#   native_range = {
#     native_network_range = "192.168.40.0/24"
#     local_ip             = "192.168.40.5"
#   }
#   secondary_vsocket = {
#     interface_ip = "192.168.40.6"
#     floating_ip  = "192.168.40.10"
#   }
#   site_location = {
#     country_code = "GB"
#     timezone     = "Europe/London"
#   }
# }
