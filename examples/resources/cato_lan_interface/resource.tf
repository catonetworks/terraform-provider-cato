// lan interface active
resource "cato_lan_interface" "lan6" {
    dest_type    = "LAN"
    interface_id = "INT_6"
    local_ip     = "192.168.198.6"
    name         = "Interface lan 6"
    site_id      = "12345"
    subnet       = "192.168.198.0/25"
}