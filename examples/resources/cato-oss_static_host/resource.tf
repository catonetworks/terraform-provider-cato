// static host 
resource "cato_static_host" "host1" {
  site_id     = cato_socket_site.site1.id
  name        = "host"
  ip          = "192.168.0.1"
}

// static host with DHCP reservation based on mac_address
resource "cato_static_host" "host2" {
  site_id     = cato_socket_site.site1.id
  name        = "host"
  ip          = "192.168.0.2"
  mac_address = "00:00:00:00:00:02"
}