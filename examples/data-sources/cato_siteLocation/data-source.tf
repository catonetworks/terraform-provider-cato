## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source Examples ###

# Example 1: Filter by country_code (ISO code)
data "cato_siteLocation" "us_locations" {
  filters {
    field     = "country_code"
    search    = "US"
    operation = "exact"
  }
}

# Example 2: Filter by state_code (ISO 3166-2 code)
data "cato_siteLocation" "california" {
  filters {
    field     = "state_code"
    search    = "US-CA"
    operation = "exact"
  }
}

# Example 3: Filter by city name with startsWith
data "cato_siteLocation" "san_cities" {
  filters {
    field     = "city"
    search    = "San"
    operation = "startsWith"
  }
  filters {
    field     = "country_code"
    search    = "US"
    operation = "exact"
  }
}

# Example 4: Filter by country_name and state_name
data "cato_siteLocation" "ny" {
  filters {
    field     = "city"
    search    = "New York"
    operation = "startsWith"
  }
  filters {
    field     = "state_name"
    search    = "New York"
    operation = "exact"
  }
  filters {
    field     = "country_name"
    search    = "United States"
    operation = "exact"
  }
}

# Example 5: Filter by city with contains operation
data "cato_siteLocation" "diego" {
  filters {
    field     = "city"
    search    = "Diego"
    operation = "contains"
  }
}

## Available Filter Fields:
# - city: Filter by city name
# - state_code: Filter by state ISO code (e.g., "US-CA", "US-NY")
# - state_name: Filter by state name (e.g., "California", "New York")
# - country_code: Filter by country ISO code (e.g., "US", "GB", "DE")
# - country_name: Filter by country name (e.g., "United States", "Germany")
#
## Important: Mutually exclusive fields
# - Use state_code OR state_name, but NOT both in the same data source
# - Use country_code OR country_name, but NOT both in the same data source
#
## Available Filter Operations:
# - exact: Exact match
# - startsWith: Matches values starting with the search string
# - endsWith: Matches values ending with the search string
# - contains: Matches values containing the search string

## Example Response Output ###
# data "cato_siteLocation" "california" returns:
# locations = [
#   {
#     city         = "Los Angeles"
#     country_code = "US"
#     country_name = "United States"
#     state_code   = "US-CA"
#     state_name   = "California"
#     timezone     = ["America/Los_Angeles"]
#   },
#   {
#     city         = "San Diego"
#     country_code = "US"
#     country_name = "United States"
#     state_code   = "US-CA"
#     state_name   = "California"
#     timezone     = ["America/Los_Angeles"]
#   },
#   {
#     city         = "San Francisco"
#     country_code = "US"
#     country_name = "United States"
#     state_code   = "US-CA"
#     state_name   = "California"
#     timezone     = ["America/Los_Angeles"]
#   },
#   # ... more cities
# ]
