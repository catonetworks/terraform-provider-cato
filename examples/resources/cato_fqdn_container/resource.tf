resource "cato_fqdn_container" "example" {
  name        = "example_fqdn_container"
  description = "Managed FQDN members"
  mode        = "INLINE"
  values      = ["example.com", "api.example.com"]

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
