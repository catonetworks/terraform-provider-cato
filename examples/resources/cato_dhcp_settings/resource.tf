resource "cato_dhcp_settings" "example" {
  # lease_time = 3600
  # relay = {
  #   enabled = true
  #   timeout = 10
  #   group = { id = cato_dhcp_relay_group.example.id }
  #   # Alternatively: group = { name = "Example relay group" }
  # }
}
# Omission preserves current lease/relay configuration.
# Enabled relay needs an effective timeout and group; disable relay to stop using it.
# Destroy removes Terraform ownership and retains account DHCP settings.
