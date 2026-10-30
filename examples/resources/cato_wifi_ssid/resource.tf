# Requires an unused Wi-Fi interface on a compatible Socket site.
# The API currently supports GUEST SSIDs only.
resource "cato_wifi_ssid" "example" {
  name                 = "Guest Wi-Fi"
  site = {
    id = "12345"
    # name = "Branch" # Use name instead of id if preferred.
  }
  network_interface_id = "WBR1"
  band                 = "BOTH"
  category             = "GUEST"
  subnet               = "10.20.3.0/24"
  local_ip             = "10.20.3.1"
  mdns_enabled         = false
  microsegmentation_enabled = false
  # enabled       = true
  # visible       = true
  # internet_only = false # Enabling requires the account/socket feature.
  # dhcp_enabled  = false # Deprecated API flag; only false is supported.
  # dhcp = {
  #   dhcp_subnet = "10.20.3.128/25"
  # }
  security = {
    mode          = "PSK"
    auth_protocol = "WPA3"
    psk = {
      passkey = var.wifi_passkey
    }
    # track_authentication = false
  }
}

variable "wifi_passkey" {
  description = "Wi-Fi PSK, 8 to 63 characters without whitespace. Retained in Terraform state because the API cannot return it."
  type        = string
  sensitive   = true
}
