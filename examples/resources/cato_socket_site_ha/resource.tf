resource "cato_socket_site_ha" "cloud_router" {
  # The site must already have its secondary GCP vSocket registered.
  site_id        = "12345"
  is_cloud_router = true
}
