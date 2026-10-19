resource "cato_dhcp_settings" "example" {
  lease_time    = 3600
  relay_enabled = false
  # relay_timeout  = 10
  # relay_group_id = cato_dhcp_relay_group.example.id
}
# Enabled relay needs a timeout and group; omitted values retain existing config.
# Destroy removes Terraform ownership and retains account DHCP settings.
