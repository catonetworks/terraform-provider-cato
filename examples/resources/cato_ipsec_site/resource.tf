// Data Source for site location
data "cato_siteLocation" "ny" {
  filters = [{
    field = "city"
    search = "New York"
    operation = "startsWith"
  },
  {
    field = "state_name"
    search = "New York"
    operation = "exact"
  },
 {
    field = "country_name"
    search = "United"
    operation = "contains"
  }]
}

// ipsec site and tunnels
resource "cato_ipsec_site" "test-dev-site" {
  name = "ipsec-dev-site"
  site_type = "CLOUD_DC"
  description = "IPSec Dev site"
  native_network_range = "172.98.10.0/24"
  site_location = {
    country_code = data.cato_siteLocation.ny.locations[1].country_code
    state_code = data.cato_siteLocation.ny.locations[1].state_code
    timezone = data.cato_siteLocation.ny.locations[1].timezone[0]
    address = "555 That Way"
  }
  ipsec = {
    primary = {
      public_cato_ip_id = "30111"
      tunnels = [
        {
          public_site_ip = "1.1.1.1"
          private_cato_ip = "172.98.10.51"
          private_site_ip = "172.98.10.61"
          psk = "abcde12345"
          last_mile_bw = {
            downstream = 10
            upstream = 10
          }
        }
      ]
    }
    secondary = {
      public_cato_ip_id = "30112"
      tunnels = [
        {
          public_site_ip = "1.1.1.2"
          private_cato_ip = "172.98.10.91"
          private_site_ip = "172.98.10.11"
          psk = "abcde12345abcde12345"
          last_mile_bw = {
            downstream = 10
            upstream = 10
          }
        }
      ]
    }    
  }  
}