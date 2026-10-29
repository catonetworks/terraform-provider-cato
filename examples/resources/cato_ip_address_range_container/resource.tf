resource "cato_ip_address_range_container" "example" {
  name        = "example_ip_address_range_container"
  description = "Managed IPv4 range members"
  mode        = "INLINE"
  values      = [{from="192.0.2.1",to="192.0.2.10"},{from="198.51.100.1",to="198.51.100.1"}]

  # URL mode alternative: set mode = "URL" and omit values above.
  # URL mode manages synchronization settings; fetched members are not Terraform-owned.
  # file_type = "CSV" # CSV or STIX
  # sync_data = {
  #   url           = "https://example.com/members.csv"
  #   time_interval = 1
  #   time_unit     = "DAY"
  #   notifications = {
  #     mailing_list       = []
  #     subscription_group = [{ name = "existing-subscription-group" }] # or { id = "existing-id" }
  #     webhook            = []
  #   }
  # }
}
