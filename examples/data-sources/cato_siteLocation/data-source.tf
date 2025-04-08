## Providers ###
provider "cato" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = var.cato_token
    account_id = var.account_id
}

### Data Source ###
data "cato_siteLocation" "ny" {
  filters = [{
    field = "city"
    search = "New York"
    operation = "startsWith"
  },
  {
    field = "state_name"
    search = "New York"
    operation = "exact"
  },
 {
    field = "country_name"
    search = "United"
    operation = "contains"
  }]
}

## Example Response Output ###
data "cato_siteLocation" "ny" {
    locations = [
        {
            city         = "New York Mills"
            country_code = "US"
            country_name = "United States"
            state_code   = "US-NY"
            state_name   = "New York"
            timezone     = "America/New_York"
        },
        {
            city         = "New York City"
            country_code = "US"
            country_name = "United States"
            state_code   = "US-NY"
            state_name   = "New York"
            timezone     = "America/New_York"
        },
    ]
}