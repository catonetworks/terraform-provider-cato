# Owns one named set, preserving the search order.
resource "cato_dns_suffix_set" "example" {
  name   = "Private search suffixes"
  suffix = ["engineering.example.com", "example.com"]
}

# Use [] to clear the list. Each suffix must be unique and contain no whitespace.
# terraform import cato_dns_suffix_set.example DNS_SET_ID
