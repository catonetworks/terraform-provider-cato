## Providers ###
provider "cato-oss" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = var.cato_token
    account_id = var.account_id
}

### Data Source ###
data "cato-oss_accountSnapshotSite" "aws-dev-site" {
	id = var.site_id
}