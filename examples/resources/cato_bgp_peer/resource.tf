resource "cato_bgp_peer" "bgp_peer" {
  site_id                  = "12345"
  name                     = "vWAN BGP connectionn"
  cato_asn                 = 65000
  peer_asn                 = 65100
  peer_ip                  = "10.0.0.1"
  metric                   = 150
  default_action           = "ACCEPT"
  advertise_summary_routes = true
  advertise_default_route  = false
  summary_route = [
    {
      route = "1.1.1.0/24"
      community = [
        {
          from = 65005
          to   = 65010
        }
      ]
    },
    {
      route = "1.1.2.0/24"
      community = [
        {
          from = 65015
          to   = 65020
        }
      ]
    }
  ]
  bfd_settings = {
    transmit_interval = 100
    receive_interval  = 100
    multiplier        = 10
  }
}