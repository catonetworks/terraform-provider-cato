resource "cato_dhcp_relay_group" "example" {
  name   = "Example relay group"
  server = ["192.0.2.53", "192.0.2.54"]
}
# ID is opaque. Existing site/range DHCP APIs can resolve the group by name;
# do not use this opaque ID where a legacy numeric entity ID is required.
