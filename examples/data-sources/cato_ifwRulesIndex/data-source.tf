## Providers ###
provider "cato" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = var.cato_token
    account_id = var.account_id
}

### Data Source ###
data "cato_ifwRulesIndex" "all_rules" {}