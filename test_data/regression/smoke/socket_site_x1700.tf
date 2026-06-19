resource "cato_socket_site" "this" {
  name            = var.socket_name
  description     = var.description
  site_type       = var.site_type
  connection_type = var.connection_type

  native_range = {
    native_network_range = var.socket_native_network_range
    local_ip             = var.socket_local_ip
    dhcp_settings = {
      dhcp_type = var.socket_dhcp_type
      ip_range  = var.socket_ip_range
    }
  }

  site_location = {
    country_code = var.socket_country_code
    timezone     = var.socket_timezone
  }
}
