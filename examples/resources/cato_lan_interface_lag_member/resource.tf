// lan interface active
resource "cato_lan_interface" "lag1" {
  dest_type     = "LAN_LAG_MASTER"
  lag_min_links = 2
  interface_id  = "INT_5"
  local_ip      = "10.12.200.6"
  name          = "MyLag5"
  site_id       = var.site_id
  subnet        = "10.12.200.0/25"
}

resource "cato_lan_interface_lag_member" "lag2" {
  depends_on   = [cato_lan_interface.lag1]
  dest_type    = "LAN_LAG_MEMBER"
  interface_id = "INT_6"
  name         = "LagMember6"
  site_id      = var.site_id
}

resource "cato_lan_interface_lag_member" "lag3" {
  depends_on   = [cato_lan_interface.lag1]
  dest_type    = "LAN_LAG_MEMBER"
  interface_id = "INT_7"
  name         = "LagMember7"
  site_id      = var.site_id
}
