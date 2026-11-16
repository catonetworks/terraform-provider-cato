

// Unlike the other seven rule types in this batch, cato_private_access_rule has no
// `at` placement block in its schema - ordering and publishing are handled entirely
// by the separate cato_private_access_rule_bulk resource (see its own example).

// allow access to a single private application
resource "cato_private_access_rule" "allow_private_app" {
  action  = "ALLOW"
  enabled = true
  name    = "Allow Private App"
  applications = [
    {
      # id   = "54321"
      name = "private-app-1"
    },
  ]
  source = {
    users = [
      {
        # id   = "0"
        name = "test user"
      },
    ]
  }
}

# cato_private_access_rule.kitchen_sink:
resource "cato_private_access_rule" "kitchen_sink" {
  action      = "BLOCK"
  description = "test"
  enabled     = true
  name        = "Private Access Kitchen Sink"
  connection_origins = [
    "REMOTE_CLIENT",
    "SITE",
  ]
  platforms = [
    "LINUX",
    "WINDOWS",
  ]
  applications = [
    {
      # id   = "54321"
      name = "private-app-1"
    },
  ]
  countries = [
    {
      # id   = "US"
      name = "United States"
    },
  ]
  devices = [
    {
      # id   = "4202"
      name = "Test Device Posture Profile"
    },
  ]
  active_period = {
    effective_from     = "2026-01-20T03:04:00Z"
    use_effective_from = true
    expires_at         = "2027-02-20T03:04:00Z"
    use_expires_at     = true
  }
  schedule = {
    active_on = "WORKING_HOURS"
    custom_timeframe = {
      from = "2026-02-20T01:02:00Z"
      to   = "2026-02-20T03:04:00Z"
    }
    custom_recurring = {
      days = ["MONDAY", "FRIDAY"]
      from = "08:04:00"
      to   = "19:30:00"
    }
  }
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
  tracking = {
    event = {
      enabled = true
    }
    alert = {
      enabled   = true
      frequency = "DAILY"
      mailing_list = [
        {
          # id   = "623603"
          name = "mailing-list-1"
        },
      ]
      subscription_group = [
        {
          # id   = "623604"
          name = "subscription-group-1"
        },
      ]
      webhook = [
        {
          # id   = "623605"
          name = "webhook-1"
        },
      ]
    }
  }
  user_attributes = {
    risk_score = {
      category = "LOW"
      operator = "LTE"
    }
  }
}
