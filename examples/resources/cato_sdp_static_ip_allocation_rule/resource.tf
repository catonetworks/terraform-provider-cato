

// allocate a static IP to a single user
resource "cato_sdp_static_ip_allocation_rule" "allocate_user_ip" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name    = "Allocate Static IP to jdoe"
    enabled = true
    ip      = "192.0.2.241"
    source = {
      user = {
        # id   = "123456"
        name = "jdoe"
      }
    }
  }
}

resource "cato_sdp_static_ip_allocation_section" "ssia_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "SSIA Section"
  }
}

# cato_sdp_static_ip_allocation_rule.kitchen_sink:
resource "cato_sdp_static_ip_allocation_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_sdp_static_ip_allocation_section.ssia_section.section.id
  }
  rule = {
    name        = "SDP Static IP Allocation Kitchen Sink"
    description = "test"
    enabled     = true
    ip          = "192.0.2.242"
    source = {
      user = {
        # id   = "123457"
        name = "Test User"
      }
    }
  }
}
