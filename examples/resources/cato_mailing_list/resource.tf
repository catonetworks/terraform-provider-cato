# Owns the entire mailing list membership; creating it sends no email.
resource "cato_mailing_list" "example" {
  name    = "Operations"
  address = ["operations@example.com"]
  # Admin references require IDs. Names are returned for display only.
  # admin = [{ id = "123" }]
}

# At least one address or administrator must remain.
# terraform import cato_mailing_list.example 12345
