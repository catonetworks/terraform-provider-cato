# cato_bgp_peer: BGP peering on a site. Below is a full "kitchen sink" exercising
# every attribute, followed by a minimal peer.

resource "cato_bgp_peer" "kitchen_sink" {
  site_id        = "12345"
  name           = "vWAN BGP connection"
  cato_asn       = 65000
  peer_asn       = 65100
  peer_ip        = "10.0.0.1"
  default_action = "ACCEPT" // ACCEPT | REJECT

  # Route advertisement controls.
  advertise_all_routes     = false
  advertise_default_route  = false
  advertise_summary_routes = true

  # Session tuning.
  metric             = 150
  hold_time          = 30
  keepalive_interval = 10
  perform_nat        = false
  md5_auth_key       = "REPLACE_ME" // pull from a sensitive variable in real configs

  # Summary routes advertised to the peer, with optional BGP communities.
  summary_route = [
    {
      route = "1.1.1.0/24"
      community = [
        { from = 65005, to = 65010 },
      ]
    },
    {
      route = "1.1.2.0/24"
      community = [
        { from = 65015, to = 65020 },
      ]
    },
  ]

  # Bidirectional Forwarding Detection.
  bfd_enabled = true
  bfd_settings = {
    transmit_interval = 100
    receive_interval  = 100
    multiplier        = 10
  }

  # Notifications for this peer's state changes.
  tracking = {
    enabled         = true
    alert_frequency = "DAILY"
    # subscription_id = "<subscription id>"
  }
}

# Minimal BGP peer (only the required attributes).
resource "cato_bgp_peer" "minimal" {
  site_id        = "12345"
  name           = "Minimal BGP peer"
  cato_asn       = 65000
  peer_asn       = 65200
  peer_ip        = "10.0.0.2"
  default_action = "ACCEPT"
}
