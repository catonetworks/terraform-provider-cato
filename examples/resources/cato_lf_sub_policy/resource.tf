# Omitted site, source, and destination reference sets default to [] within the required scope objects.
// LAN Firewall sub-policy

resource "cato_lf_sub_policy" "example" {
  name        = "example"
  description = "example sub-policy"

  at = {
    position = "FIRST_IN_POLICY"
  }
  # Optional: manage the API-created cleanup rule without a separate resource.
  # cleanup_rule = {
  #   name        = "example - Cleanup Rule"
  #   description = "Managed cleanup rule"
  #   enabled     = true
  # }
  scope = {
    source = {
      subnet = ["10.1.0.0/16"]
    }
    destination = {}
    site        = {}
    direction   = "TO"
    enabled     = true
  }
}
