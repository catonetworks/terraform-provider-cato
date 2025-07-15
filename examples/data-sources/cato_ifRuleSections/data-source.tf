## Providers ###
provider "cato" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = var.cato_token
    account_id = var.account_id
}

### Data Source ###
data "cato_ifRuleSections" "all_wf_sections" {}

data "cato_ifRuleSections" "default_section" {
  name_filter = ["Default Internet Rules"]
}

module "if_rules" {
  source = "catonetworks/bulk-if-rules/cato"
  ifw_rules_json_file_path = "config_data/all_if_rules_and_sections.json"
  section_to_start_after_id = data.cato_ifRuleSections.default_section.items[0].id
}