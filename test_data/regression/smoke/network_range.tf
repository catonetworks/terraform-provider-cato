resource "cato_network_range" "vlan" {
  site_id         = cato_socket_site.this.id
  interface_index = var.interface_index
  name            = var.vlan_range_name
  range_type      = var.range_type
  subnet          = var.vlan_range_subnet
  local_ip        = var.vlan_range_local_ip
  vlan            = var.vlan_range_vlan
  dhcp_settings = {
    dhcp_type              = var.socket_dhcp_type
    ip_range               = var.vlan_range_ip_range
    dhcp_microsegmentation = var.vlan_range_dhcp_microsegmentation
  }
}
