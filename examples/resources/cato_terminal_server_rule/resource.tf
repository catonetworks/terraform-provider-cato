

// allow a single global IP range as the terminal server host
resource "cato_terminal_server_rule" "allow_host" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name    = "Allow Terminal Server Host"
    enabled = true
    allowed_host_ip = {
      # id   = "1757826"
      name = "global_ip_range"
    }
  }
}

resource "cato_terminal_server_section" "ts_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Terminal Server Section"
  }
}

# cato_terminal_server_rule.kitchen_sink:
resource "cato_terminal_server_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_terminal_server_section.ts_section.section.id
  }
  rule = {
    name        = "Terminal Server Kitchen Sink"
    description = "test"
    enabled     = true
    allowed_host_ip = {
      # id   = "1757826"
      name = "global_ip_range"
    }
    exclude_traffic = [
      {
        # id   = "1910542"
        name = "global_ip_range2"
      },
    ]
  }
}
