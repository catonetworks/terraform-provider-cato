resource "cato_private_app" "example" {
  name                 = "private-app-example"
  description          = "private-app-example description"
  internal_app_address = "172.198.30.10"
  protocol_ports = [
    { ports = [80], protocol = "TCP" },
    { ports = [443], protocol = "TCP" },
    { port_range = { from = 6000, to = 6010 }, protocol = "TCP" }
  ]
  allow_icmp_protocol = true
  probing_enabled     = true
  private_app_probing = {
    type                 = "ICMP_PING"
    interval             = 5
    fault_threshold_down = 10
  }
  published = true
  published_app_domain = {
    connector_group_name = "conn-group-1"
    published_app_domain = "private-app-example.example.com"
  }
}
