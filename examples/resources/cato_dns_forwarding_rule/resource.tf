

// forward a single domain
resource "cato_dns_forwarding_rule" "example" {
  domain = "example.com"
  server = [
    "192.0.2.53",
  ]
}

# cato_dns_forwarding_rule.kitchen_sink:
# domain and server are plain literals with no id/name reference to another tenant
# object, so a single maximal instance (6 servers, the schema's upper bound) is the
# complete kitchen sink - no by-name/by-id duality applies to this type.
resource "cato_dns_forwarding_rule" "kitchen_sink" {
  domain = "kitchensink.example.com"
  server = [
    "203.0.113.53",
    "203.0.113.54",
    "203.0.113.55",
    "203.0.113.56",
    "203.0.113.57",
    "203.0.113.58",
  ]
}
