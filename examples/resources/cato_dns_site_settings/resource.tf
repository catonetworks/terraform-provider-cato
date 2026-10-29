resource "cato_dns_site_settings" "example" {
  site = { id = var.site_id } # Alternatively: { name = "Branch" }
  # primary_server = "192.0.2.53"
  # secondary_server = "192.0.2.54"
  # suffix           = ["example.test"]
}
# Destroy removes this site's override and restores account inheritance.
