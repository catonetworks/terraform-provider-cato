
# Socket LAN Section - Creates a section in the Socket LAN policy to organize rules
resource "cato_socket_lan_section" "example" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "My Socket LAN Section"
  }
}

# Socket LAN Section - Insert at first position
resource "cato_socket_lan_section" "first_section" {
  at = {
    position = "FIRST_IN_POLICY"
  }
  section = {
    name = "Priority Rules Section"
  }
}
