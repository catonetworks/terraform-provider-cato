resource "cato_dns_forwarding_rule" "example" {
  domain = "example.test"
  server = ["192.0.2.53", "192.0.2.54"]
}
