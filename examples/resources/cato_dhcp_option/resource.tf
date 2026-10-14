resource "cato_dhcp_option" "example" {
  tag   = 224
  type  = "ASCII"
  value = "example"
  # description = "Example DHCP option"
}
