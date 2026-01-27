resource "cato_app_connector" "test100" {
  name  = "test-connector-200"
  type  = "VIRTUAL"
  group = "conn-group-2"
  address = {
    country = { id = "US" }
    state   = "Virginia"
    city    = "Richmond"
    street  = null
    zipCode = null
  }
  timezone = "America/New_York"
  preferred_pop_location = {
    preferred_only = true
    automatic      = false
    primary        = { name = "New York_Sta" }
    secondary      = { name = "Los Angeles Sta" }
  }
}
