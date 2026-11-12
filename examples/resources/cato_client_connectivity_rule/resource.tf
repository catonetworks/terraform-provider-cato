

// allow high-confidence client connections
resource "cato_client_connectivity_rule" "allow_high_confidence" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name             = "Allow High Confidence Connections"
    enabled          = true
    action           = "ALLOW"
    confidence_level = "HIGH"
  }
}

resource "cato_client_connectivity_section" "cc_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Client Connectivity Section"
  }
}

# cato_client_connectivity_rule.kitchen_sink:
resource "cato_client_connectivity_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_client_connectivity_section.cc_section.section.id
  }
  rule = {
    name              = "Client Connectivity Kitchen Sink"
    description       = "test"
    enabled           = true
    action            = "ALLOW"
    confidence_level  = "HIGH"
    connection_origin = ["REMOTE", "REMOTE_EXTENSION"]
    platform = [
      "WINDOWS",
      "MACOS",
    ]
    country = [
      {
        # id   = "AG"
        name = "Antigua and Barbuda"
      },
    ]
    device = [
      {
        # id   = "4202"
        name = "Test Device Posture Profile"
      },
    ]
    source = {
      user = [
        {
          # id   = "0"
          name = "test user"
        },
      ]
      users_group = [
        {
          # id   = "500000000"
          name = "Test User Group"
        },
      ]
    }
    source_range = [
      {
        global_ip_range = {
          # id   = "1757826"
          name = "global_ip_range"
        }
      },
    ]
  }
}
