// static host (parent site referenced by id)
resource "cato_static_host" "host1" {
  site_id = cato_socket_site.site1.id
  name    = "host"
  ip      = "192.168.0.1"
}

// static host with DHCP reservation based on mac_address
resource "cato_static_host" "host2" {
  site_id     = cato_socket_site.site1.id
  name        = "host"
  ip          = "192.168.0.2"
  mac_address = "00:00:00:00:00:02"
}

// static host whose parent site is referenced by NAME instead of id.
// Exactly one of site_id / site_name may be set; the provider resolves the
// name to the site id via entityLookup and stores it in state. Useful for
// tenant-to-tenant replication where site ids differ but names are stable.
resource "cato_static_host" "host3" {
  site_name = "My Branch Site"
  name      = "host"
  ip        = "192.168.0.3"
}