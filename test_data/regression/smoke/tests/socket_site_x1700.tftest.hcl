# Create SOCKET_X1700
run "create_socket_x1700" {
  command = apply

  variables {
    socket_name                 = "TF-Test-Socket-X1700-Site"
    description                 = "Test Socket X1700 Site created by Terraform"
    site_type                   = "BRANCH"
    connection_type             = "SOCKET_X1700"
    socket_native_network_range = "248.248.100.0/24"
    socket_local_ip             = "248.248.100.1"
    socket_dhcp_type            = "DHCP_RANGE"
    socket_ip_range             = "248.248.100.50-248.248.100.100"
    socket_country_code         = "FR"
    socket_timezone             = "Europe/Paris"

    interface_index                   = "INT_3"
    vlan_range_name                   = "TF-Test-Socket-X1700-Site_VLAN_Range"
    range_type                        = "VLAN"
    vlan_range_subnet                 = "248.249.100.0/24"
    vlan_range_local_ip               = "248.249.100.2"
    vlan_range_vlan                   = 202
    vlan_range_ip_range               = "248.249.100.20-248.249.100.32"
    vlan_range_dhcp_microsegmentation = true
  }

  # Input field assertions
  assert {
    condition     = cato_socket_site.this.name == var.socket_name
    error_message = "Site name does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.description == var.description
    error_message = "Site description does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.site_type == var.site_type
    error_message = "Site type does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.connection_type == var.connection_type
    error_message = "Connection type does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.native_range.native_network_range == var.socket_native_network_range
    error_message = "Native network range does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.native_range.local_ip == var.socket_local_ip
    error_message = "Native range local IP does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.native_range.dhcp_settings.dhcp_type == var.socket_dhcp_type
    error_message = "DHCP type does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.native_range.dhcp_settings.ip_range == var.socket_ip_range
    error_message = "DHCP IP range does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.site_location.country_code == var.socket_country_code
    error_message = "Site country code does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.site_location.timezone == var.socket_timezone
    error_message = "Site timezone does not match expected value"
  }

  # Computed field assertions
  assert {
    condition     = cato_socket_site.this.id != null && cato_socket_site.this.id != ""
    error_message = "Site ID should be set after apply"
  }

  # VLAN network range input field assertions
  assert {
    condition     = cato_network_range.vlan.site_id == cato_socket_site.this.id
    error_message = "VLAN network range site ID does not match socket site ID"
  }

  assert {
    condition     = cato_network_range.vlan.interface_index == var.interface_index
    error_message = "VLAN network range interface index does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.name == var.vlan_range_name
    error_message = "VLAN network range name does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.range_type == var.range_type
    error_message = "VLAN network range type does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.subnet == var.vlan_range_subnet
    error_message = "VLAN network range subnet does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.local_ip == var.vlan_range_local_ip
    error_message = "VLAN network range local IP does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.vlan == var.vlan_range_vlan
    error_message = "VLAN network range VLAN ID does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.dhcp_settings.dhcp_type == var.socket_dhcp_type
    error_message = "VLAN network range DHCP type does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.dhcp_settings.ip_range == var.vlan_range_ip_range
    error_message = "VLAN network range DHCP IP range does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.dhcp_settings.dhcp_microsegmentation == var.vlan_range_dhcp_microsegmentation
    error_message = "VLAN network range DHCP microsegmentation does not match expected value"
  }

  # VLAN network range computed field assertions
  assert {
    condition     = cato_network_range.vlan.id != null && cato_network_range.vlan.id != ""
    error_message = "VLAN network range ID should be set after apply"
  }
}

# Update SOCKET_X1700
run "update_socket_x1700" {
  command = apply

  variables {
    socket_name                 = "TF-Test-Socket-X1700-Site-2"
    description                 = "Test Socket X1700 Site created by Terraform-2"
    site_type                   = "BRANCH"
    connection_type             = "SOCKET_X1700"
    socket_native_network_range = "249.248.100.0/24"
    socket_local_ip             = "249.248.100.1"
    socket_dhcp_type            = "DHCP_RANGE"
    socket_ip_range             = "249.248.100.50-249.248.100.100"
    socket_country_code         = "CZ"
    socket_timezone             = "Europe/Prague"

    interface_index                   = "INT_3"
    vlan_range_name                   = "TF-Test-Socket-X1700-Site_VLAN_Range-2"
    range_type                        = "VLAN"
    vlan_range_subnet                 = "249.249.100.0/24"
    vlan_range_local_ip               = "249.249.100.2"
    vlan_range_vlan                   = 203
    vlan_range_ip_range               = "249.249.100.20-249.249.100.32"
    vlan_range_dhcp_microsegmentation = true
  }

  # Input field assertions
  assert {
    condition     = cato_socket_site.this.name == var.socket_name
    error_message = "Site name does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.description == var.description
    error_message = "Site description does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.site_type == var.site_type
    error_message = "Site type does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.connection_type == var.connection_type
    error_message = "Connection type does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.native_range.native_network_range == var.socket_native_network_range
    error_message = "Native network range does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.native_range.local_ip == var.socket_local_ip
    error_message = "Native range local IP does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.native_range.dhcp_settings.dhcp_type == var.socket_dhcp_type
    error_message = "DHCP type does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.native_range.dhcp_settings.ip_range == var.socket_ip_range
    error_message = "DHCP IP range does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.site_location.country_code == var.socket_country_code
    error_message = "Site country code does not match expected value"
  }

  assert {
    condition     = cato_socket_site.this.site_location.timezone == var.socket_timezone
    error_message = "Site timezone does not match expected value"
  }

  # Computed field assertions
  assert {
    condition     = cato_socket_site.this.id != null && cato_socket_site.this.id != ""
    error_message = "Site ID should be set after apply"
  }

  # VLAN network range input field assertions
  assert {
    condition     = cato_network_range.vlan.site_id == cato_socket_site.this.id
    error_message = "VLAN network range site ID does not match socket site ID"
  }

  assert {
    condition     = cato_network_range.vlan.interface_index == var.interface_index
    error_message = "VLAN network range interface index does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.name == var.vlan_range_name
    error_message = "VLAN network range name does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.range_type == var.range_type
    error_message = "VLAN network range type does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.subnet == var.vlan_range_subnet
    error_message = "VLAN network range subnet does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.local_ip == var.vlan_range_local_ip
    error_message = "VLAN network range local IP does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.vlan == var.vlan_range_vlan
    error_message = "VLAN network range VLAN ID does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.dhcp_settings.dhcp_type == var.socket_dhcp_type
    error_message = "VLAN network range DHCP type does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.dhcp_settings.ip_range == var.vlan_range_ip_range
    error_message = "VLAN network range DHCP IP range does not match expected value"
  }

  assert {
    condition     = cato_network_range.vlan.dhcp_settings.dhcp_microsegmentation == var.vlan_range_dhcp_microsegmentation
    error_message = "VLAN network range DHCP microsegmentation does not match expected value"
  }

  # VLAN network range computed field assertions
  assert {
    condition     = cato_network_range.vlan.id != null && cato_network_range.vlan.id != ""
    error_message = "VLAN network range ID should be set after apply"
  }
}
