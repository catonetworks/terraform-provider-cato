## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Resource Usage ###

# Allocate an IP in a PoP location. Cato assigns the IP address --
# the configuration chooses the location, description, and type. The
# resulting ip and display name ("<pop> - <ip>") are computed.
resource "cato_allocated_ip" "portland" {
  pop_location_name = "Portland"
  description       = "AWS Bedrock Frontend"
}

# The description is the stable cross-tenant identity of an allocation
# (keep it unique per PoP location); policy rules reference the computed
# display name:
output "portland_allocation_name" {
  value = cato_allocated_ip.portland.name # e.g. "Portland - 199.27.34.88"
}
