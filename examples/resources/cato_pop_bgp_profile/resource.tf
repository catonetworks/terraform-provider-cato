# Creates a profile without attaching it to a BGP session.
resource "cato_pop_bgp_profile" "example" {
  name = "Private communities"
  # description = "Terraform managed BGP communities"
  communities = [
    { from = "65001", to = "100" },
    { from = "65001", to = "200" },
  ]
}

# Community pairs are unordered. Each value is a canonical decimal string, 0..65535.
# The API restricts community changes while the profile is attached to a subnet/session.
# terraform import cato_pop_bgp_profile.example 12345
