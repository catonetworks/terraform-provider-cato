## Providers ###
provider "cato" {
  baseurl    = "https://api.catonetworks.com/api/v1/graphql2"
  token      = var.cato_token
  account_id = var.account_id
}

### Data Source ###
data "cato_licensingInfo" "all_licenses" {
}

data "cato_licensingInfo" "pooled_bandwidth_licenses" {
  sku = "CATO_PB"
}

data "cato_licensingInfo" "active_unsassigned_site_licenses" {
  is_active   = true
  is_assigned = false
  sku         = "CATO_SITE"
}