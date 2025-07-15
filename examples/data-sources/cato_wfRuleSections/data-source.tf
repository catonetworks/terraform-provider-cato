## Providers ###
provider "cato" {
    baseurl = "https://api.catonetworks.com/api/v1/graphql2"
    token = var.cato_token
    account_id = var.account_id
}

### Data Source ###
data "cato_wfRuleSections" "all_wf_sections" {}

data "cato_wfRuleSections" "my_section" {
  name_filter = ["My WAN Section"]
}

module "wf_rules" {
  source = "catonetworks/bulk-wf-rules/cato"
  wf_rules_json_file_path = "config_data/all_wf_rules_and_sections.json"
  section_to_start_after_id = data.cato_wfRuleSections.my_section.items[0].id
}