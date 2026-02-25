resource "cato_app_connector" "example" {
  name        = "example-app-connector"
  description = "example-app-connector description"
  group_name  = "example-group"
  location = {
    address      = "123 Main St"
    city_name    = "Chicago"
    country_code = "US"
    state_code   = "US-CA"
    timezone     = "America/New_York"
  }
  preferred_pop_location = {
    automatic      = false
    preferred_only = true
    primary        = { name = "New York" }
    secondary      = { name = "Chicago" }
  }
  type = "VIRTUAL"
}

