# Registers Cato site metadata; physical connections use their own resource.
resource "cato_cloud_interconnect_site" "example" {
  name      = "Development interconnect"
  site_type = "BRANCH"
  # description = "Terraform managed Cloud Interconnect site"
  site_location = {
    country_code = "GB"
    timezone     = "Europe/London"
    # city       = "London"
    # state_code = ""
    # address    = "Site address"
  }
}

# Remove separately managed physical connections before destroying this site.
# terraform import cato_cloud_interconnect_site.example 12345
