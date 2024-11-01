// network range of type VLAN
resource "cato_network_range" "vlan100" {
  site_id    = cato_socket_site.site1.id
  name       = "VLAN100"
  range_type = "VLAN"
  subnet     = "192.168.100.0/24"
  local_ip   = "192.168.100.100"
  vlan       = "100"
}

// network range of type VLAN with DHCP RANGE
resource "cato_network_range" "vlan200" {
  site_id    = cato_socket_site.site1.id
  name       = "VLAN200"
  range_type = "VLAN"
  subnet     = "192.168.200.0/24"
  local_ip   = "192.168.200.1"
  vlan       = "200"
  dhcp_settings = {
    dhcp_type = "DHCP_RANGE"
    ip_range  = "192.168.200.100-192.168.200.150"
  }
}

// routed network 
resource "cato_network_range" "routed250" {
  site_id    = cato_socket_site.site1.id
  name       = "routed250"
  range_type = "Routed"
  subnet     = "192.168.250.0/24"
  gateway   = "192.168.25.1"
}