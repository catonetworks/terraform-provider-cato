

// enforce Always On for a single user
resource "cato_always_on_rule" "enforce_user" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name              = "Enforce Always On for jdoe"
    enabled           = true
    action            = "ENFORCE"
    allow_fail_open   = false
    allow_user_bypass = false
    anti_tamper_mode  = "OFF"
    bypass_duration = {
      time = 30
      unit = "MINUTES"
    }
    source = {
      users = [
        {
          # id   = "123456"
          name = "jdoe"
        },
      ]
    }
  }
}

resource "cato_always_on_section" "ao_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "Always-On Section"
  }
}

# cato_always_on_rule.kitchen_sink:
resource "cato_always_on_rule" "kitchen_sink" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_always_on_section.ao_section.section.id
  }
  rule = {
    name              = "Always-On Kitchen Sink"
    description       = "test"
    enabled           = true
    action            = "ENFORCE"
    allow_fail_open   = true
    allow_user_bypass = true
    anti_tamper_mode  = "PROTECT"
    bypass_duration = {
      time = 1
      unit = "HOURS"
    }
    platform = [
      "WINDOWS",
      "MACOS",
    ]
    device_posture_profile = [
      {
        # id   = "4202"
        name = "Test Device Posture Profile"
      },
    ]
    source = {
      users = [
        {
          # id   = "0"
          name = "test user"
        },
      ]
      user_groups = [
        {
          # id   = "500000000"
          name = "Test User Group"
        },
      ]
    }
  }
}
