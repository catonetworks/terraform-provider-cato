// LAN Firewall sub-policy 

resource "cato_lf_sub_policy" "example" {
  name        = "example"
  description = "example sub-policy"

  at = {
    position = "FIRST_IN_POLICY"
  }
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
