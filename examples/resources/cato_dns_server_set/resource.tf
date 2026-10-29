# This resource owns one named set. It does not replace the account-wide collection.
resource "cato_dns_server_set" "example" {
  name   = "Private resolvers"
  server = ["192.0.2.53", "198.51.100.53"]
}

# The ordered list supports up to two IPv4 resolvers. Use [] to clear it.
# terraform import cato_dns_server_set.example DNS_SET_ID
