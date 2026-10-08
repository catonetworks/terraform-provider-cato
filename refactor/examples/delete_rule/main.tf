terraform {
  required_providers {
    cato = {
      source = "catonetworks/cato-refactor"
    }
  }
}

variable "account_id" {
  type        = string
  description = "Cato account containing the existing private-access rule."
}

provider "cato" {
  account_id = var.account_id
  # baseurl and token are read from CATO_BASEURL and CATO_TOKEN.
}

# Import an existing rule into this resource, then destroy it.
# Creation and updates are outside this deletion POC.
resource "cato_private_access_rule" "existing" {}
