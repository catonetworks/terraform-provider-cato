// internet firewall section last in policy
resource "cato_wf_section" "wf_section_1" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "tf section"
  }
}

// internet firewall section after "wf_section_1" previously created
resource "cato_wf_section" "wf_section_2" {
  at = {
    position = "AFTER_SECTION"
    ref      = cato_wf_section.wf_section_1.section.id
  }
  section = {
    name = "tf section 2"
  }
}


// internet firewall rule using section id
resource "cato_wf_section" "wf_section_1" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "tf section"
  }
}

resource "cato_wf_rule" "wf_rule_1" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato_wf_section.wf_section_1.section.id
  }
  rule = {
    name        = "test"
    description = "terraform test rules"
    enabled     = false
    action      = "ALLOW"
    direction   = "BOTH"
  }
}


