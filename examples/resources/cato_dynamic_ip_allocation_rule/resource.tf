

// allocate dynamic IPs from a global IP range
resource "cato_dynamic_ip_allocation_rule" "allocate_range" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name    = "Allocate Dynamic IPs"
    enabled = true
    range = {
      global_ip_range = {
        # id   = "1757826"
        name = "global_ip_range"
      }
    }
  }
}

resource "cato_dynamic_ip_allocation_section" "dia_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "DIA Section"
  }
}

# cato_dynamic_ip_allocation_rule.kitchen_sink:
resource "cato_dynamic_ip_allocation_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_dynamic_ip_allocation_section.dia_section.section.id
  }
  rule = {
    name        = "Dynamic IP Allocation Kitchen Sink"
    description = "test"
    enabled     = true
    range = {
      global_ip_range = {
        # id   = "1757826"
        name = "global_ip_range"
      }
    }
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
    device_posture_profile = [
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
  }
}
