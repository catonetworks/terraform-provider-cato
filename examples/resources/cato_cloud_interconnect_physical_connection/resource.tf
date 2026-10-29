# The parent must be a Cloud Interconnect site with an unused HA slot.
resource "cato_cloud_interconnect_physical_connection" "example" {
  site = {
    id = "12345"
    # name = "Cloud Interconnect branch" # Use name instead of id if preferred.
  }
  ha_role = "PRIMARY"
  pop_location = {
    id = "23456"
    # name = "Interconnect PoP" # Use name instead of id if preferred.
  }
  service_provider_name = "Supported service provider"
  encapsulation_method  = "DOT1Q"
  subnet                 = "10.20.4.0/30"
  private_site_ip        = "10.20.4.1"
  private_cato_ip        = "10.20.4.2"
  upstream_bw_limit      = "100"
  downstream_bw_limit    = "100"

  # Omit tags to use API allocation; assigned active tags are returned in state.
  # Custom tags require SUPPORT_INTERCONNECT_VLAN_CONFIGURATION.
  # vlan = 123 # DOT1Q only.
  # For QINQ, set encapsulation_method = "QINQ", omit vlan, and optionally set:
  # qinq_vlan_configuration = {
  #   c_vlan = 234
  #   s_vlan = 345
  # }
}
