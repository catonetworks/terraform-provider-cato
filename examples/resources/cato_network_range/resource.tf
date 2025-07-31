// network range of type Direct
resource "cato_network_range" "direct" {
  site_id      = "142382"
  interface_id = "172922"
  name         = "Direct Network Range"
  range_type   = "Direct"
  subnet       = "192.166.100.0/24"
  local_ip     = "192.166.100.1"
  # internet_only = true 
  mdns_reflector    = true
  translated_subnet = "172.166.100.0/24"
}

resource "cato_network_range" "routed" {
  site_id       = "142382"
  interface_id  = "172922"
  name          = "Routed Network Range"
  range_type    = "Routed"
  subnet        = "192.167.100.0/24"
  gateway       = "10.0.0.3"
  internet_only = true
  # mdns_reflector = true
  translated_subnet = "172.167.100.0/24"
}

// network range of type VLAN with DHCP RANGE
resource "cato_network_range" "vlan200" {
  site_id       = cato_socket_site.site1.id
  name          = "VLAN200"
  range_type    = "VLAN"
  subnet        = "192.168.200.0/24"
  local_ip      = "192.168.200.1"
  vlan          = "200"
  internet_only = true
  # mdns_reflector = true
  dhcp_settings = {
    dhcp_type              = "DHCP_RANGE"
    ip_range               = "192.168.200.100-192.168.200.150"
    dhcp_microsegmentation = false
  }
}

// network range of type VLAN with DHCP RELAY
resource "cato_network_range" "vlan201_relay" {
  site_id    = cato_socket_site.site1.id
  name       = "VLAN201"
  range_type = "VLAN"
  subnet     = "192.168.200.0/24"
  local_ip   = "192.168.200.1"
  vlan       = "200"
  # internet_only = true
  mdns_reflector = true
  dhcp_settings = {
    dhcp_type              = "DHCP_RELAY"
    relay_group_name       = "my_dhcp_relay"
    dhcp_microsegmentation = true
  }
}

