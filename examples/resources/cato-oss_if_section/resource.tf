// internet firewall section last in policy
resource "cato-oss_if_section" "if_section_1" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "tf section"
  }
}

// internet firewall section after "if_section_1" previously created
resource "cato-oss_if_section" "if_section_2" {
  at = {
    position = "AFTER_SECTION"
    ref      = cato-oss_if_section.if_section_1.section.id
  }
  section = {
    name = "tf section 2"
  }
}


// internet firewall rule using section id
resource "cato-oss_if_section" "if_section_1" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "tf section"
  }
}

resource "cato-oss_if_rule" "if_rule_1" {
  at = {
    position = "FIRST_IN_SECTION"
    ref      = cato-oss_if_section.if_section_1.section.id
  }
  rule = {
    name        = "test"
    description = "terraform test rules"
    enabled     = false
    action      = "ALLOW"
  }
}